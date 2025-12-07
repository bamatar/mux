package mux

import (
	"fmt"
	"log"
	"net/http"
)

// Handler handles HTTP requests
type Handler func(c *Context) error

// Middleware wraps handler execution

type Middleware func(next Handler) Handler

// ErrorHandler processes handler errors
type ErrorHandler func(c *Context, err error)

// Router wraps http.ServeMux with error handling
type Router struct {
	ctx pool[Context]
	mux *http.ServeMux
	mws []Middleware

	// callbacks
	on404 http.Handler
	on405 http.Handler
	onErr ErrorHandler
}

// New creates a router
func New() *Router {

	r := new(Router)
	r.ctx = pool[Context]{}
	r.mux = http.NewServeMux()

	r.On404(func(c *Context) error {
		return c.NotFound(M{"error": "not found"})
	})

	r.On405(func(c *Context) error {
		return c.MethodNotAllowed(M{"error": "method not allowed"})
	})

	r.OnErr(func(c *Context, err error) {
		_ = c.InternalServerError(M{"error": "internal server error", "message": err.Error()})
	})

	return r
}

// On404 sets the handler for 404 responses
func (r *Router) On404(h Handler) {
	r.on404 = r.handler(h)
}

// On405 sets the handler for 405 responses
func (r *Router) On405(h Handler) {
	r.on405 = r.handler(h)
}

// OnErr sets the error handler
func (r *Router) OnErr(h ErrorHandler) {
	r.onErr = h
}

// Use adds middleware to the router
func (r *Router) Use(middlewares ...Middleware) {
	r.mws = append(r.mws, middlewares...)
}

// GET registers a handler for GET requests
func (r *Router) GET(pattern string, h Handler) {
	r.handle("GET", pattern, h)
}

// POST registers a handler for POST requests
func (r *Router) POST(pattern string, h Handler) {
	r.handle("POST", pattern, h)
}

// PUT registers a handler for PUT requests
func (r *Router) PUT(pattern string, h Handler) {
	r.handle("PUT", pattern, h)
}

// DELETE registers a handler for DELETE requests
func (r *Router) DELETE(pattern string, h Handler) {
	r.handle("DELETE", pattern, h)
}

// PATCH registers a handler for PATCH requests
func (r *Router) PATCH(pattern string, h Handler) {
	r.handle("PATCH", pattern, h)
}

// handle registers a handler for the method and path
func (r *Router) handle(method, pattern string, h Handler) {
	for i := len(r.mws) - 1; i >= 0; i-- {
		h = r.mws[i](h)
	}
	r.mux.Handle(method+" "+pattern, r.handler(h))
}

// safelyHandleError calls the error handler with panic recovery
func (r *Router) safelyHandleError(c *Context, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("mux: panic in error handler: %v", e)
		}
	}()
	r.onErr(c, err)
}

// handler wraps a Handler into http.HandlerFunc with context pooling and panic recovery
func (r *Router) handler(handlerFunc Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// acquire context
		c := r.ctx.get()
		c.attach(w, req)

		defer func() {
			if err := recover(); err != nil {
				r.safelyHandleError(c, fmt.Errorf("panic: %v", err))
			}

			// release context
			c.detach()
			r.ctx.put(c)
		}()

		// execute handler
		if err := handlerFunc(c); err != nil {
			r.safelyHandleError(c, err)
		}
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h, p := r.mux.Handler(req)

	if p != "" {
		// Route matched. Delegate to mux which sets path values
		// before calling the handler.
		r.mux.ServeHTTP(w, req)
		return
	}

	// No route matched. The mux returns an internal handler that would
	// write 404 or 405. used responder as a fake writer to capture the
	// status without sending anything to the client, then substitute
	// our custom error handler.
	rsp := responder{ResponseWriter: w}
	h.ServeHTTP(&rsp, req)

	if rsp.status == http.StatusMethodNotAllowed {
		r.on405.ServeHTTP(w, req)
		return
	}

	r.on404.ServeHTTP(w, req)
}
