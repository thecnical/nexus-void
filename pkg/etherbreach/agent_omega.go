// OMEGA-BRAIN Agent
// AI Decision Engine, Attack Chain Orchestration, Telemetry, TUI Control, Chat NL

package etherbreach

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nexus-void/nexus-void/internal/ai"
)

// OmegaBrain is the central coordinator AI
type OmegaBrain struct {
	bus    *EventBus
	state  *SharedState
	ai     *ai.AIClient
	stopCh chan struct{}
	status string
}

// NewOmegaBrain creates the brain agent
func NewOmegaBrain(client *ai.AIClient) *OmegaBrain {
	return &OmegaBrain{
		ai:     client,
		stopCh: make(chan struct{}),
		status: "idle",
	}
}

func (o *OmegaBrain) Name() string   { return "OMEGA" }
func (o *OmegaBrain) Status() string { return o.status }

func (o *OmegaBrain) Start(bus *EventBus, state *SharedState) {
	o.bus = bus
	o.state = state
	ch := bus.Subscribe(o.Name())

	for {
		select {
		case msg := <-ch:
			o.handleMessage(msg)
		case <-o.stopCh:
			return
		}
	}
}

func (o *OmegaBrain) Stop() {
	close(o.stopCh)
}

func (o *OmegaBrain) handleMessage(msg AgentMessage) {
	switch msg.Type {
	case "SCAN_RESULT":
		o.status = "analyzing"
		o.analyzeAndAttack(msg)
		o.status = "idle"
	case "ATTACK_SUCCESS":
		o.handleSuccess(msg)
	case "ATTACK_FAILURE":
		o.handleFailure(msg)
	case "AUTO_ATTACK_START":
		o.status = "orchestrating"
		o.orchestrateAutoAttack()
		o.status = "idle"
	case "USER_COMMAND":
		o.handleUserCommand(msg.Data, msg.Payload)
	case "FINDING":
		o.broadcast(fmt.Sprintf("[FINDING] %s", msg.Data))
	}
}

