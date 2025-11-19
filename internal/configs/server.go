package configs

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr            string `env:"ADDRESS" json:"address"`
	GRPCAddr        string `env:"GRPC_ADDRESS" json:"grpc_address"`
	StoreIntr       int    `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool   `env:"RESTORE" json:"restore"`
	LogLevel        string `env:"LOG_LEVEL"`
	Env             string `env:"ENVIRONMENT"`
	DSN             string `env:"DATABASE_DSN" json:"database_dsn"`
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

func (cfg *ServerConfig) UnmarshalJSON(data []byte) error {
	type Alias ServerConfig
	var temp struct {
		Alias
		StoreIntr time.Duration `json:"store_interval"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*cfg = ServerConfig(temp.Alias)
	cfg.StoreIntr = int(temp.StoreIntr.Seconds())
	return nil
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := new(ServerConfig)

	args, err := processConfigFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process config file: %w", err)
	}

	if err = parseServerConfigFields(cfg, args); err != nil {
		return nil, fmt.Errorf("failed to parse config fields from flags: %w", err)
	}
	if err = env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("server config error: %w", err)
	}
	return cfg, nil
}

func parseServerConfigFields(cfg *ServerConfig, args []string) error {
	configFieldsFlags := flag.NewFlagSet("fields", flag.ExitOnError)
	configFieldsFlags.StringVar(&cfg.Addr, "a", ":8080", "address and port to run server")
	configFieldsFlags.StringVar(&cfg.GRPCAddr, "g", ":3200", "address and port to run gRPC server")
	configFieldsFlags.IntVar(&cfg.StoreIntr, "i", 300, "metrics saves to file each time after this interval")
	configFieldsFlags.StringVar(&cfg.FileStoragePath, "f", `./metrics.json`, "file path for metrics saving")
	configFieldsFlags.BoolVar(&cfg.Restore, "r", false, "load dumped metrics at server start")
	configFieldsFlags.StringVar(&cfg.LogLevel, "l", "info", "level of logging")
	configFieldsFlags.StringVar(&cfg.Env, "e", "local", "environment: prod, local")
	configFieldsFlags.StringVar(&cfg.DSN, "d", "", "postgres data source name")
	configFieldsFlags.StringVar(&cfg.Key, "k", "", "key for making hash")
	configFieldsFlags.StringVar(&cfg.CryptoKey, "crypto-key", "./private.pem", "path to rsa private key")
	configFieldsFlags.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet in CIDR notation")
	return configFieldsFlags.Parse(args)
}
