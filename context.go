package mux

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
)

type Context struct {
	w http.ResponseWriter
	r *http.Request

	// request-scoped storage
	locals []local
}

type local struct {
	key   string
	value any
}

// Request Info

// Method returns the HTTP method
func (c *Context) Method() string {
	return c.r.Method
}

// Path returns the URL path
func (c *Context) Path() string {
	return c.r.URL.Path
}

// Context returns the request context
func (c *Context) Context() context.Context {
	return c.r.Context()
}

// Path parameters

// Param returns a path parameter by name
func (c *Context) Param(name string) string {
	return c.r.PathValue(name)
}

// Query parameters

// Query returns a query parameter by name
func (c *Context) Query(key string, fallback ...string) string {
	v := c.r.URL.Query().Get(key)
	if v == "" && len(fallback) > 0 {
		return fallback[0]
	}
	return v
}

// QueryInt parses a query parameter as int
func (c *Context) QueryInt(key string, fallback ...int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil && len(fallback) > 0 {
		return fallback[0]
	}
	return v
}

// Queries returns all query parameters
func (c *Context) Queries() url.Values {
	return c.r.URL.Query()
}

// Headers

// Header returns a request header by key
func (c *Context) Header(key string) string {
	return c.r.Header.Get(key)
}

// SetHeader sets a response header
func (c *Context) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// Cookies

// Cookie returns a request cookie value by name
func (c *Context) Cookie(name string, fallback ...string) string {
	cookie, err := c.r.Cookie(name)
	if err != nil && len(fallback) > 0 {
		return fallback[0]
	}
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

// SetCookie adds a cookie to the response
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

// Locals

// Set stores a value in locals
func (c *Context) Set(key string, value any) {
	for i := range c.locals {
		if c.locals[i].key == key {
			c.locals[i].value = value
			return
		}
	}
	if value == nil {
		return
	}
	n := len(c.locals)
	if cap(c.locals) > n {
		c.locals = c.locals[:n+1]
		c.locals[n].key = key
		c.locals[n].value = value
		return
	}
	c.locals = append(c.locals, local{key, value})
}

// Get retrieves a value from locals
func (c *Context) Get(key string) any {
	for i := range c.locals {
		if c.locals[i].key == key {
			return c.locals[i].value
		}
	}
	return nil
}

// GetString retrieves a string from locals
func (c *Context) GetString(key string) string {
	if v, ok := c.Get(key).(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int from locals
func (c *Context) GetInt(key string) int {
	if v, ok := c.Get(key).(int); ok {
		return v
	}
	return 0
}

// GetBool retrieves a bool from locals
func (c *Context) GetBool(key string) bool {
	if v, ok := c.Get(key).(bool); ok {
		return v
	}
	return false
}

// Body parsing

// Body returns the raw request body
func (c *Context) Body() ([]byte, error) {
	return io.ReadAll(c.r.Body)
}

// Bind decodes request body into v with auto-detect content type
func (c *Context) Bind(v any) error {
	switch c.ContentType() {
	case MIMEApplicationXML, MIMETextXML:
		// TODO: implement decodeXML(c, v)
		return nil
	case MIMEApplicationForm, MIMEMultipartForm:
		// TODO: implement decodeForm(c, v)
		return nil
	default:
		return decodeJSON(c, v)
	}
}

// FormValue returns a form field by name
func (c *Context) FormValue(name string) string {
	return c.r.FormValue(name)
}

// ContentType returns the parsed media type from Content-Type header
func (c *Context) ContentType() string {
	ct, _, _ := mime.ParseMediaType(c.r.Header.Get("Content-Type"))
	return ct
}

// Status Response

// Status writes a status code with empty body
func (c *Context) Status(code int) error {
	return c.Blob(code, "", nil)
}

// NoContent writes 204 No Content
func (c *Context) NoContent() error {
	return c.Blob(http.StatusNoContent, "", nil)
}

// JSON helpers

// JSON writes a JSON response
func (c *Context) JSON(status int, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Blob(status, MIMEApplicationJSON, data)
}

// OK writes 200 JSON response
func (c *Context) OK(v any) error {
	return c.JSON(http.StatusOK, v)
}

// Created writes 201 JSON response
func (c *Context) Created(v any) error {
	return c.JSON(http.StatusCreated, v)
}

// BadRequest writes 400 JSON response
func (c *Context) BadRequest(v any) error {
	return c.JSON(http.StatusBadRequest, v)
}

// Unauthorized writes 401 JSON response
func (c *Context) Unauthorized(v any) error {
	return c.JSON(http.StatusUnauthorized, v)
}

// Forbidden writes 403 JSON response
func (c *Context) Forbidden(v any) error {
	return c.JSON(http.StatusForbidden, v)
}

// NotFound writes 404 JSON response
func (c *Context) NotFound(v any) error {
	return c.JSON(http.StatusNotFound, v)
}

// MethodNotAllowed writes 405 JSON response
func (c *Context) MethodNotAllowed(v any) error {
	return c.JSON(http.StatusMethodNotAllowed, v)
}

// InternalServerError writes 500 JSON response
func (c *Context) InternalServerError(v any) error {
	return c.JSON(http.StatusInternalServerError, v)
}

// Response

// String writes a plain text response
func (c *Context) String(status int, s string) error {
	return c.Blob(status, MIMETextPlain, []byte(s))
}

// HTML writes an HTML response
func (c *Context) HTML(status int, html string) error {
	return c.Blob(status, MIMETextHTML, []byte(html))
}

// Blob writes raw bytes with content type
func (c *Context) Blob(status int, contentType string, data []byte) error {
	if contentType != "" {
		c.w.Header().Set("Content-Type", contentType)
	}

	c.w.WriteHeader(status)

	if len(data) > 0 {
		_, err := c.w.Write(data)
		return err
	}
	return nil
}
