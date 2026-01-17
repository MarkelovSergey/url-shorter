package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWriterLogger struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *ResponseWriterLogger) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriterLogger) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n

	return n, err
}

func Logging(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rwl := &ResponseWriterLogger{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rwl, r)
			dur := time.Since(start)

			logger.Info("HTTP request",
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.Duration("duration", dur),
				zap.Int("status", rwl.status),
				zap.Int("size", rwl.size),
			)
		})
	}
}
