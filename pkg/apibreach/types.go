package apibreach

import "sync"

// APITarget represents an API endpoint
type APITarget struct {
	BaseURL    string            `json:"base_url"`
	Endpoints  []APIEndpoint     `json:"endpoints"`
	AuthType   string            `json:"auth_type"`
	Headers    map[string]string `json:"headers"`
	Schemas    []string          `json:"schemas"` // OpenAPI, GraphQL, gRPC
}

type APIEndpoint struct {
	Path       string            `json:"path"`
	Method     string            `json:"method"`
	Params     []APIParam        `json:"params"`
	Auth       bool              `json:"auth"`
	Response   string            `json:"response"`
}

type APIParam struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	In     string `json:"in"` // query, path, body, header
	Required bool `json:"required"`
}

type APIVuln struct {
	Type       string  `json:"type"`
	Endpoint   string  `json:"endpoint"`
	Param      string  `json:"param"`
	Payload    string  `json:"payload"`
	Confidence float64 `json:"confidence"`
}

type AgentMessage struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Data string `json:"data"`
}

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
		select { case ch <- msg: default: }
	}
}

type SharedState struct {
	Target         APITarget `json:"target"`
	Vulnerabilities []APIVuln `json:"vulnerabilities"`
	mu              sync.RWMutex
}

type Agent interface {
	Start()
	Stop()
	Handle(msg AgentMessage)
	Name() string
	Status() string
}
