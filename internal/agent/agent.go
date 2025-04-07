package agent

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
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
	pollCount  int
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
	val := reflect.ValueOf(*a.memStats)
	for _, m := range runtimeMetrics {
		if err := a.postMetric("gauge", m, val.FieldByName(m)); err != nil {
			log.Println(err)
		}
	}
	if err := a.postMetric("gauge", "RandomValue", getRandomInt()); err != nil {
		log.Printf("Failed to post RandomValue %s\n", err.Error())
	}
	a.pollCount += 1
	if err := a.postMetric("counter", "PollCount", a.pollCount); err != nil {
		log.Printf("Failed to post PollCount %s\n", err.Error())
	}
}

func (a *Agent) postMetric(tp, nm string, val any) error {
	urlTemplate := "http://" + a.serverAddr + "/update/%s/%s/%v"
	u := fmt.Sprintf(urlTemplate, tp, nm, val)
	res, err := http.Post(u, "text/plain", nil)
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

func getRandomInt() int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Int()
}
