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
	memStats  *runtime.MemStats
	pollCount int
}

func (a *Agent) Run() error {
	go func() {
		for {
			runtime.ReadMemStats(a.memStats)
			time.Sleep(2 * time.Second)
		}
	}()
	for {
		time.Sleep(10 * time.Second)
		a.collectMetrics()
	}
	return nil
}

func (a *Agent) collectMetrics() {
	val := reflect.ValueOf(*a.memStats)
	for _, m := range runtimeMetrics {
		if err := postMetric("gauge", m, val.FieldByName(m)); err != nil {
			log.Println(err)
		}
	}
	if err := postMetric("gauge", "RandomValue", getRandomInt()); err != nil {
		log.Printf("Failed to post RandomValue %s\n", err.Error())
	}
	a.pollCount += 1
	if err := postMetric("counter", "PollCount", a.pollCount); err != nil {
		log.Printf("Failed to post PollCount %s\n", err.Error())
	}
}

func New() *Agent {
	return &Agent{
		memStats:  &runtime.MemStats{},
		pollCount: 0,
	}
}

func getRandomInt() int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	return r.Int()
}

func postMetric(tp, nm string, val any) error {
	urlTemplate := `http://localhost:8080/update/%s/%s/%v`
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
