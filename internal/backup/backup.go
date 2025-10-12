package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// ValidateFilePath проверяет корректность пути к файлу
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}

	// Проверяем что директория существует или может быть создана
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", dir, err)
		}
	}

	return nil
}

// IsFileExists проверяет существование файла
func IsFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (b *MetricsBackup) Restore() (err error) {
	if err := ValidateFilePath(b.fp); err != nil {
		return err
	}

	if !IsFileExists(b.fp) {
		// Файл не существует - это не ошибка при первом запуске
		return nil
	}

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

	return b.restoreMetrics(m)
}

// restoreMetrics восстанавливает метрики из структуры данных
func (b *MetricsBackup) restoreMetrics(m metrics) error {
	ctx := context.Background()

	for _, gauge := range m.Gauges {
		if err := b.mgp.PushGaugeMetric(ctx, gauge); err != nil {
			return fmt.Errorf("failed to restore gauge metric %s: %w", gauge.Name, err)
		}
	}

	for _, counter := range m.Counters {
		if err := b.mgp.PushCounterMetric(ctx, counter); err != nil {
			return fmt.Errorf("failed to restore counter metric %s: %w", counter.Name, err)
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

	return b.writeMetricsToFile(gauges, counters)
}

// writeMetricsToFile записывает метрики в файл
func (b *MetricsBackup) writeMetricsToFile(gauges []*models.GaugeMetric, counters []*models.CounterMetric) (err error) {
	if err := ValidateFilePath(b.fp); err != nil {
		return err
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
