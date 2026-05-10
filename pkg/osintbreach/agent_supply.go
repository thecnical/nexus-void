// EPSILON-SUPPLY — Supply Chain & Dependency Security Agent
// osv-scanner, trivy, retirejs, snyk, dependency-check, npm-audit

package osintbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// SupplyAgent analyzes dependencies and supply chain
type SupplyAgent struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewSupplyAgent(bus *EventBus, state *SharedState) *SupplyAgent {
	return &SupplyAgent{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 100),
	}
}

func (a *SupplyAgent) Name() string  { return "EPSILON-SUPPLY" }
func (a *SupplyAgent) Status() string { return "online" }

func (a *SupplyAgent) Start() {
	a.bus.Subscribe("EPSILON", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *SupplyAgent) Stop() { close(a.stopCh) }

func (a *SupplyAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SUPPLY_CHECK":
		a.supplyCheck(msg.Data)
	case "DEP_SCAN":
		a.depScan(msg.Data)
	case "TYPO_SQUAT":
		a.typoSquatCheck(msg.Data)
	}
}

func (a *SupplyAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "EPSILON", To: "ALL", Type: "LOG", Data: msg})
}

func (a *SupplyAgent) supplyCheck(repo string) {
	a.broadcast(fmt.Sprintf("[EPSILON] Supply chain analysis: %s", repo))
	a.depScan(repo)
	a.secretScanRepo(repo)
	a.typoSquatCheck(repo)
}

func (a *SupplyAgent) depScan(path string) {
	a.broadcast("[EPSILON] Scanning dependencies...")

	// osv-scanner for dependency vulns
	if out, err := exec.Command("osv-scanner", "-r", path).CombinedOutput(); err == nil {
		output := string(out)
		a.broadcast("[EPSILON] osv-scanner results:")

		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "CVE") || strings.Contains(line, "GHSA") {
				a.broadcast(fmt.Sprintf("[VULN] %s", line))
				// Extract package info
				dep := Dependency{
					Name: extractDepName(line),
					Vulns: extractCVEs(line),
				}
				a.state.mu.Lock()
				a.state.Deps = append(a.state.Deps, dep)
				a.state.mu.Unlock()
			}
		}
	} else {
		_ = out
	}

	// trivy filesystem scan
	if out, err := exec.Command("trivy", "fs", "--scanners", "vuln,misconfig,secret",
		"--format", "json", path).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, `"Vulnerabilities"`) {
			a.broadcast("[EPSILON] trivy found vulnerabilities in dependencies")
		}
		_ = output
	}

	// Check for package-lock.json or similar
	pkgManagers := []string{
		"package-lock.json",
		"yarn.lock",
		"Pipfile.lock",
		"go.mod",
		"Cargo.lock",
		"Gemfile.lock",
	}
	for _, pm := range pkgManagers {
		if out, err := exec.Command("find", path, "-name", pm, "-maxdepth", "3").CombinedOutput(); err == nil {
			if len(out) > 0 {
				a.broadcast(fmt.Sprintf("[EPSILON] Found %s dependency file", pm))
			}
		}
	}
}

func (a *SupplyAgent) secretScanRepo(path string) {
	a.broadcast("[EPSILON] Scanning repository for secrets...")

	if out, err := exec.Command("gitleaks", "detect", "--source", path, "-v").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "leak") {
			a.broadcast("[CRITICAL] Secrets found in repository!")
		}
		_ = output
	}
}

func (a *SupplyAgent) typoSquatCheck(packageName string) {
	a.broadcast(fmt.Sprintf("[EPSILON] Checking typo-squatting for: %s", packageName))

	// Generate common typosquats
	typos := generateTypos(packageName)
	for _, typo := range typos {
		// Check npm registry
		if out, err := exec.Command("npm", "view", typo, "version").CombinedOutput(); err == nil {
			if len(out) > 0 {
				a.broadcast(fmt.Sprintf("[ALERT] Typosquat package exists: %s", typo))
			}
		}
	}
}

func generateTypos(name string) []string {
	var typos []string
	if len(name) < 2 {
		return typos
	}
	// Common typo patterns
	typos = append(typos, name+"-utils")
	typos = append(typos, name+"-helpers")
	typos = append(typos, name+"-core")
	typos = append(typos, name+"js")
	// Swap adjacent characters
	for i := 0; i < len(name)-1; i++ {
		chars := []rune(name)
		chars[i], chars[i+1] = chars[i+1], chars[i]
		typos = append(typos, string(chars))
	}
	return typos
}

func extractDepName(line string) string {
	if idx := strings.Index(line, "│"); idx >= 0 {
		parts := strings.Split(line[idx:], " ")
		if len(parts) > 1 {
			return strings.Trim(parts[1], "│ ")
		}
	}
	return "unknown"
}

func extractCVEs(line string) []string {
	var cves []string
	for _, part := range strings.Split(line, " ") {
		if strings.HasPrefix(part, "CVE-") {
			cves = append(cves, part)
		} else if strings.HasPrefix(part, "GHSA-") {
			cves = append(cves, part)
		}
	}
	return cves
}
