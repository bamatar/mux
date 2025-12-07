package mux

import (
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
