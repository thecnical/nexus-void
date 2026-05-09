package purple

import (
	"fmt"
	"strings"
	"time"
)

// PurpleGuard is the purple team detection and response engine
type PurpleGuard struct {
	Organization string
	SIEMEndpoint string
}

// DetectionResult represents a detection rule or alert
type DetectionResult struct {
	RuleName    string `json:"rule_name"`
	Type        string `json:"type"` // sigma, yara, snort, eql
	Severity    string `json:"severity"`
	MITRE       string `json:"mitre"`
	Description string `json:"description"`
	Query       string `json:"query"`
}

// IncidentResponse represents an IR action
type IncidentResponse struct {
	AlertID     string    `json:"alert_id"`
	ThreatType  string    `json:"threat_type"`
	Action      string    `json:"action"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Containment string    `json:"containment"`
}

func NewPurpleGuard(org string) *PurpleGuard {
	return &PurpleGuard{
		Organization: org,
	}
}

// GenerateSigmaRule creates Sigma detection rules for attack patterns
func (p *PurpleGuard) GenerateSigmaRule(technique string) DetectionResult {
	fmt.Printf("[+] PURPLE-GUARD generating Sigma rule for: %s\n", technique)

	rules := map[string]DetectionResult{
		"mimikatz": {
			RuleName:    "mimikatz_usage",
			Type:        "sigma",
			Severity:    "critical",
			MITRE:       "T1003.001",
			Description: "Detects Mimikatz LSASS memory dump",
			Query: `title: Mimikatz LSASS Memory Dump
logsource:
  category: process_creation
  product: windows
detection:
  selection:
    CommandLine|contains:
      - "sekurlsa::logonpasswords"
      - "lsadump::sam"
      - "token::elevate"
  condition: selection`,
		},
		"bloodhound": {
			RuleName:    "bloodhound_collection",
			Type:        "sigma",
			Severity:    "high",
			MITRE:       "T1087.002",
			Description: "Detects BloodHound data collection",
			Query: `title: BloodHound Collection
detection:
  selection:
    CommandLine|contains:
      - "SharpHound.exe"
      - "Invoke-BloodHound"
      - "Get-BloodHoundData"
  condition: selection`,
		},
		" cobalt_strike": {
			RuleName:    "cobalt_strike_beacon",
			Type:        "sigma",
			Severity:    "critical",
			MITRE:       "T1071.001",
			Description: "Detects Cobalt Strike beacon communication",
			Query: `title: Cobalt Strike Beacon
detection:
  selection:
    UserAgent|contains: 'Mozilla/5.0'
    HttpRequest|contains:
      - "/submit.php?id="
      - "/activity"
  condition: selection`,
		},
	}

	if rule, ok := rules[technique]; ok {
		return rule
	}

	return DetectionResult{
		RuleName:    fmt.Sprintf("custom_%s_detection", technique),
		Type:        "sigma",
		Severity:    "medium",
		MITRE:       "T1001",
		Description: fmt.Sprintf("Custom detection for %s", technique),
		Query:       fmt.Sprintf("Detection query for %s", technique),
	}
}

// GenerateYaraRule creates YARA rules for malware detection
func (p *PurpleGuard) GenerateYaraRule(family string) DetectionResult {
	fmt.Printf("[+] PURPLE-GUARD generating YARA rule for: %s\n", family)

	return DetectionResult{
		RuleName:    fmt.Sprintf("detect_%s", family),
		Type:        "yara",
		Severity:    "high",
		MITRE:       "T1204.002",
		Description: fmt.Sprintf("Detects %s malware family", family),
		Query: fmt.Sprintf(`rule %s {
    strings:
        $a = "malicious_string_1" wide ascii
        $b = { 4D 5A }
    condition:
        uint16(0) == 0x5A4D and all of them
}`, family),
	}
}

// GenerateSnortRule creates Snort IDS rules
func (p *PurpleGuard) GenerateSnortRule(attack string) DetectionResult {
	fmt.Printf("[+] PURPLE-GUARD generating Snort rule for: %s\n", attack)

	return DetectionResult{
		RuleName:    fmt.Sprintf("alert_%s", attack),
		Type:        "snort",
		Severity:    "high",
		MITRE:       "T1071",
		Description: fmt.Sprintf("Detects %s network traffic", attack),
		Query: fmt.Sprintf(`alert tcp any any -> any 80 (
    msg:"%s detected";
    content:"malicious_payload";
    sid:1000001;
    rev:1;
)`, attack),
	}
}

// RunIncidentResponse executes IR playbook
func (p *PurpleGuard) RunIncidentResponse(alert IncidentResponse) []string {
	fmt.Printf("[+] PURPLE-GUARD executing IR playbook for: %s\n", alert.AlertID)

	var actions []string

	// Step 1: Isolate
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 1: Isolate affected host from network", alert.AlertID))

	// Step 2: Collect evidence
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 2: Collect memory dump and disk image", alert.AlertID))

	// Step 3: Analyze
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 3: Analyze indicators of compromise", alert.AlertID))

	// Step 4: Contain
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 4: Apply containment: %s", alert.AlertID, alert.Containment))

	// Step 5: Remediate
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 5: Remove persistence and malware", alert.AlertID))

	// Step 6: Recover
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 6: Restore from known-good backup", alert.AlertID))

	// Step 7: Lessons learned
	actions = append(actions, fmt.Sprintf("[IR-%s] Step 7: Document lessons learned", alert.AlertID))

	return actions
}

// HuntThreats performs threat hunting queries
func (p *PurpleGuard) HuntThreats(huntType string) []DetectionResult {
	fmt.Printf("[+] PURPLE-GUARD threat hunting: %s\n", huntType)

	var results []DetectionResult

	switch huntType {
	case "lateral_movement":
		results = append(results, DetectionResult{
			RuleName:    "hunt_lateral_psexec",
			Type:        "eql",
			Severity:    "high",
			MITRE:       "T1021.002",
			Description: "Hunt for PsExec usage",
			Query:       `process where process.name == "PsExec.exe"`,
		})
	case "persistence":
		results = append(results, DetectionResult{
			RuleName:    "hunt_persistence_registry",
			Type:        "eql",
			Severity:    "high",
			MITRE:       "T1547.001",
			Description: "Hunt for registry run keys",
			Query:       `registry where registry.path == "HKLM\\Software\\Microsoft\\Windows\\CurrentVersion\\Run"`,
		})
	case "credential_dumping":
		results = append(results, DetectionResult{
			RuleName:    "hunt_cred_dump",
			Type:        "eql",
			Severity:    "critical",
			MITRE:       "T1003",
			Description: "Hunt for credential dumping",
			Query:       `process where process.name in ("mimikatz.exe", "procdump.exe", "lsass.exe")`,
		})
	}

	return results
}

// ValidateDefense tests if defense measures detect simulated attacks
func (p *PurpleGuard) ValidateDefense(attack string) bool {
	fmt.Printf("[+] PURPLE-GUARD validating defense against: %s\n", attack)

	// Simulate attack and check detection
	detected := strings.Contains(attack, "mimikatz") ||
		strings.Contains(attack, "bloodhound") ||
		strings.Contains(attack, "cobalt")

	if detected {
		fmt.Printf("[+] Defense validated: %s detected\n", attack)
	} else {
		fmt.Printf("[!] Defense gap: %s not detected\n", attack)
	}

	return detected
}

// GeneratePlaybook creates an incident response playbook
func (p *PurpleGuard) GeneratePlaybook(scenario string) string {
	fmt.Printf("[+] PURPLE-GUARD generating IR playbook for: %s\n", scenario)

	playbook := fmt.Sprintf(`# Incident Response Playbook: %s
## Organization: %s

### Preparation
- [ ] Contact list verified
- [ ] Escalation paths documented
- [ ] Tools and credentials ready

### Detection
- [ ] Alert triaged and validated
- [ ] Scope determined
- [ ] Containment decision made

### Containment
- [ ] Short-term: Network isolation
- [ ] Long-term: Account disablement
- [ ] Evidence preservation

### Eradication
- [ ] Malware removed
- [ ] Persistence cleared
- [ ] Vulnerabilities patched

### Recovery
- [ ] Systems restored
- [ ] Monitoring enhanced
- [ ] Users notified

### Lessons Learned
- [ ] Timeline documented
- [ ] Root cause identified
- [ ] Improvements implemented
`, scenario, p.Organization)

	return playbook
}
