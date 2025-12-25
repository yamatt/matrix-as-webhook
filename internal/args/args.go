package args

import (
	"flag"
)

// Args represents parsed command-line arguments.
type Args struct {
	ConfigPath           string
	Port                 int
	GenerateRegistration string
	Server               string
	AsToken              string
}

// Parse parses the provided arguments slice.
func Parse(rawArgs []string) (Args, error) {
	fs := flag.NewFlagSet("matrix-as-webhook", flag.ContinueOnError)

	var parsed Args
	fs.StringVar(&parsed.ConfigPath, "config", "config.toml", "Path to configuration file")
	fs.IntVar(&parsed.Port, "port", 8080, "Port to listen on")
	fs.StringVar(&parsed.GenerateRegistration, "generate-registration", "", "Generate registration.yaml file at this path and exit")
	fs.StringVar(&parsed.Server, "server", "http://localhost:8080", "Server address (e.g., http://localhost:8080 or https://app.example.com)")
	fs.StringVar(&parsed.AsToken, "as-token", "", "Application Service token for registration (if empty, generated)")

	if err := fs.Parse(rawArgs); err != nil {
		return Args{}, err
	}

	return parsed, nil
}
