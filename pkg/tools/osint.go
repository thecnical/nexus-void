package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// === 1. RECON-OMEGA ===
func runReconOmega(args []string) (string, error) {
	target := "example.com"
	if len(args) > 0 {
		target = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Target: %s", target))

	// DNS recon
	hosts, err := net.LookupHost(target)
	if err == nil {
		results = append(results, fmt.Sprintf("IP Addresses: %s", strings.Join(hosts, ", ")))
	}

	// MX records
	mxRecs, _ := net.LookupMX(target)
	if len(mxRecs) > 0 {
		var mxStrs []string
		for _, mx := range mxRecs {
			mxStrs = append(mxStrs, fmt.Sprintf("%s (pref=%d)", mx.Host, mx.Pref))
		}
		results = append(results, fmt.Sprintf("Mail servers: %s", strings.Join(mxStrs, ", ")))
	}

	// NS records
	nsRecs, _ := net.LookupNS(target)
	if len(nsRecs) > 0 {
		var nsStrs []string
		for _, ns := range nsRecs {
			nsStrs = append(nsStrs, ns.Host)
		}
		results = append(results, fmt.Sprintf("Name servers: %s", strings.Join(nsStrs, ", ")))
	}

	// TXT records
	txtRecs, _ := net.LookupTXT(target)
	if len(txtRecs) > 0 {
		results = append(results, fmt.Sprintf("TXT records: %s", strings.Join(txtRecs, " | ")))
	}

	// HTTP recon
	resp, err := http.Get("http://" + target)
	if err == nil {
		results = append(results, fmt.Sprintf("HTTP Status: %s", resp.Status))
		results = append(results, fmt.Sprintf("Server: %s", resp.Header.Get("Server")))
		results = append(results, fmt.Sprintf("X-Powered-By: %s", resp.Header.Get("X-Powered-By")))
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// Extract technologies
		content := string(body)
		if strings.Contains(content, "wp-content") {
			results = append(results, "Technology: WordPress detected")
		}
		if strings.Contains(content, "drupal") {
			results = append(results, "Technology: Drupal detected")
		}
		if strings.Contains(content, "joomla") {
			results = append(results, "Technology: Joomla detected")
		}
		if strings.Contains(content, "react") {
			results = append(results, "Technology: React detected")
		}
		if strings.Contains(content, "angular") {
			results = append(results, "Technology: Angular detected")
		}
		if strings.Contains(content, "vue") {
			results = append(results, "Technology: Vue.js detected")
		}
		resp.Body.Close()
	}

	// HTTPS recon
	resp, err = http.Get("https://" + target)
	if err == nil {
		results = append(results, fmt.Sprintf("HTTPS Status: %s", resp.Status))
		resp.Body.Close()
	}

	return fmt.Sprintf("[RECON-OMEGA]\n%s", strings.Join(results, "\n")), nil
}

// === 2. EMAIL-HUNTER ===
func runEmailHunter(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	// Scrape emails from website
	resp, err := http.Get("http://" + domain)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	emailRe := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRe.FindAllString(string(body), -1)
	seen := make(map[string]bool)
	var unique []string
	for _, e := range emails {
		if !seen[e] {
			seen[e] = true
			unique = append(unique, e)
		}
	}

	// Common patterns
	patterns := []string{
		"admin@" + domain,
		"info@" + domain,
		"support@" + domain,
		"contact@" + domain,
		"sales@" + domain,
		"security@" + domain,
		"noreply@" + domain,
		"webmaster@" + domain,
		"postmaster@" + domain,
		"hostmaster@" + domain,
	}

	return fmt.Sprintf("[EMAIL-HUNTER] Domain: %s\nEmails found: %d\n%s\nCommon patterns:\n%s",
		domain, len(unique), strings.Join(unique, "\n"), strings.Join(patterns, "\n")), nil
}

// === 3. DOMAIN-MAPPER ===
func runDomainMapper(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("Domain: %s", domain))

	// Get all DNS records
	hosts, _ := net.LookupHost(domain)
	results = append(results, fmt.Sprintf("A Records: %s", strings.Join(hosts, ", ")))

	mxRecs, _ := net.LookupMX(domain)
	var mxStrs []string
	for _, mx := range mxRecs {
		mxStrs = append(mxStrs, fmt.Sprintf("%s (pref=%d)", mx.Host, mx.Pref))
	}
	results = append(results, fmt.Sprintf("MX Records: %s", strings.Join(mxStrs, ", ")))

	nsRecs, _ := net.LookupNS(domain)
	var nsStrs []string
	for _, ns := range nsRecs {
		nsStrs = append(nsStrs, ns.Host)
	}
	results = append(results, fmt.Sprintf("NS Records: %s", strings.Join(nsStrs, ", ")))

	txtRecs, _ := net.LookupTXT(domain)
	results = append(results, fmt.Sprintf("TXT Records: %s", strings.Join(txtRecs, " | ")))

	// Reverse DNS
	for _, ip := range hosts {
		names, _ := net.LookupAddr(ip)
		if len(names) > 0 {
			results = append(results, fmt.Sprintf("Reverse DNS for %s: %s", ip, strings.Join(names, ", ")))
		}
	}

	// Subdomain enumeration (basic)
	commonSubs := []string{"www", "mail", "ftp", "admin", "blog", "shop", "api", "dev", "test"}
	for _, sub := range commonSubs {
		fqdn := sub + "." + domain
		_, err := net.LookupHost(fqdn)
		if err == nil {
			results = append(results, fmt.Sprintf("Subdomain: %s [ACTIVE]", fqdn))
		}
	}

	return fmt.Sprintf("[DOMAIN-MAPPER]\n%s", strings.Join(results, "\n")), nil
}

