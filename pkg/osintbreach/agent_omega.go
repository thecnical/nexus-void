// OMEGA-BRAIN — AI Orchestrator for OSINTBREACH
// Correlates data, assigns targets, generates attack surface reports

package osintbreach

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// OsintOmega coordinates all OSINT agents
type OsintOmega struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewOsintOmega(bus *EventBus, state *SharedState) *OsintOmega {
	return &OsintOmega{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 200),
	}
}

func (o *OsintOmega) Name() string  { return "OMEGA-BRAIN" }
func (o *OsintOmega) Status() string { return "online" }

func (o *OsintOmega) Start() {
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

func (o *OsintOmega) Stop() { close(o.stopCh) }

func (o *OsintOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT":
		o.orchestrateFullRecon(msg.Data)
	case "AUTO_ATTACK":
		o.autonomousChain(msg.Data)
	case "REPORT":
		o.generateReport()
	default:
		// Telemetry logging
		fmt.Printf("\033[36m[OMEGA]\033[0m %s -> %s: %s | %s\n",
			msg.From, msg.To, msg.Type, msg.Data)
	}
}

func (o *OsintOmega) orchestrateFullRecon(domain string) {
	fmt.Println()
	fmt.Println("\033[36m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Println("\033[36m  OMEGA-BRAIN: AUTONOMOUS RECONNAISSANCE CHAIN INITIATED    \033[0m")
	fmt.Println("\033[36m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Printf("\033[33m  Target: %s\033[0m\n", domain)
	fmt.Println()

	// Phase 1: Reconnaissance (ALPHA)
	fmt.Println("\033[35m[PHASE 1/6]\033[0m ALPHA-RECON: Domain enumeration & mapping")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "ALPHA", Type: "DOMAIN_RECON", Data: domain})
	time.Sleep(3 * time.Second)

	// Phase 2: Surface Mapping (BETA)
	fmt.Println("\033[35m[PHASE 2/6]\033[0m BETA-SURFACE: Attack surface mapping & cloud hunting")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "BETA", Type: "SURFACE_MAP", Data: domain})
	time.Sleep(2 * time.Second)

	// Phase 3: People OSINT (GAMMA)
	fmt.Println("\033[35m[PHASE 3/6]\033[0m GAMMA-PERSONA: People discovery & breach checks")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GAMMA", Type: "PERSONA_HUNT", Data: domain})
	time.Sleep(2 * time.Second)

	// Phase 4: Vulnerability Scanning (DELTA)
	fmt.Println("\033[35m[PHASE 4/6]\033[0m DELTA-VULN: Vulnerability discovery")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "DELTA", Type: "VULN_SCAN", Data: domain})
	time.Sleep(2 * time.Second)

	// Phase 5: Supply Chain (EPSILON)
	fmt.Println("\033[35m[PHASE 5/6]\033[0m EPSILON-SUPPLY: Dependency & supply chain analysis")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "EPSILON", Type: "SUPPLY_CHECK", Data: "."})
	time.Sleep(2 * time.Second)

	// Phase 6: Report Generation
	fmt.Println("\033[35m[PHASE 6/6]\033[0m OMEGA-BRAIN: Attack surface correlation & report")
	o.generateReport()

	fmt.Println()
	fmt.Println("\033[32m[OMEGA] Autonomous reconnaissance chain complete.\033[0m")
	fmt.Println()
}

func (o *OsintOmega) autonomousChain(target string) {
	// AI-driven target analysis and chain selection
	fmt.Printf("\033[36m[OMEGA] Analyzing target: %s\033[0m\n", target)

	// Determine target type
	if strings.Contains(target, "/") || strings.HasPrefix(target, "http") {
		// URL target - full web recon
		o.orchestrateFullRecon(extractDomain(target))
	} else if strings.Contains(target, "@") {
		// Email target - persona hunt
		o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GAMMA", Type: "BREACH_CHECK", Data: target})
	} else {
		// Domain target - full recon
		o.orchestrateFullRecon(target)
	}
}

