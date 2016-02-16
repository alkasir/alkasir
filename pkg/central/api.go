package central

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/shared/apierrors"
	"github.com/alkasir/alkasir/pkg/shared/apiutils"
	"github.com/ant0ine/go-json-rest/rest"
	version "github.com/hashicorp/go-version"
	"github.com/thomasf/lg"
)

// apiMux creates the servermux for the json api server
func apiMux(dbclients db.Clients) (*http.ServeMux, error) {
	var routes = []*rest.Route{
		{"POST", "/v1/suggestions/new/", SuggestionToken(dbclients)},
		{"POST", "/v1/samples/", StoreSample(dbclients)},
		{"POST", "/v1/hosts/", GetHosts(dbclients)},
		{"POST", "/v1/upgrades/", GetUpgrade(dbclients)},
	}
	mux := http.NewServeMux()
	api := defaultAPI("central")
	router, err := rest.MakeRouter(routes...)
	if err != nil {
		return nil, err
	}
	api.SetApp(router)
	handler := api.MakeHandler()

	mux.Handle("/", handler)
	return mux, nil
}

func apiError(w rest.ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	if lg.V(5) {
		lg.InfoDepth(1, fmt.Sprintf("%d: %s", code, error))
	}
	err := w.WriteJson(map[string]string{
		"Error": error,
		"Ok":    "false",
	})
	if err != nil {
		lg.Error(err)
		return
	}
}

// relatedHosts .
type relatedHosts struct {
	sync.RWMutex
	items     map[string][]string
	dbclients db.Clients
}

func (r *relatedHosts) update() {
	lg.V(19).Infoln("updating related hosts..")
	related, err := r.dbclients.DB.GetRelatedHosts()
	if err != nil {
		lg.Fatal(err)
	}
	curated := make(map[string][]string, len(related))
	for k, v := range related {
		curated[strings.TrimPrefix(k, "www.")] = v
	}
	r.Lock()
	r.items = curated
	r.Unlock()
}

func (r *relatedHosts) fill(hosts []string) []string {
	items := make(map[string]bool, len(hosts))
	r.RLock()
	defer r.RUnlock()
	for _, h := range hosts {
		trHost := strings.TrimPrefix(h, "www.")
		items[trHost] = true
		if addhosts, ok := r.items[trHost]; ok {
			for _, v := range addhosts {
				trHost := strings.TrimPrefix(v, "www.")
				if trHost != "" {
					items[trHost] = true
				}
			}
		}
	}
	var result []string
	for k := range items {
		result = append(result, k)
	}
	return result
}

