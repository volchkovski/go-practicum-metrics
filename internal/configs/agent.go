// Package configs provides configuration structures and parsing logic
// for the metrics collection system. It handles command-line flags
// and environment variables for both agent and server components.
package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ServerAddr string `env:"ADDRESS"`
	ReportIntr int    `env:"REPORT_INTERVAL"`
	PollIntr   int    `env:"POLL_INTERVAL"`
	Key        string `env:"KEY"`
	RateLimit  int    `env:"RATE_LIMIT"`
	CryptoKey  string `env:"CRYPTO_KEY"`
}

func NewAgentConfig() (*AgentConfig, error) {
	cfg := new(AgentConfig)
	parseAgentFlags(cfg)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("agent config error: %w", err)
	}
	return cfg, nil
}

func parseAgentFlags(cfg *AgentConfig) {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "server address and port to push")
	flag.IntVar(&cfg.ReportIntr, "r", 10, "each time to report metrics")
	flag.IntVar(&cfg.PollIntr, "p", 2, "each time to poll metrics")
	flag.StringVar(&cfg.Key, "k", "", "key for making hash")
	flag.IntVar(&cfg.RateLimit, "l", 1, "number of simultaneous requests")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "./public.pem", "path to rsa public key")
	flag.Parse()
}
