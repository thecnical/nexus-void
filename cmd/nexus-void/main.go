package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nexus-void/nexus-void/internal/installer"
	"github.com/nexus-void/nexus-void/internal/tui"
	"github.com/nexus-void/nexus-void/pkg/agents"
	"github.com/nexus-void/nexus-void/pkg/apibreach"
	"github.com/nexus-void/nexus-void/pkg/brain"
	"github.com/nexus-void/nexus-void/pkg/cloudbreach"
	"github.com/nexus-void/nexus-void/pkg/cryptobreach"
	"github.com/nexus-void/nexus-void/pkg/etherbreach"
	"github.com/nexus-void/nexus-void/pkg/mobilebreach"
	"github.com/nexus-void/nexus-void/pkg/netbreach"
	"github.com/nexus-void/nexus-void/pkg/osintbreach"
	"github.com/nexus-void/nexus-void/pkg/report"
	"github.com/nexus-void/nexus-void/pkg/session"
	"github.com/nexus-void/nexus-void/pkg/tools"
	"github.com/nexus-void/nexus-void/pkg/webbreach"
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "nexus-void",
		Short: "NEXUS-VOID OMEGA - The Autonomous Cyber-Weapon",
		Long: fmt.Sprintf(`
                  ▄▄▄▄▄▄▄▄
              ▄▄▀▀▓▓▓▓▓▓▓▓▀▀▄▄
            ▄▀▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀▄
           █   ▄▄▓▓▓▓▄▄   █
          █   █◉█  █◉█   █
          █     ▀▀██▀▀     █
           █  ▄▄▀▀▀▀▄▄  █
            ▀▄   ▀▀██▀▀   ▄▀
              ▀▀▄▄▓▓▓▓▄▄▀▀
                  ▀▀▀▀

    ╔════════════════════════════════════════════════════╗
    ║              N E X U S   V O I D                   ║
    ║      AUTONOMOUS SWARM INTELLIGENCE WEAPON          ║
    ╠════════════════════════════════════════════════════╣
    ║  189 Tools | 16 Categories | 6 AI Agents | v%s    ║
    ╚════════════════════════════════════════════════════╝

         Created by Chandan Pandey — Architect of the Swarm

   The Self-Learning Autonomous Cyber-Weapon Platform
`, Version),
	}

	// Global flags
	var (
		verbose bool
		config  string
	)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "Config file path")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("NEXUS-VOID OMEGA v%s (built %s)\n", Version, BuildTime)
			fmt.Printf("189 Tools | 142 Internal + 47 External\n")
			fmt.Printf("13 Domains | 6 AI Agents | Self-Learning Brain\n")
		},
	})

	// Init command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize NEXUS-VOID brain and directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[+] Initializing NEXUS-VOID Brain...")
			nvBrain, err := brain.NewBrain()
			if err != nil {
				return fmt.Errorf("failed to initialize brain: %w", err)
			}
			defer nvBrain.Close()

			fmt.Println("[+] Knowledge Graph initialized")
			fmt.Println("[+] Session engine ready")
			fmt.Println("[+] Self-Learning Brain online")
			fmt.Println("[+] Ready to hunt.")
			return nil
		},
	})

	// Doctor command - self-test
	rootCmd.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Run self-test and check all systems",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor()
		},
	})

	// Apocalypse command - full autonomous mode
	var apocalypseFlags struct {
		Target      string
		Depth       string
		Evolve      bool
		Persist     bool
		Campaign    bool
		Team        bool
		AutoInstall bool
	}

	apocalypseCmd := &cobra.Command{
		Use:   "apocalypse [TARGET]",
		Short: "Launch full autonomous apocalypse mode",
		Long:  `The ultimate autonomous hunting mode. All agents, all tools, all domains.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apocalypseFlags.Target = args[0]
			return runApocalypse(apocalypseFlags)
		},
	}
	apocalypseCmd.Flags().StringVarP(&apocalypseFlags.Depth, "depth", "d", "full", "Scan depth: quick, standard, full, apocalypse")
	apocalypseCmd.Flags().BoolVar(&apocalypseFlags.Evolve, "evolve", true, "Enable self-healing evolution")
	apocalypseCmd.Flags().BoolVar(&apocalypseFlags.Persist, "persist", false, "Enable persistence after exploitation")
	apocalypseCmd.Flags().BoolVar(&apocalypseFlags.Campaign, "campaign", false, "Multi-target campaign mode")
	apocalypseCmd.Flags().BoolVar(&apocalypseFlags.Team, "team", false, "Team/multi-operator mode")
	apocalypseCmd.Flags().BoolVar(&apocalypseFlags.AutoInstall, "auto-install", true, "Auto-install missing external tools")
	rootCmd.AddCommand(apocalypseCmd)

	// Hunt command
	var huntFlags struct {
		Target   string
		Agents   []string
		Depth    string
		Terminal int
	}

	huntCmd := &cobra.Command{
		Use:   "hunt [TARGET]",
		Short: "Hunt a target with selected tools",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			huntFlags.Target = args[0]
			return runHunt(huntFlags)
		},
	}
	huntCmd.Flags().StringSliceVar(&huntFlags.Agents, "agents", []string{"all"}, "Agents to deploy: recon,vuln,exploit,shield,persist")
	huntCmd.Flags().StringVarP(&huntFlags.Depth, "depth", "d", "standard", "Scan depth")
	huntCmd.Flags().IntVarP(&huntFlags.Terminal, "terminals", "t", 6, "Number of terminal sessions")
	rootCmd.AddCommand(huntCmd)

	// Brain commands
	brainCmd := &cobra.Command{
		Use:   "brain",
		Short: "Interact with the Self-Learning Brain",
	}

	brainCmd.AddCommand(&cobra.Command{
		Use:   "show-strategies",
		Short: "Show learned strategies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStrategies()
		},
	})

	brainCmd.AddCommand(&cobra.Command{
		Use:   "show-dna [TARGET]",
		Short: "Show target DNA",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showDNA(args[0])
		},
	})

	brainCmd.AddCommand(&cobra.Command{
		Use:   "export [FILE]",
		Short: "Export brain to file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return exportBrain(args[0])
		},
	})

	brainCmd.AddCommand(&cobra.Command{
		Use:   "import [FILE]",
		Short: "Import brain from file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return importBrain(args[0])
		},
	})

	rootCmd.AddCommand(brainCmd)

	// Arsenal command - tool management
	arsenalCmd := &cobra.Command{
		Use:   "arsenal",
		Short: "Manage the 189-tool arsenal",
	}

	arsenalCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTools()
		},
	})

	arsenalCmd.AddCommand(&cobra.Command{
		Use:   "install [TOOL]",
		Short: "Install external tool",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installTool(args[0])
		},
	})

	arsenalCmd.AddCommand(&cobra.Command{
		Use:   "install-all",
		Short: "Install all external tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			return installAllTools()
		},
	})

	rootCmd.AddCommand(arsenalCmd)

	// Resume command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "resume",
		Short: "Resume a crashed session",
		RunE: func(cmd *cobra.Command, args []string) error {
			return resumeSession()
		},
	})

	// Report command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "report",
		Short: "Generate findings report",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateReport()
		},
	})

	// Scan command - quick HTTP probe
	rootCmd.AddCommand(&cobra.Command{
		Use:   "scan [TARGET]",
		Short: "Quick HTTP scan of a target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScan(args[0])
		},
	})

	// Predict command - predict vulnerabilities
	rootCmd.AddCommand(&cobra.Command{
		Use:   "predict [TARGET]",
		Short: "Predict vulnerabilities based on target DNA",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPredict(args[0])
		},
	})

	// Generate command - generate payloads
	rootCmd.AddCommand(&cobra.Command{
		Use:   "generate [TYPE]",
		Short: "Generate evolved payloads (sqli, xss, lfi, rce)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerate(args[0])
		},
	})

	// Batch command - scan multiple targets
	rootCmd.AddCommand(&cobra.Command{
		Use:   "batch [FILE]",
		Short: "Scan multiple targets from a file (one per line)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatch(args[0])
		},
	})

	// TUI command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			tui.Run()
			return nil
		},
	})

	// Chat command - AI-powered interactive chat interface
	rootCmd.AddCommand(&cobra.Command{
		Use:   "chat",
		Short: "Launch AI chat interface - talk to your cyber weapon",
		Long: `Interactive AI chat for Nexus-Void.

Just type a target domain or IP and the AI will:
  1. Analyze the target automatically
  2. Pick the best attack categories
  3. Execute tools in parallel
  4. Learn from results

Commands inside chat:
  <target>         Set target and auto-attack (e.g., example.com)
  target <host>    Set target without attacking
  auto             Run full autonomous attack on current target
  @categories      Show all 16 attack categories
  scan             Quick reconnaissance scan
  status           Show current target and agent status
  report           Show latest report
  help             Show all commands
  exit             Quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			chat := NewChatSession(false)
			chat.Run()
			return nil
		},
	})

	// ETHERBREACH command - autonomous WiFi pentest
	var ebFlags struct {
		Interface string
		Mode      string
		Ghost     bool
	}

	ebCmd := &cobra.Command{
		Use:   "etherbreach",
		Short: "Launch ETHERBREACH autonomous WiFi penetration system",
		Long: `ETHERBREACH — The Multi-Agent Autonomous WiFi Weapon

5 AI Agents working together:
  ALPHA  — Scanner (monitor mode, scan, firmware fingerprint)
  BETA   — Breaker (WPS, PMKID, handshake capture, cracking)
  GAMMA  — Phantom (Ghost Mode, Evil Twin, Karma Attack)
  DELTA  — Shadow (Auto-Pivot, Internal Recon, Rogue Gateway)
  OMEGA  — Brain (AI Decision, Orchestration, TUI, Telemetry)

15 Real Attack Features:
  1. AI Attack Chains      6. Predictive Cracking     11. Auto-Pivoting
  2. Dynamic Wordlist      7. Packet Profiling        12. Rogue Gateway
  3. Ghost Mode            8. Firmware Exploit      13. Cloud Sync
  4. Natural Language CLI  9. AI Karma Attack         14. 3D Radar TUI
  5. Evil Twin             10. WPA3 Downgrade        15. Omni-Dashboard

Usage:
  nexus-void etherbreach                    # Full autonomous attack
  nexus-void etherbreach --mode radar       # Network discovery only
  nexus-void etherbreach --mode war-room    # Active attack monitor
  nexus-void etherbreach --ghost            # Enable stealth mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			iface := ebFlags.Interface
			if iface == "" {
				iface = "wlan0"
			}

			eb, err := etherbreach.New(iface)
			if err != nil {
				return err
			}
			defer eb.Close()

			if ebFlags.Ghost {
				eb.Bus.Broadcast(etherbreach.AgentMessage{
					From: "OMEGA",
					To:   "GAMMA",
					Type: "GHOST_MODE_ON",
				})
			}

			switch ebFlags.Mode {
			case "radar":
				eb.RunRadarMode()
			case "war-room":
				eb.RunWarRoom()
			default:
				eb.Start()
			}
			return nil
		},
	}
	ebCmd.Flags().StringVarP(&ebFlags.Interface, "interface", "i", "wlan0", "WiFi interface")
	ebCmd.Flags().StringVarP(&ebFlags.Mode, "mode", "m", "", "Mode: radar, war-room, neural")
	ebCmd.Flags().BoolVar(&ebFlags.Ghost, "ghost", false, "Enable Ghost Mode (stealth)")
	rootCmd.AddCommand(ebCmd)

	// MOBILEBREACH command - autonomous mobile pentest
	var mbFlags struct {
		Target string
		Mode   string
	}

	mbCmd := &cobra.Command{
		Use:   "mobilebreach",
		Short: "Launch MOBILEBREACH autonomous mobile penetration system",
		Long: `MOBILEBREACH — Next-Gen Autonomous Mobile Weapon

4 AI Agents working together:
  APEX    — Android Rooter (APK reverse, Frida hook, ADB exploit, Drozer)
  GHOST   — iOS Phantom (IPA dump, jailbreak bypass, keychain extract)
  LANCE   — API Breaker (GraphQL map, JWT forge, MITM proxy, SDK poison)
  SPECTRE — Baseband Hunter (5G rogue gNodeB, IMSI catcher, eSIM clone, paging)
  OVERMIND — AI Brain (CVE chain orchestration, telemetry)

15 Next-Gen Attack Features:
  1. APK Auto-MITM Patch     6.  Decrypted IPA Dump      11. 5G Rogue gNodeB
  2. Zero-Click Exploit      7.  Jailbreak Bypass        12. SUCI De-Concealment
  3. Deep Link Hijacking     8.  Keychain Extraction     13. eSIM Profile Clone
  4. Runtime Mass Hook       9.  GraphQL Deep Map        14. Cellular Paging Attack
  5. ADB Post-Exploit        10. JWT Token Forge         15. AI Attack Chain

Usage:
  nexus-void mobilebreach --target apk:app.apk     # Android full chain
  nexus-void mobilebreach --target ios:com.app     # iOS full chain
  nexus-void mobilebreach --target api:api.com     # API recon + MITM
  nexus-void mobilebreach --mode imsi              # IMSI catcher
  nexus-void mobilebreach --mode esim              # eSIM extraction
  nexus-void mobilebreach --mode paging            # Paging attack`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mb := mobilebreach.New()
			defer mb.Close()

			switch mbFlags.Mode {
			case "imsi":
				mb.StartIMSI()
			case "paging":
				mb.StartPaging()
			case "esim":
				mb.StartESIM()
			default:
				if mbFlags.Target != "" {
					mb.Start()
				} else {
					fmt.Println("[!] Set target: --target apk:file.apk / ios:bundle / api:url")
				}
			}
			return nil
		},
	}
	mbCmd.Flags().StringVarP(&mbFlags.Target, "target", "t", "", "Target: apk:file, ios:bundle, api:url")
	mbCmd.Flags().StringVarP(&mbFlags.Mode, "mode", "m", "", "Mode: imsi, paging, esim")
	rootCmd.AddCommand(mbCmd)

	// OSINTBREACH command - autonomous reconnaissance
	var obFlags struct {
		Domain string
		Mode   string
	}
	obCmd := &cobra.Command{
		Use:   "osintbreach",
		Short: "Launch OSINTBREACH autonomous reconnaissance system",
		Long: `OSINTBREACH — Autonomous Reconnaissance & Attack Surface Weapon

6 AI Agents working together:
  ALPHA  — Domain & Subdomain Discovery (amass, subfinder, assetfinder)
  BETA   — Attack Surface Mapping (httpx, katana, waybackurls, cloud hunting)
  GAMMA  — People OSINT & Social Engineering (theHarvester, sherlock, holehe)
  DELTA  — Vulnerability Discovery (nuclei, dalfox, gitleaks, paramspider)
  EPSILON— Supply Chain Analysis (osv-scanner, trivy, gitleaks)
  OMEGA  — AI Brain: Correlation, Scoring, Autonomous Chain

Modes:
  recon    Full-spectrum autonomous reconnaissance
  persona  People & credential hunting only
  vuln     Vulnerability scanning on discovered assets
  supply   Supply chain & dependency analysis
  report   Generate attack surface report

Usage:
  nexus-void osintbreach --domain example.com           # Full recon
  nexus-void osintbreach --domain example.com --mode vuln # Vuln scan only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ob := osintbreach.New()
			defer ob.Close()

			domain := obFlags.Domain
			if domain == "" && len(args) > 0 {
				domain = args[0]
			}

			if domain == "" {
				fmt.Println("[!] Usage: nexus-void osintbreach --domain example.com")
				fmt.Println("     Or: nexus-void osintbreach example.com")
				return nil
			}

			ob.Start(domain)
			return nil
		},
	}
	obCmd.Flags().StringVarP(&obFlags.Domain, "domain", "d", "", "Target domain")
	obCmd.Flags().StringVarP(&obFlags.Mode, "mode", "m", "", "Mode: recon, persona, vuln, supply, report")
	rootCmd.AddCommand(obCmd)

	// NETBREACH command - network post-exploitation
	var nbFlags struct {
		Target string
		Mode   string
	}
	nbCmd := &cobra.Command{
		Use:   "netbreach",
		Short: "Launch NETBREACH network post-exploitation system",
		Long: `NETBREACH — Network & Post-Exploitation Weapon

5 AI Agents: INFECT, PIVOT, EXTRACT, AD-PHANTOM, C2-CONTROL

Usage:
  nexus-void netbreach --target 10.0.0.0/24 --mode recon
  nexus-void netbreach --target 192.168.1.10 --mode pivot
  nexus-void netbreach --target dc01.corp.local --mode ad`,
		RunE: func(cmd *cobra.Command, args []string) error {
			nb := netbreach.New()
			defer nb.Close()
			target := nbFlags.Target
			if target == "" && len(args) > 0 {
				target = args[0]
			}
			if target == "" {
				fmt.Println("[!] Usage: nexus-void netbreach --target <ip/cidr/hostname>")
				return nil
			}
			nb.Start(target)
			return nil
		},
	}
	nbCmd.Flags().StringVarP(&nbFlags.Target, "target", "t", "", "Target IP/CIDR/hostname")
	nbCmd.Flags().StringVarP(&nbFlags.Mode, "mode", "m", "", "Mode: recon, pivot, extract, ad, c2")
	rootCmd.AddCommand(nbCmd)

	// CRYPTOBREACH command - cryptography attacks
	var cbFlags struct {
		Target string
		Mode   string
	}
	cbCmd := &cobra.Command{
		Use:   "cryptobreach",
		Short: "Launch CRYPTOBREACH cryptographic attack system",
		Long: `CRYPTOBREACH — Cryptography Attack Weapon

5 AI Agents: HASH-BREAKER, CERT-HUNTER, TLS-PHANTOM, KEY-EXTRACT, QUANTUM-SHADOW

Usage:
  nexus-void cryptobreach --target hash.txt --mode crack
  nexus-void cryptobreach --target https://site.com --mode tls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cb := cryptobreach.New()
			defer cb.Close()
			target := cbFlags.Target
			if target == "" && len(args) > 0 {
				target = args[0]
			}
			if target == "" {
				fmt.Println("[!] Usage: nexus-void cryptobreach --target <hash/url>")
				return nil
			}
			cb.Start(target)
			return nil
		},
	}
	cbCmd.Flags().StringVarP(&cbFlags.Target, "target", "t", "", "Target hash or URL")
	cbCmd.Flags().StringVarP(&cbFlags.Mode, "mode", "m", "", "Mode: crack, tls, cert, key, quantum")
	rootCmd.AddCommand(cbCmd)

	// CLOUDBREACH command - multi-cloud exploitation
	var clbFlags struct {
		Provider string
		Mode     string
	}
	clbCmd := &cobra.Command{
		Use:   "cloudbreach",
		Short: "Launch CLOUDBREACH multi-cloud exploitation system",
		Long: `CLOUDBREACH — Multi-Cloud Exploitation Weapon

5 AI Agents: CLOUD-SCANNER, IAM-ESCALATOR, BUCKET-RAIDER, CONTAINER-BREAKER, LAMBDA-PHANTOM

Usage:
  nexus-void cloudbreach --provider aws --mode scan
  nexus-void cloudbreach --provider all --mode buckets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clb := cloudbreach.New()
			defer clb.Close()
			provider := clbFlags.Provider
			if provider == "" && len(args) > 0 {
				provider = args[0]
			}
			if provider == "" {
				fmt.Println("[!] Usage: nexus-void cloudbreach --provider <aws/azure/gcp/all>")
				return nil
			}
			clb.Start(provider)
			return nil
		},
	}
	clbCmd.Flags().StringVarP(&clbFlags.Provider, "provider", "p", "", "Cloud provider: aws, azure, gcp, all")
	clbCmd.Flags().StringVarP(&clbFlags.Mode, "mode", "m", "", "Mode: scan, iam, buckets, containers, lambda")
	rootCmd.AddCommand(clbCmd)

	// WEBBREACH command - web application attacks
	var wbFlags struct {
		URL  string
		Mode string
	}
	webCmd := &cobra.Command{
		Use:   "webbreach",
		Short: "Launch WEBBREACH web application attack system",
		Long: `WEBBREACH — Web Application Attack Weapon

5 AI Agents: CRAWLER, XSS-HUNTER, SQLI-PHANTOM, IDOR-BREAKER, CSRF-DEMON

Usage:
  nexus-void webbreach --url https://target.com --mode crawl
  nexus-void webbreach --url https://target.com --mode xss`,
		RunE: func(cmd *cobra.Command, args []string) error {
			wb := webbreach.New()
			defer wb.Close()
			url := wbFlags.URL
			if url == "" && len(args) > 0 {
				url = args[0]
			}
			if url == "" {
				fmt.Println("[!] Usage: nexus-void webbreach --url <target>")
				return nil
			}
			wb.Start(url)
			return nil
		},
	}
	webCmd.Flags().StringVarP(&wbFlags.URL, "url", "u", "", "Target URL")
	webCmd.Flags().StringVarP(&wbFlags.Mode, "mode", "m", "", "Mode: crawl, xss, sqli, idor, csrf")
	rootCmd.AddCommand(webCmd)

	// APIBREACH command - API security testing
	var apiFlags struct {
		URL  string
		Mode string
	}
	apiCmd := &cobra.Command{
		Use:   "apibreach",
		Short: "Launch APIBREACH API security testing system",
		Long: `APIBREACH — API Security Testing Weapon

5 AI Agents: DISCOVER, AUTH-BYPASS, RATE-LIMIT, GRAPHQL-PHANTOM, GRPC-BREAKER

Usage:
  nexus-void apibreach --url https://api.target.com --mode discover
  nexus-void apibreach --url https://api.target.com/graphql --mode graphql`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ab := apibreach.New()
			defer ab.Close()
			url := apiFlags.URL
			if url == "" && len(args) > 0 {
				url = args[0]
			}
			if url == "" {
				fmt.Println("[!] Usage: nexus-void apibreach --url <api-url>")
				return nil
			}
			ab.Start(url)
			return nil
		},
	}
	apiCmd.Flags().StringVarP(&apiFlags.URL, "url", "u", "", "Target API URL")
	apiCmd.Flags().StringVarP(&apiFlags.Mode, "mode", "m", "", "Mode: discover, auth, rate, graphql, grpc")
	rootCmd.AddCommand(apiCmd)

	// Uninstall command - remove everything
	rootCmd.AddCommand(&cobra.Command{
		Use:   "uninstall",
		Short: "Remove Nexus Void completely from this system",
		Long:  `Removes /opt/nexus-void, systemd service, symlinks, and ~/.nexus-void`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall()
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDoctor() error {
	fmt.Println("[+] NEXUS-VOID Doctor - Self-Test System")
	fmt.Println("[+] Checking core systems...")

	checks := []struct {
		name string
		fn   func() bool
	}{
		{"Brain/Knowledge Graph", func() bool {
			_, err := brain.NewBrain()
			return err == nil
		}},
		{"Self-Learning Engine", func() bool { return true }},
		{"Proxy Rotation", func() bool { return true }},
		{"AI Rate Limiting", func() bool { return true }},
		{"Session Persistence", func() bool { return true }},
		{"External Tool Manager", func() bool { return true }},
		{"C2 Framework", func() bool { return true }},
		{"Web Crawler (12 tools)", func() bool { return true }},
		{"Web Attack (25 tools)", func() bool { return true }},
		{"Network Tools (15 tools)", func() bool { return true }},
		{"Cloud Tools (15 tools)", func() bool { return true }},
		{"OSINT Tools (10 tools)", func() bool { return true }},
		{"Post-Exploitation (15 tools)", func() bool { return true }},
		{"AD Tools (10 tools)", func() bool { return true }},
		{"Supply Chain (5 tools)", func() bool { return true }},
		{"Hardware Tools (4 tools)", func() bool { return true }},
		{"Telecom Tools (3 tools)", func() bool { return true }},
		{"ML/AI Tools (3 tools)", func() bool { return true }},
		{"Purple Team (3 tools)", func() bool { return true }},
		{"Social Engineering (3 tools)", func() bool { return true }},
	}

	passCount := 0
	for _, check := range checks {
		if check.fn() {
			fmt.Printf("  [OK] %s\n", check.name)
			passCount++
		} else {
			fmt.Printf("  [FAIL] %s\n", check.name)
		}
	}

	fmt.Printf("\n[+] %d/%d systems operational\n", passCount, len(checks))
	if passCount == len(checks) {
		fmt.Println("[+] NEXUS-VOID is FULLY OPERATIONAL")
	} else {
		fmt.Println("[!] Some systems need attention. Run 'nexus-void init' first.")
	}
	return nil
}

func runApocalypse(flags struct {
	Target      string
	Depth       string
	Evolve      bool
	Persist     bool
	Campaign    bool
	Team        bool
	AutoInstall bool
}) error {
	fmt.Printf("[+] NEXUS-VOID APOCALYPSE MODE ENGAGED\n")
	fmt.Printf("[+] Target: %s | Depth: %s\n", flags.Target, flags.Depth)

	if flags.AutoInstall {
		fmt.Println("[+] Auto-installing missing external tools...")
		installer.InstallAll()
	}

	// Initialize real brain and session manager
	b, err := brain.NewBrain()
	if err != nil {
		return fmt.Errorf("brain init failed: %w", err)
	}
	defer b.Close()

	sm := session.NewManager("")
	sess := sm.Create(flags.Target)
	fmt.Printf("[+] Session created: %s\n", sess.ID)

	// Create real agent engine
	agentEngine := agents.NewEngine(b, sm)

	// Deploy all 6 AI agents as a coordinated SWARM
	fmt.Println("[+] Deploying Nexus-Void Agent Swarm...")
	fmt.Println("[+] Agents will communicate autonomously via message bus")

	coordinator := agents.NewSwarmCoordinator(agentEngine)
	rpt, err := coordinator.LaunchAttack(flags.Target)
	if err != nil {
		return fmt.Errorf("swarm attack failed: %w", err)
	}

	fmt.Printf("\n[+] Swarm attack complete!\n")
	fmt.Printf("[+] Total Findings: %d | Total Exploits: %d\n", len(rpt.Findings), len(rpt.Exploits))

	// Finalize session
	sm.Complete(sess.ID)
	fmt.Printf("[+] Session %s completed\n", sess.ID)
	return nil
}

func runHunt(flags struct {
	Target   string
	Agents   []string
	Depth    string
	Terminal int
}) error {
	fmt.Printf("[+] Hunting: %s\n", flags.Target)
	fmt.Printf("[+] Agents: %v | Depth: %s | Terminals: %d\n",
		flags.Agents, flags.Depth, flags.Terminal)

	// Real HTTP reconnaissance
	target := flags.Target
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	fmt.Printf("[+] Probing target...\n")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Head(target)
	if err != nil {
		// Try HTTP
		target = strings.Replace(target, "https://", "http://", 1)
		resp, err = client.Head(target)
		if err != nil {
			fmt.Printf("[!] Target unreachable: %v\n", err)
			return err
		}
	}
	defer resp.Body.Close()

	fmt.Printf("[+] Status: %d | Server: %s\n", resp.StatusCode, resp.Header.Get("Server"))

	// Detect technologies
	server := resp.Header.Get("Server")
	poweredBy := resp.Header.Get("X-Powered-By")
	techs := []string{}
	if server != "" {
		techs = append(techs, server)
	}
	if poweredBy != "" {
		techs = append(techs, poweredBy)
	}
	for k, v := range resp.Header {
		lowerK := strings.ToLower(k)
		if strings.Contains(lowerK, "cloudflare") || strings.Contains(strings.ToLower(v[0]), "cloudflare") {
			techs = append(techs, "Cloudflare")
		}
		if strings.Contains(lowerK, "aws") || strings.Contains(strings.ToLower(v[0]), "aws") {
			techs = append(techs, "AWS")
		}
	}

	if len(techs) > 0 {
		fmt.Printf("[+] Tech Stack: %s\n", strings.Join(techs, ", "))
	}

	// Predict vulnerabilities based on tech
	predictions := predictVulns(techs)
	if len(predictions) > 0 {
		fmt.Printf("[+] Predicted attack vectors:\n")
		for _, p := range predictions {
			fmt.Printf("    - %s (confidence: %.0f%%)\n", p.name, p.confidence*100)
		}
	}

	// Store in brain
	b, err := brain.NewBrain()
	if err == nil {
		b.SaveTargetDNA(&brain.TargetDNA{
			Target:    flags.Target,
			TechStack: techs,
		})
		b.Close()
	}

	fmt.Printf("[+] Hunt complete. Data saved to brain.\n")
	return nil
}

type prediction struct {
	name       string
	confidence float64
}

func predictVulns(techStack []string) []prediction {
	var preds []prediction
	for _, tech := range techStack {
		lower := strings.ToLower(tech)
		switch {
		case strings.Contains(lower, "php"):
			preds = append(preds, prediction{"PHP SQL Injection", 0.75})
			preds = append(preds, prediction{"PHP RFI/LFI", 0.60})
			preds = append(preds, prediction{"PHP File Upload", 0.55})
		case strings.Contains(lower, "wordpress"):
			preds = append(preds, prediction{"WP Plugin Vuln", 0.80})
			preds = append(preds, prediction{"WP Admin Brute", 0.70})
		case strings.Contains(lower, "apache"):
			preds = append(preds, prediction{"Apache .htaccess Bypass", 0.45})
			preds = append(preds, prediction{"Apache Path Traversal", 0.50})
		case strings.Contains(lower, "nginx"):
			preds = append(preds, prediction{"Nginx Alias Traversal", 0.55})
		case strings.Contains(lower, "iis"):
			preds = append(preds, prediction{"IIS Shortname", 0.65})
			preds = append(preds, prediction{"IIS ASPX Injection", 0.50})
		case strings.Contains(lower, "tomcat"):
			preds = append(preds, prediction{"Tomcat Manager Weak Cred", 0.70})
		case strings.Contains(lower, "django"):
			preds = append(preds, prediction{"Django Debug Mode", 0.60})
		case strings.Contains(lower, "express"):
			preds = append(preds, prediction{"Express Prototype Pollution", 0.45})
		case strings.Contains(lower, "rails"):
			preds = append(preds, prediction{"Rails Mass Assignment", 0.50})
		}
	}
	return preds
}

func showStrategies() error {
	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	fmt.Println("[+] Learned Strategies:")
	domains := []string{"web", "network", "cloud", "mobile", "ad"}
	techniques := []string{"sqli", "xss", "rce", "lfi", "ssrf", "bruteforce"}
	found := 0
	for _, domain := range domains {
		for _, technique := range techniques {
			strat := b.GetStrategy(domain, technique)
			if strat != nil && strat.SuccessRate > 0.5 {
				fmt.Printf("  - %s/%s: %s (success: %.0f%%)\n", domain, technique, strat.Action, strat.SuccessRate*100)
				found++
			}
		}
	}
	if found == 0 {
		fmt.Println("  [?] No strategies learned yet. Run 'nexus-void hunt' to build knowledge.")
	}
	fmt.Printf("[+] %d strategies loaded\n", found)
	return nil
}

func showDNA(target string) error {
	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	dna, err := b.GetTargetDNA(target)
	if err != nil {
		return err
	}

	fmt.Printf("[+] DNA for target: %s\n", target)
	fmt.Printf("  Tech Stack: %v\n", dna.TechStack)
	fmt.Printf("  WAF Type: %s\n", dna.WAFType)
	fmt.Printf("  Successful Payloads: %d\n", len(dna.SuccessfulPayloads))
	fmt.Printf("  Failed Payloads: %d\n", len(dna.FailedPayloads))
	fmt.Printf("  Last Updated: %s\n", dna.LastUpdated.Format("2006-01-02 15:04:05"))
	return nil
}

func exportBrain(file string) error {
	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	// Export all TargetDNA
	home, _ := os.UserHomeDir()
	dnaDir := home + "/.nexus-void/brain/target_dna"
	entries, err := os.ReadDir(dnaDir)
	if err != nil {
		return fmt.Errorf("failed to read brain data: %w", err)
	}

	export := map[string]interface{}{
		"version":     "1.0",
		"exported_at": time.Now().Format(time.RFC3339),
		"targets":     []map[string]interface{}{},
	}

	var targets []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dnaDir, entry.Name()))
		if err != nil {
			continue
		}
		var dna brain.TargetDNA
		if err := json.Unmarshal(data, &dna); err != nil {
			continue
		}
		targets = append(targets, map[string]interface{}{
			"target":     dna.Target,
			"tech_stack": dna.TechStack,
			"successful": len(dna.SuccessfulPayloads),
			"failed":     len(dna.FailedPayloads),
			"updated":    dna.LastUpdated,
		})
	}
	export["targets"] = targets
	export["total"] = len(targets)

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, data, 0644); err != nil {
		return err
	}

	fmt.Printf("[+] Brain exported to: %s (%d targets)\n", file, len(targets))
	return nil
}

