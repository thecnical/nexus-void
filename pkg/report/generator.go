package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Report represents a complete engagement report
type Report struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Target      string       `json:"target"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     time.Time    `json:"end_time"`
	Findings    []Finding    `json:"findings"`
	Exploits    []Exploit    `json:"exploits"`
	Statistics  Statistics   `json:"statistics"`
	AgentEvents []AgentEvent `json:"agent_events"`
}

// Finding represents a discovered vulnerability
type Finding struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Parameter   string  `json:"parameter"`
	Payload     string  `json:"payload"`
	Evidence    string  `json:"evidence"`
	Confidence  int     `json:"confidence"`
	CVSS        float64 `json:"cvss"`
	Remediation string  `json:"remediation"`
}

// Exploit represents a successful exploit
type Exploit struct {
	Type      string    `json:"type"`
	URL       string    `json:"url"`
	Payload   string    `json:"payload"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

// Statistics holds engagement metrics
type Statistics struct {
	TotalFindings      int    `json:"total_findings"`
	CriticalCount      int    `json:"critical_count"`
	HighCount          int    `json:"high_count"`
	MediumCount        int    `json:"medium_count"`
	LowCount           int    `json:"low_count"`
	ExploitsSuccessful int    `json:"exploits_successful"`
	ToolsUsed          int    `json:"tools_used"`
	Duration           string `json:"duration"`
}

