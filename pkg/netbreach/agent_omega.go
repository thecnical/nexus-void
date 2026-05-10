// OMEGA-BRAIN — AI Orchestrator for NETBREACH

package netbreach

import (
	"fmt"
	"strings"
	"time"
)

// NetOmega coordinates all NETBREACH agents
type NetOmega struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewNetOmega(bus *EventBus, state *SharedState) *NetOmega {
	return &NetOmega{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 200)}
}

func (o *NetOmega) Name() string  { return "OMEGA-BRAIN" }
func (o *NetOmega) Status() string { return "online" }

func (o *NetOmega) Start() {
	o.bus.Subscribe("OMEGA", o.msgCh)
	for {
		select {
		case msg := <-o.msgCh:
			o.Handle(msg)
		case <-o.stopCh:
			return
		}
	}
}

func (o *NetOmega) Stop() { close(o.stopCh) }

func (o *NetOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT":
		o.orchestrate(msg.Data)
	case "AUTO_ATTACK":
		o.autonomousChain(msg.Data)
	default:
		fmt.Printf("\033[36m[OMEGA]\033[0m %s -> %s: %s | %s\n", msg.From, msg.To, msg.Type, msg.Data)
	}
}

func (o *NetOmega) orchestrate(target string) {
	fmt.Println()
	fmt.Println("\033[31m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Println("\033[31m  OMEGA-BRAIN: NETWORK POST-EXPLOITATION CHAIN INITIATED     \033[0m")
	fmt.Println("\033[31m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Printf("\033[33m  Target: %s\033[0m\n", target)
	fmt.Println()

	fmt.Println("\033[35m[PHASE 1/6]\033[0m INFECT: Reconnaissance & payload generation")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "INFECT", Type: "RECON", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 2/6]\033[0m PIVOT: Lateral movement & protocol checks")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "PIVOT", Type: "PIVOT", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 3/6]\033[0m EXTRACT: Credential harvesting")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "EXTRACT", Type: "EXTRACT", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 4/6]\033[0m AD-PHANTOM: Active Directory takeover")
	if strings.Contains(target, ".") {
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "AD-PHANTOM", Type: "AD_ATTACK", Data: target})
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 5/6]\033[0m C2-CONTROL: Persistence & C2 deployment")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "C2-CONTROL", Type: "C2_DEPLOY", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 6/6]\033[0m OMEGA: Report generation")
	o.generateReport()

	fmt.Println()
	fmt.Println("\033[32m[OMEGA] Network post-exploitation chain complete.\033[0m")
	fmt.Println()
}

func (o *NetOmega) autonomousChain(target string) {
	fmt.Printf("\033[36m[OMEGA] Analyzing target: %s\033[0m\n", target)
	if strings.Contains(target, "/") {
		// CIDR range
		o.orchestrate(target)
	} else {
		o.orchestrate(target)
	}
}

func (o *NetOmega) generateReport() {
	o.state.mu.RLock()
	defer o.state.mu.RUnlock()

	fmt.Println()
	fmt.Println("\033[31m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[31m║\033[33m           N E T B R E A C H   R E P O R T                     \033[31m║\033[0m")
	fmt.Println("\033[31m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Printf("\033[31m║\033[0m  Sessions:       \033[33m%-40d\033[31m║\033[0m\n", len(o.state.Sessions))
	fmt.Printf("\033[31m║\033[0m  Credentials:   \033[31m%-40d\033[31m║\033[0m\n", len(o.state.Creds))
	fmt.Printf("\033[31m║\033[0m  AD Objects:    \033[33m%-40d\033[31m║\033[0m\n", len(o.state.ADObjects))
	fmt.Printf("\033[31m║\033[0m  Tunnels:       \033[32m%-40d\033[31m║\033[0m\n", len(o.state.Tunnels))
	fmt.Printf("\033[31m║\033[0m  Paths:         \033[33m%-40d\033[31m║\033[0m\n", len(o.state.Paths))
	fmt.Println("\033[31m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()
}
