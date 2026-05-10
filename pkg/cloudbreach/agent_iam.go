// IAM-ESCALATOR — Cloud Privilege Escalation Agent
// pacu, enumerate-iam, aws_pwn, cloudsplaining

package cloudbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// IAMEscalator handles privilege escalation
type IAMEscalator struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewIAMEscalator(bus *EventBus, state *SharedState) *IAMEscalator {
	return &IAMEscalator{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *IAMEscalator) Name() string  { return "IAM-ESCALATOR" }
func (a *IAMEscalator) Status() string { return "online" }

func (a *IAMEscalator) Start() {
	a.bus.Subscribe("IAM-ESCALATOR", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *IAMEscalator) Stop() { close(a.stopCh) }

func (a *IAMEscalator) Handle(msg AgentMessage) {
	switch msg.Type {
	case "ESCALATE":
		a.escalate(msg.Data)
	}
}

func (a *IAMEscalator) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "IAM-ESCALATOR", To: "ALL", Type: "LOG", Data: msg})
}

func (a *IAMEscalator) escalate(provider string) {
	a.broadcast(fmt.Sprintf("[IAM] Privilege escalation scan: %s", provider))

	if strings.ToLower(provider) == "aws" || provider == "all" {
		// pacu privesc scan
		a.broadcast("[IAM] Running pacu IAM privilege escalation scan...")
		if out, err := exec.Command("pacu", "--module", "iam__privesc_scan").CombinedOutput(); err == nil {
			output := string(out)
			if strings.Contains(output, "VULNERABLE") || strings.Contains(output, "PRIVESC") {
				a.broadcast("[CRITICAL] IAM privilege escalation path found!")
			}
			_ = output
		}

		// enumerate-iam
		if out, err := exec.Command("enumerate-iam", "--access-key", "AKIA...", "--secret-key", "...").CombinedOutput(); err == nil {
			a.broadcast("[IAM] Current permissions enumerated")
			_ = out
		}

		// cloudsplaining
		if out, err := exec.Command("cloudsplaining", "scan", "--input", "default").CombinedOutput(); err == nil {
			output := string(out)
			if strings.Contains(output, "HIGH") {
				a.broadcast("[CRITICAL] Dangerous IAM policies found!")
			}
			_ = output
		}
	}
}
