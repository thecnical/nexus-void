package webbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// IDORAgent tests for insecure direct object references
type IDORAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewIDORAgent(bus *EventBus, state *SharedState) *IDORAgent {
	return &IDORAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *IDORAgent) Name() string  { return "IDOR-BREAKER" }
func (a *IDORAgent) Status() string { return "online" }
func (a *IDORAgent) Start() {
	a.bus.Subscribe("IDOR-BREAKER", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *IDORAgent) Stop() { close(a.stopCh) }
func (a *IDORAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN": a.scanIDOR(msg.Data)
	}
}
func (a *IDORAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "IDOR-BREAKER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *IDORAgent) scanIDOR(url string) {
	a.broadcast(fmt.Sprintf("[IDOR] Testing %s", url))

	// ffuf for ID enumeration
	if out, err := exec.Command("ffuf", "-u", url+"/FUZZ", "-w", "/usr/share/wordlists/numbers.txt", "-mc", "200").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Status: 200") {
			a.broadcast("[CRITICAL] IDOR detected — valid IDs accessible without auth!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, Vulnerability{Type: "IDOR", URL: url, Confidence: 0.85})
			a.state.mu.Unlock()
		}
	}
}
