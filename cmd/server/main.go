package main

import (
	"github.com/volchkovski/go-practicum-metrics/internal/server"
	"log"
)

func main() {
	s := server.New()
	log.Fatal(s.Run())
}
