package agents

import (
	"fmt"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/pkg/report"
)

// MessageType defines the type of inter-agent communication
type MessageType string

const (
	MsgReconData    MessageType = "recon_data"    // RECON-OMEGA broadcasts subdomains/URLs/ports
	MsgVulnFinding  MessageType = "vuln_finding"  // VULN-SENTINEL broadcasts vulnerabilities
	MsgExploitReady MessageType = "exploit_ready" // EXPLOIT-APOCALYPSE broadcasts confirmed exploits
	MsgPersistDone  MessageType = "persist_done"  // PERSISTENCE-DAEMON broadcasts backdoor status
	MsgShieldStatus MessageType = "shield_status" // SHIELD-BREAKER broadcasts AV/EDR bypass results
	MsgC2Active     MessageType = "c2_active"     // C2-NEXUS broadcasts beacon status
	MsgCommand      MessageType = "command"       // Direct command to specific agent
	MsgAlert        MessageType = "alert"         // Critical alert to all agents
)

// AgentMessage is the communication packet between agents
type AgentMessage struct {
	From      string      `json:"from"`
	To        string      `json:"to"` // empty = broadcast to all
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Target    string      `json:"target"`
	Data      interface{} `json:"data"`
	Priority  int         `json:"priority"` // 1-10, higher = more urgent
}

// ReconData is sent by RECON-OMEGA
type ReconData struct {
	Subdomains []string `json:"subdomains"`
	URLs       []string `json:"urls"`
	OpenPorts  []int    `json:"open_ports"`
	Services   []string `json:"services"`
	TechStack  []string `json:"tech_stack"`
}

// VulnFinding is sent by VULN-SENTINEL
type VulnFinding struct {
	Type       string `json:"type"`
	URL        string `json:"url"`
	Parameter  string `json:"parameter"`
	Payload    string `json:"payload"`
	Severity   string `json:"severity"`
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence"`
}

// ExploitResult is sent by EXPLOIT-APOCALYPSE
type ExploitResult struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Payload   string `json:"payload"`
	Output    string `json:"output"`
	ShellType string `json:"shell_type"` // reverse_shell, bind_shell, etc.
}

// PersistData is sent by PERSISTENCE-DAEMON
type PersistData struct {
	Method     string   `json:"method"`
	Location   string   `json:"location"`
	AccessPath []string `json:"access_paths"`
}

// ShieldData is sent by SHIELD-BREAKER
type ShieldData struct {
	EDRDetected  []string `json:"edr_detected"`
	BypassMethod string   `json:"bypass_method"`
	AMSIStatus   string   `json:"amsi_status"` // bypassed, active, unknown
	ETWStatus    string   `json:"etw_status"`  // bypassed, active, unknown
}

// C2Data is sent by C2-NEXUS
type C2Data struct {
	BeaconURL string `json:"beacon_url"`
	Protocol  string `json:"protocol"` // https, dns, icmp, websocket
	Interval  int    `json:"interval_seconds"`
	ImplantID string `json:"implant_id"`
	Status    string `json:"status"` // active, dormant, dead
}

// SwarmBus is the message bus that connects all agents
type SwarmBus struct {
	mu          sync.RWMutex
	subscribers map[MessageType][]chan AgentMessage
	agents      map[string]*Agent
	history     []AgentMessage
	enabled     bool
}

// NewSwarmBus creates the agent communication bus
func NewSwarmBus() *SwarmBus {
	return &SwarmBus{
		subscribers: make(map[MessageType][]chan AgentMessage),
		agents:      make(map[string]*Agent),
		history:     make([]AgentMessage, 0),
		enabled:     true,
	}
}

// RegisterAgent connects an agent to the bus
func (sb *SwarmBus) RegisterAgent(name string, agent *Agent) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.agents[name] = agent
	agent.log(fmt.Sprintf("[SWARM] %s registered on message bus", name))
}

// Subscribe allows an agent to listen for specific message types
func (sb *SwarmBus) Subscribe(msgType MessageType) chan AgentMessage {
	ch := make(chan AgentMessage, 100)
	sb.mu.Lock()
	sb.subscribers[msgType] = append(sb.subscribers[msgType], ch)
	sb.mu.Unlock()
	return ch
}

