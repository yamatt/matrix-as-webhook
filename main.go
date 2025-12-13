package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to configuration file")
	port := flag.Int("port", 8008, "Port to listen on")
	flag.Parse()

	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Printf("Warning: Could not load config file: %v. Using defaults.", err)
		config = NewDefaultConfig()
	}

	server := NewAppServer(config)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting Matrix Application Server on %s", addr)

	if err := http.ListenAndServe(addr, server.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
