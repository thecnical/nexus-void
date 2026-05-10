// MOBILEBREACH v2 — Next-Gen Autonomous Mobile Penetration System
// Core data structures for 5 AI agents: APEX, GHOST, LANCE, SPECTRE, OVERMIND

package mobilebreach

import (
	"sync"
	"time"
)

// ─── Mobile Target Types ──────────────────────────────────────────────

type MobileTarget struct {
	ID          string
	Type        string // apk, ipa, api, device, cellular
	Name        string
	Path        string // file path or URL
	BundleID    string // com.example.app
	Version     string
	OSVersion   string // Android 13, iOS 17.2
	Platform    string // android, ios
	APIBaseURL  string
	APITech     string // graphql, rest, grpc
	Severity    string // critical, high, medium, low
	Findings    []string
}

// AppInfo holds reverse-engineering results
type AppInfo struct {
	PackageName       string
	VersionName       string
	MinSDK            int
	TargetSDK         int
	Permissions       []string
	Activities        []string
	Services          []string
	Receivers         []string
	Providers         []string
	DeepLinks         []string
	HardcodedSecrets  []string // API keys, tokens found
	SSLPinned         bool     // certificate pinning detected
	RootDetection     bool     // root/jailbreak detection detected
	AntiDebug         bool
	AntiVM            bool
	Obfuscated        bool
}

// DeviceProfile holds connected device info
type DeviceProfile struct {
	Serial     string
	Model      string
	OSVersion  string
	Rooted     bool
	Jailbroken bool
	Connected  bool
	Interface  string // adb, usb, network
}

// ─── 5G / Cellular Data ─────────────────────────────────────────────

type CellularTarget struct {
	IMSI       string
	SUCI       string // 5G concealed IMSI
	GUTI       string // Globally Unique Temporary Identifier
	TAC        int    // Tracking Area Code
	MCC        int    // Mobile Country Code
	MNC        int    // Mobile Network Code
	CellID     string
	Band       string // 5G, LTE, GSM
	SST        int    // Slice/Service Type
	SD         int    // Slice Differentiator
	SignalDBM  int
	IsRogue    bool
	LastSeen   time.Time
}

type SIMProfile struct {
	IMSI      string
	ICCID     string
	KI        string // Authentication key
	OPC       string // Operator key
	IsESIM    bool
	Carrier   string
	Country   string
}

// ─── API / Backend Data ───────────────────────────────────────────

type APITarget struct {
	BaseURL       string
	Endpoints     []APIEndpoint
	GraphQLSchema string
	SSLInfo       SSLInfo
	Headers       map[string]string
	SDKKeys       []SDKKey // Firebase, OneSignal, etc.
}

type APIEndpoint struct {
	Path        string
	Method      string
	StatusCode  int
	Tech        string // graphql, rest, grpc, websocket
	AuthRequired bool
	Vulnerable  bool
	ParamCount  int
}

type SSLInfo struct {
	Pinned           bool
	Protocols        []string
	CipherSuites     []string
	CertExpiry       time.Time
	SelfSigned       bool
	WeakCipherFound  bool
}

type SDKKey struct {
	Name    string // Firebase, OneSignal, AppsFlyer, Amplitude
	Key     string
	Scope   string // read, write, admin
	Valid   bool
}

// ─── CVE / Exploit Chain Data ───────────────────────────────────────

type CVEEntry struct {
	ID           string
	CVSS         float64
	Platform     string // android, ios
	AffectedVer  string // e.g. "Android 13-14"
	ExploitType  string // zero-click, one-click, local
	Tool         string // metasploit, custom, frida
	Chainable    bool   // can be chained with other CVEs
}

type AttackChain struct {
	Name     string
	Steps    []AttackStep
	CVEs     []CVEEntry
	Confidence float64
}

type AttackStep struct {
	Agent    string
	Action   string
	Tool     string
	Timeout  time.Duration
	Depends  []int // indices of prerequisite steps
}

// ─── Agent Communication ────────────────────────────────────────────

type AgentMessage struct {
	From      string
	To        string // ALL, APEX, GHOST, LANCE, SPECTRE
	Type      string // APK_REVERSE, IPA_DUMP, API_RECON, CELL_SCAN, etc.
	Data      string
	Timestamp time.Time
}

type EventBus struct {
	subscribers map[string]chan AgentMessage
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan AgentMessage),
	}
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

// ─── Shared State ───────────────────────────────────────────────────

type SharedState struct {
	mu           sync.RWMutex
	Targets      []*MobileTarget
	CurrentTarget *MobileTarget
	Devices      []*DeviceProfile
	CellTargets  []*CellularTarget
	SIMProfiles  []*SIMProfile
	APITargets   []*APITarget
	AppInfo      *AppInfo
	CVEList      []CVEEntry
	AttackPlan   *AttackChain
	Results      *AttackResults
}

type AttackResults struct {
	Findings      []string
	Credentials   []Credential
	ExploitsRun   []string
	APIDumps      []string
	CellularDumps []string
}

type Credential struct {
	Type     string // api_key, jwt_token, password, cert
	Value    string
	Source   string // apk, ios_keychain, api_response
	Severity string
}

// ─── Agent Interface ────────────────────────────────────────────────

type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}

// ─── CVE Database (In-Memory) ───────────────────────────────────────

type CVEDatabase struct {
	entries []CVEEntry
	mu      sync.RWMutex
}

func NewCVEDatabase() *CVEDatabase {
	db := &CVEDatabase{}
	db.LoadDefaults()
	return db
}

func (db *CVEDatabase) LoadDefaults() {
	db.entries = []CVEEntry{
		{ID: "CVE-2025-48593", CVSS: 9.8, Platform: "android", AffectedVer: "Android 13-14", ExploitType: "zero-click", Tool: "frida", Chainable: true},
		{ID: "CVE-2024-0044", CVSS: 8.5, Platform: "android", AffectedVer: "Android 12-14", ExploitType: "local", Tool: "adb", Chainable: true},
		{ID: "CVE-2023-XXXXX", CVSS: 7.8, Platform: "android", AffectedVer: "Android 10-13", ExploitType: "one-click", Tool: "drozer", Chainable: true},
		{ID: "CVE-2024-232XX", CVSS: 9.0, Platform: "ios", AffectedVer: "iOS 16-17", ExploitType: "zero-click", Tool: "frida", Chainable: false},
		{ID: "CVE-2023-419XX", CVSS: 8.2, Platform: "ios", AffectedVer: "iOS 15-16", ExploitType: "local", Tool: "libimobiledevice", Chainable: true},
		{ID: "CVE-2024-QUANT", CVSS: 9.5, Platform: "api", AffectedVer: "JWT libs < 4.0", ExploitType: "remote", Tool: "jwt_tool", Chainable: true},
		{ID: "CVE-2024-GRAPH", CVSS: 8.0, Platform: "api", AffectedVer: "GraphQL default", ExploitType: "remote", Tool: "graphqlmap", Chainable: true},
	}
}

func (db *CVEDatabase) Lookup(platform, version string) []CVEEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var results []CVEEntry
	for _, e := range db.entries {
		if e.Platform == platform {
			results = append(results, e)
		}
	}
	return results
}

func (db *CVEDatabase) GetChainable(platform string) []CVEEntry {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var results []CVEEntry
	for _, e := range db.entries {
		if e.Platform == platform && e.Chainable {
			results = append(results, e)
		}
	}
	return results
}
