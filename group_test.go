package mux

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// -----------------------------------------------------------------------------
// Group Creation
// -----------------------------------------------------------------------------

func TestGroupFromRouter(t *testing.T) {
	r := New()
	g := r.Group("/api")

	if g.prefix != "/api" {
		t.Errorf("expected prefix /api, got %s", g.prefix)
	}
	if g.router != r {
		t.Error("expected group to reference router")
	}
}

func TestGroupNested(t *testing.T) {
	r := New()
	api := r.Group("/api")
	v1 := api.Group("/v1")

	if v1.prefix != "/api/v1" {
		t.Errorf("expected prefix /api/v1, got %s", v1.prefix)
	}
	if v1.router != r {
		t.Error("expected nested group to reference router")
	}
}

func TestGroupDeeplyNested(t *testing.T) {
	r := New()
	l1 := r.Group("/l1")
	l2 := l1.Group("/l2")
	l3 := l2.Group("/l3")
	l4 := l3.Group("/l4")

	if l4.prefix != "/l1/l2/l3/l4" {
		t.Errorf("expected prefix /l1/l2/l3/l4, got %s", l4.prefix)
	}
}

// -----------------------------------------------------------------------------
// Group Middleware
// -----------------------------------------------------------------------------

func TestGroupUse(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var called bool
	mw := func(next Handler) Handler {
		return func(c *Context) error {
			called = true
			return next(c)
		}
	}

	api.Use(mw)
	api.GET("/test", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !called {
		t.Error("middleware was not called")
	}
}

func TestGroupUseMultiple(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var order []int
	mw1 := func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, 1)
			return next(c)
		}
	}
	mw2 := func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, 2)
			return next(c)
		}
	}

	api.Use(mw1, mw2)
	api.GET("/test", func(c *Context) error {
		order = append(order, 3)
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

func TestGroupMiddlewareInheritance(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var parentCalled, childCalled bool
	parentMw := func(next Handler) Handler {
		return func(c *Context) error {
			parentCalled = true
			return next(c)
		}
	}
	childMw := func(next Handler) Handler {
		return func(c *Context) error {
			childCalled = true
			return next(c)
		}
	}

	api.Use(parentMw)
	v1 := api.Group("/v1")
	v1.Use(childMw)

	v1.GET("/test", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !parentCalled {
		t.Error("parent middleware was not called")
	}
	if !childCalled {
		t.Error("child middleware was not called")
	}
}

func TestGroupMiddlewareOrder(t *testing.T) {
	r := New()

	var order []string

	routerMw := func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "router")
			return next(c)
		}
	}
	groupMw := func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "group")
			return next(c)
		}
	}
	nestedMw := func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "nested")
			return next(c)
		}
	}

	r.Use(routerMw)
	api := r.Group("/api")
	api.Use(groupMw)
	v1 := api.Group("/v1")
	v1.Use(nestedMw)

	v1.GET("/test", func(c *Context) error {
		order = append(order, "handler")
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	expected := []string{"router", "group", "nested", "handler"}
	if len(order) != len(expected) {
		t.Errorf("expected %v, got %v", expected, order)
		return
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected %v, got %v", expected, order)
			break
		}
	}
}

func TestGroupMiddlewareIsolation(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var g1Called, g2Called bool
	mw1 := func(next Handler) Handler {
		return func(c *Context) error {
			g1Called = true
			return next(c)
		}
	}
	mw2 := func(next Handler) Handler {
		return func(c *Context) error {
			g2Called = true
			return next(c)
		}
	}

	g1 := api.Group("/g1")
	g1.Use(mw1)
	g1.GET("/test", func(c *Context) error { return c.OK(nil) })

	g2 := api.Group("/g2")
	g2.Use(mw2)
	g2.GET("/test", func(c *Context) error { return c.OK(nil) })

	// Request to g1 should only call mw1
	req := httptest.NewRequest("GET", "/api/g1/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !g1Called {
		t.Error("g1 middleware should be called")
	}
	if g2Called {
		t.Error("g2 middleware should not be called")
	}
}

// -----------------------------------------------------------------------------
// Group HTTP Methods
// -----------------------------------------------------------------------------

