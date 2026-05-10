// ETHERBREACH - Main Engine Coordinator
// Boots all 5 agents, manages event bus, starts autonomous attack

package etherbreach

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/internal/ai"
	"github.com/nexus-void/nexus-void/pkg/wireless"
)

// EtherBreach is the main autonomous WiFi attack engine
type EtherBreach struct {
	Bus     *EventBus
	State   *SharedState
	AI      *ai.AIClient
	Specter *wireless.WirelessSpecter
	OUI     *OUIDatabase
	Agents  map[string]Agent
	StopCh  chan struct{}
	mu      sync.RWMutex
}

// Agent interface for all swarm members
type Agent interface {
	Name() string
	Start(bus *EventBus, state *SharedState)
	Stop()
	Status() string
}

// New creates the full EtherBreach engine with all 5 agents
func New(iface string) (*EtherBreach, error) {
	if iface == "" {
		iface = "wlan0"
	}

	bus := NewEventBus()
	state := NewSharedState()
	state.Adapter = iface
	state.MonitorIface = iface + "mon"

	// Create output dirs
	for _, dir := range []string{state.HandshakesDir, state.WordlistsDir, state.CaptiveDir, "/tmp/nexus-void"} {
		os.MkdirAll(dir, 0755)
	}

	client := ai.NewClient()
	client.LoadAPIKeys()

	// Create engine
	eb := &EtherBreach{
		Bus:     bus,
		State:   state,
		AI:      client,
		Specter: wireless.NewWirelessSpecter(iface),
		OUI:     NewOUIDatabase(),
		Agents:  make(map[string]Agent),
		StopCh:  make(chan struct{}),
	}

	// Instantiate all 5 agents
	eb.Agents["ALPHA"] = NewAlphaScanner()
	eb.Agents["BETA"] = NewBetaBreaker()
	eb.Agents["GAMMA"] = NewGammaPhantom()
	eb.Agents["DELTA"] = NewDeltaShadow()
	eb.Agents["OMEGA"] = NewOmegaBrain(client)

	return eb, nil
}

// Start launches all agents and begins autonomous operation
func (eb *EtherBreach) Start() {
	fmt.Println()
	fmt.Println("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[35m           E T H E R B R E A C H   E N G I N E               \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[32m      AUTONOMOUS MULTI-AGENT WiFi PENETRATION SYSTEM           \033[36m║\033[0m")
	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[36m║\033[33m  5 Agents | 15 Features | Real Tools | AI-Driven              \033[36m║\033[0m")
	fmt.Println("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	// Step 1: ALPHA detects adapter + enables monitor mode
	fmt.Println("[OMEGA] Initiating agent swarm...")
	fmt.Println("[ALPHA] Detecting wireless adapter...")

	adapter := eb.detectAdapter()
	if adapter == "" {
		fmt.Println("\033[31m[!] No WiFi adapter detected. Cannot proceed.\033[0m")
		fmt.Println("[!] Required: USB WiFi adapter with monitor mode support")
		return
	}

	eb.State.mu.Lock()
	eb.State.Adapter = adapter
	eb.State.MonitorIface = adapter + "mon"
	eb.State.mu.Unlock()

	fmt.Printf("[ALPHA] Adapter found: %s\n", adapter)
	fmt.Println("[ALPHA] Enabling monitor mode...")

	if err := eb.enableMonitorMode(adapter); err != nil {
		fmt.Printf("\033[31m[!] Monitor mode failed: %v\033[0m\n", err)
		fmt.Println("[!] Try: sudo airmon-ng start", adapter)
		return
	}

	fmt.Printf("[ALPHA] Monitor mode active on %s\n", eb.State.MonitorIface)

	// Step 2: Start all agents on the event bus
	for name, agent := range eb.Agents {
		go agent.Start(eb.Bus, eb.State)
		fmt.Printf("[OMEGA] Agent %s deployed\n", name)
	}

	// Step 3: OMEGA triggers autonomous scan
	time.Sleep(1 * time.Second)
	eb.Bus.Broadcast(AgentMessage{
		From:   "OMEGA",
		To:     "ALPHA",
		Type:   "SCAN_START",
		Data:   "full_scan",
		Target: nil,
	})

	// Step 4: Listen for completion or stop signal
	<-eb.StopCh
	fmt.Println("[OMEGA] Engine shutdown complete.")
}

// Stop gracefully halts all agents
func (eb *EtherBreach) Stop() {
	for _, agent := range eb.Agents {
		agent.Stop()
	}
	select {
	case <-eb.StopCh:
		// Already closed
	default:
		close(eb.StopCh)
	}
}

// Close is an alias for Stop for external API consistency
func (eb *EtherBreach) Close() {
	eb.Stop()
}

// detectAdapter finds the best WiFi interface
func (eb *EtherBreach) detectAdapter() string {
	// Method 1: iw dev
	if out, err := exec.Command("iw", "dev").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Interface") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
			}
		}
	}

	// Method 2: ip link show
	if out, err := exec.Command("ip", "link", "show").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "wlan") || strings.Contains(line, "wlx") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					iface := strings.Trim(parts[1], ":")
					return iface
				}
			}
		}
	}

	// Method 3: lsusb for known WiFi chipset
	if out, err := exec.Command("lsusb").CombinedOutput(); err == nil {
		if strings.Contains(string(out), "RTL88") || strings.Contains(string(out), "Atheros") ||
			strings.Contains(string(out), "MediaTek") || strings.Contains(string(out), "Ralink") ||
			strings.Contains(string(out), "802.11") {
			return "wlan0" // Best guess
		}
	}

	return ""
}

