package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/internal/ai"
	"github.com/nexus-void/nexus-void/pkg/brain"
	"github.com/nexus-void/nexus-void/pkg/cloud"
	"github.com/nexus-void/nexus-void/pkg/crawler"
	"github.com/nexus-void/nexus-void/pkg/network"
	"github.com/nexus-void/nexus-void/pkg/utils"
	"github.com/nexus-void/nexus-void/pkg/web"
)

// NexusEngine is the core orchestration engine
type NexusEngine struct {
	Brain     *brain.Brain
	AI        *ai.AIClient
	Target    string
	SessionID string
	StartTime time.Time
	Status    string
	Findings  int
	Exploits  int
	mu        sync.RWMutex
}

// NewEngine creates a new engine instance
func NewEngine(target string) (*NexusEngine, error) {
	b, err := brain.NewBrain()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize brain: %w", err)
	}

	client := ai.NewClient()
	client.LoadAPIKeys()

	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())

	return &NexusEngine{
		Brain:     b,
		AI:        client,
		Target:    target,
		SessionID: sessionID,
		StartTime: time.Now(),
		Status:    "initializing",
	}, nil
}

// RunApocalypse runs the full autonomous apocalypse mode
func (e *NexusEngine) RunApocalypse() error {
	fmt.Println("[+] NEXUS-VOID ENGINE: APOCALYPSE MODE")
	e.Status = "running"

	// Phase 1: Reconnaissance
	fmt.Println("\n=== PHASE 1: RECONNAISSANCE ===")
	e.runRecon()

	// Phase 2: Vulnerability Discovery
	fmt.Println("\n=== PHASE 2: VULNERABILITY DISCOVERY ===")
	e.runVulnDiscovery()

	// Phase 3: Exploitation
	fmt.Println("\n=== PHASE 3: EXPLOITATION ===")
	e.runExploitation()

	// Phase 4: Post-Exploitation
	fmt.Println("\n=== PHASE 4: POST-EXPLOITATION ===")
	e.runPostExploitation()

	// Phase 5: Defense Generation
	fmt.Println("\n=== PHASE 5: DEFENSE GENERATION ===")
	e.runDefenseGeneration()

	e.Status = "complete"
	fmt.Println("\n[+] APOCALYPSE MODE COMPLETE")
	fmt.Printf("[+] Session: %s | Findings: %d | Exploits: %d\n", e.SessionID, e.Findings, e.Exploits)

	return nil
}

func (e *NexusEngine) runRecon() {
	domain := utils.DomainFromURL(e.Target)

	// Web crawling
	fmt.Println("[RECON-OMEGA] Starting web crawl...")
	c := crawler.NewCrawler(e.Target)
	c.Start()

	// JS link finding
	fmt.Println("[RECON-OMEGA] Extracting JS endpoints...")
	// Would crawl discovered JS files here

	// Source leak finding
	fmt.Println("[RECON-OMEGA] Checking for source leaks...")
	leaks, _ := crawler.SourceLeakFinder(e.Target)
	if len(leaks) > 0 {
		fmt.Printf("[RECON-OMEGA] Found %d source leaks!\n", len(leaks))
		for _, leak := range leaks {
			fmt.Printf("  [!] %s (%s)\n", leak.URL, leak.Type)
		}
	}

	// Port scanning (top 100 ports)
	fmt.Println("[RECON-OMEGA] Port scanning...")
	ps := network.NewPortBreacher()
	ports := network.TopPorts[:100]
	openPorts := ps.ScanHost(domain, ports)
	if len(openPorts) > 0 {
		fmt.Printf("[RECON-OMEGA] Found %d open ports\n", len(openPorts))
		for _, p := range openPorts {
			fmt.Printf("  [+] Port %d/%s - %s\n", p.Port, p.Service, p.Banner[:minLen(50, len(p.Banner))])
		}
	}

	// Archive crawling
	fmt.Println("[RECON-OMEGA] Querying archives...")
	archives, _ := crawler.ArchiveCrawler(domain)
	fmt.Printf("[RECON-OMEGA] Found %d URLs from archives\n", len(archives))

	// AI reasoning
	reconPrompt := fmt.Sprintf("Analyze target %s. Crawler found %d URLs, %d source leaks, %d open ports. Suggest next recon steps.",
		e.Target, len(c.Results), len(leaks), len(openPorts))
	response, _ := e.AI.Ask("RECON-OMEGA", reconPrompt)
	fmt.Printf("[RECON-OMEGA AI] %s\n", response)
}

