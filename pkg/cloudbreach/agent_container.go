// CONTAINER-BREAKER — Container/K8s Escape Agent
// peirates, kube-hunter, amicontained, cdk

package cloudbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// ContainerBreaker handles container escape
type ContainerBreaker struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewContainerBreaker(bus *EventBus, state *SharedState) *ContainerBreaker {
	return &ContainerBreaker{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *ContainerBreaker) Name() string  { return "CONTAINER-BREAKER" }
func (a *ContainerBreaker) Status() string { return "online" }

func (a *ContainerBreaker) Start() {
	a.bus.Subscribe("CONTAINER-BREAKER", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *ContainerBreaker) Stop() { close(a.stopCh) }

func (a *ContainerBreaker) Handle(msg AgentMessage) {
	switch msg.Type {
	case "BREAK":
		a.breakContainer(msg.Data)
	}
}

func (a *ContainerBreaker) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "CONTAINER-BREAKER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *ContainerBreaker) breakContainer(cluster string) {
	a.broadcast(fmt.Sprintf("[CONTAINER] Container escape analysis: %s", cluster))

	// amicontained - check container capabilities
	a.broadcast("[CONTAINER] Checking container capabilities...")
	if out, err := exec.Command("amicontained").CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "CAP_SYS_ADMIN") || strings.Contains(line, "privileged") {
				a.broadcast(fmt.Sprintf("[CRITICAL] %s", line))
			}
		}
		_ = output
	}

	// kube-hunter for Kubernetes
	a.broadcast("[CONTAINER] Scanning Kubernetes with kube-hunter...")
	if out, err := exec.Command("kube-hunter", "--remote", cluster, "--active").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "Vulnerability") {
			a.broadcast("[CRITICAL] K8s vulnerabilities found!")
		}
		_ = output
	}

	// peirates for K8s privilege escalation
	a.broadcast("[CONTAINER] K8s privilege escalation with peirates...")
	if out, err := exec.Command("peirates", "--enumerate").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "token") || strings.Contains(output, "secret") {
			a.broadcast("[CRITICAL] K8s service account tokens accessible!")
		}
		_ = output
	}

	// CDK - Container Drift K8s
	a.broadcast("[CONTAINER] Container escape with CDK...")
	if out, err := exec.Command("cdk", "evaluate", "--full").CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "exploit") || strings.Contains(output, "escape") {
			a.broadcast("[CRITICAL] Container escape vectors found!")
		}
		_ = output
	}
}
