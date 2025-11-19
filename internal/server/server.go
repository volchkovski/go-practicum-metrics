// Package server provides the main server application logic and startup functionality.
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/rsakey"

	"github.com/volchkovski/go-practicum-metrics/internal/backup"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/grpcserver"
	"github.com/volchkovski/go-practicum-metrics/internal/httpserver"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	"github.com/volchkovski/go-practicum-metrics/internal/routers"
	"github.com/volchkovski/go-practicum-metrics/internal/services"
	"github.com/volchkovski/go-practicum-metrics/internal/storage/mem"
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
		strg = mem.NewMemStorage()
		logger.Log.Infoln("Memory storage in use")
	} else {
		if strg, err = pg.New(cfg.DSN); err != nil {
			logger.Log.Errorf("Postgres creation failed: %s", err.Error())
			return
		}
		logger.Log.Infoln("Postgres storage in use")
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

	privRSA, err := rsakey.GetPrivateKey(cfg.CryptoKey)
	if err != nil && !errors.Is(err, rsakey.ErrEmptyPath) {
		logger.Log.Errorf("Failed to load rsa private key: %s", err.Error())
		return
	}

	router := routers.NewMetricRouter(cfg.Key, privRSA, cfg.TrustedSubnet, service)
	hs := httpserver.New(router, cfg.Addr)

	// Start gRPC server if address is configured
	var gs *grpcserver.GRPCServer
	if cfg.GRPCAddr != "" {
		gs = grpcserver.New(cfg.GRPCAddr, service, logger.Log, cfg.TrustedSubnet)
		gs.Start()
		logger.Log.Infoln("gRPC server enabled on", cfg.GRPCAddr)
	}

	hs.Start()
	b.Start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Wait for errors or interrupt
	if gs != nil {
		select {
		case err = <-hs.Notify():
			return gracefulShutdown(hs, gs, b)
		case err = <-gs.Notify():
			return gracefulShutdown(hs, gs, b)
		case err = <-b.Notify():
			return gracefulShutdown(hs, gs, b)
		case s := <-interrupt:
			logger.Log.Infoln("server - Run - signal: " + s.String())
			return gracefulShutdown(hs, gs, b)
		}
	} else {
		select {
		case err = <-hs.Notify():
			return
		case err = <-b.Notify():
			return
		case s := <-interrupt:
			logger.Log.Infoln("server - Run - signal: " + s.String())
			return gracefulShutdown(hs, nil, b)
		}
	}
}

func gracefulShutdown(hs *httpserver.HTTPServer, gs *grpcserver.GRPCServer, b *backup.MetricsBackup) error {
	logger.Log.Infoln("Starting graceful shutdown...")

	// Создаем контекст с таймаутом для shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем backup и делаем финальное сохранение
	logger.Log.Infoln("Stopping backup and saving final metrics...")
	if err := b.Stop(); err != nil {
		logger.Log.Errorf("Error during backup final save: %v", err)
	}

	// Останавливаем gRPC сервер
	if gs != nil {
		logger.Log.Infoln("Shutting down gRPC server...")
		if err := gs.Shutdown(); err != nil {
			logger.Log.Errorf("Error during gRPC server shutdown: %v", err)
		}
	}

	// Останавливаем HTTP сервер
	logger.Log.Infoln("Shutting down HTTP server...")
	if err := hs.Shutdown(ctx); err != nil {
		logger.Log.Errorf("Error during HTTP server shutdown: %v", err)
		return err
	}

	logger.Log.Infoln("Server graceful shutdown completed")
	return nil
}
