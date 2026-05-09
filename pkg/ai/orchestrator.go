package ai

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/pkg/agents"
	"github.com/nexus-void/nexus-void/pkg/brain"
	"github.com/nexus-void/nexus-void/pkg/report"
	"github.com/nexus-void/nexus-void/pkg/session"
	"github.com/nexus-void/nexus-void/pkg/tools"
)

// TargetProfile is the AI's understanding of the target
type TargetProfile struct {
	Target     string   `json:"target"`
	IsWeb      bool     `json:"is_web"`
	IsCloud    bool     `json:"is_cloud"`
	IsAD       bool     `json:"is_ad"`
	IsNetwork  bool     `json:"is_network"`
	IsWireless bool     `json:"is_wireless"`
	IsMobile   bool     `json:"is_mobile"`
	HasAPI     bool     `json:"has_api"`
	TechStack  []string `json:"tech_stack"`
	OpenPorts  []int    `json:"open_ports"`
	Services   []string `json:"services"`
	Categories []string `json:"categories"`
	Confidence float64  `json:"confidence"` // 0-1
}

// Strategy is the AI's attack plan
type Strategy struct {
	Target     string        `json:"target"`
	Phases     []Phase       `json:"phases"`
	Categories []string      `json:"categories"`
	Reasoning  string        `json:"reasoning"`
	EstTime    time.Duration `json:"est_time"`
	Parallel   bool          `json:"parallel"`
	Confidence float64       `json:"confidence"`
}

// Phase represents one attack phase
type Phase struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Tools      []string `json:"tools"`
	DependsOn  []string `json:"depends_on"` // Previous phases required
}

// Orchestrator is the AI brain for attack planning
type Orchestrator struct {
	registry    *tools.Registry
	brain       *brain.Brain
	agentEngine *agents.Engine
	swarm       *agents.SwarmCoordinator
	learned     map[string][]string // target fingerprint -> successful categories
	mu          sync.RWMutex
}

// NewOrchestrator creates the AI attack planner
func NewOrchestrator(brain *brain.Brain, engine *agents.Engine) *Orchestrator {
	return &Orchestrator{
		registry:    tools.NewRegistry(),
		brain:       brain,
		agentEngine: engine,
		learned:     make(map[string][]string),
	}
}

