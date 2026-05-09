package intel

import (
	"fmt"
	"strings"
)

// MitigationGuide maps vulnerability types to fixes
type MitigationGuide struct {
	VulnType    string   `json:"vuln_type"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	FixSteps    []string `json:"fix_steps"`
	References  []string `json:"references"`
	Tools       []string `json:"tools"`
}

// MitigationDB is the vulnerability fix database
var MitigationDB = map[string]MitigationGuide{
	"sqli": {
		VulnType:    "SQL Injection",
		Severity:    "critical",
		Description: "User input is directly concatenated into SQL queries",
		FixSteps: []string{
			"Use parameterized queries (prepared statements)",
			"Use ORM frameworks that handle escaping automatically",
			"Implement input validation with whitelist approach",
			"Apply least privilege to database accounts",
			"Enable SQL injection WAF rules",
			"Use stored procedures where possible",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html",
			"https://portswigger.net/web-security/sql-injection",
		},
		Tools: []string{"sqlmap (for testing)", "Burp Suite", "OWASP ZAP"},
	},
	"xss": {
		VulnType:    "Cross-Site Scripting (XSS)",
		Severity:    "high",
		Description: "Untrusted data is rendered in browser without encoding",
		FixSteps: []string{
			"Encode all output based on context (HTML, JS, CSS, URL)",
			"Implement Content Security Policy (CSP)",
			"Use modern frameworks with auto-escaping",
			"Validate and sanitize all input",
			"Use HttpOnly and Secure flags on cookies",
			"Implement X-XSS-Protection header",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html",
			"https://portswigger.net/web-security/cross-site-scripting",
		},
		Tools: []string{"XSStrike", "xsser", "DOM XSS Scanner"},
	},
	"lfi": {
		VulnType:    "Local File Inclusion",
		Severity:    "high",
		Description: "User-controlled paths allow reading arbitrary files",
		FixSteps: []string{
			"Never use user input in file paths directly",
			"Use whitelisting of allowed files/directories",
			"Use chroot jails or sandboxed file access",
			"Disable dangerous PHP functions: include, require, fopen",
			"Run web server with minimal privileges",
		},
		References: []string{
			"https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/11.1-Testing_for_Local_File_Inclusion",
		},
		Tools: []string{"Liffy", "LFISuite"},
	},
	"ssrf": {
		VulnType:    "Server-Side Request Forgery",
		Severity:    "high",
		Description: "Server makes requests to attacker-controlled URLs",
		FixSteps: []string{
			"Whitelist allowed destination hosts/IPs",
			"Disable URL schemas like file://, gopher://, dict://",
			"Validate and sanitize all URLs",
			"Use URL parser that rejects internal IPs",
			"Implement network segmentation",
			"Block metadata endpoints (169.254.169.254)",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html",
			"https://portswigger.net/web-security/ssrf",
		},
		Tools: []string{"SSRFmap", "Burp Collaborator"},
	},
	"rce": {
		VulnType:    "Remote Code Execution",
		Severity:    "critical",
		Description: "Attacker can execute arbitrary commands on server",
		FixSteps: []string{
			"Never pass user input to system/exec/shell functions",
			"Use parameterized APIs instead of shell commands",
			"Implement strict input validation",
			"Use sandboxing (containers, seccomp, AppArmor)",
			"Disable dangerous functions in PHP/Python",
			"Apply WAF rules for command injection",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html",
		},
		Tools: []string{"commix", "command-injection-payload-list"},
	},
	"csrf": {
		VulnType:    "Cross-Site Request Forgery",
		Severity:    "medium",
		Description: "Attacker can perform actions on behalf of authenticated users",
		FixSteps: []string{
			"Implement anti-CSRF tokens on all state-changing requests",
			"Use SameSite cookie attribute",
			"Validate Referer/Origin headers",
			"Require re-authentication for sensitive actions",
			"Use custom request headers for AJAX",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html",
		},
		Tools: []string{"CSRFTester", "Burp Suite Pro"},
	},
	"jwt": {
		VulnType:    "JWT Weakness",
		Severity:    "high",
		Description: "JSON Web Token implementation vulnerabilities",
		FixSteps: []string{
			"Use strong signing algorithms (RS256, ES256) - avoid none/HS256",
			"Validate algorithm header strictly",
			"Keep signing keys secure (use HSM/KMS)",
			"Implement token expiration and rotation",
			"Use short-lived access tokens + refresh tokens",
			"Validate all claims (iss, aud, exp, nbf)",
		},
		References: []string{
			"https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html",
			"https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/",
		},
		Tools: []string{"jwt_tool", "jwt.io"},
	},
}

// GetMitigation returns fix guide for a vulnerability type
func GetMitigation(vulnType string) (MitigationGuide, bool) {
	// Normalize vuln type
	vt := strings.ToLower(strings.TrimSpace(vulnType))

	// Try exact match
	if guide, ok := MitigationDB[vt]; ok {
		return guide, true
	}

	// Try partial match
	for key, guide := range MitigationDB {
		if strings.Contains(vt, key) || strings.Contains(key, vt) {
			return guide, true
		}
	}

	return MitigationGuide{}, false
}

// GenerateMitigationReport builds full mitigation report
func GenerateMitigationReport(vulnTypes []string) string {
	var report []string
	report = append(report, "╔═══════════════════════════════════════════════════════════════╗")
	report = append(report, "║            MITIGATION GUIDE & REMEDIATION PLAN                ║")
	report = append(report, "╚═══════════════════════════════════════════════════════════════╝")
	report = append(report, "")

	for _, vt := range vulnTypes {
		if guide, ok := GetMitigation(vt); ok {
			report = append(report, fmt.Sprintf("\n[%s] %s", strings.ToUpper(guide.Severity), guide.VulnType))
			report = append(report, fmt.Sprintf("Description: %s", guide.Description))
			report = append(report, "")
			report = append(report, "Fix Steps:")
			for i, step := range guide.FixSteps {
				report = append(report, fmt.Sprintf("  %d. %s", i+1, step))
			}
			report = append(report, "")
			report = append(report, "References:")
			for _, ref := range guide.References {
				report = append(report, fmt.Sprintf("  • %s", ref))
			}
			report = append(report, "")
		}
	}

	return strings.Join(report, "\n")
}
