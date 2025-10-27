// Package configs provides configuration structures and parsing logic
// for the metrics collection system. It handles command-line flags
// and environment variables for both agent and server components.
package configs

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	ServerAddr string `env:"ADDRESS" json:"address"`
	ReportIntr int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollIntr   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	Key        string `env:"KEY" json:"key"`
	RateLimit  int    `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey  string `env:"CRYPTO_KEY" json:"crypto_key"`
}

func (ac *AgentConfig) UnmarshalJSON(data []byte) error {
	type Alias AgentConfig
	var temp struct {
		Alias
		ReportIntr time.Duration `json:"report_interval"`
		PollIntr   time.Duration `json:"poll_interval"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*ac = AgentConfig(temp.Alias)
	ac.ReportIntr = int(temp.ReportIntr.Seconds())
	ac.PollIntr = int(temp.PollIntr.Seconds())
	return nil
}

func NewAgentConfig() (*AgentConfig, error) {
	cfg := new(AgentConfig)

	args, err := processConfigFile(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process config file: %w", err)
	}

	if err = parseAgentFlags(cfg, args); err != nil {
		return nil, fmt.Errorf("failed to parse config fields from flags: %w", err)
	}
	if err = env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("agent config error: %w", err)
	}
	return cfg, nil
}

func parseAgentFlags(cfg *AgentConfig, args []string) error {
	configFieldsFlags := flag.NewFlagSet("fields", flag.ExitOnError)
	configFieldsFlags.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "server address and port to push")
	configFieldsFlags.IntVar(&cfg.ReportIntr, "r", 10, "each time to report metrics")
	configFieldsFlags.IntVar(&cfg.PollIntr, "p", 2, "each time to poll metrics")
	configFieldsFlags.StringVar(&cfg.Key, "k", "", "key for making hash")
	configFieldsFlags.IntVar(&cfg.RateLimit, "l", 1, "number of simultaneous requests")
	configFieldsFlags.StringVar(&cfg.CryptoKey, "crypto-key", "./public.pem", "path to rsa public key")
	return configFieldsFlags.Parse(args)
}
