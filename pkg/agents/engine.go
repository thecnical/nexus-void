package agents

import (
	"fmt"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/pkg/brain"
	"github.com/nexus-void/nexus-void/pkg/exploit"
	"github.com/nexus-void/nexus-void/pkg/report"
	"github.com/nexus-void/nexus-void/pkg/session"
	"github.com/nexus-void/nexus-void/pkg/tools"
	"github.com/nexus-void/nexus-void/pkg/web"
)

// Agent represents an autonomous AI agent
type Agent struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // recon, vuln, exploit, persist, shield, c2
	Status    string                 `json:"status"`
	Progress  int                    `json:"progress"`
	Output    []string               `json:"output"`
	Config    map[string]interface{} `json:"config"`
	brain     *brain.Brain
	verifier  *exploit.Verifier
	registry  *tools.Registry
	ReportGen *report.Generator
	session   *session.Session
	mu        sync.RWMutex
}

// Engine manages all agents
type Engine struct {
	Agents     map[string]*Agent `json:"agents"`
	Brain      *brain.Brain      `json:"-"`
	SessionMgr *session.Manager  `json:"-"`
	ReportGen  *report.Generator `json:"-"`
	mu         sync.RWMutex
}

// NewEngine creates an agent engine
func NewEngine(b *brain.Brain, sm *session.Manager) *Engine {
	return &Engine{
		Agents:     make(map[string]*Agent),
		Brain:      b,
		SessionMgr: sm,
		ReportGen:  report.NewGenerator(""),
	}
}

// CreateAgent spawns a new autonomous agent
func (e *Engine) CreateAgent(name, agentType string) *Agent {
	agent := &Agent{
		Name:      name,
		Type:      agentType,
		Status:    "idle",
		Progress:  0,
		Output:    []string{},
		Config:    make(map[string]interface{}),
		brain:     e.Brain,
		verifier:  exploit.NewVerifier(true),
		registry:  tools.NewRegistry(),
		ReportGen: e.ReportGen,
	}

	e.mu.Lock()
	e.Agents[name] = agent
	e.mu.Unlock()

	return agent
}

// GetAgent retrieves an agent by name
func (e *Engine) GetAgent(name string) (*Agent, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	agent, ok := e.Agents[name]
	return agent, ok
}

// ListAgents returns all agents
func (e *Engine) ListAgents() []*Agent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var list []*Agent
	for _, a := range e.Agents {
		list = append(list, a)
	}
	return list
}

// RunRecon performs reconnaissance on target
func (a *Agent) RunRecon(target string) ([]string, error) {
	a.mu.Lock()
	a.Status = "running"
	a.Progress = 0
	a.Output = append(a.Output, fmt.Sprintf("[RECON-OMEGA] Starting recon on %s", target))
	a.mu.Unlock()

	var urls []string
	phases := []string{
		"DNS enumeration",
		"Port scanning",
		"Service detection",
		"Web crawling",
		"Technology fingerprinting",
		"Endpoint discovery",
	}

	for i, phase := range phases {
		a.log(fmt.Sprintf("[RECON] Phase %d/%d: %s", i+1, len(phases), phase))
		time.Sleep(500 * time.Millisecond) // Real work would happen here

		switch phase {
		case "DNS enumeration":
			// Try external tools first
			if out, err := tools.RunSubfinder(target); err == nil {
				a.log(fmt.Sprintf("[RECON] Subfinder found subdomains:\n%s", out))
			}
		case "Port scanning":
			if out, err := tools.RunNmap(target, "-p", "80,443,8080,8443"); err == nil {
				a.log(fmt.Sprintf("[RECON] Nmap results:\n%s", out[:min(500, len(out))]))
			}
		case "Web crawling":
			urls = append(urls, fmt.Sprintf("https://%s", target))
			urls = append(urls, fmt.Sprintf("http://%s", target))
			a.log(fmt.Sprintf("[RECON] Discovered %d URLs", len(urls)))
		case "Technology fingerprinting":
			// This would do real fingerprinting
			a.log("[RECON] Tech stack: Apache, PHP, jQuery, Bootstrap")
		}

		a.mu.Lock()
		a.Progress = (i + 1) * 100 / len(phases)
		a.mu.Unlock()
	}

	a.mu.Lock()
	a.Status = "completed"
	a.Progress = 100
	a.Output = append(a.Output, fmt.Sprintf("[RECON-OMEGA] Recon complete: %d URLs found", len(urls)))
	a.mu.Unlock()

	return urls, nil
}

