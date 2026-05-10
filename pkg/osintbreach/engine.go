// OSINTBREACH Engine
// Coordinates 6 agents for full-spectrum autonomous reconnaissance

package osintbreach

import (
	"fmt"
	"os/exec"
)

// OsintBreach is the main engine
type OsintBreach struct {
	Bus      *EventBus
	State    *SharedState
	Agents   []Agent
	StopCh   chan struct{}
}

func New() *OsintBreach {
	bus := NewEventBus()
	state := &SharedState{
		Subdomains: make(map[string]*Subdomain),
		People:     make(map[string]*Person),
	}

	ob := &OsintBreach{
		Bus:    bus,
		State:  state,
		StopCh: make(chan struct{}),
	}

	ob.Agents = []Agent{
		NewReconAgent(bus, state),
		NewSurfaceAgent(bus, state),
		NewPersonaAgent(bus, state),
		NewVulnAgent(bus, state),
		NewSupplyAgent(bus, state),
		NewOsintOmega(bus, state),
	}

	return ob
}

func (ob *OsintBreach) Start(domain string) {
	fmt.Println()
	fmt.Println("\033[36m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m║\033[35m         O S I N T B R E A C H   E N G I N E                  \033[36m║\033[0m")
	fmt.Println("\033[36m║\033[32m    AUTONOMOUS RECONNAISSANCE & ATTACK SURFACE WEAPON        \033[36m║\033[0m")
	fmt.Println("\033[36m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[36m║\033[33m  6 Agents | Full Spectrum | AI-Driven | Self-Learning        \033[36m║\033[0m")
	fmt.Println("\033[36m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	for _, agent := range ob.Agents {
		go agent.Start()
		fmt.Printf("[OMEGA] Agent %s deployed\n", agent.Name())
	}

	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "INIT",
		Data: domain,
	})

	<-ob.StopCh
	fmt.Println("[OMEGA] OSINTBREACH shutdown.")
}

func (ob *OsintBreach) Close() {
	for _, agent := range ob.Agents {
		agent.Stop()
	}
	select {
	case <-ob.StopCh:
	default:
		close(ob.StopCh)
	}
}

func (ob *OsintBreach) Recon(domain string) {
	ob.State.mu.Lock()
	ob.State.Target = &Target{Domain: domain}
	ob.State.mu.Unlock()

	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALPHA",
		Type: "DOMAIN_RECON",
		Data: domain,
	})
}

func (ob *OsintBreach) SurfaceScan(domain string) {
	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "BETA",
		Type: "SURFACE_MAP",
		Data: domain,
	})
}

func (ob *OsintBreach) PersonaHunt(domain string) {
	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "GAMMA",
		Type: "PERSONA_HUNT",
		Data: domain,
	})
}

func (ob *OsintBreach) VulnScan(target string) {
	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "DELTA",
		Type: "VULN_SCAN",
		Data: target,
	})
}

func (ob *OsintBreach) SupplyCheck(repo string) {
	ob.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "EPSILON",
		Type: "SUPPLY_CHECK",
		Data: repo,
	})
}

func (ob *OsintBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

func (ob *OsintBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	ob.Bus.Broadcast(AgentMessage{From: source, To: "ALL", Type: "LOG", Data: msg})
}

func (ob *OsintBreach) AddVuln(v Vulnerability) {
	ob.State.mu.Lock()
	ob.State.Vulns = append(ob.State.Vulns, v)
	ob.State.mu.Unlock()
	ob.Log("OMEGA", fmt.Sprintf("VULN [%s] %s on %s", v.Severity, v.Name, v.URL))
}

func (ob *OsintBreach) AddSecret(s Secret) {
	ob.State.mu.Lock()
	ob.State.Secrets = append(ob.State.Secrets, s)
	ob.State.mu.Unlock()
	ob.Log("OMEGA", fmt.Sprintf("SECRET [%s] found at %s", s.Type, s.URL))
}

func (ob *OsintBreach) GetSurfaceReport() *AttackSurface {
	ob.State.mu.RLock()
	defer ob.State.mu.RUnlock()

	as := &AttackSurface{
		Domain: ob.State.Target.Domain,
	}
	for _, sub := range ob.State.Subdomains {
		if sub.Live {
			as.LiveHosts++
		}
		if sub.Takeover {
			as.Score += 20
		}
	}
	as.ExposedAPIs = len(ob.State.APIs)
	as.LeakedCreds = len(ob.State.People)
	as.VulnCount = len(ob.State.Vulns)
	as.CloudExposures = len(ob.State.CloudAssets)
	as.SupplyChainRisk = len(ob.State.Deps)

	as.Score += float64(as.LiveHosts) * 2
	as.Score += float64(as.ExposedAPIs) * 5
	as.Score += float64(as.LeakedCreds) * 10
	as.Score += float64(as.VulnCount) * 15
	as.Score += float64(as.CloudExposures) * 8

	if as.Score >= 100 {
		as.Priority = "critical"
	} else if as.Score >= 50 {
		as.Priority = "high"
	} else {
		as.Priority = "medium"
	}

	return as
}
