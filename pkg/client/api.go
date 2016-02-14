package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alkasir/alkasir/pkg/browsercode"
	"github.com/alkasir/alkasir/pkg/central/client"
	"github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/debugexport"
	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
	"github.com/alkasir/alkasir/pkg/pac"
	"github.com/alkasir/alkasir/pkg/service"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/shared/apierrors"
	"github.com/alkasir/alkasir/pkg/shared/apiutils"
	"github.com/alkasir/alkasir/pkg/shared/fielderrors"
	"github.com/alkasir/alkasir/pkg/shared/middlewares"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/thomasf/lg"
	"h12.me/socks"
)

// ClientAPIVersion is used to verify that the browser/extension and client can
// speak to eachother.
//
const ClientAPIVersion = 2

// routes for alkasir-client
var routes = []*rest.Route{

	// user resources // edit
	{"GET", "/pac/", GetPAC},
	{"GET", "/pac/#whatever/", GetPAC},
	{"GET", "/notifications/", GetNotifications},
	{"GET", "/settings/", GetUserSettings},
	{"POST", "/settings/", PostUserSettings},
	{"GET", "/status/summary/", GetStatusSummary},
	{"GET", "/version/", GetVersion},
	{"GET", "/suggestions/:id/", GetSuggestion},
	{"GET", "/suggestions/", GetSuggestions},
	{"POST", "/suggestions/:id/submit/", SubmitSuggestion},
	{"POST", "/suggestions/", CreateSuggestion},

	{"POST", "/transports/traffic/", PostTransportTraffic},
	{"GET", "/transports/traffic/", GetTransportTraffic},

	{"GET", "/export/debug/", GetDebug},
	{"POST", "/export/chrome-extension/", PostExportChromeExtension},
	{"POST", "/browsercode/toclipboard/", PostBrowsercodeToClipboard},

	{"POST", "/log/#sender/", PostLog},
	{"POST", "/connections/validate/", ValidateConnectionString},
	{"GET", "/connections/", GetConnections},
	{"POST", "/connections/", PostConnection},
	{"POST", "/connections/:id/toggle/", ToggleConnection},
	{"DELETE", "/connections/:id/", DeleteConnection},

	// dev resources
	{"GET", "/services/", GetAllServices},
	{"GET", "/methods/", GetAllMethods},
	{"POST", "/methods/:id/test/", TestMethod},
	{"GET", "/hostpatterns/", GetAllHostPatterns},
}

func GetPAC(w rest.ResponseWriter, r *rest.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.(http.ResponseWriter).Write(pac.GenPAC())
}

type NotificationAction struct {
	Name  string `json:"name"`
	Route string `json:"route"`
}

// Notification represents a user notifcation event which also can have some
// interactivity configured.
type Notification struct {
	ID          string               `json:"id"`        // UUID
	EventTime   time.Time            `json:"eventTime"` // A timestamp associated with the notification
	Level       string               `json:"level"`     // info,success,danger,warning: for styling
	Priority    int                  `json:"priority"`  // Priority ranges from -2 to 2. -2 is lowest priority. 2 is highest. Zero is default.
	Title       string               `json:"title"`     // header
	Message     string               `json:"message"`   // body
	Actions     []NotificationAction `json:"actions"`
	Dismissable bool                 `json:"dismissable"` // can the user dismiss this
}

var notificationIDgen, _ = shared.NewIDGen("notify")

func prepareNotifications(notifications []Notification) []Notification {
	for k, v := range notifications {
		if v.ID == "" {
			v.ID = notificationIDgen.New()
			notifications[k] = v
		}
	}
	return notifications
}

