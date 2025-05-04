package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr            string `env:"ADDRESS"`
	StoreIntr       int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	LogLevel        string `env:"LOG_LEVEL"`
	Env             string `env:"ENVIRONMENT"`
	DSN             string `env:"DATABASE_DSN"`
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := new(ServerConfig)
	parseServerFlags(cfg)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}
	return cfg, nil
}

func parseServerFlags(cfg *ServerConfig) {
	flag.StringVar(&cfg.Addr, "a", ":8080", "address and port to run server")
	flag.IntVar(&cfg.StoreIntr, "i", 300, "metrics saves to file each time after this interval")
	flag.StringVar(&cfg.FileStoragePath, "f", `./metrics.json`, "file path for metrics saving")
	flag.BoolVar(&cfg.Restore, "r", false, "load dumped metrics at server start")
	flag.StringVar(&cfg.LogLevel, "l", "info", "level of logging")
	flag.StringVar(&cfg.Env, "e", "local", "environment: prod, local")
	flag.StringVar(&cfg.DSN, "d", "user=postgres password=postgres host=localhost port=5432 dbname=postgres sslmode=disable", "postgres data source name")
	flag.Parse()
}
