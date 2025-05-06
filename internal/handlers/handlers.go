package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"

	"github.com/go-chi/chi/v5"
)

func CollectMetricHandler(s MetricPusher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")
		val := chi.URLParam(r, "val")

		if err := collectMetric(s, tp, nm, val); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func collectMetric(s MetricPusher, tp, nm, val string) error {
	switch MetricType(tp) {
	case GaugeType:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("value for gauge type must be float: %w", err)
		}
		if err = s.PushGaugeMetric(&m.GaugeMetric{Name: nm, Value: v}); err != nil {
			return fmt.Errorf("failed to push gauge metric: %w", err)
		}

	case CounterType:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("value for counter must be integer: %w", err)
		}
		if err = s.PushCounterMetric(&m.CounterMetric{Name: nm, Value: v}); err != nil {
			return fmt.Errorf("failed to push counter metric: %w", err)
		}
	default:
		return fmt.Errorf("allowed metric types: %s, %s", GaugeType, CounterType)
	}
	return nil
}

func MetricHandler(s MetricGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")
		var metricVal string
		switch MetricType(tp) {
		case GaugeType:
			m, err := s.GetGaugeMetric(nm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			metricVal = strconv.FormatFloat(m.Value, 'f', -1, 64)
		case CounterType:
			m, err := s.GetCounterMetric(nm)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			metricVal = strconv.FormatInt(m.Value, 10)
		default:
			msg := fmt.Sprintf("Allowed metric types: %s, %s", GaugeType, CounterType)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, metricVal); err != nil {
			http.Error(w, "Failed to write body: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func AllMetricsHandler(s AllMetricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type metric struct {
			Name  string
			Value string
		}
		var metrics []metric

		gaugeMetrics, err := s.GetAllGaugeMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, gm := range gaugeMetrics {
			metrics = append(metrics, metric{gm.Name, strconv.FormatFloat(gm.Value, 'f', 2, 64)})
		}

		counterMetrics, err := s.GetAllCounterMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, cm := range counterMetrics {
			metrics = append(metrics, metric{cm.Name, strconv.FormatInt(cm.Value, 10)})
		}

		t, err := template.New("AllMetrics").Parse(HTMLAllMetrics)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse html template: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		err = t.Execute(w, metrics)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to put metrics into html template: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func CollectMetricHandlerJSON(s MetricPusher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric m.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := collectMetricJSON(s, metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func collectMetricJSON(s MetricPusher, metric m.Metrics) error {
	switch MetricType(metric.MType) {
	case GaugeType:
		if err := s.PushGaugeMetric(&m.GaugeMetric{Name: metric.ID, Value: *metric.Value}); err != nil {
			return fmt.Errorf("failed to push gauge metric: %w", err)
		}
	case CounterType:
		if err := s.PushCounterMetric(&m.CounterMetric{Name: metric.ID, Value: *metric.Delta}); err != nil {
			return fmt.Errorf("failed to push counter metric: %w", err)
		}
	default:
		return fmt.Errorf("allowed metric types: %s, %s", GaugeType, CounterType)
	}
	return nil
}

func MetricHandlerJSON(s MetricGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metric := new(m.Metrics)

		if err := json.NewDecoder(r.Body).Decode(metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metric, err := metricJSON(s, metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func metricJSON(s MetricGetter, metric *m.Metrics) (*m.Metrics, error) {
	switch MetricType(metric.MType) {
	case GaugeType:
		gm, err := s.GetGaugeMetric(metric.ID)
		if err != nil {
			return nil, err
		}
		metric.Value = &gm.Value
	case CounterType:
		cm, err := s.GetCounterMetric(metric.ID)
		if err != nil {
			return nil, err
		}
		metric.Delta = &cm.Value
	default:
		return nil, fmt.Errorf("invalid metric type. Allowed: %s, %s", GaugeType, CounterType)
	}
	return metric, nil
}

func PingDB(s DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.PingDB(); err != nil {
			logger.Log.Infof("PingDB error: %s", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func CollectMetricsHandlerJSON(s MetricsPusher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metrics []m.Metrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			logger.Log.Errorf("collectMetricsHandlerJson - failed to decode request body: %s", err.Error())
			http.Error(w, "Failed to decode request body", http.StatusInternalServerError)
			return
		}
		gauges := make([]*m.GaugeMetric, 0, 50)
		counters := make([]*m.CounterMetric, 0, 10)
		for _, metric := range metrics {
			switch MetricType(metric.MType) {
			case GaugeType:
				gauge := m.GaugeMetric{
					Name:  metric.ID,
					Value: *metric.Value,
				}
				gauges = append(gauges, &gauge)
			case CounterType:
				counter := m.CounterMetric{
					Name:  metric.ID,
					Value: *metric.Delta,
				}
				counters = append(counters, &counter)
			default:
				msg := fmt.Sprintf("Invalid metric type %s. Allowed metric types: %s, %s", metric.MType, GaugeType, CounterType)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
		}
		if err := s.PushMetrics(gauges, counters); err != nil {
			logger.Log.Errorf("Failed to push metrics: %s", err.Error())
			http.Error(w, "Error during metrics pushing", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
