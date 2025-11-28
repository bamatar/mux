package mux

// Handler is the interface for handling HTTP requests.
type Handler interface {
	Handle(ctx *Context) error
}

// HandlerFunc allows using functions as handlers.
type HandlerFunc func(ctx *Context) error

func (f HandlerFunc) Handle(ctx *Context) error {
	return f(ctx)
}

// ErrorHandler processes errors returned from handlers.
type ErrorHandler func(ctx *Context, err error)