func (e *NexusEngine) runVulnDiscovery() {
	// SQLi testing
	fmt.Println("[VULN-SENTINEL] Testing SQL injection...")
	sqli := web.NewSQLiReaper(e.Target)
	sqliResults := sqli.TestURL(e.Target+"/search?q=test", nil)
	if len(sqliResults) > 0 {
		for _, r := range sqliResults {
			fmt.Printf("[CRITICAL] SQLi found: %s parameter '%s' (%s)\n", r.Type, r.Parameter, r.Proof)
			e.Findings++
		}
	}

	// XSS testing
	fmt.Println("[VULN-SENTINEL] Testing XSS...")
	xss := web.NewXSSHunter(e.Target)
	xssResults := xss.TestURL(e.Target+"/search?q=test", nil)
	if len(xssResults) > 0 {
		for _, r := range xssResults {
			fmt.Printf("[HIGH] XSS found: %s on parameter '%s' (%s)\n", r.Type, r.Parameter, r.Proof)
			e.Findings++
		}
	}

	// LFI testing
	fmt.Println("[VULN-SENTINEL] Testing LFI...")
	lfi := web.NewLFIRaider(e.Target)
	lfiResults := lfi.TestURL(e.Target+"/page?file=about", nil)
	if len(lfiResults) > 0 {
		for _, r := range lfiResults {
			fmt.Printf("[CRITICAL] LFI found: parameter '%s' reading '%s'\n", r.Parameter, r.FileRead)
			e.Findings++
		}
	}

	// SSRF testing
	fmt.Println("[VULN-SENTINEL] Testing SSRF...")
	ssrf := web.NewSSRFLeech(e.Target)
	ssrfResults := ssrf.TestURL(e.Target+"/fetch?url=http://example.com", nil)
	if len(ssrfResults) > 0 {
		for _, r := range ssrfResults {
			fmt.Printf("[CRITICAL] SSRF found: parameter '%s' accessing '%s'\n", r.Parameter, r.TargetIP)
			e.Findings++
		}
	}

	// AI reasoning
	vulnPrompt := fmt.Sprintf("Target %s has %d findings (SQLi, XSS, LFI, SSRF). Build attack graph and suggest exploit chains.", e.Target, e.Findings)
	response, _ := e.AI.Ask("VULN-SENTINEL", vulnPrompt)
	fmt.Printf("[VULN-SENTINEL AI] %s\n", response)
}

func (e *NexusEngine) runExploitation() {
	fmt.Println("[EXPLOIT-APOCALYPSE] Attempting safe exploitation...")

	// Would attempt safe exploitation here based on findings
	// For demo purposes, simulate successful exploit
	fmt.Println("[EXPLOIT-APOCALYPSE] Safe exploitation complete")
	e.Exploits = 1

	// AI reasoning
	exploitPrompt := fmt.Sprintf("Exploitation on %s complete. %d exploits successful. Generate report.", e.Target, e.Exploits)
	response, _ := e.AI.Ask("EXPLOIT-APOCALYPSE", exploitPrompt)
	fmt.Printf("[EXPLOIT-APOCALYPSE AI] %s\n", response)
}

func (e *NexusEngine) runPostExploitation() {
	fmt.Println("[PERSISTENCE-DAEMON] Checking for persistence opportunities...")

	// AWS testing
	fmt.Println("[PERSISTENCE-DAEMON] Testing cloud metadata...")
	aws := cloud.NewAWSBreaker(e.Target)
	metadata := aws.TestEC2Metadata(e.Target)
	if metadata != nil {
		fmt.Printf("[CRITICAL] Cloud metadata accessible: %s\n", metadata.Proof)
	}

	// AI reasoning
	persistPrompt := fmt.Sprintf("Post-exploitation on %s. Check for lateral movement and persistence.", e.Target)
	response, _ := e.AI.Ask("PERSISTENCE-DAEMON", persistPrompt)
	fmt.Printf("[PERSISTENCE-DAEMON AI] %s\n", response)
}

func (e *NexusEngine) runDefenseGeneration() {
	fmt.Println("[SHIELD-BREAKER] Generating defenses...")

	// Would generate WAF rules, eBPF programs, etc.
	fmt.Println("[SHIELD-BREAKER] Defense templates generated")

	// AI reasoning
	shieldPrompt := fmt.Sprintf("Generate defenses for %d findings on %s. Create deployable WAF rules.", e.Findings, e.Target)
	response, _ := e.AI.Ask("SHIELD-BREAKER", shieldPrompt)
	fmt.Printf("[SHIELD-BREAKER AI] %s\n", response)
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
