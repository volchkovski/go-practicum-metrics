package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"

	"github.com/go-chi/chi/v5"
)

var (
	ErrInvalidType    = fmt.Errorf("allowed metric types: %s, %s", GaugeType, CounterType)
	ErrMetricNotFound = errors.New("metric is not found")
)

var (
	AllowedMetricTypesMsg = fmt.Sprintf("Allowed metric types: %s, %s", GaugeType, CounterType)
	CanceledReqMsg        = "Request is canceled"
)

func CollectMetricHandler(s MetricPusher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")
		val := chi.URLParam(r, "val")

		ctx := r.Context()
		errs := make(chan error, 1)

		go func() {
			errs <- collectMetric(ctx, s, tp, nm, val)
			close(errs)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case err := <-errs:
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func collectMetric(ctx context.Context, s MetricPusher, tp, nm, val string) error {
	switch MetricType(tp) {
	case GaugeType:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("value for gauge type must be float: %w", err)
		}
		if err = s.PushGaugeMetric(ctx, &m.GaugeMetric{Name: nm, Value: v}); err != nil {
			return fmt.Errorf("failed to push gauge metric: %w", err)
		}

	case CounterType:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("value for counter must be integer: %w", err)
		}
		if err = s.PushCounterMetric(ctx, &m.CounterMetric{Name: nm, Value: v}); err != nil {
			return fmt.Errorf("failed to push counter metric: %w", err)
		}
	default:
		return ErrInvalidType
	}
	return nil
}

func MetricHandler(s MetricGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")

		type result struct {
			mvalue string
			err    error
		}

		ctx := r.Context()
		resultChan := make(chan result, 1)

		go func() {
			mvalue, err := metricValue(ctx, s, tp, nm)
			resultChan <- result{mvalue: mvalue, err: err}
			close(resultChan)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case res := <-resultChan:
			if res.err != nil {
				if errors.Is(res.err, ErrMetricNotFound) {
					http.Error(w, res.err.Error(), http.StatusNotFound)
					return
				}
				http.Error(w, AllowedMetricTypesMsg, http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			if _, err := fmt.Fprint(w, res.mvalue); err != nil {
				http.Error(w, "Failed to write body: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func metricValue(ctx context.Context, s MetricGetter, tp, nm string) (string, error) {
	switch MetricType(tp) {
	case GaugeType:
		m, err := s.GetGaugeMetric(ctx, nm)
		if err != nil {
			return "", ErrMetricNotFound
		}
		return strconv.FormatFloat(m.Value, 'f', -1, 64), nil
	case CounterType:
		m, err := s.GetCounterMetric(ctx, nm)
		if err != nil {
			return "", ErrMetricNotFound
		}
		return strconv.FormatInt(m.Value, 10), nil
	default:
		return "", ErrInvalidType
	}
}

func AllMetricsHandler(s AllMetricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type result struct {
			metrics []Metric
			err     error
		}

		ctx := r.Context()
		resultChan := make(chan result, 1)

		go func() {
			metrics, err := allMetrics(ctx, s)
			resultChan <- result{metrics: metrics, err: err}
			close(resultChan)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case res := <-resultChan:
			if res.err != nil {
				http.Error(w, res.err.Error(), http.StatusInternalServerError)
				return
			}
			t, err := template.New("AllMetrics").Parse(HTMLAllMetrics)
			if err != nil {
				msg := fmt.Sprintf("Failed to parse html template: %s", err.Error())
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			err = t.Execute(w, res.metrics)
			if err != nil {
				msg := fmt.Sprintf("Failed to put metrics into html template: %s", err.Error())
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
		}
	}
}

func allMetrics(ctx context.Context, s AllMetricsGetter) ([]Metric, error) {
	metrics := make([]Metric, 0, 50)

	gaugeMetrics, err := s.GetAllGaugeMetrics(ctx)
	if err != nil {
		return nil, err
	}
	for _, gm := range gaugeMetrics {
		metric := Metric{gm.Name, strconv.FormatFloat(gm.Value, 'f', 2, 64)}
		metrics = append(metrics, metric)
	}

	counterMetrics, err := s.GetAllCounterMetrics(ctx)
	if err != nil {
		return nil, err
	}
	for _, cm := range counterMetrics {
		metric := Metric{cm.Name, strconv.FormatInt(cm.Value, 10)}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func CollectMetricHandlerJSON(s MetricPusher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric m.Metrics

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		errs := make(chan error, 1)

		go func() {
			errs <- collectMetricJSON(ctx, s, metric)
			close(errs)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case err := <-errs:
			if err != nil {
				if errors.Is(err, ErrInvalidType) {
					http.Error(w, AllowedMetricTypesMsg, http.StatusBadRequest)
					return
				}
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
}

func collectMetricJSON(ctx context.Context, s MetricPusher, metric m.Metrics) error {
	switch MetricType(metric.MType) {
	case GaugeType:
		if err := s.PushGaugeMetric(ctx, &m.GaugeMetric{Name: metric.ID, Value: *metric.Value}); err != nil {
			return fmt.Errorf("failed to push gauge metric: %w", err)
		}
	case CounterType:
		if err := s.PushCounterMetric(ctx, &m.CounterMetric{Name: metric.ID, Value: *metric.Delta}); err != nil {
			return fmt.Errorf("failed to push counter metric: %w", err)
		}
	default:
		return ErrInvalidType
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

		type result struct {
			metric *m.Metrics
			err    error
		}

		ctx := r.Context()
		resultChan := make(chan result, 1)

		go func() {
			metric, err := metricJSON(ctx, s, metric)
			resultChan <- result{metric: metric, err: err}
			close(resultChan)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case res := <-resultChan:
			if res.err != nil {
				if errors.Is(res.err, ErrMetricNotFound) {
					http.Error(w, res.err.Error(), http.StatusNotFound)
					return
				}
				if errors.Is(res.err, ErrInvalidType) {
					http.Error(w, res.err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(res.metric); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func metricJSON(ctx context.Context, s MetricGetter, metric *m.Metrics) (*m.Metrics, error) {
	switch MetricType(metric.MType) {
	case GaugeType:
		gm, err := s.GetGaugeMetric(ctx, metric.ID)
		if err != nil {
			return nil, ErrMetricNotFound
		}
		metric.Value = &gm.Value
	case CounterType:
		cm, err := s.GetCounterMetric(ctx, metric.ID)
		if err != nil {
			return nil, ErrMetricNotFound
		}
		metric.Delta = &cm.Value
	default:
		return nil, ErrInvalidType
	}
	return metric, nil
}

func PingDB(s DBPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		errs := make(chan error, 1)

		go func() {
			errs <- s.PingDB(ctx)
			close(errs)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case err := <-errs:
			if err != nil {
				logger.Log.Infof("PingDB error: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
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
		ctx := r.Context()
		errs := make(chan error, 1)

		go func() {
			errs <- collectMetrics(ctx, s, metrics)
			close(errs)
		}()

		select {
		case <-ctx.Done():
			http.Error(w, CanceledReqMsg, http.StatusRequestTimeout)
			return
		case err := <-errs:
			if err != nil {
				if errors.Is(err, ErrInvalidType) {
					http.Error(w, AllowedMetricTypesMsg, http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
}

func collectMetrics(ctx context.Context, s MetricsPusher, metrics []m.Metrics) error {
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
			return ErrInvalidType
		}
	}
	if err := s.PushMetrics(ctx, gauges, counters); err != nil {
		return fmt.Errorf("failed to push metrics: %s", err.Error())
	}
	return nil
}
