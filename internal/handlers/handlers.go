package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func CollectMetricHandler(s MetricsWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")
		val := chi.URLParam(r, "val")

		switch MetricType(tp) {
		case GaugeType:
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				http.Error(w, "Value for gauge type must be float", http.StatusBadRequest)
				return
			}
			s.WriteGauge(nm, v)
		case CounterType:
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				http.Error(w, "Value for counter type must be integer", http.StatusBadRequest)
				return
			}
			s.WriteCounter(nm, v)
		default:
			msg := fmt.Sprintf("Allowed metric types: %s, %s", GaugeType, CounterType)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func MetricHandler(s MetricsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tp := chi.URLParam(r, "tp")
		nm := chi.URLParam(r, "nm")
		var metricVal string
		switch MetricType(tp) {
		case GaugeType:
			m, err := s.ReadGauge(nm)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			metricVal = strconv.FormatFloat(m, 'f', -1, 64)
		case CounterType:
			m, err := s.ReadCounter(nm)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			metricVal = strconv.FormatInt(m, 10)
		default:
			msg := fmt.Sprintf("Allowed metric types: %s, %s", GaugeType, CounterType)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		fmt.Fprint(w, metricVal)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}
}

func AllMetricsHandler(s AllMetricsReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Metric struct {
			Name  string
			Value string
		}
		var metrics []Metric
		for nm, val := range s.ReadAllGauges() {
			metrics = append(metrics, Metric{nm, strconv.FormatFloat(val, 'f', 2, 64)})
		}
		for nm, val := range s.ReadAllCounters() {
			metrics = append(metrics, Metric{nm, strconv.FormatInt(val, 10)})
		}
		t, err := template.New("AllMetrics").Parse(HTMLAllMetrics)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, metrics)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
	}
}
