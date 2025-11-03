package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressWriter(t *testing.T) {
	t.Run("new compress writer", func(t *testing.T) {
		rec := httptest.NewRecorder()
		cw := newCompressWriter(rec)

		assert.NotNil(t, cw)
		assert.NotNil(t, cw.zw)
		assert.Equal(t, rec, cw.ResponseWriter)
	})

	t.Run("write with compression", func(t *testing.T) {
		rec := httptest.NewRecorder()
		cw := newCompressWriter(rec)

		data := []byte("test data for compression")
		n, err := cw.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		err = cw.Close()
		assert.NoError(t, err)

		// Verify the data was compressed
		reader, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)

		assert.Equal(t, data, decompressed)
	})

	t.Run("close nil writer", func(t *testing.T) {
		cw := &compressWriter{zw: nil}
		err := cw.Close()
		assert.NoError(t, err)
	})
}

func TestCompressReader(t *testing.T) {
	t.Run("successful decompression", func(t *testing.T) {
		originalData := []byte("test data for decompression")

		// Create compressed data
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write(originalData)
		require.NoError(t, err)
		err = gzw.Close()
		require.NoError(t, err)

		// Create reader
		cr, err := newCompressReader(io.NopCloser(&buf))
		require.NoError(t, err)

		// Read decompressed data
		result, err := io.ReadAll(cr)
		require.NoError(t, err)

		assert.Equal(t, originalData, result)

		err = cr.Close()
		assert.NoError(t, err)
	})

	t.Run("invalid gzip data", func(t *testing.T) {
		invalidData := bytes.NewBufferString("not gzip data")

		cr, err := newCompressReader(io.NopCloser(invalidData))
		assert.Error(t, err)
		assert.Nil(t, cr)
	})

	t.Run("read from compress reader", func(t *testing.T) {
		originalData := []byte("test data")

		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write(originalData)
		require.NoError(t, err)
		err = gzw.Close()
		require.NoError(t, err)

		cr, err := newCompressReader(io.NopCloser(&buf))
		require.NoError(t, err)

		readBuf := make([]byte, len(originalData))
		n, err := cr.Read(readBuf)

		// EOF can be expected when reading all data
		if err != nil && err != io.EOF {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, len(originalData), n)
		assert.Equal(t, originalData, readBuf)
	})
}

func TestWithCompress(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading body", http.StatusInternalServerError)
			return
		}
		_, err = w.Write([]byte("response: " + string(body)))
		if err != nil {
			http.Error(w, "error writing response", http.StatusInternalServerError)
			return
		}
	})

	compressedHandler := WithCompress(handler)

	t.Run("request without gzip", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("test body"))
		rec := httptest.NewRecorder()

		compressedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: test body", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("request with accept-encoding gzip", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("test body"))
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		compressedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		// Decompress and verify
		reader, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)

		assert.Equal(t, "response: test body", string(decompressed))
	})

	t.Run("request with compressed body", func(t *testing.T) {
		originalBody := "compressed request body"

		// Compress the body
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write([]byte(originalBody))
		require.NoError(t, err)
		err = gzw.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		rec := httptest.NewRecorder()

		compressedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: "+originalBody, rec.Body.String())
	})

	t.Run("request with invalid compressed body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("invalid gzip data"))
		req.Header.Set("Content-Encoding", "gzip")
		rec := httptest.NewRecorder()

		compressedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid gzip encoding")
	})

	t.Run("request with both gzip accept and compressed body", func(t *testing.T) {
		originalBody := "test data"

		// Compress the request body
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		_, err := gzw.Write([]byte(originalBody))
		require.NoError(t, err)
		err = gzw.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		compressedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		// Decompress response
		reader, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)

		assert.Equal(t, "response: "+originalBody, string(decompressed))
	})
}
