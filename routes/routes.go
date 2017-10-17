package routes

import (
	"encoding/json"
	"net/http"
)

type Arguments map[interface{}]interface{}

type Handler func(*http.Request, Arguments) (int, interface{})

type IntermediateHandler func(http.ResponseWriter, *http.Request, Arguments) (int, interface{})

type Decorator interface {
	Decorate(IntermediateHandler) IntermediateHandler
}

type ErrorBody struct {
	Error string `json:"error"`
}

func Register(pattern string, handler Handler, decorators ...Decorator) error {
	stage := baseIntermediate(handler)
	l := len(decorators) - 1
	for i := range decorators {
		d := decorators[l-i]
		stage = d.Decorate(stage)
	}
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		arguments := make(Arguments)
		status, output := stage(w, r, arguments)
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(output)
	})
	return nil
}

func baseIntermediate(handler Handler) IntermediateHandler {
	return func(w http.ResponseWriter, r *http.Request, a Arguments) (int, interface{}) {
		return handler(r, a)
	}
}
