// CRYPTOBREACH Engine
// Coordinates 6 agents for cryptographic attacks

package cryptobreach

import (
	"fmt"
	"os/exec"
)

// CryptoBreach is the main engine
type CryptoBreach struct {
	Bus    *EventBus
	State  *SharedState
	Agents []Agent
	StopCh chan struct{}
}

func New() *CryptoBreach {
	bus := NewEventBus()
	state := &SharedState{}

	cb := &CryptoBreach{
		Bus:    bus,
		State:  state,
		StopCh: make(chan struct{}),
	}

	cb.Agents = []Agent{
		NewHashBreaker(bus, state),
		NewCertHunter(bus, state),
		NewTLSPhantom(bus, state),
		NewKeyExtract(bus, state),
		NewQuantumShadow(bus, state),
		NewCryptoOmega(bus, state),
	}

	return cb
}

func (cb *CryptoBreach) Start(target string) {
	fmt.Println()
	fmt.Println("\033[35m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[35m║\033[33m      C R Y P T O B R E A C H   E N G I N E                    \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[32m    CRYPTOGRAPHIC ATTACK WEAPON SYSTEM                        \033[35m║\033[0m")
	fmt.Println("\033[35m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[35m║\033[36m  6 Agents | Hash Crack | TLS Downgrade | Key Extract        \033[35m║\033[0m")
	fmt.Println("\033[35m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	for _, agent := range cb.Agents {
		go agent.Start()
		fmt.Printf("[OMEGA] Agent %s deployed\n", agent.Name())
	}

	cb.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "INIT",
		Data: target,
	})

	<-cb.StopCh
	fmt.Println("[OMEGA] CRYPTOBREACH shutdown.")
}

func (cb *CryptoBreach) Close() {
	for _, agent := range cb.Agents {
		agent.Stop()
	}
	select {
	case <-cb.StopCh:
	default:
		close(cb.StopCh)
	}
}

func (cb *CryptoBreach) CrackHash(hash string) {
	cb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "HASH-BREAKER", Type: "CRACK", Data: hash})
}

func (cb *CryptoBreach) ScanCert(url string) {
	cb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "CERT-HUNTER", Type: "SCAN_CERT", Data: url})
}

func (cb *CryptoBreach) AttackTLS(url string) {
	cb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "TLS-PHANTOM", Type: "ATTACK_TLS", Data: url})
}

func (cb *CryptoBreach) ExtractKey(file string) {
	cb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "KEY-EXTRACT", Type: "EXTRACT_KEY", Data: file})
}

func (cb *CryptoBreach) QuantumCheck(target string) {
	cb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "QUANTUM-SHADOW", Type: "QUANTUM_CHECK", Data: target})
}

func (cb *CryptoBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

func (cb *CryptoBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	cb.Bus.Broadcast(AgentMessage{From: source, To: "ALL", Type: "LOG", Data: msg})
}
