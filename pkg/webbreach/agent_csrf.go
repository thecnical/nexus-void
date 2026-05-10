package webbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// CSRFAgent hunts cross-site request forgery
type CSRFAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewCSRFAgent(bus *EventBus, state *SharedState) *CSRFAgent {
	return &CSRFAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *CSRFAgent) Name() string  { return "CSRF-DEMON" }
func (a *CSRFAgent) Status() string { return "online" }
func (a *CSRFAgent) Start() {
	a.bus.Subscribe("CSRF-DEMON", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *CSRFAgent) Stop() { close(a.stopCh) }
func (a *CSRFAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN": a.scanCSRF(msg.Data)
	}
}
func (a *CSRFAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "CSRF-DEMON", To: "ALL", Type: "LOG", Data: msg})
}

func (a *CSRFAgent) scanCSRF(url string) {
	a.broadcast(fmt.Sprintf("[CSRF] Testing %s", url))

	// nikto for CSRF checks
	if out, err := exec.Command("nikto", "-h", url, "-C", "all").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "CSRF") || strings.Contains(output, "No CSRF") {
			a.broadcast("[WARN] CSRF token missing or weak!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, Vulnerability{Type: "CSRF", URL: url, Confidence: 0.80})
			a.state.mu.Unlock()
		}
	}
}