func GetNotifications(w rest.ResponseWriter, r *rest.Request) {
	if UITEST {
		GetNotificationsUITest(w, r)
		return
	}

	var notifications []Notification
	conf := clientconfig.Get()
	// if setup is not done it is the only notification
	if !conf.Settings.Local.UserSetup() {
		notifications = append(notifications, Notification{
			Level:   "info",
			Title:   "setup_required_title",
			Message: "setup_required_firstrun_message",
			Actions: []NotificationAction{
				{"action_continue", "/setup-guide/"},
			},
		})
		notifications = prepareNotifications(notifications)
		w.WriteJson(notifications)
		return
	}
	// put the global status of alkasir first
	if service.TransportOk() {
		notifications = append(notifications, Notification{
			Level:   "success",
			Title:   "status_title",
			Message: "status_ok_message",
		})

		notifications = append(notifications, Notification{
			Level:   "info",
			Title:   "suggest_this_title",
			Message: "suggest_this_message",
			Actions: []NotificationAction{
				{"action_continue", "/suggestions/"},
				{"action_help", "/docs/en/report-url/"},
			},
		})

	}

	if !service.TransportOk() {
		notifications = append(notifications, Notification{
			Level:   "danger",
			Title:   "status_title",
			Message: "transport_error_message",
			Actions: []NotificationAction{
				{"action_help", "/docs/en/index"},
			},
		})
	}

	notifications = prepareNotifications(notifications)
	w.WriteJson(notifications)
}

// ConnectionSetting .
type ConnectionSetting struct {
	ID        string `json:"id"`      // runtime only connection ID.
	Name      string `json:"name"`    // display name  generated from hash of connection string.
	Encoded   string `json:"encoded"` // only sent from client to browser, never the other way around.
	Disabled  bool   `json:"disabled"`
	Protected bool   `json:"protected"` // if true, the user cannot delete or modify this connection.
}

type UserSettings struct {
	Language            string   `json:"language"`
	LanguageOptions     []string `json:"languageOptions"`
	CountryCode         string   `json:"countryCode"`
	ClientAutoUpdate    bool     `json:"clientAutoUpdate"`
	BlocklistAutoUpdate bool     `json:"blocklistAutoUpdate"`
}

func GetConnections(w rest.ResponseWriter, r *rest.Request) {
	conf := clientconfig.Get()

	var cSettings []ConnectionSetting
	connections := conf.Settings.Connections

	for _, v := range connections {
		cSettings = append(cSettings, ConnectionSetting{
			ID:        v.ID,
			Disabled:  v.Disabled,
			Protected: v.Protected,
			Name:      v.DisplayName(),
		})

	}
	w.WriteJson(cSettings)
}

func ToggleConnection(w rest.ResponseWriter, r *rest.Request) {
	ID := r.PathParam("id")

	err := clientconfig.Update(func(conf *clientconfig.Config) error {
		// validate that the id exists, if supplied
		all := conf.Settings.Connections
		foundIdx := 0
		found := false
		nEnabled := 0
		for k, v := range all {
			if !v.Disabled {
				nEnabled += 1
			}
			if v.ID == ID {
				found = true
				foundIdx = k
			}

		}

		if !found {
			apiutils.WriteRestError(w, apierrors.NewNotFound("connection", ID))
			return nil
		}
		if nEnabled < 2 && !conf.Settings.Connections[foundIdx].Disabled {
			apiutils.WriteRestError(w,
				apierrors.NewForbidden("connection", ID, errors.New("one connection must be enabled")))
			return nil
		}
		conf.Settings.Connections[foundIdx].Disabled = !conf.Settings.Connections[foundIdx].Disabled
		service.UpdateConnections(conf.Settings.Connections)

		return nil
	})
	if err != nil {
		lg.Errorln(err)
	}
	clientconfig.Write()
	w.WriteJson(true)

}

func DeleteConnection(w rest.ResponseWriter, r *rest.Request) {
	ID := r.PathParam("id")

	err := clientconfig.Update(func(conf *clientconfig.Config) error {

		all := conf.Settings.Connections
		var result []shared.Connection

		// validate that the id exists
		found := false
		nEnabled := 0
		for _, v := range all {
			if v.ID == ID {
				found = true
				if v.Protected {
					apiutils.WriteRestError(w, apierrors.NewForbidden("connection", ID, nil))
					return nil
				}
			} else {
				if !v.Disabled {
					nEnabled += 1
				}
				result = append(result, v)
			}
		}
		if !found {
			apiutils.WriteRestError(w, apierrors.NewNotFound("connection", ID))
			return nil
		}
		if nEnabled < 1 {
			apiutils.WriteRestError(w,
				apierrors.NewForbidden("connection", ID, errors.New("one connection must be enabled")))
			return nil
		}
		service.UpdateConnections(result)
		conf.Settings.Connections = result
		return nil
	})
	if err != nil {
		lg.Errorln(err)
	}
	clientconfig.Write()
	w.WriteJson(true)

}

