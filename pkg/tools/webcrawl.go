package tools

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// crawlResult holds findings from a crawl
type crawlResult struct {
	URLs       []string
	Forms      []map[string]string
	Comments   []string
	Endpoints  []string
	Leaks      []string
	Parameters []string
}

// === 1. CRAWLER-X ===
func runCrawlerX(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Extract all links
	linkRe := regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	links := linkRe.FindAllStringSubmatch(content, -1)
	var urls []string
	for _, m := range links {
		if len(m) > 1 {
			urls = append(urls, resolveURL(target, m[1]))
		}
	}
	// Extract JS files
	jsRe := regexp.MustCompile(`(?i)src=["']([^"']+\.js)["']`)
	jsFiles := jsRe.FindAllStringSubmatch(content, -1)
	var js []string
	for _, m := range jsFiles {
		if len(m) > 1 {
			js = append(js, resolveURL(target, m[1]))
		}
	}
	// Extract forms
	formRe := regexp.MustCompile(`(?i)<form[^>]*action=["']([^"']*)["']`)
	forms := formRe.FindAllStringSubmatch(content, -1)

	return fmt.Sprintf("[CRAWLER-X] Target: %s\nLinks found: %d\nJS files: %d\nForms: %d\nURLs:\n%s\nJS:\n%s",
		target, len(urls), len(js), len(forms),
		strings.Join(urls, "\n"), strings.Join(js, "\n")), nil
}

// === 2. JS-LINK-FINDER ===
func runJSLinkFinder(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Extract JS src
	jsRe := regexp.MustCompile(`(?i)src=["']([^"']+\.js)["']`)
	matches := jsRe.FindAllStringSubmatch(content, -1)
	var jsURLs []string
	for _, m := range matches {
		if len(m) > 1 {
			jsURLs = append(jsURLs, resolveURL(target, m[1]))
		}
	}

	// For each JS file, extract API endpoints
	var endpoints []string
	for _, jsURL := range jsURLs {
		jsResp, err := http.Get(jsURL)
		if err != nil {
			continue
		}
		jsBody, _ := io.ReadAll(jsResp.Body)
		jsResp.Body.Close()
		jsContent := string(jsBody)

		// Find API endpoints in JS
		apiRe := regexp.MustCompile(`["'](/api/[^"']+)["']|["']([^"']*/v\d+/[^"']*)["']`)
		apiMatches := apiRe.FindAllStringSubmatch(jsContent, -1)
		for _, am := range apiMatches {
			if len(am) > 1 && am[1] != "" {
				endpoints = append(endpoints, am[1])
			} else if len(am) > 2 && am[2] != "" {
				endpoints = append(endpoints, am[2])
			}
		}
		// Find fetch/axios URLs
		fetchRe := regexp.MustCompile(`(?i)fetch\(["']([^"']+)["']\)|axios\.(?:get|post|put|delete)\(["']([^"']+)["']\)`)
		fetchMatches := fetchRe.FindAllStringSubmatch(jsContent, -1)
		for _, fm := range fetchMatches {
			if len(fm) > 1 && fm[1] != "" {
				endpoints = append(endpoints, fm[1])
			} else if len(fm) > 2 && fm[2] != "" {
				endpoints = append(endpoints, fm[2])
			}
		}
	}

	return fmt.Sprintf("[JS-LINK-FINDER] Target: %s\nJS files: %d\nAPI endpoints: %d\n%s",
		target, len(jsURLs), len(endpoints), strings.Join(endpoints, "\n")), nil
}

