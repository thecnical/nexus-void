package apibreach

import (
	"testing"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	ch := make(chan AgentMessage, 10)
	bus.Subscribe("TEST", ch)

	bus.Broadcast(AgentMessage{From: "A", To: "TEST", Type: "PING", Data: "hello"})

	select {
	case msg := <-ch:
		if msg.Data != "hello" {
			t.Errorf("expected 'hello', got %q", msg.Data)
		}
	default:
		t.Error("expected message on channel")
	}
}

func TestSharedState(t *testing.T) {
	state := &SharedState{}
	state.mu.Lock()
	state.Target.BaseURL = "https://api.example.com"
	state.Vulnerabilities = append(state.Vulnerabilities, APIVuln{Type: "JWT-Bypass", Endpoint: "/api/auth", Confidence: 0.85})
	state.mu.Unlock()

	state.mu.RLock()
	if len(state.Vulnerabilities) != 1 {
		t.Errorf("expected 1 vuln, got %d", len(state.Vulnerabilities))
	}
	state.mu.RUnlock()
}

func TestAPIBreachNew(t *testing.T) {
	ab := New()
	if ab == nil {
		t.Fatal("expected non-nil APIBreach")
	}
	if ab.bus == nil {
		t.Error("expected non-nil bus")
	}
	if ab.state == nil {
		t.Error("expected non-nil state")
	}
}
