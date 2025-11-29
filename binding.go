package mux

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type M map[string]any

// decodeJSON parses JSON request body into v
func decodeJSON(c *Context, v any) error {
	// limit request body to 1MB
	maxBytes := 1_048_576
	c.r.Body = http.MaxBytesReader(c.w, c.r.Body, int64(maxBytes))

	decoder := json.NewDecoder(c.r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(v)
	if err == nil {
		// ensure body contains only one JSON value
		err = decoder.Decode(&struct{}{})
		if err != io.EOF {
			return errors.New("body must only contain a single JSON value")
		}

		return nil
	}

	var syntaxError *json.SyntaxError
	var maxBytesError *http.MaxBytesError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError

	// programmer error
	if errors.As(err, &invalidUnmarshalError) {
		panic(err)
	}

	// empty body
	if errors.Is(err, io.EOF) {
		return errors.New("body must be valid JSON")
	}

	// unexpected EOF (https://github.com/golang/go/issues/25956)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return errors.New("body contains badly-formed JSON")
	}

	// body too large
	if errors.As(err, &maxBytesError) {
		return fmt.Errorf("body must not exceed %d bytes", maxBytesError.Limit)
	}

	// unknown field
	if strings.HasPrefix(err.Error(), "json: unknown field ") {
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("body contains unknown field %s", fieldName)
	}

	// syntax error
	if errors.As(err, &syntaxError) {
		return errors.New("body contains badly-formed JSON")
	}

	// type mismatch
	if errors.As(err, &unmarshalTypeError) {
		if unmarshalTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect type for field %q", unmarshalTypeError.Field)
		}

		return fmt.Errorf("body contains incorrect type at position %d", unmarshalTypeError.Offset)
	}

	return err
}
