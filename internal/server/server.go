package server

import (
	"net/http"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/routers"
	"github.com/volchkovski/go-practicum-metrics/internal/services"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
)

func Run(cfg *configs.ServerConfig) error {
	storage := storage.NewMemStorage()

	service := services.NewMetricService(storage)

	router := routers.NewMetricRouter(service)

	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		return err
	}
	return nil
}