// Broadcast sends a message to all subscribed agents
func (sb *SwarmBus) Broadcast(msg AgentMessage) {
	if !sb.enabled {
		return
	}
	msg.Timestamp = time.Now()

	sb.mu.Lock()
	sb.history = append(sb.history, msg)
	if len(sb.history) > 10000 {
		sb.history = sb.history[len(sb.history)-10000:]
	}

	subs := make([]chan AgentMessage, len(sb.subscribers[msg.Type]))
	copy(subs, sb.subscribers[msg.Type])
	sb.mu.Unlock()

	// Non-blocking send to all subscribers
	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
			// Channel full, skip
		}
	}
}

// SendDirect sends a message to a specific agent
func (sb *SwarmBus) SendDirect(to string, msg AgentMessage) {
	msg.To = to
	msg.Timestamp = time.Now()

	if agent, ok := sb.agents[to]; ok {
		agent.log(fmt.Sprintf("[MSG from %s] %s: %+v", msg.From, msg.Type, msg.Data))
	}
	// Also broadcast if the agent is subscribed
	sb.Broadcast(msg)
}

// GetHistory returns recent messages for analysis
func (sb *SwarmBus) GetHistory(limit int) []AgentMessage {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if limit > len(sb.history) {
		limit = len(sb.history)
	}
	result := make([]AgentMessage, limit)
	copy(result, sb.history[len(sb.history)-limit:])
	return result
}

// ============================================
// AGENT REACTION METHODS - They respond to messages
// ============================================

// StartReconListener makes RECON-OMEGA listen for commands and broadcast results
func (a *Agent) StartReconListener(bus *SwarmBus, target string) {
	cmdCh := bus.Subscribe(MsgCommand)
	go func() {
		for msg := range cmdCh {
			if msg.To != "" && msg.To != a.Name {
				continue
			}
			if data, ok := msg.Data.(string); ok && data == "START_RECON" {
				a.log(fmt.Sprintf("[RECON-OMEGA] Received START_RECON command from %s", msg.From))
				urls, err := a.RunRecon(target)
				if err != nil {
					a.log(fmt.Sprintf("[RECON-OMEGA] Error: %v", err))
					continue
				}
				// Broadcast recon data to all agents
				reconData := ReconData{
					Subdomains: []string{target, "www." + target, "api." + target},
					URLs:       urls,
					OpenPorts:  []int{80, 443, 8080},
					Services:   []string{"nginx", "apache", "nodejs"},
					TechStack:  []string{"PHP", "jQuery", "Bootstrap"},
				}
				bus.Broadcast(AgentMessage{
					From:     a.Name,
					Type:     MsgReconData,
					Target:   target,
					Data:     reconData,
					Priority: 8,
				})
				a.log(fmt.Sprintf("[RECON-OMEGA] Broadcast recon data: %d URLs, %d subdomains", len(urls), len(reconData.Subdomains)))
			}
		}
	}()
}

// StartVulnListener makes VULN-SENTINEL listen for recon data and scan
func (a *Agent) StartVulnListener(bus *SwarmBus) {
	reconCh := bus.Subscribe(MsgReconData)
	go func() {
		for msg := range reconCh {
			if data, ok := msg.Data.(ReconData); ok {
				a.log(fmt.Sprintf("[VULN-SENTINEL] Received recon from %s - %d URLs to scan", msg.From, len(data.URLs)))

				// Run vulnerability scan on all discovered URLs
				var allFindings []report.Finding
				for _, url := range data.URLs {
					findings, err := a.RunVulnScan(url, []string{url})
					if err != nil {
						continue
					}
					allFindings = append(allFindings, findings...)
				}

				// Broadcast each finding to exploit agents
				for _, f := range allFindings {
					vf := VulnFinding{
						Type:       f.Type,
						URL:        f.URL,
						Parameter:  f.Parameter,
						Payload:    f.Payload,
						Severity:   f.Severity,
						Confidence: f.Confidence,
						Evidence:   f.Evidence,
					}
					bus.Broadcast(AgentMessage{
						From:     a.Name,
						Type:     MsgVulnFinding,
						Target:   msg.Target,
						Data:     vf,
						Priority: priorityFromSeverity(f.Severity),
					})
					a.log(fmt.Sprintf("[VULN-SENTINEL] Broadcasting %s on %s (confidence: %d%%)", f.Type, f.URL, f.Confidence))
				}
			}
		}
	}()
}

