package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (s *testStorage) WriteGauge(nm string, val float64) error {
	s.gauges[nm] = val
	return nil
}

func (s *testStorage) WriteCounter(nm string, val int64) error {
	s.counters[nm] += val
	return nil
}

func TestCollectMetricHandler(t *testing.T) {
	type value struct {
		method string
		target string
	}
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name  string
		value value
		want  want
	}{
		{
			name: "incorrect method",
			value: value{
				method: http.MethodGet,
				target: "http://localhost:8080/update/gauge/SomeMetric/123",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusMethodNotAllowed,
				body:        "Allowed only Post method\n",
			},
		},
		{
			name: "wrong metric type",
			value: value{
				method: http.MethodPost,
				target: "http://localhost:8080/update/gayge/SomeMetric/123",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "Allowed metric types: gauge, counter\n",
			},
		},
		{
			name: "no metric name",
			value: value{
				method: http.MethodPost,
				target: "http://localhost:8080/update/gauge",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "Metric name is empty\n",
			},
		},
		{
			name: "no metric value",
			value: value{
				method: http.MethodPost,
				target: "http://localhost:8080/update/gauge/someMetric",
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "Metric value is empty\n",
			},
		},
		{
			name: "test positive value",
			value: value{
				method: http.MethodPost,
				target: "http://localhost:8080/update/gauge/SomeMetric/123",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				body:        "",
			},
		},
		{
			name: "test negative value",
			value: value{
				method: http.MethodPost,
				target: "http://localhost:8080/update/counter/SomeMetric/-123",
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				body:        "",
			},
		},
	}
	s := &testStorage{
		gauges:   map[string]float64{},
		counters: map[string]int64{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.value.method, tt.value.target, nil)
			w := httptest.NewRecorder()
			h := CollectMetricHandler(s)
			h(w, r)
			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.body, w.Body.String())
		})
	}
}
