package router

import (
	"testing"

	"github.com/yamatt/go-as-webhook/internal/config"
)

func TestResolve_CELSelectors(t *testing.T) {
	cfg := &config.Config{Routes: []config.RouteConfig{
		{Name: "alerts", Selector: "event.type == 'm.room.message' && event.content.body.contains('alert')", WebhookURL: "http://static.example/alerts", Method: "POST"},
		{Name: "notify", Selector: "event.content.body.contains('notify')", WebhookURL: "http://dyn.example/notify", Method: "POST"},
	}}
	res, err := NewResolver(cfg)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	event := map[string]interface{}{
		"type":    "m.room.message",
		"content": map[string]interface{}{"body": "please alert and notify team"},
	}

	targets, err := res.Resolve(event)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
}

func TestResolve_NoMatch(t *testing.T) {
	cfg := &config.Config{Routes: []config.RouteConfig{
		{Name: "alerts", Selector: "event.content.body.contains('alert')", WebhookURL: "http://static.example/alerts"},
	}}
	res, err := NewResolver(cfg)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	event := map[string]interface{}{"content": map[string]interface{}{"body": "no match"}}
	targets, err := res.Resolve(event)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(targets) != 0 {
		t.Fatalf("expected 0 targets, got %d", len(targets))
	}
}
