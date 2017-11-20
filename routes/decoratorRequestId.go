package routes

import (
	"net/http"

	"github.com/satori/go.uuid"
)

// RequestId is a decorator which adds a request ID to incoming requests.  It
// stores it in the request context, sets the logger up to log it, and adds it
// to the response in the `X-Request-ID` HTTP header.
type RequestId struct{}

func (_ RequestId) Decorate(h Handler) Handler {
	h.Func = getRequestIdFunc(h.Func)
	return h
}

func getRequestIdFunc(hf HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, a Arguments) (int, interface{}) {
		requestId := uuid.NewV1().String()
		r = requestWithLoggedContextValue(r, contextKeyRequestId, "requestId", requestId)
		w.Header()["X-Request-ID"] = []string{requestId}
		return hf(w, r, a)
	}
}
