package middleware

import (
	"net/http"
	"strings"
	"time"

	"webhook-forge/pkg/logger"
)

// RequestLogger is a middleware that logs all incoming requests with IP address information
type RequestLogger struct {
	logger logger.Logger
}

// responseWriter is a wrapper for http.ResponseWriter that captures status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Status returns the HTTP status code
func (rw *responseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

// Size returns the response size
func (rw *responseWriter) Size() int {
	return rw.size
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger(logger logger.Logger) *RequestLogger {
	return &RequestLogger{
		logger: logger,
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (common for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2, ...), take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header (used by some proxies)
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr from the request
	// RemoteAddr is in the form "IP:port", so strip the port
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Remove brackets from IPv6 addresses
	ip = strings.TrimPrefix(ip, "[")
	ip = strings.TrimSuffix(ip, "]")

	return ip
}

// Middleware returns an http.Handler middleware function
func (m *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := getClientIP(r)

		// Create response writer wrapper
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
			size:           0,
		}

		// Log request start
		m.logger.Info("Request started",
			logger.Field{Key: "method", Value: r.Method},
			logger.Field{Key: "path", Value: r.URL.Path},
			logger.Field{Key: "ip", Value: clientIP})

		// Call the next handler with our wrapped response writer
		next.ServeHTTP(rw, r)

		// Log request completion with status code and response size
		duration := time.Since(start)

		// Use appropriate log level based on status code
		logMsg := "Request completed"
		if rw.Status() >= 500 {
			m.logger.Error(logMsg,
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "path", Value: r.URL.Path},
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "status", Value: rw.Status()},
				logger.Field{Key: "size", Value: rw.Size()},
				logger.Field{Key: "duration_ms", Value: duration.Milliseconds()})
		} else if rw.Status() >= 400 {
			m.logger.Warn(logMsg,
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "path", Value: r.URL.Path},
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "status", Value: rw.Status()},
				logger.Field{Key: "size", Value: rw.Size()},
				logger.Field{Key: "duration_ms", Value: duration.Milliseconds()})
		} else {
			m.logger.Info(logMsg,
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "path", Value: r.URL.Path},
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "status", Value: rw.Status()},
				logger.Field{Key: "size", Value: rw.Size()},
				logger.Field{Key: "duration_ms", Value: duration.Milliseconds()})
		}
	})
}
