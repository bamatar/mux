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

var defaultNotFound handler = func(ctx *Context) error {
	// TODO: implement 404 handler
	return nil
}

var defaultMethodNotAllowed handler = func(ctx *Context) error {
	// TODO: implement 405 handler
	return nil
}

// ServeHTTP handles the request lifecycle
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{w: w, r: r}
	if err := h(ctx); err != nil {
		// TODO: handle error
	}
}
