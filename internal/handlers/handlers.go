package handlers

import (
	"fmt"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
	"net/http"
	"regexp"
	"strconv"
)

func CollectMetricHandler(s storage.Storage) http.Handler {
	re := regexp.MustCompile(`/?update/([a-zA-Z]+)/([a-zA-Z]+)/(\d+(\.\d+)?)`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Allowed only Post method", http.StatusMethodNotAllowed)
			return
		}
		result := re.FindAllStringSubmatch(r.URL.Path, -1)
		if result == nil {
			http.Error(w, "Invalid path to push metric", http.StatusBadRequest)
			return
		}
		// [["/update/adfa/someMetric/527.00" "adfa" "someMetric" "527.00" ".00"]]
		matches := result[0]
		if matches[2] == "" {
			http.NotFound(w, r)
			return
		}
		switch matches[1] {
		case storage.Gauge:
			v, err := strconv.ParseFloat(matches[3], 64)
			if err != nil {
				http.Error(w, "Value for gauge type must be float", http.StatusBadRequest)
				return
			}
			s.WriteGauge(matches[2], v)
		case storage.Counter:
			v, err := strconv.ParseInt(matches[3], 10, 64)
			if err != nil {
				http.Error(w, "Value for counter type must be integer", http.StatusBadRequest)
				return
			}
			s.WriteCounter(matches[2], v)
		default:
			msg := fmt.Sprintf("Allow metric types: %s, %s", storage.Gauge, storage.Counter)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
