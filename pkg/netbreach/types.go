// NETBREACH — Network & Post-Exploitation Weapon
// 6 AI Agents: INFECT, PIVOT, EXTRACT, AD-PHANTOM, C2-CONTROL, OMEGA-BRAIN

package netbreach

import (
	"sync"
	"time"
)

// ─── Target & Network Types ─────────────────────────────────────────

type Target struct {
	ID        string
	IP        string
	Hostname  string
	OS        string
	Domain    string
	Sessions  []Session
	Creds     []Credential
	ADObjects []ADObject
	Tunnels   []Tunnel
	Score     float64
}

type Session struct {
	ID       string
	Protocol string // smb, ssh, winrm, wmi, rdp, mssql
	User     string
	Hash     string
	Token    string
	Active   bool
}

type Credential struct {
	Username string
	Password string
	Hash     string
	Type     string // plaintext, ntlm, kerberos, ticket
	Source   string
}

type ADObject struct {
	Name       string
	Type       string // user, group, computer, gpo, ou
	SID        string
	DistinguishedName string
	Members    []string
	Privileges []string
	Vulnerable bool
}

type Tunnel struct {
	ID       string
	Type     string // chisel, ligolo, ssh, dnscat2
	Local    string
	Remote   string
	Active   bool
}

type PrivilegePath struct {
	Steps    []string
	Target   string
	Method   string
	Exploitable bool
}

// ─── Agent Communication ────────────────────────────────────────────

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

// ─── Shared State ──────────────────────────────────────────────────

type SharedState struct {
	mu          sync.RWMutex
	Target      *Target
	Sessions    map[string]*Session
	Creds       []Credential
	ADObjects   []ADObject
	Tunnels     []Tunnel
	Paths       []PrivilegePath
	Finished    bool
}

// ─── Agent Interface ──────────────────────────────────────────────

type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}
