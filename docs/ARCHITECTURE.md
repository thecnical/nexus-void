# NEXUS-VOID OMEGA Architecture

## System Overview

NEXUS-VOID OMEGA is built on a modular, multi-agent architecture with a central autonomous brain that coordinates all operations.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PRESENTATION LAYER                                │
├──────────────┬────────────────┬─────────────────┬───────────────────────────┤
│  CLI Core    │   TUI          │ React Dashboard │   VSCode Extension        │
│  (cobra)     │  (bubbletea)   │  (websocket)    │   (future)                │
└──────────────┴────────────────┴─────────────────┴───────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ORCHESTRATION LAYER                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  NexusEngine    ┌──────────────────────────────────────────────────────┐  │
│  (5-Phase       │         AUTONOMOUS BRAIN v3.0                        │  │
│   Pipeline)     ├──────────┬──────────┬──────────┬────────────────────┤  │
│                 │  Memory   │  Learn   │  Evolve  │  Reason / Predict  │  │
│                 │  Module   │  Module  │  Module  │  Module            │  │
│                 └──────────┴──────────┴──────────┴────────────────────┘  │
│                                    │                                        │
│                 ┌──────────────────────────────────────────────────────┐   │
│                 │      MULTI-AGENT COORDINATOR (10 Agents)             │   │
│                 │  RECON │ VULN │ EXPLOIT │ PERSIST │ DEFENSE │ C2 │   │   │
│                 └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TOOL LAYER (189 Tools)                              │
├──────────────┬──────────────┬──────────────┬──────────────┬───────────────┤
│   Web (12)   │ Network (8)  │  Cloud (6)   │    AD (10)   │ Post-Ex (8)   │
├──────────────┼──────────────┼──────────────┼──────────────┼───────────────┤
│  OSINT (6)   │Wireless (7)  │Hardware (7)  │Telecom (5)   │  ML (3)       │
├──────────────┼──────────────┼──────────────┼──────────────┼───────────────┤
│ Social (3)   │Purple (4)    │Supply (3)    │  C2 (4)      │               │
└──────────────┴──────────────┴──────────────┴──────────────┴───────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                      EXTERNAL TOOL INTEGRATION (47)                         │
│  Nmap │ SQLMap │ Metasploit │ Burp │ Nuclei │ BloodHound │ Impacket │ etc  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Autonomous Brain v3.0

The brain is the central reasoning engine with 5 cognitive modules:

#### Memory Module
- **Short-term memory**: Active session data, current target fingerprint
- **Long-term memory**: Persistent SQLite storage of all historical data
- **Semantic indexing**: Fast recall by subject/predicate/object

#### Learn Module
- **Strategy tracking**: Records success/failure rates per technique
- **Pattern discovery**: Identifies recurring vulnerability patterns
- **Adaptive scoring**: Adjusts technique priority based on past performance

#### Evolve Module
- **Genetic algorithm**: Mutates payloads across generations
- **DNA archive**: Stores evolved knowledge per target
- **Crossover & mutation**: Combines successful payloads

#### Reason Module
- **Rule-based inference**: IF technology=X THEN test=Y
- **Fact knowledge base**: Stores known vulnerability relationships
- **Logical deduction**: Chains facts to derive new attack paths

#### Predict Module
- **Heuristic prediction**: Estimates vulnerability likelihood
- **Feature weighting**: Technologies → vulnerability scores
- **Model training**: Adapts predictions based on outcomes

### 2. Multi-Agent Coordinator

10 specialized agents work in parallel with task delegation:

```
┌────────────────────────────────────────────────────────┐
│                    TASK QUEUE                             │
│  [recon] [vuln_scan] [exploit] [c2] [payload_gen]       │
└────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
      ┌─────┴─────┐  ┌────┴────┐  ┌──────┴──────┐
      │  Agent 1  │  │ Agent 2 │  │   Agent 3   │
      │  RECON    │  │  VULN   │  │   EXPLOIT   │
      └─────┬─────┘  └────┬────┘  └──────┬──────┘
            │               │               │
            └───────────────┴───────────────┘
                            │
              ┌─────────────┴─────────────┐
              │      RESULT COLLECTOR      │
              │   (updates Brain + Brain)  │
              └────────────────────────────┘
```

### 3. 5-Phase Attack Pipeline