// QuickProbe does 5-second target reconnaissance
func (o *Orchestrator) QuickProbe(target string) *TargetProfile {
	profile := &TargetProfile{
		Target:     target,
		TechStack:  []string{},
		OpenPorts:  []int{},
		Services:   []string{},
		Categories: []string{},
	}

	// 1. HTTP Probe - is it a website?
	client := &http.Client{Timeout: 3 * time.Second}
	urlsToTry := []string{
		"https://" + target,
		"http://" + target,
		"https://www." + target,
	}

	for _, url := range urlsToTry {
		resp, err := client.Get(url)
		if err == nil {
			profile.IsWeb = true
			profile.OpenPorts = append(profile.OpenPorts, 80)
			if strings.HasPrefix(url, "https") {
				profile.OpenPorts = append(profile.OpenPorts, 443)
			}

			// Fingerprint from headers
			server := resp.Header.Get("Server")
			if server != "" {
				profile.Services = append(profile.Services, server)
				profile.TechStack = append(profile.TechStack, server)
			}
			poweredBy := resp.Header.Get("X-Powered-By")
			if poweredBy != "" {
				profile.TechStack = append(profile.TechStack, poweredBy)
			}

			// Check for API endpoints
			if resp.StatusCode == 404 {
				apiResp, _ := client.Get(url + "/api")
				if apiResp != nil && apiResp.StatusCode != 404 {
					profile.HasAPI = true
					if apiResp.Body != nil {
						apiResp.Body.Close()
					}
				}
			}
			resp.Body.Close()
			break
		}
	}

	// 2. DNS Resolution - is it a domain?
	ips, err := net.LookupIP(target)
	if err == nil && len(ips) > 0 {
		profile.IsNetwork = true
		profile.Services = append(profile.Services, fmt.Sprintf("Resolves to %s", ips[0].String()))
	}

	// 3. MX Records - is it a mail server?
	mxRecords, _ := net.LookupMX(target)
	if len(mxRecords) > 0 {
		profile.Services = append(profile.Services, "Mail server detected")
	}

	// 4. Quick port scan (top 5 ports)
	ports := []int{21, 22, 25, 53, 80, 110, 143, 443, 445, 993, 995, 3306, 3389, 5432, 8080, 8443}
	var wg sync.WaitGroup
	portChan := make(chan int, 5)
	var portMu sync.Mutex

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, fmt.Sprintf("%d", p)), 1*time.Second)
			if err == nil {
				conn.Close()
				portMu.Lock()
				profile.OpenPorts = append(profile.OpenPorts, p)
				portMu.Unlock()
				portChan <- p
			}
		}(port)
	}

	go func() {
		wg.Wait()
		close(portChan)
	}()

	for p := range portChan {
		switch p {
		case 445:
			profile.Services = append(profile.Services, "SMB (Windows/AD likely)")
			profile.IsAD = true
		case 53:
			profile.Services = append(profile.Services, "DNS")
		case 22:
			profile.Services = append(profile.Services, "SSH")
		case 25, 587:
			profile.Services = append(profile.Services, "SMTP")
		case 3306:
			profile.Services = append(profile.Services, "MySQL")
			profile.TechStack = append(profile.TechStack, "MySQL")
		case 5432:
			profile.Services = append(profile.Services, "PostgreSQL")
			profile.TechStack = append(profile.TechStack, "PostgreSQL")
		case 3389:
			profile.Services = append(profile.Services, "RDP (Windows)")
		case 8080, 8443:
			profile.Services = append(profile.Services, "Alt-Web")
			profile.IsWeb = true
		}
	}

	// 5. Classify target
	o.Classify(profile)

	// 6. Calculate confidence
	if profile.IsWeb {
		profile.Confidence += 0.3
	}
	if len(profile.OpenPorts) > 0 {
		profile.Confidence += 0.2
	}
	if len(profile.TechStack) > 0 {
		profile.Confidence += 0.2
	}
	if profile.IsAD {
		profile.Confidence += 0.2
	}
	if profile.Confidence > 1.0 {
		profile.Confidence = 1.0
	}

	return profile
}

// Classify determines categories based on probe results
func (o *Orchestrator) Classify(p *TargetProfile) {
	categories := []string{}

	if p.IsWeb {
		categories = append(categories, "web_crawling", "web_app")
		if p.HasAPI {
			categories = append(categories, "web_app") // API testing included
		}
	}

	if p.IsNetwork || len(p.OpenPorts) > 0 {
		categories = append(categories, "network")
	}

	if p.IsAD {
		categories = append(categories, "ad")
	}

	// Web + Database = SQLi high probability
	if p.IsWeb && containsAny(p.TechStack, []string{"MySQL", "PostgreSQL", "MariaDB", "MSSQL"}) {
		categories = append(categories, "web_app")
	}

	// Always add OSINT for reconnaissance
	categories = append(categories, "osint")

	// Cloud indicators
	if containsAny(p.Services, []string{"Amazon", "CloudFront", "AWS", "Azure", "Google"}) {
		categories = append(categories, "cloud")
	}

	// Crypto if HTTPS
	for _, port := range p.OpenPorts {
		if port == 443 || port == 8443 {
			categories = append(categories, "crypto")
			break
		}
	}

	// Remove duplicates
	p.Categories = uniqueStrings(categories)
}

// BuildStrategy creates the optimal attack plan
func (o *Orchestrator) BuildStrategy(target string, profile *TargetProfile, manualCats []string) *Strategy {
	categories := profile.Categories
	if len(manualCats) > 0 {
		categories = manualCats // User override
	}

	strategy := &Strategy{
		Target:     target,
		Categories: categories,
		Parallel:   true,
		Confidence: profile.Confidence,
	}

	// Check learned strategies
	o.mu.RLock()
	fingerprint := o.fingerprintTarget(profile)
	if learnedCats, ok := o.learned[fingerprint]; ok {
		strategy.Categories = learnedCats
		strategy.Reasoning = "Using learned strategy from previous similar target"
	}
	o.mu.RUnlock()

	if strategy.Reasoning == "" {
		strategy.Reasoning = o.generateReasoning(profile, categories)
	}

	// Build phases based on category dependencies
	strategy.Phases = o.buildPhases(categories)

	// Estimate time (rough)
	strategy.EstTime = time.Duration(len(categories)*30) * time.Second

	return strategy
}

