package cloudbreach

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

func TestCloudBreachNew(t *testing.T) {
	clb := New()
	if clb == nil {
		t.Fatal("expected non-nil CloudBreach")
	}
	if clb.Bus == nil {
		t.Error("expected non-nil bus")
	}
}
