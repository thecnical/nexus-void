// KEY-EXTRACT — Key Recovery & Extraction Agent
// rsactftool, openssl

package cryptobreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// KeyExtract handles key recovery
type KeyExtract struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewKeyExtract(bus *EventBus, state *SharedState) *KeyExtract {
	return &KeyExtract{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *KeyExtract) Name() string  { return "KEY-EXTRACT" }
func (a *KeyExtract) Status() string { return "online" }

func (a *KeyExtract) Start() {
	a.bus.Subscribe("KEY-EXTRACT", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *KeyExtract) Stop() { close(a.stopCh) }

func (a *KeyExtract) Handle(msg AgentMessage) {
	switch msg.Type {
	case "EXTRACT_KEY":
		a.extractKey(msg.Data)
	}
}

func (a *KeyExtract) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "KEY-EXTRACT", To: "ALL", Type: "LOG", Data: msg})
}

func (a *KeyExtract) extractKey(file string) {
	a.broadcast(fmt.Sprintf("[KEY] Analyzing key material: %s", file))

	// Check RSA key with openssl
	if out, err := exec.Command("openssl", "rsa", "-in", file, "-check", "-noout").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "RSA key ok") {
			a.broadcast("[KEY] Valid RSA key detected")
		}
		_ = output
	}

	// Try RsaCtfTool for factorization
	if out, err := exec.Command("RsaCtfTool", "--attack", "all", "--key", file).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "private") {
			a.broadcast("[CRITICAL] Private key recovered via factorization!")
		}
		_ = output
	} else {
		// Try weak key attacks
		a.broadcast("[KEY] Attempting weak key attacks...")
		if out2, err2 := exec.Command("RsaCtfTool", "--attack", "smallq,factordb", "--key", file).CombinedOutput(); err2 == nil {
			output := string(out2)
			if strings.Contains(output, "private") {
				a.broadcast("[CRITICAL] Private key recovered!")
			}
			_ = output
		}
	}
}
