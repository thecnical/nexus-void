package intel

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ScreenshotCapturer takes screenshots of vulnerable URLs
type ScreenshotCapturer struct {
	outputDir string
	client    *http.Client
}

// NewScreenshotCapturer creates screenshot engine
func NewScreenshotCapturer(outputDir string) *ScreenshotCapturer {
	if outputDir == "" {
		outputDir = "reports/screenshots"
	}
	os.MkdirAll(outputDir, 0755)
	return &ScreenshotCapturer{
		outputDir: outputDir,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// CapturePage fetches the HTML/error page from a vulnerable URL
func (s *ScreenshotCapturer) CapturePage(url string, vulnType string) (string, error) {
	// Sanitize filename
	filename := sanitizeFilename(url, vulnType) + ".html"
	filepath_ := filepath.Join(s.outputDir, filename)

	// Fetch the page
	resp, err := s.client.Get(url)
	if err != nil {
		// Save error info even if page doesn't load
		errContent := fmt.Sprintf("<!-- Screenshot for %s (%s) -->\n<!-- Error: %v -->\n", url, vulnType, err)
		os.WriteFile(filepath_, []byte(errContent), 0644)
		return filepath_, nil
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Wrap in screenshot HTML with metadata
	screenshotHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Nexus-Void Screenshot - %s</title>
<style>
body { font-family: monospace; background: #1a1a2e; color: #eee; padding: 20px; }
.header { background: #16213e; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
.meta { color: #0f3460; }
.url { color: #e94560; font-size: 1.2em; }
.vuln { color: #f39c12; font-weight: bold; }
.timestamp { color: #95a5a6; }
.content { background: #0f0f23; padding: 20px; border-radius: 8px; overflow: auto; }
pre { white-space: pre-wrap; word-break: break-all; }
.tag { display: inline-block; background: #e94560; color: white; padding: 3px 8px; border-radius: 4px; margin: 2px; }
</style>
</head>
<body>
<div class="header">
<h1>NEXUS-VOID SCREENSHOT</h1>
<div class="url">URL: %s</div>
<div class="vuln">Vulnerability: %s</div>
<div class="timestamp">Captured: %s</div>
<div class="meta">Status: %d | Content-Type: %s | Length: %d bytes</div>
</div>
<div class="content">
<h3>Response Body:</h3>
<pre>%s</pre>
</div>
<div class="header" style="margin-top:20px">
<h3>HTTP Headers:</h3>
<pre>%s</pre>
</div>
</body>
</html>`,
		url,
		url,
		vulnType,
		time.Now().Format("2006-01-02 15:04:05"),
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		len(body),
		escapeHTML(string(body)),
		formatHeaders(resp.Header),
	)

	err = os.WriteFile(filepath_, []byte(screenshotHTML), 0644)
	if err != nil {
		return "", err
	}

	return filepath_, nil
}

// CaptureVulnerabilityFindings creates screenshots for all findings
func CaptureVulnerabilityFindings(findings []struct {
	URL       string
	Type      string
	Parameter string
	Payload   string
}, outputDir string) []string {
	capturer := NewScreenshotCapturer(outputDir)
	var captured []string

	for _, f := range findings {
		if f.URL == "" {
			continue
		}
		// Append payload to URL for evidence
		testURL := f.URL
		if f.Parameter != "" && f.Payload != "" {
			if strings.Contains(testURL, "?") {
				testURL += "&" + f.Parameter + "=" + f.Payload
			} else {
				testURL += "?" + f.Parameter + "=" + f.Payload
			}
		}

		path, err := capturer.CapturePage(testURL, f.Type)
		if err == nil {
			captured = append(captured, path)
		}
	}

	return captured
}

func sanitizeFilename(url, vuln string) string {
	name := strings.ReplaceAll(url, "https://", "")
	name = strings.ReplaceAll(name, "http://", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "&", "_")
	name = strings.ReplaceAll(name, "=", "_")
	if len(name) > 100 {
		name = name[:100]
	}
	return fmt.Sprintf("%s_%s_%d", name, vuln, time.Now().Unix())
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func formatHeaders(h http.Header) string {
	var lines []string
	for k, v := range h {
		lines = append(lines, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}
	return strings.Join(lines, "\n")
}
