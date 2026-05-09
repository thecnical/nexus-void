package intel

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
	"time"
)

// EmailConfig holds SMTP settings
type EmailConfig struct {
	SMTPHost  string
	SMTPPort  string
	From      string
	Password  string
	To        []string
	EnableTLS bool
}

// ReportEmail is the alert email content
type ReportEmail struct {
	Target     string
	Timestamp  string
	Findings   int
	Exploits   int
	Severity   string
	Categories []string
	Summary    string
}

const emailTemplate = `From: Nexus-Void AI <{{.From}}>
To: {{.To}}
Subject: [NEXUS-VOID] CRITICAL: {{.Severity}} findings on {{.Target}}
MIME-Version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
<head>
<style>
body { font-family: Arial, sans-serif; background: #1a1a2e; color: #eee; padding: 20px; }
.header { background: #16213e; padding: 20px; border-radius: 10px; text-align: center; }
.severity-critical { color: #e74c3c; font-weight: bold; font-size: 1.5em; }
.severity-high { color: #f39c12; font-weight: bold; font-size: 1.3em; }
.severity-medium { color: #3498db; font-weight: bold; }
.stats { display: flex; justify-content: space-around; margin: 20px 0; }
.stat-box { background: #0f3460; padding: 15px; border-radius: 8px; text-align: center; min-width: 120px; }
.stat-number { font-size: 2em; color: #e94560; }
.footer { margin-top: 30px; padding: 15px; background: #16213e; border-radius: 8px; }
</style>
</head>
<body>
<div class="header">
<h1>NEXUS-VOID AUTONOMOUS ASSAULT REPORT</h1>
<div class="severity-{{.SeverityClass}}">SEVERITY: {{.Severity}}</div>
<p>Target: <strong>{{.Target}}</strong></p>
<p>Time: {{.Timestamp}}</p>
</div>

<div class="stats">
<div class="stat-box">
<div class="stat-number">{{.Findings}}</div>
<div>Findings</div>
</div>
<div class="stat-box">
<div class="stat-number">{{.Exploits}}</div>
<div>Exploits</div>
</div>
<div class="stat-box">
<div class="stat-number">{{.Categories}}</div>
<div>Categories</div>
</div>
</div>

<div class="footer">
<p><strong>Summary:</strong></p>
<p>{{.Summary}}</p>
<p style="color: #95a5a6; font-size: 0.9em;">
This is an automated alert from Nexus-Void AI Swarm.<br>
Full report available in the dashboard.
</p>
</div>
</body>
</html>`

// EmailNotifier sends attack alerts via email
type EmailNotifier struct {
	config EmailConfig
}

// NewEmailNotifier creates email alert system
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{config: config}
}

// SendAlert sends email when findings exceed threshold
func (e *EmailNotifier) SendAlert(target string, findings, exploits int, categories []string) error {
	if e.config.SMTPHost == "" {
		return fmt.Errorf("SMTP not configured")
	}

	severity := "LOW"
	severityClass := "medium"
	if exploits > 0 {
		severity = "CRITICAL"
		severityClass = "critical"
	} else if findings > 5 {
		severity = "HIGH"
		severityClass = "high"
	} else if findings > 0 {
		severity = "MEDIUM"
		severityClass = "medium"
	}

	data := struct {
		From          string
		To            string
		Target        string
		Timestamp     string
		Findings      int
		Exploits      int
		Categories    int
		Severity      string
		SeverityClass string
		Summary       string
	}{
		From:          e.config.From,
		To:            e.config.To[0],
		Target:        target,
		Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
		Findings:      findings,
		Exploits:      exploits,
		Categories:    len(categories),
		Severity:      severity,
		SeverityClass: severityClass,
		Summary:       fmt.Sprintf("Nexus-Void completed a full-spectrum attack on %s using %d categories. Found %d vulnerabilities with %d confirmed exploits.", target, len(categories), findings, exploits),
	}

	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Send via SMTP
	addr := fmt.Sprintf("%s:%s", e.config.SMTPHost, e.config.SMTPPort)
	auth := smtp.PlainAuth("", e.config.From, e.config.Password, e.config.SMTPHost)

	msg := body.Bytes()
	err = smtp.SendMail(addr, auth, e.config.From, e.config.To, msg)
	if err != nil {
		return err
	}

	return nil
}

// AutoAlert checks if alert should be sent and sends it
func AutoAlert(config EmailConfig, target string, findings, exploits int, categories []string) {
	// Only alert on findings
	if findings == 0 {
		return
	}

	notifier := NewEmailNotifier(config)
	if err := notifier.SendAlert(target, findings, exploits, categories); err != nil {
		fmt.Printf("[EMAIL] Alert not sent: %v\n", err)
	} else {
		fmt.Printf("[EMAIL] Alert sent to %v\n", config.To)
	}
}
