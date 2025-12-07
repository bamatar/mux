package mux

import (
	"net/http/httptest"
	"testing"
)

func TestResponseWriterStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	rw.WriteHeader(201)

	if rw.Status() != 201 {
		t.Errorf("expected status 201, got %d", rw.Status())
	}
	if rec.Code != 201 {
		t.Errorf("expected underlying status 201, got %d", rec.Code)
	}
}

func TestResponseWriterSize(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	rw.WriteHeader(200)
	n, err := rw.Write([]byte("hello"))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if rw.Size() != 5 {
		t.Errorf("expected size 5, got %d", rw.Size())
	}
}

func TestResponseWriterMultipleWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	rw.WriteHeader(200)
	rw.Write([]byte("hello"))
	rw.Write([]byte(" world"))

	if rw.Size() != 11 {
		t.Errorf("expected size 11, got %d", rw.Size())
	}
	if rec.Body.String() != "hello world" {
		t.Errorf("expected 'hello world', got %s", rec.Body.String())
	}
}

func TestResponseWriterDefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	if rw.Status() != 0 {
		t.Errorf("expected default status 0, got %d", rw.Status())
	}
}

func TestResponseWriterDefaultSize(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &ResponseWriter{ResponseWriter: rec}

	if rw.Size() != 0 {
		t.Errorf("expected default size 0, got %d", rw.Size())
	}
}
