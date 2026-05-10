// GAMMA-PERSONA — People OSINT & Social Engineering Agent
// theHarvester, sherlock, holehe, h8mail, phoneinfoga

package osintbreach

import (
	"fmt"
	"os/exec"
	"strings"
)

// PersonaAgent handles people reconnaissance
type PersonaAgent struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	msgCh   chan AgentMessage
}

func NewPersonaAgent(bus *EventBus, state *SharedState) *PersonaAgent {
	return &PersonaAgent{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 100),
	}
}

func (a *PersonaAgent) Name() string  { return "GAMMA-PERSONA" }
func (a *PersonaAgent) Status() string { return "online" }

func (a *PersonaAgent) Start() {
	a.bus.Subscribe("GAMMA", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *PersonaAgent) Stop() { close(a.stopCh) }

func (a *PersonaAgent) Handle(msg AgentMessage) {
	switch msg.Type {
	case "PERSONA_HUNT":
		a.personaHunt(msg.Data)
	case "EMAIL_ENUM":
		a.emailEnum(msg.Data)
	case "SOCIAL_HUNT":
		a.socialHunt(msg.Data)
	case "BREACH_CHECK":
		a.breachCheck(msg.Data)
	}
}

func (a *PersonaAgent) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "GAMMA", To: "ALL", Type: "LOG", Data: msg})
}

func (a *PersonaAgent) personaHunt(domain string) {
	a.broadcast(fmt.Sprintf("[GAMMA] People OSINT hunt: %s", domain))

	// theHarvester for emails and names
	a.emailEnum(domain)

	// Check breaches for found emails
	a.breachCheck(domain)

	// Social media hunting with sherlock
	a.socialHunt(domain)
}

func (a *PersonaAgent) emailEnum(domain string) {
	if out, err := exec.Command("theHarvester", "-d", domain, "-b", "all", "-l", "500").CombinedOutput(); err == nil {
		output := string(out)
		a.broadcast("[GAMMA] theHarvester complete")

		// Parse emails
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "@") && strings.Contains(line, domain) {
				email := strings.TrimSpace(line)
				email = strings.TrimPrefix(email, "[*] ")
				email = strings.TrimPrefix(email, "[+] ")

				a.state.mu.Lock()
				if _, ok := a.state.People[email]; !ok {
					a.state.People[email] = &Person{
						Email: email,
						Social: make(map[string]string),
					}
				}
				a.state.mu.Unlock()
			}
		}
	} else {
		a.broadcast(fmt.Sprintf("[!] theHarvester failed: %v", err))
	}

	// holehe - check where emails are registered
	a.state.mu.RLock()
	var emails []string
	for email := range a.state.People {
		emails = append(emails, email)
	}
	a.state.mu.RUnlock()

	for _, email := range emails {
		if out, err := exec.Command("holehe", email).CombinedOutput(); err == nil {
			output := string(out)
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, "[+]") {
					a.broadcast(fmt.Sprintf("[GAMMA] %s registered on: %s", email, line))
				}
			}
		}
	}
}

func (a *PersonaAgent) breachCheck(domain string) {
	a.broadcast("[GAMMA] Checking credential breaches...")

	// h8mail for breach checks
	a.state.mu.RLock()
	var emails []string
	for email := range a.state.People {
		emails = append(emails, email)
	}
	a.state.mu.RUnlock()

	if len(emails) > 0 {
		// h8mail -t targets.txt -c config.ini
		a.broadcast(fmt.Sprintf("[GAMMA] Checking %d emails for breaches", len(emails)))
		for _, email := range emails {
			if out, err := exec.Command("h8mail", "-t", email).CombinedOutput(); err == nil {
				output := string(out)
				if strings.Contains(output, "Password") || strings.Contains(output, "Hash") {
					a.broadcast(fmt.Sprintf("[CRITICAL] Breached credentials for %s", email))
				}
			}
		}
	}
}

func (a *PersonaAgent) socialHunt(domain string) {
	// sherlock - hunt usernames across social platforms
	a.broadcast("[GAMMA] Social media hunt with sherlock...")

	// Extract potential usernames from domain
	username := strings.Split(domain, ".")[0]
	if out, err := exec.Command("sherlock", username, "--timeout", "5").CombinedOutput(); err == nil {
		output := string(out)
		found := 0
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "+") {
				found++
			}
		}
		a.broadcast(fmt.Sprintf("[GAMMA] sherlock found %d social profiles for %s", found, username))
	}
}