func PostConnection(w rest.ResponseWriter, r *rest.Request) {
	form := ConnectionSetting{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewBadRequest(err.Error()))
		return
	}

	err = clientconfig.Update(func(conf *clientconfig.Config) error {

		all := conf.Settings.Connections

		var connection shared.Connection // decoded connection
		var invalids fielderrors.ValidationErrorList

		// Verify that the encoded field is set
		{
			if form.Encoded == "" {
				invalids = append(invalids, fielderrors.NewFieldRequired("encoded"))
			}
		}

		// verify that the connection is decodeable
		{
			var err error
			connection, err = shared.DecodeConnection(form.Encoded)
			if err != nil {
				invalids = append(invalids, fielderrors.NewFieldInvalid("encoded", form.Encoded, "invalid formatting"))
			}
		}

		// validate that the id exists, if supplied
		foundIdx := 0
		{
			if form.ID != "" {
				found := false
				for k, v := range all {
					if v.ID == form.ID {
						found = true
						foundIdx = k
					}
				}
				if !found {
					invalids = append(invalids, fielderrors.NewFieldNotFound("id", form.ID))
				}
			}
		}

		// validate that the connection doesnt alreay exist
		{
			encoded, err := connection.Encode()
			if err != nil {
				apiutils.WriteRestError(w, err)
				return nil
			}

			found := false
			for _, v := range all {
				enc2, err := v.Encode()
				if err != nil {
					lg.Errorln(err)
					continue
				}
				if enc2 == encoded {
					found = true
				}
			}
			if found {
				invalids = append(invalids, fielderrors.NewFieldDuplicate("encoded", form.Encoded))
			}
		}

		// end of field validations
		if len(invalids) > 0 {
			apiutils.WriteRestError(w, apierrors.NewInvalid("post-connection", "", invalids))
			return nil
		}

		// add connection to settings and save
		if form.ID == "" {
			connection.EnsureID()
			conf.Settings.Connections = append(conf.Settings.Connections, connection)
		} else {
			conf.Settings.Connections[foundIdx] = connection
		}
		service.UpdateConnections(conf.Settings.Connections)
		return nil
	})
	if err != nil {
		lg.Errorln(err)
	}
	clientconfig.Write()
	w.WriteJson(true)

}

func GetUserSettings(w rest.ResponseWriter, r *rest.Request) {
	conf := clientconfig.Get()

	response := UserSettings{
		Language:            conf.Settings.Local.Language,
		LanguageOptions:     LanguageOptions,
		CountryCode:         conf.Settings.Local.CountryCode,
		ClientAutoUpdate:    conf.Settings.Local.ClientAutoUpdate,
		BlocklistAutoUpdate: conf.Settings.Local.BlocklistAutoUpdate,
	}
	w.WriteJson(response)
}

func PostUserSettings(w rest.ResponseWriter, r *rest.Request) {
	form := UserSettings{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewBadRequest(err.Error()))
		return
	}

	changed := false // true if config really was updated
	err = clientconfig.Update(func(conf *clientconfig.Config) error {

		s := &conf.Settings.Local
		prevLang := s.Language
		if ValidLanguage(form.Language) {
			s.Language = form.Language
		} else {
			s.Language = "en"
		}
		if prevLang != s.Language {
			ui.Language(s.Language)
			changed = true
		}
		if s.CountryCode != form.CountryCode {
			s.CountryCode = form.CountryCode
			changed = true
		}
		if s.ClientAutoUpdate != form.ClientAutoUpdate {
			s.ClientAutoUpdate = form.ClientAutoUpdate
			changed = true
		}
		if s.BlocklistAutoUpdate != form.BlocklistAutoUpdate {
			s.BlocklistAutoUpdate = form.BlocklistAutoUpdate
			changed = true
		}

		return nil
	})
	if changed {
		err := clientconfig.Write()
		if err != nil {
			lg.Errorln(err)
		}
	}

	if err != nil {
		lg.Errorln(err)
	}

}

var lastBlocklistChange time.Time

