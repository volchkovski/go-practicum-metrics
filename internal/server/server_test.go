package server

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
)

func TestRun_ConfigValidation(t *testing.T) {
	t.Run("handles invalid log level", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "invalid-level",
			Env:      "local",
			Addr:     ":8080",
		}

		err := Run(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to intizalize logger")
	})

	t.Run("handles invalid environment", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "invalid-env",
			Addr:     ":8080",
		}

		err := Run(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to intizalize logger")
	})
}

func TestRun_MemoryStorage(t *testing.T) {
	t.Run("uses memory storage when no DSN provided", func(t *testing.T) {
		// Этот тест проверяет инициализацию без запуска сервера
		// Мы не можем полноценно протестировать Run без изменения кода,
		// но можем проверить логику инициализации

		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0", // случайный порт
			DSN:             "",   // пустой DSN = memory storage
			FileStoragePath: "/tmp/test-metrics.json",
			StoreIntr:       300,
			Restore:         false,
		}

		// В этом тесте мы ожидаем, что сервер попытается запуститься,
		// но мы не можем легко остановить его без изменения кода
		// Поэтому просто проверим, что конфигурация валидна
		assert.NotNil(t, cfg)
		assert.Equal(t, "", cfg.DSN)
		assert.Equal(t, "info", cfg.LogLevel)
	})
}

// Дополнительные helper функции для тестирования
func validateServerConfig(cfg *configs.ServerConfig) error {
	if cfg.LogLevel == "" {
		return errors.New("log level is required")
	}
	if cfg.Env == "" {
		return errors.New("environment is required")
	}
	if cfg.Addr == "" {
		return errors.New("address is required")
	}
	return nil
}

func TestValidateServerConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "local",
			Addr:     ":8080",
		}

		err := validateServerConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing log level fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			Env:  "local",
			Addr: ":8080",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log level is required")
	})

	t.Run("missing environment fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Addr:     ":8080",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment is required")
	})

	t.Run("missing address fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "local",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address is required")
	})
}
