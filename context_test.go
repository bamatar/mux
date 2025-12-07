package mux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// Request Info
// -----------------------------------------------------------------------------

func TestContextMethod(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		return c.OK(M{"method": c.Method()})
	})

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["method"] != "POST" {
		t.Errorf("expected POST, got %v", body["method"])
	}
}

func TestContextPath(t *testing.T) {
	r := New()
	r.GET("/users/list", func(c *Context) error {
		return c.OK(M{"path": c.Path()})
	})

	req := httptest.NewRequest("GET", "/users/list", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["path"] != "/users/list" {
		t.Errorf("expected /users/list, got %v", body["path"])
	}
}

func TestContextContext(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		ctx := c.Context()
		if ctx == nil {
			t.Error("expected non-nil context")
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}

// -----------------------------------------------------------------------------
// Query Parameters
// -----------------------------------------------------------------------------

func TestContextQuery(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"name": c.Query("name")})
	})

	req := httptest.NewRequest("GET", "/?name=john", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["name"] != "john" {
		t.Errorf("expected john, got %v", body["name"])
	}
}

func TestContextQueryFallback(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"name": c.Query("name", "default")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["name"] != "default" {
		t.Errorf("expected default, got %v", body["name"])
	}
}

func TestContextQueryInt(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"page": c.QueryInt("page")})
	})

	req := httptest.NewRequest("GET", "/?page=5", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["page"] != float64(5) {
		t.Errorf("expected 5, got %v", body["page"])
	}
}

func TestContextQueryIntFallback(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"page": c.QueryInt("page", 1)})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["page"] != float64(1) {
		t.Errorf("expected 1, got %v", body["page"])
	}
}

func TestContextQueryIntInvalid(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"page": c.QueryInt("page", 10)})
	})

	req := httptest.NewRequest("GET", "/?page=invalid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["page"] != float64(10) {
		t.Errorf("expected 10, got %v", body["page"])
	}
}

func TestContextQueries(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		q := c.Queries()
		return c.OK(M{"a": q.Get("a"), "b": q.Get("b")})
	})

	req := httptest.NewRequest("GET", "/?a=1&b=2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["a"] != "1" || body["b"] != "2" {
		t.Errorf("expected a=1 b=2, got %v", body)
	}
}

// -----------------------------------------------------------------------------
// Headers
// -----------------------------------------------------------------------------

func TestContextHeader(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"auth": c.Header("Authorization")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["auth"] != "Bearer token123" {
		t.Errorf("expected Bearer token123, got %v", body["auth"])
	}
}

func TestContextSetHeader(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.SetHeader("X-Custom", "value")
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("X-Custom") != "value" {
		t.Errorf("expected X-Custom=value, got %s", rec.Header().Get("X-Custom"))
	}
}

// -----------------------------------------------------------------------------
// Cookies
// -----------------------------------------------------------------------------

func TestContextCookie(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"session": c.Cookie("session")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["session"] != "abc123" {
		t.Errorf("expected abc123, got %v", body["session"])
	}
}

func TestContextCookieFallback(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"session": c.Cookie("session", "default")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["session"] != "default" {
		t.Errorf("expected default, got %v", body["session"])
	}
}

func TestContextSetCookie(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.SetCookie(&http.Cookie{Name: "session", Value: "xyz789"})
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "xyz789" {
		t.Errorf("expected session=xyz789, got %v", cookies)
	}
}

// -----------------------------------------------------------------------------
// Locals
// -----------------------------------------------------------------------------

