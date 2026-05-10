// OSINTBREACH — Autonomous Reconnaissance & Attack Surface Weapon
// 6 AI Agents: ALPHA-RECON, BETA-SURFACE, GAMMA-PERSONA, DELTA-VULN, EPSILON-SUPPLY, OMEGA-BRAIN

package osintbreach

import (
	"sync"
	"time"
)

// ─── Target & Asset Types ────────────────────────────────────────────

type Target struct {
	ID          string
	Domain      string
	IPs         []string
	Subdomains  []Subdomain
	People      []Person
	Credentials []CredentialLeak
	CloudAssets []CloudAsset
	APIs        []APIEndpoint
	Vulns       []Vulnerability
	Secrets     []Secret
	Deps        []Dependency
	Score       float64 // attack surface score
}

type Subdomain struct {
	Name       string
	IP         string
	StatusCode int
	Tech       string
	Ports      []int
	Live       bool
	Takeover   bool // subdomain takeover possible
}

type Person struct {
	Name        string
	Email       string
	Phone       string
	Social      map[string]string // platform -> username
	Role        string
	BreachCount int
}

type CredentialLeak struct {
	Email    string
	Password string // hashed/fragment
	Source   string // haveibeenpwned, breach db
	Date     time.Time
}

type CloudAsset struct {
	Type     string // s3, azure-blob, gcp-bucket, cloudfront
	Name     string
	URL      string
	Public   bool
	Writable bool
	Region   string
}

type APIEndpoint struct {
	URL        string
	Method     string
	StatusCode int
	Tech       string
	AuthType   string
	Params     []string
	Headers    map[string]string
}

type Vulnerability struct {
	Name       string
	Severity   string // critical, high, medium, low
	URL        string
	Type       string // xss, sqli, ssrf, lfi, rce, info-disclosure
	Tool       string
	Confidence float64
	Poc        string
}

type Secret struct {
	Type   string // api_key, token, password, aws_key, private_key
	Value  string
	Source string // git, js, response
	URL    string
}

type Dependency struct {
	Name       string
	Version    string
	Type       string   // npm, pip, maven, go, cargo
	Vulns      []string // CVE IDs
	Direct     bool
	Transitive bool
}

// ─── Agent Communication ──────────────────────────────────────────────

type AgentMessage struct {
	From      string
	To        string
	Type      string
	Data      string
	Timestamp time.Time
}

type EventBus struct {
	subscribers map[string]chan AgentMessage
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[string]chan AgentMessage)}
}

func (eb *EventBus) Subscribe(agent string, ch chan AgentMessage) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[agent] = ch
}

func (eb *EventBus) Broadcast(msg AgentMessage) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	msg.Timestamp = time.Now()
	for _, ch := range eb.subscribers {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ─── Shared State ────────────────────────────────────────────────────

type SharedState struct {
	mu          sync.RWMutex
	Target      *Target
	Subdomains  map[string]*Subdomain
	People      map[string]*Person
	Vulns       []Vulnerability
	Secrets     []Secret
	CloudAssets []CloudAsset
	APIs        []APIEndpoint
	Deps        []Dependency
	Finished    bool
}

// ─── Agent Interface ────────────────────────────────────────────────

type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}

// ─── Attack Chain ────────────────────────────────────────────────────

type ReconStep struct {
	Agent   string
	Action  string
	Tool    string
	Input   string
	Output  string
	Timeout time.Duration
}

type AttackSurface struct {
	Domain          string
	Score           float64
	LiveHosts       int
	ExposedAPIs     int
	LeakedCreds     int
	VulnCount       int
	CloudExposures  int
	SupplyChainRisk int
	Priority        string // critical, high, medium
}
