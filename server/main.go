package main

import (
	"log"
)

func main() {
	s := NewServer()
	controlAddr := ":7001"
	httpAddr := ":8080"

	log.Printf("Starting server...")
	log.Printf("Control/Data listener will start on %s", controlAddr)
	log.Printf("HTTP routing enabled on %s", httpAddr)

	// Start HTTP listener in background
	go s.listenHTTP(httpAddr)

	// Run main control and data listener
	if err := s.Run(controlAddr); err != nil {
		log.Fatal(err)
	}
}
