package client

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/pkg/browser"
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/browsercode"
	"github.com/alkasir/alkasir/pkg/client/internal/config"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/res"
)

func startInternalHTTPServer(authKey string) error {

	lg.V(15).Infoln("starting internal api server")
	mux, err := createServeMux()
	if err != nil {
		lg.Error("failed to create router")
		return err
	}

	auth := Auth{Key: authKey, wrapped: mux}
	conf := clientconfig.Get()
	listener, err := net.Listen("tcp", conf.Settings.Local.ClientBindAddr)
	if err != nil {
		lg.Warning(err)
		listener, err = net.Listen("tcp", "127.0.0.1:")
	}
	if err != nil {
		lg.Warning(err)
		ui.Notify("Could not bind any local port (bootstrap)")
		lg.Errorln("Could not bind any local port (bootstrap)")
		return err
	}
	url := fmt.Sprintf("http://%s?authKey=%s", listener.Addr().String(), authKey)
	go func() {
		for {
			select {
			case <-ui.Actions.CopyBrowserCodeToClipboard:
				bc := browsercode.BrowserCode{Key: authKey}
				err := bc.SetHostport(listener.Addr().String())
				if err != nil {
					lg.Errorln(err)
					continue
				}
				err = bc.CopyToClipboard()
				if err != nil {
					lg.Errorln(err)
				}

			case <-ui.Actions.OpenInBrowser:
				browser.OpenURL(url + "#/")

			case <-ui.Actions.Help:
				browser.OpenURL(url + "#/docs/__/index")
			}
		}
	}()

	doneC := make(chan bool, 1)
	go func() {
		defer listener.Close()
		err = http.Serve(listener, auth)
		if err != nil {
			doneC <- false
		}
	}()
	select {
	case ok := <-doneC:
		if !ok {
			return errors.New("Could not start internal http server")
		}
	case <-time.After(time.Duration(time.Millisecond * 200)):
		return nil
	}
	return nil
}

func createServeMux() (*http.ServeMux, error) {
	mux := http.NewServeMux()
	if hotEnabled {
		for _, v := range []string{"img", "generated", "fonts"} {
			mux.Handle(fmt.Sprintf("/%s/", v),
				http.StripPrefix(
					fmt.Sprintf("/%s/", v),
					http.FileServer(http.Dir(fmt.Sprintf("res/%s", v)))))
		}
	} else {
		fileserver := http.FileServer(&assetfs.AssetFS{Asset: res.Asset, AssetDir: res.AssetDir})
		mux.Handle("/img/", fileserver)
		mux.Handle("/generated/", fileserver)
		mux.Handle("/fonts/", fileserver)
	}
	mux.HandleFunc("/", HtmlHandler)
	err := AddRoutes(mux)
	if err != nil {
		return nil, err
	}
	return mux, nil
}

// HtmlHandler is a HandlerFunc that serves all pages in the internal browser
// using a single html template.
func HtmlHandler(w http.ResponseWriter, r *http.Request) {
	templates := loadTemplates()
	err := templates.ExecuteTemplate(w, "page.html", struct {
		Title    string
		PageData interface{}
	}{
		Title: "Alkasir",
	},
	)
	if err != nil {
		lg.Infof("err: %+v", err)
	}
}

// load all html templates from bindata.
func loadTemplates() (t *template.Template) {
	lg.V(30).Infoln("Loading templates...")
	t = template.New("_asdf")
	allAssets := res.AssetNames()

	for _, path := range allAssets {
		if strings.HasPrefix(path, "templates/") &&
			strings.HasSuffix(path, "html") {
			data, err := res.Asset(path)
			if err != nil {
				panic(err)
			}
			_, err = t.New(strings.TrimPrefix(path, "templates/")).Parse(string(data))
			if err != nil {
				panic(err)
			}
		}
	}
	return
}

// Auth .
type Auth struct {
	Key     string
	wrapped http.Handler
}

func (a Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authenticated := false

	if !authenticated {
		key := r.URL.Query().Get("authKey")
		if key == a.Key {
			authenticated = true
			http.SetCookie(w, &http.Cookie{Name: "authKey", Value: key})

		}
	}

	if !authenticated {
		ah := r.Header.Get("Authorization")
		if strings.HasPrefix(ah, "Bearer ") {
			key := strings.TrimPrefix(ah, "Bearer ")
			if key == a.Key {
				authenticated = true
			}
		}
	}

	if !authenticated {
		keyc, err := r.Cookie("authKey")
		if err == nil {
			if keyc.Value == a.Key {
				authenticated = true
			}
		}
	}

	if authenticated {
		a.wrapped.ServeHTTP(w, r)
		return
	}
	lg.V(50).Infoln("unauthenticated call to", r.URL.String())
	w.WriteHeader(401)

}
