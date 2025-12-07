package mux

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

// -----------------------------------------------------------------------------
// Router Creation
// -----------------------------------------------------------------------------

func TestNew(t *testing.T) {
	r := New()

	if r == nil {
		t.Fatal("expected router, got nil")
	}
	if r.mux == nil {
		t.Error("expected mux to be initialized")
	}
	if r.on404 == nil {
		t.Error("expected on404 handler")
	}
	if r.on405 == nil {
		t.Error("expected on405 handler")
	}
	if r.onErr == nil {
		t.Error("expected onErr handler")
	}
}

// -----------------------------------------------------------------------------
// HTTP Methods
// -----------------------------------------------------------------------------

func TestRouterGET(t *testing.T) {
	r := New()
	r.GET("/test", func(c *Context) error {
		return c.OK(M{"method": "GET"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRouterPOST(t *testing.T) {
	r := New()
	r.POST("/test", func(c *Context) error {
		return c.Created(M{"method": "POST"})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestRouterPUT(t *testing.T) {
	r := New()
	r.PUT("/test", func(c *Context) error {
		return c.OK(M{"method": "PUT"})
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRouterDELETE(t *testing.T) {
	r := New()
	r.DELETE("/test", func(c *Context) error {
		return c.NoContent()
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 204 {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestRouterPATCH(t *testing.T) {
	r := New()
	r.PATCH("/test", func(c *Context) error {
		return c.OK(M{"method": "PATCH"})
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Path Parameters
// -----------------------------------------------------------------------------

func TestRouterPathParams(t *testing.T) {
	r := New()
	r.GET("/users/{id}", func(c *Context) error {
		return c.OK(M{"id": c.Param("id")})
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)

	if body["id"] != "123" {
		t.Errorf("expected id=123, got %v", body["id"])
	}
}

func TestRouterMultiplePathParams(t *testing.T) {
	r := New()
	r.GET("/users/{userID}/posts/{postID}", func(c *Context) error {
		return c.OK(M{
			"userID": c.Param("userID"),
			"postID": c.Param("postID"),
		})
	})

	req := httptest.NewRequest("GET", "/users/10/posts/20", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)

	if body["userID"] != "10" {
		t.Errorf("expected userID=10, got %v", body["userID"])
	}
	if body["postID"] != "20" {
		t.Errorf("expected postID=20, got %v", body["postID"])
	}
}

func TestRouterWildcard(t *testing.T) {
	r := New()
	r.GET("/files/{path...}", func(c *Context) error {
		return c.OK(M{"path": c.Param("path")})
	})

	req := httptest.NewRequest("GET", "/files/a/b/c", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)

	if body["path"] != "a/b/c" {
		t.Errorf("expected path=a/b/c, got %v", body["path"])
	}
}

// -----------------------------------------------------------------------------
// 404 Handler
// -----------------------------------------------------------------------------

func TestDefault404(t *testing.T) {
	r := New()

	req := httptest.NewRequest("GET", "/notfound", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "not found" {
		t.Errorf("expected 'not found', got %v", body["error"])
	}
}

func TestCustom404(t *testing.T) {
	r := New()
	r.On404(func(c *Context) error {
		return c.JSON(404, M{"custom": "not found", "path": c.Path()})
	})

	req := httptest.NewRequest("GET", "/missing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["custom"] != "not found" {
		t.Errorf("expected custom message, got %v", body)
	}
}

// -----------------------------------------------------------------------------
// 405 Handler
// -----------------------------------------------------------------------------

func TestDefault405(t *testing.T) {
	r := New()
	r.GET("/resource", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/resource", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %v", body["error"])
	}
}

func TestCustom405(t *testing.T) {
	r := New()
	r.On405(func(c *Context) error {
		return c.JSON(405, M{"custom": "wrong method", "method": c.Method()})
	})
	r.GET("/resource", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("DELETE", "/resource", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["custom"] != "wrong method" {
		t.Errorf("expected custom message, got %v", body)
	}
}

func Test404vs405(t *testing.T) {
	r := New()
	r.GET("/exists", func(c *Context) error {
		return c.OK(nil)
	})

	// 404 - path doesn't exist
	req404 := httptest.NewRequest("GET", "/notexists", nil)
	rec404 := httptest.NewRecorder()
	r.ServeHTTP(rec404, req404)

	// 405 - path exists but wrong method
	req405 := httptest.NewRequest("POST", "/exists", nil)
	rec405 := httptest.NewRecorder()
	r.ServeHTTP(rec405, req405)

	if rec404.Code != 404 {
		t.Errorf("expected 404, got %d", rec404.Code)
	}
	if rec405.Code != 405 {
		t.Errorf("expected 405, got %d", rec405.Code)
	}
}

// -----------------------------------------------------------------------------
// Error Handler
// -----------------------------------------------------------------------------

func TestDefaultErrorHandler(t *testing.T) {
	r := New()
	r.GET("/error", func(c *Context) error {
		return errors.New("something failed")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "internal server error" {
		t.Errorf("expected 'internal server error', got %v", body["error"])
	}
}

func TestCustomErrorHandler(t *testing.T) {
	r := New()

	var capturedErr error
	r.OnErr(func(c *Context, err error) {
		capturedErr = err
		_ = c.JSON(500, M{"custom_error": err.Error()})
	})

	r.GET("/error", func(c *Context) error {
		return errors.New("custom failure")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if capturedErr == nil {
		t.Fatal("expected error to be captured")
	}
	if capturedErr.Error() != "custom failure" {
		t.Errorf("expected 'custom failure', got %v", capturedErr.Error())
	}
}

// -----------------------------------------------------------------------------
// Router Middleware
// -----------------------------------------------------------------------------

func TestRouterUse(t *testing.T) {
	r := New()

	var called bool
	mw := func(next Handler) Handler {
		return func(c *Context) error {
			called = true
			return next(c)
		}
	}

	r.Use(mw)
	r.GET("/test", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !called {
		t.Error("middleware was not called")
	}
}

func TestRouterUseMultiple(t *testing.T) {
	r := New()

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

	r.Use(mw1, mw2)
	r.GET("/test", func(c *Context) error {
		order = append(order, 3)
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("expected order [1,2,3], got %v", order)
	}
}

func TestRouterMiddlewareOrder(t *testing.T) {
	r := New()

	var order []string
	r.Use(func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "before1")
			err := next(c)
			order = append(order, "after1")
			return err
		}
	})
	r.Use(func(next Handler) Handler {
		return func(c *Context) error {
			order = append(order, "before2")
			err := next(c)
			order = append(order, "after2")
			return err
		}
	})

	r.GET("/test", func(c *Context) error {
		order = append(order, "handler")
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	expected := []string{"before1", "before2", "handler", "after2", "after1"}
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

// -----------------------------------------------------------------------------
// Panic Recovery
// -----------------------------------------------------------------------------

func TestPanicRecovery(t *testing.T) {
	r := New()
	r.GET("/panic", func(c *Context) error {
		panic("unexpected panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	r.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected 500 after panic, got %d", rec.Code)
	}
}

func TestPanicRecoveryWithCustomErrorHandler(t *testing.T) {
	r := New()

	var capturedErr error
	r.OnErr(func(c *Context, err error) {
		capturedErr = err
		_ = c.JSON(500, M{"panic": true})
	})

	r.GET("/panic", func(c *Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if capturedErr == nil {
		t.Fatal("expected panic error to be captured")
	}
	if capturedErr.Error() != "panic: test panic" {
		t.Errorf("expected 'panic: test panic', got %v", capturedErr.Error())
	}
}

func TestPanicInErrorHandler(t *testing.T) {
	r := New()
	r.OnErr(func(c *Context, err error) {
		panic("panic in error handler")
	})

	r.GET("/panic", func(c *Context) error {
		panic("first panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	rec := httptest.NewRecorder()

	// Should not panic even if error handler panics
	r.ServeHTTP(rec, req)
}

// -----------------------------------------------------------------------------
// Context Pooling
// -----------------------------------------------------------------------------

func TestContextPooling(t *testing.T) {
	r := New()
	r.GET("/test/{id}", func(c *Context) error {
		id := c.Param("id")
		c.Set("id", id)
		if c.GetString("id") != id {
			t.Error("context data mismatch")
		}
		return c.OK(M{"id": id})
	})

	var wg sync.WaitGroup
	var errors atomic.Int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				req := httptest.NewRequest("GET", "/test/"+string(rune('A'+id%26)), nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				if rec.Code != 200 {
					errors.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()
	if errors.Load() > 0 {
		t.Errorf("had %d errors in concurrent test", errors.Load())
	}
}

func TestContextClearedBetweenRequests(t *testing.T) {
	r := New()

	var prevValue any
	callCount := 0

	r.GET("/test", func(c *Context) error {
		callCount++
		if callCount == 1 {
			c.Set("secret", "sensitive-data")
		} else {
			prevValue = c.Get("secret")
		}
		return c.OK(nil)
	})

	// First request sets data
	req1 := httptest.NewRequest("GET", "/test", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	// Second request should not see previous data
	req2 := httptest.NewRequest("GET", "/test", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if prevValue != nil {
		t.Errorf("context not cleared between requests: got %v", prevValue)
	}
}

// -----------------------------------------------------------------------------
// Edge Cases
// -----------------------------------------------------------------------------

func TestEmptyRouter(t *testing.T) {
	r := New()

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestRootRoute(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"root": true})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTrailingSlash(t *testing.T) {
	r := New()
	r.GET("/path/", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/path/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandlerReturnsNil(t *testing.T) {
	r := New()
	r.GET("/nil", func(c *Context) error {
		c.SetHeader("X-Custom", "value")
		return nil
	})

	req := httptest.NewRequest("GET", "/nil", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMultipleMethodsSamePath(t *testing.T) {
	r := New()
	r.GET("/resource", func(c *Context) error { return c.OK(M{"m": "GET"}) })
	r.POST("/resource", func(c *Context) error { return c.OK(M{"m": "POST"}) })
	r.PUT("/resource", func(c *Context) error { return c.OK(M{"m": "PUT"}) })

	methods := []string{"GET", "POST", "PUT"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/resource", nil)
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
