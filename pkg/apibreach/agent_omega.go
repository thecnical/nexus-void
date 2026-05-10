package apibreach

import (
	"fmt"
	"time"
)

type APIOmega struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewAPIOmega(bus *EventBus, state *SharedState) *APIOmega {
	return &APIOmega{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 200)}
}
func (o *APIOmega) Name() string  { return "OMEGA-BRAIN" }
func (o *APIOmega) Status() string { return "online" }
func (o *APIOmega) Start() {
	o.bus.Subscribe("OMEGA", o.msgCh)
	for { select { case msg := <-o.msgCh: o.Handle(msg); case <-o.stopCh: return } }
}
func (o *APIOmega) Stop() { close(o.stopCh) }
func (o *APIOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT": o.orchestrate(msg.Data)
	}
}

func (o *APIOmega) orchestrate(url string) {
	fmt.Println("\033[34m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Println("\033[34m  OMEGA-BRAIN: API EXPLOITATION CHAIN INITIATED              \033[0m")
	fmt.Println("\033[34m═══════════════════════════════════════════════════════════════\033[0m")

	fmt.Println("\033[35m[PHASE 1/6]\033[0m DISCOVER: Endpoint enumeration")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "DISCOVER", Type: "DISCOVER", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 2/6]\033[0m AUTH-BYPASS: Token/JWT weakness")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "AUTH-BYPASS", Type: "TEST", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 3/6]\033[0m RATE-LIMIT: Throttle bypass")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "RATE-LIMIT", Type: "TEST", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 4/6]\033[0m GRAPHQL-PHANTOM: Query injection")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GRAPHQL-PHANTOM", Type: "TEST", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 5/6]\033[0m GRPC-BREAKER: Proto introspection")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "GRPC-BREAKER", Type: "TEST", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 6/6]\033[0m OMEGA: Report generation")
	o.generateReport()
	fmt.Println("\033[32m[OMEGA] API exploitation chain complete.\033[0m")
}

func (o *APIOmega) generateReport() {
	o.state.mu.RLock()
	defer o.state.mu.RUnlock()
	fmt.Println("\033[34m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[34m║\033[33m      A P I B R E A C H   R E P O R T                         \033[34m║\033[0m")
	fmt.Println("\033[34m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Printf("\033[34m║\033[0m  API Vulnerabilities: \033[33m%-32d\033[34m║\033[0m\n", len(o.state.Vulnerabilities))
	for _, v := range o.state.Vulnerabilities {
		fmt.Printf("\033[34m║\033[0m    - %s on %s\n", v.Type, v.Endpoint)
	}
	fmt.Println("\033[34m╚═══════════════════════════════════════════════════════════════╝\033[0m")
}
