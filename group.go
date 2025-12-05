package mux

// Group wraps Router with nested paths
type Group struct {
	prefix string
	router *Router
}

// Group creates a router group
func (r *Router) Group(prefix string) *Group {
	return &Group{
		router: r,
		prefix: prefix,
	}
}

// Group creates a router group
func (g *Group) Group(prefix string) *Group {
	return &Group{
		router: g.router,
		prefix: g.prefix + prefix,
	}
}

// GET registers a handler for GET requests
func (g *Group) GET(pattern string, h Handler) {
	g.router.handle("GET", g.prefix+pattern, h)
}

// POST registers a handler for POST requests
func (g *Group) POST(pattern string, h Handler) {
	g.router.handle("POST", g.prefix+pattern, h)
}

// PUT registers a handler for PUT requests
func (g *Group) PUT(pattern string, h Handler) {
	g.router.handle("PUT", g.prefix+pattern, h)
}

// DELETE registers a handler for DELETE requests
func (g *Group) DELETE(pattern string, h Handler) {
	g.router.handle("DELETE", g.prefix+pattern, h)
}

// PATCH registers a handler for PATCH requests
func (g *Group) PATCH(pattern string, h Handler) {
	g.router.handle("PATCH", g.prefix+pattern, h)
}