// === 3. FORM-HUNTER ===
func runFormHunter(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Find all forms
	formRe := regexp.MustCompile(`(?i)<form[^>]*>(.*?)</form>`)
	forms := formRe.FindAllString(content, -1)
	var results []string
	for i, form := range forms {
		actionRe := regexp.MustCompile(`(?i)action=["']([^"']*)["']`)
		actionMatch := actionRe.FindStringSubmatch(form)
		action := ""
		if len(actionMatch) > 1 {
			action = actionMatch[1]
		}
		methodRe := regexp.MustCompile(`(?i)method=["']([^"']*)["']`)
		methodMatch := methodRe.FindStringSubmatch(form)
		method := "GET"
		if len(methodMatch) > 1 {
			method = strings.ToUpper(methodMatch[1])
		}
		// Extract inputs
		inputRe := regexp.MustCompile(`(?i)<input[^>]*name=["']([^"']+)["']`)
		inputs := inputRe.FindAllStringSubmatch(form, -1)
		var inputNames []string
		for _, inp := range inputs {
			if len(inp) > 1 {
				inputNames = append(inputNames, inp[1])
			}
		}
		results = append(results, fmt.Sprintf("Form #%d: action=%s method=%s inputs=%v",
			i+1, action, method, inputNames))
	}

	return fmt.Sprintf("[FORM-HUNTER] Target: %s\nForms found: %d\n%s",
		target, len(forms), strings.Join(results, "\n")), nil
}

// === 4. PARAMETER-DISCOVERER ===
func runParameterDiscoverer(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Discover parameters from URLs in the page
	paramRe := regexp.MustCompile(`[?&]([a-zA-Z_][a-zA-Z0-9_]*)=`)
	params := paramRe.FindAllStringSubmatch(content, -1)
	paramSet := make(map[string]bool)
	for _, p := range params {
		if len(p) > 1 {
			paramSet[p[1]] = true
		}
	}

	// Common hidden parameters to test
	commonParams := []string{"debug", "test", "env", "config", "admin", "api", "key", "token",
		"secret", "password", "user", "id", "page", "file", "path", "url", "redirect", "callback",
		"next", "return", "dest", "target", "view", "mode", "action", "cmd", "exec", "run",
		"source", "include", "require", "load", "import", "data", "query", "search", "filter",
		"sort", "order", "limit", "offset", "page_size", "format", "type", "version", "v"}

	var discovered []string
	for _, param := range commonParams {
		testURL := target + "?" + param + "=test"
		tr, err := http.Get(testURL)
		if err != nil {
			continue
		}
		// If parameter is reflected or changes response, it's valid
		trBody, _ := io.ReadAll(tr.Body)
		tr.Body.Close()
		if strings.Contains(string(trBody), "test") || tr.StatusCode != resp.StatusCode {
			discovered = append(discovered, param)
		}
	}

	for p := range paramSet {
		discovered = append(discovered, p)
	}

	return fmt.Sprintf("[PARAMETER-DISCOVERER] Target: %s\nParameters found: %d\n%s",
		target, len(discovered), strings.Join(discovered, "\n")), nil
}

// === 5. ARCHIVE-CRAWLER ===
func runArchiveCrawler(args []string) (string, error) {
	domain := "example.com"
	if len(args) > 0 {
		domain = args[0]
	}
	waybackURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=original", domain)
	resp, err := http.Get(waybackURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Parse URLs from CDX response
	var urls []string
	lines := strings.Split(content, "\n")
	seen := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "[\"original\"]" {
			continue
		}
		// Remove JSON brackets and quotes
		line = strings.Trim(line, "[]")
		line = strings.Trim(line, "\"")
		if line != "" && !seen[line] {
			urls = append(urls, line)
			seen[line] = true
		}
	}

	return fmt.Sprintf("[ARCHIVE-CRAWLER] Domain: %s\nArchive URLs: %d\n%s",
		domain, len(urls), strings.Join(urls[:min(len(urls), 50)], "\n")), nil
}

