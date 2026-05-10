// ALPHA-RECON — Domain & Subdomain Discovery Agent
// amass, subfinder, assetfinder, findomain, dnsx, naabu

package osintbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// ReconAgent handles DNS and subdomain reconnaissance
type ReconAgent struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewReconAgent(bus *EventBus, state *SharedState) *ReconAgent {
	return &ReconAgent{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 100),
	}
}

func (a *ReconAgent) Name() string  { return "ALPHA-RECON" }
func (a *ReconAgent) Status() string { return "online" }

func (a *ReconAgent) Start() {
	a.bus.Subscribe("ALPHA", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *ReconAgent) Stop() { close(a.stopCh) }

func (a *ReconAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "DOMAIN_RECON":
		a.fullRecon(msg.Data)
	case "SUBDOMAIN_ENUM":
		a.subdomainEnum(msg.Data)
	case "PORT_SCAN":
		a.portScan(msg.Data)
	}
}

func (a *ReconAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "ALPHA", To: "ALL", Type: "LOG", Data: msg})
}

// Full reconnaissance pipeline
func (a *ReconAgent) fullRecon(domain string) {
	a.broadcast(fmt.Sprintf("[ALPHA] Full reconnaissance on: %s", domain))

	// Phase 1: Passive subdomain enumeration
	a.subdomainEnum(domain)

	// Phase 2: DNS resolution and IP mapping
	a.dnsResolution(domain)

	// Phase 3: Port scan on discovered hosts
	a.portScan(domain)

	// Phase 4: Whois and ASN info
	a.whoisLookup(domain)
}

func (a *ReconAgent) subdomainEnum(domain string) {
	var allSubs []string

	// Tool 1: subfinder (passive sources)
	if out, err := exec.Command("subfinder", "-d", domain, "-all", "-silent").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				allSubs = append(allSubs, line)
			}
		}
		a.broadcast(fmt.Sprintf("[ALPHA] subfinder found %d subdomains", len(strings.Split(string(out), "\n"))-1))
	}

	// Tool 2: assetfinder
	if out, err := exec.Command("assetfinder", "--subs-only", domain).CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" && !contains(allSubs, line) {
				allSubs = append(allSubs, line)
			}
		}
		a.broadcast("[ALPHA] assetfinder complete")
	}

	// Tool 3: amass enum (passive)
	if out, err := exec.Command("amass", "enum", "-passive", "-d", domain, "-nocolor").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" && !contains(allSubs, line) {
				allSubs = append(allSubs, line)
			}
		}
		a.broadcast("[ALPHA] amass passive enum complete")
	}

	// Tool 4: findomain
	if out, err := exec.Command("findomain", "-t", domain, "-q").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" && !contains(allSubs, line) {
				allSubs = append(allSubs, line)
			}
		}
		a.broadcast("[ALPHA] findomain complete")
	}

	// Deduplicate and store
	a.state.mu.Lock()
	for _, sub := range allSubs {
		if _, ok := a.state.Subdomains[sub]; !ok {
			a.state.Subdomains[sub] = &Subdomain{Name: sub}
		}
	}
	count := len(a.state.Subdomains)
	a.state.mu.Unlock()

	a.broadcast(fmt.Sprintf("[ALPHA] Total unique subdomains: %d", count))

	// Trigger BETA to probe for live hosts
	a.bus.Broadcast(AgentMessage{From: "ALPHA", To: "BETA", Type: "SUBDOMAINS_FOUND", Data: domain})
}

func (a *ReconAgent) dnsResolution(domain string) {
	// Resolve all subdomains with dnsx
	a.state.mu.RLock()
	var subs []string
	for sub := range a.state.Subdomains {
		subs = append(subs, sub)
	}
	a.state.mu.RUnlock()

	if len(subs) == 0 {
		return
	}

	// Write subs to temp file for dnsx
	// dnsx -l subs.txt -a -resp
	a.broadcast("[ALPHA] Resolving DNS with dnsx...")

	if out, err := exec.Command("dnsx", "-d", domain, "-a", "-resp", "-silent").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line == "" {
				continue
			}
			// Parse "subdomain [IP]" format
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				sub := strings.TrimSpace(parts[0])
				ip := strings.Trim(parts[1], "[]")
				a.state.mu.Lock()
				if s, ok := a.state.Subdomains[sub]; ok {
					s.IP = ip
				}
				a.state.mu.Unlock()
			}
		}
	}
}

func (a *ReconAgent) portScan(domain string) {
	a.broadcast("[ALPHA] Port scanning with naabu...")

	// naabu -host domain -top-ports 100
	if out, err := exec.Command("naabu", "-host", domain, "-top-ports", "100", "-silent").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if line == "" {
				continue
			}
			// Format: host:port
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				host := parts[0]
				// port is parts[1]
				a.state.mu.Lock()
				if s, ok := a.state.Subdomains[host]; ok {
					s.Live = true
				}
				a.state.mu.Unlock()
			}
		}
		a.broadcast("[ALPHA] naabu port scan complete")
	}
}

func (a *ReconAgent) whoisLookup(domain string) {
	if out, err := exec.Command("whois", domain).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Registrar") {
			a.broadcast("[ALPHA] Whois data retrieved")
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
