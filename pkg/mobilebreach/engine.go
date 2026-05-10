// MobileBreach Engine
// Coordinates APEX, GHOST, LANCE, SPECTRE, OVERMIND agents

package mobilebreach

import (
	"fmt"
	"os/exec"
	"time"
)

// MobileBreach is the main engine
type MobileBreach struct {
	Bus    *EventBus
	State  *SharedState
	Agents []Agent
	CVE    *CVEDatabase
	StopCh chan struct{}
}

func New() *MobileBreach {
	bus := NewEventBus()
	state := &SharedState{Results: &AttackResults{}}
	cve := NewCVEDatabase()

	mb := &MobileBreach{
		Bus:    bus,
		State:  state,
		CVE:    cve,
		StopCh: make(chan struct{}),
	}

	// Spawn 5 agents
	mb.Agents = []Agent{
		NewAndroidRooter(bus, state),
		NewIOSPhantom(bus, state),
		NewAPIBreaker(bus, state),
		NewBasebandHunter(bus, state),
		NewMobileOmega(bus, state, cve),
	}

	return mb
}

// Start runs the full autonomous mobile attack
func (mb *MobileBreach) Start() {
	fmt.Println()
	fmt.Println("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[35m         M O B I L E B R E A C H   E N G I N E               \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[32m    NEXT-GEN AUTONOMOUS MOBILE PENETRATION SYSTEM             \033[36m║\033[0m")
	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[36m║\033[33m  5 Agents | 15 Features | Real Tools | AI-Driven              \033[36m║\033[0m")
	fmt.Println("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	// Start all agents
	for _, agent := range mb.Agents {
		go agent.Start()
		fmt.Printf("[OVERMIND] Agent %s deployed\n", agent.Name())
	}

	time.Sleep(500 * time.Millisecond)

	// Trigger OMEGA to begin orchestration
	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "ALL",
		Type: "INIT",
		Data: "mobilebreach_start",
	})

	<-mb.StopCh
	fmt.Println("[OVERMIND] MobileBreach engine shutdown.")
}

// Close gracefully stops all agents
func (mb *MobileBreach) Close() {
	for _, agent := range mb.Agents {
		agent.Stop()
	}
	select {
	case <-mb.StopCh:
	default:
		close(mb.StopCh)
	}
}

// ScanAPK runs APEX reverse engineering on an APK
func (mb *MobileBreach) ScanAPK(path string) *AppInfo {
	mb.State.mu.Lock()
	mb.State.CurrentTarget = &MobileTarget{Type: "apk", Path: path}
	mb.State.mu.Unlock()

	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "APEX",
		Type: "APK_REVERSE",
		Data: path,
	})
	return mb.State.AppInfo
}

// ScanAPI runs LANCE recon on a mobile API
func (mb *MobileBreach) ScanAPI(url string) *APITarget {
	mb.State.mu.Lock()
	mb.State.CurrentTarget = &MobileTarget{Type: "api", APIBaseURL: url}
	mb.State.mu.Unlock()

	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "LANCE",
		Type: "API_RECON",
		Data: url,
	})
	return nil
}

// StartIMSI runs SPECTRE IMSI catcher
func (mb *MobileBreach) StartIMSI() {
	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "SPECTRE",
		Type: "IMSI_CATCHER",
		Data: "start",
	})
}

// StartPaging runs SPECTRE paging attack
func (mb *MobileBreach) StartPaging() {
	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "SPECTRE",
		Type: "PAGING_ATTACK",
		Data: "start",
	})
}

// StartESIM runs SPECTRE eSIM extraction
func (mb *MobileBreach) StartESIM() {
	mb.Bus.Broadcast(AgentMessage{
		From: "OVERMIND",
		To:   "SPECTRE",
		Type: "ESIM_EXTRACT",
		Data: "start",
	})
}

// EnsureTool checks if a tool is available
func (mb *MobileBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Tool missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

// Log broadcasts a log to all agents
func (mb *MobileBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	mb.Bus.Broadcast(AgentMessage{
		From: source,
		To:   "ALL",
		Type: "LOG",
		Data: msg,
	})
}

// AddFinding records a finding
func (mb *MobileBreach) AddFinding(f string) {
	mb.State.mu.Lock()
	mb.State.Results.Findings = append(mb.State.Results.Findings, f)
	mb.State.mu.Unlock()
	mb.Log("OVERMIND", f)
}
