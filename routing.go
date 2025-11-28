package mux

import (
	"net/http"
	"strings"
)

// allowed maps paths to their registered methods
type allowed map[string][]string

// Router wraps http.ServeMux with error handling
type Router struct {
	// mux routes with method prefix
	mux *http.ServeMux

	// bare routes without method for 405 detection
	bare *http.ServeMux

	// root is true for root router, false for groups
	root bool

	// prefix is the path prefix for all routes
	prefix string

	// allowed maps patterns to their registered methods
	allowed allowed

	// notFound handles 404 responses
	notFound handler

	// methodNotAllowed handles 405 responses
	methodNotAllowed handler
}

// New creates a router
func New() *Router {
	return &Router{
		mux:              http.NewServeMux(),
		bare:             http.NewServeMux(),
		root:             true,
		allowed:          make(allowed),
		notFound:         defaultNotFound,
		methodNotAllowed: defaultMethodNotAllowed,
	}
}

// Group creates a router with a prefix
func (r *Router) Group(prefix string) *Router {
	return &Router{
		mux:              r.mux,
		bare:             r.bare,
		prefix:           r.prefix + prefix,
		allowed:          r.allowed,
		notFound:         r.notFound,
		methodNotAllowed: r.methodNotAllowed,
	}
}

// SetNotFound sets the handler for 404 responses
func (r *Router) SetNotFound(h Handler) {
	r.notFound = h.Handle
}

// SetMethodNotAllowed sets the handler for 405 responses
func (r *Router) SetMethodNotAllowed(h Handler) {
	r.methodNotAllowed = h.Handle
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
func (r *Router) handle(method, path string, h Handler) {
	pattern := r.prefix + path
	r.mux.Handle(method+" "+pattern, handler(h.Handle))

	// Register to bare once per pattern for 405 detection
	if r.allowed[pattern] == nil {
		r.bare.Handle(pattern, http.NotFoundHandler())
	}

	r.allowed[pattern] = append(r.allowed[pattern], method)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !r.root {
		panic("mux: cannot serve from group router")
	}

	h, pattern := r.mux.Handler(req)
	if pattern != "" {
		h.ServeHTTP(w, req)
		return
	}

	// 405
	_, pattern = r.bare.Handler(req)
	if methods, ok := r.allowed[pattern]; ok {
		w.Header().Set("Allow", strings.Join(methods, ", "))
		r.methodNotAllowed.ServeHTTP(w, req)
		return
	}

	// 404
	r.notFound.ServeHTTP(w, req)
}
