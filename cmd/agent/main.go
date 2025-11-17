package main

import (
	"fmt"
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/agent"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	showBuildInfo()
	cfg, err := configs.NewAgentConfig()
	if err != nil {
		log.Fatal(err)
	}
	a, err := agent.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func showBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
