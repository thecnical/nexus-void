package apibreach

import (
	"fmt"
	"os/exec"
)

type RateLimitAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewRateLimitAgent(bus *EventBus, state *SharedState) *RateLimitAgent {
	return &RateLimitAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *RateLimitAgent) Name() string  { return "RATE-LIMIT" }
func (a *RateLimitAgent) Status() string { return "online" }
func (a *RateLimitAgent) Start() {
	a.bus.Subscribe("RATE-LIMIT", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *RateLimitAgent) Stop() { close(a.stopCh) }
func (a *RateLimitAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "TEST": a.testRate(msg.Data)
	}
}
func (a *RateLimitAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "RATE-LIMIT", To: "ALL", Type: "LOG", Data: msg})
}

func (a *RateLimitAgent) testRate(url string) {
	a.broadcast(fmt.Sprintf("[RATE] Stressing %s", url))
	// Use curl rapid-fire
	if out, err := exec.Command("bash", "-c", "for i in {1..50}; do curl -s -o /dev/null -w '%{http_code}' "+url+"; done").CombinedOutput(); err == nil {
		output := string(out)
		if len(output) > 0 {
			a.broadcast(fmt.Sprintf("[RATE] Response codes: %s", output))
		}
	}
}
