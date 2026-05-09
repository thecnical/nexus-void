package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Session represents an engagement session
type Session struct {
	ID          string       `json:"id"`
	Target      string       `json:"target"`
	StartTime   time.Time    `json:"start_time"`
	LastActive  time.Time    `json:"last_active"`
	Status      string       `json:"status"` // running, paused, completed
	Phase       string       `json:"phase"`  // recon, vulnerability, exploit, post-exploit, report
	Findings    int          `json:"findings"`
	Exploits    int          `json:"exploits"`
	Agents      []AgentState `json:"agents"`
	Checkpoints []Checkpoint `json:"checkpoints"`
	Data        SessionData  `json:"data"`
}

// AgentState tracks agent progress
type AgentState struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Output   string `json:"output"`
}

// Checkpoint represents a savable state
type Checkpoint struct {
	Timestamp time.Time `json:"timestamp"`
	Phase     string    `json:"phase"`
	FilePath  string    `json:"file_path"`
}

// SessionData holds all session-specific data
type SessionData struct {
	URLs            []string            `json:"urls"`
	Parameters      map[string][]string `json:"parameters"`
	TechStack       []string            `json:"tech_stack"`
	Vulnerabilities []Vuln              `json:"vulnerabilities"`
	Credentials     []Credential        `json:"credentials"`
	Screenshots     []string            `json:"screenshots"`
}

// Vuln represents a found vulnerability
type Vuln struct {
	Type       string `json:"type"`
	URL        string `json:"url"`
	Parameter  string `json:"parameter"`
	Severity   string `json:"severity"`
	Confidence int    `json:"confidence"`
}

// Credential represents found credentials
type Credential struct {
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Source   string `json:"source"`
}

// Manager handles session persistence
type Manager struct {
	sessionsDir string
	sessions    map[string]*Session
	mu          sync.RWMutex
}

// NewManager creates a session manager
func NewManager(sessionsDir string) *Manager {
	if sessionsDir == "" {
		sessionsDir = filepath.Join(os.Getenv("HOME"), ".nexus-void", "sessions")
	}
	os.MkdirAll(sessionsDir, 0755)

	m := &Manager{
		sessionsDir: sessionsDir,
		sessions:    make(map[string]*Session),
	}

	// Load existing sessions
	m.loadAll()
	return m
}

// Create starts a new session
func (m *Manager) Create(target string) *Session {
	s := &Session{
		ID:          fmt.Sprintf("SES-%d", time.Now().UnixNano()),
		Target:      target,
		StartTime:   time.Now(),
		LastActive:  time.Now(),
		Status:      "running",
		Phase:       "recon",
		Agents:      []AgentState{},
		Checkpoints: []Checkpoint{},
		Data: SessionData{
			URLs:            []string{},
			Parameters:      make(map[string][]string),
			TechStack:       []string{},
			Vulnerabilities: []Vuln{},
			Credentials:     []Credential{},
			Screenshots:     []string{},
		},
	}

	m.mu.Lock()
	m.sessions[s.ID] = s
	m.mu.Unlock()

	m.save(s)
	return s
}

// Get retrieves a session by ID
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	return s, ok
}

// GetActive returns the most recent active session
func (m *Manager) GetActive() *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var active *Session
	for _, s := range m.sessions {
		if s.Status == "running" {
			if active == nil || s.LastActive.After(active.LastActive) {
				active = s
			}
		}
	}
	return active
}

// List returns all sessions
func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*Session
	for _, s := range m.sessions {
		list = append(list, s)
	}
	return list
}

// Update saves session state
func (m *Manager) Update(s *Session) error {
	s.LastActive = time.Now()
	m.mu.Lock()
	m.sessions[s.ID] = s
	m.mu.Unlock()
	return m.save(s)
}

// Checkpoint creates a recovery point
func (m *Manager) Checkpoint(s *Session, phase string) error {
	cp := Checkpoint{
		Timestamp: time.Now(),
		Phase:     phase,
		FilePath:  filepath.Join(m.sessionsDir, fmt.Sprintf("%s_%s_%d.json", s.ID, phase, time.Now().Unix())),
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(cp.FilePath, data, 0644); err != nil {
		return err
	}

	s.Checkpoints = append(s.Checkpoints, cp)
	return m.Update(s)
}

// Restore recovers a session from checkpoint
func (m *Manager) Restore(sessionID string, checkpointIdx int) (*Session, error) {
	s, ok := m.Get(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if checkpointIdx >= len(s.Checkpoints) {
		return nil, fmt.Errorf("checkpoint index out of range")
	}

	cp := s.Checkpoints[checkpointIdx]
	data, err := os.ReadFile(cp.FilePath)
	if err != nil {
		return nil, fmt.Errorf("checkpoint file not found: %w", err)
	}

	var restored Session
	if err := json.Unmarshal(data, &restored); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	restored.Status = "running"
	restored.LastActive = time.Now()

	m.mu.Lock()
	m.sessions[restored.ID] = &restored
	m.mu.Unlock()

	m.save(&restored)
	return &restored, nil
}

// Pause marks session as paused
func (m *Manager) Pause(id string) error {
	s, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("session not found")
	}
	s.Status = "paused"
	m.Checkpoint(s, s.Phase)
	return m.Update(s)
}

// Resume marks session as running
func (m *Manager) Resume(id string) error {
	s, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("session not found")
	}
	s.Status = "running"
	return m.Update(s)
}

// Complete marks session as finished
func (m *Manager) Complete(id string) error {
	s, ok := m.Get(id)
	if !ok {
		return fmt.Errorf("session not found")
	}
	s.Status = "completed"
	return m.Update(s)
}

// Delete removes a session
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()

	// Remove session file
	path := filepath.Join(m.sessionsDir, fmt.Sprintf("%s.json", id))
	os.Remove(path)

	// Remove checkpoints
	pattern := filepath.Join(m.sessionsDir, fmt.Sprintf("%s_*", id))
	matches, _ := filepath.Glob(pattern)
	for _, match := range matches {
		os.Remove(match)
	}

	return nil
}

// save persists session to disk
func (m *Manager) save(s *Session) error {
	path := filepath.Join(m.sessionsDir, fmt.Sprintf("%s.json", s.ID))
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// loadAll restores sessions from disk
func (m *Manager) loadAll() {
	files, err := os.ReadDir(m.sessionsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.sessionsDir, file.Name()))
		if err != nil {
			continue
		}

		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}

		m.sessions[s.ID] = &s
	}
}

// AddAgent registers an agent to a session
func (s *Session) AddAgent(name string) {
	for i, a := range s.Agents {
		if a.Name == name {
			s.Agents[i].Status = "running"
			return
		}
	}
	s.Agents = append(s.Agents, AgentState{Name: name, Status: "running", Progress: 0})
}

// UpdateAgent updates agent progress
func (s *Session) UpdateAgent(name string, progress int, output string) {
	for i, a := range s.Agents {
		if a.Name == name {
			s.Agents[i].Progress = progress
			if output != "" {
				s.Agents[i].Output = output
			}
			return
		}
	}
}

// AddVulnerability adds a found vulnerability
func (s *Session) AddVulnerability(v Vuln) {
	s.Data.Vulnerabilities = append(s.Data.Vulnerabilities, v)
	s.Findings++
}

// AddURL adds a discovered URL
func (s *Session) AddURL(url string) {
	for _, u := range s.Data.URLs {
		if u == url {
			return
		}
	}
	s.Data.URLs = append(s.Data.URLs, url)
}

// AddCredential adds discovered credentials
func (s *Session) AddCredential(c Credential) {
	s.Data.Credentials = append(s.Data.Credentials, c)
}