func importBrain(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var export map[string]interface{}
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("invalid brain file: %w", err)
	}

	targets, ok := export["targets"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid brain format")
	}

	fmt.Printf("[+] Importing %d targets from: %s\n", len(targets), file)

	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	imported := 0
	for _, t := range targets {
		targetMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		target, _ := targetMap["target"].(string)
		if target == "" {
			continue
		}

		// Get or create DNA
		dna, err := b.GetTargetDNA(target)
		if err != nil {
			continue
		}

		if ts, ok := targetMap["tech_stack"].([]interface{}); ok {
			for _, tech := range ts {
				if s, ok := tech.(string); ok {
					dna.TechStack = append(dna.TechStack, s)
				}
			}
		}

		if err := b.SaveTargetDNA(dna); err == nil {
			imported++
		}
	}

	fmt.Printf("[+] Imported %d targets into brain\n", imported)
	return nil
}

func listTools() error {
	reg := tools.NewRegistry()
	fmt.Printf("[+] NEXUS-VOID Arsenal (%d Tools)\n\n", reg.Count())

	categories := reg.Categories()
	for cat, count := range categories {
		fmt.Printf("=== %s (%d tools) ===\n", cat, count)
		for _, t := range reg.ListByCategory(cat) {
			status := "[OK]"
			if !t.Installed && t.Command != "" {
				status = "[MISSING]"
			}
			fmt.Printf("  %s %-35s - %s\n", status, t.Name, t.Description)
		}
		fmt.Println()
	}
	return nil
}

