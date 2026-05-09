package osint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// OSINTPhantom is the open-source intelligence gathering engine
type OSINTPhantom struct {
	Target string
}

// OSINTResult represents a piece of intelligence
type OSINTResult struct {
	Type       string `json:"type"` // email, subdomain, employee, credential, social
	Value      string `json:"value"`
	Source     string `json:"source"` // where it was found
	Confidence int    `json:"confidence"`
}

func NewOSINTPhantom(target string) *OSINTPhantom {
	return &OSINTPhantom{Target: target}
}

// HarvestEmails discovers email addresses from various sources
func (o *OSINTPhantom) HarvestEmails(domain string) []OSINTResult {
	fmt.Printf("[+] OSINT-PHANTOM harvesting emails for: %s\n", domain)

	var results []OSINTResult

	// Common patterns
	patterns := []string{
		"info@", "admin@", "support@", "contact@", "sales@",
		"security@", "abuse@", "postmaster@", "webmaster@", "hostmaster@",
		"noreply@", "no-reply@", "help@", "marketing@", "careers@",
		"jobs@", "hr@", "media@", "press@", "legal@", "billing@",
	}

	for _, prefix := range patterns {
		email := prefix + domain
		results = append(results, OSINTResult{
			Type:       "email",
			Value:      email,
			Source:     "pattern_guess",
			Confidence: 30,
		})
	}

	// Query HaveIBeenPwned-style APIs (simulated)
	// Query GitHub for leaked emails (simulated)

	fmt.Printf("[+] OSINT-PHANTOM found %d emails\n", len(results))
	return results
}

// DiscoverSubdomains uses multiple sources
func (o *OSINTPhantom) DiscoverSubdomains(domain string) []OSINTResult {
	fmt.Printf("[+] OSINT-PHANTOM discovering subdomains for: %s\n", domain)

	var results []OSINTResult

	// Common subdomains
	commonSubs := []string{
		"www", "mail", "ftp", "localhost", "webmail", "smtp", "pop", "ns1", "ns2",
		"ns3", "dns1", "dns2", "mx", "mx1", "mx2", "mail1", "mail2", "www2",
		"www1", "ns", "ns1", "ns2", "webmaster", "admin", "imap", "meeting",
		"calendar", "wiki", "docs", "git", "jenkins", "jira", "confluence",
		"staging", "dev", "test", "uat", "qa", "prod", "production", "api",
		"api1", "api2", "api3", "v1", "v2", "v3", "beta", "alpha", "demo",
		"portal", "vpn", "remote", "gateway", "secure", "app", "apps", "mobile",
		"m", "cdn", "static", "assets", "media", "img", "images", "css", "js",
		"blog", "news", "forum", "shop", "store", "cart", "checkout", "pay",
		"payment", "payments", "billing", "invoice", "account", "accounts",
		"auth", "login", "signin", "sso", "oauth", "id", "identity",
		"console", "dashboard", "panel", "cp", "control", "manage", "manager",
		"status", "monitor", "monitoring", "logs", "logging", "trace",
		"search", "es", "elastic", "kibana", "grafana", "prometheus",
		"gitlab", "github", "bitbucket", "svn", "cvs", "repo", "repository",
		"registry", "docker", "k8s", "kubernetes", "kube", "helm",
		"db", "database", "sql", "mysql", "postgres", "mongo", "redis",
		"cache", "memcache", "rabbitmq", "kafka", "mq", "queue",
		"backup", "backups", "archive", "archives", "old", "legacy",
		"temp", "tmp", "test1", "test2", "dev1", "dev2", "staging1", "staging2",
		"ftp", "sftp", "ssh", "telnet", "rlogin", "vnc", "rdp",
		"exchange", "owa", "autodiscover", "lync", "lyncdiscover",
		"sip", "sipe", "xmpp", "jabber", "im", "chat",
		"support", "help", "ticket", "tickets", "servicedesk", "helpdesk",
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50)

	for _, sub := range commonSubs {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(s string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			fqdn := s + "." + domain
			resp, err := utils.Fetch("http://"+fqdn, nil)
			if err != nil {
				return
			}
			resp.Body.Close()

			mu.Lock()
			results = append(results, OSINTResult{
				Type:       "subdomain",
				Value:      fqdn,
				Source:     "dns_resolution",
				Confidence: 100,
			})
			mu.Unlock()
		}(sub)
	}

	wg.Wait()

	fmt.Printf("[+] OSINT-PHANTOM found %d subdomains\n", len(results))
	return results
}

// SearchGitHubDorks searches GitHub for sensitive files
func (o *OSINTPhantom) SearchGitHubDorks(domain string) []OSINTResult {
	fmt.Printf("[+] OSINT-PHANTOM searching GitHub dorks for: %s\n", domain)

	var results []OSINTResult

	dorks := []string{
		"extension:json apikeys",
		"extension:json auth",
		"extension:yaml password",
		"extension:env DB_PASSWORD",
		"extension:sql dump",
		"extension:backup sql",
		"extension:pem private",
		"extension:ppk private",
		"extension:log password",
		"extension:xml password",
	}

	for _, dork := range dorks {
		results = append(results, OSINTResult{
			Type:       "github_dork",
			Value:      dork + " " + domain,
			Source:     "github",
			Confidence: 50,
		})
	}

	return results
}

// SearchShodan simulates Shodan queries
func (o *OSINTPhantom) SearchShodan(query string) []OSINTResult {
	fmt.Printf("[+] OSINT-PHANTOM querying Shodan: %s\n", query)

	var results []OSINTResult

	// Simulated Shodan results
	results = append(results, OSINTResult{
		Type:       "shodan_host",
		Value:      query,
		Source:     "shodan",
		Confidence: 80,
	})

	return results
}

// FindSocialMedia discovers social media accounts
func (o *OSINTPhantom) FindSocialMedia(company string) []OSINTResult {
	fmt.Printf("[+] OSINT-PHANTOM finding social media for: %s\n", company)

	var results []OSINTResult

	platforms := []struct {
		name string
		url  string
	}{
		{"linkedin", "https://linkedin.com/company/"},
		{"twitter", "https://twitter.com/"},
		{"facebook", "https://facebook.com/"},
		{"github", "https://github.com/"},
		{"youtube", "https://youtube.com/"},
		{"instagram", "https://instagram.com/"},
	}

	for _, platform := range platforms {
		results = append(results, OSINTResult{
			Type:       "social_media",
			Value:      platform.url + strings.ToLower(company),
			Source:     "platform_guess",
			Confidence: 40,
		})
	}

	return results
}

// ExtractMetadata extracts metadata from documents
func (o *OSINTPhantom) ExtractMetadata(url string) map[string]string {
	metadata := make(map[string]string)

	resp, err := http.Get(url)
	if err != nil {
		return metadata
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Extract emails
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(bodyStr, -1)
	if len(emails) > 0 {
		metadata["emails"] = strings.Join(utils.Unique(emails), ", ")
	}

	// Extract phone numbers
	phoneRegex := regexp.MustCompile(`\+?\d{1,3}[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	phones := phoneRegex.FindAllString(bodyStr, -1)
	if len(phones) > 0 {
		metadata["phones"] = strings.Join(utils.Unique(phones), ", ")
	}

	return metadata
}

func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
