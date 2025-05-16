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
		client:     NewRestyClient(),
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
	metrics := make([]*m.Metrics, 0, 50)
	for _, metricName := range runtimeMetricNames {
		v, ok := gaugeVal(a.memStats, metricName)
		if !ok {
			log.Printf("Failed to get gauge value for %s", metricName)
			continue
		}
		metric := &m.Metrics{ID: metricName, MType: "gauge", Value: &v}
		metrics = append(metrics, metric)
	}
	rv := getRandomFloat()
	metric := &m.Metrics{ID: "RandomValue", MType: "gauge", Value: &rv}
	metrics = append(metrics, metric)

	a.pollCount += 1
	metric = &m.Metrics{ID: "PollCount", MType: "counter", Delta: &a.pollCount}
	metrics = append(metrics, metric)

	if err := a.postMetrics(metrics); err != nil {
		log.Printf("Failed to post metrics: %s", err.Error())
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

func (a *Agent) postMetrics(metrics []*m.Metrics) error {
	url := "http://" + a.serverAddr + "/updates/"

	p, err := json.Marshal(metrics)
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
