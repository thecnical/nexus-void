package recon

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Result holds reconnaissance findings
type Result struct {
	Target     string            `json:"target"`
	StatusCode int               `json:"status_code"`
	Server     string            `json:"server"`
	Title      string            `json:"title"`
	TechStack  []string          `json:"tech_stack"`
	Headers    map[string]string `json:"headers"`
	Endpoints  []string          `json:"endpoints"`
	Subdomains []string          `json:"subdomains"`
	Emails     []string          `json:"emails"`
	Duration   time.Duration     `json:"duration"`
	Timestamp  time.Time         `json:"timestamp"`
}

var (
	techPatterns = map[string][]string{
		"WordPress":  {`wp-content`, `wp-includes`, `/xmlrpc.php`},
		"Drupal":     {`Drupal`, `sites/default`},
		"Joomla":     {`Joomla`, `/administrator/`},
		"PHP":        {`.php`, `X-Powered-By:.*PHP`},
		"ASP.NET":    {`ASP.NET`, `__VIEWSTATE`, `.aspx`},
		"Node.js":    {`Express`, `Next.js`, `node.js`},
		"Python":     {`Python`, `Django`, `Flask`, `Werkzeug`},
		"Ruby":       {`Ruby`, `Rails`},
		"Java":       {`JSP`, `Servlet`, `Spring`, `Tomcat`},
		"Nginx":      {`nginx`},
		"Apache":     {`Apache`},
		"IIS":        {`IIS`, `Microsoft-IIS`},
		"Cloudflare": {`cloudflare`},
		"AWS":        {`aws`, `amazonaws`, `x-amz`},
		"React":      {`react`, `__REACT_INSPECTOR__`},
		"Vue.js":     {`vue`, `v-if`, `v-for`},
		"Angular":    {`ng-`, `angular`},
		"jQuery":     {`jquery`},
		"Bootstrap":  {`bootstrap`},
		"Laravel":    {`laravel`},
		"Shopify":    {`shopify`, `myshopify`},
		"Magento":    {`magento`},
	}

	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
)

// Scan performs full reconnaissance on a target
func Scan(target string) (*Result, error) {
	start := time.Now()
	result := &Result{
		Target:    target,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}

	// Normalize URL
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	// Fetch main page
	resp, err := httpClient.Get(target)
	if err != nil {
		// Try HTTP if HTTPS fails
		target = strings.Replace(target, "https://", "http://", 1)
		resp, err = httpClient.Get(target)
		if err != nil {
			return result, nil // Return partial result
		}
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Server = resp.Header.Get("Server")

	// Capture headers
	for k, v := range resp.Header {
		result.Headers[k] = strings.Join(v, ", ")
	}

	// Read body
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	bodyStr := string(body)

	// Extract title
	if m := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`).FindStringSubmatch(bodyStr); len(m) > 1 {
		result.Title = strings.TrimSpace(m[1])
	}

	// Detect tech stack
	lowerBody := strings.ToLower(bodyStr)
	for tech, patterns := range techPatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerBody, strings.ToLower(pattern)) ||
				strings.Contains(strings.ToLower(result.Server), strings.ToLower(pattern)) {
				result.TechStack = appendUnique(result.TechStack, tech)
				break
			}
		}
	}

	// Find endpoints
	endpointRegex := regexp.MustCompile(`(?:href|src|action)=["']([^"']+)["']`)
	matches := endpointRegex.FindAllStringSubmatch(bodyStr, -1)
	for _, m := range matches {
		if len(m) > 1 {
			endpoint := m[1]
			if strings.HasPrefix(endpoint, "/") || strings.HasPrefix(endpoint, "http") {
				result.Endpoints = appendUnique(result.Endpoints, endpoint)
			}
		}
	}

	// Find emails
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindAllString(bodyStr, -1)
	for _, e := range emails {
		result.Emails = appendUnique(result.Emails, e)
	}

	// Subdomain brute force (common prefixes)
	commonSubs := []string{"www", "mail", "ftp", "admin", "api", "blog", "shop", "dev", "staging", "test", "portal", "cdn"}
	var wg sync.WaitGroup
	subChan := make(chan string, len(commonSubs))
	for _, sub := range commonSubs {
		wg.Add(1)
		go func(prefix string) {
			defer wg.Done()
			subURL := fmt.Sprintf("https://%s.%s", prefix, u.Host)
			if r, err := httpClient.Head(subURL); err == nil {
				r.Body.Close()
				if r.StatusCode < 400 {
					subChan <- prefix + "." + u.Host
				}
			}
		}(sub)
	}

	go func() {
		wg.Wait()
		close(subChan)
	}()

	for sub := range subChan {
		result.Subdomains = append(result.Subdomains, sub)
	}

	result.Duration = time.Since(start)
	return result, nil
}

// QuickProbe just checks if target is alive and gets basic info
func QuickProbe(target string) (int, string, error) {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	resp, err := httpClient.Head(target)
	if err != nil {
		target = strings.Replace(target, "https://", "http://", 1)
		resp, err = httpClient.Head(target)
		if err != nil {
			return 0, "", err
		}
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Server"), nil
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
