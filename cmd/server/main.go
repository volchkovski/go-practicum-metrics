package main

import (
	"fmt"
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/server"
	_ "net/http/pprof"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	showBuildInfo()
	cfg, err := configs.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Run(cfg))
}

func showBuildInfo() {
	fmt.Printf("Build version: %q\n", buildVersion)
	fmt.Printf("Build date: %q\n", buildDate)
	fmt.Printf("Build commit: %q\n", buildCommit)
}
