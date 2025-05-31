package routers

import (
	"bytes"
	"errors"
	"github.com/volchkovski/go-practicum-metrics/internal/hasher"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	headers  http.Header
	body     string
	mock     func()
	expected expected
}

const secretKey = "test"

func headersWithHash(h http.Header, b []byte) http.Header {
	if h == nil {
		h = make(http.Header)
	}
	hshr := hasher.New(secretKey)
	hash := hshr.Hash(b)
	h.Set(hasher.HashHeaderKey, hash)
	return h
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, headers http.Header) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	b, err := io.ReadAll(body)
	require.NoError(t, err)
	req.Body = io.NopCloser(bytes.NewBuffer(b))

	headers = headersWithHash(headers, b)
	req.Header = headers

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func testIter(ts *httptest.Server, tc test) func(*testing.T) {
	return func(t *testing.T) {
		tc.mock()
		resp, body := testRequest(t, ts, tc.method, tc.path, strings.NewReader(tc.body), tc.headers)
		defer func() {
			require.NoError(t, resp.Body.Close())
		}()
		assert.Equal(t, tc.expected.status, resp.StatusCode)
		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, tc.expected.contentType)
		if tc.expected.body != "" && strings.Contains(contentType, "application/json") {
			assert.JSONEq(t, tc.expected.body, body)
		} else if tc.expected.body != "" {
			assert.Equal(t, tc.expected.body, body)
		}
	}
}

func TestRouterUpdateMetric(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(secretKey, service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "update valid gauge",
			path:   "/update/gauge/test/123",
			method: http.MethodPost,
			mock: func() {
				service.EXPECT().
					PushGaugeMetric(gomock.Any(), &m.GaugeMetric{Name: "test", Value: float64(123)}).
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
					PushCounterMetric(gomock.Any(), &m.CounterMetric{Name: "test", Value: int64(123)}).
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
	r := NewMetricRouter(secretKey, service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "get valid gauge",
			path:   "/value/gauge/test",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetGaugeMetric(gomock.Any(), "test").
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
				service.EXPECT().GetCounterMetric(gomock.Any(), "test").
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
				service.EXPECT().GetGaugeMetric(gomock.Any(), "test").
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
				service.EXPECT().GetCounterMetric(gomock.Any(), "test").
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
	r := NewMetricRouter(secretKey, service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []test{
		{
			name:   "get valid all metrics",
			path:   "",
			method: http.MethodGet,
			mock: func() {
				service.EXPECT().GetAllGaugeMetrics(gomock.Any()).
					Return([]*m.GaugeMetric{
						{Name: "test", Value: float64(123)},
					}, nil)
				service.EXPECT().GetAllCounterMetrics(gomock.Any()).
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

func TestRouterMetricJSON(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(secretKey, service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	tests := []test{
		{
			name:    "get valid gauge metric",
			path:    "/value",
			method:  http.MethodPost,
			body:    `{"id": "test", "type": "gauge"}`,
			headers: headers,
			mock: func() {
				service.EXPECT().GetGaugeMetric(gomock.Any(), "test").
					Return(&m.GaugeMetric{Name: "test", Value: float64(1.1)}, nil)
			},
			expected: expected{
				contentType: "application/json",
				status:      http.StatusOK,
				body:        `{"id": "test", "type": "gauge", "value": 1.1}`,
			},
		},
		{
			name:    "get counter metric",
			path:    "/value",
			method:  http.MethodPost,
			body:    `{"id": "test", "type": "counter"}`,
			headers: headers,
			mock: func() {
				service.EXPECT().GetCounterMetric(gomock.Any(), "test").
					Return(&m.CounterMetric{Name: "test", Value: int64(1)}, nil)
			},
			expected: expected{
				contentType: "application/json",
				status:      http.StatusOK,
				body:        `{"id": "test", "type": "counter", "delta": 1}`,
			},
		},
		{
			name:    "update gauge metric",
			path:    "/update",
			method:  http.MethodPost,
			body:    `{"id": "test", "type": "gauge", "value": 123.0}`,
			headers: headers,
			mock: func() {
				service.EXPECT().
					PushGaugeMetric(gomock.Any(), &m.GaugeMetric{Name: "test", Value: float64(123)}).
					Return(nil)
			},
			expected: expected{
				contentType: "application/json",
				status:      http.StatusOK,
				body:        `{"id": "test", "type": "gauge", "value": 123.0}`,
			},
		},
		{
			name:    "update counter metric",
			path:    "/update",
			method:  http.MethodPost,
			body:    `{"id": "test", "type": "counter", "delta": 123}`,
			headers: headers,
			mock: func() {
				service.EXPECT().
					PushCounterMetric(gomock.Any(), &m.CounterMetric{Name: "test", Value: int64(123)}).
					Return(nil)
			},
			expected: expected{
				contentType: "application/json",
				status:      http.StatusOK,
				body:        `{"id": "test", "type": "counter", "delta": 123}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, testIter(ts, tc))
	}
}

func TestRouterMetricsJSON(t *testing.T) {
	mockCtl := gomock.NewController(t)

	service := NewMockmetricsProcessor(mockCtl)
	r := NewMetricRouter(secretKey, service)
	ts := httptest.NewServer(r)
	defer ts.Close()

	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	tests := []test{
		{
			name:   "post gauge and counter metrics",
			path:   "/updates/",
			method: http.MethodPost,
			body: `[
				{"id": "testGauge1", "type": "gauge", "value": 1.0},
				{"id": "testCounter1", "type": "counter", "delta": 1}
			]`,
			headers: headers,
			mock: func() {
				service.EXPECT().PushMetrics(
					gomock.Any(),
					[]*m.GaugeMetric{
						{Name: "testGauge1", Value: 1.0},
					},
					[]*m.CounterMetric{
						{Name: "testCounter1", Value: 1},
					},
				).Return(nil)
			},
			expected: expected{
				contentType: "",
				status:      http.StatusOK,
				body:        "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, testIter(ts, tc))
	}
}
