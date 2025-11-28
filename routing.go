package mux

import "net/http"

// Router wraps http.ServeMux with error handling
type Router struct {
	mux *http.ServeMux
}

// New creates a router
func New() *Router {
	return &Router{mux: http.NewServeMux()}
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

// handle registers a handler for the method and pattern
func (r *Router) handle(method, pattern string, h Handler) {
	r.mux.Handle(method+" "+pattern, handler(h.Handle))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
