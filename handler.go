package mux

import (
	"net/http"
)

// Handler handles HTTP requests
type Handler interface {
	Handle(ctx *Context) error
}

// handler adapts Handler to http.Handler
type handler func(ctx *Context) error

// HandlerFunc allows using functions as handlers
type HandlerFunc func(ctx *Context) error

func (f HandlerFunc) Handle(ctx *Context) error {
	return f(ctx)
}

// ErrorHandler processes handler errors
type ErrorHandler func(ctx *Context, err error)

// ServeHTTP handles the request lifecycle
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{}
	if err := h(ctx); err != nil {
		// TODO: handle error
	}
}
