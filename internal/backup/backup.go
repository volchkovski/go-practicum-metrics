package backup

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
)

type metrics struct {
	Gauges   []*models.GaugeMetric   `json:"gauges"`
	Counters []*models.CounterMetric `json:"counters"`
}

type metricsGetPusher interface {
	handlers.MetricPusher
	handlers.AllMetricsGetter
}

type MetricsBackup struct {
	mgp      metricsGetPusher
	fp       string
	interval time.Duration
	notify   chan error
}

func NewMetricsBackup(mgp metricsGetPusher, fp string, intr int) *MetricsBackup {
	return &MetricsBackup{
		mgp:      mgp,
		fp:       fp,
		interval: time.Duration(intr) * time.Second,
		notify:   make(chan error, 1),
	}
}

func (b *MetricsBackup) Notify() chan error {
	return b.notify
}

func (b *MetricsBackup) Restore() (err error) {
	file, err := os.OpenFile(b.fp, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	var m metrics
	err = json.NewDecoder(file).Decode(&m)
	if err != nil {
		return
	}
	for _, gauge := range m.Gauges {
		err = b.mgp.PushGaugeMetric(context.Background(), gauge)
		if err != nil {
			return
		}
	}
	for _, counter := range m.Counters {
		err = b.mgp.PushCounterMetric(context.Background(), counter)
		if err != nil {
			return
		}
	}
	return nil
}

func (b *MetricsBackup) Start() {
	go func() {
		for {
			time.Sleep(b.interval)
			err := b.dumpMetrics()
			if err != nil {
				b.notify <- err
				break
			}
		}
		close(b.notify)
	}()
}

func (b *MetricsBackup) dumpMetrics() (err error) {
	ctx := context.Background()
	gauges, err := b.mgp.GetAllGaugeMetrics(ctx)
	if err != nil {
		return
	}
	counters, err := b.mgp.GetAllCounterMetrics(ctx)
	if err != nil {
		return
	}
	file, err := os.Create(b.fp)
	if err != nil {
		return
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()

	m := metrics{
		Gauges:   gauges,
		Counters: counters,
	}

	err = json.NewEncoder(file).Encode(m)
	if err != nil {
		return
	}
	return nil
}