// StatusSummary is a composite type of various alkasir client
type StatusSummary struct {
	TransportOk         bool      `json:"transportOk"`         // true if there are no problems with transports
	BrowserOk           bool      `json:"browserOk"`           // true if the browser extension is connected
	CentralOk           bool      `json:"centralOk"`           // true if the communication with the central server is good
	CountryCode         string    `json:"countryCode"`         // The currently active country code
	LastBlocklistCheck  time.Time `json:"lastBlocklistCheck"`  // Time blocklist was last checked for
	LastBlocklistChange time.Time `json:"lastBlocklistChange"` // Time blocklist was last updated
	AlkasirVersion      string    `json:"alkasirVersion"`      // Current alkasir version (TODO remove? version ping is also available)
}

// JSON api function
func GetStatusSummary(w rest.ResponseWriter, r *rest.Request) {
	conf := clientconfig.Get()
	version := VERSION
	if version == "" {
		version = "development version"
	}
	summary := &StatusSummary{
		AlkasirVersion:      version,
		CountryCode:         conf.Settings.Local.CountryCode,
		TransportOk:         service.TransportOk(),
		LastBlocklistChange: lastBlocklistChange,
		BrowserOk:           true,
		CentralOk:           true,
	}

	w.WriteJson(summary)
	return
}

// GetVersion returns the alkasir-client version.
func GetVersion(w rest.ResponseWriter, r *rest.Request) {
	version := VERSION
	if version == "" {
		version = "development version"
	}
	summary := &struct {
		AlkasirVersion   string `json:"alkasirVersion"`
		ClientAPIVersion int    `json:"apiVersion"`
	}{
		AlkasirVersion:   version,
		ClientAPIVersion: ClientAPIVersion,
	}
	w.WriteJson(summary)
	return
}

// CreateSuggestion .
func CreateSuggestion(w rest.ResponseWriter, r *rest.Request) {

	form := shared.BrowserSuggestionTokenRequest{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		// apiError(w, err.Error(), http.StatusInternalServerError)
		apiutils.WriteRestError(w, err)
		return
	}

	var invalids fielderrors.ValidationErrorList

	// parse and validate url.
	URL := strings.TrimSpace(form.URL)
	if URL == "" {
		invalids = append(invalids, fielderrors.NewFieldRequired("URL"))
		// apiError(w, "no or empty URL", http.StatusBadRequest)

	}
	u, err := url.Parse(URL)
	if err != nil {
		invalids = append(invalids, fielderrors.NewFieldInvalid("URL", URL, err.Error()))
		// apiError(w, fmt.Sprintf("%s is not a valid URL", URL), http.StatusBadRequest)

	}

	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(u.Host)
		if err != nil {
			invalids = append(invalids, fielderrors.NewFieldInvalid("URL", URL, err.Error()))
		}
	}

	if !shared.AcceptedURL(u) {
		invalids = append(invalids, fielderrors.NewFieldValueNotSupported("URL", URL, nil))
	}

	if len(invalids) > 0 {
		apiutils.WriteRestError(w, apierrors.NewInvalid("create-suggestion", "URL", invalids))
		return
	}

	s := client.NewSuggestion(u.String())
	measurers, err := measure.DefaultMeasurements(form.URL)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}

	for _, v := range measurers {
		m, err := v.Measure()
		if err != nil {
			lg.Errorf("could not measure: %s", err.Error())
		} else {
			switch m.Type() {
			case sampletypes.DNSQuery, sampletypes.HTTPHeader:
				err = s.AddMeasurement(m)
				if err != nil {
					lg.Errorln(err.Error())
					return
				}
			default:
				lg.Warningf("unsupported sample type: %s", m.Type().String())
			}
		}
	}
}

func GetSuggestion(w rest.ResponseWriter, r *rest.Request) {
	ID := r.PathParam("id")
	form, ok := client.GetSuggestion(ID)
	if !ok {
		apiutils.WriteRestError(w, apierrors.NewNotFound("suggestion", ID))
		return
	}
	w.WriteJson(form)
}

