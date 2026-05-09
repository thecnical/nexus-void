package crawler

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// CrawlResult represents a discovered URL/asset
type CrawlResult struct {
	URL         string            `json:"url"`
	Source      string            `json:"source"` // Where it was found
	StatusCode  int               `json:"status_code"`
	Title       string            `json:"title"`
	ContentType string            `json:"content_type"`
	IsForm      bool              `json:"is_form"`
	Parameters  map[string]string `json:"parameters"`
	Depth       int               `json:"depth"`
	Timestamp   time.Time         `json:"timestamp"`
}

// Crawler is the base web crawler engine
type Crawler struct {
	BaseURL           string
	MaxDepth          int
	MaxConcurrency    int
	FollowRedirects   bool
	IncludeSubdomains bool
	Results           []CrawlResult
	Visited           map[string]bool
	mu                sync.RWMutex
	client            *http.Client
}

func NewCrawler(baseURL string) *Crawler {
	return &Crawler{
		BaseURL:           baseURL,
		MaxDepth:          3,
		MaxConcurrency:    50,
		FollowRedirects:   true,
		IncludeSubdomains: true,
		Results:           []CrawlResult{},
		Visited:           make(map[string]bool),
		client:            utils.GetHTTPClient(),
	}
}

func (c *Crawler) Start() error {
	fmt.Printf("[+] Starting crawler on: %s (max depth: %d)\n", c.BaseURL, c.MaxDepth)

	// Parse base URL
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	// Seed with base URL
	c.crawl(base.String(), base, 0)

	fmt.Printf("[+] Crawler complete. Found %d unique URLs\n", len(c.Results))
	return nil
}

func (c *Crawler) crawl(pageURL string, base *url.URL, depth int) {
	if depth > c.MaxDepth {
		return
	}

	c.mu.Lock()
	if c.Visited[pageURL] {
		c.mu.Unlock()
		return
	}
	c.Visited[pageURL] = true
	c.mu.Unlock()

	resp, err := utils.Fetch(pageURL, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyStr := string(body)

	// Extract title
	title := extractTitle(bodyStr)

	result := CrawlResult{
		URL:         pageURL,
		Source:      "crawler",
		StatusCode:  resp.StatusCode,
		Title:       title,
		ContentType: resp.Header.Get("Content-Type"),
		IsForm:      strings.Contains(bodyStr, "<form"),
		Depth:       depth,
		Timestamp:   time.Now(),
	}

	c.mu.Lock()
	c.Results = append(c.Results, result)
	c.mu.Unlock()

	// Extract links
	links := extractLinks(bodyStr, pageURL, base)

	// Crawl discovered links concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.MaxConcurrency)

	for _, link := range links {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(l string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			c.crawl(l, base, depth+1)
		}(link)
	}

	wg.Wait()
}

func extractTitle(html string) string {
	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func extractLinks(html, pageURL string, base *url.URL) []string {
	var links []string

	// Extract href attributes
	hrefRe := regexp.MustCompile(`href=["']([^"']+)["']`)
	matches := hrefRe.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			absolute := resolveURL(link, pageURL, base)
			if absolute != "" {
				links = append(links, absolute)
			}
		}
	}

	// Extract src attributes
	srcRe := regexp.MustCompile(`src=["']([^"']+)["']`)
	srcMatches := srcRe.FindAllStringSubmatch(html, -1)

	for _, match := range srcMatches {
		if len(match) > 1 {
			link := match[1]
			absolute := resolveURL(link, pageURL, base)
			if absolute != "" {
				links = append(links, absolute)
			}
		}
	}

	// Extract action attributes (forms)
	actionRe := regexp.MustCompile(`action=["']([^"']+)["']`)
	actionMatches := actionRe.FindAllStringSubmatch(html, -1)

	for _, match := range actionMatches {
		if len(match) > 1 {
			link := match[1]
			absolute := resolveURL(link, pageURL, base)
			if absolute != "" {
				links = append(links, absolute)
			}
		}
	}

	return utils.Unique(links)
}

