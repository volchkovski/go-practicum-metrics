package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	flagServerAddr string
	flagRepIntr    int
	flagPollIntr   int
)

func parseFlags() error {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server address and port to push")
	flag.IntVar(&flagRepIntr, "r", 10, "each time to report metrics")
	flag.IntVar(&flagPollIntr, "p", 2, "each time to poll metrics")
	flag.Parse()
	if envServerAddr, ok := os.LookupEnv("ADDRESS"); ok {
		flagServerAddr = envServerAddr
	}
	if envRepIntr, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		val, err := strconv.Atoi(envRepIntr)
		if err != nil {
			return err
		}
		flagRepIntr = val
	}
	if envPollIntr, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		val, err := strconv.Atoi(envPollIntr)
		if err != nil {
			return err
		}
		flagPollIntr = val
	}
	return nil
}
