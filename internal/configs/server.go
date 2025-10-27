package configs

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr            string `env:"ADDRESS" json:"address"`
	StoreIntr       int    `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool   `env:"RESTORE" json:"restore"`
	LogLevel        string `env:"LOG_LEVEL"`
	Env             string `env:"ENVIRONMENT"`
	DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY" json:"crypto_key"`
}

func (cfg *ServerConfig) UnmarshalJSON(data []byte) error {
	type Alias ServerConfig
	var temp struct {
		Alias
		StoreIntr string `json:"store_interval"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*cfg = ServerConfig(temp.Alias)
	d, err := time.ParseDuration(temp.StoreIntr)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}
	cfg.StoreIntr = int(d.Seconds())
	return nil
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := new(ServerConfig)

	configFileFlags := flag.NewFlagSet("config", flag.ExitOnError)
	var c, config string
	configFileFlags.StringVar(&c, "c", "", "path to config file")
	configFileFlags.StringVar(&config, "config", "", "path to config file")
	if err := configFileFlags.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	envConfigFile := os.Getenv("CONFIG")
	for _, path := range []string{c, config, envConfigFile} {
		if path != "" {
			if err := parseConfigFile(cfg, path); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}
	if err := parseConfigFields(cfg, configFileFlags.Args()); err != nil {
		return nil, fmt.Errorf("failed to parse config fields from flags: %w", err)
	}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}
	return cfg, nil
}

func parseConfigFile(cfg *ServerConfig, path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	data, err := io.ReadAll(f)
	if err != nil {
		return
	}
	return json.Unmarshal(data, cfg)
}

func parseConfigFields(cfg *ServerConfig, args []string) error {
	configFieldsFlags := flag.NewFlagSet("fields", flag.ExitOnError)
	configFieldsFlags.StringVar(&cfg.Addr, "a", ":8080", "address and port to run server")
	configFieldsFlags.IntVar(&cfg.StoreIntr, "i", 300, "metrics saves to file each time after this interval")
	configFieldsFlags.StringVar(&cfg.FileStoragePath, "f", `./metrics.json`, "file path for metrics saving")
	configFieldsFlags.BoolVar(&cfg.Restore, "r", false, "load dumped metrics at server start")
	configFieldsFlags.StringVar(&cfg.LogLevel, "l", "info", "level of logging")
	configFieldsFlags.StringVar(&cfg.Env, "e", "local", "environment: prod, local")
	configFieldsFlags.StringVar(&cfg.DSN, "d", "", "postgres data source name")
	configFieldsFlags.StringVar(&cfg.Key, "k", "", "key for making hash")
	configFieldsFlags.StringVar(&cfg.CryptoKey, "crypto-key", "./private.pem", "path to rsa private key")
	return configFieldsFlags.Parse(args)
}
