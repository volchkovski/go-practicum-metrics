package main

import (
	"github.com/volchkovski/go-practicum-metrics/internal/agent"
)

func main() {
	parseFlags()
	a := agent.New(flagServerAddr, flagRepIntr, flagPollIntr)
	a.Run()
}
