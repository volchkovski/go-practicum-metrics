package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"go.uber.org/mock/gomock"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	rw := handlers.NewMockMetricsReadWriter(mockCtrl)
	rw.EXPECT().WriteCounter("known", int64(123))
	rw.EXPECT().WriteGauge("known", float64(123))
	rw.EXPECT().ReadCounter("known").Return(int64(1), nil)
	rw.EXPECT().ReadGauge("known").Return(float64(1), nil)
	rw.EXPECT().ReadGauge("unknown").Return(float64(0), errors.New("unknown metric"))
	rw.EXPECT().ReadAllCounters().Return(map[string]int64{"Test1": int64(1)})
	rw.EXPECT().ReadAllGauges().Return(map[string]float64{"Test2": float64(2)})

	ts := httptest.NewServer(metricRouter(rw))
	defer ts.Close()

	type expected struct {
		contentType string
		status      int
		body        string
	}
	type test struct {
		name     string
		path     string
		method   string
		expected expected
	}

	tests := []test{
		{
			name:   "update gauge",
			path:   "/update/counter/known/123",
			method: http.MethodPost,
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "update counter",
			path:   "/update/gauge/known/123",
			method: http.MethodPost,
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "get value known metric",
			path:   "/value/gauge/known",
			method: http.MethodGet,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusOK,
				body:        "1.000",
			},
		},
		{
			name:   "get value known metric",
			path:   "/value/counter/known",
			method: http.MethodGet,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusOK,
				body:        "1",
			},
		},
		{
			name:   "get value unknown metric",
			path:   "/value/gauge/unknown",
			method: http.MethodGet,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "get all metrics",
			path:   "",
			method: http.MethodGet,
			expected: expected{
				contentType: "text/html; charset=utf-8",
				status:      http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "update invalid metric type",
			path:   "/update/gayge/known/123",
			method: http.MethodPost,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "update invalid metric value",
			path:   "/update/gauge/known/test",
			method: http.MethodPost,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "update metric without name and value",
			path:   "/update/gauge",
			method: http.MethodPost,
			expected: expected{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusNotFound,
				body:        "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.path)
			defer resp.Body.Close()
			assert.Equal(t, tt.expected.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.expected.status, resp.StatusCode)
			if tt.expected.body != "" {
				assert.Equal(t, tt.expected.body, body)
			}
		})
	}
}