func TestContextSetGet(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("key", "value")
		return c.OK(M{"key": c.Get("key")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["key"] != "value" {
		t.Errorf("expected value, got %v", body["key"])
	}
}

func TestContextSetOverwrite(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("key", "first")
		c.Set("key", "second")
		return c.OK(M{"key": c.Get("key")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["key"] != "second" {
		t.Errorf("expected second, got %v", body["key"])
	}
}

func TestContextSetNil(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("key", nil)
		return c.OK(M{"key": c.Get("key")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["key"] != nil {
		t.Errorf("expected nil, got %v", body["key"])
	}
}

func TestContextGetMissing(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		v := c.Get("missing")
		if v != nil {
			t.Errorf("expected nil, got %v", v)
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}

func TestContextGetString(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("name", "john")
		return c.OK(M{"name": c.GetString("name")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["name"] != "john" {
		t.Errorf("expected john, got %v", body["name"])
	}
}

func TestContextGetStringWrongType(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("num", 123)
		return c.OK(M{"num": c.GetString("num")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["num"] != "" {
		t.Errorf("expected empty string, got %v", body["num"])
	}
}

func TestContextGetInt(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("count", 42)
		return c.OK(M{"count": c.GetInt("count")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["count"] != float64(42) {
		t.Errorf("expected 42, got %v", body["count"])
	}
}

func TestContextGetIntWrongType(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("count", "not an int")
		return c.OK(M{"count": c.GetInt("count")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["count"] != float64(0) {
		t.Errorf("expected 0, got %v", body["count"])
	}
}

func TestContextGetBool(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("active", true)
		return c.OK(M{"active": c.GetBool("active")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["active"] != true {
		t.Errorf("expected true, got %v", body["active"])
	}
}

func TestContextGetBoolWrongType(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		c.Set("active", "yes")
		return c.OK(M{"active": c.GetBool("active")})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["active"] != false {
		t.Errorf("expected false, got %v", body["active"])
	}
}

// -----------------------------------------------------------------------------
// Body Parsing
// -----------------------------------------------------------------------------

func TestContextBody(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		body, err := c.Body()
		if err != nil {
			return err
		}
		return c.OK(M{"body": string(body)})
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader("raw body content"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["body"] != "raw body content" {
		t.Errorf("expected raw body content, got %v", body["body"])
	}
}

func TestContextFormValue(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		return c.OK(M{"name": c.FormValue("name")})
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader("name=john"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["name"] != "john" {
		t.Errorf("expected john, got %v", body["name"])
	}
}

func TestContextContentType(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		return c.OK(M{"ct": c.ContentType()})
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var body M
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body["ct"] != "application/json" {
		t.Errorf("expected application/json, got %v", body["ct"])
	}
}

// -----------------------------------------------------------------------------
// Response Helpers
// -----------------------------------------------------------------------------

func TestContextStatus(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.Status(202)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 202 {
		t.Errorf("expected 202, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body, got %s", rec.Body.String())
	}
}

func TestContextNoContent(t *testing.T) {
	r := New()
	r.DELETE("/", func(c *Context) error {
		return c.NoContent()
	})

	req := httptest.NewRequest("DELETE", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 204 {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestContextJSON(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.JSON(200, M{"key": "value"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestContextOK(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.OK(M{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestContextCreated(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		return c.Created(M{"id": 1})
	})

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestContextBadRequest(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		return c.BadRequest(M{"error": "invalid"})
	})

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestContextUnauthorized(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.Unauthorized(M{"error": "unauthorized"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestContextForbidden(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.Forbidden(M{"error": "forbidden"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestContextNotFound(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.NotFound(M{"error": "not found"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestContextMethodNotAllowed(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.MethodNotAllowed(M{"error": "method not allowed"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestContextInternalServerError(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.InternalServerError(M{"error": "server error"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// -----------------------------------------------------------------------------
// Response Types
// -----------------------------------------------------------------------------

func TestContextString(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.String(200, "hello world")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected text/plain, got %s", rec.Header().Get("Content-Type"))
	}
	if rec.Body.String() != "hello world" {
		t.Errorf("expected hello world, got %s", rec.Body.String())
	}
}

func TestContextHTML(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.HTML(200, "<h1>Hello</h1>")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "text/html" {
		t.Errorf("expected text/html, got %s", rec.Header().Get("Content-Type"))
	}
	if rec.Body.String() != "<h1>Hello</h1>" {
		t.Errorf("expected <h1>Hello</h1>, got %s", rec.Body.String())
	}
}

func TestContextBlob(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.Blob(200, "application/octet-stream", []byte{0x01, 0x02, 0x03})
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got %s", rec.Header().Get("Content-Type"))
	}
	if len(rec.Body.Bytes()) != 3 {
		t.Errorf("expected 3 bytes, got %d", len(rec.Body.Bytes()))
	}
}

func TestContextBlobEmptyContentType(t *testing.T) {
	r := New()
	r.GET("/", func(c *Context) error {
		return c.Blob(200, "", []byte("data"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Type") != "" {
		t.Errorf("expected empty content-type, got %s", rec.Header().Get("Content-Type"))
	}
}

// -----------------------------------------------------------------------------
// Internal
// -----------------------------------------------------------------------------

func TestContextDetach(t *testing.T) {
	c := &Context{}
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if c.Get("key1") != "value1" {
		t.Error("expected value1")
	}

	c.detach()

	if c.Get("key1") != nil {
		t.Error("expected nil after detach")
	}
	if c.Get("key2") != nil {
		t.Error("expected nil after detach")
	}
	if len(c.locals) != 0 {
		t.Errorf("expected empty locals, got %d", len(c.locals))
	}
}
