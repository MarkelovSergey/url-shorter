package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr := gzipReaderPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		gzipReaderPool.Put(zr)
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	err := c.zr.Close()
	gzipReaderPool.Put(c.zr)

	if closeErr := c.r.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

type gzipWriterWithContentType struct {
	http.ResponseWriter
	gzipWriter     *gzip.Writer
	shouldCompress func(string) bool
	headerWritten  bool
	gzipEnabled    bool
}

func (w *gzipWriterWithContentType) WriteHeader(statusCode int) {
	if !w.headerWritten {
		contentType := w.ResponseWriter.Header().Get("Content-Type")
		w.gzipEnabled = w.shouldCompress(contentType)

		if w.gzipEnabled {
			gzw := gzipWriterPool.Get().(*gzip.Writer)
			gzw.Reset(w.ResponseWriter)
			w.gzipWriter = gzw
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			w.ResponseWriter.Header().Del("Content-Length")
		}

		w.headerWritten = true
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipWriterWithContentType) Write(b []byte) (int, error) {
	if !w.headerWritten {
		contentType := w.ResponseWriter.Header().Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(b)
			w.ResponseWriter.Header().Set("Content-Type", contentType)
		}

		w.gzipEnabled = w.shouldCompress(contentType)

		if w.gzipEnabled {
			gzw := gzipWriterPool.Get().(*gzip.Writer)
			gzw.Reset(w.ResponseWriter)
			w.gzipWriter = gzw
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			w.ResponseWriter.Header().Del("Content-Length")
		}

		w.headerWritten = true
	}

	if w.gzipEnabled && w.gzipWriter != nil {
		return w.gzipWriter.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

func Gzipping(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := newGzipReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()
		}

		if supportsGzip {
			gzw := &gzipWriterWithContentType{
				ResponseWriter: w,
				shouldCompress: func(ct string) bool {
					return strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html")
				},
			}

			next.ServeHTTP(gzw, r)

			if gzw.gzipWriter != nil {
				gzw.gzipWriter.Close()
				gzipWriterPool.Put(gzw.gzipWriter)
			}

			return
		}

		next.ServeHTTP(w, r)
	})
}
