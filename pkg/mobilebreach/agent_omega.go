// OVERMIND — MOBILE-OMEGA Agent
// AI orchestration, CVE chain selection, agent coordination, telemetry

package mobilebreach

import (
	"fmt"
	"strings"
	"time"

	"github.com/nexus-void/nexus-void/internal/ai"
)

// MobileOmega is the AI brain for mobile attacks
type MobileOmega struct {
	bus    *EventBus
	state  *SharedState
	cve    *CVEDatabase
	stopCh chan struct{}
	msgCh  chan AgentMessage
	ai     *ai.AIClient
}

func NewMobileOmega(bus *EventBus, state *SharedState, cve *CVEDatabase) *MobileOmega {
	aiClient := ai.NewClient()
	return &MobileOmega{
		bus:    bus,
		state:  state,
		cve:    cve,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 50),
		ai:     aiClient,
	}
}

func (o *MobileOmega) Name() string   { return "OVERMIND" }
func (o *MobileOmega) Status() string { return "online" }

func (o *MobileOmega) Start() {
	o.bus.Subscribe("OVERMIND", o.msgCh)
	for {
		select {
		case msg := <-o.msgCh:
			o.Handle(msg)
		case <-o.stopCh:
			return
		}
	}
}

func (o *MobileOmega) Stop() {
	close(o.stopCh)
}

func (o *MobileOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT":
		o.broadcast("[OVERMIND] MobileBreach swarm initialized. Standing by for targets.")
	case "AUTO_ATTACK":
		o.runAutoChain(msg.Data)
	case "APK_ANALYZED":
		o.onAPKAnalyzed()
	case "DEVICE_CONNECTED":
		o.onDeviceConnected(msg.Data)
	case "API_ENDPOINT_FOUND":
		o.onAPIFound(msg.Data)
	case "CELL_TARGET_FOUND":
		o.onCellFound(msg.Data)
	}
}

func (o *MobileOmega) broadcast(msg string) {
	o.bus.Broadcast(AgentMessage{From: "OVERMIND", To: "ALL", Type: "LOG", Data: msg})
}

// ─── Feature 15: AI Attack Chain ─────────────────────────────────
func (o *MobileOmega) runAutoChain(target string) {
	o.broadcast(fmt.Sprintf("[OVERMIND] Building AI attack chain for: %s", target))

	// Step 1: Determine target type
	platform := o.detectPlatform(target)
	o.broadcast(fmt.Sprintf("[OVERMIND] Target platform detected: %s", platform))

	// Step 2: Look up CVEs for platform
	cves := o.cve.Lookup(platform, "")
	o.broadcast(fmt.Sprintf("[OVERMIND] %d chainable CVEs available for %s", len(cves), platform))

	// Step 3: Build attack plan
	plan := o.buildAttackPlan(platform, cves)
	if plan == nil {
		o.broadcast("[!] No attack plan could be generated for this target.")
		return
	}

	o.state.mu.Lock()
	o.state.AttackPlan = plan
	o.state.mu.Unlock()

	o.broadcast(fmt.Sprintf("[OVERMIND] Attack plan: %s | %d steps | Confidence: %.0f%%",
		plan.Name, len(plan.Steps), plan.Confidence*100))

	// Step 4: Execute plan step by step
	for i, step := range plan.Steps {
		o.broadcast(fmt.Sprintf("[OVERMIND] Step %d/%d: %s -> %s (%s)",
			i+1, len(plan.Steps), step.Agent, step.Action, step.Tool))

		o.bus.Broadcast(AgentMessage{
			From: "OVERMIND",
			To:   step.Agent,
			Type: step.Action,
			Data: target,
		})

		time.Sleep(step.Timeout)
	}

	o.broadcast("[OVERMIND] Attack chain execution complete. Collecting results...")
}

func (o *MobileOmega) buildAttackPlan(platform string, cves []CVEEntry) *AttackChain {
	switch platform {
	case "android":
		return o.buildAndroidChain(cves)
	case "ios":
		return o.buildIOSChain(cves)
	case "api":
		return o.buildAPIChain(cves)
	case "cellular":
		return o.buildCellularChain()
	default:
		return o.buildGenericChain()
	}
}

