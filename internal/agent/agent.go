package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

var headers = map[string]string{"Content-Encoding": "gzip", "Content-Type": "application/json"}

type Agent struct {
	memStats   *runtime.MemStats
	repIntr    time.Duration
	pollIntr   time.Duration
	serverAddr string
	pollCount  int64
	client     *resty.Client
}

func New(cfg *configs.AgentConfig) *Agent {
	return &Agent{
		memStats:   &runtime.MemStats{},
		repIntr:    time.Duration(cfg.ReportIntr) * time.Second,
		pollIntr:   time.Duration(cfg.PollIntr) * time.Second,
		serverAddr: cfg.ServerAddr,
		pollCount:  0,
		client:     resty.New().SetHeaders(headers),
	}
}

func (a *Agent) Run() {
	go func() {
		for {
			runtime.ReadMemStats(a.memStats)
			time.Sleep(a.pollIntr)
		}
	}()
	for {
		time.Sleep(a.repIntr)
		a.collectMetrics()
	}
}

func (a *Agent) collectMetrics() {
	for _, metric := range runtimeMetrics {
		v, ok := gaugeVal(a.memStats, metric)
		if !ok {
			log.Printf("Failed to get gauge value for %s", metric)
			continue
		}
		if err := a.postMetric(&m.Metrics{ID: metric, MType: "gauge", Value: &v}); err != nil {
			log.Println(err)
		}
	}
	rv := getRandomFloat()
	if err := a.postMetric(&m.Metrics{ID: "RandomValue", MType: "gauge", Value: &rv}); err != nil {
		log.Printf("Failed to post RandomValue %s\n", err.Error())
	}
	a.pollCount += 1
	if err := a.postMetric(&m.Metrics{ID: "PollCount", MType: "counter", Delta: &a.pollCount}); err != nil {
		log.Printf("Failed to post PollCount %s\n", err.Error())
	}
}

func gaugeVal(stat *runtime.MemStats, fname string) (float64, bool) {
	field := reflect.ValueOf(*stat).FieldByName(fname)
	if field.IsValid() {
		switch field.Kind() {
		case reflect.Uint32, reflect.Uint64:
			return float64(field.Uint()), true
		case reflect.Float64:
			return field.Float(), true
		default:
			return float64(0), false
		}
	}
	return float64(0), false
}

func (a *Agent) postMetric(metric *m.Metrics) error {
	url := "http://" + a.serverAddr + "/update"

	p, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	var buff bytes.Buffer

	cw, err := gzip.NewWriterLevel(&buff, gzip.BestSpeed)
	if err != nil {
		return err
	}
	if _, err = cw.Write(p); err != nil {
		return err
	}

	if err = cw.Close(); err != nil {
		return err
	}

	res, err := a.client.R().SetBody(&buff).Post(url)
	if err != nil {
		return err
	}
	statusCode := res.StatusCode()
	if statusCode == http.StatusOK {
		return nil
	}
	body := res.Body()
	return fmt.Errorf("got bad response status: %d body: %s", statusCode, string(body))
}

func getRandomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Float64()
}
