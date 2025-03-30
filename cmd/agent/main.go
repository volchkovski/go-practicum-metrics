package main

import (
	"log"

	"github.com/volchkovski/go-practicum-metrics/internal/agent"
)

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}
	a := agent.New(flagServerAddr, flagRepIntr, flagPollIntr)
	a.Run()
}
