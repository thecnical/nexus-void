package apibreach

import (
	"fmt"
	"os/exec"
	"strings"
)

type DiscoveryAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewDiscoveryAgent(bus *EventBus, state *SharedState) *DiscoveryAgent {
	return &DiscoveryAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *DiscoveryAgent) Name() string  { return "DISCOVER" }
func (a *DiscoveryAgent) Status() string { return "online" }
func (a *DiscoveryAgent) Start() {
	a.bus.Subscribe("DISCOVER", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *DiscoveryAgent) Stop() { close(a.stopCh) }
func (a *DiscoveryAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "DISCOVER": a.discover(msg.Data)
	}
}
func (a *DiscoveryAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "DISCOVER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *DiscoveryAgent) discover(url string) {
	a.broadcast(fmt.Sprintf("[DISCOVER] Mapping API %s", url))
	// arjun parameter discovery
	if out, err := exec.Command("arjun", "-u", url, "-oJ", "arjun.json").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Parameters found") {
			a.broadcast("[DISCOVER] Hidden parameters found!")
		}
	}
	// ffuf for endpoint enumeration
	if out, err := exec.Command("ffuf", "-u", url+"/api/v1/FUZZ", "-w", "/usr/share/wordlists/api-endpoints.txt").CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "Status: 200") {
				a.broadcast(fmt.Sprintf("[DISCOVER] Valid endpoint: %s", line))
			}
		}
	}
}
