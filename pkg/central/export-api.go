// public/partner api, runs on own port because the need to route through different networks.
package central

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/shared/jwtmw"
	"github.com/alkasir/alkasir/pkg/shared/linkheader"
	"github.com/ant0ine/go-json-rest/rest"
)

// apiMux creates the servermux for the json api server
func apiMuxExport(dbclients db.Clients, secretKey []byte) (*http.ServeMux, error) {
	jwtm := &jwtmw.JWTMiddleware{
		Key:        secretKey,
		Realm:      "jwt auth",
		Timeout:    time.Hour,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string) bool {
			ok, cred, err := dbclients.DB.GetExportAPIAuthCredentials(userId)
			switch {
			case !ok:
				return false
			case err != nil:
				return false
			}
			ok, err = cred.IsValid(password)
			if err != nil {
				return false
			}
			return ok
		},
	}
	var routes = []*rest.Route{
		{"POST", "/login", jwtm.LoginHandler},
		{"GET", "/v1/blocked/", GetBlockedHostsExport(dbclients)},
		{"GET", "/v1/samples/", GetSamplesExport(dbclients)},
		{"GET", "/v1/simple_samples/", GetSimpleSamplesExport(dbclients)},
	}
	mux := http.NewServeMux()
	api := defaultAPI("export_api")

	api.Use(&rest.IfMiddleware{
		Condition: func(request *rest.Request) bool {
			return request.URL.Path != "/login"
		},
		IfTrue: jwtm,
	})

	router, err := rest.MakeRouter(routes...)
	if err != nil {
		return nil, err
	}
	api.SetApp(router)
	handler := api.MakeHandler()

	mux.Handle("/", handler)
	return mux, nil
}

// GetBlockedHostsExport
func GetBlockedHostsExport(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		req := shared.BlockedContentRequest{}
		q := r.Request.URL.Query()

		if q.Get("id_max") != "" {
			idmax, err := strconv.Atoi(q.Get("id_max"))
			if err != nil {
				panic(err)
			}
			req.IDMax = idmax
		}

		hosts, nextpage, err := dbclients.DB.GetExportBlockedHosts(req)

		lh := linkheader.NewLinkHeader(r.URL, "id_max")
		r.URL.RequestURI()
		if len(hosts) > 0 {
			lh.Current(hosts[0].ID)
		}
		// len(hosts) > 0
		if nextpage != "" {
			lh.Next(hosts[len(hosts)-1].ID)
		}

		lh.SetHeader(w.Header())
		err = w.WriteJson(hosts)
		if err != nil {
			panic(err)
		}

	}
}

func GetSamplesExport(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		req := shared.ExportSampleRequest{}
		q := r.Request.URL.Query()

		if q.Get("id_max") != "" {
			idmax, err := strconv.Atoi(q.Get("id_max"))
			if err != nil {
				panic(err)
			}
			req.IDMax = idmax
		}

		samples, nextpage, err := dbclients.DB.GetExportSamples(req)

		lh := linkheader.NewLinkHeader(r.URL, "id_max")
		r.URL.RequestURI()
		if len(samples) > 0 {
			lh.Current(samples[0].ID)
		}
		if nextpage != "" {
			lh.Next(samples[len(samples)-1].ID)
		}
		lh.SetHeader(w.Header())
		err = w.WriteJson(samples)
		if err != nil {
			panic(err)
		}
	}
}

func GetSimpleSamplesExport(dbclients db.Clients) func(w rest.ResponseWriter, r *rest.Request) {
	return func(w rest.ResponseWriter, r *rest.Request) {
		req := shared.ExportSimpleSampleRequest{}
		q := r.Request.URL.Query()

		if q.Get("id_max") != "" {
			idmax, err := strconv.Atoi(q.Get("id_max"))
			if err != nil {
				panic(err)
			}
			req.IDMax = idmax
		}

		samples, nextpage, err := dbclients.DB.GetExportSimpleSamples(req)

		lh := linkheader.NewLinkHeader(r.URL, "id_max")
		r.URL.RequestURI()
		if len(samples) > 0 {
			lh.Current(samples[0].ID)
		}
		if nextpage != "" {
			lh.Next(samples[len(samples)-1].ID)
		}

		lh.SetHeader(w.Header())
		err = w.WriteJson(samples)
		if err != nil {
			panic(err)
		}
	}
}
