package central

import (
	"github.com/alkasir/alkasir/pkg/shared/middlewares"
	"github.com/alkasir/alkasir/pkg/shared/middlewares/prometheusMW"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/thomasf/lg"
)

func defaultAPI(servername string) *rest.Api {
	api := rest.NewApi()
	api.Use(&prometheusMW.PrometheusMiddleware{ServiceName: servername})
	if lg.V(100) {
		api.Use(&middlewares.AccessLogApacheMiddleware{
			Format: "%S\033[0m \033[36;1m%Dμs\033[0m \"%r\" \033[1;30m%u \"%{User-Agent}i\"\033[0m",
		})
	} else {
		api.Use(&middlewares.AccessLogApacheMiddleware{
			Format: "%S\033[0m \033[36;1m%Dμs\033[0m \"%r\" \033[1;30m%u \033[0m",
		})
	}
	api.Use([]rest.Middleware{
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{},
		// &rest.GzipMiddleware{},
		// &rest.ContentTypeCheckerMiddleware{},
	}...)

	return api
}

// validCountryCode returns true if the supplied country code is handeled by
// the server.
func validCountryCode(cc string) bool {
	if cc == "__" {
		return false
	}
	validCountryCodesMu.RLock()
	_, ok := validCountryCodes[cc]
	validCountryCodesMu.RUnlock()
	return ok
}