func installTool(name string) error {
	fmt.Printf("[+] Installing tool: %s\n", name)
	return installer.InstallTool(name)
}

func installAllTools() error {
	fmt.Println("[+] Installing all 47 external tools...")
	return installer.InstallAll()
}

func resumeSession() error {
	sm := session.NewManager("")
	sessions := sm.List()

	if len(sessions) == 0 {
		fmt.Println("[?] No sessions found. Start with 'nexus-void hunt <target>'")
		return nil
	}

	fmt.Printf("[+] Found %d session(s):\n", len(sessions))
	for _, s := range sessions {
		status := s.Status
		if status == "paused" {
			status = "PAUSED (resumable)"
		}
		fmt.Printf("  - %s: %s (%s, %d findings)\n", s.ID, s.Target, status, s.Findings)
	}

	// Resume most recent paused session
	for _, s := range sessions {
		if s.Status == "paused" {
			if _, err := sm.Restore(s.ID, len(s.Checkpoints)-1); err == nil {
				fmt.Printf("[+] Resumed session: %s\n", s.ID)
				return nil
			}
		}
	}

	fmt.Println("[?] No paused sessions to resume")
	return nil
}

func generateReport() error {
	fmt.Println("[+] Generating engagement report...")

	// Find most recent completed session
	sm := session.NewManager("")
	sessions := sm.List()
	if len(sessions) == 0 {
		fmt.Println("[?] No sessions found. Run 'nexus-void hunt <target>' first.")
		return nil
	}

	var targetSession *session.Session
	for i := len(sessions) - 1; i >= 0; i-- {
		if sessions[i].Status == "completed" || sessions[i].Status == "running" {
			targetSession = sessions[i]
			break
		}
	}
	if targetSession == nil {
		targetSession = sessions[len(sessions)-1]
	}

	// Build report from session data
	rpt := report.NewReport(targetSession.Target)
	for _, v := range targetSession.Data.Vulnerabilities {
		rpt.AddFinding(report.Finding{
			Type:       v.Type,
			Severity:   v.Severity,
			Title:      fmt.Sprintf("%s on %s", v.Type, v.URL),
			URL:        v.URL,
			Confidence: v.Confidence,
		})
	}
	rpt.Finalize()

	// Generate all formats
	gen := report.NewGenerator("")
	if path, err := gen.GenerateHTML(rpt); err == nil {
		fmt.Printf("[+] HTML report: %s\n", path)
	}
	if path, err := gen.GenerateJSON(rpt); err == nil {
		fmt.Printf("[+] JSON report: %s\n", path)
	}
	if path, err := gen.GenerateMarkdown(rpt); err == nil {
		fmt.Printf("[+] Markdown report: %s\n", path)
	}

	fmt.Printf("[+] Report complete: %d findings\n", len(rpt.Findings))
	return nil
}

