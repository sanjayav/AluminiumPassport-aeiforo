package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapper, r)

		// Log request details
		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v %s %s",
			r.Method,
			r.RequestURI,
			wrapper.statusCode,
			duration,
			r.RemoteAddr,
			r.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StructuredLoggingMiddleware provides more detailed logging
func StructuredLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapper, r)

		// Calculate duration
		duration := time.Since(start)

		// Log structured data
		logEntry := map[string]interface{}{
			"timestamp":      start.Format(time.RFC3339),
			"method":         r.Method,
			"uri":            r.RequestURI,
			"status":         wrapper.statusCode,
			"duration_ms":    duration.Milliseconds(),
			"remote_addr":    r.RemoteAddr,
			"user_agent":     r.UserAgent(),
			"referer":        r.Referer(),
			"content_length": r.ContentLength,
		}

		// Add user info if available
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			logEntry["user_id"] = userID
		}

		// Add request ID if available
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			logEntry["request_id"] = requestID
		}

		// Log based on status code
		if wrapper.statusCode >= 500 {
			log.Printf("ERROR: %+v", logEntry)
		} else if wrapper.statusCode >= 400 {
			log.Printf("WARN: %+v", logEntry)
		} else {
			log.Printf("INFO: %+v", logEntry)
		}
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to request headers for downstream handlers
		r.Header.Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple timestamp-based ID (in production, use UUID or similar)
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
