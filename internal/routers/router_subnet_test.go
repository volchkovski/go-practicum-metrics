package routers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRouterWithTrustedSubnet(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	service := NewMockmetricsProcessor(mockCtl)

	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		path           string
		method         string
		body           string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "Empty trusted subnet allows request",
			trustedSubnet: "",
			xRealIP:       "192.168.1.1",
			path:          "/update",
			method:        http.MethodPost,
			body:          `{"id": "test", "type": "gauge", "value": 123.0}`,
			mockSetup: func() {
				service.EXPECT().PushGaugeMetric(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "IP in trusted subnet allows request",
			trustedSubnet: "192.168.1.0/24",
			xRealIP:       "192.168.1.100",
			path:          "/update",
			method:        http.MethodPost,
			body:          `{"id": "test", "type": "gauge", "value": 123.0}`,
			mockSetup: func() {
				service.EXPECT().PushGaugeMetric(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			path:           "/update",
			method:         http.MethodPost,
			body:           `{"id": "test", "type": "gauge", "value": 123.0}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing X-Real-IP header returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			path:           "/update",
			method:         http.MethodPost,
			body:           `{"id": "test", "type": "gauge", "value": 123.0}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:          "Batch updates with IP in trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			xRealIP:       "192.168.1.50",
			path:          "/updates/",
			method:        http.MethodPost,
			body:          `[{"id": "test1", "type": "gauge", "value": 1.0}]`,
			mockSetup: func() {
				service.EXPECT().PushMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Batch updates with IP not in trusted subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			path:           "/updates/",
			method:         http.MethodPost,
			body:           `[{"id": "test1", "type": "gauge", "value": 1.0}]`,
			mockSetup:      func() {},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			r := NewMetricRouter(secretKey, nil, tt.trustedSubnet, service)
			ts := httptest.NewServer(r)
			defer ts.Close()

			headers := make(http.Header)
			headers.Add("Content-Type", "application/json")
			if tt.xRealIP != "" {
				headers.Add("X-Real-IP", tt.xRealIP)
			}

			resp, _ := testRequest(t, ts, tt.method, tt.path, strings.NewReader(tt.body), headers)
			defer func() {
				require.NoError(t, resp.Body.Close())
			}()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
