// DELTA-VULN — Automated Vulnerability Discovery Agent
// nuclei, dalfox, sqlmap, gitleaks, trufflehog, ffuf, paramspider

package osintbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// VulnAgent discovers and tests vulnerabilities
type VulnAgent struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewVulnAgent(bus *EventBus, state *SharedState) *VulnAgent {
	return &VulnAgent{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 100),
	}
}

func (a *VulnAgent) Name() string  { return "DELTA-VULN" }
func (a *VulnAgent) Status() string { return "online" }

func (a *VulnAgent) Start() {
	a.bus.Subscribe("DELTA", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *VulnAgent) Stop() { close(a.stopCh) }

func (a *VulnAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "VULN_SCAN":
		a.vulnScan(msg.Data)
	case "LIVE_HOSTS_FOUND":
		a.scanLiveHosts(msg.Data)
	case "SECRET_SCAN":
		a.secretScan(msg.Data)
	case "SUBDOMAIN_TAKEOVER":
		a.subdomainTakeoverCheck(msg.Data)
	}
}

func (a *VulnAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "DELTA", To: "ALL", Type: "LOG", Data: msg})
}

func (a *VulnAgent) vulnScan(target string) {
	a.broadcast(fmt.Sprintf("[DELTA] Vulnerability scan: %s", target))
	a.scanLiveHosts(target)
	a.secretScan(target)
	a.subdomainTakeoverCheck(target)
}

func (a *VulnAgent) scanLiveHosts(domain string) {
	a.state.mu.RLock()
	var liveHosts []string
	for sub, info := range a.state.Subdomains {
		if info.Live {
			liveHosts = append(liveHosts, "https://"+sub)
		}
	}
	a.state.mu.RUnlock()

	if len(liveHosts) == 0 {
		liveHosts = []string{"https://" + domain}
	}

	a.broadcast(fmt.Sprintf("[DELTA] Scanning %d live hosts", len(liveHosts)))

	// nuclei for vulnerability scanning
	for _, host := range liveHosts {
		out, err := exec.Command("nuclei", "-u", host,
			"-severity", "critical,high,medium",
			"-silent", "-jsonl").CombinedOutput()
		if err == nil && len(out) > 0 {
			a.broadcast("[DELTA] nuclei findings:")
			for _, line := range strings.Split(string(out), "\n") {
				if line == "" {
					continue
				}
				// Parse JSONL output
				if strings.Contains(line, `"info"`) {
					severity := "medium"
					if strings.Contains(line, `"severity":"critical"`) {
						severity = "critical"
					} else if strings.Contains(line, `"severity":"high"`) {
						severity = "high"
					}
					name := extractJSONField(line, "name")
					vuln := Vulnerability{
						Name:     name,
						Severity: severity,
						URL:      host,
						Tool:     "nuclei",
					}
					a.state.mu.Lock()
					a.state.Vulns = append(a.state.Vulns, vuln)
					a.state.mu.Unlock()
					a.broadcast(fmt.Sprintf("[VULN][%s] %s on %s", severity, name, host))
				}
			}
		}
	}

	// dalfox for XSS
	for _, host := range liveHosts {
		out, err := exec.Command("dalfox", "url", host, "-S", "-silent").CombinedOutput()
		if err == nil && len(out) > 0 {
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "[POC]") || strings.Contains(line, "VULN") {
					vuln := Vulnerability{
						Name:     "XSS",
						Severity: "high",
						URL:      host,
						Tool:     "dalfox",
						Poc:      line,
					}
					a.state.mu.Lock()
					a.state.Vulns = append(a.state.Vulns, vuln)
					a.state.mu.Unlock()
					a.broadcast(fmt.Sprintf("[CRITICAL] XSS found on %s", host))
				}
			}
		}
	}

	// paramspider for parameter discovery then test
	a.paramDiscovery(domain)
}

func (a *VulnAgent) paramDiscovery(domain string) {
	a.broadcast("[DELTA] Parameter discovery with paramspider...")
	if out, err := exec.Command("paramspider", "-d", domain).CombinedOutput(); err == nil {
		output := string(out)
		if len(output) > 0 {
			a.broadcast(fmt.Sprintf("[DELTA] paramspider found parameters:\n%s", output))
		}
	}
}

func (a *VulnAgent) secretScan(domain string) {
	a.broadcast("[DELTA] Secret scanning with gitleaks + trufflehog...")

	// gitleaks detect (if we have git repos)
	if out, err := exec.Command("gitleaks", "detect", "--source", ".", "-v").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "leak") {
			a.broadcast("[CRITICAL] gitleaks found exposed secrets!")
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "Rule") {
					secret := Secret{
						Type:   "api_key",
						Source: "git",
						URL:    domain,
					}
					a.state.mu.Lock()
					a.state.Secrets = append(a.state.Secrets, secret)
					a.state.mu.Unlock()
				}
			}
		}
	}

	// trufflehog filesystem scan
	if out, err := exec.Command("trufflehog", "filesystem", ".", "--json").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Verified") && strings.Contains(output, "true") {
			a.broadcast("[CRITICAL] trufflehog found verified secrets!")
		}
		_ = output
	}
}

func (a *VulnAgent) subdomainTakeoverCheck(domain string) {
	a.broadcast("[DELTA] Subdomain takeover check...")

	a.state.mu.RLock()
	var checkSubs []string
	for sub, info := range a.state.Subdomains {
		if !info.Live && info.StatusCode == 0 {
			checkSubs = append(checkSubs, sub)
		}
	}
	a.state.mu.RUnlock()

	if len(checkSubs) == 0 {
		return
	}

	for _, sub := range checkSubs {
		// Check for known takeover signatures
		if out, err := exec.Command("httpx", "-u", "http://"+sub, "-status-code", "-silent").CombinedOutput(); err == nil {
			output := string(out)
			if strings.Contains(output, "NXDOMAIN") ||
				strings.Contains(output, "NoSuchBucket") ||
				strings.Contains(output, "Fastly") ||
				strings.Contains(output, "Heroku") {
				a.broadcast(fmt.Sprintf("[CRITICAL] Subdomain takeover possible: %s", sub))
				a.state.mu.Lock()
				if s, ok := a.state.Subdomains[sub]; ok {
					s.Takeover = true
				}
				a.state.mu.Unlock()
			}
			_ = output
		}
	}
}

func extractJSONField(json, field string) string {
	// Simple extraction of "field":"value"
	prefix := fmt.Sprintf(`"%s":"`, field)
	idx := strings.Index(json, prefix)
	if idx >= 0 {
		start := idx + len(prefix)
		end := strings.Index(json[start:], `"`)
		if end >= 0 {
			return json[start : start+end]
		}
	}
	return ""
}
