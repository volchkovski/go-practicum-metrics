package main

import (
	"github.com/volchkovski/go-practicum-metrics/internal/collector"
)

func main() {
	c := collector.New()
	if err := c.Run(); err != nil {
		panic(err)
	}
}
