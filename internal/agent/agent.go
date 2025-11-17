// Package agent provides the metrics collection agent that gathers
// system metrics and sends them to a metrics server via HTTP.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/rsakey"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/volchkovski/go-practicum-metrics/internal/hasher"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"

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
	publicRSA  *rsa.PublicKey
}

func New(cfg *configs.AgentConfig) (*Agent, error) {
	pubRSA, err := rsakey.GetPublicKey(cfg.CryptoKey)
	if err != nil && !errors.Is(err, rsakey.ErrEmptyPath) {
		return nil, err
	}
	return &Agent{
		mstorage:   NewMetricsStorage(),
		repIntr:    time.Duration(cfg.ReportIntr) * time.Second,
		pollIntr:   time.Duration(cfg.PollIntr) * time.Second,
		serverAddr: cfg.ServerAddr,
		pollCount:  atomic.Int64{},
		client:     NewRestyClient(),
		key:        cfg.Key,
		rateLimit:  cfg.RateLimit,
		publicRSA:  pubRSA,
	}, nil
}

func (a *Agent) Run() error {
	if err := logger.Initialize("debug", "local"); err != nil {
		return fmt.Errorf("logger initialization fail: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	metricsChunks := make(chan []*m.Metrics)

	logger.Log.Infoln("Starting workers")
	var wg sync.WaitGroup
	a.startPostWorkers(&wg, metricsChunks)

	logger.Log.Infoln("Starting metric collection")
	a.startMetricCollection(ctx)

	repTicker := time.NewTicker(a.repIntr)
	defer repTicker.Stop()

	logger.Log.Infoln("Starting reporting loop")
	a.runReportingLoop(ctx, &wg, metricsChunks, repTicker)
	return nil
}

func (a *Agent) runReportingLoop(ctx context.Context, wg *sync.WaitGroup, metricsChunks chan<- []*m.Metrics, repTicker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			logger.Log.Infoln("agent - runReportingLoop: " + ctx.Err().Error())
			// Отправляем последние метрики перед завершением
			logger.Log.Infoln("Sending final metrics before shutdown...")
			lastMetrics := a.mstorage.ReadMetrics()
			if len(lastMetrics) > 0 {
				metricsChunks <- lastMetrics
			}
			// Закрываем канал и ждем завершения всех воркеров
			close(metricsChunks)
			logger.Log.Infoln("Waiting for workers to finish...")
			wg.Wait()
			logger.Log.Infoln("Agent graceful shutdown completed")
			return
		case <-repTicker.C:
			metricsChunks <- a.mstorage.ReadMetrics()
		}
	}
}

func (a *Agent) startMetricCollection(ctx context.Context) {
	go func() {
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
	}()
}

func (a *Agent) startPostWorkers(wg *sync.WaitGroup, metricsChunks <-chan []*m.Metrics) {
	logger.Log.Debugf("Workers number: %d", a.rateLimit)
	for i := 0; i < a.rateLimit; i++ {
		wg.Add(1)
		go a.postWorker(wg, metricsChunks)
	}
}

func (a *Agent) postWorker(wg *sync.WaitGroup, metricsChunks <-chan []*m.Metrics) {
	defer wg.Done()
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
	logger.Log.Infoln("Posting metrics")
	if len(metrics) == 0 {
		logger.Log.Warn("empty metrics slice")
		return nil
	}
	url := "http://" + a.serverAddr + "/updates/"

	p, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	if a.publicRSA != nil {
		if p, err = rsa.EncryptPKCS1v15(cryptorand.Reader, a.publicRSA, p); err != nil {
			return fmt.Errorf("failed to encrypt: %w", err)
		}
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

	// Add X-Real-IP header with the host's IP address
	if localIP := getLocalIP(); localIP != "" {
		req = req.SetHeader("X-Real-IP", localIP)
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

// getLocalIP returns the non-loopback local IP of the host.
// It returns an empty string if no suitable IP is found.
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Log.Warnf("Failed to get interface addresses: %v", err)
		return ""
	}

	for _, address := range addrs {
		// Check if the address is an IP address (not a network)
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	logger.Log.Warn("No suitable local IP address found")
	return ""
}
