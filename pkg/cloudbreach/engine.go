// CLOUDBREACH Engine
// Coordinates 6 agents for multi-cloud exploitation

package cloudbreach

import (
	"fmt"
	"os/exec"
)

// CloudBreach is the main engine
type CloudBreach struct {
	Bus    *EventBus
	State  *SharedState
	Agents []Agent
	StopCh chan struct{}
}

func New() *CloudBreach {
	bus := NewEventBus()
	state := &SharedState{}

	clb := &CloudBreach{
		Bus:    bus,
		State:  state,
		StopCh: make(chan struct{}),
	}

	clb.Agents = []Agent{
		NewCloudScanner(bus, state),
		NewIAMEscalator(bus, state),
		NewBucketRaider(bus, state),
		NewContainerBreaker(bus, state),
		NewLambdaPhantom(bus, state),
		NewCloudOmega(bus, state),
	}

	return clb
}

func (clb *CloudBreach) Start(target string) {
	fmt.Println()
	fmt.Println("\033[34m╔═══════════════════════════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[34m║\033[33m      C L O U D B R E A C H   E N G I N E                      \033[34m║\033[0m")
	fmt.Println("\033[34m║\033[32m    MULTI-CLOUD EXPLOITATION WEAPON SYSTEM                    \033[34m║\033[0m")
	fmt.Println("\033[34m╠═══════════════════════════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[34m║\033[36m  6 Agents | AWS/Azure/GCP | IAM | Buckets | K8s | Lambda    \033[34m║\033[0m")
	fmt.Println("\033[34m╚═══════════════════════════════════════════════════════════════╝\033[0m")
	fmt.Println()

	for _, agent := range clb.Agents {
		go agent.Start()
		fmt.Printf("[OMEGA] Agent %s deployed\n", agent.Name())
	}

	clb.Bus.Broadcast(AgentMessage{
		From: "OMEGA",
		To:   "ALL",
		Type: "INIT",
		Data: target,
	})

	<-clb.StopCh
	fmt.Println("[OMEGA] CLOUDBREACH shutdown.")
}

func (clb *CloudBreach) Close() {
	for _, agent := range clb.Agents {
		agent.Stop()
	}
	select {
	case <-clb.StopCh:
	default:
		close(clb.StopCh)
	}
}

func (clb *CloudBreach) Scan(provider string) {
	clb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "CLOUD-SCANNER", Type: "SCAN", Data: provider})
}

func (clb *CloudBreach) Escalate(provider string) {
	clb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "IAM-ESCALATOR", Type: "ESCALATE", Data: provider})
}

func (clb *CloudBreach) RaidBuckets(provider string) {
	clb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "BUCKET-RAIDER", Type: "RAID", Data: provider})
}

func (clb *CloudBreach) BreakContainers(cluster string) {
	clb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "CONTAINER-BREAKER", Type: "BREAK", Data: cluster})
}

func (clb *CloudBreach) BackdoorLambda(funcName string) {
	clb.Bus.Broadcast(AgentMessage{From: "OMEGA", To: "LAMBDA-PHANTOM", Type: "BACKDOOR", Data: funcName})
}

func (clb *CloudBreach) EnsureTool(name string) bool {
	if _, err := exec.LookPath(name); err != nil {
		fmt.Printf("\033[33m[!] Missing: %s. Install: nexus-void arsenal install %s\033[0m\n", name, name)
		return false
	}
	return true
}

func (clb *CloudBreach) Log(source, msg string) {
	fmt.Printf("[%s] %s\n", source, msg)
	clb.Bus.Broadcast(AgentMessage{From: source, To: "ALL", Type: "LOG", Data: msg})
}
