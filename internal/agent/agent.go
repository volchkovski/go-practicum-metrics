package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/volchkovski/go-practicum-metrics/internal/hasher"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

type Agent struct {
	mstorage   *MetricsStorage
	repIntr    time.Duration
	pollIntr   time.Duration
	serverAddr string
	pollCount  atomic.Int64
	client     *resty.Client
	key        string
	rateLimit  int
}

func New(cfg *configs.AgentConfig) *Agent {
	return &Agent{
		mstorage:   NewMetricsStorage(),
		repIntr:    time.Duration(cfg.ReportIntr) * time.Second,
		pollIntr:   time.Duration(cfg.PollIntr) * time.Second,
		serverAddr: cfg.ServerAddr,
		pollCount:  atomic.Int64{},
		client:     NewRestyClient(),
		key:        cfg.Key,
		rateLimit:  cfg.RateLimit,
	}
}

func (a *Agent) Run() {
	if err := logger.Initialize("debug", "local"); err != nil {
		panic(fmt.Sprintf("logger initialization fail: %v", err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsChunks := make(chan []*m.Metrics)
	defer close(metricsChunks)

	for i := 0; i < a.rateLimit; i++ {
		go a.postWorker(metricsChunks)
	}

	go func(ctx context.Context) {
		pollTicker := time.NewTicker(a.pollIntr)
		defer pollTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				a.collectAllMetrics()
			}
		}
	}(ctx)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	repTicker := time.NewTicker(a.repIntr)
	defer repTicker.Stop()

	for {
		select {
		case s := <-interrupt:
			logger.Log.Infoln("server - Run - signal: " + s.String())
			return
		case <-repTicker.C:
			metricsChunks <- a.mstorage.ReadMetrics()
		}
	}
}

func (a *Agent) postWorker(metricsChunks chan []*m.Metrics) {
	for metricsChunk := range metricsChunks {
		if err := a.postMetrics(metricsChunk); err != nil {
			logger.Log.Errorf("Failed to post metrics: %v", err)
		}
	}
}

func (a *Agent) collectAllMetrics() {
	metricsCh := make(chan *m.Metrics)

	var wg sync.WaitGroup
	wg.Add(2)

	go a.collectRuntimeMetrics(&wg, metricsCh)
	go a.collectExtraMetrics(&wg, metricsCh)

	go func() {
		wg.Wait()
		close(metricsCh)
	}()

	metrics := make([]*m.Metrics, 0, 50)
	for metric := range metricsCh {
		metrics = append(metrics, metric)
	}

	pollCount := a.pollCount.Add(1)
	metrics = append(metrics, &m.Metrics{ID: "PollCount", MType: "counter", Delta: &pollCount})

	a.mstorage.ReplaceMetrics(metrics...)
}

func (a *Agent) collectRuntimeMetrics(wg *sync.WaitGroup, metricsCh chan<- *m.Metrics) {
	defer wg.Done()

	memStats := new(runtime.MemStats)
	runtime.ReadMemStats(memStats)

	for _, metricName := range runtimeMetricNames {
		v, ok := gaugeVal(memStats, metricName)
		if !ok {
			log.Printf("Failed to get gauge value for %s", metricName)
			continue
		}
		metricsCh <- &m.Metrics{ID: metricName, MType: "gauge", Value: &v}
	}
	rv := getRandomFloat()
	metricsCh <- &m.Metrics{ID: "RandomValue", MType: "gauge", Value: &rv}
}

func (a *Agent) collectExtraMetrics(wg *sync.WaitGroup, metricsCh chan<- *m.Metrics) {
	// TotalMemory FreeMemory CPUutilization1
	defer wg.Done()
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Errorf("Failed to get virtual memorty stats: %v", err)
		return
	}
	totalMem := float64(vmStat.Total)
	metricsCh <- &m.Metrics{ID: "TotalMemory", MType: "gauge", Value: &totalMem}

	freeMem := float64(vmStat.Free)
	metricsCh <- &m.Metrics{ID: "FreeMemory", MType: "gauge", Value: &freeMem}

	CPUPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		logger.Log.Errorf("Failed to get cpu percent: %v", err)
		return
	}
	if len(CPUPercent) == 0 {
		logger.Log.Error("Length of cpuPercent is 0")
		return
	}
	metricsCh <- &m.Metrics{ID: "CPUutilization1", MType: "gauge", Value: &CPUPercent[0]}
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
	logger.Log.Infoln("im happening")
	if len(metrics) == 0 {
		logger.Log.Warn("empty metrics slice")
		return nil
	}
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

	req := a.client.R()
	if a.key != "" {
		hshr := hasher.New(a.key)
		req = req.SetHeader(hasher.HashHeaderKey, hshr.Hash(buff.Bytes()))
	}
	req = req.SetBody(&buff)
	res, err := req.Post(url)
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
