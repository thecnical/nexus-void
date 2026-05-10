package apibreach

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

type APIBreach struct {
	bus    *EventBus
	state  *SharedState
	agents []Agent
	wg     sync.WaitGroup
	stopCh chan struct{}
}

func New() *APIBreach {
	bus := NewEventBus()
	state := &SharedState{}
	return &APIBreach{bus: bus, state: state, stopCh: make(chan struct{})}
}

func (ab *APIBreach) Start(target string) {
	ab.printBanner()
	fmt.Printf("\033[33m[INIT]\033[0m Target API: %s\n", target)
	ab.state.mu.Lock()
	ab.state.Target.BaseURL = target
	ab.state.mu.Unlock()

	ab.agents = []Agent{
		NewDiscoveryAgent(ab.bus, ab.state),
		NewAuthBypassAgent(ab.bus, ab.state),
		NewRateLimitAgent(ab.bus, ab.state),
		NewGraphQLAgent(ab.bus, ab.state),
		NewGRPCAgent(ab.bus, ab.state),
		NewAPIOmega(ab.bus, ab.state),
	}
	for _, agent := range ab.agents {
		ab.wg.Add(1)
		go func(a Agent) { defer ab.wg.Done(); a.Start() }(agent)
	}
	time.Sleep(500 * time.Millisecond)
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "OMEGA", Type: "INIT", Data: target})
}

func (ab *APIBreach) Close() {
	close(ab.stopCh)
	for _, a := range ab.agents {
		a.Stop()
	}
	ab.wg.Wait()
}

func (ab *APIBreach) Discover(url string) {
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "DISCOVER", Type: "DISCOVER", Data: url})
}
func (ab *APIBreach) AuthBypass(url string) {
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "AUTH-BYPASS", Type: "TEST", Data: url})
}
func (ab *APIBreach) RateLimit(url string) {
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "RATE-LIMIT", Type: "TEST", Data: url})
}
func (ab *APIBreach) GraphQL(url string) {
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "GRAPHQL-PHANTOM", Type: "TEST", Data: url})
}
func (ab *APIBreach) GRPC(url string) {
	ab.bus.Broadcast(AgentMessage{From: "USER", To: "GRPC-BREAKER", Type: "TEST", Data: url})
}

func (ab *APIBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[31m[!] %s not found\033[0m\n", name)
		return false
	}
	return true
}

func (ab *APIBreach) printBanner() {
	fmt.Println()
	fmt.Println("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m █████╗ ██████╗ ██╗    ██████╗ ██████╗ ███████╗ █████╗   \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m██╔══██╗██╔══██╗██║    ██╔══██╗██╔══██╗██╔════╝██╔══██╗  \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m███████║██████╔╝██║    ██████╔╝██████╔╝█████╗  ███████║  \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m██╔══██║██╔═══╝ ██║    ██╔══██╗██╔══██╗██╔══╝  ██╔══██║  \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m██║  ██║██║     ██║    ██████╔╝██║  ██║███████╗██║  ██║  \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[1;36m╚═╝  ╚═╝╚═╝     ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝  \033[0m  \033[36m║\033[0m")
	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[36m║\033[0m  \033[33mAPI Attack Weapon — REST | GraphQL | gRPC | Auth Bypass\033[0m   \033[36m║\033[0m")
	fmt.Println("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()
}
