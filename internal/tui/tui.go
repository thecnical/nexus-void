package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(1, 2).
			Width(80)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF88")).
			Bold(true)

	findingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D"))

	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	highStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6600"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(1, 2).
			Width(78)
)

// Terminal represents a single agent terminal
type Terminal struct {
	Name     string
	Status   string
	Output   []string
	Active   bool
	Progress int
}

// Model is the TUI model
type Model struct {
	terminals   []Terminal
	activeTab   int
	width       int
	height      int
	spinner     spinner.Model
	showSpinner bool
	messages    []string
	findings    int
	exploits    int
	evolutions  int
	started     time.Time
}

func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))

	return Model{
		terminals: []Terminal{
			{Name: "RECON-OMEGA", Status: "idle", Output: []string{}, Active: true},
			{Name: "VULN-SENTINEL", Status: "idle", Output: []string{}, Active: false},
			{Name: "EXPLOIT-APOCALYPSE", Status: "idle", Output: []string{}, Active: false},
			{Name: "PERSISTENCE-DAEMON", Status: "idle", Output: []string{}, Active: false},
			{Name: "SHIELD-BREAKER", Status: "idle", Output: []string{}, Active: false},
			{Name: "C2-NEXUS", Status: "idle", Output: []string{}, Active: false},
		},
		spinner:     s,
		showSpinner: true,
		messages:    []string{},
		started:     time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % len(m.terminals)
		case "left":
			m.activeTab = (m.activeTab - 1 + len(m.terminals)) % len(m.terminals)
		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
		case "3":
			m.activeTab = 2
		case "4":
			m.activeTab = 3
		case "5":
			m.activeTab = 4
		case "6":
			m.activeTab = 5
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		// Simulate activity
		m.simulateActivity()
		return m, tickCmd()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) simulateActivity() {
	// Add simulated messages based on time
	elapsed := time.Since(m.started)

	if elapsed > 2*time.Second && len(m.messages) < 1 {
		m.messages = append(m.messages, "[RECON] Starting subdomain enumeration...")
		m.terminals[0].Status = "scanning"
		m.terminals[0].Output = append(m.terminals[0].Output, "Starting recon on target...")
	}

	if elapsed > 5*time.Second && len(m.messages) < 2 {
		m.messages = append(m.messages, "[RECON] Found 247 subdomains")
		m.terminals[0].Output = append(m.terminals[0].Output, "Found 247 subdomains")
		m.terminals[0].Progress = 30
	}

	if elapsed > 8*time.Second && len(m.messages) < 3 {
		m.messages = append(m.messages, "[VULN] Building attack graph...")
		m.terminals[0].Status = "complete"
		m.terminals[1].Status = "analyzing"
		m.terminals[1].Output = append(m.terminals[1].Output, "Analyzing 247 assets...")
		m.terminals[0].Progress = 100
		m.terminals[1].Progress = 20
	}

	if elapsed > 12*time.Second && len(m.messages) < 4 {
		m.messages = append(m.messages, "[EXPLOIT] Attempting SQLi on api.target.com")
		m.terminals[1].Status = "complete"
		m.terminals[2].Status = "exploiting"
		m.terminals[2].Output = append(m.terminals[2].Output, "Testing SQLi payloads...")
		m.terminals[1].Progress = 100
		m.terminals[2].Progress = 15
		m.findings = 1
	}

	if elapsed > 15*time.Second && len(m.messages) < 5 {
		m.messages = append(m.messages, "[EXPLOIT] SQLi PROVEN - Database: target_prod")
		m.terminals[2].Output = append(m.terminals[2].Output, "SQL Injection confirmed!")
		m.terminals[2].Output = append(m.terminals[2].Output, "Database: target_prod")
		m.terminals[2].Progress = 50
		m.findings = 3
		m.exploits = 1
	}

	if elapsed > 18*time.Second && len(m.messages) < 6 {
		m.messages = append(m.messages, "[SHIELD] Generating WAF bypass for payload #3")
		m.terminals[4].Status = "evolving"
		m.terminals[4].Output = append(m.terminals[4].Output, "Payload blocked by WAF")
		m.terminals[4].Output = append(m.terminals[4].Output, "Generating mutations...")
		m.evolutions = 3
	}

	if elapsed > 22*time.Second && len(m.messages) < 7 {
		m.messages = append(m.messages, "[SHIELD] WAF bypass successful - Score: 97/100")
		m.terminals[4].Status = "complete"
		m.terminals[4].Output = append(m.terminals[4].Output, "Bypass found! Score: 97")
		m.evolutions = 5
	}
}

func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("  NEXUS-VOID OMEGA - Autonomous Cyber-Weapon  ") + "\n\n")

	// Stats bar
	elapsed := time.Since(m.started)
	stats := fmt.Sprintf("  Target: %s | Elapsed: %s | Findings: %d | Exploits: %d | Evolutions: %d  ",
		"target.com", elapsed.Round(time.Second), m.findings, m.exploits, m.evolutions)
	b.WriteString(boxStyle.Render(stats) + "\n\n")

	// Terminal tabs
	b.WriteString("  ")
	for i, term := range m.terminals {
		style := dimStyle
		if i == m.activeTab {
			style = statusStyle
		}
		if term.Status != "idle" {
			style = style.Bold(true)
		}
		b.WriteString(fmt.Sprintf("[%d] %s ", i+1, style.Render(term.Name)))
	}
	b.WriteString("\n\n")

	// Active terminal content
	active := m.terminals[m.activeTab]
	b.WriteString(fmt.Sprintf("  %s Terminal: %s %s\n",
		m.spinner.View(),
		statusStyle.Render(active.Name),
		infoStyle.Render(fmt.Sprintf("(%s)", active.Status))))

	b.WriteString("  " + strings.Repeat("─", 76) + "\n")

	// Show last few lines of output
	start := len(active.Output) - 15
	if start < 0 {
		start = 0
	}
	for _, line := range active.Output[start:] {
		if strings.Contains(line, "PROVEN") || strings.Contains(line, "confirmed") {
			b.WriteString("  " + criticalStyle.Render(line) + "\n")
		} else if strings.Contains(line, "found") || strings.Contains(line, "Found") {
			b.WriteString("  " + findingStyle.Render(line) + "\n")
		} else {
			b.WriteString("  " + infoStyle.Render(line) + "\n")
		}
	}

	if len(active.Output) == 0 {
		b.WriteString("  " + dimStyle.Render("Waiting for activity...") + "\n")
	}

	b.WriteString("  " + strings.Repeat("─", 76) + "\n\n")

	// Live messages
	b.WriteString("  " + statusStyle.Render("LIVE FEED:") + "\n")
	for i := len(m.messages) - 5; i < len(m.messages); i++ {
		if i < 0 {
			continue
		}
		msg := m.messages[i]
		if strings.Contains(msg, "PROVEN") || strings.Contains(msg, "bypass") {
			b.WriteString("  " + criticalStyle.Render("> "+msg) + "\n")
		} else {
			b.WriteString("  " + infoStyle.Render("> "+msg) + "\n")
		}
	}

	// Help
	b.WriteString("\n  " + dimStyle.Render("Press 1-6 to switch terminals | q to quit") + "\n")

	return b.String()
}

// Run starts the TUI
func Run() {
	fmt.Println("[+] Starting NEXUS-VOID TUI...")
	m := NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
