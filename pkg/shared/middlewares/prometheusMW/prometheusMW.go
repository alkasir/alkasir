package prometheusMW

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

//PrometheusMiddleware .
type PrometheusMiddleware struct {
	ServiceName string
}

// MiddlewareFunc makes PrometheusMiddleware implement the rest.Middleware interface.
func (mw *PrometheusMiddleware) MiddlewareFunc(handler rest.HandlerFunc) rest.HandlerFunc {

	requests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_total", mw.ServiceName),
		Help: "Total api requests",
	})
	prometheus.MustRegister(requests)

	status500 := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_500", mw.ServiceName),
		Help: "Total api status 500 responses",
	})
	prometheus.MustRegister(status500)

	status5xx := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_5xx", mw.ServiceName),
		Help: "Total api status 5xx responses",
	})
	prometheus.MustRegister(status5xx)

	status400 := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_400", mw.ServiceName),
		Help: "Total api status 400 responses",
	})

	prometheus.MustRegister(status400)

	status404 := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_404", mw.ServiceName),
		Help: "Total api status 404 responses",
	})

	prometheus.MustRegister(status404)

	status4xx := prometheus.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_api_req_4xx", mw.ServiceName),
		Help: "Total api status 4xx responses",
	})

	prometheus.MustRegister(status4xx)

	return func(writer rest.ResponseWriter, request *rest.Request) {
		handler(writer, request)
		requests.Add(1)
		if request.Env["STATUS_CODE"] != nil {
			s := request.Env["STATUS_CODE"].(int)
			switch {
			case s == 500:
				status500.Inc()
			case s > 500:
				status5xx.Inc()
			case s == 400:
				status400.Inc()
			case s == 404:
				status404.Inc()
			case s > 400:
				status4xx.Inc()
			}
		}
	}
}