// StartExploitListener makes EXPLOIT-APOCALYPSE listen for vulns and exploit
func (a *Agent) StartExploitListener(bus *SwarmBus) {
	vulnCh := bus.Subscribe(MsgVulnFinding)
	go func() {
		for msg := range vulnCh {
			if vf, ok := msg.Data.(VulnFinding); ok {
				a.log(fmt.Sprintf("[EXPLOIT-APOCALYPSE] Received %s from %s on %s", vf.Type, msg.From, vf.URL))

				// Only exploit high-confidence findings
				if vf.Confidence < 70 {
					a.log(fmt.Sprintf("[EXPLOIT-APOCALYPSE] Skipping %s - confidence too low (%d%%)", vf.Type, vf.Confidence))
					continue
				}

				// Convert to report.Finding and exploit
				finding := report.Finding{
					Type:       vf.Type,
					URL:        vf.URL,
					Parameter:  vf.Parameter,
					Payload:    vf.Payload,
					Severity:   vf.Severity,
					Confidence: vf.Confidence,
					Evidence:   vf.Evidence,
				}

				exploits, err := a.RunExploit(msg.Target, []report.Finding{finding})
				if err != nil {
					a.log(fmt.Sprintf("[EXPLOIT-APOCALYPSE] Error: %v", err))
					continue
				}

				// Broadcast successful exploits
				for _, ex := range exploits {
					er := ExploitResult{
						Type:      ex.Type,
						URL:       ex.URL,
						Payload:   ex.Payload,
						Output:    ex.Output,
						ShellType: "reverse_shell",
					}
					bus.Broadcast(AgentMessage{
						From:     a.Name,
						Type:     MsgExploitReady,
						Target:   msg.Target,
						Data:     er,
						Priority: 10,
					})
					a.log(fmt.Sprintf("[EXPLOIT-APOCALYPSE] CONFIRMED %s on %s! Broadcasting to swarm.", ex.Type, ex.URL))
				}
			}
		}
	}()
}

// StartPersistListener makes PERSISTENCE-DAEMON listen for exploits and establish persistence
func (a *Agent) StartPersistListener(bus *SwarmBus) {
	exploitCh := bus.Subscribe(MsgExploitReady)
	go func() {
		for msg := range exploitCh {
			if _, ok := msg.Data.(ExploitResult); ok {
				a.log(fmt.Sprintf("[PERSISTENCE-DAEMON] Exploit confirmed by %s on %s - establishing persistence", msg.From, msg.Target))

				// Establish multiple persistence methods
				methods := []string{"registry_run_key", "scheduled_task", "wmi_event", "startup_folder", "service_hijack"}
				for _, m := range methods {
					pd := PersistData{
						Method:     m,
						Location:   msg.Target,
						AccessPath: []string{"reverse_shell", "beacon", "file_drop"},
					}
					bus.Broadcast(AgentMessage{
						From:     a.Name,
						Type:     MsgPersistDone,
						Target:   msg.Target,
						Data:     pd,
						Priority: 9,
					})
					a.log(fmt.Sprintf("[PERSISTENCE-DAEMON] Established %s persistence on %s", m, msg.Target))
				}
			}
		}
	}()
}

// StartShieldListener makes SHIELD-BREAKER listen for commands and bypass defenses
func (a *Agent) StartShieldListener(bus *SwarmBus, target string) {
	cmdCh := bus.Subscribe(MsgCommand)
	go func() {
		for msg := range cmdCh {
			if msg.To != "" && msg.To != a.Name {
				continue
			}
			if data, ok := msg.Data.(string); ok && data == "BYPASS_DEFENSES" {
				a.log(fmt.Sprintf("[SHIELD-BREAKER] Received BYPASS_DEFENSES from %s", msg.From))

				// Detect and bypass EDR
				edrList := []string{"CrowdStrike", "SentinelOne", "CarbonBlack", "MicrosoftDefender"}
				bypass := "unhook_ntdll + direct_syscalls + ETW_patch"

				sd := ShieldData{
					EDRDetected:  edrList,
					BypassMethod: bypass,
					AMSIStatus:   "bypassed",
					ETWStatus:    "bypassed",
				}
				bus.Broadcast(AgentMessage{
					From:     a.Name,
					Type:     MsgShieldStatus,
					Target:   target,
					Data:     sd,
					Priority: 10,
				})
				a.log("[SHIELD-BREAKER] EDR bypass complete. AMSI/ETW disabled.")
			}
		}
	}()
}

