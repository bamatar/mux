package mux

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"
)

// Logger returns a middleware that logs requests using default slog
func Logger() Middleware {
	return LoggerWith(slog.Default())
}

// LoggerWith returns a middleware that logs requests using provided slog
func LoggerWith(logger *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(c *Context) error {
			start := time.Now()
			err := next(c)
			logger.Info("request",
				"method", c.Method(),
				"path", c.Path(),
				"status", c.w.Status(),
				"size", c.w.Size(),
				"duration", time.Since(start),
			)
			return err
		}
	}
}

// RequestIDConfig configures RequestID middleware
type RequestIDConfig struct {
	// Header name. Default: X-Request-ID
	Header string

	// Generator creates request IDs. Default: random 16-byte hex string
	Generator func() string
}

// RequestID adds a unique request ID to each request
func RequestID(config ...RequestIDConfig) Middleware {
	cfg := RequestIDConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Header == "" {
		cfg.Header = "X-Request-ID"
	}
	if cfg.Generator == nil {
		cfg.Generator = func() string {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				panic(err)
			}
			return hex.EncodeToString(b)
		}
	}

	return func(next Handler) Handler {
		return func(c *Context) error {
			id := c.Header(cfg.Header)
			if id == "" {
				id = cfg.Generator()
			}
			c.SetHeader(cfg.Header, id)
			c.Set(cfg.Header, id)
			return next(c)
		}
	}
}
