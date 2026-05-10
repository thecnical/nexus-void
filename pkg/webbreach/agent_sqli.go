package webbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// SQLiAgent hunts SQL injection
type SQLiAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewSQLiAgent(bus *EventBus, state *SharedState) *SQLiAgent {
	return &SQLiAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *SQLiAgent) Name() string  { return "SQLI-PHANTOM" }
func (a *SQLiAgent) Status() string { return "online" }
func (a *SQLiAgent) Start() {
	a.bus.Subscribe("SQLI-PHANTOM", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *SQLiAgent) Stop() { close(a.stopCh) }
func (a *SQLiAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN": a.scanSQLi(msg.Data)
	}
}
func (a *SQLiAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "SQLI-PHANTOM", To: "ALL", Type: "LOG", Data: msg})
}

func (a *SQLiAgent) scanSQLi(url string) {
	a.broadcast(fmt.Sprintf("[SQLI] Testing %s", url))

	// sqlmap
	if out, err := exec.Command("sqlmap", "-u", url, "--batch", "--level=2", "--risk=1").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "is vulnerable") {
			a.broadcast("[CRITICAL] SQL Injection confirmed!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, Vulnerability{Type: "SQLi", URL: url, Confidence: 0.95})
			a.state.mu.Unlock()
		}
	}
}
