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
	if targets[0].SendBody != true {
		t.Errorf("expected SendBody default true, got %v", targets[0].SendBody)
	}
	if targets[0].StopOnMatch != false {
		t.Errorf("expected StopOnMatch default false, got %v", targets[0].StopOnMatch)
	}
}

func TestResolve_StopOnMatch(t *testing.T) {
	sendBodyTrue := true
	cfg := &config.Config{Routes: []config.RouteConfig{
		{Name: "stop_route", Selector: "true", WebhookURL: "http://example/stop", Method: "POST", StopOnMatch: true, SendBody: &sendBodyTrue},
		{Name: "should_not_match", Selector: "true", WebhookURL: "http://example/other", Method: "POST"},
	}}
	res, err := NewResolver(cfg)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	event := map[string]interface{}{"content": map[string]interface{}{"body": "test"}}
	targets, err := res.Resolve(event)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target (stop_on_match should prevent others), got %d", len(targets))
	}
	if targets[0].Name != "stop_route" {
		t.Errorf("expected stop_route, got %s", targets[0].Name)
	}
	if !targets[0].StopOnMatch {
		t.Errorf("expected StopOnMatch true, got false")
	}
}

func TestResolve_SendBody(t *testing.T) {
	sendBodyFalse := false
	cfg := &config.Config{Routes: []config.RouteConfig{
		{Name: "no_body", Selector: "true", WebhookURL: "http://example/no", Method: "POST", SendBody: &sendBodyFalse},
	}}
	res, err := NewResolver(cfg)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}

	event := map[string]interface{}{"content": map[string]interface{}{"body": "test"}}
	targets, err := res.Resolve(event)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].SendBody != false {
		t.Errorf("expected SendBody false, got %v", targets[0].SendBody)
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
