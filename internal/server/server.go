package server

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/volchkovski/go-practicum-metrics/internal/backup"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/httpserver"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	"github.com/volchkovski/go-practicum-metrics/internal/routers"
	"github.com/volchkovski/go-practicum-metrics/internal/services"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
)

func Run(cfg *configs.ServerConfig) error {
	if err := logger.Initialize("debug"); err != nil {
		return fmt.Errorf("failed to intizalize logger: %w", err)
	}
	defer logger.Log.Sync()

	storage := storage.NewMemStorage()

	service := services.NewMetricService(storage)

	b := backup.NewMetricsBackup(service, cfg.FileStoragePath, cfg.StoreIntr)

	if cfg.Restore {
		if err := b.Restore(); err != nil && err != io.EOF {
			return err
		}
	}

	router := routers.NewMetricRouter(service)
	httpserver := httpserver.New(router, cfg.Addr)

	httpserver.Start()
	b.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-httpserver.Notify():
		return err
	case err := <-b.Notify():
		return err
	case s := <-interrupt:
		logger.Log.Infoln("server - Run - signal: " + s.String())
	}

	return nil
}