// GetSuggestions returns all local/remote suggestion sessions, can be filtered
// by time of creation.
func GetSuggestions(w rest.ResponseWriter, r *rest.Request) {
	var res []client.Suggestion
	all := client.AllSuggestions()
	after := r.URL.Query().Get("after")

	if after == "" {
		res = all
	} else {
		after, err := time.Parse(time.RFC3339, after)
		if err != nil {
			lg.Warningln("Wrong time format: %s", after)
			return
		}
		for _, v := range all {
			if v.CreatedAt.After(after) {
				res = append(res, v)
			}
		}
	}
	w.WriteJson(res)
}

// SubmitSuggestion initiates the comminication with Central for a Submission
// session.
func SubmitSuggestion(w rest.ResponseWriter, r *rest.Request) {

	// // TODO This is the response that must be forwarded from central/api and parsed by client and passed on to the browser.
	// apiutils.WriteRestError(w,
	// 	apierrors.NewInvalid("object", "suggestion",
	// 		fielderrors.ValidationErrorList{
	// 			fielderrors.NewFieldValueNotSupported("URL", "...", []string{})}))
	// return

	ID := r.PathParam("id")
	suggestion, ok := client.GetSuggestion(ID)
	if !ok {
		apiutils.WriteRestError(w, apierrors.NewNotFound("suggestion", ID))
		return
	}
	wanip := shared.GetPublicIPAddr()
	if wanip == nil {
		lg.Warning("could not resolve public ip addr")
	}

	conf := clientconfig.Get()
	restclient, err := NewRestClient()
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}

	tokenResp, err := suggestion.RequestToken(
		restclient, wanip, conf.Settings.Local.CountryCode)
	if err != nil {
		if apiutils.IsNetError(err) {
			apiutils.WriteRestError(w, apierrors.NewServerTimeout("alkasir-central", "request-submission-token", 0))
		} else {
			apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		}
		return
	}
	n, err := suggestion.SendSamples(restclient)
	if err != nil {
		lg.Warningln("error sending samples", err.Error())
	}

	lg.V(5).Infoln("sent ", n)
	// FIXME PRESENTATION: just add the url locally
	u, err := url.Parse(suggestion.URL)
	if err != nil {
		lg.Errorln(err)
	} else {
		err := clientconfig.Update(func(conf *clientconfig.Config) error {
			conf.BlockedHosts.Add(u.Host)
			lastBlocklistChange = time.Now()

			pac.UpdateBlockedList(conf.BlockedHostsCentral.Hosts,
				conf.BlockedHosts.Hosts)
			return nil
		})
		if err != nil {
			lg.Errorln(err)
		}
	}
	w.WriteJson(tokenResp)
}

var (
	transportTraffic    shared.TransportTraffic
	transportTrafficLog []shared.TransportTraffic
	transportTrafficMu  sync.RWMutex
)

func PostTransportTraffic(w rest.ResponseWriter, r *rest.Request) {
	form := shared.TransportTraffic{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}
	transportTrafficMu.Lock()
	defer transportTrafficMu.Unlock()
	transportTraffic = form
	if lg.V(10) {
		if len(transportTrafficLog) == 6 {
			lg.Infof("transport traffic: %.0fkb/s %.0fkb/s %.0fkb/s %.0fkb/s  %.0fkb/s %.0fkb/s",
				(transportTrafficLog[0].Throughput)/1024,
				(transportTrafficLog[1].Throughput)/1024,
				(transportTrafficLog[2].Throughput)/1024,
				(transportTrafficLog[3].Throughput)/1024,
				(transportTrafficLog[4].Throughput)/1024,
				(transportTrafficLog[5].Throughput)/1024,
			)
			transportTrafficLog = make([]shared.TransportTraffic, 0)
		}
		if transportTraffic.Throughput > 1024 {
			transportTrafficLog = append(transportTrafficLog, form)
		}
	}
	response := true
	w.WriteJson(response)
}

// GetTransportTraffic returns all current traffic entries and empties the queue
func GetTransportTraffic(w rest.ResponseWriter, r *rest.Request) {
	transportTrafficMu.RLock()
	defer transportTrafficMu.RUnlock()
	w.WriteJson(&transportTraffic)
	return
}

