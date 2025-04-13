package server

import (
	"fmt"
	"net/http"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
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

	router := routers.NewMetricRouter(service)

	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		return err
	}
	return nil
}
