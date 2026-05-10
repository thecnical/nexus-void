// BUCKET-RAIDER — S3/Azure Blob/GCP Bucket Exploit Agent
// s3scanner, gcpbucketbrute, cloud_enum

package cloudbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// BucketRaider handles cloud storage exploitation
type BucketRaider struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewBucketRaider(bus *EventBus, state *SharedState) *BucketRaider {
	return &BucketRaider{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *BucketRaider) Name() string  { return "BUCKET-RAIDER" }
func (a *BucketRaider) Status() string { return "online" }

func (a *BucketRaider) Start() {
	a.bus.Subscribe("BUCKET-RAIDER", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *BucketRaider) Stop() { close(a.stopCh) }

func (a *BucketRaider) Handle(msg AgentMessage) {
	switch msg.Type {
	case "RAID":
		a.raidBuckets(msg.Data)
	}
}

func (a *BucketRaider) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "BUCKET-RAIDER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *BucketRaider) raidBuckets(provider string) {
	a.broadcast(fmt.Sprintf("[BUCKET] Raiding cloud storage: %s", provider))

	if strings.ToLower(provider) == "aws" || provider == "all" {
		// s3scanner
		a.broadcast("[BUCKET] Scanning S3 buckets with s3scanner...")
		if out, err := exec.Command("s3scanner", "scan", "--buckets-file", "buckets.txt").CombinedOutput(); err == nil {
			output := string(out)
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "public") || strings.Contains(line, "Found") {
					a.broadcast(fmt.Sprintf("[CRITICAL] %s", line))
					bucket := Bucket{Name: line, Provider: "aws", Public: true}
					a.state.mu.Lock()
					a.state.Buckets = append(a.state.Buckets, bucket)
					a.state.mu.Unlock()
				}
			}
			_ = output
		}

		// aws s3 ls
		if out, err := exec.Command("aws", "s3", "ls").CombinedOutput(); err == nil {
			output := string(out)
			a.broadcast(fmt.Sprintf("[BUCKET] AWS S3 buckets found: %d", len(strings.Split(output, "\n"))))
			_ = output
		}
	}

	if strings.ToLower(provider) == "gcp" || provider == "all" {
		// gcpbucketbrute
		if out, err := exec.Command("python3", "gcpbucketbrute.py", "-k", "target").CombinedOutput(); err == nil {
			output := string(out)
			if strings.Contains(output, "FOUND") {
				a.broadcast("[CRITICAL] Exposed GCP buckets found!")
			}
			_ = output
		}
	}

	if strings.ToLower(provider) == "azure" || provider == "all" {
		// cloud_enum for Azure blobs
		if out, err := exec.Command("python3", "cloud_enum.py", "-k", "target", "-m", "azure").CombinedOutput(); err == nil {
			output := string(out)
			if strings.Contains(output, "FOUND") {
				a.broadcast("[CRITICAL] Exposed Azure blobs found!")
			}
			_ = output
		}
	}
}