// ExecuteStrategy runs the attack plan
func (o *Orchestrator) ExecuteStrategy(strategy *Strategy) (*report.Report, error) {
	target := strategy.Target

	// Create session
	sm := session.NewManager("")
	sess := sm.Create(target)

	// Create agent engine and swarm
	agentEngine := agents.NewEngine(o.brain, sm)
	coordinator := agents.NewSwarmCoordinator(agentEngine)

	// Run full swarm attack
	rpt, err := coordinator.LaunchAttack(target)
	if err != nil {
		return nil, err
	}

	// Learn from results
	o.Learn(strategy, rpt)

	sm.Complete(sess.ID)
	return rpt, nil
}

// ExecuteCategories runs specific categories in parallel
func (o *Orchestrator) ExecuteCategories(target string, categories []string) (*report.Report, error) {
	fmt.Printf("\n[AI] Executing %d categories in parallel against %s\n", len(categories), target)

	registry := tools.NewRegistry()
	var allFindings []report.Finding
	var allExploits []report.Exploit
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, cat := range categories {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()
			fmt.Printf("[AI] Running %s category...\n", category)

			// Get all tools for this category
			toolsList := registry.ListByCategory(tools.ToolCategory(category))
			if len(toolsList) == 0 {
				// Try matching partial category names
				toolsList = registry.List()
				var matched []*tools.Tool
				for _, t := range toolsList {
					if strings.Contains(strings.ToLower(string(t.Category)), strings.ToLower(category)) {
						matched = append(matched, t)
					}
				}
				toolsList = matched
			}

			// Run top tools from each category
			for _, tool := range toolsList {
				if tool.GoFunc == nil {
					continue
				}
				output, err := tool.GoFunc([]string{target})
				if err != nil {
					continue
				}
				mu.Lock()
				allFindings = append(allFindings, report.Finding{
					Type:     tool.Name,
					URL:      target,
					Evidence: output,
				})
				mu.Unlock()
			}
		}(cat)
	}

	wg.Wait()

	// Build report
	rpt := report.NewReport(target)
	for _, f := range allFindings {
		rpt.AddFinding(f)
	}
	for _, e := range allExploits {
		rpt.AddExploit(e)
	}
	rpt.Finalize()

	fmt.Printf("[AI] Parallel execution complete: %d findings\n", len(allFindings))
	return rpt, nil
}

// Learn remembers what worked
func (o *Orchestrator) Learn(strategy *Strategy, rpt *report.Report) {
	if len(rpt.Findings) == 0 {
		return
	}

	fingerprint := strategy.Target + strings.Join(strategy.Categories, ",")
	o.mu.Lock()
	o.learned[fingerprint] = strategy.Categories
	o.mu.Unlock()

	fmt.Printf("[AI] Learned: Target '%s' responded to categories: %v\n",
		strategy.Target, strategy.Categories)
}

// GetCategories returns all 15 categories for user selection
func (o *Orchestrator) GetCategories() []struct {
	ID          string
	Name        string
	Description string
} {
	return []struct {
		ID          string
		Name        string
		Description string
	}{
		{"web_crawling", "Web Crawling", "Deep web crawling, endpoint discovery, source leak detection"},
		{"web_app", "Web Application", "SQLi, XSS, LFI, RCE, SSRF, JWT, GraphQL, deserialization"},
		{"network", "Network", "Port scanning, service fingerprinting, SMB, LDAP, DNS, MITM"},
		{"cloud", "Cloud", "AWS, Azure, GCP, Kubernetes, Docker, serverless exploitation"},
		{"crypto", "Cryptography", "Hash cracking, TLS analysis, JWT attacks, PRNG testing"},
		{"osint", "OSINT", "Reconnaissance, email hunting, breach checking, GitHub harvesting"},
		{"post_exploit", "Post-Exploitation", "Privilege escalation, lateral movement, persistence, credential harvesting"},
		{"c2", "C2 & Implant", "Beacons, implants, steganography, DNS/ICMP/gRPC tunneling"},
		{"wireless", "Wireless & Mobile", "WiFi cracking, Bluetooth, Android/iOS, baseband"},
		{"hardware", "Hardware", "JTAG, SPI, I2C, firmware dumping, chip decapsulation"},
		{"telecom", "Telecom", "SS7, Diameter, GTP, SIP/VoIP, SMS interception"},
		{"ad", "Active Directory", "Kerberoasting, DCSync, pass-the-hash, golden ticket, RBCD"},
		{"purple", "Purple Team", "Detection testing, Sigma rules, honeypots, EDR evasion"},
		{"supply_chain", "Supply Chain", "Dependency scanning, typosquatting, CI/CD attacks"},
		{"social_eng", "Social Engineering", "Phishing, spear phishing, smishing, vishing, QR codes"},
		{"ml", "ML & AI", "Payload evolution, vulnerability prediction, WAF bypass, anomaly detection"},
	}
}