func (o *OsintOmega) generateReport() {
	o.state.mu.RLock()
	defer o.state.mu.RUnlock()

	fmt.Println()
	fmt.Println("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[33m           O S I N T B R E A C H   R E P O R T               \033[36m║\033[0m")
	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")

	if o.state.Target != nil {
		fmt.Printf("\033[36m║\033[0m  Target: \033[32m%-50s\033[36m║\033[0m\n", o.state.Target.Domain)
	}

	// Subdomains
	liveCount := 0
	takeoverCount := 0
	for _, sub := range o.state.Subdomains {
		if sub.Live {
			liveCount++
		}
		if sub.Takeover {
			takeoverCount++
		}
	}
	fmt.Printf("\033[36m║\033[0m  Subdomains Discovered:  \033[33m%-33d\033[36m║\033[0m\n", len(o.state.Subdomains))
	fmt.Printf("\033[36m║\033[0m  Live Hosts:             \033[32m%-33d\033[36m║\033[0m\n", liveCount)
	fmt.Printf("\033[36m║\033[0m  Subdomain Takeovers:    \033[31m%-33d\033[36m║\033[0m\n", takeoverCount)

	// People
	fmt.Printf("\033[36m║\033[0m  People Identified:      \033[33m%-33d\033[36m║\033[0m\n", len(o.state.People))

	// APIs
	fmt.Printf("\033[36m║\033[0m  API Endpoints Found:   \033[33m%-33d\033[36m║\033[0m\n", len(o.state.APIs))

	// Vulnerabilities
	critical, high, medium := 0, 0, 0
	for _, v := range o.state.Vulns {
		switch v.Severity {
		case "critical":
			critical++
		case "high":
			high++
		case "medium":
			medium++
		}
	}
	fmt.Printf("\033[36m║\033[0m  Vulnerabilities:        CRITICAL: \033[31m%-2d\033[0m  HIGH: \033[33m%-2d\033[0m  MED: \033[32m%-11d\033[36m║\033[0m\n",
		critical, high, medium)

	// Secrets
	fmt.Printf("\033[36m║\033[0m  Secrets Exposed:        \033[31m%-33d\033[36m║\033[0m\n", len(o.state.Secrets))

	// Cloud Assets
	fmt.Printf("\033[36m║\033[0m  Cloud Assets:          \033[33m%-33d\033[36m║\033[0m\n", len(o.state.CloudAssets))

	// Dependencies
	fmt.Printf("\033[36m║\033[0m  Vulnerable Dependencies: \033[31m%-33d\033[36m║\033[0m\n", len(o.state.Deps))

	// Attack Surface Score
	score := float64(liveCount)*2 + float64(len(o.state.APIs))*5 +
		float64(len(o.state.People))*3 + float64(critical)*20 +
		float64(high)*10 + float64(len(o.state.Secrets))*15 +
		float64(len(o.state.CloudAssets))*8 + float64(takeoverCount)*25

	var priority, color string
	if score >= 100 {
		priority = "CRITICAL"
		color = "\033[31m"
	} else if score >= 50 {
		priority = "HIGH"
		color = "\033[33m"
	} else {
		priority = "MEDIUM"
		color = "\033[32m"
	}

	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Printf("\033[36m║\033[0m  ATTACK SURFACE SCORE:   %s%-33.0f\033[0m\033[36m║\033[0m\n", color, score)
	fmt.Printf("\033[36m║\033[0m  PRIORITY:               %s%-33s\033[0m\033[36m║\033[0m\n", color, priority)
	fmt.Println("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	// Recommendations
	if critical > 0 || high > 0 {
		fmt.Println("\033[31m[!] IMMEDIATE ACTION REQUIRED\033[0m")
		fmt.Println("    - Patch critical/high vulnerabilities")
		fmt.Println("    - Rotate exposed secrets/API keys")
		fmt.Println("    - Fix subdomain takeovers")
		fmt.Println()
	}
	if len(o.state.Secrets) > 0 {
		fmt.Println("\033[33m[!] SECRET EXPOSURE DETECTED\033[0m")
		fmt.Println("    - Review git history")
		fmt.Println("    - Enable secret scanning in CI/CD")
		fmt.Println()
	}
	if takeoverCount > 0 {
		fmt.Println("\033[31m[!] SUBDOMAIN TAKEOVER POSSIBLE\033[0m")
		fmt.Println("    - Claim dangling DNS records")
		fmt.Println("    - Audit subdomain provisioning")
		fmt.Println()
	}
}

func (o *OsintOmega) checkTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	if idx := strings.Index(url, "/"); idx >= 0 {
		url = url[:idx]
	}
	return url
}
