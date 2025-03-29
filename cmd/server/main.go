package main

import (
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/server"
)

func main() {
	parseFlags()
	s := server.New()
	log.Fatal(s.Run(flagRunAddr))
}
