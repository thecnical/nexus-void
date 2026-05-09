# NEXUS-VOID AI Agent System

## Overview

NEXUS-VOID OMEGA operates 10 autonomous AI agents that work in parallel, share intelligence through the Brain, and coordinate attacks via the Agent Coordinator.

## Agent Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  AGENT COORDINATOR                          │
│                    (Orchestrator)                           │
├─────────────────────────────────────────────────────────────┤
│  Task Queue → Dispatch → Agent Worker → Result Collector    │
│                                                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │ Agent 1 │  │ Agent 2 │  │ Agent 3 │  │ Agent N │        │
│  │  Goroutine│  │ Goroutine│  │ Goroutine│  │ Goroutine│      │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## Agent Types

### 1. RECON-OMEGA
**Specialization:** Reconnaissance & OSINT
**Capabilities:**
- Web crawling and endpoint discovery
- Subdomain enumeration (brute, DNS, certificate transparency)
- Technology fingerprinting (Wappalyzer-style detection)
- Port scanning (TCP/UDP with service detection)
- OSINT gathering (emails, social media, metadata)
- GitHub dorking for exposed credentials

**Status Values:** `idle` → `scanning` → `analyzing` → `complete`

### 2. VULN-SENTINEL
**Specialization:** Vulnerability Discovery
**Capabilities:**
- SQL injection testing (error-based, boolean-based, time-based)
- XSS detection (reflected, stored, DOM-based)
- LFI/RFI exploitation with wrapper bypasses
- SSRF testing against internal services
- JWT weakness analysis (none alg, weak secrets)
- Cloud misconfiguration scanning (S3, IAM, Lambda)

**Status Values:** `idle` → `scanning` → `analyzing` → `reporting`

### 3. EXPLOIT-APOCALYPSE
**Specialization:** Active Exploitation
**Capabilities:**
- Automated exploit selection from brain knowledge
- Multi-vector coordinated attacks
- Payload delivery via HTTP, DNS, SMB
- Privilege escalation (kernel exploits, misconfigs)
- Buffer overflow and format string testing
- Race condition detection

**Status Values:** `idle` → `planning` → `exploiting` → `complete`

### 4. PERSISTENCE-DAEMON
**Specialization:** Post-Exploitation & Persistence
**Capabilities:**
- Linux persistence (cron, systemd, SSH keys, .bashrc)
- Windows persistence (registry, scheduled tasks, WMI events, services)
- Credential dumping (LSASS, SAM, NTDS.dit, Kerberos tickets)
- Lateral movement (SSH, WMI, PsExec, WinRM, RDP)
- Data exfiltration (DNS tunneling, HTTP covert channels)
- Token impersonation and pass-the-hash

**Status Values:** `idle` → `exploiting` → `persisting` → `complete`

### 5. SHIELD-BREAKER
**Specialization:** Defense Evasion
**Capabilities:**
- AMSI bypass techniques
- EDR/AV evasion (process injection, unhooking)
- Obfuscation (base64, XOR, AES, custom encoding)
- Living-off-the-land (PowerShell, WMI, certutil)
- Anti-forensics (log clearing, timestomping, artifact wiping)
- Sandbox evasion (sleep, mouse checks, VM detection)

**Status Values:** `idle` → `analyzing` → `evolving` → `complete`

### 6. C2-NEXUS
**Specialization:** Command & Control
**Capabilities:**
- Malleable C2 profile management (4 built-in profiles)
- Beacon registration and tracking
- Task queueing and result retrieval
- AES-GCM encrypted communications
- Traffic mimicry (GitHub, Slack, AWS API patterns)
- Fallback communication channels

**Status Values:** `idle` → `planning` → `beaconing` → `complete`

### 7. NEURAL-PHOENIX
**Specialization:** ML Payload Evolution
**Capabilities:**
- Genetic algorithm payload generation
- SQLi payload mutation (encoding, comment injection, case variation)
- XSS payload evolution (event handlers, SVG vectors, polyglots)
- Command injection payload adaptation
- Payload scoring (bypass probability 0-100)
- Cross-generation learning (best parents selected)

**Status Values:** `idle` → `evolving` → `scoring` → `complete`

### 8. OSINT-PHANTOM
**Specialization:** Open-Source Intelligence
**Capabilities:**
- Email harvesting from search engines
- Subdomain discovery via DNS, certificates, archives
- GitHub repository dorking for secrets
- Social media profile enumeration
- Metadata extraction from documents
- Shodan/Censys integration for exposed services

**Status Values:** `idle` → `scanning` → `analyzing` → `complete`

### 9. SOCIAL-MANIPULATOR
**Specialization:** Social Engineering
**Capabilities:**
- AI-crafted phishing email generation
- Vishing (voice phishing) script creation
- Pretext scenario development
- Target psychological profiling
- USB bait content generation
- Spear-phishing payload customization

**Status Values:** `idle` → `profiling` → `crafting` → `complete`

### 10. PURPLE-GUARD
**Specialization:** Purple Team Defense
**Capabilities:**
- Sigma rule generation for SIEM detection
- YARA rule creation for malware identification
- Snort/Suricata rule development for network detection
- EQL (Event Query Language) threat hunting queries
- Incident response playbook generation
- Defense validation and gap analysis
- MITRE ATT&CK technique mapping

**Status Values:** `idle` → `analyzing` → `generating` → `complete`

## Agent Lifecycle

```
Spawn → Idle → Task Assignment → Execution → Result → Learn → Idle
  │                                           │
  └────────────────── Error ──────────────────┘
                       ↓
              Recovery / Retry
```

## Task Dispatch System

Tasks are dispatched based on priority and agent capability:

```go
// Priority levels
PriorityCritical = 10  // Active exploitation
PriorityHigh     = 8   // Vulnerability discovery
PriorityNormal   = 5   // Reconnaissance
PriorityLow      = 2   // Background learning
```

## Inter-Agent Communication

Agents communicate through the Brain's shared memory:

1. **RECON-OMEGA** discovers technology stack → stores in Target DNA
2. **VULN-SENTINEL** reads Target DNA → predicts vulnerabilities
3. **EXPLOIT-APOCALYPSE** reads predictions → executes attacks
4. **NEURAL-PHOENIX** evolves payloads based on EXPLOIT results
5. **PURPLE-GUARD** analyzes all findings → generates detection rules

## Agent Monitoring

Each agent exposes:
- **Status**: Current phase (idle, scanning, exploiting, etc.)
- **Progress**: 0-100% completion for current task
- **Output**: Last 100 log lines
- **Confidence**: Success probability (0.0 - 1.0)
- **Errors**: Count of failures in current session

## Real-Time Dashboard

Agents are visualized in the React dashboard:
- Live status indicators with color coding
- Progress bars for active tasks
- Terminal output streaming
- Confidence metrics over time
