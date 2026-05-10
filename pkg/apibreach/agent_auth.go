package apibreach

import (
	"fmt"
	"os/exec"
	"strings"
)

type AuthBypassAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewAuthBypassAgent(bus *EventBus, state *SharedState) *AuthBypassAgent {
	return &AuthBypassAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *AuthBypassAgent) Name() string  { return "AUTH-BYPASS" }
func (a *AuthBypassAgent) Status() string { return "online" }
func (a *AuthBypassAgent) Start() {
	a.bus.Subscribe("AUTH-BYPASS", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *AuthBypassAgent) Stop() { close(a.stopCh) }
func (a *AuthBypassAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "TEST": a.testAuth(msg.Data)
	}
}
func (a *AuthBypassAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "AUTH-BYPASS", To: "ALL", Type: "LOG", Data: msg})
}

func (a *AuthBypassAgent) testAuth(url string) {
	a.broadcast(fmt.Sprintf("[AUTH] Testing %s", url))
	// jwt_tool for JWT weakness
	if out, err := exec.Command("jwt_tool", url, "-t").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "vulnerable") {
			a.broadcast("[CRITICAL] JWT auth bypass possible!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, APIVuln{Type: "JWT-Bypass", Endpoint: url, Confidence: 0.90})
			a.state.mu.Unlock()
		}
	}
}
