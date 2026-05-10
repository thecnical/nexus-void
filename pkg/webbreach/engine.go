package webbreach

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// WebBreach is the web application attack engine
type WebBreach struct {
	bus    *EventBus
	state  *SharedState
	agents []Agent
	wg     sync.WaitGroup
	stopCh chan struct{}
}

func New() *WebBreach {
	bus := NewEventBus()
	state := &SharedState{}
	return &WebBreach{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
	}
}

func (wb *WebBreach) Start(target string) {
	wb.printBanner()
	fmt.Printf("\033[33m[INIT]\033[0m Target: %s\n", target)

	wb.state.mu.Lock()
	wb.state.Target.URL = target
	wb.state.mu.Unlock()

	wb.agents = []Agent{
		NewCrawlerAgent(wb.bus, wb.state),
		NewXSSAgent(wb.bus, wb.state),
		NewSQLiAgent(wb.bus, wb.state),
		NewIDORAgent(wb.bus, wb.state),
		NewCSRFAgent(wb.bus, wb.state),
		NewWebOmega(wb.bus, wb.state),
	}

	for _, agent := range wb.agents {
		wb.wg.Add(1)
		go func(a Agent) {
			defer wb.wg.Done()
			a.Start()
		}(agent)
	}

	// Wait a moment for all agents to subscribe
	time.Sleep(500 * time.Millisecond)

	// Launch orchestrator
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "OMEGA", Type: "INIT", Data: target})
}

func (wb *WebBreach) Close() {
	close(wb.stopCh)
	for _, agent := range wb.agents {
		agent.Stop()
	}
	wb.wg.Wait()
}

func (wb *WebBreach) Crawl(url string) {
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "CRAWLER", Type: "CRAWL", Data: url})
}
func (wb *WebBreach) XSSScan(url string) {
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "XSS-HUNTER", Type: "SCAN", Data: url})
}
func (wb *WebBreach) SQLiScan(url string) {
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "SQLI-PHANTOM", Type: "SCAN", Data: url})
}
func (wb *WebBreach) IDORScan(url string) {
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "IDOR-BREAKER", Type: "SCAN", Data: url})
}
func (wb *WebBreach) CSRFScan(url string) {
	wb.bus.Broadcast(AgentMessage{From: "USER", To: "CSRF-DEMON", Type: "SCAN", Data: url})
}

func (wb *WebBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[31m[!] %s not found in PATH\033[0m\n", name)
		return false
	}
	return true
}

func (wb *WebBreach) Log(msg string) {
	fmt.Printf("\033[36m[WEBBREACH]\033[0m %s\n", msg)
}

func (wb *WebBreach) printBanner() {
	fmt.Println()
	fmt.Println("\033[35m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[1;35m██╗    ██╗███████╗██████╗ ██████╗ ██████╗ ███████╗ █████╗ \033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[1;35m██║    ██║██╔════╝██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗\033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[1;35m██║ █╗ ██║█████╗  ██████╔╝██████╔╝██████╔╝█████╗  ███████║\033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[1;35m██║███╗██║██╔══╝  ██╔══██╗██╔══██╗██╔══██╗██╔══╝  ██╔══██║\033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[1;35m╚███╔███╔╝███████╗██████╔╝██║  ██║██████╔╝███████╗██║  ██║\033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m║\033[0m   \033[1;35m╚══╝╚══╝ ╚══════╝╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝\033[0m  \033[35m║\033[0m")
	fmt.Println("\033[35m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[35m║\033[0m  \033[33mWeb Application Attack Weapon  —  XSS | SQLi | IDOR | CSRF\033[0m   \033[35m║\033[0m")
	fmt.Println("\033[35m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()
}
