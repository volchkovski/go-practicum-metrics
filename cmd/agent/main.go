package main

import (
	"github.com/volchkovski/go-practicum-metrics/internal/agent"
	"log"
)

func main() {
	a := agent.New()
	log.Fatal(a.Run())
}
