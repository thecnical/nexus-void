package webbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// XSSAgent hunts cross-site scripting
type XSSAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewXSSAgent(bus *EventBus, state *SharedState) *XSSAgent {
	return &XSSAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *XSSAgent) Name() string  { return "XSS-HUNTER" }
func (a *XSSAgent) Status() string { return "online" }
func (a *XSSAgent) Start() {
	a.bus.Subscribe("XSS-HUNTER", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *XSSAgent) Stop() { close(a.stopCh) }
func (a *XSSAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN": a.scanXSS(msg.Data)
	}
}
func (a *XSSAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "XSS-HUNTER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *XSSAgent) scanXSS(url string) {
	a.broadcast(fmt.Sprintf("[XSS] Hunting on %s", url))

	// dalfox
	if out, err := exec.Command("dalfox", "url", url, "-b", "xss.report").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "[VULN]") {
			a.broadcast("[CRITICAL] XSS vulnerability found!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, Vulnerability{Type: "XSS", URL: url, Confidence: 0.95})
			a.state.mu.Unlock()
		}
	}

	// XSStrike
	if out, err := exec.Command("python3", "XSStrike.py", "-u", url).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "vulnerable") {
			a.broadcast("[CRITICAL] XSStrike confirmed XSS!")
		}
	}
}
