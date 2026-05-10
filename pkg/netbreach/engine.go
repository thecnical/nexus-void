// NETBREACH Engine
// Coordinates 6 agents for full-spectrum network post-exploitation

package netbreach

import (
	"fmt"
	"os/exec"
)

// NetBreach is the main engine
type NetBreach struct {
	Bus    *EventBus
	State  *SharedState
	Agents []Agent
	StopCh chan struct{}
}

func New() *NetBreach {
	bus := NewEventBus()
	state := &SharedState{
		Sessions: make(map[string]*Session),
	}

	nb := &NetBreach{
		Bus:    bus,
		State:  state,
		StopCh: make(chan struct{}),
	}

	nb.Agents = []Agent{
		NewInfectAgent(bus, state),
		NewPivotAgent(bus, state),
		NewExtractAgent(bus, state),
		NewADPhantomAgent(bus, state),
		NewC2ControlAgent(bus, state),
		NewNetOmega(bus, state),
	}

	return nb
}

func (nb *NetBreach) Start(target string) {
	fmt.Println()
	fmt.Println("\033[31m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[31m║\033[33m         N E T B R E A C H   E N G I N E                        \033[31m║\033[0m")
	fmt.Println("\033[31m║\033[32m    NETWORK & POST-EXPLOITATION WEAPON SYSTEM                \033[31m║\033[0m")
	fmt.Println("\033[31m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[31m║\033[36m  6 Agents | AD Takeover | Lateral Move | C2 | Extraction   \033[31m║\033[0m")
	fmt.Println("\033[31m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	for _, agent := range nb.Agents {
		go agent.Start()
		fmt.Printf("[OMEGA] Agent %s deployed\n", agent.Name())
	}

	nb.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "INIT",
		Data: target,
	})

	<-nb.StopCh
	fmt.Println("[OMEGA] NETBREACH shutdown.")
}

func (nb *NetBreach) Close() {
	for _, agent := range nb.Agents {
		agent.Stop()
	}
	select {
	case <-nb.StopCh:
	default:
		close(nb.StopCh)
	}
}

func (nb *NetBreach) Recon(target string) {
	nb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "INFECT", Type: "RECON", Data: target})
}

func (nb *NetBreach) Pivot(target string) {
	nb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "PIVOT", Type: "PIVOT", Data: target})
}

func (nb *NetBreach) Extract(target string) {
	nb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "EXTRACT", Type: "EXTRACT", Data: target})
}

func (nb *NetBreach) ADAttack(target string) {
	nb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "AD-PHANTOM", Type: "AD_ATTACK", Data: target})
}

func (nb *NetBreach) C2Deploy(target string) {
	nb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "C2-CONTROL", Type: "C2_DEPLOY", Data: target})
}

func (nb *NetBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

func (nb *NetBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	nb.Bus.Broadcast(AgentMessage{From: source, To: "ALL", Type: "LOG", Data: msg})
}
