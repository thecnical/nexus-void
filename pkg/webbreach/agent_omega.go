package webbreach

import (
	"fmt"
	"time"
)

// WebOmega coordinates web agents
type WebOmega struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewWebOmega(bus *EventBus, state *SharedState) *WebOmega {
	return &WebOmega{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 200)}
}
func (o *WebOmega) Name() string  { return "OMEGA-BRAIN" }
func (o *WebOmega) Status() string { return "online" }
func (o *WebOmega) Start() {
	o.bus.Subscribe("OMEGA", o.msgCh)
	for { select { case msg := <-o.msgCh: o.Handle(msg); case <-o.stopCh: return } }
}
func (o *WebOmega) Stop() { close(o.stopCh) }
func (o *WebOmega) Handle(msg AgentMessage) {
	switch msg.Type {
	case "INIT": o.orchestrate(msg.Data)
	}
}

func (o *WebOmega) orchestrate(url string) {
	fmt.Println("\033[34m═══════════════════════════════════════════════════════════════\033[0m")
	fmt.Println("\033[34m  OMEGA-BRAIN: WEB EXPLOITATION CHAIN INITIATED              \033[0m")
	fmt.Println("\033[34m═══════════════════════════════════════════════════════════════\033[0m")

	fmt.Println("\033[35m[PHASE 1/6]\033[0m CRAWLER: Endpoint discovery")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "CRAWLER", Type: "CRAWL", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 2/6]\033[0m XSS-HUNTER: Reflected/DOM XSS")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "XSS-HUNTER", Type: "SCAN", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 3/6]\033[0m SQLI-PHANTOM: Injection testing")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "SQLI-PHANTOM", Type: "SCAN", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 4/6]\033[0m IDOR-BREAKER: Access control")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "IDOR-BREAKER", Type: "SCAN", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 5/6]\033[0m CSRF-DEMON: Token bypass")
	o.bus.Broadcast(AgentMessage{From: "OMEGA", To: "CSRF-DEMON", Type: "SCAN", Data: url})
	time.Sleep(2 * time.Second)

	fmt.Println("\033[35m[PHASE 6/6]\033[0m OMEGA: Report generation")
	o.generateReport()
	fmt.Println("\033[32m[OMEGA] Web exploitation chain complete.\033[0m")
}

func (o *WebOmega) generateReport() {
	o.state.mu.RLock()
	defer o.state.mu.RUnlock()
	fmt.Println("\033[34m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[34m║\033[33m      W E B B R E A C H   R E P O R T                         \033[34m║\033[0m")
	fmt.Println("\033[34m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Printf("\033[34m║\033[0m  URLs Crawled:      \033[33m%-34d\033[34m║\033[0m\n", len(o.state.CrawledURLs))
	fmt.Printf("\033[34m║\033[0m  Vulnerabilities:   \033[33m%-34d\033[34m║\033[0m\n", len(o.state.Vulnerabilities))
	for _, v := range o.state.Vulnerabilities {
		fmt.Printf("\033[34m║\033[0m    - %s on %s\n", v.Type, v.URL)
	}
	fmt.Println("\033[34m╚═══════════════════════════════════════════════════════════════╝\033[0m")
}
