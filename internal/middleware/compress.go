package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.zw != nil {
		c.Header().Set("Content-Encoding", "gzip")
	}
	return c.zw.Write(p)
}

func (c *compressWriter) Close() error {
	if c.zw == nil {
		return nil
	}
	return c.zw.Close()
}

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

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

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err1 := c.ReadCloser.Close()
	err2 := c.zr.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

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
			// Создаем временный буфер для тела запроса
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			_ = r.Body.Close()

			// Создаем новый reader для распакованных данных
			cr, err := newCompressReader(io.NopCloser(strings.NewReader(string(body))))
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