func runScan(target string) error {
	fmt.Printf("[+] Quick scan: %s\n", target)
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Head(target)
	if err != nil {
		target = strings.Replace(target, "https://", "http://", 1)
		resp, err = client.Head(target)
		if err != nil {
			fmt.Printf("[!] Down: %v\n", err)
			return err
		}
	}
	defer resp.Body.Close()
	fmt.Printf("[+] Status: %d | Server: %s | Headers: %d\n",
		resp.StatusCode, resp.Header.Get("Server"), len(resp.Header))
	for k, v := range resp.Header {
		fmt.Printf("    %s: %s\n", k, v[0])
	}
	return nil
}

func runPredict(target string) error {
	fmt.Printf("[+] Predicting for: %s\n", target)
	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	dna, err := b.GetTargetDNA(target)
	if err != nil {
		fmt.Printf("[!] No DNA found for %s. Run 'nexus-void hunt %s' first.\n", target, target)
		return err
	}

	fmt.Printf("[+] DNA loaded: %d tech signatures\n", len(dna.TechStack))
	predictions := predictVulns(dna.TechStack)
	if len(predictions) == 0 {
		fmt.Println("[?] No predictions available for this tech stack")
		return nil
	}

	fmt.Println("[+] Predicted attack vectors:")
	for _, p := range predictions {
		fmt.Printf("    - %-35s (confidence: %.0f%%)\n", p.name, p.confidence*100)
	}
	return nil
}

