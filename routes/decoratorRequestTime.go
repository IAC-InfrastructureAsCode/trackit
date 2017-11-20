package routes

import (
	"net/http"
	"time"
)

// RequestTime is a decorator which adds a request time to incoming requests.
// It stores it in the request context and sets the logger up to log it.
type RequestTime struct{}

func (d RequestTime) Decorate(h Handler) Handler {
	h.Func = getRequestTimeFunc(h.Func)
	return h
}

func getRequestTimeFunc(hf HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, a Arguments) (int, interface{}) {
		now := time.Now()
		r = requestWithLoggedContextValue(r, contextKeyRequestTime, "requestTime", now)
		return hf(w, r, a)
	}
}