// Update list of blocked hosts for an IP address.
func GetHosts(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	relh := relatedHosts{
		dbclients: dbclients,
	}
	relh.update()
	go func() {
		for range time.NewTicker(10 * time.Minute).C {
			relh.update()
		}
	}()
	return func(w rest.ResponseWriter, r *rest.Request) {

		// HANDLE USERIP BEGIN
		req := shared.UpdateHostlistRequest{}
		err := r.DecodeJsonPayload(&req)
		if err != nil {
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}

		// parse/validate client ip address.
		IP := req.ClientAddr
		if IP == nil {
			apiError(w, "bad ClientAddr", http.StatusBadRequest)
			return
		}

		// resolve ip to asn.
		var ASN int
		ASNres, err := dbclients.Internet.IP2ASN(IP)
		if err != nil {
			lg.Errorln(shared.SafeClean(err.Error()))
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}
		if ASNres != nil {
			ASN = ASNres.ASN
		} else {
			lg.Warningf("no ASN lookup result for IP: %s ", shared.SafeClean(IP.String()))
		}

		// resolve ip to country code.
		countryCode := dbclients.Maxmind.IP2CountryCode(IP)

		req.ClientAddr = net.IPv4zero
		IP = net.IPv4zero
		// HANDLE USERIP END

		hosts, err := dbclients.DB.GetBlockedHosts(countryCode, ASN)
		if err != nil {
			apiError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		simpleData := struct {
			ClientVersion string `json:"version"` // client version idientifer
		}{
			req.ClientVersion,
		}
		data, err := json.Marshal(simpleData)
		if err != nil {
			lg.Errorln(err)
		}

		ss := db.SimpleSample{
			CountryCode: countryCode,
			ASN:         ASN,
			Type:        "ClientBlocklistUpdate",
			OriginID:    req.UpdateID,
			Data:        data,
		}
		err = dbclients.DB.InsertSimpleSample(ss)
		if err != nil {
			lg.Errorf("error persisting simplesample %v", ss)
		}

		err = w.WriteJson(shared.UpdateHostlistResponse{
			Hosts: relh.fill(hosts),
		})
		if err != nil {
			lg.Error(err)
			return
		}

	}
}

// SuggestionToken JSON API method.
func SuggestionToken(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		// HANDLE USERIP BEGIN
		req := shared.SuggestionTokenRequest{}
		err := r.DecodeJsonPayload(&req)
		if err != nil {
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}

		// validate country code.
		if !validCountryCode(req.CountryCode) {
			apiError(w,
				fmt.Sprintf("invalid country code: %s", req.CountryCode),
				http.StatusBadRequest)
			return
		}

		// parse/validate client ip address.

		IP := req.ClientAddr
		if IP == nil {
			apiError(w, "bad ClientAddr", http.StatusBadRequest)
			return
		}

		// parse and validate url.
		URL := strings.TrimSpace(req.URL)
		if URL == "" {
			apiError(w, "no or empty URL", http.StatusBadRequest)
			return
		}

		u, err := url.Parse(URL)
		if err != nil {
			apiError(w, fmt.Sprintf("%s is not a valid URL", URL), http.StatusBadRequest)
			return
		}

		if !shared.AcceptedURL(u) {
			apiError(w, fmt.Sprintf("%s is not a valid URL", URL), http.StatusBadRequest)
			return
		}

		// resolve ip to asn.
		var ASN int
		ASNres, err := dbclients.Internet.IP2ASN(IP)
		if err != nil {
			lg.Errorln(shared.SafeClean(err.Error()))
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}
		if ASNres != nil {
			ASN = ASNres.ASN
		} else {
			lg.Warningf("no ASN lookup result for IP: %s ", shared.SafeClean(IP.String()))
		}

		// reoslve ip to country code.
		countryCode := dbclients.Maxmind.IP2CountryCode(IP)

		// resolve ip to city geonameid
		geoCityID := dbclients.Maxmind.IP2CityGeoNameID(IP)

		req.ClientAddr = net.IPv4zero
		IP = net.IPv4zero
		// HANDLE USERIP END

		{
			supported, err := dbclients.DB.IsURLAllowed(u, countryCode)
			if err != nil {
				// TODO: standardize http status codes
				apiError(w, err.Error(), http.StatusForbidden)
				return
			}
			if !supported {
				lg.Infof("got request for unsupported URL %s")
				w.WriteJson(shared.SuggestionTokenResponse{
					Ok:  false,
					URL: req.URL,
				})
				return
			}
		}
		// start new submission token session
		token := db.SessionTokens.New(URL)

		// create newclienttoken sample data
		sample := shared.NewClientTokenSample{
			URL:         URL,
			CountryCode: req.CountryCode,
		}
		sampleData, err := json.Marshal(sample)
		if err != nil {
			lg.Errorln(err)
			apiError(w, "error #20150424-002542-CEST", http.StatusInternalServerError)
			return
		}

		// create extraData
		extra := shared.IPExtraData{
			CityGeoNameID: geoCityID,
		}
		extraData, err := json.Marshal(extra)
		if err != nil {
			lg.Errorln(err)
			apiError(w, "error #20150427-211052-CEST", http.StatusInternalServerError)
			return
		}

		// insert into db
		{
			err := dbclients.DB.InsertSample(db.Sample{
				Host:        u.Host,
				CountryCode: countryCode,
				ASN:         ASN,
				Type:        "NewClientToken",
				Origin:      "Central",
				Token:       token,
				Data:        sampleData,
				ExtraData:   extraData,
			})
			if err != nil {
				apiError(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// queue central measurements
		measurements, err := measure.DefaultMeasurements(req.URL)
		if err != nil {
			lg.Warningf("could not create standard measurements: %s", err.Error())
		} else {
			queueMeasurements(token, measurements...)
		}

		// write json response
		{
			err := w.WriteJson(shared.SuggestionTokenResponse{
				Ok:    true,
				URL:   URL,
				Token: token,
			})
			if err != nil {
				lg.Errorln(err.Error())
				return
			}
		}

	}
}

var clientSampleTypes = map[string]bool{
	"HTTPHeader": true,
	"DNSQuery":   true,
}

// StoreSample JSON API method.
func StoreSample(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		// HANDLE USERIP BEGIN
		req := shared.StoreSampleRequest{}
		err := r.DecodeJsonPayload(&req)
		if err != nil {
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}

		// validate sample type.
		if _, ok := clientSampleTypes[req.SampleType]; !ok {
			apiError(w, "invalid sample type: "+req.SampleType, http.StatusBadRequest)
			return
		}

		// get/validate suggestion session token.
		tokenData, validToken := db.SessionTokens.Get(req.Token)
		if !validToken {
			apiError(w, "invalid token", http.StatusBadRequest)
			return
		}

		// parse/validate url.
		URL := strings.TrimSpace(req.URL)
		if tokenData.URL != URL {
			lg.V(2).Infof("invalid URL %s for token session, was expecting %s", req.URL, tokenData.URL)
			apiError(w, "invalid URL for token session", http.StatusBadRequest)
			return
		}
		u, err := url.Parse(URL)
		if err != nil {
			apiError(w, fmt.Sprintf("%s is not a valid URL", URL), http.StatusBadRequest)
			return
		}

		// parse/validate client ip address.
		IP := net.ParseIP(req.ClientAddr)
		if IP == nil {
			apiError(w, "bad ClientAddr", http.StatusBadRequest)
			return
		}

		// resolve ip to asn.
		var ASN int
		ASNres, err := dbclients.Internet.IP2ASN(IP)
		if err != nil {
			lg.Errorln(shared.SafeClean(err.Error()))
			apiError(w, shared.SafeClean(err.Error()), http.StatusInternalServerError)
			return
		}
		if ASNres != nil {
			ASN = ASNres.ASN
		} else {
			lg.Warningf("no ASN lookup result for IP: %s ", shared.SafeClean(IP.String()))
		}

		countryCode := dbclients.Maxmind.IP2CountryCode(IP)

		req.ClientAddr = ""
		IP = net.IPv4zero
		// HANDLE USERIP END

		// insert into db
		{
			err := dbclients.DB.InsertSample(db.Sample{
				Host:        u.Host,
				CountryCode: countryCode,
				ASN:         ASN,
				Type:        req.SampleType,
				Origin:      "Client",
				Token:       req.Token,
				Data:        []byte(req.Data),
			})
			if err != nil {
				lg.Errorln(err.Error())
				apiError(w, "error 20150424-011948-CEST", http.StatusInternalServerError)
				return
			}
		}

		w.WriteJson(shared.StoreSampleResponse{
			Ok: true,
		})
		return
	}
}

// Update list of blocked hosts for an IP address.
func GetUpgrade(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		req := shared.BinaryUpgradeRequest{}
		err := r.DecodeJsonPayload(&req)
		// TODO: proper validation response
		if err != nil {
			apiError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if req.FromVersion == "" {
			apiError(w, "fromversion required", http.StatusInternalServerError)
			return
		}
		v, err := version.NewVersion(req.FromVersion)
		if err != nil {
			lg.Warningln(err)
			apiError(w, fmt.Sprintf("invalid fromVersion %s", req.FromVersion), http.StatusInternalServerError)
			return
		}

		if req.Artifact == "" {
			apiError(w, "artifact required", http.StatusInternalServerError)
			return

		}
		res, found, err := dbclients.DB.GetUpgrade(
			db.GetUpgradeQuery{
				Artifact: req.Artifact,
			})
		if err != nil {
			apiError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !found {
			apiutils.WriteRestError(w,
				apierrors.NewNotFound(
					"upgrade", fmt.Sprintf("%s-%s", req.Artifact, req.FromVersion)))
			return
		}

		serverVersion, err := version.NewVersion(res.Version)
		if err != nil {
			apiError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if serverVersion.LessThan(v) || serverVersion.Equal(v) {
			apiutils.WriteRestError(w,
				apierrors.NewNotFound(
					"upgrade", fmt.Sprintf("%s-%s", req.Artifact, req.FromVersion)))
			return

		}
		w.WriteJson(shared.BinaryUpgradeResponse{
			Artifact:         res.Artifact,
			Version:          res.Version,
			CreatedAt:        res.CreatedAt,
			SHA256Sum:        res.SHA256Sum,
			ED25519Signature: res.ED25519Signature,
		})

	}
}