func TestGroupAllMethods(t *testing.T) {
	r := New()
	api := r.Group("/api")

	api.GET("/resource", func(c *Context) error { return c.OK(M{"m": "GET"}) })
	api.POST("/resource", func(c *Context) error { return c.OK(M{"m": "POST"}) })
	api.PUT("/resource", func(c *Context) error { return c.OK(M{"m": "PUT"}) })
	api.DELETE("/resource", func(c *Context) error { return c.OK(M{"m": "DELETE"}) })
	api.PATCH("/resource", func(c *Context) error { return c.OK(M{"m": "PATCH"}) })

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/resource", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != 200 {
				t.Errorf("expected 200, got %d", rec.Code)
			}

			var body M
			json.Unmarshal(rec.Body.Bytes(), &body)
			if body["m"] != method {
				t.Errorf("expected %s, got %v", method, body["m"])
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Group with Path Parameters
// -----------------------------------------------------------------------------

func TestGroupWithPathParams(t *testing.T) {
	r := New()
	api := r.Group("/api")
	users := api.Group("/users")

	users.GET("/{id}", func(c *Context) error {
		return c.OK(M{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/api/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["id"] != "123" {
		t.Errorf("expected id=123, got %v", body["id"])
	}
}

// -----------------------------------------------------------------------------
// Group Middleware with Context
// -----------------------------------------------------------------------------

func TestGroupMiddlewareSetContext(t *testing.T) {
	r := New()
	api := r.Group("/api")

	authMw := func(next Handler) Handler {
		return func(c *Context) error {
			c.Set("user_id", "user123")
			return next(c)
		}
	}

	api.Use(authMw)
	api.GET("/me", func(c *Context) error {
		return c.OK(M{"user_id": c.GetString("user_id")})
	})

	req := httptest.NewRequest("GET", "/api/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["user_id"] != "user123" {
		t.Errorf("expected user123, got %v", body["user_id"])
	}
}

func TestGroupMiddlewareShortCircuit(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var handlerCalled bool
	authMw := func(next Handler) Handler {
		return func(c *Context) error {
			return c.Unauthorized(M{"error": "unauthorized"})
		}
	}

	api.Use(authMw)
	api.GET("/protected", func(c *Context) error {
		handlerCalled = true
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/api/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if handlerCalled {
		t.Error("handler should not be called when middleware short-circuits")
	}
}

// -----------------------------------------------------------------------------
// Edge Cases
// -----------------------------------------------------------------------------

func TestGroupEmptyPrefix(t *testing.T) {
	r := New()
	g := r.Group("")
	g.GET("/test", func(c *Context) error {
		return c.OK(M{"path": c.Path()})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGroupMultipleSamePrefix(t *testing.T) {
	r := New()

	g1 := r.Group("/api")
	g2 := r.Group("/api")

	g1.GET("/one", func(c *Context) error { return c.OK(M{"g": 1}) })
	g2.GET("/two", func(c *Context) error { return c.OK(M{"g": 2}) })

	req1 := httptest.NewRequest("GET", "/api/one", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest("GET", "/api/two", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec1.Code != 200 || rec2.Code != 200 {
		t.Errorf("expected both 200, got %d and %d", rec1.Code, rec2.Code)
	}
}

func TestGroupDoesNotInheritRouterMiddleware(t *testing.T) {
	r := New()

	var routerMwCalled bool
	routerMw := func(next Handler) Handler {
		return func(c *Context) error {
			routerMwCalled = true
			return next(c)
		}
	}

	// Add middleware to router AFTER creating group
	api := r.Group("/api")
	r.Use(routerMw)

	api.GET("/test", func(c *Context) error {
		return c.OK(nil)
	})

	// Router middleware should still be called (applied at handle time)
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !routerMwCalled {
		t.Error("router middleware should be called")
	}
}

func TestNestedGroupDoesNotAffectParent(t *testing.T) {
	r := New()
	api := r.Group("/api")

	var nestedMwCalled bool
	nestedMw := func(next Handler) Handler {
		return func(c *Context) error {
			nestedMwCalled = true
			return next(c)
		}
	}

	// Add handler to parent before creating nested group
	api.GET("/parent", func(c *Context) error { return c.OK(nil) })

	// Create nested group with middleware
	v1 := api.Group("/v1")
	v1.Use(nestedMw)
	v1.GET("/nested", func(c *Context) error { return c.OK(nil) })

	// Request to parent should not trigger nested middleware
	req := httptest.NewRequest("GET", "/api/parent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if nestedMwCalled {
		t.Error("nested middleware should not be called for parent route")
	}
}
