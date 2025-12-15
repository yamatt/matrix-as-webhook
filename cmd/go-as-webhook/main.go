package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yamatt/go-as-webhook/internal/args"
	"github.com/yamatt/go-as-webhook/internal/config"
	"github.com/yamatt/go-as-webhook/internal/server"
)

func main() {
	cliArgs, err := args.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	cfg, err := config.Load(cliArgs.ConfigPath)
	if err != nil {
		log.Printf("Warning: Could not load config file: %v. Using defaults.", err)
		cfg = config.NewDefault()
	}

	srv := server.NewAppServer(cfg)

	addr := fmt.Sprintf(":%d", cliArgs.Port)
	log.Printf("Starting Matrix Application Server on %s", addr)

	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