// AgentEvent represents an agent action
type AgentEvent struct {
	AgentName string    `json:"agent_name"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// Generator creates reports
type Generator struct {
	outputDir string
}

// NewGenerator creates a report generator
func NewGenerator(outputDir string) *Generator {
	if outputDir == "" {
		outputDir = "reports"
	}
	os.MkdirAll(outputDir, 0755)
	return &Generator{outputDir: outputDir}
}

// GenerateHTML creates an HTML report
func (g *Generator) GenerateHTML(r *Report) (string, error) {
	fileName := fmt.Sprintf("%s_%s.html", sanitize(r.Target), r.ID)
	path := filepath.Join(g.outputDir, fileName)

	tmpl := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>NEXUS-VOID Report: {{.Target}}</title>
<style>
body { font-family: 'Segoe UI', Arial, sans-serif; background: #0a0a1a; color: #e0e0e0; margin: 0; padding: 2rem; }
.container { max-width: 1200px; margin: 0 auto; }
h1 { color: #ff6b6b; border-bottom: 2px solid #ff6b6b; padding-bottom: 1rem; }
.summary { background: #1a1a2e; padding: 1.5rem; border-radius: 8px; margin: 1rem 0; }
.stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 1rem; margin: 1rem 0; }
.stat-box { background: #12122a; padding: 1rem; border-radius: 6px; text-align: center; border: 1px solid #2a2a4e; }
.stat-value { font-size: 2rem; font-weight: bold; color: #00ff88; }
.stat-label { font-size: 0.85rem; color: #888; }
.finding { background: #1a1a2e; padding: 1rem; margin: 0.5rem 0; border-radius: 6px; border-left: 4px solid; }
.critical { border-color: #ff3333; }
.high { border-color: #ff6b6b; }
.medium { border-color: #ffaa00; }
.low { border-color: #00ccff; }
.severity-badge { display: inline-block; padding: 0.2rem 0.6rem; border-radius: 4px; font-size: 0.75rem; font-weight: bold; }
.severity-critical { background: #ff3333; }
.severity-high { background: #ff6b6b; }
.severity-medium { background: #ffaa00; color: #000; }
.severity-low { background: #00ccff; color: #000; }
table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
th { background: #1a1a2e; padding: 0.75rem; text-align: left; border-bottom: 2px solid #2a2a4e; }
td { padding: 0.75rem; border-bottom: 1px solid #2a2a4e; }
.payload { background: #12122a; padding: 0.75rem; border-radius: 4px; font-family: monospace; overflow-x: auto; }
</style>
</head>
<body>
<div class="container">
<h1>NEXUS-VOID OMEGA - Engagement Report</h1>

<div class="summary">
<h2>Executive Summary</h2>
<p><strong>Target:</strong> {{.Target}}</p>
<p><strong>Duration:</strong> {{.Statistics.Duration}}</p>
<p><strong>Report ID:</strong> {{.ID}}</p>
<p><strong>Generated:</strong> {{.EndTime.Format "2006-01-02 15:04:05"}}</p>
</div>

<div class="stat-grid">
<div class="stat-box">
<div class="stat-value">{{.Statistics.TotalFindings}}</div>
<div class="stat-label">Total Findings</div>
</div>
<div class="stat-box">
<div class="stat-value" style="color:#ff3333">{{.Statistics.CriticalCount}}</div>
<div class="stat-label">Critical</div>
</div>
<div class="stat-box">
<div class="stat-value" style="color:#ff6b6b">{{.Statistics.HighCount}}</div>
<div class="stat-label">High</div>
</div>
<div class="stat-box">
<div class="stat-value" style="color:#ffaa00">{{.Statistics.MediumCount}}</div>
<div class="stat-label">Medium</div>
</div>
<div class="stat-box">
<div class="stat-value" style="color:#00ccff">{{.Statistics.LowCount}}</div>
<div class="stat-label">Low</div>
</div>
<div class="stat-box">
<div class="stat-value" style="color:#00ff88">{{.Statistics.ExploitsSuccessful}}</div>
<div class="stat-label">Exploits</div>
</div>
</div>

<h2>Findings ({{len .Findings}})</h2>
{{range .Findings}}
<div class="finding {{.Severity}}">
<span class="severity-badge severity-{{.Severity}}">{{upper .Severity}}</span>
<strong>{{.Title}}</strong>
<p><strong>URL:</strong> {{.URL}}</p>
<p><strong>Parameter:</strong> {{.Parameter}}</p>
<p><strong>Confidence:</strong> {{.Confidence}}%</p>
<p><strong>Evidence:</strong> {{.Evidence}}</p>
{{if .Payload}}<div class="payload">{{.Payload}}</div>{{end}}
{{if .Remediation}}<p><strong>Remediation:</strong> {{.Remediation}}</p>{{end}}
</div>
{{end}}

<h2>Exploits ({{len .Exploits}})</h2>
<table>
<tr><th>Type</th><th>Target</th><th>Payload</th><th>Timestamp</th></tr>
{{range .Exploits}}
<tr>
<td>{{.Type}}</td>
<td>{{.URL}}</td>
<td><code>{{.Payload}}</code></td>
<td>{{.Timestamp.Format "15:04:05"}}</td>
</tr>
{{end}}
</table>

<h2>Agent Activity</h2>
<table>
<tr><th>Agent</th><th>Action</th><th>Target</th><th>Result</th></tr>
{{range .AgentEvents}}
<tr>
<td>{{.AgentName}}</td>
<td>{{.Action}}</td>
<td>{{.Target}}</td>
<td>{{.Result}}</td>
</tr>
{{end}}
</table>

<footer style="margin-top: 3rem; padding-top: 1rem; border-top: 1px solid #2a2a4e; color: #888; font-size: 0.85rem;">
Generated by NEXUS-VOID OMEGA | Classification: CONFIDENTIAL
</footer>
</div>
</body>
</html>`

	t := template.New("report").Funcs(template.FuncMap{
		"upper": strings.ToUpper,
	})
	t, err := t.Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file error: %w", err)
	}
	defer f.Close()

	if err := t.Execute(f, r); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	return path, nil
}

// GenerateJSON creates a JSON report
func (g *Generator) GenerateJSON(r *Report) (string, error) {
	fileName := fmt.Sprintf("%s_%s.json", sanitize(r.Target), r.ID)
	path := filepath.Join(g.outputDir, fileName)

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}

	return path, nil
}

