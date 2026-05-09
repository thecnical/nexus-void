# NEXUS-VOID Autonomous Brain v3.0

## Overview

The Brain is the central cognitive engine of NEXUS-VOID OMEGA. It learns from every session, reasons about targets, predicts vulnerabilities, and evolves attack strategies using genetic algorithms.

## Cognitive Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     BRAIN v3.0                              │
├──────────────┬──────────────┬──────────────┬──────────────┤
│   MEMORY     │    LEARN     │    EVOLVE    │   REASON     │
│   MODULE     │   MODULE     │   MODULE     │   /PREDICT   │
├──────────────┼──────────────┼──────────────┼──────────────┤
│ Short-term   │ Strategies   │ Generations  │ Facts        │
│ Long-term    │ Patterns     │ DNA Archive  │ Rules        │
│ Cache        │ Success Map  │ Mutations    │ Models       │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

## Memory Module

### Short-Term Memory
- Active session data (current target, phase, findings)
- Agent status snapshots
- Pending task queue state
- Real-time exploit results

### Long-Term Memory (SQLite)
- Historical strategies with success rates
- Target DNA profiles
- Payload evolution archive
- Fact knowledge base
- Session checkpoints

### Storage Schema
```sql
CREATE TABLE strategies (
    id TEXT PRIMARY KEY,
    domain TEXT,
    technique TEXT,
    success_rate REAL,
    use_count INTEGER,
    created_at DATETIME
);

CREATE TABLE patterns (
    id TEXT PRIMARY KEY,
    type TEXT,
    signature TEXT,
    confidence REAL,
    occurrences INTEGER
);

CREATE TABLE generations (
    id TEXT PRIMARY KEY,
    payload TEXT,
    score INTEGER,
    target_type TEXT,
    mutation_type TEXT,
    created_at DATETIME
);
```

## Learn Module

### Strategy Learning
The brain tracks every technique's performance:

```
Strategy: web:sqli
├── Success Rate: 0.85
├── Use Count: 47
├── Avg Time: 1200ms
└── Created: 2024-01-15

Strategy: ad:kerberoast
├── Success Rate: 0.92
├── Use Count: 23
├── Avg Time: 4500ms
└── Created: 2024-02-01
```

### Adaptive Prioritization
```
When attacking a WordPress target:
1. Try SQL injection (success rate: 0.85) ← FIRST
2. Try XSS (success rate: 0.72)         ← SECOND
3. Try LFI (success rate: 0.68)        ← THIRD
```

### Pattern Discovery
The brain identifies patterns like:
- "WordPress + Plugin X → SQL injection"
- "Apache + mod_php → path traversal"
- "IIS + ASP.NET → deserialization"

## Evolve Module

### Genetic Algorithm

```
Generation 0:  ' OR 1=1 --
       ↓ mutate
Generation 1:  ' OR 1=1 /*
       ↓ mutate
Generation 2:  %2527 OR 1=1-- -
       ↓ mutate
Generation 3:  /*!50000'*/ OR '1'='1'-- -
       ↓ mutate
Generation 4:  0x27204f5220313d312d2d (hex encoded)
```

### Scoring System
| Factor | Points |
|--------|--------|
| Base score | 50 |
| Length > 20 chars | +5 |
| Length > 50 chars | +5 |
| Encoding diversity | +10 |
| Technology match | +10 |
| Past success | +10 |
| Too short (<5) | -20 |
| **Maximum** | **100** |

### Mutation Types
- **Expansion**: Add comments, null bytes, extra characters
- **Contraction**: Remove unnecessary parts
- **Substitution**: Replace characters with encoded variants

## Reason Module

### Knowledge Base
Pre-loaded facts about vulnerability relationships:

```
Fact: WordPress → has → plugin_sqli (confidence: 0.85)
Fact: Java → has → deserialization_rce (confidence: 0.90)
Fact: Node.js → has → prototype_pollution (confidence: 0.70)
Fact: PHP → has → lfi_wrapper (confidence: 0.80)
```

### Inference Rules
```
IF technology contains "WordPress" THEN test_sqli_plugin
IF technology contains "Apache" THEN test_path_traversal
IF technology contains "Java" THEN test_spring_rce
IF technology contains "Node.js" THEN test_prototype_pollution
IF technology contains "PHP" THEN test_lfi_wrapper
IF technology contains ".NET" THEN test_deserialization
IF technology contains "Cloud" THEN test_ssrf_metadata
```

