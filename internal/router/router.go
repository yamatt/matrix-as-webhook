package router

import (
	"encoding/json"
	"log"

	"github.com/google/cel-go/cel"
	"github.com/yamatt/go-as-webhook/internal/config"
)

type Target struct {
	Name   string
	URL    string
	Method string
}

type compiledRoute struct {
	conf config.RouteConfig
	prog cel.Program
}

// Resolver evaluates CEL selectors to pick webhook targets.
type Resolver struct {
	routes []compiledRoute
}

func NewResolver(cfg *config.Config) (*Resolver, error) {
	env, err := cel.NewEnv(
		cel.Variable("event", cel.DynType),
	)
	if err != nil {
		return nil, err
	}

	crs := make([]compiledRoute, 0, len(cfg.Routes))
	for _, rc := range cfg.Routes {
		ast, issues := env.Compile(rc.Selector)
		if issues != nil && issues.Err() != nil {
			return nil, issues.Err()
		}
		prog, err := env.Program(ast)
		if err != nil {
			return nil, err
		}
		crs = append(crs, compiledRoute{conf: rc, prog: prog})
	}
	return &Resolver{routes: crs}, nil
}

// Resolve returns targets for the given event (as struct or map).
func (r *Resolver) Resolve(event interface{}) ([]Target, error) {
	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	var js any
	if err := json.Unmarshal(b, &js); err != nil {
		return nil, err
	}

	var out []Target
	for _, rt := range r.routes {
		val, _, err := rt.prog.Eval(map[string]any{"event": js})
		if err != nil {
			log.Printf("Router: selector eval error for route '%s': %v", rt.conf.Name, err)
			continue
		}
		matched := false
		if b, ok := val.Value().(bool); ok {
			matched = b
		}
		if matched {
			log.Printf("Router: selector matched for route '%s' -> %s", rt.conf.Name, rt.conf.WebhookURL)
			m := rt.conf.Method
			if m == "" {
				m = "POST"
			}
			out = append(out, Target{Name: rt.conf.Name, URL: rt.conf.WebhookURL, Method: m})
		} else {
			log.Printf("Router: selector did not match for route '%s'", rt.conf.Name)
		}
	}
	return out, nil
}