func resolveURL(link, pageURL string, base *url.URL) string {
	// Skip anchors, javascript, mailto, tel
	if strings.HasPrefix(link, "#") ||
		strings.HasPrefix(link, "javascript:") ||
		strings.HasPrefix(link, "mailto:") ||
		strings.HasPrefix(link, "tel:") {
		return ""
	}

	u, err := url.Parse(link)
	if err != nil {
		return ""
	}

	if u.IsAbs() {
		// Absolute URL - check if same domain
		if u.Hostname() != base.Hostname() {
			// Different domain - skip unless subdomains allowed
			if !isSubdomain(u.Hostname(), base.Hostname()) {
				return ""
			}
		}
		return u.String()
	}

	// Relative URL
	baseURL, _ := url.Parse(pageURL)
	if baseURL == nil {
		return ""
	}

	resolved := baseURL.ResolveReference(u)
	return resolved.String()
}

func isSubdomain(host, domain string) bool {
	return strings.HasSuffix(host, "."+domain) || host == domain
}

// JSLinkFinder extracts API endpoints from JavaScript files
func JSLinkFinder(jsURL string) ([]string, error) {
	fmt.Printf("[+] Analyzing JS file: %s\n", jsURL)

	resp, err := utils.Fetch(jsURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsContent := string(body)

	var endpoints []string

	// Extract API paths from fetch/XHR calls
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`fetch\(["']([^"']+)["']`),
		regexp.MustCompile(`axios\.(?:get|post|put|delete)\(["']([^"']+)["']`),
		regexp.MustCompile(`\$\.(?:get|post|ajax)\s*\(\s*["']([^"']+)["']`),
		regexp.MustCompile(`url\s*:\s*["']([^"']+)["']`),
		regexp.MustCompile(`path\s*:\s*["']([^"']+)["']`),
		regexp.MustCompile(`route\s*:\s*["']([^"']+)["']`),
		regexp.MustCompile(`endpoint\s*:\s*["']([^"']+)["']`),
		regexp.MustCompile(`api\/[^\s"']+`),
		regexp.MustCompile(`v\d+\/[^\s"']+`),
		regexp.MustCompile(`\/api\/[^\s"']+`),
		regexp.MustCompile(`\/graphql`),
		regexp.MustCompile(`\/rest\/[^\s"']+`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(jsContent, -1)
		for _, match := range matches {
			if len(match) > 1 {
				endpoints = append(endpoints, match[1])
			} else if len(match) > 0 {
				endpoints = append(endpoints, match[0])
			}
		}
	}

	// Extract from string literals that look like paths
	pathRe := regexp.MustCompile(`["'](/[a-zA-Z0-9_\-/]+(?:\.[a-z]+)?)["']`)
	pathMatches := pathRe.FindAllStringSubmatch(jsContent, -1)
	for _, match := range pathMatches {
		if len(match) > 1 {
			path := match[1]
			// Filter common non-API paths
			if !isStaticAsset(path) {
				endpoints = append(endpoints, path)
			}
		}
	}

	return utils.Unique(endpoints), nil
}

