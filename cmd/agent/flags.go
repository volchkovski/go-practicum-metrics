package main

import (
	"flag"
)

var (
	flagServerAddr string
	flagRepIntr    int
	flagPollIntr   int
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server address and port to push")
	flag.IntVar(&flagRepIntr, "r", 10, "each time to report metrics")
	flag.IntVar(&flagPollIntr, "p", 2, "each time to poll metrics")
	flag.Parse()
}