// === 6. SPA-CRAWLER ===
func runSPACrawler(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Find React/Vue/Angular routes
	routeRe := regexp.MustCompile(`(?i)(?:path|route|url)\s*[:=]\s*["']([^"']+)["']`)
	routes := routeRe.FindAllStringSubmatch(content, -1)
	var foundRoutes []string
	for _, r := range routes {
		if len(r) > 1 {
			foundRoutes = append(foundRoutes, r[1])
		}
	}

	// Find router configurations
	routerRe := regexp.MustCompile(`(?i)(?:Router|Switch|Routes)\s*.*?\{([^}]*)\}`)
	routerMatches := routerRe.FindAllStringSubmatch(content, -1)
	for _, rm := range routerMatches {
		if len(rm) > 1 {
			inner := rm[1]
			innerRoutes := regexp.MustCompile(`["']([^"']+)["']`).FindAllStringSubmatch(inner, -1)
			for _, ir := range innerRoutes {
				if len(ir) > 1 && strings.HasPrefix(ir[1], "/") {
					foundRoutes = append(foundRoutes, ir[1])
				}
			}
		}
	}

	// Hash-based routes
	hashRe := regexp.MustCompile(`(?i)window\.location\.hash\s*=\s*["']([^"']+)["']`)
	hashMatches := hashRe.FindAllStringSubmatch(content, -1)
	for _, hm := range hashMatches {
		if len(hm) > 1 {
			foundRoutes = append(foundRoutes, hm[1])
		}
	}

	return fmt.Sprintf("[SPA-CRAWLER] Target: %s\nRoutes found: %d\n%s",
		target, len(foundRoutes), strings.Join(foundRoutes, "\n")), nil
}

// === 7. RENDER-SPIDER ===
func runRenderSpider(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Extract dynamic content indicators
	dynamic := []string{}
	if strings.Contains(content, "<script") {
		dynamic = append(dynamic, "JavaScript detected")
	}
	if strings.Contains(content, "React") || strings.Contains(content, "react") {
		dynamic = append(dynamic, "React app detected")
	}
	if strings.Contains(content, "vue") || strings.Contains(content, "Vue") {
		dynamic = append(dynamic, "Vue app detected")
	}
	if strings.Contains(content, "angular") || strings.Contains(content, "Angular") {
		dynamic = append(dynamic, "Angular app detected")
	}
	if strings.Contains(content, "data-") {
		dynamic = append(dynamic, "Dynamic data attributes")
	}

	// Extract all resources
	resourceRe := regexp.MustCompile(`(?i)(?:src|href)=["']([^"']+)["']`)
	resources := resourceRe.FindAllStringSubmatch(content, -1)
	var urls []string
	for _, r := range resources {
		if len(r) > 1 {
			urls = append(urls, resolveURL(target, r[1]))
		}
	}

	return fmt.Sprintf("[RENDER-SPIDER] Target: %s\nDynamic indicators: %d\n%s\nResources: %d\n%s",
		target, len(dynamic), strings.Join(dynamic, "\n"), len(urls), strings.Join(urls[:min(len(urls), 30)], "\n")), nil
}

// === 8. COMMENT-HARVESTER ===
func runCommentHarvester(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Extract HTML comments
	commentRe := regexp.MustCompile(`<!--(.*?)-->`)
	comments := commentRe.FindAllString(content, -1)
	var interesting []string
	for _, c := range comments {
		lower := strings.ToLower(c)
		if strings.Contains(lower, "todo") || strings.Contains(lower, "fixme") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "secret") ||
			strings.Contains(lower, "api") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "debug") || strings.Contains(lower, "internal") {
			interesting = append(interesting, strings.TrimSpace(c))
		}
	}

	return fmt.Sprintf("[COMMENT-HARVESTER] Target: %s\nTotal comments: %d\nInteresting: %d\n%s",
		target, len(comments), len(interesting), strings.Join(interesting, "\n")), nil
}

// === 9. SOURCE-LEAK-FINDER ===
func runSourceLeakFinder(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	leakPaths := []string{
		".git/", ".git/config", ".svn/", ".hg/", ".env", ".env.local",
		".env.production", "config.php", "config.inc", "wp-config.php",
		"settings.py", "database.yml", "backup.sql", "dump.sql", "db.sql",
		".htaccess", "web.config", "Dockerfile", "docker-compose.yml",
		"package.json", "requirements.txt", "composer.json", "Gemfile",
		"robots.txt", "sitemap.xml", "crossdomain.xml", "clientaccesspolicy.xml",
		".DS_Store", "Thumbs.db", "*.bak", "*.old", "*.orig", "*.swp", "*.tmp",
		"api-docs", "swagger.json", "swagger.yaml", "openapi.json",
	}
	var found []string
	for _, path := range leakPaths {
		leakURL := target + "/" + path
		resp, err := http.Get(leakURL)
		if err != nil {
			continue
		}
		if resp.StatusCode == 200 {
			found = append(found, leakURL+" [EXPOSED]")
		}
		resp.Body.Close()
	}

	return fmt.Sprintf("[SOURCE-LEAK-FINDER] Target: %s\nLeaks found: %d\n%s",
		target, len(found), strings.Join(found, "\n")), nil
}

