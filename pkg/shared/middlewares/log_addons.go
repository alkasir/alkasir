package middlewares

import (
	"github.com/ant0ine/go-json-rest/rest"
)

type AccessLogApacheErrorMiddleware struct {
	*AccessLogApacheMiddleware
}

func noopHandlerFunc(rest.ResponseWriter, *rest.Request) { return }

// MiddlewareFunc only logs http response codes > 200
func (mw *AccessLogApacheErrorMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	lh := mw.AccessLogApacheMiddleware.MiddlewareFunc(noopHandlerFunc)

	return func(w rest.ResponseWriter, r *rest.Request) {
		h(w, r)
		code := r.Env["STATUS_CODE"]
		if code, ok := code.(int); ok {
			if code > 299 {
				lh(w, r)
			}
		}
	}
}
