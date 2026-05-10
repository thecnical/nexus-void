// TLS-PHANTOM — TLS Downgrade & MITM Agent
// bettercap, sslstrip

package cryptobreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// TLSPhantom handles TLS attacks
type TLSPhantom struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewTLSPhantom(bus *EventBus, state *SharedState) *TLSPhantom {
	return &TLSPhantom{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *TLSPhantom) Name() string  { return "TLS-PHANTOM" }
func (a *TLSPhantom) Status() string { return "online" }

func (a *TLSPhantom) Start() {
	a.bus.Subscribe("TLS-PHANTOM", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *TLSPhantom) Stop() { close(a.stopCh) }

func (a *TLSPhantom) Handle(msg AgentMessage) {
	switch msg.Type {
	case "ATTACK_TLS":
		a.attackTLS(msg.Data)
	}
}

func (a *TLSPhantom) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "TLS-PHANTOM", To: "ALL", Type: "LOG", Data: msg})
}

func (a *TLSPhantom) attackTLS(url string) {
	a.broadcast(fmt.Sprintf("[TLS] Attacking TLS on: %s", url))

	// Check for weak protocols with openssl
	if out, err := exec.Command("openssl", "s_client", "-connect", url+":443",
		"-tls1", "-brief").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "TLSv1.0") || strings.Contains(output, "TLSv1.1") {
			a.broadcast("[CRITICAL] TLS 1.0/1.1 downgrade possible!")
			a.state.mu.Lock()
			a.state.TLSIssues = append(a.state.TLSIssues, "TLS 1.0/1.1 supported")
			a.state.mu.Unlock()
		}
		_ = output
	}

	// Check cipher suites
	if out, err := exec.Command("nmap", "--script", "ssl-enum-ciphers", "-p", "443", url).CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "weak") || strings.Contains(line, "NULL") ||
				strings.Contains(line, "EXPORT") || strings.Contains(line, "RC4") {
				a.broadcast(fmt.Sprintf("[CRITICAL] Weak cipher: %s", line))
			}
		}
		_ = output
	}
}
