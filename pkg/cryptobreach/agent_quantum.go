// QUANTUM-SHADOW — Quantum-Vulnerability Detection Agent

package cryptobreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// QuantumShadow checks for quantum-vulnerable algorithms
type QuantumShadow struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewQuantumShadow(bus *EventBus, state *SharedState) *QuantumShadow {
	return &QuantumShadow{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *QuantumShadow) Name() string  { return "QUANTUM-SHADOW" }
func (a *QuantumShadow) Status() string { return "online" }

func (a *QuantumShadow) Start() {
	a.bus.Subscribe("QUANTUM-SHADOW", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *QuantumShadow) Stop() { close(a.stopCh) }

func (a *QuantumShadow) Handle(msg AgentMessage) {
	switch msg.Type {
	case "QUANTUM_CHECK":
		a.quantumCheck(msg.Data)
	}
}

func (a *QuantumShadow) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "QUANTUM-SHADOW", To: "ALL", Type: "LOG", Data: msg})
}

func (a *QuantumShadow) quantumCheck(target string) {
	a.broadcast(fmt.Sprintf("[QUANTUM] Checking quantum vulnerabilities on: %s", target))

	// Check for RSA keys < 2048 bits (vulnerable to Shor's algorithm)
	if out, err := exec.Command("openssl", "s_client", "-connect", target+":443",
		"-showcerts", "</dev/null", "2>/dev/null").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "RSA Public-Key") {
			// Extract key size
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "Public-Key") {
					a.broadcast(fmt.Sprintf("[QUANTUM] %s", line))
					if strings.Contains(line, "1024") || strings.Contains(line, "512") {
						a.broadcast("[CRITICAL] Weak RSA key - quantum vulnerable!")
						a.state.mu.Lock()
						a.state.QuantumVulns = append(a.state.QuantumVulns, "Weak RSA key")
						a.state.mu.Unlock()
					}
				}
			}
		}
		_ = output
	}

	// Check for ECC curves
	if out, err := exec.Command("openssl", "s_client", "-connect", target+":443",
		"-curves", "secp256r1", "</dev/null", "2>/dev/null").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "ECDSA") {
			a.broadcast("[QUANTUM] ECDSA detected - vulnerable to Shor's algorithm")
			a.state.mu.Lock()
			a.state.QuantumVulns = append(a.state.QuantumVulns, "ECDSA vulnerable to quantum")
			a.state.mu.Unlock()
		}
		_ = output
	}

	a.broadcast("[QUANTUM] Recommendation: Migrate to CRYSTALS-Kyber / CRYSTALS-Dilithium")
}
