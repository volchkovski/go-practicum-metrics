package main

import (
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/agent"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
)

func main() {
	cfg, err := configs.NewAgentConfig()
	if err != nil {
		log.Fatal(err)
	}
	a := agent.New(cfg)
	a.Run()
}
