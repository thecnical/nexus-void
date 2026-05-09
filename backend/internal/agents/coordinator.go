package agents

import (
	"fmt"
	"sync"
	"time"

	"github.com/nexus-void/backend/internal/brain"
)

// AgentType represents different AI agent specializations
type AgentType string

const (
	AgentRecon       AgentType = "RECON-OMEGA"
	AgentVuln        AgentType = "VULN-SENTINEL"
	AgentExploit     AgentType = "EXPLOIT-APOCALYPSE"
	AgentPostExploit AgentType = "PERSISTENCE-DAEMON"
	AgentDefense     AgentType = "SHIELD-BREAKER"
	AgentC2          AgentType = "C2-NEXUS"
	AgentML          AgentType = "NEURAL-PHOENIX"
	AgentOSINT       AgentType = "OSINT-PHANTOM"
	AgentSocial      AgentType = "SOCIAL-MANIPULATOR"
	AgentPurple      AgentType = "PURPLE-GUARD"
)

// Agent represents an autonomous AI agent
type Agent struct {
	ID          string      `json:"id"`
	Type        AgentType   `json:"type"`
	Status      AgentStatus `json:"status"`
	Progress    int         `json:"progress"`
	Output      []string    `json:"output"`
	CurrentTask *Task       `json:"current_task,omitempty"`
	StartTime   time.Time   `json:"start_time"`
	LastUpdate  time.Time   `json:"last_update"`
	Confidence  float64     `json:"confidence"`
	Errors      int         `json:"errors"`
	mu          sync.Mutex  `json:"-"`
}

// AgentStatus represents the current state of an agent
type AgentStatus string

const (
	StatusIdle       AgentStatus = "idle"
	StatusPlanning   AgentStatus = "planning"
	StatusScanning   AgentStatus = "scanning"
	StatusAnalyzing  AgentStatus = "analyzing"
	StatusExploiting AgentStatus = "exploiting"
	StatusEvolving   AgentStatus = "evolving"
	StatusReporting  AgentStatus = "reporting"
	StatusComplete   AgentStatus = "complete"
	StatusError      AgentStatus = "error"
	StatusRecovering AgentStatus = "recovering"
)

// Task represents a unit of work assigned to an agent
type Task struct {
	ID           string      `json:"id"`
	Type         string      `json:"type"`
	Target       string      `json:"target"`
	Payload      string      `json:"payload"`
	Priority     int         `json:"priority"`
	Status       string      `json:"status"`
	Result       interface{} `json:"result,omitempty"`
	StartedAt    time.Time   `json:"started_at"`
	CompletedAt  *time.Time  `json:"completed_at,omitempty"`
	Dependencies []string    `json:"dependencies"`
}