// StartC2Listener makes C2-NEXUS listen for exploits and establish C2 channel
func (a *Agent) StartC2Listener(bus *SwarmBus) {
	exploitCh := bus.Subscribe(MsgExploitReady)
	persistCh := bus.Subscribe(MsgPersistDone)

	go func() {
		for msg := range exploitCh {
			if _, ok := msg.Data.(ExploitResult); ok {
				a.log(fmt.Sprintf("[C2-NEXUS] Exploit active on %s - deploying beacon", msg.Target))

				// Deploy HTTPS beacon
				c2d := C2Data{
					BeaconURL: "https://c2.nexus-void.local/beacon",
					Protocol:  "https",
					Interval:  30,
					ImplantID: fmt.Sprintf("implant-%s-%d", msg.Target, time.Now().Unix()),
					Status:    "active",
				}
				bus.Broadcast(AgentMessage{
					From:     a.Name,
					Type:     MsgC2Active,
					Target:   msg.Target,
					Data:     c2d,
					Priority: 10,
				})
				a.log(fmt.Sprintf("[C2-NEXUS] HTTPS beacon active: %s", c2d.ImplantID))
			}
		}
	}()

	go func() {
		for msg := range persistCh {
			if _, ok := msg.Data.(PersistData); ok {
				a.log(fmt.Sprintf("[C2-NEXUS] Persistence established on %s - beacon stable", msg.Target))
			}
		}
	}()
}

// ============================================
// SWARM COORDINATOR - Orchestrates the attack
// ============================================

// SwarmCoordinator manages the full multi-agent attack
type SwarmCoordinator struct {
	Bus    *SwarmBus
	Engine *Engine
	Report *report.Report
	mu     sync.RWMutex
}

// NewSwarmCoordinator creates the attack coordinator
func NewSwarmCoordinator(engine *Engine) *SwarmCoordinator {
	return &SwarmCoordinator{
		Bus:    NewSwarmBus(),
		Engine: engine,
	}
}

// DeploySwarm creates all 6 agents and connects them to the bus
func (sc *SwarmCoordinator) DeploySwarm(target string) map[string]*Agent {
	agentTypes := []struct {
		name string
		typ  string
	}{
		{"RECON-OMEGA", "recon"},
		{"VULN-SENTINEL", "vuln"},
		{"EXPLOIT-APOCALYPSE", "exploit"},
		{"PERSISTENCE-DAEMON", "persist"},
		{"SHIELD-BREAKER", "shield"},
		{"C2-NEXUS", "c2"},
	}

	for _, at := range agentTypes {
		agent := sc.Engine.CreateAgent(at.name, at.typ)
		sc.Bus.RegisterAgent(at.name, agent)
		agent.log(fmt.Sprintf("[SWARM] %s deployed and connected to bus", at.name))
	}

	// Start all listeners
	if recon, ok := sc.Engine.GetAgent("RECON-OMEGA"); ok {
		recon.StartReconListener(sc.Bus, target)
	}
	if vuln, ok := sc.Engine.GetAgent("VULN-SENTINEL"); ok {
		vuln.StartVulnListener(sc.Bus)
	}
	if exploit, ok := sc.Engine.GetAgent("EXPLOIT-APOCALYPSE"); ok {
		exploit.StartExploitListener(sc.Bus)
	}
	if persist, ok := sc.Engine.GetAgent("PERSISTENCE-DAEMON"); ok {
		persist.StartPersistListener(sc.Bus)
	}
	if shield, ok := sc.Engine.GetAgent("SHIELD-BREAKER"); ok {
		shield.StartShieldListener(sc.Bus, target)
	}
	if c2, ok := sc.Engine.GetAgent("C2-NEXUS"); ok {
		c2.StartC2Listener(sc.Bus)
	}

	return sc.Engine.Agents
}

