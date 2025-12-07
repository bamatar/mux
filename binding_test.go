package mux

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type testPayload struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestDecodeJSON(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(M{"name": p.Name, "email": p.Email, "age": p.Age})
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john","email":"john@example.com","age":30}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestDecodeJSONEmptyBody(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "body must be valid JSON") {
		t.Errorf("expected 'body must be valid JSON', got %s", rec.Body.String())
	}
}

func TestDecodeJSONSyntaxError(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "badly-formed JSON") {
		t.Errorf("expected 'badly-formed JSON', got %s", rec.Body.String())
	}
}

func TestDecodeJSONUnexpectedEOF(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john"`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "badly-formed JSON") {
		t.Errorf("expected 'badly-formed JSON', got %s", rec.Body.String())
	}
}

func TestDecodeJSONUnknownField(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john","unknown":"field"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unknown field") {
		t.Errorf("expected 'unknown field', got %s", rec.Body.String())
	}
}

func TestDecodeJSONTypeMismatch(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john","age":"not a number"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "incorrect type") {
		t.Errorf("expected 'incorrect type', got %s", rec.Body.String())
	}
}

func TestDecodeJSONMultipleValues(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john"}{"name":"jane"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "single JSON value") {
		t.Errorf("expected 'single JSON value', got %s", rec.Body.String())
	}
}

func TestDecodeJSONBodyTooLarge(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p testPayload
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	// Create a body larger than 1MB
	largeBody := `{"name":"` + strings.Repeat("a", 1_048_577) + `"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "must not exceed") {
		t.Errorf("expected 'must not exceed', got %s", rec.Body.String())
	}
}

func TestDecodeJSONInvalidTarget(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p string
		// Passing non-pointer should panic
		if err := c.Bind(p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"john"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Should recover from panic and return 500
	r.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected 500 after panic, got %d", rec.Code)
	}
}

func TestDecodeJSONTypeMismatchAtPosition(t *testing.T) {
	r := New()
	r.POST("/", func(c *Context) error {
		var p []int
		if err := c.Bind(&p); err != nil {
			return c.BadRequest(M{"error": err.Error()})
		}
		return c.OK(nil)
	})

	req := httptest.NewRequest("POST", "/", strings.NewReader(`["a", "b"]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "incorrect type at position") {
		t.Errorf("expected 'incorrect type at position', got %s", rec.Body.String())
	}
}