// Coordinator manages all agents and orchestrates operations
type Coordinator struct {
	brain     *brain.Brain
	agents    map[string]*Agent
	taskQueue chan *Task
	results   chan *TaskResult
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// TaskResult represents the outcome of a task
type TaskResult struct {
	TaskID    string      `json:"task_id"`
	AgentID   string      `json:"agent_id"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewCoordinator creates a new agent coordinator
func NewCoordinator(b *brain.Brain) *Coordinator {
	c := &Coordinator{
		brain:     b,
		agents:    make(map[string]*Agent),
		taskQueue: make(chan *Task, 100),
		results:   make(chan *TaskResult, 100),
		stopChan:  make(chan struct{}),
	}

	// Initialize default agents
	c.spawnAgent(AgentRecon)
	c.spawnAgent(AgentVuln)
	c.spawnAgent(AgentExploit)
	c.spawnAgent(AgentPostExploit)
	c.spawnAgent(AgentDefense)
	c.spawnAgent(AgentC2)
	c.spawnAgent(AgentML)
	c.spawnAgent(AgentOSINT)
	c.spawnAgent(AgentSocial)
	c.spawnAgent(AgentPurple)

	// Start the orchestrator
	go c.orchestrator()
	go c.resultCollector()

	return c
}

// spawnAgent creates and registers a new agent
func (c *Coordinator) spawnAgent(agentType AgentType) *Agent {
	agent := &Agent{
		ID:         fmt.Sprintf("%s-%d", agentType, time.Now().UnixNano()),
		Type:       agentType,
		Status:     StatusIdle,
		Output:     make([]string, 0),
		StartTime:  time.Now(),
		Confidence: 1.0,
	}

	c.mu.Lock()
	c.agents[agent.ID] = agent
	c.mu.Unlock()

	// Start the agent's worker
	go c.agentWorker(agent)

	return agent
}

// agentWorker is the main loop for each agent
func (c *Coordinator) agentWorker(agent *Agent) {
	for {
		select {
		case <-c.stopChan:
			return
		case task := <-c.taskQueue:
			// Check if this agent can handle this task
			if c.canHandle(agent, task) {
				c.executeTask(agent, task)
			} else {
				// Put back in queue for another agent
				go func(t *Task) {
					time.Sleep(100 * time.Millisecond)
					c.taskQueue <- t
				}(task)
			}

		default:
			if agent.Status != StatusIdle && agent.Status != StatusError {
				// Agent is busy, continue working
				c.simulateWork(agent)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// canHandle checks if an agent can handle a specific task type
func (c *Coordinator) canHandle(agent *Agent, task *Task) bool {
	switch task.Type {
	case "recon", "crawl", "portscan", "subdomain":
		return agent.Type == AgentRecon || agent.Type == AgentOSINT
	case "vuln_scan", "sqli", "xss", "lfi", "ssrf", "jwt":
		return agent.Type == AgentVuln || agent.Type == AgentML
	case "exploit", "rce", "privesc":
		return agent.Type == AgentExploit
	case "persistence", "lateral", "exfil":
		return agent.Type == AgentPostExploit
	case "defense_evasion", "sigma", "yara":
		return agent.Type == AgentDefense || agent.Type == AgentPurple
	case "c2", "beacon", "implant":
		return agent.Type == AgentC2
	case "payload_gen", "evolve", "mutate":
		return agent.Type == AgentML
	case "osint", "social", "phishing":
		return agent.Type == AgentOSINT || agent.Type == AgentSocial
	default:
		return true
	}
}

// executeTask runs a task on an agent
func (c *Coordinator) executeTask(agent *Agent, task *Task) {
	agent.mu.Lock()
	agent.Status = StatusAnalyzing
	agent.CurrentTask = task
	agent.LastUpdate = time.Now()
	agent.mu.Unlock()

	c.log(agent, fmt.Sprintf("[+] Starting task %s: %s on %s", task.ID, task.Type, task.Target))

	// Simulate work phases
	phases := []AgentStatus{StatusScanning, StatusAnalyzing, StatusExploiting, StatusReporting}
	for _, phase := range phases {
		agent.mu.Lock()
		agent.Status = phase
		agent.Progress += 25
		agent.mu.Unlock()

		c.log(agent, fmt.Sprintf("[*] Phase: %s (%d%%)", phase, agent.Progress))
		time.Sleep(500 * time.Millisecond)
	}

	// Generate result
	result := &TaskResult{
		TaskID:    task.ID,
		AgentID:   agent.ID,
		Success:   true,
		Data:      map[string]string{"finding": fmt.Sprintf("%s completed on %s", task.Type, task.Target)},
		Timestamp: time.Now(),
	}

	c.results <- result

	agent.mu.Lock()
	agent.Status = StatusIdle
	agent.Progress = 0
	agent.CurrentTask = nil
	agent.LastUpdate = time.Now()
	agent.mu.Unlock()

	c.log(agent, fmt.Sprintf("[+] Task %s completed", task.ID))
}

// simulateWork simulates background agent activity
func (c *Coordinator) simulateWork(agent *Agent) {
	// Background learning/reflection
	if time.Since(agent.LastUpdate) > 10*time.Second {
		agent.mu.Lock()
		agent.Status = StatusIdle
		agent.Progress = 0
		agent.mu.Unlock()
	}
}

// orchestrator coordinates multi-agent operations
func (c *Coordinator) orchestrator() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.coordinateAgents()
		}
	}
}

// coordinateAgents manages agent collaboration and task delegation
func (c *Coordinator) coordinateAgents() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check for idle agents and assign learning tasks
	for _, agent := range c.agents {
		if agent.Status == StatusIdle {
			// Assign background learning or idle optimization
			c.brain.Learn.GetBestStrategy(string(agent.Type))
		}
	}
}

// resultCollector processes task results
func (c *Coordinator) resultCollector() {
	for {
		select {
		case <-c.stopChan:
			return
		case result := <-c.results:
			if result.Success {
				// Update brain with success
				c.brain.Learn.RecordSuccess("generic", "generic", 1000)
			} else {
				c.brain.Learn.RecordFailure("generic", "generic")
			}
		}
	}
}

// DispatchTask adds a task to the queue
func (c *Coordinator) DispatchTask(taskType, target string, priority int) string {
	task := &Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Type:      taskType,
		Target:    target,
		Priority:  priority,
		Status:    "pending",
		StartedAt: time.Now(),
	}

	c.taskQueue <- task
	return task.ID
}

// GetAgentStatus returns all agent statuses
func (c *Coordinator) GetAgentStatus() []*Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agents := make([]*Agent, 0, len(c.agents))
	for _, agent := range c.agents {
		agents = append(agents, agent)
	}
	return agents
}

// AgentCount returns the number of active agents
func (c *Coordinator) AgentCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.agents)
}

// Stop shuts down all agents
func (c *Coordinator) Stop() {
	close(c.stopChan)
}

// log adds a log entry to an agent
func (c *Coordinator) log(agent *Agent, msg string) {
	agent.mu.Lock()
	agent.Output = append(agent.Output, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	if len(agent.Output) > 100 {
		agent.Output = agent.Output[len(agent.Output)-100:]
	}
	agent.mu.Unlock()
}
