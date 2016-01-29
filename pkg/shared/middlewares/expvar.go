package middlewares

import (
	"expvar"
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"
)

//ExpvarMiddleware .
type ExpvarMiddleware struct {
	ServiceName string
}

// MiddlewareFunc makes ExpvarMiddleware implement the rest.Middleware interface.
func (mw *ExpvarMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {
	requests := expvar.NewInt(fmt.Sprintf("%s-api_req_total", mw.ServiceName))
	status500 := expvar.NewInt(fmt.Sprintf("%s-api_req_500", mw.ServiceName))
	status5xx := expvar.NewInt(fmt.Sprintf("%s-api_req_5xx", mw.ServiceName))
	status400 := expvar.NewInt(fmt.Sprintf("%s-api_req_400", mw.ServiceName))
	status404 := expvar.NewInt(fmt.Sprintf("%s-api_req_404", mw.ServiceName))
	status4xx := expvar.NewInt(fmt.Sprintf("%s-api_req_4xx", mw.ServiceName))
	return func(writer rest.ResponseWriter, request *rest.Request) {
		handler(writer, request)
		requests.Add(1)
		if request.Env["STATUS_CODE"] != nil {
			s := request.Env["STATUS_CODE"].(int)
			switch {
			case s == 500:
				status500.Add(1)
			case s > 500:
				status5xx.Add(1)
			case s == 400:
				status400.Add(1)
			case s == 404:
				status404.Add(1)
			case s > 400:
				status4xx.Add(1)
			}
		}
	}
}