func (o *MobileOmega) buildAndroidChain(cves []CVEEntry) *AttackChain {
	chain := &AttackChain{
		Name:       "ANDROID-FULL-CHAIN",
		Confidence: 0.85,
	}

	// Phase 1: APEX - Reverse APK
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "APEX",
		Action:  "APK_REVERSE",
		Tool:    "jadx+apktool",
		Timeout: 30 * time.Second,
	})

	// Phase 2: APEX - MITM patch if SSL pinning detected
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "APEX",
		Action:  "APK_MITM_PATCH",
		Tool:    "apk-mitm",
		Timeout: 60 * time.Second,
		Depends: []int{0},
	})

	// Phase 3: APEX - Deep link hijack
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "APEX",
		Action:  "DEEP_LINK_HIJACK",
		Tool:    "adb",
		Timeout: 20 * time.Second,
		Depends: []int{0},
	})

	// Phase 4: APEX - Frida hook
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "APEX",
		Action:  "FRIDA_HOOK",
		Tool:    "frida",
		Timeout: 30 * time.Second,
		Depends: []int{1},
	})

	// Phase 5: LANCE - API recon from hardcoded URLs
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "LANCE",
		Action:  "API_RECON",
		Tool:    "httpx+arjun",
		Timeout: 45 * time.Second,
		Depends: []int{0},
	})

	// Phase 6: LANCE - SDK poison from extracted keys
	chain.Steps = append(chain.Steps, AttackStep{
		Agent:   "LANCE",
		Action:  "SDK_POISON",
		Tool:    "curl",
		Timeout: 30 * time.Second,
		Depends: []int{4},
	})

	// Add CVE-specific steps
	for _, cve := range cves {
		if cve.ExploitType == "zero-click" {
			chain.Steps = append(chain.Steps, AttackStep{
				Agent:   "APEX",
				Action:  "ZERO_CLICK_EXPLOIT",
				Tool:    cve.Tool,
				Timeout: 60 * time.Second,
			})
			chain.Confidence += 0.05
		}
	}

	return chain
}

func (o *MobileOmega) buildIOSChain(cves []CVEEntry) *AttackChain {
	chain := &AttackChain{
		Name:       "IOS-FULL-CHAIN",
		Confidence: 0.80,
	}

	chain.Steps = []AttackStep{
		{Agent: "GHOST", Action: "IPA_DUMP", Tool: "frida-ios-dump", Timeout: 60 * time.Second},
		{Agent: "GHOST", Action: "JAILBREAK_BYPASS", Tool: "frida", Timeout: 30 * time.Second, Depends: []int{0}},
		{Agent: "GHOST", Action: "KEYCHAIN_DUMP", Tool: "idevicebackup2", Timeout: 45 * time.Second, Depends: []int{1}},
		{Agent: "GHOST", Action: "EXTENSION_ABUSE", Tool: "frida", Timeout: 30 * time.Second, Depends: []int{0}},
		{Agent: "LANCE", Action: "API_RECON", Tool: "httpx", Timeout: 30 * time.Second, Depends: []int{0}},
	}

	return chain
}

func (o *MobileOmega) buildAPIChain(cves []CVEEntry) *AttackChain {
	chain := &AttackChain{
		Name:       "API-FULL-CHAIN",
		Confidence: 0.90,
	}

	chain.Steps = []AttackStep{
		{Agent: "LANCE", Action: "API_RECON", Tool: "httpx+arjun+nuclei", Timeout: 60 * time.Second},
		{Agent: "LANCE", Action: "GRAPHQL_MAP", Tool: "graphqlmap", Timeout: 45 * time.Second, Depends: []int{0}},
		{Agent: "LANCE", Action: "JWT_FORGE", Tool: "jwt_tool", Timeout: 60 * time.Second, Depends: []int{0}},
		{Agent: "LANCE", Action: "MITM_START", Tool: "mitmproxy", Timeout: 30 * time.Second, Depends: []int{0}},
		{Agent: "LANCE", Action: "SDK_POISON", Tool: "curl", Timeout: 30 * time.Second, Depends: []int{0}},
	}

	return chain
}

func (o *MobileOmega) buildCellularChain() *AttackChain {
	chain := &AttackChain{
		Name:       "CELLULAR-FULL-CHAIN",
		Confidence: 0.75,
	}

	chain.Steps = []AttackStep{
		{Agent: "SPECTRE", Action: "CELL_SCAN", Tool: "gr-gsm+hackrf", Timeout: 60 * time.Second},
		{Agent: "SPECTRE", Action: "IMSI_CATCHER", Tool: "PacketRusher+srsRAN", Timeout: 120 * time.Second, Depends: []int{0}},
		{Agent: "SPECTRE", Action: "PAGING_ATTACK", Tool: "gr-gsm", Timeout: 60 * time.Second, Depends: []int{0}},
		{Agent: "SPECTRE", Action: "SMS_INTERCEPT", Tool: "srsUE", Timeout: 90 * time.Second, Depends: []int{1}},
		{Agent: "SPECTRE", Action: "ESIM_EXTRACT", Tool: "pySim", Timeout: 45 * time.Second},
	}

	return chain
}

func (o *MobileOmega) buildGenericChain() *AttackChain {
	return &AttackChain{
		Name:       "GENERIC-RECON",
		Confidence: 0.50,
		Steps: []AttackStep{
			{Agent: "APEX", Action: "APK_REVERSE", Tool: "jadx", Timeout: 30 * time.Second},
			{Agent: "LANCE", Action: "API_RECON", Tool: "httpx", Timeout: 30 * time.Second},
		},
	}
}

