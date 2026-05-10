package webbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// CrawlerAgent discovers endpoints and tech stack
type CrawlerAgent struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewCrawlerAgent(bus *EventBus, state *SharedState) *CrawlerAgent {
	return &CrawlerAgent{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}
func (a *CrawlerAgent) Name() string  { return "CRAWLER" }
func (a *CrawlerAgent) Status() string { return "online" }
func (a *CrawlerAgent) Start() {
	a.bus.Subscribe("CRAWLER", a.msgCh)
	for { select { case msg := <-a.msgCh: a.Handle(msg); case <-a.stopCh: return } }
}
func (a *CrawlerAgent) Stop() { close(a.stopCh) }
func (a *CrawlerAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "CRAWL": a.crawl(msg.Data)
	}
}
func (a *CrawlerAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "CRAWLER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *CrawlerAgent) crawl(url string) {
	a.broadcast(fmt.Sprintf("[CRAWL] Mapping %s", url))

	// katana fast crawl
	if out, err := exec.Command("katana", "-u", url, "-d", "5", "-j").CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "http") {
				a.state.mu.Lock()
				a.state.CrawledURLs = append(a.state.CrawledURLs, line)
				a.state.mu.Unlock()
			}
		}
		a.broadcast(fmt.Sprintf("[CRAWL] Katana found %d URLs", len(a.state.CrawledURLs)))
	}

	// whatweb fingerprinting
	if out, err := exec.Command("whatweb", "-a", "3", url).CombinedOutput(); err == nil {
		a.broadcast(fmt.Sprintf("[CRAWL] Tech: %s", string(out)))
	}

	// httpx probe
	if out, err := exec.Command("httpx", "-u", url, "-tech-detect").CombinedOutput(); err == nil {
		a.broadcast(fmt.Sprintf("[CRAWL] httpx: %s", string(out)))
	}
}