func isStaticAsset(path string) bool {
	staticExts := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	lower := strings.ToLower(path)
	for _, ext := range staticExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// FormHunter discovers all forms on a page
func FormHunter(pageURL string) ([]FormInfo, error) {
	resp, err := utils.Fetch(pageURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(body)

	var forms []FormInfo

	// Extract form tags
	formRe := regexp.MustCompile(`<form[^>]*>(.*?)</form>`)
	formMatches := formRe.FindAllStringSubmatch(html, -1)

	for _, match := range formMatches {
		if len(match) < 2 {
			continue
		}

		formHTML := match[0]
		innerHTML := match[1]

		form := FormInfo{
			Action:     extractAttr(formHTML, "action"),
			Method:     strings.ToUpper(extractAttr(formHTML, "method")),
			Enctype:    extractAttr(formHTML, "enctype"),
			Parameters: []string{},
		}

		if form.Method == "" {
			form.Method = "GET"
		}

		// Extract input fields
		inputRe := regexp.MustCompile(`<input[^>]*>`)
		inputMatches := inputRe.FindAllString(innerHTML, -1)

		for _, input := range inputMatches {
			name := extractAttr(input, "name")
			if name != "" {
				form.Parameters = append(form.Parameters, name)
			}
		}

		forms = append(forms, form)
	}

	return forms, nil
}

type FormInfo struct {
	Action     string   `json:"action"`
	Method     string   `json:"method"`
	Enctype    string   `json:"enctype"`
	Parameters []string `json:"parameters"`
}

func extractAttr(tag, attr string) string {
	re := regexp.MustCompile(attr + `=["']([^"']*)["']`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// SourceLeakFinder checks for exposed source files
func SourceLeakFinder(baseURL string) ([]LeakFinding, error) {
	leakPaths := []string{
		"/.git/HEAD",
		"/.git/config",
		"/.svn/entries",
		"/.env",
		"/.env.local",
		"/.env.production",
		"/.env.development",
		"/config.php",
		"/config.inc.php",
		"/wp-config.php",
		"/configuration.php",
		"/.htaccess",
		"/.htpasswd",
		"/robots.txt",
		"/sitemap.xml",
		"/sitemap.xml.gz",
		"/backup.zip",
		"/backup.tar.gz",
		"/backup.sql",
		"/dump.sql",
		"/database.sql",
		"/db.sql",
		"/phpinfo.php",
		"/info.php",
		"/.DS_Store",
		"/Thumbs.db",
		"/server-status",
		"/.well-known/security.txt",
		"/crossdomain.xml",
		"/clientaccesspolicy.xml",
		"/.dockerignore",
		"/docker-compose.yml",
		"/Dockerfile",
		"/package.json",
		"/composer.json",
		"/composer.lock",
		"/yarn.lock",
		"/requirements.txt",
		"/Gemfile",
		"/Gemfile.lock",
		"/pom.xml",
		"/build.gradle",
		"/.npmrc",
		"/.pypirc",
		"/web.config",
		"/Global.asax",
		"/appsettings.json",
		"/connectionStrings.config",
	}

	var findings []LeakFinding

	for _, path := range leakPaths {
		target := baseURL + path
		resp, err := utils.Fetch(target, nil)
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			content := string(body)

			// Check if it's actually the file (not a 404 page)
			if len(content) > 0 && len(content) < 100000 {
				findings = append(findings, LeakFinding{
					URL:     target,
					Type:    path,
					Size:    len(content),
					Preview: truncateString(content, 200),
				})
			}
		}
		resp.Body.Close()
	}

	return findings, nil
}

type LeakFinding struct {
	URL     string `json:"url"`
	Type    string `json:"type"`
	Size    int    `json:"size"`
	Preview string `json:"preview"`
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ArchiveCrawler queries Wayback Machine and CommonCrawl
func ArchiveCrawler(domain string) ([]string, error) {
	fmt.Printf("[+] Querying archives for: %s\n", domain)

	var urls []string

	// Wayback Machine CDX API
	waybackURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&matchType=domain&fl=original&collapse=urlkey", domain)
	resp, err := utils.Fetch(waybackURL, nil)
	if err == nil {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				urls = append(urls, line)
			}
		}
		resp.Body.Close()
	}

	fmt.Printf("[+] Found %d URLs from archives\n", len(urls))
	return utils.Unique(urls), nil
}

// RobotsAbuser parses robots.txt and meta robots tags
func RobotsAbuser(baseURL string) ([]string, error) {
	robotsURL := baseURL + "/robots.txt"
	resp, err := utils.Fetch(robotsURL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var disallowed []string
	lines := strings.Split(string(body), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Disallow:") || strings.HasPrefix(line, "Allow:") {
			path := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "Disallow:"), "Allow:"))
			if path != "" && path != "/" {
				disallowed = append(disallowed, baseURL+path)
			}
		}
	}

	return disallowed, nil
}
