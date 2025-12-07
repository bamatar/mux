package mux

import "net/http"

// ResponseWriter wraps http.ResponseWriter to track status and size
type ResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Status returns the response status code
func (r *ResponseWriter) Status() int {
	return r.status
}

// Size returns the number of bytes written
func (r *ResponseWriter) Size() int {
	return r.size
}

// WriteHeader captures status and writes header
func (r *ResponseWriter) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write captures size and writes data
func (r *ResponseWriter) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}
