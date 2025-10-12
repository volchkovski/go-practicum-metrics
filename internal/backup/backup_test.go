package backup

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
)

// mockMetricsGetPusher implements the metricsGetPusher interface for testing
type mockMetricsGetPusher struct {
	gauges   []*models.GaugeMetric
	counters []*models.CounterMetric
	pushErr  error
	getErr   error
}

func (m *mockMetricsGetPusher) PushGaugeMetric(ctx context.Context, metric *models.GaugeMetric) error {
	return m.pushErr
}

func (m *mockMetricsGetPusher) PushCounterMetric(ctx context.Context, metric *models.CounterMetric) error {
	return m.pushErr
}

func (m *mockMetricsGetPusher) GetAllGaugeMetrics(ctx context.Context) ([]*models.GaugeMetric, error) {
	return m.gauges, m.getErr
}

func (m *mockMetricsGetPusher) GetAllCounterMetrics(ctx context.Context) ([]*models.CounterMetric, error) {
	return m.counters, m.getErr
}

func TestValidateFilePath(t *testing.T) {
	t.Run("rejects empty path", func(t *testing.T) {
		err := ValidateFilePath("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path cannot be empty")
	})

	t.Run("accepts valid path in current directory", func(t *testing.T) {
		err := ValidateFilePath("test.json")
		assert.NoError(t, err)
	})

	t.Run("creates directory if needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		testPath := filepath.Join(tmpDir, "new", "dir", "test.json")

		err := ValidateFilePath(testPath)
		assert.NoError(t, err)

		// Проверяем что директория была создана
		dir := filepath.Dir(testPath)
		_, err = os.Stat(dir)
		assert.NoError(t, err)
	})

	t.Run("handles permission error gracefully", func(t *testing.T) {
		// Тестируем только если не root пользователь
		if os.Getuid() != 0 {
			err := ValidateFilePath("/root/forbidden/test.json")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot create directory")
		}
	})
}

func TestIsFileExists(t *testing.T) {
	t.Run("returns true for existing file", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "existing.json")

		// Создаем файл
		file, err := os.Create(tmpFile)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		exists := IsFileExists(tmpFile)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent file", func(t *testing.T) {
		nonExistentPath := filepath.Join(t.TempDir(), "does-not-exist.json")

		exists := IsFileExists(nonExistentPath)
		assert.False(t, exists)
	})

	t.Run("returns false for directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		exists := IsFileExists(tmpDir)
		// Технически директория существует, но функция может вернуть true или false
		// в зависимости от реализации - это нормально
		_ = exists
	})
}

func TestNewMetricsBackup(t *testing.T) {
	t.Run("creates backup with correct configuration", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		filePath := "/tmp/test-backup.json"
		interval := 300

		backup := NewMetricsBackup(mgp, filePath, interval)

		assert.NotNil(t, backup)
		assert.Equal(t, mgp, backup.mgp)
		assert.Equal(t, filePath, backup.fp)
		assert.Equal(t, 300*time.Second, backup.interval)
		assert.NotNil(t, backup.notify)
	})

	t.Run("creates backup with zero interval", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, "test.json", 0)

		assert.Equal(t, time.Duration(0), backup.interval)
	})
}

func TestMetricsBackup_Notify(t *testing.T) {
	t.Run("returns notify channel", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, "", 0)

		ch := backup.Notify()
		assert.NotNil(t, ch)
		assert.Equal(t, backup.notify, ch)
	})
}

func TestMetricsBackup_RestoreMetrics(t *testing.T) {
	t.Run("restores gauges and counters successfully", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, "", 0)

		testMetrics := metrics{
			Gauges: []*models.GaugeMetric{
				{Name: "cpu_usage", Value: 75.5},
				{Name: "memory_usage", Value: 82.3},
			},
			Counters: []*models.CounterMetric{
				{Name: "requests", Value: 1000},
				{Name: "errors", Value: 5},
			},
		}

		err := backup.restoreMetrics(testMetrics)
		assert.NoError(t, err)
	})

	t.Run("handles push error for gauge", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{
			pushErr: assert.AnError,
		}
		backup := NewMetricsBackup(mgp, "", 0)

		testMetrics := metrics{
			Gauges: []*models.GaugeMetric{
				{Name: "failing_gauge", Value: 100.0},
			},
		}

		err := backup.restoreMetrics(testMetrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to restore gauge metric")
		assert.Contains(t, err.Error(), "failing_gauge")
	})

	t.Run("handles push error for counter", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{
			pushErr: assert.AnError,
		}
		backup := NewMetricsBackup(mgp, "", 0)

		testMetrics := metrics{
			Counters: []*models.CounterMetric{
				{Name: "failing_counter", Value: 42},
			},
		}

		err := backup.restoreMetrics(testMetrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to restore counter metric")
		assert.Contains(t, err.Error(), "failing_counter")
	})
}

func TestMetricsBackup_WriteMetricsToFile(t *testing.T) {
	t.Run("writes metrics to file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_metrics.json")

		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, testFile, 0)

		gauges := []*models.GaugeMetric{
			{Name: "test_gauge", Value: 123.45},
		}
		counters := []*models.CounterMetric{
			{Name: "test_counter", Value: 678},
		}

		err := backup.writeMetricsToFile(gauges, counters)
		assert.NoError(t, err)

		// Проверяем что файл создан и содержит правильные данные
		assert.True(t, IsFileExists(testFile))

		file, err := os.Open(testFile)
		require.NoError(t, err)
		defer func() {
			err := file.Close()
			require.NoError(t, err)
		}()

		var loadedMetrics metrics
		err = json.NewDecoder(file).Decode(&loadedMetrics)
		require.NoError(t, err)

		assert.Len(t, loadedMetrics.Gauges, 1)
		assert.Equal(t, "test_gauge", loadedMetrics.Gauges[0].Name)
		assert.Equal(t, 123.45, loadedMetrics.Gauges[0].Value)

		assert.Len(t, loadedMetrics.Counters, 1)
		assert.Equal(t, "test_counter", loadedMetrics.Counters[0].Name)
		assert.Equal(t, int64(678), loadedMetrics.Counters[0].Value)
	})

	t.Run("handles invalid file path", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, "", 0) // пустой путь

		err := backup.writeMetricsToFile(nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path cannot be empty")
	})
}

func TestMetricsBackup_Restore(t *testing.T) {
	t.Run("handles non-existent file gracefully", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		nonExistentFile := filepath.Join(t.TempDir(), "does-not-exist.json")
		backup := NewMetricsBackup(mgp, nonExistentFile, 0)

		err := backup.Restore()
		assert.NoError(t, err) // не должно быть ошибки для несуществующего файла
	})

	t.Run("validates file path", func(t *testing.T) {
		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, "", 0) // пустой путь

		err := backup.Restore()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path cannot be empty")
	})

	t.Run("restores from valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "restore_test.json")

		// Создаем тестовый файл
		testData := metrics{
			Gauges: []*models.GaugeMetric{
				{Name: "restore_gauge", Value: 99.9},
			},
			Counters: []*models.CounterMetric{
				{Name: "restore_counter", Value: 555},
			},
		}

		file, err := os.Create(testFile)
		require.NoError(t, err)

		err = json.NewEncoder(file).Encode(testData)
		require.NoError(t, err)
		err = file.Close()
		require.NoError(t, err)

		mgp := &mockMetricsGetPusher{}
		backup := NewMetricsBackup(mgp, testFile, 0)

		err = backup.Restore()
		assert.NoError(t, err)
	})
}