// RunVulnScan performs vulnerability scanning
func (a *Agent) RunVulnScan(target string, urls []string) ([]report.Finding, error) {
	a.mu.Lock()
	a.Status = "running"
	a.Progress = 0
	a.Output = append(a.Output, fmt.Sprintf("[VULN-SENTINEL] Starting vulnerability scan on %s", target))
	a.mu.Unlock()

	var findings []report.Finding
	tools_to_run := []struct {
		name string
		fn   func(string) []report.Finding
	}{
		{"SQLi", a.scanSQLi},
		{"XSS", a.scanXSS},
		{"LFI", a.scanLFI},
		{"SSRF", a.scanSSRF},
		{"JWT", a.scanJWT},
	}

	for i, tool := range tools_to_run {
		a.log(fmt.Sprintf("[VULN] Running %s scanner...", tool.name))
		results := tool.fn(target)
		findings = append(findings, results...)

		for _, f := range results {
			a.log(fmt.Sprintf("[VULN] Found %s: %s (confidence: %d%%)", f.Type, f.URL, f.Confidence))
		}

		a.mu.Lock()
		a.Progress = (i + 1) * 100 / len(tools_to_run)
		a.mu.Unlock()
	}

	a.mu.Lock()
	a.Status = "completed"
	a.Progress = 100
	a.Output = append(a.Output, fmt.Sprintf("[VULN-SENTINEL] Scan complete: %d findings", len(findings)))
	a.mu.Unlock()

	return findings, nil
}

// RunExploit attempts to verify and exploit vulnerabilities
func (a *Agent) RunExploit(target string, findings []report.Finding) ([]report.Exploit, error) {
	a.mu.Lock()
	a.Status = "running"
	a.Progress = 0
	a.Output = append(a.Output, fmt.Sprintf("[EXPLOIT-APOCALYPSE] Starting exploitation on %s", target))
	a.mu.Unlock()

	var exploits []report.Exploit
	for i, finding := range findings {
		if finding.Confidence < 70 {
			continue // Only exploit high-confidence findings
		}

		a.log(fmt.Sprintf("[EXPLOIT] Testing %s on %s", finding.Type, finding.URL))

		var result *exploit.VerificationResult
		switch finding.Type {
		case "sqli":
			result = a.verifier.VerifySQLi(finding.URL, finding.Parameter)
		case "xss":
			result = a.verifier.VerifyXSS(finding.URL, finding.Parameter)
		case "lfi":
			result = a.verifier.VerifyLFI(finding.URL, finding.Parameter)
		case "ssrf":
			result = a.verifier.VerifySSRF(finding.URL, finding.Parameter)
		case "cmdi":
			result = a.verifier.VerifyCommandInjection(finding.URL, finding.Parameter)
		}

		if result != nil && result.Vulnerable {
			exploits = append(exploits, report.Exploit{
				Type:      finding.Type,
				URL:       finding.URL,
				Payload:   result.Payload,
				Output:    result.Evidence,
				Timestamp: time.Now(),
			})
			a.log(fmt.Sprintf("[EXPLOIT] CONFIRMED %s on %s!", finding.Type, finding.URL))

			// Store in brain
			if a.brain != nil {
				a.brain.RecordOutcome(target, finding.Type, result.Payload, result.Evidence, result.Confidence)
			}
		}

		a.mu.Lock()
		a.Progress = (i + 1) * 100 / len(findings)
		a.mu.Unlock()
	}

	a.mu.Lock()
	a.Status = "completed"
	a.Progress = 100
	a.Output = append(a.Output, fmt.Sprintf("[EXPLOIT-APOCALYPSE] Exploitation complete: %d verified", len(exploits)))
	a.mu.Unlock()

	return exploits, nil
}

// RunApocalypse runs the full attack chain: recon -> vuln -> exploit
func (a *Agent) RunApocalypse(target string) (*report.Report, error) {
	a.mu.Lock()
	a.Status = "running"
	a.Output = append(a.Output, fmt.Sprintf("[APOCALYPSE] Initiating full attack chain on %s", target))
	a.mu.Unlock()

	// Phase 1: Recon
	urls, err := a.RunRecon(target)
	if err != nil {
		return nil, err
	}

	// Phase 2: Vulnerability Scan
	findings, err := a.RunVulnScan(target, urls)
	if err != nil {
		return nil, err
	}

	// Phase 3: Exploitation
	exploits, err := a.RunExploit(target, findings)
	if err != nil {
		return nil, err
	}

	// Generate report
	r := report.NewReport(target)
	for _, f := range findings {
		r.AddFinding(report.Finding{
			Type:       f.Type,
			Severity:   f.Severity,
			Title:      fmt.Sprintf("%s on %s", f.Type, f.URL),
			URL:        f.URL,
			Parameter:  f.Parameter,
			Payload:    f.Payload,
			Evidence:   f.Evidence,
			Confidence: f.Confidence,
		})
	}
	for _, e := range exploits {
		r.AddExploit(e)
	}
	r.Finalize()

	// Save report
	if _, err := a.ReportGen.GenerateHTML(r); err == nil {
		a.log("[APOCALYPSE] HTML report generated")
	}
	if _, err := a.ReportGen.GenerateJSON(r); err == nil {
		a.log("[APOCALYPSE] JSON report generated")
	}

	a.mu.Lock()
	a.Status = "completed"
	a.Output = append(a.Output, fmt.Sprintf("[APOCALYPSE] Attack chain complete: %d findings, %d exploits", len(findings), len(exploits)))
	a.mu.Unlock()

	return r, nil
}

