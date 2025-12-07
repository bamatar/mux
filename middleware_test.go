package mux

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	r := New()
	r.Use(LoggerWith(logger))
	r.GET("/test", func(c *Context) error {
		return c.OK(M{"message": "hello"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	log := buf.String()

	if !strings.Contains(log, "method=GET") {
		t.Errorf("expected method=GET, got %s", log)
	}
	if !strings.Contains(log, "path=/test") {
		t.Errorf("expected path=/test, got %s", log)
	}
	if !strings.Contains(log, "status=200") {
		t.Errorf("expected status=200, got %s", log)
	}
	if !strings.Contains(log, "duration=") {
		t.Errorf("expected duration=, got %s", log)
	}
}

func TestLoggerStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	r := New()
	r.Use(LoggerWith(logger))
	r.POST("/users", func(c *Context) error {
		return c.Created(M{"id": 1})
	})

	req := httptest.NewRequest("POST", "/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	log := buf.String()

	if !strings.Contains(log, "status=201") {
		t.Errorf("expected status=201, got %s", log)
	}
}

func TestLoggerSize(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	r := New()
	r.Use(LoggerWith(logger))
	r.GET("/test", func(c *Context) error {
		return c.String(200, "hello")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	log := buf.String()

	if !strings.Contains(log, "size=5") {
		t.Errorf("expected size=5, got %s", log)
	}
}

func TestLoggerWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	r := New()
	r.Use(LoggerWith(logger))
	r.GET("/error", func(c *Context) error {
		return c.InternalServerError(M{"error": "failed"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	log := buf.String()

	if !strings.Contains(log, "status=500") {
		t.Errorf("expected status=500, got %s", log)
	}
}

func TestLoggerDefault(t *testing.T) {
	r := New()
	r.Use(Logger())
	r.GET("/test", func(c *Context) error {
		return c.OK(nil)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
