package netbreach

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

func TestNetBreachNew(t *testing.T) {
	nb := New()
	if nb == nil {
		t.Fatal("expected non-nil NetBreach")
	}
	if nb.Bus == nil {
		t.Error("expected non-nil bus")
	}
	if nb.State == nil {
		t.Error("expected non-nil state")
	}
}