// adds api routes to given mix router
func AddRoutes(mux *http.ServeMux) error {
	api := rest.NewApi()

	logger := log.New(nil, "", 0)
	lg.CopyLoggerTo("INFO", logger)

	loggerWarning := log.New(nil, "", 0)
	lg.CopyLoggerTo("WARNING", loggerWarning)

	if lg.V(100) {
		api.Use(&middlewares.AccessLogApacheMiddleware{
			Format: "%S\033[0m \033[36;1m%Dμs\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m",
		})
	} else if lg.V(20) {
		api.Use(&middlewares.AccessLogApacheMiddleware{
			Format: "%s %Dμs %r",
		})
	} else if lg.V(3) {
		api.Use(&middlewares.AccessLogApacheErrorMiddleware{
			&middlewares.AccessLogApacheMiddleware{
				Format: "%s %Dμs %r",
			},
		})
	}
	api.Use(
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		// &rest.ContentTypeCheckerMiddleware{},
	)
	if lg.V(5) {
		api.Use(
			&rest.RecoverMiddleware{
				EnableResponseStackTrace: true,
				Logger: logger,
			},
		)
	} else {
		api.Use(
			&rest.RecoverMiddleware{
				Logger: logger,
			},
		)
	}

	router, err := rest.MakeRouter(routes...)
	if err != nil {
		panic(err)
	}
	api.SetApp(router)
	handler := api.MakeHandler()
	mux.Handle("/api/", http.StripPrefix("/api", handler))
	return err
}

// A link to another entity.
type Link struct {
	ID   string // The Id of whatever is being linked
	Name string // Usually the name of the linked object, might also be some kind of title.
}

type MethodListItem struct {
	ID       string
	Service  Link
	Name     string
	Protocol string
	BindAddr string
	Running  bool
}

func toMethodListItem(m service.Method) MethodListItem {
	return MethodListItem{
		ID:       m.ID,
		Name:     m.Name,
		Service:  Link{m.Service.ID, m.Service.Name},
		Protocol: m.Protocol,
		BindAddr: m.BindAddr,
		Running:  m.Service.Running(),
	}
}

// api integration
func GetAllMethods(w rest.ResponseWriter, r *rest.Request) {
	methods := service.ManagedServices.AllMethods()
	items := make([]MethodListItem, 0)
	for _, m := range methods {
		items = append(items, toMethodListItem(*m))
	}

	w.WriteJson(items)
	return
}

type ServiceListItem struct {
	ID      string
	Name    string
	Running bool
	Methods []Link
}

func toServiceListItem(s service.Service) ServiceListItem {
	methods := make([]Link, 0)
	for _, method := range s.Methods.All() {
		if method != nil {
			methods = append(methods, Link{method.ID, method.Name})
		}
	}
	return ServiceListItem{
		ID:      s.ID,
		Name:    s.Name,
		Running: s.Running(),
		Methods: methods,
	}
}

// JSON api function
func GetAllServices(w rest.ResponseWriter, r *rest.Request) {
	services := service.ManagedServices.AllServices()
	items := make([]ServiceListItem, 0)
	for _, value := range services {
		items = append(items, toServiceListItem(*value))
	}
	w.WriteJson(items)
	return
}

// HostPatternListItem
type HostPatternListItem struct {
	Pattern    string   // The url pattern itself
	Categories []string // Which lists this pattern belongs to (blocked/direct/...)
}

func toHostPatternListItem(pattern string, cateogires ...string) HostPatternListItem {
	return HostPatternListItem{
		Pattern:    pattern,
		Categories: cateogires,
	}
}

// Fetch and list all host patterns
func GetAllHostPatterns(w rest.ResponseWriter, r *rest.Request) {
	conf := clientconfig.Get()

	items := make([]HostPatternListItem, 0)

	for _, value := range conf.BlockedHosts.Hosts {
		items = append(items, toHostPatternListItem(value, "blocked"))
	}

	for _, value := range conf.DirectHosts.Hosts {
		items = append(items, toHostPatternListItem(value, "direct"))
	}

	for _, value := range conf.BlockedHostsCentral.Hosts {
		items = append(items, toHostPatternListItem(value, "blocked-central"))
	}

	w.WriteJson(items)
	return
}

