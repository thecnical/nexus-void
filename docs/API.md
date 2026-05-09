# NEXUS-VOID API Reference

## Base URL

```
Local:  http://localhost:8080
Render: https://your-app.onrender.com
```

## WebSocket API

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?operator=username');
```

### Incoming Messages

#### Welcome
```json
{
  "type": "welcome",
  "timestamp": "2024-03-15T10:30:00Z",
  "data": {
    "version": "2.0",
    "agents": 10,
    "clients": 3,
    "message": "Connected to NEXUS-VOID OMEGA Backend"
  }
}
```

#### Agent Update
```json
{
  "type": "agent_update",
  "timestamp": "2024-03-15T10:30:05Z",
  "data": {
    "agent": "RECON-OMEGA",
    "status": "scanning",
    "progress": 45
  }
}
```

#### Finding
```json
{
  "type": "finding",
  "timestamp": "2024-03-15T10:31:00Z",
  "data": {
    "type": "sqli",
    "proof": "id=1 AND 1=1 --",
    "severity": "critical"
  }
}
```

#### Session Started
```json
{
  "type": "session_started",
  "timestamp": "2024-03-15T10:30:00Z",
  "data": {
    "session_id": "sess_abc123",
    "target": "target.com",
    "operator": "admin"
  }
}
```

### Outgoing Commands

#### Start Session
```json
{
  "type": "start_session",
  "target": "target.com"
}
```

#### Stop Session
```json
{
  "type": "stop_session",
  "session_id": "sess_abc123"
}
```

#### Dispatch Task
```json
{
  "type": "dispatch_task",
  "task_type": "recon",
  "target": "target.com",
  "priority": 10
}
```

## REST API

### GET /api/status

Server health and status.

**Response:**
```json
{
  "version": "2.0",
  "clients": 3,
  "agents": 10,
  "uptime": "running",
  "status": "operational"
}
```

### GET /api/agents

List all agent statuses.

**Response:**
```json
[
  {
    "id": "RECON-OMEGA-12345",
    "type": "RECON-OMEGA",
    "status": "idle",
    "progress": 0,
    "output": [],
    "confidence": 1.0,
    "errors": 0
  },
  {
    "id": "VULN-SENTINEL-67890",
    "type": "VULN-SENTINEL",
    "status": "scanning",
    "progress": 45,
    "output": ["[10:30:00] Starting scan..."],
    "confidence": 0.85,
    "errors": 0
  }
]
```

### GET /api/sessions

List active sessions.

**Response:**
```json
{
  "sess_abc123": {
    "id": "sess_abc123",
    "target": "target.com",
    "start_time": "2024-03-15T10:30:00Z",
    "status": "running",
    "findings": 5,
    "exploits": 2,
    "agents": ["RECON-OMEGA", "VULN-SENTINEL"]
  }
}
```

### GET /api/brain/stats

Brain learning statistics.

**Response:**
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

### GET /api/brain/strategies

Learned strategies ranked by success rate.

**Response:**
```json
[
  {
    "id": "ad:kerberoast",
    "domain": "ad",
    "technique": "kerberoast",
    "success_rate": 0.92,
    "use_count": 23,
    "avg_time_ms": 4500,
    "created_at": "2024-02-01T00:00:00Z",
    "updated_at": "2024-03-10T00:00:00Z"
  },
  {
    "id": "web:sqli",
    "domain": "web",
    "technique": "sqli",
    "success_rate": 0.85,
    "use_count": 47,
    "avg_time_ms": 1200,
    "created_at": "2024-01-15T00:00:00Z",
    "updated_at": "2024-03-12T00:00:00Z"
  }
]
```

## CLI API

### Initialize
```bash
nexus-void init
```

### Reconnaissance
```bash
nexus-void recon https://target.com
```

### Vulnerability Hunt
```bash
nexus-void hunt https://target.com
```

### Full Autonomous Attack
```bash
nexus-void apocalypse https://target.com
```

### Arsenal List
```bash
nexus-void arsenal list
```

### Doctor (Self-Test)
```bash
nexus-void doctor
```

### Report
```bash
nexus-void report
```

## Brain Go API

### Initialize Brain
```go
import "github.com/nexus-void/nexus-void/pkg/brain"

b, err := brain.Initialize("~/.nexus-void")
if err != nil {
    log.Fatal(err)
}
defer b.Close()
```

### Record Success/Failure
```go
// Record a successful technique
b.Learn.RecordSuccess("sqli", "web", 1200.0)

// Record a failed technique
b.Learn.RecordFailure("rce", "web")
```

### Get Best Strategy
```go
best := b.Learn.GetBestStrategy("web")
if best != nil {
    fmt.Printf("Best technique: %s (%.0f%% success)\n", 
        best.Technique, best.SuccessRate*100)
}
```

### Evolve Payload
```go
best := b.Evolve.EvolvePayload("' OR 1=1 --", "sqli", 30)
fmt.Printf("Best payload: %s (score: %d)\n", best.Payload, best.Score)
```

### Predict Vulnerability
```go
dna := b.GetTargetDNA("target.com")
if dna == nil {
    dna = &brain.TargetDNA{Target: "target.com"}
}

sqliProb := b.Predict.PredictVulnerability(dna, "sqli")
fmt.Printf("SQLi likelihood: %.0f%%\n", sqliProb*100)
```

### Create Session
```go
session := b.NewSession("target.com")
fmt.Printf("Session ID: %s\n", session.ID)
```

### Update Session
```go
b.UpdateSession(session.ID, map[string]interface{}{
    "findings": 5,
    "exploits": 2,
    "status": "completed",
})
```

## Agent Coordinator API

### Dispatch Task
```go
import "github.com/nexus-void/nexus-void/backend/internal/agents"

coord := agents.NewCoordinator(brain)
taskID := coord.DispatchTask("recon", "target.com", 10)
fmt.Printf("Task dispatched: %s\n", taskID)
```

### Get Agent Status
```go
for _, agent := range coord.GetAgentStatus() {
    fmt.Printf("%s: %s (%d%%)\n", 
        agent.Type, agent.Status, agent.Progress)
}
```

## Error Codes

| Code | Meaning |
|------|---------|
| 400 | Bad Request |
| 404 | Not Found |
| 500 | Internal Server Error |
| 1001 | Brain Database Error |
| 1002 | Agent Not Found |
| 1003 | Task Queue Full |
| 1004 | Invalid Payload |
| 1005 | Evolution Failed |

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| WebSocket | 60 messages/minute |
| REST API | 100 requests/minute |
| Task Dispatch | 10 tasks/second |
| Payload Evolution | 1000/sec |