```
Phase 1: RECONNAISSANCE
├── Crawl target (all subdomains, endpoints, APIs)
├── Port scan (TCP/UDP, service detection)
├── Technology fingerprint (Wappalyzer, WhatWeb)
├── OSINT gathering (emails, subdomains, GitHub dorks)
└── Target DNA creation (technology + behavior profile)

Phase 2: VULNERABILITY DISCOVERY
├── Predict vulnerabilities from Target DNA
├── Test SQL injection (error-based, blind, time-based)
├── Test XSS (reflected, stored, DOM)
├── Test LFI/RFI (wrapper bypass, log poisoning)
├── Test SSRF (internal services, cloud metadata)
├── Test JWT (none alg, weak secrets, kid injection)
├── Test Cloud misconfigs (S3, IAM, Lambda)
└── Test Network (default creds, known CVEs)

Phase 3: EXPLOITATION
├── Genetic payload evolution (score-based mutation)
├── Multi-vector attack coordination
├── C2 beacon deployment (malleable profile)
├── Privilege escalation (Linux/Windows)
└── Lateral movement (SSH, WMI, PsExec, WinRM)

Phase 4: POST-EXPLOITATION
├── Persistence (cron, systemd, registry, scheduled tasks)
├── Credential dumping (LSASS, SAM, NTDS.dit)
├── Data exfiltration (DNS tunnel, HTTP covert)
├── Keylogging & screenshot capture
└── Token impersonation & pass-the-hash

Phase 5: DEFENSE BYPASS
├── EDR/AV evasion (AMSIsi bypass, process injection)
├── Obfuscation (base64, XOR, custom encoding)
├── Living-off-the-land techniques
├── Anti-forensics (log clearing, timestomping)
└── Sigma/YARA rule generation for detection
```

## Data Flow

```
Target URL
    │
    ▼
┌─────────────┐
│   Brain     │ ← Creates Target DNA, predicts vulnerabilities
│  (Reason)   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│  Agent Coordinator dispatches tasks  │
│  to specialized agents               │
└─────────────────────────────────────┘
       │
   ┌───┴───┬───────┬───────┐
   ▼       ▼       ▼       ▼
RECON   VULN   EXPLOIT  DEFENSE
   │       │       │       │
   └───────┴───┬───┴───────┘
               ▼
        ┌────────────┐
        │   Results   │
        │  (Brain     │
        │   learns)   │
        └────────────┘
               │
               ▼
        ┌────────────┐
        │  Dashboard  │ ← WebSocket real-time streaming
        │   Report    │
        └────────────┘
```

## Backend Architecture

The backend server is a standalone Go module for private cloud deployment:

```
backend/
├── cmd/server/         # Entry point
├── internal/
│   ├── server/         # WebSocket + HTTP API
│   ├── brain/          # Upgraded Brain v3.0
│   └── agents/         # Agent coordinator
├── go.mod              # Standalone module
└── Dockerfile          # Container build
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ws` | WebSocket | Real-time dashboard connection |
| `/api/status` | GET | Server health & stats |
| `/api/agents` | GET | Active agent statuses |
| `/api/sessions` | GET | Active operation sessions |
| `/api/brain/stats` | GET | Brain learning statistics |
| `/api/brain/strategies` | GET | Learned strategies ranked by success |

## Storage Architecture

```
~/.nexus-void/
├── brain/
│   ├── knowledge_graph.db    # SQLite brain database
│   ├── sessions/             # Saved session states
│   ├── target_dna/           # Evolved target profiles
│   ├── exploit_dna/          # Payload evolution archive
│   ├── ai_cache/             # AI response cache
│   └── learned_strategies/   # Persistent strategy data
├── logs/
├── reports/
└── wordlists/
```

## Security Model

- All AI reasoning happens locally (brain database)
- Optional external AI providers (OpenRouter, Groq, Ollama) for payload generation
- C2 communications use AES-GCM encryption
- Malleable profiles mimic legitimate traffic patterns
- Team server supports multi-operator authentication

## Performance

| Metric | Value |
|--------|-------|
| Binary Size | ~20 MB |
| Memory (idle) | ~50 MB |
| Memory (active) | ~200-500 MB |
| Concurrent Agents | 10 |
| Payload Generations/sec | ~1000 |
| WebSocket Clients | Unlimited |
| Database | SQLite (auto-migration) |

## Extension Points

- **New Tools**: Add to `pkg/<domain>/<tool>.go`
- **New Agents**: Add to `backend/internal/agents/coordinator.go`
- **New C2 Profiles**: Add to `pkg/c2/c2.go`
- **New Reason Rules**: Add to `brain.loadInitialKnowledge()`
- **New Dashboard Panels**: Add React components to `dashboard/src/components/`
