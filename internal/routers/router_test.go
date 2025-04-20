package routers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/mock/gomock"
)

type expected struct {
	contentType string
	status      int
	body        string
}
type test struct {
	name     string
	path     string
	method   string
	mock     func()
	expected expected
}

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

func testIter(ts *httptest.Server, tc test) func(*testing.T) {
	return func(t *testing.T) {
		tc.mock()
		resp, body := testRequest(t, ts, tc.method, tc.path)
		defer resp.Body.Close()
		assert.Contains(t, resp.Header.Get("Content-Type"), tc.expected.contentType)
		assert.Equal(t, tc.expected.status, resp.StatusCode)
		if tc.expected.body != "" {
			assert.Equal(t, tc.expected.body, body)
		}
	}
}

func TestRouterUpdateMetric(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "update valid gauge",
			path:   "/update/gauge/test/123",
			method: http.MethodPost,
			mock: func() {
				service.EXPECT().
					PushGaugeMetric(&m.GaugeMetric{Name: "test", Value: float64(123)}).
					Return(nil)
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "update valid counter",
			path:   "/update/counter/test/123",
			method: http.MethodPost,
			mock: func() {
				service.EXPECT().
					PushCounterMetric(&m.CounterMetric{Name: "test", Value: int64(123)}).
					Return(nil)
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "update invalid metric type",
			path:   "/update/gayge/test/123",
			method: http.MethodPost,
			mock:   func() {},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "update invalid metric value",
			path:   "/update/gauge/known/test",
			method: http.MethodPost,
			mock:   func() {},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "update metric without name and value",
			path:   "/update/gauge",
			method: http.MethodPost,
			mock:   func() {},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusNotFound,
				body:        "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, testIter(ts, tc))
	}
}

func TestRouterGetMetric(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "get valid gauge",
			path:   "/value/gauge/test",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetGaugeMetric("test").
					Return(&m.GaugeMetric{Name: "test", Value: float64(1.1)}, nil)
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "1.1",
			},
		},
		{
			name:   "get valid counter",
			path:   "/value/counter/test",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetCounterMetric("test").
					Return(&m.CounterMetric{Name: "test", Value: int64(1)}, nil)
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusOK,
				body:        "1",
			},
		},
		{
			name:   "get not existing gauge",
			path:   "/value/gauge/test",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetGaugeMetric("test").
					Return(nil, errors.New("not existing metric"))
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "get not existing counter",
			path:   "/value/counter/test",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetCounterMetric("test").
					Return(nil, errors.New("not existing metric"))
			},
			expected: expected{
				contentType: "text/plain",
				status:      http.StatusNotFound,
				body:        "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, testIter(ts, tc))
	}
}

func TestRouterAllMetricsHTML(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "get valid all metrics",
			path:   "",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetAllGaugeMetrics().
					Return([]*m.GaugeMetric{
						{Name: "test", Value: float64(123)},
					}, nil)
				service.EXPECT().GetAllCounterMetrics().
					Return([]*m.CounterMetric{
						{Name: "test", Value: int64(123)},
					}, nil)
			},
			expected: expected{
				contentType: "text/html",
				status:      http.StatusOK,
				body:        "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, testIter(ts, tc))
	}
}
