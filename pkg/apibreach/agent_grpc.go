package apibreach

import (
	"fmt"
	"os/exec"
)

type GRPCAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewGRPCAgent(bus *EventBus, state *SharedState) *GRPCAgent {
	return &GRPCAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *GRPCAgent) Name() string  { return "GRPC-BREAKER" }
func (a *GRPCAgent) Status() string { return "online" }
func (a *GRPCAgent) Start() {
	a.bus.Subscribe("GRPC-BREAKER", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *GRPCAgent) Stop() { close(a.stopCh) }
func (a *GRPCAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "TEST": a.testGRPC(msg.Data)
	}
}
func (a *GRPCAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "GRPC-BREAKER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *GRPCAgent) testGRPC(url string) {
	a.broadcast(fmt.Sprintf("[GRPC] Probing %s", url))
	if out, err := exec.Command("grpcurl", "-plaintext", url, "list").CombinedOutput(); err == nil {
		a.broadcast(fmt.Sprintf("[GRPC] Services: %s", string(out)))
	}
}
