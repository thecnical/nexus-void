// BETA-SURFACE — Attack Surface Mapping Agent
// httpx, waybackurls, gau, katana, nuclei fingerprinting

package osintbreach

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SurfaceAgent maps live hosts and tech stacks
type SurfaceAgent struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewSurfaceAgent(bus *EventBus, state *SharedState) *SurfaceAgent {
	return &SurfaceAgent{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 100),
	}
}

func (a *SurfaceAgent) Name() string  { return "BETA-SURFACE" }
func (a *SurfaceAgent) Status() string { return "online" }

func (a *SurfaceAgent) Start() {
	a.bus.Subscribe("BETA", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *SurfaceAgent) Stop() { close(a.stopCh) }

func (a *SurfaceAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SURFACE_MAP":
		a.surfaceMap(msg.Data)
	case "SUBDOMAINS_FOUND":
		a.probeSubdomains(msg.Data)
	case "TECH_FINGERPRINT":
		a.techFingerprint(msg.Data)
	case "CLOUD_HUNT":
		a.cloudHunt(msg.Data)
	}
}

func (a *SurfaceAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "BETA", To: "ALL", Type: "LOG", Data: msg})
}

func (a *SurfaceAgent) surfaceMap(domain string) {
	a.broadcast(fmt.Sprintf("[BETA] Mapping attack surface: %s", domain))
	a.probeSubdomains(domain)
	a.archiveDiscovery(domain)
	a.cloudHunt(domain)
}

func (a *SurfaceAgent) probeSubdomains(domain string) {
	// httpx probe all discovered subdomains
	a.state.mu.RLock()
	var subs []string
	for sub := range a.state.Subdomains {
		subs = append(subs, sub)
	}
	a.state.mu.RUnlock()

	if len(subs) == 0 {
		subs = []string{domain}
	}

	a.broadcast(fmt.Sprintf("[BETA] httpx probing %d hosts...", len(subs)))

	// Use httpx with tech detection
	for _, host := range subs {
		out, err := exec.Command("httpx", "-u", "https://"+host,
			"-title", "-tech-detect", "-status-code", "-web-server", "-silent").CombinedOutput()
		if err == nil && len(out) > 0 {
			line := string(out)
			a.broadcast(fmt.Sprintf("[BETA] %s", line))

			// Parse status code
			code := extractStatusCode(line)
			a.state.mu.Lock()
			if s, ok := a.state.Subdomains[host]; ok {
				s.StatusCode = code
				s.Live = code > 0
				s.Tech = extractTech(line)
			}
			a.state.mu.Unlock()
		}
	}

	// Trigger DELTA to scan live hosts
	a.bus.Broadcast(AgentMessage{From: "BETA", To: "DELTA", Type: "LIVE_HOSTS_FOUND", Data: domain})
}

func (a *SurfaceAgent) archiveDiscovery(domain string) {
	// waybackurls for URL discovery
	a.broadcast("[BETA] Archive URL discovery with waybackurls...")
	if out, err := exec.Command("waybackurls", domain).CombinedOutput(); err == nil {
		urls := strings.Split(string(out), "\n")
		a.broadcast(fmt.Sprintf("[BETA] waybackurls found %d historical URLs", len(urls)))

		// Search for API endpoints, interesting params
		for _, url := range urls {
			if strings.Contains(url, "/api/") || strings.Contains(url, "/graphql") ||
				strings.Contains(url, "/swagger") || strings.Contains(url, "/v1/") {
				a.state.mu.Lock()
				a.state.APIs = append(a.state.APIs, APIEndpoint{URL: url})
				a.state.mu.Unlock()
			}
		}
	}

	// gau - GetAllUrls
	if out, err := exec.Command("gau", domain, "--subs").CombinedOutput(); err == nil {
		urls := strings.Split(string(out), "\n")
		a.broadcast(fmt.Sprintf("[BETA] gau found %d URLs", len(urls)))
	}

	// katana crawl
	if out, err := exec.Command("katana", "-u", "https://"+domain, "-d", "3", "-silent").CombinedOutput(); err == nil {
		urls := strings.Split(string(out), "\n")
		a.broadcast(fmt.Sprintf("[BETA] katana crawled %d URLs", len(urls)))
	}
}

func (a *SurfaceAgent) techFingerprint(domain string) {
	// wappalyzer or nuclei -t technologies
	a.broadcast("[BETA] Technology fingerprinting...")
	if out, err := exec.Command("nuclei", "-u", "https://"+domain,
		"-t", "technologies/", "-silent").CombinedOutput(); err == nil {
		if len(out) > 0 {
			a.broadcast("[BETA] Tech detected:\n" + string(out))
		}
	}
}

func (a *SurfaceAgent) cloudHunt(domain string) {
	a.broadcast("[BETA] Hunting cloud assets...")

	// Check for S3 buckets
	bucketPatterns := []string{
		domain + "-backup", domain + "-assets", domain + "-data",
		"prod-" + domain, "dev-" + domain, "staging-" + domain,
	}
	for _, bucket := range bucketPatterns {
		bucketName := strings.ReplaceAll(bucket, ".", "-")
		cmd := exec.Command("aws", "s3", "ls", "s3://"+bucketName)
		if out, err := cmd.CombinedOutput(); err == nil {
			a.broadcast(fmt.Sprintf("[CRITICAL] Exposed S3 bucket: %s", bucketName))
			a.state.mu.Lock()
			a.state.CloudAssets = append(a.state.CloudAssets, CloudAsset{
				Type:   "s3",
				Name:   bucketName,
				URL:    "s3://" + bucketName,
				Public: true,
			})
			a.state.mu.Unlock()
			_ = out
		}
	}

	// Check for Azure blobs
	for _, bucket := range bucketPatterns {
		blobName := strings.ReplaceAll(bucket, ".", "")
		url := fmt.Sprintf("https://%s.blob.core.windows.net/", blobName)
		cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
		if out, err := cmd.CombinedOutput(); err == nil {
			if string(out) == "200" || string(out) == "404" {
				a.broadcast(fmt.Sprintf("[+] Azure blob found: %s", url))
			}
		}
	}
}

func extractStatusCode(line string) int {
	// Extract [200] or similar from httpx output
	if idx := strings.Index(line, "["); idx >= 0 {
		endIdx := strings.Index(line[idx:], "]")
		if endIdx > 0 {
			codeStr := line[idx+1 : idx+endIdx]
			if code, err := strconv.Atoi(codeStr); err == nil {
				return code
			}
		}
	}
	return 0
}

func extractTech(line string) string {
	if idx := strings.Index(line, "["); idx >= 0 {
		// Find tech stack after status code
		parts := strings.Split(line, " ")
		if len(parts) > 3 {
			return strings.Join(parts[3:], " ")
		}
	}
	return ""
}