// === 10. SITEMAP-ABUSER ===
func runSitemapAbuser(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	sitemapURL := target + "/sitemap.xml"
	resp, err := http.Get(sitemapURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	type URLSet struct {
		URLs []struct {
			Loc string `xml:"url>loc"`
		} `xml:"url"`
	}
	var urlset URLSet
	if err := xml.Unmarshal(body, &urlset); err != nil {
		// Try regex fallback
		locRe := regexp.MustCompile(`<loc>([^<]+)</loc>`)
		matches := locRe.FindAllStringSubmatch(string(body), -1)
		for _, m := range matches {
			if len(m) > 1 {
				urlset.URLs = append(urlset.URLs, struct {
					Loc string `xml:"url>loc"`
				}{Loc: m[1]})
			}
		}
	}

	var urls []string
	for _, u := range urlset.URLs {
		urls = append(urls, u.Loc)
	}

	return fmt.Sprintf("[SITEMAP-ABUSER] Target: %s\nURLs in sitemap: %d\n%s",
		target, len(urls), strings.Join(urls[:min(len(urls), 50)], "\n")), nil
}

// === 11. ROBOTS-ABUSER ===
func runRobotsAbuser(args []string) (string, error) {
	target := "http://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	robotsURL := target + "/robots.txt"
	resp, err := http.Get(robotsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	var disallowed []string
	var allowed []string
	var sitemaps []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Disallow:") {
			disallowed = append(disallowed, strings.TrimSpace(strings.TrimPrefix(line, "Disallow:")))
		} else if strings.HasPrefix(line, "Allow:") {
			allowed = append(allowed, strings.TrimSpace(strings.TrimPrefix(line, "Allow:")))
		} else if strings.HasPrefix(line, "Sitemap:") {
			sitemaps = append(sitemaps, strings.TrimSpace(strings.TrimPrefix(line, "Sitemap:")))
		}
	}

	return fmt.Sprintf("[ROBOTS-ABUSER] Target: %s\nDisallowed: %d\nAllowed: %d\nSitemaps: %d\nDisallowed paths:\n%s\nSitemaps:\n%s",
		target, len(disallowed), len(allowed), len(sitemaps),
		strings.Join(disallowed, "\n"), strings.Join(sitemaps, "\n")), nil
}

// === 12. URL-PARAMETER-FUZZER ===
func runURLParameterFuzzer(args []string) (string, error) {
	target := "http://example.com/page"
	if len(args) > 0 {
		target = args[0]
	}
	wordlist := []string{"id", "page", "file", "path", "url", "redirect", "debug", "test",
		"admin", "user", "name", "password", "token", "key", "secret", "api",
		"action", "cmd", "exec", "source", "include", "data", "query", "search",
		"filter", "sort", "order", "limit", "offset", "format", "type", "version"}

	var found []string
	baseResp, _ := http.Get(target)
	var baseBody string
	if baseResp != nil {
		b, _ := io.ReadAll(baseResp.Body)
		baseBody = string(b)
		baseResp.Body.Close()
	}

	for _, param := range wordlist {
		testURL := target + "?" + param + "=FUZZ"
		resp, err := http.Get(testURL)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// If response differs from base, parameter is processed
		if string(body) != baseBody || resp.StatusCode != baseResp.StatusCode {
			found = append(found, param+" [REFLECTED/PROCESSED]")
		}
	}

	return fmt.Sprintf("[URL-PARAMETER-FUZZER] Target: %s\nParameters found: %d\n%s",
		target, len(found), strings.Join(found, "\n")), nil
}

// Helper: resolve relative URLs
func resolveURL(base, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	baseURL, _ := url.Parse(base)
	refURL, _ := url.Parse(ref)
	if baseURL != nil && refURL != nil {
		return baseURL.ResolveReference(refURL).String()
	}
	return base + "/" + ref
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
