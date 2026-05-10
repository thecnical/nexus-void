// HASH-BREAKER — Password Hash Cracking Agent
// hashcat, john, hashid

package cryptobreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// HashBreaker handles hash cracking
type HashBreaker struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewHashBreaker(bus *EventBus, state *SharedState) *HashBreaker {
	return &HashBreaker{bus: bus, state: state, stopCh: make(chan struct{}), msgCh: make(chan AgentMessage, 100)}
}

func (a *HashBreaker) Name() string  { return "HASH-BREAKER" }
func (a *HashBreaker) Status() string { return "online" }

func (a *HashBreaker) Start() {
	a.bus.Subscribe("HASH-BREAKER", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *HashBreaker) Stop() { close(a.stopCh) }

func (a *HashBreaker) Handle(msg AgentMessage) {
	switch msg.Type {
	case "CRACK":
		a.crackHash(msg.Data)
	case "IDENTIFY":
		a.identifyHash(msg.Data)
	}
}

func (a *HashBreaker) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "HASH-BREAKER", To: "ALL", Type: "LOG", Data: msg})
}

func (a *HashBreaker) crackHash(hash string) {
	a.broadcast(fmt.Sprintf("[HASH] Cracking: %s", hash))

	// Identify hash type first
	hashType := a.identifyHash(hash)
	a.broadcast(fmt.Sprintf("[HASH] Detected type: %s", hashType))

	// Write hash to temp file
	// hashcat cracking
	var cmd *exec.Cmd
	switch hashType {
	case "NTLM":
		cmd = exec.Command("hashcat", "-m", "1000", "-a", "0", hash, "/usr/share/wordlists/rockyou.txt")
	case "MD5":
		cmd = exec.Command("hashcat", "-m", "0", "-a", "0", hash, "/usr/share/wordlists/rockyou.txt")
	case "SHA256":
		cmd = exec.Command("hashcat", "-m", "1400", "-a", "0", hash, "/usr/share/wordlists/rockyou.txt")
	case "bcrypt":
		cmd = exec.Command("hashcat", "-m", "3200", "-a", "0", hash, "/usr/share/wordlists/rockyou.txt")
	default:
		cmd = exec.Command("hashcat", "-a", "3", hash, "?a?a?a?a?a?a")
	}

	if out, err := cmd.CombinedOutput(); err == nil {
		output := string(out)
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, hash) && strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					plaintext := parts[len(parts)-1]
					a.broadcast(fmt.Sprintf("[CRACKED] %s -> %s", hashType, plaintext))
					result := HashCrackResult{Hash: hash, Plaintext: plaintext, Type: hashType, Method: "hashcat"}
					a.state.mu.Lock()
					a.state.CrackResults = append(a.state.CrackResults, result)
					a.state.mu.Unlock()
				}
			}
		}
		_ = output
	} else {
		// john fallback
		a.broadcast("[HASH] hashcat failed, trying john...")
		if out2, err2 := exec.Command("john", "--wordlist=/usr/share/wordlists/rockyou.txt",
			fmt.Sprintf("--format=%s", hashType), hash).CombinedOutput(); err2 == nil {
			a.broadcast("[HASH] john attempt complete")
			_ = out2
		}
	}
}

func (a *HashBreaker) identifyHash(hash string) string {
	if out, err := exec.Command("hashid", "-j", hash).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, "MD5") {
			return "MD5"
		} else if strings.Contains(output, "SHA-256") {
			return "SHA256"
		} else if strings.Contains(output, "NTLM") {
			return "NTLM"
		} else if strings.Contains(output, "bcrypt") {
			return "bcrypt"
		}
		_ = output
	}
	return "unknown"
}