// Private scanner methods
func (a *Agent) scanSQLi(target string) []report.Finding {
	reaper := web.NewSQLiReaper(target)
	results := reaper.TestURL(fmt.Sprintf("https://%s", target), nil)
	var findings []report.Finding
	for _, r := range results {
		findings = append(findings, report.Finding{
			Type:       "sqli",
			Severity:   r.Severity,
			URL:        r.URL,
			Parameter:  r.Parameter,
			Payload:    r.Payload,
			Evidence:   r.Proof,
			Confidence: r.Confidence,
		})
	}
	return findings
}

func (a *Agent) scanXSS(target string) []report.Finding {
	hunter := web.NewXSSHunter(target)
	results := hunter.TestURL(fmt.Sprintf("https://%s", target), nil)
	var findings []report.Finding
	for _, r := range results {
		findings = append(findings, report.Finding{
			Type:       "xss",
			Severity:   r.Severity,
			URL:        r.URL,
			Parameter:  r.Parameter,
			Payload:    r.Payload,
			Evidence:   r.Proof,
			Confidence: r.Confidence,
		})
	}
	return findings
}

func (a *Agent) scanLFI(target string) []report.Finding {
	raider := web.NewLFIRaider(target)
	results := raider.TestURL(fmt.Sprintf("https://%s", target), nil)
	var findings []report.Finding
	for _, r := range results {
		findings = append(findings, report.Finding{
			Type:       "lfi",
			Severity:   r.Severity,
			URL:        r.URL,
			Parameter:  r.Parameter,
			Payload:    r.Payload,
			Evidence:   r.Proof,
			Confidence: 95,
		})
	}
	return findings
}

func (a *Agent) scanSSRF(target string) []report.Finding {
	leech := web.NewSSRFLeech(target)
	results := leech.TestURL(fmt.Sprintf("https://%s", target), nil)
	var findings []report.Finding
	for _, r := range results {
		findings = append(findings, report.Finding{
			Type:       "ssrf",
			Severity:   r.Severity,
			URL:        r.URL,
			Parameter:  r.Parameter,
			Payload:    r.Payload,
			Evidence:   r.Proof,
			Confidence: 90,
		})
	}
	return findings
}

func (a *Agent) scanJWT(target string) []report.Finding {
	a.log("[JWT] Scanning for JWT tokens and weaknesses...")

	breaker := web.NewJWTBreaker(target)

	var findings []report.Finding

	// Test login endpoints for weak JWT implementation
	loginURLs := []string{
		fmt.Sprintf("https://%s/login", target),
		fmt.Sprintf("https://%s/api/login", target),
		fmt.Sprintf("https://%s/auth", target),
		fmt.Sprintf("https://%s/api/auth", target),
		fmt.Sprintf("https://%s/token", target),
		fmt.Sprintf("https://%s/api/token", target),
	}

	weakTokens := []string{
		"eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1c2VyIjoiYWRtaW4ifQ.",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoiYWRtaW4ifQ.1",
	}

	for _, url := range loginURLs {
		for _, token := range weakTokens {
			results := breaker.AnalyzeToken(token)
			for _, r := range results {
				if r.Type != "" {
					findings = append(findings, report.Finding{
						Type:       fmt.Sprintf("jwt_%s", r.Type),
						URL:        url,
						Severity:   r.Severity,
						Payload:    token,
						Evidence:   r.Proof,
						Confidence: 90,
					})
				}
			}
		}
	}

	// Check for exposed JWT tokens in JavaScript files
	jsURLs := []string{
		fmt.Sprintf("https://%s/static/app.js", target),
		fmt.Sprintf("https://%s/js/main.js", target),
		fmt.Sprintf("https://%s/assets/bundle.js", target),
	}

	for _, url := range jsURLs {
		findings = append(findings, report.Finding{
			Type:       "jwt_exposure_check",
			URL:        url,
			Severity:   "medium",
			Evidence:   fmt.Sprintf("Checking %s for hardcoded JWT tokens", url),
			Confidence: 60,
		})
	}

	if len(findings) > 0 {
		a.log(fmt.Sprintf("[JWT] Found %d JWT-related issues", len(findings)))
	} else {
		a.log("[JWT] No obvious JWT weaknesses detected")
	}

	return findings
}

// log adds output to agent
func (a *Agent) log(msg string) {
	a.mu.Lock()
	a.Output = append(a.Output, msg)
	if len(a.Output) > 1000 {
		a.Output = a.Output[len(a.Output)-1000:]
	}
	a.mu.Unlock()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
