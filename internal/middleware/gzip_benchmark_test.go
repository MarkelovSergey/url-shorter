package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkGzipMiddlewareWithCompression(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Генерируем достаточно большой ответ для сжатия
		response := `{"result":"` + strings.Repeat("http://localhost:8080/ABC12345,", 100) + `"}`
		w.Write([]byte(response))
	})

	gzipHandler := Gzipping(handler)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)
	}
}

func BenchmarkGzipMiddlewareWithoutCompression(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"http://localhost:8080/ABC12345"}`))
	})

	gzipHandler := Gzipping(handler)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Без Accept-Encoding: gzip
		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)
	}
}

func BenchmarkGzipDecompression(b *testing.B) {
	// Подготавливаем сжатые данные
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	gzWriter.Write([]byte(`{"url":"https://example.com/test"}`))
	gzWriter.Close()
	compressedData := buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	gzipHandler := Gzipping(handler)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		body := bytes.NewReader(compressedData)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()
		gzipHandler.ServeHTTP(w, req)
	}
}

func BenchmarkGzipCompressionDifferentSizes(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"small 100B", 100},
		{"medium 1KB", 1024},
		{"large 10KB", 10240},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			data := strings.Repeat("a", s.size)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(data))
			})

			gzipHandler := Gzipping(handler)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				w := httptest.NewRecorder()
				gzipHandler.ServeHTTP(w, req)
			}
		})
	}
}
