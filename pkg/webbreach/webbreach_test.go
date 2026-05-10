package webbreach

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
	state.Target.URL = "https://example.com"
	state.Vulnerabilities = append(state.Vulnerabilities, Vulnerability{Type: "XSS", URL: "https://example.com", Confidence: 0.9})
	state.mu.Unlock()

	state.mu.RLock()
	if len(state.Vulnerabilities) != 1 {
		t.Errorf("expected 1 vuln, got %d", len(state.Vulnerabilities))
	}
	state.mu.RUnlock()
}

func TestWebBreachNew(t *testing.T) {
	wb := New()
	if wb == nil {
		t.Fatal("expected non-nil WebBreach")
	}
	if wb.bus == nil {
		t.Error("expected non-nil bus")
	}
	if wb.state == nil {
		t.Error("expected non-nil state")
	}
}