// analyzeAndAttack receives scan results and decides attack chain
func (o *OmegaBrain) analyzeAndAttack(msg AgentMessage) {
	var targets []*NetworkTarget
	if msg.Payload != nil {
		if t, ok := msg.Payload["targets"].([]*NetworkTarget); ok {
			targets = t
		}
	}

	if len(targets) == 0 {
		o.broadcast("[OMEGA] No targets found. Nothing to attack.")
		return
	}

	// AI Decision: rank networks by vulnerability
	target := o.pickBestTarget(targets)
	if target == nil {
		o.broadcast("[OMEGA] No suitable target found.")
		return
	}

	o.state.mu.Lock()
	o.state.CurrentTarget = target
	o.state.mu.Unlock()

	o.broadcast("")
	o.broadcast(fmt.Sprintf("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m"))
	o.broadcast(fmt.Sprintf("\033[36m║\033[35m           O M E G A   T A R G E T   S E L E C T E D        \033[36m║\033[0m"))
	o.broadcast(fmt.Sprintf("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m"))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  SSID:     %-49s \033[36m║\033[0m", target.SSID))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  BSSID:    %-49s \033[36m║\033[0m", target.BSSID))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  Security: %-49s \033[36m║\033[0m", target.Security))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  Brand:    %-49s \033[36m║\033[0m", target.Brand))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  Channel: %-49d \033[36m║\033[0m", target.Channel))
	o.broadcast(fmt.Sprintf("\033[36m║\033[33m  Signal:   %-49d dBm\033[36m║\033[0m", target.Signal))
	o.broadcast(fmt.Sprintf("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m"))
	o.broadcast("")

	// Build attack chain
	plan := o.buildAttackPlan(target)
	o.broadcast(fmt.Sprintf("[OMEGA] Attack Plan: %s", plan.Reasoning))

	// Execute attack chain
	o.executePlan(plan)
}

// pickBestTarget ranks networks by exploitability
func (o *OmegaBrain) pickBestTarget(targets []*NetworkTarget) *NetworkTarget {
	var best *NetworkTarget
	bestScore := -999

	for _, t := range targets {
		score := 0

		// Signal strength (closer = better)
		score += t.Signal

		// Security type (weaker = higher score)
		switch t.Security {
		case "OPEN":
			score += 1000 // Instant win
		case "WEP":
			score += 500
		case "WPA":
			score += 200
		case "WPA2":
			score += 100
		case "WPA3":
			score += 50 // Harder but possible with downgrade
		}

		// Client count (more clients = more deauth targets)
		score += len(t.Clients) * 20

		// Known brand with default creds
		if t.Brand != "Unknown" {
			score += 30
		}

		if score > bestScore {
			bestScore = score
			best = t
		}
	}

	return best
}

// buildAttackPlan creates the optimal attack sequence
func (o *OmegaBrain) buildAttackPlan(target *NetworkTarget) *AttackPlan {
	plan := &AttackPlan{
		Target:     target,
		AgentVotes: make(map[string]string),
	}

	// Agent voting
	plan.AgentVotes["ALPHA"] = "WPS"
	plan.AgentVotes["BETA"] = "PMKID"
	plan.AgentVotes["GAMMA"] = "EVIL_TWIN"
	plan.AgentVotes["DELTA"] = "PIVOT"

	switch target.Security {
	case "OPEN":
		plan.Reasoning = "OPEN network. Connect directly, no password needed."
		plan.Steps = []AttackStep{
			{Name: "Connect", Tool: "wpa_supplicant", Command: []string{"wpa_supplicant"}, Agent: "DELTA"},
		}
	case "WEP":
		plan.Reasoning = "WEP is trivially breakable with aircrack-ng."
		plan.Steps = []AttackStep{
			{Name: "Capture IVs", Tool: "airodump-ng", Command: []string{"airodump-ng"}, Agent: "BETA"},
			{Name: "Crack WEP", Tool: "aircrack-ng", Command: []string{"aircrack-ng"}, Agent: "BETA"},
		}
	case "WPA", "WPA2":
		plan.Reasoning = "WPA2 network. Fastest path: WPS → PMKID → Handshake → Evil Twin"
		plan.Steps = []AttackStep{
			{Name: "Test WPS", Tool: "reaver", Command: []string{"reaver"}, Agent: "BETA", Timeout: 300 * time.Second},
			{Name: "Capture PMKID", Tool: "hcxdumptool", Command: []string{"hcxdumptool"}, Agent: "BETA", Timeout: 60 * time.Second},
			{Name: "Deauth + Handshake", Tool: "aireplay-ng", Command: []string{"aireplay-ng"}, Agent: "BETA", Timeout: 120 * time.Second},
			{Name: "Crack", Tool: "aircrack-ng", Command: []string{"aircrack-ng"}, Agent: "BETA", Timeout: 300 * time.Second},
			{Name: "Evil Twin Fallback", Tool: "airbase-ng", Command: []string{"airbase-ng"}, Agent: "GAMMA", Timeout: 600 * time.Second},
		}
	case "WPA3":
		plan.Reasoning = "WPA3-SAE network. Try downgrade to WPA2 first."
		plan.Steps = []AttackStep{
			{Name: "WPA3 Downgrade", Tool: "dragonshift", Command: []string{"dragonshift"}, Agent: "BETA", Timeout: 120 * time.Second},
			{Name: "Capture PMKID", Tool: "hcxdumptool", Command: []string{"hcxdumptool"}, Agent: "BETA", Timeout: 60 * time.Second},
			{Name: "Evil Twin", Tool: "airbase-ng", Command: []string{"airbase-ng"}, Agent: "GAMMA", Timeout: 600 * time.Second},
		}
	}

	// Always add pivot as final step
	plan.Steps = append(plan.Steps, AttackStep{
		Name: "Auto-Pivot", Tool: "nmap", Command: []string{"nmap"}, Agent: "DELTA", Timeout: 300 * time.Second,
	})

	plan.EstimatedTime = time.Duration(len(plan.Steps)) * 60 * time.Second
	return plan
}

// executePlan runs the attack chain step by step
func (o *OmegaBrain) executePlan(plan *AttackPlan) {
	for i, step := range plan.Steps {
		o.broadcast(fmt.Sprintf("\033[36m[CHAIN %d/%d] %s | Agent: %s | Tool: %s\033[0m",
			i+1, len(plan.Steps), step.Name, step.Agent, step.Tool))

		// Enable ghost mode before attacks
		if step.Agent == "BETA" || step.Agent == "GAMMA" {
			o.state.mu.RLock()
			ghost := o.state.GhostMode
			o.state.mu.RUnlock()
			if ghost {
				o.broadcast("[GHOST] MAC morphing active for stealth.")
			}
		}

		// Dispatch to agent
		msg := AgentMessage{
			From:     "OMEGA",
			To:       step.Agent,
			Type:     "ATTACK_" + strings.ToUpper(step.Name),
			Target:   plan.Target,
			Data:     step.Tool,
			Priority: 10,
		}

		switch step.Agent {
		case "BETA":
			switch step.Name {
			case "Test WPS":
				msg.Type = "ATTACK_WPS"
			case "Capture PMKID":
				msg.Type = "ATTACK_PMKID"
			case "Deauth + Handshake":
				msg.Type = "ATTACK_HANDSHAKE"
			case "Crack":
				msg.Type = "CRACK_CAPTURE"
			case "WPA3 Downgrade":
				msg.Type = "ATTACK_WPA3_DOWNGRADE"
			}
		case "GAMMA":
			if step.Name == "Evil Twin Fallback" || step.Name == "Evil Twin" {
				msg.Type = "EVIL_TWIN_START"
			}
		case "DELTA":
			if step.Name == "Auto-Pivot" {
				msg.Type = "PIVOT_CONNECT"
				msg.Payload = map[string]interface{}{}
			}
		}

		o.bus.Broadcast(msg)

		// Wait for agent response (with timeout)
		select {
		case <-time.After(step.Timeout):
			// Timeout, continue to next step
		case <-o.stopCh:
			return
		}
	}

	o.broadcast("\033[32m[OMEGA] Attack chain complete.\033[0m")
}

// handleSuccess processes a successful attack step
func (o *OmegaBrain) handleSuccess(msg AgentMessage) {
	o.broadcast(fmt.Sprintf("\033[32m[OMEGA] SUCCESS from %s: %s\033[0m", msg.From, msg.Data))

	// If password found, immediately trigger pivot
	if msg.Target != nil {
		o.state.mu.Lock()
		for _, cred := range o.state.Results.Passwords {
			if cred.Target == msg.Target.SSID {
				o.broadcast(fmt.Sprintf("[OMEGA] Password found: %s. Triggering DELTA pivot...", cred.Target))
				o.bus.Broadcast(AgentMessage{
					From:    "OMEGA",
					To:      "DELTA",
					Type:    "PIVOT_CONNECT",
					Target:  msg.Target,
					Payload: map[string]interface{}{"password": cred.Password},
				})
				break
			}
		}
		o.state.mu.Unlock()
	}
}

// handleFailure retries or escalates to next attack
func (o *OmegaBrain) handleFailure(msg AgentMessage) {
	o.broadcast(fmt.Sprintf("\033[31m[OMEGA] FAILURE from %s: %s\033[0m", msg.From, msg.Data))
	o.broadcast("[OMEGA] Escalating to next attack vector...")
}

// orchestrateAutoAttack triggers the full autonomous chain
func (o *OmegaBrain) orchestrateAutoAttack() {
	o.broadcast("[OMEGA] Full autonomous attack initiated.")
	// Trigger scan first
	o.bus.Broadcast(AgentMessage{
		From: "OMEGA", To: "ALPHA", Type: "SCAN_START", Data: "auto",
	})
}

// handleUserCommand parses natural language WiFi commands
func (o *OmegaBrain) handleUserCommand(cmd string, payload map[string]interface{}) {
	cmd = strings.ToLower(cmd)

	if strings.Contains(cmd, "scan") && strings.Contains(cmd, "wifi") {
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "ALPHA", Type: "SCAN_START"})
	} else if strings.Contains(cmd, "attack") || strings.Contains(cmd, "hack") {
		if o.state.CurrentTarget != nil {
			plan := o.buildAttackPlan(o.state.CurrentTarget)
			o.executePlan(plan)
		} else {
			o.broadcast("[!] No target selected. Run scan first.")
		}
	} else if strings.Contains(cmd, "evil") || strings.Contains(cmd, "twin") {
		if o.state.CurrentTarget != nil {
			o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GAMMA", Type: "EVIL_TWIN_START", Target: o.state.CurrentTarget})
		}
	} else if strings.Contains(cmd, "ghost") || strings.Contains(cmd, "stealth") {
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GAMMA", Type: "GHOST_MODE_ON"})
	} else if strings.Contains(cmd, "pivot") {
		if o.state.CurrentTarget != nil {
			o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "DELTA", Type: "PIVOT_CONNECT", Target: o.state.CurrentTarget})
		}
	}
}

// Telemetry sends data to cloud relay
func (o *OmegaBrain) Telemetry(data map[string]interface{}) {
	if data == nil {
		return
	}
	// TODO: WebSocket/SSE connection to remote relay
	// For now, log locally
	jsonData, _ := json.Marshal(data)
	o.broadcast(fmt.Sprintf("[TELEMETRY] %s", string(jsonData)))
}

func (o *OmegaBrain) broadcast(msg string) {
	o.bus.Broadcast(AgentMessage{
		From: o.Name(), To: "ALL", Type: "LOG", Data: msg,
	})
}
