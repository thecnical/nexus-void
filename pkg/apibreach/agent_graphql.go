package apibreach

import (
	"fmt"
	"os/exec"
	"strings"
)

type GraphQLAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewGraphQLAgent(bus *EventBus, state *SharedState) *GraphQLAgent {
	return &GraphQLAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *GraphQLAgent) Name() string  { return "GRAPHQL-PHANTOM" }
func (a *GraphQLAgent) Status() string { return "online" }
func (a *GraphQLAgent) Start() {
	a.bus.Subscribe("GRAPHQL-PHANTOM", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *GraphQLAgent) Stop() { close(a.stopCh) }
func (a *GraphQLAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "TEST": a.testGraphQL(msg.Data)
	}
}
func (a *GraphQLAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "GRAPHQL-PHANTOM", To: "ALL", Type: "LOG", Data: msg})
}

func (a *GraphQLAgent) testGraphQL(url string) {
	a.broadcast(fmt.Sprintf("[GRAPHQL] Testing %s", url))
	if out, err := exec.Command("graphqlmap", "-u", url).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "introspection") || strings.Contains(output, "vulnerable") {
			a.broadcast("[CRITICAL] GraphQL introspection enabled!")
			a.state.mu.Lock()
			a.state.Vulnerabilities = append(a.state.Vulnerabilities, APIVuln{Type: "GraphQL", Endpoint: url, Confidence: 0.95})
			a.state.mu.Unlock()
		}
	}
}
