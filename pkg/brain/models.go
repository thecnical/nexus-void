package brain

import (
	"gorm.io/gorm"
	"time"
)

// Node represents any entity in the knowledge graph
type Node struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"index"` // target, subdomain, endpoint, vuln, exploit, defense, credential
	Label     string         `json:"label"`
	Data      string         `json:"data" gorm:"type:text"` // JSON blob
	TargetID  string         `json:"target_id" gorm:"index"`
	SessionID string         `json:"session_id" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Edge represents relationships between nodes
type Edge struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	SourceID  string    `json:"source_id" gorm:"index"`
	TargetID  string    `json:"target_id" gorm:"index"`
	Relation  string    `json:"relation" gorm:"index"` // has_subdomain, has_vulnerability, exploits, generates_defense
	Weight    float64   `json:"weight"`                // Confidence 0-1
	SessionID string    `json:"session_id" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

// Finding represents a discovered vulnerability
type Finding struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	TargetID       string    `json:"target_id" gorm:"index"`
	SessionID      string    `json:"session_id" gorm:"index"`
	Type           string    `json:"type"`     // sqli, xss, rce, lfi, etc.
	Severity       string    `json:"severity"` // critical, high, medium, low, info
	Title          string    `json:"title"`
	Description    string    `json:"description" gorm:"type:text"`
	URL            string    `json:"url"`
	Parameter      string    `json:"parameter"`
	Payload        string    `json:"payload" gorm:"type:text"`
	Evidence       string    `json:"evidence" gorm:"type:text"` // request/response
	CVSS           float64   `json:"cvss"`
	CWE            string    `json:"cwe"`
	CVE            string    `json:"cve"`
	ProofOfConcept bool      `json:"proof_of_concept"`
	Status         string    `json:"status"` // discovered, confirmed, exploited, verified
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Exploit represents a proven exploit
type Exploit struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	FindingID  string    `json:"finding_id" gorm:"index"`
	SessionID  string    `json:"session_id" gorm:"index"`
	Type       string    `json:"type"`
	Payload    string    `json:"payload" gorm:"type:text"`
	Command    string    `json:"command" gorm:"type:text"`
	Output     string    `json:"output" gorm:"type:text"`
	SafeMode   bool      `json:"safe_mode"` // Only harmless commands (whoami, id)
	Evidence   string    `json:"evidence" gorm:"type:text"`
	Screenshot string    `json:"screenshot"` // Path to screenshot
	Timestamp  time.Time `json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
}

// Defense represents a generated countermeasure
type Defense struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	FindingID     string    `json:"finding_id" gorm:"index"`
	SessionID     string    `json:"session_id" gorm:"index"`
	Type          string    `json:"type"` // waf_rule, iptables, ebpf, middleware, nginx
	Title         string    `json:"title"`
	Code          string    `json:"code" gorm:"type:text"`   // Deployable code
	Config        string    `json:"config" gorm:"type:text"` // Configuration
	Language      string    `json:"language"`                // python, go, c, yaml
	Effectiveness float64   `json:"effectiveness"`           // 0-100 score
	Verified      bool      `json:"verified"`
	Deployed      bool      `json:"deployed"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AgentEvent represents AI agent reasoning and actions
type AgentEvent struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index"`
	AgentName string    `json:"agent_name" gorm:"index"` // RECON-OMEGA, EXPLOIT-APOCALYPSE, etc.
	EventType string    `json:"event_type"`              // reasoning, action, finding, error
	Message   string    `json:"message" gorm:"type:text"`
	Context   string    `json:"context" gorm:"type:text"` // AI prompt/response
	ToolUsed  string    `json:"tool_used"`
	Target    string    `json:"target"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

// Credential represents discovered credentials
type Credential struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index"`
	TargetID  string    `json:"target_id" gorm:"index"`
	Type      string    `json:"type"` // password, hash, token, key, cookie
	Username  string    `json:"username"`
	Secret    string    `json:"secret" gorm:"type:text"`
	Source    string    `json:"source"` // Where found
	Valid     bool      `json:"valid"`
	CreatedAt time.Time `json:"created_at"`
}

// ToolRun tracks execution of each tool
type ToolRun struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	SessionID  string     `json:"session_id" gorm:"index"`
	ToolName   string     `json:"tool_name"`
	Command    string     `json:"command" gorm:"type:text"`
	Status     string     `json:"status"` // running, success, failed, timeout
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	OutputSize int64      `json:"output_size"`
	Findings   int        `json:"findings"`
	Errors     int        `json:"errors"`
}
