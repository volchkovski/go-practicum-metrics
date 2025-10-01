package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

// compressWriter wraps http.ResponseWriter to provide gzip compression for responses.
// It automatically sets the Content-Encoding header when data is written.
type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter that writes gzip-compressed data
// to the provided ResponseWriter.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

// Write compresses the data using gzip and writes it to the underlying ResponseWriter.
// It automatically sets the Content-Encoding header to "gzip" on first write.
func (c *compressWriter) Write(p []byte) (int, error) {
	if c.zw != nil {
		c.Header().Set("Content-Encoding", "gzip")
	}
	return c.zw.Write(p)
}

// Close finalizes the gzip stream. Must be called to ensure all compressed
// data is properly written to the underlying ResponseWriter.
func (c *compressWriter) Close() error {
	if c.zw == nil {
		return nil
	}
	return c.zw.Close()
}

// compressReader wraps an io.ReadCloser to provide gzip decompression for request bodies.
// It reads compressed data from the source and provides uncompressed data to consumers.
type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader that decompresses gzip data
// from the provided ReadCloser. The original reader will be closed when
// this compressReader is closed.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		_ = r.Close()
		return nil, err
	}
	return &compressReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}

// Read decompresses data from the underlying gzip reader.
func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

// Close closes both the gzip reader and the underlying ReadCloser.
// It returns the first error encountered, if any.
func (c *compressReader) Close() error {
	err1 := c.ReadCloser.Close()
	err2 := c.zr.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// WithCompress is HTTP middleware that handles gzip compression and decompression.
//
// For responses: If the client accepts gzip encoding (Accept-Encoding: gzip),
// the response will be automatically compressed.
//
// For requests: If the request body is gzip-compressed (Content-Encoding: gzip),
// it will be automatically decompressed using streaming approach without
// loading the entire body into memory.
//
// This middleware is safe for large request/response bodies as it uses
// streaming compression/decompression.
func WithCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// Обработка сжатого ответа
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					logger.Log.Errorf("Failed to close compressWriter: %v", err)
				}
			}()
		}

		// Обработка сжатого запроса
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// Заменяем r.Body на распакованный reader напрямую, не читая в память
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip encoding", http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					logger.Log.Errorf("Failed to close compressReader: %v", err)
				}
			}()
		}

		h.ServeHTTP(ow, r)
	})
}
