package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
)

// MockMetricService implements all the required interfaces for testing
type MockMetricService struct {
	mock.Mock
}

func (m *MockMetricService) PushGaugeMetric(ctx context.Context, metric *models.GaugeMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricService) PushCounterMetric(ctx context.Context, metric *models.CounterMetric) error {
	args := m.Called(ctx, metric)
	return args.Error(0)
}

func (m *MockMetricService) GetGaugeMetric(ctx context.Context, name string) (*models.GaugeMetric, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GaugeMetric), args.Error(1)
}

func (m *MockMetricService) GetCounterMetric(ctx context.Context, name string) (*models.CounterMetric, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CounterMetric), args.Error(1)
}

func (m *MockMetricService) GetAllGaugeMetrics(ctx context.Context) ([]*models.GaugeMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.GaugeMetric), args.Error(1)
}

func (m *MockMetricService) GetAllCounterMetrics(ctx context.Context) ([]*models.CounterMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CounterMetric), args.Error(1)
}

func (m *MockMetricService) PingDB(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMetricService) PushMetrics(ctx context.Context, gauges []*models.GaugeMetric, counters []*models.CounterMetric) error {
	args := m.Called(ctx, gauges, counters)
	return args.Error(0)
}

func TestCollectMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValue    string
		setupMock      func(*MockMetricService)
		expectedStatus int
	}{
		{
			name:        "valid gauge metric",
			metricType:  "gauge",
			metricName:  "cpu_usage",
			metricValue: "75.5",
			setupMock: func(m *MockMetricService) {
				expectedMetric := &models.GaugeMetric{Name: "cpu_usage", Value: 75.5}
				m.On("PushGaugeMetric", mock.Anything, expectedMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "valid counter metric",
			metricType:  "counter",
			metricName:  "requests",
			metricValue: "42",
			setupMock: func(m *MockMetricService) {
				expectedMetric := &models.CounterMetric{Name: "requests", Value: 42}
				m.On("PushCounterMetric", mock.Anything, expectedMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid metric type",
			metricType:     "invalid",
			metricName:     "test",
			metricValue:    "123",
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid gauge value",
			metricType:     "gauge",
			metricName:     "test",
			metricValue:    "not_a_number",
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid counter value",
			metricType:     "counter",
			metricName:     "test",
			metricValue:    "not_a_number",
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error for gauge",
			metricType:  "gauge",
			metricName:  "test",
			metricValue: "123.45",
			setupMock: func(m *MockMetricService) {
				m.On("PushGaugeMetric", mock.Anything, mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error for counter",
			metricType:  "counter",
			metricName:  "test",
			metricValue: "123",
			setupMock: func(m *MockMetricService) {
				m.On("PushCounterMetric", mock.Anything, mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := CollectMetricHandler(mockService)

			// Create router to handle URL parameters
			r := chi.NewRouter()
			r.Post("/update/{tp}/{nm}/{val}", handler)

			url := fmt.Sprintf("/update/%s/%s/%s", tc.metricType, tc.metricName, tc.metricValue)
			req := httptest.NewRequest(http.MethodPost, url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		setupMock      func(*MockMetricService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "existing gauge metric",
			metricType: "gauge",
			metricName: "cpu_usage",
			setupMock: func(m *MockMetricService) {
				metric := &models.GaugeMetric{Name: "cpu_usage", Value: 75.5}
				m.On("GetGaugeMetric", mock.Anything, "cpu_usage").Return(metric, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "75.5",
		},
		{
			name:       "existing counter metric",
			metricType: "counter",
			metricName: "requests",
			setupMock: func(m *MockMetricService) {
				metric := &models.CounterMetric{Name: "requests", Value: 42}
				m.On("GetCounterMetric", mock.Anything, "requests").Return(metric, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:       "non-existing gauge metric",
			metricType: "gauge",
			metricName: "missing",
			setupMock: func(m *MockMetricService) {
				m.On("GetGaugeMetric", mock.Anything, "missing").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "non-existing counter metric",
			metricType: "counter",
			metricName: "missing",
			setupMock: func(m *MockMetricService) {
				m.On("GetCounterMetric", mock.Anything, "missing").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid metric type",
			metricType:     "invalid",
			metricName:     "test",
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := MetricHandler(mockService)

			// Create router to handle URL parameters
			r := chi.NewRouter()
			r.Get("/value/{tp}/{nm}", handler)

			url := fmt.Sprintf("/value/%s/%s", tc.metricType, tc.metricName)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, strings.TrimSpace(rr.Body.String()))
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestCollectMetricHandlerJSON(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockMetricService)
		expectedStatus int
	}{
		{
			name:        "valid gauge metric JSON",
			requestBody: `{"id":"cpu_usage","type":"gauge","value":75.5}`,
			setupMock: func(m *MockMetricService) {
				expectedMetric := &models.GaugeMetric{Name: "cpu_usage", Value: 75.5}
				m.On("PushGaugeMetric", mock.Anything, expectedMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "valid counter metric JSON",
			requestBody: `{"id":"requests","type":"counter","delta":42}`,
			setupMock: func(m *MockMetricService) {
				expectedMetric := &models.CounterMetric{Name: "requests", Value: 42}
				m.On("PushCounterMetric", mock.Anything, expectedMetric).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid":json}`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid metric type in JSON",
			requestBody:    `{"id":"test","type":"invalid","value":123}`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			requestBody: `{"id":"test","type":"gauge","value":123.45}`,
			setupMock: func(m *MockMetricService) {
				m.On("PushGaugeMetric", mock.Anything, mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := CollectMetricHandlerJSON(mockService)

			req := httptest.NewRequest(http.MethodPost, "/update/", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestMetricHandlerJSON(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockMetricService)
		expectedStatus int
		checkResponse  func(*testing.T, string)
	}{
		{
			name:        "existing gauge metric",
			requestBody: `{"id":"cpu_usage","type":"gauge"}`,
			setupMock: func(m *MockMetricService) {
				metric := &models.GaugeMetric{Name: "cpu_usage", Value: 75.5}
				m.On("GetGaugeMetric", mock.Anything, "cpu_usage").Return(metric, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response models.Metrics
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "cpu_usage", response.ID)
				assert.Equal(t, "gauge", response.MType)
				assert.NotNil(t, response.Value)
				assert.Equal(t, 75.5, *response.Value)
			},
		},
		{
			name:        "existing counter metric",
			requestBody: `{"id":"requests","type":"counter"}`,
			setupMock: func(m *MockMetricService) {
				metric := &models.CounterMetric{Name: "requests", Value: 42}
				m.On("GetCounterMetric", mock.Anything, "requests").Return(metric, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response models.Metrics
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "requests", response.ID)
				assert.Equal(t, "counter", response.MType)
				assert.NotNil(t, response.Delta)
				assert.Equal(t, int64(42), *response.Delta)
			},
		},
		{
			name:        "non-existing metric",
			requestBody: `{"id":"missing","type":"gauge"}`,
			setupMock: func(m *MockMetricService) {
				m.On("GetGaugeMetric", mock.Anything, "missing").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid":json}`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid metric type",
			requestBody:    `{"id":"test","type":"invalid"}`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := MetricHandlerJSON(mockService)

			req := httptest.NewRequest(http.MethodPost, "/value/", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.checkResponse != nil {
				tc.checkResponse(t, rr.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPingDB(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMetricService)
		expectedStatus int
	}{
		{
			name: "successful ping",
			setupMock: func(m *MockMetricService) {
				m.On("PingDB", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "failed ping",
			setupMock: func(m *MockMetricService) {
				m.On("PingDB", mock.Anything).Return(errors.New("connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := PingDB(mockService)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestCollectMetricsHandlerJSON(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*MockMetricService)
		expectedStatus int
	}{
		{
			name: "valid multiple metrics",
			requestBody: `[
				{"id":"cpu_usage","type":"gauge","value":75.5},
				{"id":"requests","type":"counter","delta":42}
			]`,
			setupMock: func(m *MockMetricService) {
				expectedGauges := []*models.GaugeMetric{
					{Name: "cpu_usage", Value: 75.5},
				}
				expectedCounters := []*models.CounterMetric{
					{Name: "requests", Value: 42},
				}
				m.On("PushMetrics", mock.Anything, expectedGauges, expectedCounters).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    `[{"invalid":json}]`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid metric type",
			requestBody:    `[{"id":"test","type":"invalid","value":123}]`,
			setupMock:      func(m *MockMetricService) {}, // No calls expected
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: `[
				{"id":"test","type":"gauge","value":123.45}
			]`,
			setupMock: func(m *MockMetricService) {
				m.On("PushMetrics", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := CollectMetricsHandlerJSON(mockService)

			req := httptest.NewRequest(http.MethodPost, "/updates/", strings.NewReader(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAllMetricsHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMetricService)
		expectedStatus int
	}{
		{
			name: "successful metrics retrieval",
			setupMock: func(m *MockMetricService) {
				gauges := []*models.GaugeMetric{
					{Name: "cpu_usage", Value: 75.5},
				}
				counters := []*models.CounterMetric{
					{Name: "requests", Value: 42},
				}
				m.On("GetAllGaugeMetrics", mock.Anything).Return(gauges, nil)
				m.On("GetAllCounterMetrics", mock.Anything).Return(counters, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "gauge metrics error",
			setupMock: func(m *MockMetricService) {
				m.On("GetAllGaugeMetrics", mock.Anything).Return(nil, errors.New("gauge error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockMetricService{}
			tc.setupMock(mockService)

			handler := AllMetricsHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			if tc.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestContextCancellation(t *testing.T) {
	mockService := &MockMetricService{}

	mockService.On("PushGaugeMetric", mock.Anything, mock.Anything).Return(nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := CollectMetricHandler(mockService)

	r := chi.NewRouter()
	r.Post("/update/{tp}/{nm}/{val}", handler)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.True(t, rr.Code == http.StatusRequestTimeout || rr.Code == http.StatusOK)
}
