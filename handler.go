package mux

import (
	"log"
	"net/http"
)

// Handler handles HTTP requests
type Handler func(c *Context) error

// ErrorHandler processes handler errors
type ErrorHandler func(ctx *Context, err error)

var defaultNotFound Handler = func(c *Context) error {
	return c.NotFound(M{"error": "not found"})
}

var defaultMethodNotAllowed Handler = func(c *Context) error {
	return c.MethodNotAllowed(M{"error": "method not allowed"})
}

var defaultErrorHandler ErrorHandler = func(c *Context, err error) {
	response := M{
		"error":   "internal server error",
		"details": err.Error(),
	}
	if err := c.InternalServerError(response); err != nil {
		log.Printf("mux: failed to write error response: %v", err)
	}
}

func (h Handler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("mux: Handler.ServeHTTP should not be called directly")
}