### Reasoning Example
```
Input: Target runs WordPress + WooCommerce + MySQL
Facts matched:
  - WordPress → plugin_sqli (0.85)
  - PHP → lfi_wrapper (0.80)
Inferences:
  → Test WooCommerce SQL injection
  → Test WordPress plugin LFI
  → Test theme file inclusion
```

## Predict Module

### Heuristic Prediction
Predicts vulnerability likelihood based on technology stack:

```
Target: WordPress + MySQL + Apache + PHP

SQLi prediction:    0.70 (PHP + MySQL)
XSS prediction:     0.55 (WordPress plugins)
RCE prediction:     0.35 (no Java/Node.js)
SSRF prediction:    0.45 (PHP + Cloud)
LFI prediction:     0.80 (PHP - high confidence)
```

### Technology → Vulnerability Mapping
```
PHP:        SQLi +0.4, LFI +0.3, RCE +0.1
Java:       RCE +0.4, Deserialization +0.3
Node.js:    Prototype Pollution +0.3, RCE +0.2
Python:     SSTI +0.3, RCE +0.2
.NET:       Deserialization +0.4, SQLi +0.1
WordPress:  SQLi +0.3, XSS +0.2
```

### Model Training
The prediction models improve over time:
```
Model: sqli_predictor
├── Accuracy: 0.82
├── Trained on: 156 targets
├── Last updated: 2024-03-15
└── Features: [php, mysql, wordpress, asp, mssql]
```

## Target DNA

### Structure
```json
{
  "target": "target.com",
  "fingerprint": "sha256_hash",
  "technologies": ["WordPress", "PHP", "MySQL", "Apache"],
  "behaviors": ["redirects_to_https", "rate_limited", "waf_detected"],
  "weaknesses": ["sqli", "lfi", "xss"],
  "successful_paths": ["sqli_plugin", "lfi_wrapper"],
  "failed_paths": ["rce_spring", "ssrf_metadata"],
  "confidence_map": {
    "sqli": 0.85,
    "xss": 0.72,
    "lfi": 0.68
  },
  "last_updated": "2024-03-15T10:30:00Z"
}
```

### DNA Evolution
Target DNA is updated after every interaction:
```
Session 1: Discovers WordPress → DNA += {technologies: [WordPress]}
Session 2: Finds SQLi → DNA += {weaknesses: [sqli], successful_paths: [sqli_plugin]}
Session 3: Fails RCE → DNA += {failed_paths: [rce_php]}
```

## Session Management

### Session Lifecycle
```
CREATE → RUNNING → [PAUSED] → COMPLETED
   │         │         │          │
   ↓         ↓         ↓          ↓
RECON    EXPLOIT   SAVE      REPORT
```

### Checkpoint System
Sessions automatically checkpoint every 30 seconds:
```
checkpoint_001: {phase: "recon", progress: 45%}
checkpoint_002: {phase: "vuln_scan", progress: 60%}
checkpoint_003: {phase: "exploit", progress: 20%}
```

## Statistics

The brain exports comprehensive learning statistics:

```json
{
  "strategies_learned": 142,
  "patterns_discovered": 89,
  "generations_created": 3450,
  "targets_profiled": 67,
  "active_sessions": 3,
  "facts_known": 156,
  "models_trained": 8
}
```

## API

### Initialize
```go
brain, err := brain.Initialize("~/.nexus-void")
```

### Record Success
```go
brain.Learn.RecordSuccess("sqli", "web", 1200.0)
```

### Record Failure
```go
brain.Learn.RecordFailure("rce", "web")
```

### Get Best Strategy
```go
strategy := brain.Learn.GetBestStrategy("web")
// Returns strategy with highest success rate for web domain
```

### Evolve Payload
```go
best := brain.Evolve.EvolvePayload(
    "' OR 1=1 --",
    "sqli",
    30, // max generations
)
// Returns best scoring payload after evolution
```

### Predict Vulnerability
```go
dna := brain.GetTargetDNA("target.com")
likelihood := brain.Predict.PredictVulnerability(dna, "sqli")
// Returns 0.0-1.0 probability
```

### Update Target DNA
```go
brain.UpdateTargetDNA("target.com", &brain.TargetDNA{
    Technologies: []string{"WordPress", "PHP"},
    Weaknesses:   []string{"sqli"},
})
```

## Performance

| Operation | Time |
|-----------|------|
| Brain initialization | ~200ms |
| Strategy lookup | ~1ms |
| Payload evolution (30 gen) | ~50ms |
| Prediction | ~2ms |
| DNA save | ~10ms |
| SQLite query | ~5ms |
