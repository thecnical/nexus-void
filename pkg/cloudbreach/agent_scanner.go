// CLOUD-SCANNER — Multi-Cloud Recon & Misconfig Agent
// scoutSuite, prowler, pacu, steampipe

package cloudbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// CloudScanner handles cloud reconnaissance
type CloudScanner struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewCloudScanner(bus *EventBus, state *SharedState) *CloudScanner {
	return &CloudScanner{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *CloudScanner) Name() string  { return "CLOUD-SCANNER" }
func (a *CloudScanner) Status() string { return "online" }

func (a *CloudScanner) Start() {
	a.bus.Subscribe("CLOUD-SCANNER", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *CloudScanner) Stop() { close(a.stopCh) }

func (a *CloudScanner) Handle(msg AgentMessage) {
	switch msg.Type {
	case "SCAN":
		a.scanCloud(msg.Data)
	}
}

func (a *CloudScanner) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "CLOUD-SCANNER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *CloudScanner) scanCloud(provider string) {
	a.broadcast(fmt.Sprintf("[SCAN] Cloud reconnaissance: %s", provider))

	switch strings.ToLower(provider) {
	case "aws":
		a.scanAWS()
	case "azure":
		a.scanAzure()
	case "gcp":
		a.scanGCP()
	case "all":
		a.scanAWS()
		a.scanAzure()
		a.scanGCP()
	}
}

func (a *CloudScanner) scanAWS() {
	a.broadcast("[SCAN] Scanning AWS with scoutSuite...")
	if out, err := exec.Command("scout", "aws", "--report-dir", "/tmp/scout-aws").CombinedOutput(); err == nil {
		a.broadcast("[SCAN] ScoutSuite AWS scan complete")
		_ = out
	}

	// prowler
	a.broadcast("[SCAN] Scanning AWS with prowler...")
	if out, err := exec.Command("prowler", "aws", "--quick-inventory").CombinedOutput(); err == nil {
		a.broadcast("[SCAN] Prowler AWS scan complete")
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "FAIL") || strings.Contains(line, "HIGH") {
				a.broadcast(fmt.Sprintf("[CRITICAL] %s", line))
			}
		}
		_ = out
	}

	// pacu quick enum
	a.broadcast("[SCAN] Enumerating AWS with pacu...")
	if out, err := exec.Command("pacu", "--module", "iam__enum_users_roles_policies_groups",
		"--json").CombinedOutput(); err == nil {
		a.broadcast("[SCAN] Pacu IAM enumeration complete")
		_ = out
	}
}

func (a *CloudScanner) scanAzure() {
	a.broadcast("[SCAN] Scanning Azure with scoutSuite...")
	if out, err := exec.Command("scout", "azure", "--report-dir", "/tmp/scout-azure").CombinedOutput(); err == nil {
		a.broadcast("[SCAN] ScoutSuite Azure scan complete")
		_ = out
	}
}

func (a *CloudScanner) scanGCP() {
	a.broadcast("[SCAN] Scanning GCP with scoutSuite...")
	if out, err := exec.Command("scout", "gcp", "--report-dir", "/tmp/scout-gcp").CombinedOutput(); err == nil {
		a.broadcast("[SCAN] ScoutSuite GCP scan complete")
		_ = out
	}
}
