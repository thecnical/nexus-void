// CRYPTOBREACH — Cryptography Attack Weapon
// 6 AI Agents: HASH-BREAKER, CERT-HUNTER, TLS-PHANTOM, KEY-EXTRACT, QUANTUM-SHADOW, OMEGA-BRAIN

package cryptobreach

import (
	"sync"
	"time"
)

// ─── Crypto Target Types ────────────────────────────────────────────

type Target struct {
	ID          string
	Hash        string
	HashType    string
	URL         string
	Certs       []Certificate
	TLSConfig   *TLSConfig
	Keys        []KeyMaterial
	Score       float64
}

type Certificate struct {
	Subject    string
	Issuer     string
	NotBefore  time.Time
	NotAfter   time.Time
	SANs       []string
	Weak       bool
	SelfSigned bool
	Expired    bool
}

type TLSConfig struct {
	Version      string
	CipherSuites []string
	WeakCiphers  []string
	Downgradable bool
	HSTS         bool
}

type KeyMaterial struct {
	Type       string // rsa, ecc, dsa, dh
	Size       int
	Weak       bool
	Factorable bool
	Extracted  bool
}

type HashCrackResult struct {
	Hash       string
	Plaintext  string
	Type       string
	Time       time.Duration
	Method     string
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
	mu            sync.RWMutex
	Target        *Target
	CrackResults  []HashCrackResult
	Certs         []Certificate
	TLSIssues     []string
	Keys          []KeyMaterial
	QuantumVulns  []string
	Finished      bool
}

// ─── Agent Interface ──────────────────────────────────────────────

type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}
