// CERT-HUNTER — Certificate & TLS Attacks Agent
// testssl.sh, sslyze, certipy

package cryptobreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// CertHunter handles certificate analysis
type CertHunter struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewCertHunter(bus *EventBus, state *SharedState) *CertHunter {
	return &CertHunter{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *CertHunter) Name() string  { return "CERT-HUNTER" }
func (a *CertHunter) Status() string { return "online" }

func (a *CertHunter) Start() {
	a.bus.Subscribe("CERT-HUNTER", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *CertHunter) Stop() { close(a.stopCh) }

func (a *CertHunter) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN_CERT":
		a.scanCert(msg.Data)
	}
}

func (a *CertHunter) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "CERT-HUNTER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *CertHunter) scanCert(url string) {
	a.broadcast(fmt.Sprintf("[CERT] Scanning certificates on: %s", url))

	// testssl.sh comprehensive scan
	if out, err := exec.Command("testssl.sh", "--color", "0", "-U", "--VULNERABLE", url).CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "VULNERABLE") {
				a.broadcast(fmt.Sprintf("[CRITICAL] %s", line))
			} else if strings.Contains(line, "NOT ok") {
				a.broadcast(fmt.Sprintf("[WARN] %s", line))
			}
		}
		_ = output
	}

	// sslyze scan
	if out, err := exec.Command("sslyze", "--certinfo", "--heartbleed", "--robot", url).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "VULNERABLE") {
			a.broadcast("[CRITICAL] SSL/TLS vulnerability detected!")
		}
		if strings.Contains(output, "Heartbleed") {
			a.broadcast("[CRITICAL] Heartbleed vulnerability found!")
		}
		_ = output
	}
}
