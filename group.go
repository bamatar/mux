package mux

// Group wraps Router with nested paths
type Group struct {
	prefix string
	router *Router
	mws    []Middleware
}

// Group creates a router group
func (r *Router) Group(prefix string) *Group {
	return &Group{
		router: r,
		prefix: prefix,
	}
}

// Group creates a nested group
func (g *Group) Group(prefix string) *Group {
	return &Group{
		router: g.router,
		prefix: g.prefix + prefix,
		mws:    append([]Middleware{}, g.mws...),
	}
}

// Use adds middleware to the group
func (g *Group) Use(middlewares ...Middleware) {
	g.mws = append(g.mws, middlewares...)
}

// GET registers a handler for GET requests
func (g *Group) GET(pattern string, h Handler) {
	g.handle("GET", pattern, h)
}

// POST registers a handler for POST requests
func (g *Group) POST(pattern string, h Handler) {
	g.handle("POST", pattern, h)
}

// PUT registers a handler for PUT requests
func (g *Group) PUT(pattern string, h Handler) {
	g.handle("PUT", pattern, h)
}

// DELETE registers a handler for DELETE requests
func (g *Group) DELETE(pattern string, h Handler) {
	g.handle("DELETE", pattern, h)
}

// PATCH registers a handler for PATCH requests
func (g *Group) PATCH(pattern string, h Handler) {
	g.handle("PATCH", pattern, h)
}

func (g *Group) handle(method, pattern string, h Handler) {
	for i := len(g.mws) - 1; i >= 0; i-- {
		h = g.mws[i](h)
	}
	g.router.handle(method, g.prefix+pattern, h)
}