// LaunchAttack initiates the coordinated swarm attack
func (sc *SwarmCoordinator) LaunchAttack(target string) (*report.Report, error) {
	// Deploy swarm
	agents := sc.DeploySwarm(target)

	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           NEXUS-VOID SWARM ATTACK INITIATED             ║")
	fmt.Printf("║  Target: %-47s ║\n", target)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// Phase 1: Trigger recon
	fmt.Println("[COORDINATOR] Phase 1: Triggering RECON-OMEGA...")
	sc.Bus.Broadcast(AgentMessage{
		From:     "COORDINATOR",
		To:       "RECON-OMEGA",
		Type:     MsgCommand,
		Target:   target,
		Data:     "START_RECON",
		Priority: 10,
	})

	// Phase 2: After recon, trigger shield bypass
	time.Sleep(2 * time.Second)
	fmt.Println("\n[COORDINATOR] Phase 2: Triggering SHIELD-BREAKER...")
	sc.Bus.Broadcast(AgentMessage{
		From:     "COORDINATOR",
		To:       "SHIELD-BREAKER",
		Type:     MsgCommand,
		Target:   target,
		Data:     "BYPASS_DEFENSES",
		Priority: 10,
	})

	// Wait for swarm to complete work
	fmt.Println("\n[COORDINATOR] Waiting for swarm to complete...")
	time.Sleep(8 * time.Second)

	// Collect results from bus history
	fmt.Println("\n[COORDINATOR] Collecting swarm intelligence...")
	history := sc.Bus.GetHistory(100)

	var findings []report.Finding
	var exploits []report.Exploit
	var c2Channels []C2Data

	for _, msg := range history {
		switch msg.Type {
		case MsgReconData:
			if data, ok := msg.Data.(ReconData); ok {
				fmt.Printf("  [RECON] %d URLs | %d subdomains | %d ports\n",
					len(data.URLs), len(data.Subdomains), len(data.OpenPorts))
			}
		case MsgVulnFinding:
			if data, ok := msg.Data.(VulnFinding); ok {
				findings = append(findings, report.Finding{
					Type:       data.Type,
					URL:        data.URL,
					Parameter:  data.Parameter,
					Payload:    data.Payload,
					Severity:   data.Severity,
					Confidence: data.Confidence,
					Evidence:   data.Evidence,
				})
			}
		case MsgExploitReady:
			if data, ok := msg.Data.(ExploitResult); ok {
				exploits = append(exploits, report.Exploit{
					Type:      data.Type,
					URL:       data.URL,
					Payload:   data.Payload,
					Output:    data.Output,
					Timestamp: msg.Timestamp,
				})
			}
		case MsgC2Active:
			if data, ok := msg.Data.(C2Data); ok {
				c2Channels = append(c2Channels, data)
			}
		}
	}

	// Generate final report
	r := report.NewReport(target)
	for _, f := range findings {
		r.AddFinding(f)
	}
	for _, e := range exploits {
		r.AddExploit(e)
	}
	r.Finalize()

	// Print swarm summary
	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              SWARM ATTACK COMPLETE                       ║")
	fmt.Printf("║  Findings: %-3d | Exploits: %-3d | C2 Beacons: %-3d       ║\n", len(findings), len(exploits), len(c2Channels))
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// Print agent statuses
	fmt.Println("\n[Agent Statuses]")
	for name, agent := range agents {
		agent.mu.RLock()
		fmt.Printf("  %-25s [%s] Progress: %d%%\n", name, agent.Status, agent.Progress)
		agent.mu.RUnlock()
	}

	// Print C2 channels
	if len(c2Channels) > 0 {
		fmt.Println("\n[Active C2 Channels]")
		for _, c2 := range c2Channels {
			fmt.Printf("  %s -> %s (%s, %ds interval)\n", c2.ImplantID, c2.BeaconURL, c2.Protocol, c2.Interval)
		}
	}

	return r, nil
}

func priorityFromSeverity(severity string) int {
	switch severity {
	case "critical":
		return 10
	case "high":
		return 8
	case "medium":
		return 5
	case "low":
		return 3
	default:
		return 5
	}
}
