package handlers

import (
	"fmt"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func splitURLPath(u *url.URL) []string {
	p := strings.Trim(u.Path, "/")
	return strings.Split(p, "/")
}

func CollectMetricHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Allowed only Post method", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		// example: update/gauge/someMetric/527
		parts := splitURLPath(r.URL)
		if len(parts) < 3 {
			http.Error(w, "Metric name is empty", http.StatusNotFound)
			return
		}
		if len(parts) < 4 {
			http.Error(w, "Metric value is empty", http.StatusBadRequest)
			return
		}
		t, nm, val := parts[1], parts[2], parts[3]
		switch t {
		case storage.Gauge:
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				http.Error(w, "Value for gauge type must be float", http.StatusBadRequest)
				return
			}
			s.WriteGauge(nm, v)
		case storage.Counter:
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				http.Error(w, "Value for counter type must be integer", http.StatusBadRequest)
				return
			}
			s.WriteCounter(nm, v)
		default:
			msg := fmt.Sprintf("Allowed metric types: %s, %s", storage.Gauge, storage.Counter)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
