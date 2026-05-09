<pre>
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

    ╔══════════════════════════════════════════╗
    ║   NEXUS-VOID  AI  ASSAULT  TERMINAL      ║
    ║   AUTONOMOUS SWARM INTELLIGENCE WEAPON   ║
    ╚══════════════════════════════════════════╝

    189 Tools | 16 Categories | 6 AI Agents | Self-Learning | Auto-Weaponized
</pre>

# NEXUS-VOID OMEGA v3.0

> **The world's first fully autonomous, self-learning, multi-agent offensive security platform with AI-driven chat orchestration.**

[![Version](https://img.shields.io/badge/version-3.0-red.svg)](https://github.com/nexus-void/nexus-void)
[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Tools](https://img.shields.io/badge/tools-189-orange.svg)](docs/ARSENAL.md)
[![AI Agents](https://img.shields.io/badge/agents-6-purple.svg)](docs/AGENTS.md)
[![Categories](https://img.shields.io/badge/categories-16-critical.svg)](docs/CATEGORIES.md)

---

## What Makes Nexus-Void Different

**This is NOT a normal pentest tool. This is an autonomous cyber-weapon.**

You give it a target. It thinks, plans, executes, and reports — all by itself.

### Core Philosophy

```
YOU:     "example.com"
AI:      *Analyzes target* *Selects categories* *Deploys agents* *Attacks* *Reports*
YOU:     Read the report. Fix the bugs.
```

**Zero manual commands needed.** Just type the target and the AI does everything.

---

## Platform Overview

**189 Tools** | **142 Internal Go Tools** | **47 External Auto-Installed Tools** | **16 Attack Categories** | **6 AI Agents** | **Self-Learning Brain v3.0**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      NEXUS-VOID OMEGA v3.0                          │
├─────────────────────────────────────────────────────────────────────┤
│  CLI Chat │ TUI Dashboard │ React Dashboard │ Team Server │ REST API │
├─────────────────────────────────────────────────────────────────────┤
│              AI ORCHESTRATOR (Target → Strategy → Attack)            │
├─────────────────────────────────────────────────────────────────────┤
│  Phase 1: Recon   → QuickProbe + OSINT + CVE Mapping              │
│  Phase 2: Intel   → ExploitDB + DarkWeb + LinkedIn + Mitigation   │
│  Phase 3: Attack  → Multi-Agent Swarm + Parallel Categories        │
│  Phase 4: Enrich  → Screenshots + Email + Jira + MISP              │
│  Phase 5: Evolve  → Genetic Payload + Auto-Weaponize + Purple     │
├─────────────────────────────────────────────────────────────────────┤
│  6 AI AGENTS COMMUNICATING VIA MESSAGE BUS                         │
│  RECON-OMEGA │ VULN-SENTINEL │ EXPLOIT-APOCALYPSE                  │
│  PERSISTENCE-DAEMON │ SHIELD-BREAKER │ C2-NEXUS                     │
├─────────────────────────────────────────────────────────────────────┤
│  16 CATEGORIES (ALL RUNNABLE IN PARALLEL)                           │
│  web_crawling │ web_app │ network │ cloud │ crypto │ osint           │
│  post_exploit │ c2 │ wireless │ hardware │ telecom │ ad             │
│  purple │ supply_chain │ social_eng │ ml                             │
├─────────────────────────────────────────────────────────────────────┤
│         142 Internal Go Modules │ 47 External Tools                 │
└─────────────────────────────────────────────────────────────────────┘
```

---

## AI Chat Interface (The Boss Mode)

**The chat interface is the main interaction point. Not a generic chat — a pure attack orchestrator.**

```bash
$ nexus-void chat

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

    ╔══════════════════════════════════════════╗
    ║   NEXUS-VOID  AI  ASSAULT  TERMINAL      ║
    ║   AUTONOMOUS SWARM INTELLIGENCE WEAPON   ║
    ╚══════════════════════════════════════════╝

[AI-SWARM] Autonomous attack system initialized.
           I am not a chatbot. I am your cyber-weapon.
           Give me a target and I will hunt, exploit, and report.

nexus-void ▸ example.com
```

### Commands

| Command | Action |
|---------|--------|
| `example.com` | **Instant full-spectrum assault** (just type the target) |
| `auto` | Attack last target with everything |
| `all` | **ALL 16 categories simultaneously** (MAX POWER) |
| `target <host>` | Set target without attacking |
| `scan` | Quick reconnaissance probe |
| `cve` | Map tech stack to known CVEs (NVD API) |
| `breach` | Dark web credential check (HaveIBeenPwned) |
| `osint` | LinkedIn employee enumeration |
| `@swarm` | Deploy all 6 agents simultaneously |
| `exploitdb` | Find public exploits for findings |
| `weaponize` | Auto-generate exploit scripts |
| `evolve` | AI payload evolution (genetic algorithm) |
| `purple` | Generate detection rules + EDR test |
| `mitigate` | Full OWASP remediation report |
| `@categories` | Show all 16 attack categories |
| `1,3,5,7` | Select categories by number |
| `web,network,ad` | Select categories by name |
| `status` | Show target, agents, strategy |
| `voice` | Toggle voice narration mode |
| `help` | Show all commands |
| `exit` | Abort mission |

---

## AI Flow — Step by Step

### What happens when you type a target?

```
YOU: example.com

Step 1: QUICK PROBE (5 seconds)
  ├── HTTP probe → nginx/1.18, PHP/7.4
  ├── DNS resolve → 93.184.216.34
  ├── Port scan → 80, 443, 8080 OPEN
  ├── MX check → mail server found
  └── Tech stack → [nginx, PHP, MySQL, jQuery]

Step 2: AI CLASSIFICATION
  ├── Web app detected → web_crawling + web_app
  ├── MySQL detected → SQLi high probability
  ├── Mail server → osint add
  └── Result: [web_crawling, web_app, osint, crypto]

Step 3: AUTO-CVE MAPPING (Phase 2)
  ├── NVD API → nginx CVEs
  ├── NVD API → PHP CVEs
  └── Found: CVE-2021-23017 (9.8 CRITICAL), CVE-2019-11043 (9.8 CRITICAL)

Step 4: STRATEGY BUILD
  ├── Categories: 4
  ├── Phases: 3 (Recon → Attack → Intel)
  ├── Parallel: true (goroutines)
  └── Est. Time: 2m30s

Step 5: SWARM DEPLOYMENT
  ├── RECON-OMEGA → Subdomains, URLs
  ├── VULN-SENTINEL → SQLi, XSS, LFI
  ├── EXPLOIT-APOCALYPSE → Verify exploits
  ├── PERSISTENCE-DAEMON → Backdoor paths
  ├── SHIELD-BREAKER → EDR bypass test
  └── C2-NEXUS → Beacon deployment

Step 6: PARALLEL EXECUTION
  ├── goroutine 1: web_crawling
  ├── goroutine 2: web_app (SQLi, XSS, LFI, SSRF)
  ├── goroutine 3: osint (emails, subdomains)
  └── goroutine 4: crypto (TLS, JWT)

Step 7: SCREENSHOT CAPTURE (Phase 4)
  ├── example.com_sqli_12345.html
  ├── example.com_xss_12346.html
  └── Saved to: reports/screenshots/

Step 8: EXPLOITDB CROSS-REFERENCE
  ├── CVE-2021-23017 → public exploit found
  └── Auto-weaponize triggered

Step 9: AUTO-EMAIL ALERT (Phase 4)
  └── "CRITICAL: 15 findings on example.com"

Step 10: JIRA TICKETS (Phase 4)
  └── WEB-001, WEB-002, WEB-003 auto-created

Step 11: MISP THREAT INTEL (Phase 4)
  └── IOCs pushed to threat sharing platform

Step 12: MITIGATION GUIDE
  └── OWASP-based remediation report generated

Step 13: AI LEARNING
  └── Target fingerprint saved for next time

Step 14: FINAL REPORT
  ├── HTML report with screenshots
  ├── JSON export for automation
  └── Dashboard real-time update

╔═════════════════════════════════════════════════════════════════════╗
║              ASSAULT COMPLETE - TARGET BREACHED                      ║
╠═════════════════════════════════════════════════════════════════════╣
║  Target:     example.com                                           ║
║  Duration:   2m15s                                                   ║
║  Findings:   15                                                      ║
║  Exploits:   4                                                       ║
║  Categories: 4                                                       ║
╚═════════════════════════════════════════════════════════════════════╝
```

---

## 6 AI Agents

All agents communicate via a real-time message bus. They coordinate autonomously.

| Agent | Role | What It Does |
|-------|------|--------------|
| **RECON-OMEGA** | Reconnaissance | Subdomains, URLs, ports, tech fingerprint |
| **VULN-SENTINEL** | Vulnerability Discovery | SQLi, XSS, LFI, SSRF, JWT, cloud misconfigs |
| **EXPLOIT-APOCALYPSE** | Active Exploitation | Verify vulnerabilities, chain attacks |
| **PERSISTENCE-DAEMON** | Post-Exploitation | Persistence, lateral movement, credential dump |
| **SHIELD-BREAKER** | EDR/AV Evasion | AMSI bypass, ETW patching, direct syscalls |
| **C2-NEXUS** | Command & Control | Malleable C2 profiles, beacon deployment |

---

## 16 Attack Categories (All Runnable in Parallel)

| # | Category | Tools | Description |
|---|----------|-------|-------------|
| 1 | **web_crawling** | 12 | URL discovery, sitemap, JavaScript analysis |
| 2 | **web_app** | 25 | SQLi, XSS, LFI, SSRF, JWT, command injection |
| 3 | **network** | 15 | Port scan, service enum, banner grab, OS fingerprint |
| 4 | **cloud** | 15 | AWS, Azure, GCP misconfig, S3 buckets, IAM |
| 5 | **crypto** | 8 | TLS analysis, JWT weaknesses, cert abuse |
| 6 | **osint** | 10 | Emails, subdomains, GitHub dorks, social media |
| 7 | **post_exploit** | 15 | Persistence, privilege escalation, lateral move |
| 8 | **c2** | 8 | Malleable profiles, beacon, exfiltration |
| 9 | **wireless** | 8 | WiFi scan, WPA crack, evil twin, Bluetooth |
| 10 | **hardware** | 4 | USB, UART, CAN bus, OBD-II |
| 11 | **telecom** | 3 | SS7, 5G NAS, Modbus, DNP3 |
| 12 | **ad** | 10 | Kerberoasting, DCSync, pass-the-hash, ACL abuse |
| 13 | **purple** | 5 | Sigma rules, YARA, EDR evasion testing |
| 14 | **supply_chain** | 5 | Dependency confusion, typosquatting, SBOM |
| 15 | **social_eng** | 3 | Phishing generation, vishing, pretext |
| 16 | **ml** | 5 | AI payload evolution, model poisoning, fuzzing |

---

## Phase 2: Intelligence Enrichment (100% Real)

### Auto-CVE Mapping
- Queries **National Vulnerability Database (NVD)** API
- Maps detected tech stack to known CVEs
- Scores severity: Critical / High / Medium / Low
- File: `pkg/intel/cve_mapper.go`

### ExploitDB Integration
- Cross-references findings with **ExploitDB** database
- Finds public exploits for confirmed CVEs
- File: `pkg/intel/exploitdb.go`

### Dark Web Intelligence
- Queries **HaveIBeenPwned** API for credential breaches
- Checks email patterns for target domain
- File: `pkg/intel/breach_intel.go`

### LinkedIn OSINT
- Employee enumeration and role extraction
- Organizational chart building
- Privileged target identification
- File: `pkg/intel/osint_linkedin.go`

### Mitigation Guide
- OWASP-based remediation for 7 vulnerability types
- Step-by-step fix instructions
- File: `pkg/intel/mitigation.go`

---

## Phase 3: Advanced Capabilities (100% Real)

### Voice Mode
- Toggle with `voice` command
- AI narrates every operation step
- File: `cmd/nexus-void/chat.go`

### Auto-Weaponization
- `weaponize` command generates 5 exploit scripts:
  - `python_exploit_sqli.py` → SQLi to RCE chain
  - `bash_reverse_shell.sh` → Multi-platform reverse shell
  - `powershell_lateral.ps1` → WMI + PowerShell remoting
  - `go_beacon_implant.go` → Custom C2 beacon
  - `yaml_k8s_escape.yml` → Kubernetes pod escape
- File: `cmd/nexus-void/chat.go`

### AI Payload Evolution (Genetic Algorithm)
- `evolve` command runs 5-generation genetic mutation:
  - Generation 1: 50 payloads → 12 survivors
  - Generation 2: Crossover + mutation → 48 children
  - Generation 3: 48 payloads → 8 high-fitness
  - Generation 4: 8 elites → 32 evolved variants
  - Generation 5: 32 payloads → 5 optimal selected
- File: `backend/internal/brain/brain.go`

### Purple Team Automation
- `purple` command generates:
  - Sigma detection rules (5 types)
  - EDR evasion testing (AMSI bypass, ETW patch, syscalls)
  - Blue team import-ready rules
- File: `cmd/nexus-void/chat.go`

---

## Phase 4: Full-Stack Integration (100% Real)

### Screenshot Auto-Capture
- Every vulnerable URL gets an HTML screenshot
- Includes response body, headers, metadata
- Saved to: `reports/screenshots/`
- File: `pkg/intel/screenshot.go`

### MISP/OpenCTI Integration
- Push IOCs to MISP threat sharing platform
- Auto-create events with findings
- Enrich with community threat intel
- File: `pkg/intel/misp.go`

### Auto-Report Email
- SMTP email alerts when findings are discovered
- Severity-based priority (Critical → Immediate)
- HTML-formatted report in email body
- File: `pkg/intel/emailer.go`

### Jira Ticket Auto-Creation
- Creates ticket per vulnerability
- Labels: `nexus-void`, `auto-discovered`
- Priority maps to severity
- File: `pkg/intel/jira.go`

### Multi-Target Campaign Mode
- Attack 100+ domains simultaneously
- Shared strategy across similar targets
- Aggregated reporting
- File: `pkg/ai/orchestrator.go`

---

## Quick Install

### Option 1: PowerShell (Windows)
```powershell
git clone https://github.com/nexus-void/nexus-void.git
cd nexus-void
go build -o nexus-void.exe ./cmd/nexus-void
```

### Option 2: Manual (All Platforms)
```bash
git clone https://github.com/nexus-void/nexus-void.git
cd nexus-void
go build -o nexus-void ./cmd/nexus-void
```

### Option 3: Docker
```bash
docker build -t nexus-void .
docker run -it nexus-void chat
```

---

## Usage Examples

### Example 1: Web Application (One Shot)
```bash
$ nexus-void chat
nexus-void ▸ testphp.vulnweb.com

[AI-SWARM] Target acquired: testphp.vulnweb.com
[TARGET-ANALYSIS] Confidence: 85%
  ✓ Web application stack detected
  ✓ Exposed services: [80, 443]
  ✓ Technology fingerprint: [Apache, PHP, MySQL]
  → Recommended: [web_crawling, web_app, osint, crypto]

[?] DEPLOY OPTIONS:
      auto    → Full AI-driven autonomous assault
      @swarm  → Deploy all 6 agents simultaneously
      cve     → Map known CVEs to tech stack
```

### Example 2: Internal Network Infrastructure
```bash
nexus-void ▸ 192.168.1.1

[TARGET-ANALYSIS] Confidence: 70%
  ✓ Network infrastructure present
  ✓ Exposed services: [22, 80, 445, 3389]
  ✓ Active Directory indicators
  → Recommended: [network, ad, post_exploit, c2]
```

### Example 3: Maximum Firepower (All 16 Categories)
```bash
nexus-void ▸ target example.com
nexus-void ▸ all

╔═════════════════════════════════════════════════════════════════════╗
║         ALL 16 CATEGORIES - MAXIMUM FIREPOWER                      ║
╚═════════════════════════════════════════════════════════════════════╝
[AI] Deploying ALL 16 categories simultaneously against example.com
```

### Example 4: Manual Category Selection
```bash
nexus-void ▸ target example.com
nexus-void ▸ @categories
nexus-void ▸ 1,3,5,7
# or
nexus-void ▸ web,network,ad
```

---

## Backend Server (Team Dashboard)

```bash
# Terminal 1: Start team server
cd backend
go run ./cmd/server -addr :8080

# Terminal 2: Start React dashboard
cd dashboard && npm start
```

Access at `http://localhost:3000` with real-time WebSocket streaming.

---

## Self-Learning Brain v3.0

The brain remembers everything:

- **Memory Module** — Short-term and long-term storage
- **Learn Module** — Tracks success/failure per technique per domain
- **Evolve Module** — Genetic payload mutation across generations
- **Reason Module** — Rule-based autonomous decision making
- **Predict Module** — Heuristic vulnerability prediction per tech stack

```
Attack example.com → Brain saves strategy →
Next similar target → Brain recalls best strategy →
Success rate improves over time
```

---

## File Structure

```
nexus-void/
├── cmd/nexus-void/          # CLI application
│   ├── main.go              # Main entry + subcommands
│   └── chat.go              # AI chat interface (BOSS MODE)
├── pkg/
│   ├── ai/
│   │   └── orchestrator.go  # AI target analysis + strategy + execution
│   ├── agents/
│   │   ├── engine.go        # 6 AI agent implementations
│   │   └── swarm.go         # Multi-agent message bus coordination
│   ├── brain/               # Self-learning brain (Memory/Learn/Evolve/Reason/Predict)
│   ├── intel/               # Phase 2-4 intelligence modules
│   │   ├── cve_mapper.go    # NVD API CVE mapping
│   │   ├── exploitdb.go     # ExploitDB cross-reference
│   │   ├── breach_intel.go  # HaveIBeenPwned dark web check
│   │   ├── osint_linkedin.go # LinkedIn employee enumeration
│   │   ├── mitigation.go    # OWASP remediation guides
│   │   ├── screenshot.go    # Auto-capture vulnerable URLs
│   │   ├── misp.go          # MISP threat intel integration
│   │   ├── emailer.go       # SMTP auto-alert on findings
│   │   └── jira.go          # Auto-ticket creation
│   ├── web/                 # Web attack tools (SQLi, XSS, LFI, JWT, etc.)
│   ├── network/             # Network tools (port scan, SYN, banner grab)
│   ├── tools/               # 142 internal Go tools
│   └── report/              # Report generation (HTML, JSON)
├── backend/                 # Team server + Brain API
│   ├── internal/brain/      # Brain v3.0 with SQLite + GORM
│   └── internal/server/     # WebSocket team server
├── dashboard/               # React real-time dashboard
└── README.md                # This file
```

---

## 100% Real vs Partial — Audit Report

| Component | Status | File |
|-----------|--------|------|
| **AI Orchestrator** | 100% Real — Target analysis, classification, strategy, parallel execution, learning | `pkg/ai/orchestrator.go` |
| **Chat Interface** | 100% Real — REPL, ANSI colors, ASCII logo, 20+ commands | `cmd/nexus-void/chat.go` |
| **6 AI Agents** | 100% Real — Full implementations with message bus | `pkg/agents/` |
| **Auto-CVE Mapping** | 100% Real — NVD API queries with scoring | `pkg/intel/cve_mapper.go` |
| **ExploitDB** | 100% Real — Cross-reference with search | `pkg/intel/exploitdb.go` |
| **Dark Web Intel** | 100% Real — HIBP breach checking | `pkg/intel/breach_intel.go` |
| **LinkedIn OSINT** | 100% Real — Employee enum with role mapping | `pkg/intel/osint_linkedin.go` |
| **Mitigation Guide** | 100% Real — OWASP-based remediation | `pkg/intel/mitigation.go` |
| **Screenshot Capture** | 100% Real — HTML screenshots with metadata | `pkg/intel/screenshot.go` |
| **MISP Integration** | 100% Real — IOC push + event creation | `pkg/intel/misp.go` |
| **Auto-Email** | 100% Real — SMTP alerts on findings | `pkg/intel/emailer.go` |
| **Jira Tickets** | 100% Real — Per-vulnerability ticket creation | `pkg/intel/jira.go` |
| **Voice Mode** | 100% Real — Toggle narration | `cmd/nexus-void/chat.go` |
| **Auto-Weaponization** | 100% Real — 5 exploit script templates | `cmd/nexus-void/chat.go` |
| **Payload Evolution** | 100% Real — 5-gen genetic algorithm | `backend/internal/brain/brain.go` |
| **Purple Team** | 100% Real — Sigma rules + EDR testing | `cmd/nexus-void/chat.go` |
| **Port Scan** | 100% Real — Parallel TCP connect + banner grab | `pkg/network/portscan.go` |
| **SYN Scan** | 100% Real — Fast-connect fallback + IPv6 support | `pkg/network/portscan.go` |
| **JWT Scanner** | 100% Real — Token analysis + weakness detection | `pkg/agents/engine.go` |
| **Brain v3.0** | 100% Real — SQLite + GORM with all 5 modules | `backend/internal/brain/brain.go` |
| **Base64 Encode** | 100% Real — Actual base64 encoding | `backend/internal/brain/brain.go` |
| **IPv6 Support** | 100% Real — net.JoinHostPort everywhere | All network files |

---

## License

MIT License — See [LICENSE](LICENSE) for details.

> **Warning:** This tool is for **authorized security testing only**. Unauthorized access to computer systems is illegal.

---

<p align="center">
<b>NEXUS VOID v3.0</b><br>
Autonomous Swarm Intelligence Weapon<br>
Created by <b>Chandan Pandey</b> | <a href="https://cybermindcli.com">cybermindcli.com</a><br>
Built for the future of offensive security
</p>

---

## Full System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    USER INTERFACES                                          │
├──────────────┬──────────────┬──────────────────┬─────────────────┬──────────────────────────┤
│  CLI Chat    │  TUI Mode    │  React Dashboard │  REST API       │  WebSocket Real-Time     │
│  (Terminal)  │  (Fancy UI)  │  (Browser)       │  (HTTP/JSON)    │  (Live Stream)           │
└──────┬───────┴──────┬───────┴────────┬─────────┴────────┬────────┴──────────┬───────────────┘
       │              │                │                │                   │
       └──────────────┴────────────────┴────────────────┘                   │
                              │                                              │
                    ┌─────────┴──────────┐                                   │
                    │  AI ORCHESTRATOR   │◄─────────────────────────────────┘
                    │  (Brain + Strategy)│
                    └─────────┬──────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼────────┐  ┌────────▼────────┐  ┌──────▼──────┐
│   COGNITION    │  │   EXECUTION     │  │  REPORTING  │
│                │  │                 │  │             │
│  • Target DNA  │  │  • Multi-Agent  │  │  • HTML     │
│  • Strategy    │  │    Swarm        │  │  • JSON     │
│  • Prediction  │  │  • Parallel     │  │  • Email    │
│  • Learning    │  │    Categories   │  │  • Jira     │
│  • Evolution   │  │  • Exploit Chain  │  │  • MISP     │
└───────┬────────┘  └────────┬────────┘  │  • Screens  │
        │                    │            └──────┬──────┘
        │                    │                   │
        │            ┌───────┴────────┐        │
        │            │  16 CATEGORIES │        │
        │            │  (Parallel)    │        │
        │            └───────┬────────┘        │
        │                    │                 │
        │     ┌──────┬──────┼──────┬──────┐   │
        │     ▼      ▼      ▼      ▼      ▼   │
        │  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐│
        │  │Web │ │Net │ │Cloud│ │AD  │ │C2  ││ ... etc
        │  └────┘ └────┘ └────┘ └────┘ └────┘│
        │                                    │
        └──────────────┬─────────────────────┘
                       │
              ┌────────▼────────┐
              │  DATA LAYER     │
              │                 │
              │  • SQLite Brain │
              │  • GORM ORM     │
              │  • Target DNA   │
              │  • Session Log  │
              │  • Strategy DB  │
              └─────────────────┘
```

---

## Attack Categories Architecture

Each category runs as an independent goroutine with its own toolset:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         16 ATTACK CATEGORIES                                   │
│                     (All Runnable In Parallel)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ WEB CRAWL   │  │ WEB APP     │  │ NETWORK     │  │ CLOUD       │      │
│  │ ─────────── │  │ ─────────── │  │ ─────────── │  │ ─────────── │      │
│  │ • Spider    │  │ • SQLi      │  │ • Port Scan │  │ • S3 Bucket │      │
│  │ • Sitemap   │  │ • XSS       │  │ • Banner    │  │ • IAM Abuse │      │
│  │ • JS Parse  │  │ • LFI       │  │ • OS Fp     │  │ • Container │      │
│  │ • Form Map  │  │ • SSRF      │  │ • SYN Scan  │  │ • K8s Escape│      │
│  │             │  │ • JWT       │  │             │  │             │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ OSINT       │  │ CRYPTO      │  │ POST-EXP    │  │ C2          │      │
│  │ ─────────── │  │ ─────────── │  │ ─────────── │  │ ─────────── │      │
│  │ • LinkedIn  │  │ • TLS Audit │  │ • Persist   │  │ • Beacon    │      │
│  │ • GitHub    │  │ • JWT Crack │  │ • PrivEsc   │  │ • Exfil     │      │
│  │ • Shodan    │  │ • Cert Abus │  │ • Lateral   │  │ • Profile   │      │
│  │ • HIBP      │  │ • Hash Crack│  │ • Dump Cred │  │ • Domain    │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ WIRELESS    │  │ HARDWARE    │  │ TELECOM     │  │ AD          │      │
│  │ ─────────── │  │ ─────────── │  │ ─────────── │  │ ─────────── │      │
│  │ • WiFi Scan │  │ • UART      │  │ • SS7       │  │ • Kerberoast│      │
│  │ • WPA Crack │  │ • CAN Bus   │  │ • 5G NAS    │  │ • DCSync    │      │
│  │ • Evil Twin │  │ • USB       │  │ • Modbus    │  │ • Pass Hash │      │
│  │ • Bluetooth │  │ • OBD-II    │  │ • DNP3      │  │ • ACL Abuse │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ PURPLE      │  │ SUPPLY CHAIN│  │ SOCIAL ENG  │  │ ML/AI       │      │
│  │ ─────────── │  │ ─────────── │  │ ─────────── │  │ ─────────── │      │
│  │ • Sigma     │  │ • Dep Conf  │  │ • Phishing  │  │ • GA Evolve │      │
│  │ • YARA      │  │ • TypoSquat │  │ • Vishing   │  │ • Model Poi │      │
│  │ • EDR Test  │  │ • SBOM      │  │ • Pretext   │  │ • Fuzz AI   │      │
│  │ • Blue Team │  │ • Audit     │  │ • Deepfake  │  │ • Adversar  │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## How Nexus Void Works (Visual Flow)

```
  YOU                              NEXUS VOID AI
   │                                    │
   │   example.com                      │
   │ ──────────────────────────────────>│
   │                                    │
   │                              ┌─────┴─────┐
   │                              │ QUICKPROBE│
   │                              │  (5 sec)  │
   │                              │ HTTP/DNS/ │
   │                              │ PORT/TECH │
   │                              └─────┬─────┘
   │                                    │
   │                              ┌─────┴─────┐
   │                              │ AI CLASS  │
   │                              │ Web? Net? │
   │                              │ Cloud? AD?│
   │                              └─────┬─────┘
   │                                    │
   │                              ┌─────┴─────┐
   │                              │ CVE MAP   │
   │                              │ NVD API   │
   │                              └─────┬─────┘
   │                                    │
   │                              ┌─────┴─────┐
   │                              │ STRATEGY  │
   │                              │ Build Plan│
   │                              └─────┬─────┘
   │                                    │
   │        ┌─────────────────────────────┼─────────────────────────────┐
   │        │                             │                             │
   │   ┌────┴────┐  ┌────┴────┐  ┌────┴────┐  ┌────┴────┐  ┌────┴────┐
   │   │ AGENT 1 │  │ AGENT 2 │  │ AGENT 3 │  │ AGENT 4 │  │ AGENT 5 │
   │   │ RECON   │  │ VULN    │  │ EXPLOIT │  │ SHIELD  │  │ C2      │
   │   └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
   │        │            │            │            │            │
   │        └────────────┴────────────┴────────────┴────────────┘
   │                                    │
   │                              ┌─────┴─────┐
   │                              │ REPORT    │
   │                              │ + Learn   │
   │                              └─────┬─────┘
   │                                    │
   │  ┌─────────────────────────────────┘
   │  │
   │  ▼  HTML Report + Email + Jira + MISP + Screenshots
   │
   │  "15 findings, 4 exploits, 2m15s"
   │ <──────────────────────────────────│
```

---

## AI Conversation Mode

Nexus Void is not just a tool — it is a conversational cybersecurity AI.

### You Can Ask It Anything:

| Question | Response |
|----------|----------|
| `"who are you"` | Full self-introduction with all capabilities |
| `"who created you"` | Information about Chandan Pandey and cybermindcli.com |
| `"security tips"` | 10 best practices with explanations |
| `"attack plan"` | Full PTES methodology breakdown |
| `"hi" / "namaste"` | Greeting in multiple languages |
| `"what is SQLi"` | Detailed SQL injection explanation |
| `"what is XSS"` | Cross-site scripting breakdown |
| `"what is RCE"` | Remote code execution details |
| `"cloud security"` | AWS/Azure/GCP security overview |
| `"phishing"` | Social engineering deep dive |

### Multi-Language Support

Works in **English, Hindi (Hinglish), Urdu, and more**:
- `"tum kaun ho"` → AI introduces itself
- `"tumhe kisne banaya"` → Shows creator info
- `"kya kar sakte ho"` → Lists all capabilities
- `"kaise ho"` → Friendly greeting

---

## Render Deployment (Backend)

Deploy the Nexus Void backend to Render for free:

### Step 1: Push to GitHub

```bash
git init
git add .
git commit -m "Nexus Void v3.0 - Full Stack"
git branch -M main
git remote add origin https://github.com/nexus-void/nexus-void.git
git push -u origin main
```

### Step 2: Create Render Account

1. Go to [render.com](https://render.com) and sign up
2. Click **"New +"** → **"Web Service"**
3. Connect your GitHub repository
4. Select the `backend` directory as root

### Step 3: Configure Environment Variables

In Render dashboard, add these env vars:

```
GO_ENV=production
PORT=8080
NEXUS_BRAIN_DIR=/var/lib/nexus-void/brain
NEXUS_AI_OPENROUTER_KEY=your_key_here        (optional)
NEXUS_JIRA_URL=https://yourcompany.atlassian.net  (optional)
NEXUS_JIRA_USER=your_email                        (optional)
NEXUS_JIRA_TOKEN=your_api_token                   (optional)
NEXUS_EMAIL_SMTP_HOST=smtp.gmail.com              (optional)
NEXUS_EMAIL_SMTP_PORT=587                         (optional)
NEXUS_EMAIL_USER=your_email                       (optional)
NEXUS_EMAIL_PASS=your_app_password                (optional)
```

### Step 4: Deploy

Render will auto-build from `backend/Dockerfile` and deploy.

Your backend will be live at: `https://nexus-void-backend.onrender.com`

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/status` | GET | Server health check |
| `/api/sessions` | GET | List all sessions |
| `/api/agents` | GET | List all agent statuses |
| `/api/brain/stats` | GET | Brain statistics |
| `/api/brain/strategies` | GET | Learned strategies |
| `/api/recon` | POST | Start reconnaissance |
| `/api/scan` | POST | Start vulnerability scan |
| `/ws` | WebSocket | Real-time updates |

---

## GitHub Repository Setup

### Recommended Structure

Create **3 separate repos** for maximum virality:

```
nexus-void/           (main monorepo - this one)
├── README.md         (comprehensive docs)
├── cmd/              (CLI)
├── pkg/              (core libraries)
├── backend/          (server)
├── dashboard/        (React UI)
├── install.sh        (Linux installer)
└── Dockerfile        (container)

nexus-void-backend/   (standalone backend for Render)
├── go.mod
├── cmd/server/
├── internal/
├── Dockerfile
└── render.yaml

nexus-void-dashboard/ (standalone frontend)
├── package.json
├── src/
├── public/
└── README.md
```

### Splitting for GitHub

```bash
# Backend repo
git subtree split --prefix=backend --branch backend-main
git push https://github.com/nexus-void/nexus-void-backend.git backend-main:main

# Dashboard repo
git subtree split --prefix=dashboard --branch dashboard-main
git push https://github.com/nexus-void/nexus-void-dashboard.git dashboard-main:main
```

---

## Linux Installation (One Command)

```bash
# Download and install
curl -fsSL https://raw.githubusercontent.com/nexus-void/nexus-void/main/install.sh | bash

# Or wget
wget -qO- https://raw.githubusercontent.com/nexus-void/nexus-void/main/install.sh | bash
```

### What the installer does:

1. Checks for Go (installs if missing)
2. Checks for Git (installs if missing)
3. Clones the repository to `/opt/nexus-void`
4. Builds the CLI binary
5. Builds the backend server
6. Creates symlinks in `/usr/local/bin`
7. Sets up systemd service for backend
8. Prints usage instructions

### Post-Install Commands

```bash
# Start CLI chat
nexus-void chat

# Start backend server
nexus-server -addr :8080

# Or as systemd service
sudo systemctl start nexus-void-backend
sudo systemctl enable nexus-void-backend
```

---

## Dashboard Screenshots

The React dashboard provides:

- **Real-time WebSocket** connection to backend
- **Live agent status** with progress bars
- **Finding feed** with severity coloring
- **Brain statistics** showing learned strategies
- **Target scan bar** for instant deployment
- **Dark cyberpunk theme** matching the CLI aesthetic

```bash
cd dashboard
npm install
npm start
# Opens http://localhost:3000
```

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| **CLI** | Go 1.24 + Cobra + bufio |
| **Backend** | Go 1.24 + Gorilla WebSocket + GORM |
| **Database** | SQLite (pure-Go via glebarez) |
| **Dashboard** | React 18 + TypeScript + Recharts + Lucide |
| **Container** | Docker + Alpine Linux |
| **Deployment** | Render (Web Service) |
| **Brain** | 5 AI modules (Memory/Learn/Evolve/Reason/Predict) |

---

## Made With Love by Chandan Pandey

<p align="center">
<b>Chandan Pandey</b> — Architect of the Swarm<br>
<a href="https://cybermindcli.com">cybermindcli.com</a> | <a href="https://github.com/chandanpandey">GitHub</a><br>
<i>Defend with wisdom. Attack with precision.</i>
</p>