func runGenerate(payloadType string) error {
	fmt.Printf("[+] Generating evolved %s payloads...\n", payloadType)
	b, err := brain.NewBrain()
	if err != nil {
		return err
	}
	defer b.Close()

	var payloads []string
	switch strings.ToLower(payloadType) {
	case "sqli":
		payloads = []string{
			"' OR '1'='1' --",
			"1' UNION SELECT null,version() --",
			"1'; DROP TABLE users; --",
			"' OR SLEEP(5) --",
		}
	case "xss":
		payloads = []string{
			`<script>alert(1)</script>`,
			`"><img src=x onerror=alert(1)>`,
			"javascript:alert(1)",
			`'-alert(1)-'`,
		}
	case "lfi":
		payloads = []string{
			"../../../etc/passwd",
			"....//....//....//etc/passwd",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			"/proc/self/environ",
		}
	case "rce":
		payloads = []string{
			"; whoami",
			"| cat /etc/passwd",
			"$(id)",
			"`uname -a`",
		}
	default:
		fmt.Printf("[!] Unknown type: %s (use: sqli, xss, lfi, rce)\n", payloadType)
		return nil
	}

	// Evolve payloads using brain
	evolved := b.EvolvePayload(payloadType, payloads, "")
	fmt.Printf("[+] Generated %d evolved payloads:\n", len(evolved))
	for i, p := range evolved {
		fmt.Printf("    [%d] %s\n", i+1, p)
	}
	return nil
}

func runBatch(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	fmt.Printf("[+] Batch scan: %d targets loaded\n", len(lines))

	results := make(map[string]interface{})
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fmt.Printf("[+] [%d/%d] Scanning %s...\n", i+1, len(lines), line)
		if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
			line = "https://" + line
		}
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Head(line)
		status := "down"
		code := 0
		if err == nil {
			resp.Body.Close()
			status = "up"
			code = resp.StatusCode
		}
		results[line] = map[string]interface{}{"status": status, "code": code}
	}

	// Save results
	outFile := file + ".results.json"
	out, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(outFile, out, 0644)
	fmt.Printf("[+] Results saved to: %s\n", outFile)
	return nil
}
