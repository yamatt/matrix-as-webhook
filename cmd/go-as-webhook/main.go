package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yamatt/go-as-webhook/internal/args"
	"github.com/yamatt/go-as-webhook/internal/config"
	"github.com/yamatt/go-as-webhook/internal/registration"
	"github.com/yamatt/go-as-webhook/internal/server"
)

func main() {
	cliArgs, err := args.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Handle generate-registration flag
	if cliArgs.GenerateRegistration != "" {
		if err := generateRegistrationFile(cliArgs); err != nil {
			log.Fatalf("Failed to generate registration: %v", err)
		}
		return
	}

	cfg, err := config.Load(cliArgs.ConfigPath)
	if err != nil {
		log.Printf("Warning: Could not load config file: %v. Using defaults.", err)
		cfg = config.NewDefault()
	}

	// print found config routes
	for _, route := range cfg.Routes {
		log.Printf("Loaded route: Name=%s", route.Name)
	}

	srv := server.NewAppServer(cfg)

	addr := fmt.Sprintf(":%d", cliArgs.Port)
	log.Printf("Starting Matrix Application Server on %s", addr)

	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// generateRegistrationFile creates and writes a registration file
func generateRegistrationFile(cliArgs args.Args) error {
	reg, err := registration.Generate(cliArgs.Server, cliArgs.AsToken)
	if err != nil {
		return err
	}

	if err := reg.WriteToFile(cliArgs.GenerateRegistration); err != nil {
		return err
	}

	fmt.Printf("Registration file generated at: %s\n", cliArgs.GenerateRegistration)
	fmt.Printf("Configuration:\n")
	fmt.Printf("  - Server URL: %s\n", reg.Url)
	fmt.Printf("  - AS Token: %s\n", reg.AsToken)
	fmt.Printf("  - HS Token: %s\n", reg.HsToken)
	// Namespaces are omitted unless configured separately

	return nil
}
