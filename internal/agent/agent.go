package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

var runtimeMetrics = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

type Agent struct {
	memStats   *runtime.MemStats
	repIntr    time.Duration
	pollIntr   time.Duration
	serverAddr string
	pollCount  int64
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
		}
	}
	return float64(0), false
}

func (a *Agent) postMetric(metric *m.Metrics) error {
	url := "http://" + a.serverAddr + "/update"
	var buff bytes.Buffer
	if err := json.NewEncoder(&buff).Encode(metric); err != nil {
		return err
	}
	res, err := http.Post(url, "application/json", &buff)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}
	msg, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body %s", err.Error())
	}
	return fmt.Errorf("status: %d description: %s", res.StatusCode, string(msg))
}

func New(cfg *configs.AgentConfig) *Agent {
	return &Agent{
		memStats:   &runtime.MemStats{},
		repIntr:    time.Duration(cfg.ReportIntr) * time.Second,
		pollIntr:   time.Duration(cfg.PollIntr) * time.Second,
		serverAddr: cfg.ServerAddr,
		pollCount:  0,
	}
}

func getRandomFloat() float64 {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Float64()
}
