package args

import (
	"flag"
)

// Args represents parsed command-line arguments.
type Args struct {
	ConfigPath string
	Port       int
}

// Parse parses the provided arguments slice.
func Parse(rawArgs []string) (Args, error) {
	fs := flag.NewFlagSet("matrix-as-webhook", flag.ContinueOnError)

	var parsed Args
	fs.StringVar(&parsed.ConfigPath, "config", "config.toml", "Path to configuration file")
	fs.IntVar(&parsed.Port, "port", 8080, "Port to listen on")

	if err := fs.Parse(rawArgs); err != nil {
		return Args{}, err
	}

	return parsed, nil
}