func (o *MobileOmega) detectPlatform(target string) string {
	target = strings.ToLower(target)
	if strings.HasSuffix(target, ".apk") || strings.Contains(target, "android") {
		return "android"
	}
	if strings.HasSuffix(target, ".ipa") || strings.HasPrefix(target, "com.") || strings.Contains(target, "ios") {
		return "ios"
	}
	if strings.HasPrefix(target, "http") || strings.Contains(target, ".com") || strings.Contains(target, ".io") {
		return "api"
	}
	if strings.Contains(target, "5g") || strings.Contains(target, "lte") || strings.Contains(target, "gsm") {
		return "cellular"
	}
	return "unknown"
}

// ─── Event Handlers ──────────────────────────────────────────────

func (o *MobileOmega) onAPKAnalyzed() {
	o.state.mu.RLock()
	info := o.state.AppInfo
	o.state.mu.RUnlock()

	if info == nil {
		return
	}

	// Auto-decide next steps based on APK analysis
	if info.SSLPinned {
		o.broadcast("[OVERMIND] SSL pinning detected. Auto-triggering MITM patch...")
		o.bus.Broadcast(AgentMessage{From: "OVERMIND", To: "APEX", Type: "APK_MITM_PATCH", Data: ""})
	}

	if info.RootDetection {
		o.broadcast("[OVERMIND] Root detection found. Auto-triggering Frida bypass...")
		o.bus.Broadcast(AgentMessage{From: "OVERMIND", To: "APEX", Type: "FRIDA_HOOK", Data: ""})
	}

	if len(info.HardcodedSecrets) > 0 {
		o.broadcast(fmt.Sprintf("[OVERMIND] %d secrets found. Auto-triggering API recon...", len(info.HardcodedSecrets)))
		o.bus.Broadcast(AgentMessage{From: "OVERMIND", To: "LANCE", Type: "SDK_POISON", Data: ""})
	}
}

func (o *MobileOmega) onDeviceConnected(data string) {
	o.broadcast(fmt.Sprintf("[OVERMIND] Device connected: %s", data))
}

func (o *MobileOmega) onAPIFound(data string) {
	o.broadcast(fmt.Sprintf("[OVERMIND] API endpoint discovered: %s", data))
	if strings.Contains(data, "graphql") {
		o.broadcast("[OVERMIND] GraphQL detected! Triggering deep map...")
		o.bus.Broadcast(AgentMessage{From: "OVERMIND", To: "LANCE", Type: "GRAPHQL_MAP", Data: data})
	}
}

func (o *MobileOmega) onCellFound(data string) {
	o.broadcast(fmt.Sprintf("[OVERMIND] Cellular target: %s", data))
}

// ─── AI Decision (Groq/OpenRouter) ─────────────────────────────────
func (o *MobileOmega) aiDecision(query string) string {
	if o.ai == nil {
		return o.fallbackDecision(query)
	}

	prompt := fmt.Sprintf(`You are OVERMIND, a black-hat AI mobile pentest orchestrator.
Target: %s
Available CVEs: %d
Available agents: APEX(Android), GHOST(iOS), LANCE(API), SPECTRE(Cellular)

Determine the optimal attack chain. Respond ONLY with the agent name and action.
Example: "APEX:APK_REVERSE -> LANCE:API_RECON"`, query, len(o.cve.entries))

	resp, err := o.ai.Ask("OVERMIND", prompt)
	if err != nil {
		return o.fallbackDecision(query)
	}
	return resp
}

func (o *MobileOmega) fallbackDecision(query string) string {
	query = strings.ToLower(query)
	if strings.Contains(query, "apk") || strings.Contains(query, "android") {
		return "APEX:APK_REVERSE -> APEX:APK_MITM_PATCH -> APEX:FRIDA_HOOK -> LANCE:API_RECON -> LANCE:SDK_POISON"
	}
	if strings.Contains(query, "ios") || strings.Contains(query, "ipa") {
		return "GHOST:IPA_DUMP -> GHOST:JAILBREAK_BYPASS -> GHOST:KEYCHAIN_DUMP -> LANCE:API_RECON"
	}
	if strings.Contains(query, "api") || strings.Contains(query, "graphql") {
		return "LANCE:API_RECON -> LANCE:GRAPHQL_MAP -> LANCE:JWT_FORGE -> LANCE:MITM_START"
	}
	if strings.Contains(query, "5g") || strings.Contains(query, "cellular") {
		return "SPECTRE:CELL_SCAN -> SPECTRE:IMSI_CATCHER -> SPECTRE:PAGING_ATTACK -> SPECTRE:SMS_INTERCEPT"
	}
	return "APEX:APK_REVERSE -> LANCE:API_RECON"
}