// enableMonitorMode runs airmon-ng to enable monitor mode
func (eb *EtherBreach) enableMonitorMode(iface string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("monitor mode only supported on Linux")
	}

	// Kill interfering processes
	exec.Command("sudo", "airmon-ng", "check", "kill").Run()
	time.Sleep(500 * time.Millisecond)

	// Enable monitor mode
	cmd := exec.Command("sudo", "airmon-ng", "start", iface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Verify
	if _, err := exec.LookPath("iwconfig"); err == nil {
		if out, err := exec.Command("iwconfig").CombinedOutput(); err == nil {
			if !strings.Contains(string(out), iface+"mon") {
				return fmt.Errorf("monitor interface not created")
			}
		}
	}

	return nil
}

// ScanNetworks performs a real WiFi scan and returns discovered targets
func (eb *EtherBreach) ScanNetworks() []*NetworkTarget {
	fmt.Println("[ETHERBREACH] Scanning for wireless networks...")

	results := eb.Specter.ScanNetworks()
	var targets []*NetworkTarget

	for _, r := range results {
		target := &NetworkTarget{
			SSID:     r.SSID,
			BSSID:    r.BSSID,
			Channel:  r.Channel,
			Security: r.Security,
			Signal:   -50,
		}
		target.Brand = eb.OUI.Lookup(r.BSSID)
		targets = append(targets, target)
	}

	eb.State.mu.Lock()
	eb.State.Targets = targets
	eb.State.mu.Unlock()

	fmt.Printf("[ETHERBREACH] Found %d networks\n", len(targets))
	return targets
}

// RunAutoAttack starts the full autonomous chain
func (eb *EtherBreach) RunAutoAttack() {
	eb.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "AUTO_ATTACK_START",
		Data: "scan_and_attack",
	})
}

// RunRadarMode starts the TUI radar visualization
func (eb *EtherBreach) RunRadarMode() {
	// Will be implemented in radar_tui.go
	fmt.Println("[OMEGA] Starting Radar Mode TUI...")
	// TODO: launch bubbletea radar
}

// RunWarRoom starts the active attack monitoring TUI
func (eb *EtherBreach) RunWarRoom() {
	fmt.Println("[OMEGA] Starting War Room TUI...")
	// TODO: launch bubbletea war room
}

// GetResults returns current attack results
func (eb *EtherBreach) GetResults() *AttackResults {
	eb.State.mu.RLock()
	defer eb.State.mu.RUnlock()
	return eb.State.Results
}

// AddFinding appends a finding and broadcasts it
func (eb *EtherBreach) AddFinding(f string) {
	eb.State.mu.Lock()
	eb.State.Results.Findings = append(eb.State.Results.Findings, f)
	eb.State.mu.Unlock()

	eb.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "FINDING",
		Data: f,
	})
}

// EnsurePackage checks and warns about missing tools
func (eb *EtherBreach) EnsurePackage(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Tool missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

// Log broadcasts a log message to all agents
func (eb *EtherBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	eb.Bus.Broadcast(AgentMessage{
		From: source,
		To:   "ALL",
		Type: "LOG",
		Data: msg,
	})
}
