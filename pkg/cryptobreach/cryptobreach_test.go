package cryptobreach

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

func TestCryptoBreachNew(t *testing.T) {
	cb := New()
	if cb == nil {
		t.Fatal("expected non-nil CryptoBreach")
	}
	if cb.Bus == nil {
		t.Error("expected non-nil bus")
	}
}