// GenerateMarkdown creates a Markdown report
func (g *Generator) GenerateMarkdown(r *Report) (string, error) {
	fileName := fmt.Sprintf("%s_%s.md", sanitize(r.Target), r.ID)
	path := filepath.Join(g.outputDir, fileName)

	content := fmt.Sprintf(`# NEXUS-VOID OMEGA Engagement Report

**Target:** %s  
**Duration:** %s  
**Report ID:** %s  
**Generated:** %s

## Executive Summary

| Metric | Value |
|--------|-------|
| Total Findings | %d |
| Critical | %d |
| High | %d |
| Medium | %d |
| Low | %d |
| Successful Exploits | %d |
| Tools Used | %d |

## Findings

`, r.Target, r.Statistics.Duration, r.ID, r.EndTime.Format(time.RFC3339),
		r.Statistics.TotalFindings, r.Statistics.CriticalCount, r.Statistics.HighCount,
		r.Statistics.MediumCount, r.Statistics.LowCount, r.Statistics.ExploitsSuccessful,
		r.Statistics.ToolsUsed)

	for _, f := range r.Findings {
		content += fmt.Sprintf("### [%s] %s\n\n", f.Severity, f.Title)
		content += fmt.Sprintf("- **URL:** %s\n", f.URL)
		content += fmt.Sprintf("- **Parameter:** %s\n", f.Parameter)
		content += fmt.Sprintf("- **Confidence:** %d%%\n", f.Confidence)
		content += fmt.Sprintf("- **Evidence:** %s\n", f.Evidence)
		if f.Payload != "" {
			content += fmt.Sprintf("- **Payload:** `%s`\n", f.Payload)
		}
		if f.Remediation != "" {
			content += fmt.Sprintf("- **Remediation:** %s\n", f.Remediation)
		}
		content += "\n"
	}

	content += "## Exploits\n\n"
	for _, e := range r.Exploits {
		content += fmt.Sprintf("- **%s** on %s at %s\n", e.Type, e.URL, e.Timestamp.Format("15:04:05"))
		content += fmt.Sprintf("  - Payload: `%s`\n", e.Payload)
		if e.Output != "" {
			content += fmt.Sprintf("  - Output: `%s`\n", e.Output)
		}
	}

	content += "\n## Agent Activity\n\n"
	content += "| Agent | Action | Target | Result |\n"
	content += "|-------|--------|--------|--------|\n"
	for _, e := range r.AgentEvents {
		content += fmt.Sprintf("| %s | %s | %s | %s |\n", e.AgentName, e.Action, e.Target, e.Result)
	}

	content += "\n---\n*Generated by NEXUS-VOID OMEGA | Classification: CONFIDENTIAL*\n"

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}

	return path, nil
}

// NewReport creates a new report instance
func NewReport(target string) *Report {
	return &Report{
		ID:          fmt.Sprintf("RPT-%d", time.Now().Unix()),
		Title:       fmt.Sprintf("NEXUS-VOID Engagement: %s", target),
		Target:      target,
		StartTime:   time.Now(),
		Findings:    []Finding{},
		Exploits:    []Exploit{},
		Statistics:  Statistics{},
		AgentEvents: []AgentEvent{},
	}
}

// AddFinding adds a finding to the report
func (r *Report) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
	r.Statistics.TotalFindings++
	switch f.Severity {
	case "critical":
		r.Statistics.CriticalCount++
	case "high":
		r.Statistics.HighCount++
	case "medium":
		r.Statistics.MediumCount++
	case "low":
		r.Statistics.LowCount++
	}
}

// AddExploit adds an exploit to the report
func (r *Report) AddExploit(e Exploit) {
	r.Exploits = append(r.Exploits, e)
	r.Statistics.ExploitsSuccessful++
}

// AddAgentEvent adds an agent event
func (r *Report) AddAgentEvent(e AgentEvent) {
	r.AgentEvents = append(r.AgentEvents, e)
}

// Finalize completes the report
func (r *Report) Finalize() {
	r.EndTime = time.Now()
	duration := r.EndTime.Sub(r.StartTime)
	r.Statistics.Duration = fmt.Sprintf("%dh %dm %ds", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
}

func sanitize(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}
