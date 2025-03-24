package main

import (
	"github.com/volchkovski/go-practicum-metrics/internal/agent"
)

func main() {
	a := agent.New()
	a.Run()
}
