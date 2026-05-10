// OMEGA-BRAIN — AI Orchestrator for CRYPTOBREACH

package cryptobreach

import (
	"fmt"
	"strings"
	"time"
)

// CryptoOmega coordinates all CRYPTOBREACH agents
type CryptoOmega struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewCryptoOmega(bus *EventBus, state *SharedState) *CryptoOmega {
	return &CryptoOmega{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 200)}
}

func (o *CryptoOmega) Name() string  { return "OMEGA-BRAIN" }
func (o *CryptoOmega) Status() string { return "online" }

func (o *CryptoOmega) Start() {
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

func (o *CryptoOmega) Stop() { close(o.stopCh) }

func (o *CryptoOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT":
		o.orchestrate(msg.Data)
	case "AUTO_ATTACK":
		o.autonomousChain(msg.Data)
	default:
		fmt.Printf("\033[35m[OMEGA]\033[0m %s -> %s: %s | %s\n", msg.From, msg.To, msg.Type, msg.Data)
	}
}

func (o *CryptoOmega) orchestrate(target string) {
	fmt.Println()
	fmt.Println("\033[35m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Println("\033[35m  OMEGA-BRAIN: CRYPTO ATTACK CHAIN INITIATED                  \033[0m")
	fmt.Println("\033[35m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Printf("\033[33m  Target: %s\033[0m\n", target)
	fmt.Println()

	fmt.Println("\033[35m[PHASE 1/6]\033[0m HASH-BREAKER: Hash identification & cracking")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "HASH-BREAKER", Type: "CRACK", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 2/6]\033[0m CERT-HUNTER: Certificate analysis")
	if strings.HasPrefix(target, "http") {
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "CERT-HUNTER", Type: "SCAN_CERT", Data: target})
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 3/6]\033[0m TLS-PHANTOM: TLS downgrade & cipher analysis")
	if strings.HasPrefix(target, "http") {
		host := strings.TrimPrefix(target, "https://")
		host = strings.TrimPrefix(host, "http://")
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "TLS-PHANTOM", Type: "ATTACK_TLS", Data: host})
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 4/6]\033[0m KEY-EXTRACT: Key recovery attempts")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "KEY-EXTRACT", Type: "EXTRACT_KEY", Data: target})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 5/6]\033[0m QUANTUM-SHADOW: Quantum vulnerability check")
	if strings.HasPrefix(target, "http") {
		host := strings.TrimPrefix(target, "https://")
		host = strings.TrimPrefix(host, "http://")
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "QUANTUM-SHADOW", Type: "QUANTUM_CHECK", Data: host})
	}
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 6/6]\033[0m OMEGA: Report generation")
	o.generateReport()

	fmt.Println()
	fmt.Println("\033[32m[OMEGA] Cryptographic attack chain complete.\033[0m")
	fmt.Println()
}

func (o *CryptoOmega) autonomousChain(target string) {
	fmt.Printf("\033[35m[OMEGA] Analyzing crypto target: %s\033[0m\n", target)
	o.orchestrate(target)
}

func (o *CryptoOmega) generateReport() {
	o.state.mu.RLock()
	defer o.state.mu.RUnlock()

	fmt.Println()
	fmt.Println("\033[35m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[35m║\033[33m      C R Y P T O B R E A C H   R E P O R T                   \033[35m║\033[0m")
	fmt.Println("\033[35m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Printf("\033[35m║\033[0m  Cracked Hashes:  \033[33m%-36d\033[35m║\033[0m\n", len(o.state.CrackResults))
	fmt.Printf("\033[35m║\033[0m  TLS Issues:      \033[31m%-36d\033[35m║\033[0m\n", len(o.state.TLSIssues))
	fmt.Printf("\033[35m║\033[0m  Cert Issues:     \033[33m%-36d\033[35m║\033[0m\n", len(o.state.Certs))
	fmt.Printf("\033[35m║\033[0m  Quantum Vulns:   \033[31m%-36d\033[35m║\033[0m\n", len(o.state.QuantumVulns))
	fmt.Printf("\033[35m║\033[0m  Weak Keys:       \033[33m%-36d\033[35m║\033[0m\n", len(o.state.Keys))
	fmt.Println("\033[35m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()
}
