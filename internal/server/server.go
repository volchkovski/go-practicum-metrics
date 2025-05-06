package server

import (
	"errors"
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
	"github.com/volchkovski/go-practicum-metrics/internal/storage/pg"
)

func Run(cfg *configs.ServerConfig) (err error) {
	if err = logger.Initialize(cfg.LogLevel, cfg.Env); err != nil {
		return fmt.Errorf("failed to intizalize logger: %w", err)
	}

	defer func() {
		if errSync := logger.Log.Sync(); errSync != nil {
			err = errors.Join(err, errSync)
		}
	}()

	var strg services.MetricStorage
	if cfg.DSN == "" {
		strg = storage.NewMemStorage()
	} else {
		if strg, err = pg.New(cfg.DSN); err != nil {
			return
		}
	}
	service := services.NewMetricService(strg)
	defer func() {
		if errServiceClose := service.Close(); errServiceClose != nil {
			err = errors.Join(err, errServiceClose)
		}
	}()

	b := backup.NewMetricsBackup(service, cfg.FileStoragePath, cfg.StoreIntr)

	if cfg.Restore {
		if err = b.Restore(); err != nil && !errors.Is(err, io.EOF) {
			return
		}
	}

	router := routers.NewMetricRouter(service)
	httpserver := httpserver.New(router, cfg.Addr)

	httpserver.Start()
	b.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-httpserver.Notify():
		return
	case err = <-b.Notify():
		return
	case s := <-interrupt:
		logger.Log.Infoln("server - Run - signal: " + s.String())
	}

	return nil
}
