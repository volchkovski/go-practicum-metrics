package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(code int) {
	if code < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(code)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func WithCompress(h http.Handler) http.Handler {
	compressFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			logger.Log.Debugln("Created compressWriter")
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					logger.Log.Errorf("Failed to close compressWriter: %s", err.Error())
				}
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			logger.Log.Debug("Created compressReader")
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					logger.Log.Errorf("Failed to close compressReader: %s", err.Error())
				}
			}()
		}

		h.ServeHTTP(ow, r)

	}
	return http.HandlerFunc(compressFn)
}
