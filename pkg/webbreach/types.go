package webbreach

import (
	"sync"
)

// WebTarget represents a web application target
type WebTarget struct {
	URL        string            `json:"url"`
	TechStack  []string          `json:"tech_stack"`
	Headers    map[string]string `json:"headers"`
	Forms      []Form            `json:"forms"`
	Endpoints  []Endpoint        `json:"endpoints"`
	Parameters []Parameter       `json:"parameters"`
}

type Form struct {
	Action string            `json:"action"`
	Method string            `json:"method"`
	Fields map[string]string `json:"fields"`
}

type Endpoint struct {
	Path   string   `json:"path"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

type Parameter struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source string `json:"source"` // query, body, header, cookie
}

type Vulnerability struct {
	Type       string  `json:"type"`
	URL        string  `json:"url"`
	Param      string  `json:"param"`
	Payload    string  `json:"payload"`
	Confidence float64 `json:"confidence"`
	Evidence   string  `json:"evidence"`
}

// AgentMessage is the event bus message format
type AgentMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Data string `json:"data"`
}

// EventBus for agent communication
type EventBus struct {
	subscribers map[string]chan AgentMessage
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: make(map[string]chan AgentMessage)}
}

func (eb *EventBus) Subscribe(name string, ch chan AgentMessage) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.subscribers[name] = ch
}

func (eb *EventBus) Broadcast(msg AgentMessage) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, ch := range eb.subscribers {
		select {
		case ch <- msg:
		default:
		}
	}
}

// SharedState holds all discovered data
type SharedState struct {
	Target          WebTarget       `json:"target"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	CrawledURLs     []string        `json:"crawled_urls"`
	mu              sync.RWMutex
}

// Agent interface for all web agents
type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}