// === 4. SOCIAL-SCANNER ===
func runSocialScanner(args []string) (string, error) {
	username := "example"
	if len(args) > 0 {
		username = args[0]
	}

	platforms := []struct {
		name string
		url  string
	}{
		{"Twitter/X", "https://twitter.com/" + username},
		{"Instagram", "https://instagram.com/" + username},
		{"GitHub", "https://github.com/" + username},
		{"LinkedIn", "https://linkedin.com/in/" + username},
		{"Facebook", "https://facebook.com/" + username},
		{"Reddit", "https://reddit.com/user/" + username},
		{"YouTube", "https://youtube.com/@" + username},
		{"TikTok", "https://tiktok.com/@" + username},
		{"Twitch", "https://twitch.tv/" + username},
		{"Pinterest", "https://pinterest.com/" + username},
	}

	var results []string
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}

	for _, p := range platforms {
		resp, err := client.Get(p.url)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: ERROR", p.name))
			continue
		}
		status := fmt.Sprintf("%s: %s", p.name, resp.Status)
		if resp.StatusCode == 200 {
			status += " [FOUND]"
		} else if resp.StatusCode == 404 {
			status += " [NOT FOUND]"
		}
		resp.Body.Close()
		results = append(results, status)
	}

	return fmt.Sprintf("[SOCIAL-SCANNER] Username: %s\nPlatforms checked: %d\n%s",
		username, len(platforms), strings.Join(results, "\n")), nil
}

// === 5. BREACH-CHECKER ===
func runBreachChecker(args []string) (string, error) {
	email := "test@example.com"
	if len(args) > 0 {
		email = args[0]
	}

	// Check Have I Been Pwned API
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://haveibeenpwned.com/api/v3/breachedaccount/"+email, nil)
	req.Header.Set("User-Agent", "Nexus-Void-Breach-Checker")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var breaches []string
	if resp.StatusCode == 200 {
		var breachData []map[string]interface{}
		json.Unmarshal(body, &breachData)
		for _, b := range breachData {
			if name, ok := b["Name"].(string); ok {
				breaches = append(breaches, name)
			}
		}
	}

	return fmt.Sprintf("[BREACH-CHECKER] Email: %s\nBreaches found: %d\n%s",
		email, len(breaches), strings.Join(breaches, "\n")), nil
}

