package configs

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type (
	AgentConfig struct {
		ServerAddr string `env:"ADDRESS"`
		ReportIntr int    `env:"REPORT_INTERVAL"`
		PollIntr   int    `env:"POLL_INTERVAL"`
	}

	ServerConfig struct {
		Addr string `env:"ADDRESS"`
	}
)

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
	flag.Parse()
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
	flag.Parse()
}