func getHostPatternLists() map[string]clientconfig.HostsFile {
	conf := clientconfig.Get()
	lists := make(map[string]clientconfig.HostsFile)
	lists["blocked"] = conf.BlockedHosts
	lists["direct"] = conf.DirectHosts
	lists["blocked-central"] = conf.BlockedHostsCentral
	return lists
}

// Response from a transport connection test
type MethodTestResult struct {
	MethodID string
	// TestURL  string
	Ok      bool
	Message string
}

func testSocks5Internet(addr string) (err error) {
	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, addr)
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial:                  dialSocksProxy,
			DisableKeepAlives:     true,
			ResponseHeaderTimeout: time.Duration(time.Second * 5),
		},
	}
	resp, err := httpClient.Get("http://ip.appspot.com")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

// TestMethod tests a transport method
func TestMethod(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	method := service.ManagedServices.Method(id)
	if method == nil {
		rest.NotFound(w, r)
		return
	}
	err := testSocks5Internet(method.BindAddr)
	result := &MethodTestResult{
		MethodID: id,
		Ok:       err == nil,
	}
	if result.Ok {
		result.Message = "Test successful conneting via " +
			method.Protocol + " / " +
			method.BindAddr + "."
	} else {
		result.Message = "Test FAILED conneting via " +
			method.Protocol + " / " +
			method.BindAddr + "."
	}
	w.WriteJson(result)
}

type PostLogRequest struct {
	Level   lg.Level `json:"level"`
	Context string   `json:"context"`
	Message string   `json:"message"`
}

// PostLogRequest
func PostLog(w rest.ResponseWriter, r *rest.Request) {
	form := PostLogRequest{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		lg.Warning(err)
		apiutils.ErrToAPIStatus(err)
		// apiutils.WriteRestError(w, apierrors.NewBadRequest("can't decode json"))
		// apiutils.WriteRestError(w, apierrors.NewBadRequest("can't decode json"))
		return
	}
	sender := r.PathParam("sender")
	lg.V(form.Level).Infof("{%s} %s: %s", sender, form.Context, form.Message)
	w.WriteJson(true)
}

type ValidateConnectionStringRequest struct {
	ConnectionString string `json:"connectionString"`
}

type ValidateConnectionStringResponse struct {
	Ok   bool   `json:"ok"`
	Name string `json:"name"`
}

func ValidateConnectionString(w rest.ResponseWriter, r *rest.Request) {
	form := ValidateConnectionStringRequest{}
	err := r.DecodeJsonPayload(&form)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewBadRequest("can't decode json"))
		return
	}

	c, err := shared.DecodeConnection(form.ConnectionString)

	if err != nil {
		w.WriteJson(ValidateConnectionStringResponse{
			Ok: false,
		})
		return
	}

	w.WriteJson(ValidateConnectionStringResponse{
		Ok:   true,
		Name: c.DisplayName(),
	})
}

func writeStatusSuccess(w rest.ResponseWriter) {
	result := &shared.Status{
		Status: shared.StatusSuccess,
		Code:   http.StatusOK,
		// Details: &shared.StatusDetails{
		// Name: name,
		// Kind: scope.Kind,
		// },
	}
	w.WriteHeader(result.Code)
	w.WriteJson(&result)
}

func GetDebug(w rest.ResponseWriter, r *rest.Request) {
	response := debugexport.NewDebugResposne(VERSION, clientconfig.Get())
	if r.URL.Query().Get("inbrowser") != "true" {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition",
			fmt.Sprintf("inline; filename=alkasir-debuginfo\"%s.txt\"", response.Header.ID))
	}

	_ = w.WriteJson(response)
}

func PostBrowsercodeToClipboard(w rest.ResponseWriter, r *rest.Request) {
	conf := clientconfig.Get()
	ak := conf.Settings.Local.ClientAuthKey
	addr := conf.Settings.Local.ClientBindAddr

	bc := browsercode.BrowserCode{Key: ak}
	err := bc.SetHostport(addr)
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}

	err = bc.CopyToClipboard()
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}

	w.WriteJson(true)
}

func PostExportChromeExtension(w rest.ResponseWriter, r *rest.Request) {

	err := saveChromeExtension()
	if err != nil {
		apiutils.WriteRestError(w, apierrors.NewInternalError(err))
		return
	}

	w.WriteJson(true)
}