// === 6. GITHUB-HARVESTER ===
func runGitHubHarvester(args []string) (string, error) {
	org := "example"
	if len(args) > 0 {
		org = args[0]
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var results []string

	// Get repos
	resp, err := client.Get("https://api.github.com/users/" + org + "/repos")
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var repos []map[string]interface{}
	json.Unmarshal(body, &repos)
	for _, repo := range repos {
		if name, ok := repo["name"].(string); ok {
			results = append(results, fmt.Sprintf("Repo: %s", name))
			// Check for common secret files
			secretFiles := []string{".env", "config.json", "secrets.yml", "credentials.json"}
			for _, sf := range secretFiles {
				resp, _ := client.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", org, name, sf))
				if resp != nil && resp.StatusCode == 200 {
					results = append(results, fmt.Sprintf("  -> %s [EXPOSED]", sf))
					resp.Body.Close()
				}
			}
		}
	}

	return fmt.Sprintf("[GITHUB-HARVESTER] Org/User: %s\nRepos: %d\n%s",
		org, len(repos), strings.Join(results, "\n")), nil
}

// === 7. SHODAN-SCANNER ===
func runShodanScanner(args []string) (string, error) {
	target := "8.8.8.8"
	if len(args) > 0 {
		target = args[0]
	}

	// Shodan API requires key, so we use their web endpoint
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://www.shodan.io/host/" + target)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Extract open ports from HTML
	portRe := regexp.MustCompile(`Port (\d+)`)
	ports := portRe.FindAllStringSubmatch(string(body), -1)
	var foundPorts []string
	for _, p := range ports {
		if len(p) > 1 {
			foundPorts = append(foundPorts, p[1])
		}
	}

	return fmt.Sprintf("[SHODAN-SCANNER] Target: %s\nOpen ports from Shodan: %d\n%s",
		target, len(foundPorts), strings.Join(foundPorts, ", ")), nil
}

// === 8. PASTEBIN-HUNTER ===
func runPastebinHunter(args []string) (string, error) {
	query := "password"
	if len(args) > 0 {
		query = args[0]
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://pastebin.com/search?q=" + query)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Extract paste IDs
	pasteRe := regexp.MustCompile(`href="/([a-zA-Z0-9]{8})"`)
	pastes := pasteRe.FindAllStringSubmatch(string(body), -1)
	var pasteIDs []string
	seen := make(map[string]bool)
	for _, p := range pastes {
		if len(p) > 1 && !seen[p[1]] {
			seen[p[1]] = true
			pasteIDs = append(pasteIDs, p[1])
		}
	}

	return fmt.Sprintf("[PASTEBIN-HUNTER] Query: %s\nPastes found: %d\n%s",
		query, len(pasteIDs), strings.Join(pasteIDs, "\n")), nil
}

// === 9. WHOIS-ABUSER ===
func runWHOISAbuser(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}

	// Use WHOIS via whois API or external command
	output, err := RunExternalSafe("whois", domain)
	if err != nil {
		// Fallback: use whois API
		resp, err := http.Get("https://rdap.org/domain/" + domain)
		if err != nil {
			return "", err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var rdap map[string]interface{}
		json.Unmarshal(body, &rdap)
		output = string(body)
	}

	return fmt.Sprintf("[WHOIS-ABUSER] Domain: %s\n%s", domain, output), nil
}

// === 10. METADATA-EXTRACTOR ===
func runMetadataExtractor(args []string) (string, error) {
	url := "http://example.com/document.pdf"
	if len(args) > 0 {
		url = args[0]
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// Extract PDF metadata
	meta := []string{}
	if strings.Contains(string(body), "%PDF") {
		meta = append(meta, "Type: PDF Document")
		// Extract Author
		authorRe := regexp.MustCompile(`/Author\(([^)]+)\)`)
		if m := authorRe.FindStringSubmatch(string(body)); len(m) > 1 {
			meta = append(meta, "Author: "+m[1])
		}
		// Extract Title
		titleRe := regexp.MustCompile(`/Title\(([^)]+)\)`)
		if m := titleRe.FindStringSubmatch(string(body)); len(m) > 1 {
			meta = append(meta, "Title: "+m[1])
		}
		// Extract Creator
		creatorRe := regexp.MustCompile(`/Creator\(([^)]+)\)`)
		if m := creatorRe.FindStringSubmatch(string(body)); len(m) > 1 {
			meta = append(meta, "Creator: "+m[1])
		}
		// Extract Producer
		prodRe := regexp.MustCompile(`/Producer\(([^)]+)\)`)
		if m := prodRe.FindStringSubmatch(string(body)); len(m) > 1 {
			meta = append(meta, "Producer: "+m[1])
		}
	}
	if strings.HasSuffix(url, ".docx") || strings.HasSuffix(url, ".xlsx") {
		meta = append(meta, "Type: Office Document")
		meta = append(meta, "Note: Use OOXML parser for full metadata extraction")
	}

	return fmt.Sprintf("[METADATA-EXTRACTOR] URL: %s\nFile size: %d bytes\n%s",
		url, len(body), strings.Join(meta, "\n")), nil
}
