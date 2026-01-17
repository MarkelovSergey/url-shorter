package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ResponseWriterLogger обертка для http.ResponseWriter с логированием.
type ResponseWriterLogger struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader записывает статус-код и сохраняет его.
func (rw *ResponseWriterLogger) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write записывает данные и подсчитывает размер ответа.
func (rw *ResponseWriterLogger) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n

	return n, err
}

// Logging создает мидлвар для логирования HTTP-запросов.
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
