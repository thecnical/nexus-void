// CLOUDBREACH — Multi-Cloud Exploitation Weapon
// 6 AI Agents: CLOUD-SCANNER, IAM-ESCALATOR, BUCKET-RAIDER, CONTAINER-BREAKER, LAMBDA-PHANTOM, OMEGA-BRAIN

package cloudbreach

import (
	"sync"
	"time"
)

// ─── Cloud Target Types ───────────────────────────────────────────

type Target struct {
	ID        string
	Provider  string // aws, azure, gcp, all
	AccountID string
	Regions   []string
	Buckets   []Bucket
	IAMPols   []IAMPolicy
	Containers []Container
	Lambdas   []LambdaFunc
	Score     float64
}

type Bucket struct {
	Name     string
	Provider string
	Region   string
	Public   bool
	Writable bool
	Listable bool
	Objects  []string
}

type IAMPolicy struct {
	Name       string
	Arn        string
	Actions    []string
	Resources  []string
	Privesc    bool
	OverlyPermissive bool
}

type Container struct {
	Name      string
	Cluster   string
	Namespace string
	EscapePossible bool
	Privileged bool
	Mounts    []string
}

type LambdaFunc struct {
	Name       string
	Runtime    string
	Handler    string
	Role       string
	Backdoored bool
	EnvVars    map[string]string
}

type CloudCred struct {
	Provider   string
	AccessKey  string
	SecretKey  string
	Token      string
	Valid      bool
	Perms      []string
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
	Buckets     []Bucket
	IAMPolicies []IAMPolicy
	Containers  []Container
	Lambdas     []LambdaFunc
	Creds       []CloudCred
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
