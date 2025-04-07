package main

import (
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/server"
)

func main() {
	cfg, err := configs.NewServerConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Run(cfg))
}