// Helper methods
func (o *Orchestrator) generateReasoning(p *TargetProfile, categories []string) string {
	reasons := []string{}
	if p.IsWeb {
		reasons = append(reasons, "web application detected")
	}
	if p.IsAD {
		reasons = append(reasons, "Active Directory indicators found")
	}
	if len(p.OpenPorts) > 0 {
		reasons = append(reasons, fmt.Sprintf("open ports: %v", p.OpenPorts))
	}
	if len(p.TechStack) > 0 {
		reasons = append(reasons, fmt.Sprintf("tech stack: %v", p.TechStack))
	}

	return fmt.Sprintf("Selected %d categories based on: %s. Estimated confidence: %.0f%%",
		len(categories), strings.Join(reasons, ", "), p.Confidence*100)
}

func (o *Orchestrator) buildPhases(categories []string) []Phase {
	phases := []Phase{}

	// Phase 1: Always recon first
	reconCats := []string{}
	for _, cat := range categories {
		if cat == "osint" || cat == "web_crawling" {
			reconCats = append(reconCats, cat)
		}
	}
	if len(reconCats) > 0 {
		phases = append(phases, Phase{
			Name:       "Reconnaissance",
			Categories: reconCats,
		})
	}

	// Phase 2: Network/Cloud/AD (can run parallel)
	infraCats := []string{}
	for _, cat := range categories {
		if cat == "network" || cat == "cloud" || cat == "ad" || cat == "crypto" {
			infraCats = append(infraCats, cat)
		}
	}
	if len(infraCats) > 0 {
		phases = append(phases, Phase{
			Name:       "Infrastructure",
			Categories: infraCats,
			DependsOn:  []string{"Reconnaissance"},
		})
	}

	// Phase 3: Web App attacks
	webCats := []string{}
	for _, cat := range categories {
		if cat == "web_app" {
			webCats = append(webCats, cat)
		}
	}
	if len(webCats) > 0 {
		phases = append(phases, Phase{
			Name:       "Web Application Attack",
			Categories: webCats,
			DependsOn:  []string{"Reconnaissance"},
		})
	}

	// Phase 4: Post-Exploit / C2 / Wireless / Hardware / Telecom
	advancedCats := []string{}
	for _, cat := range categories {
		if cat == "post_exploit" || cat == "c2" || cat == "wireless" ||
			cat == "hardware" || cat == "telecom" || cat == "supply_chain" ||
			cat == "social_eng" || cat == "purple" || cat == "ml" {
			advancedCats = append(advancedCats, cat)
		}
	}
	if len(advancedCats) > 0 {
		phases = append(phases, Phase{
			Name:       "Advanced Operations",
			Categories: advancedCats,
			DependsOn:  []string{"Infrastructure", "Web Application Attack"},
		})
	}

	return phases
}

func (o *Orchestrator) fingerprintTarget(p *TargetProfile) string {
	var parts []string
	if p.IsWeb {
		parts = append(parts, "web")
	}
	if p.IsAD {
		parts = append(parts, "ad")
	}
	parts = append(parts, p.TechStack...)
	return strings.Join(parts, "|")
}

func containsAny(haystack []string, needles []string) bool {
	for _, h := range haystack {
		for _, n := range needles {
			if strings.Contains(strings.ToLower(h), strings.ToLower(n)) {
				return true
			}
		}
	}
	return false
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
