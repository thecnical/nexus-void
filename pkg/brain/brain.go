package brain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	BrainDir        = ".nexus-void/brain"
	KnowledgeDB     = "knowledge_graph.db"
	SessionsDir     = "sessions"
	TargetDNADir    = "target_dna"
	ExploitDNADir   = "exploit_dna"
	AICacheDir      = "ai_cache"
	StrategiesDir   = "learned_strategies"
	MaxGenerations  = 30
	EvolveThreshold = 0.95
)

type Brain struct {
	DB       *gorm.DB
	homeDir  string
	mu       sync.RWMutex
	Memory   *MemoryModule
	Learn    *LearnModule
	Evolve   *EvolveModule
	sessions map[string]*Session
}

// MemoryModule stores all historical data
type MemoryModule struct {
	brain *Brain
	cache map[string]interface{}
	mu    sync.RWMutex
}

// LearnModule recognizes patterns and adapts strategies
type LearnModule struct {
	brain      *Brain
	strategies map[string]*Strategy
	mu         sync.RWMutex
}

// EvolveModule uses genetic algorithms to mutate payloads
type EvolveModule struct {
	brain       *Brain
	generations map[string][]*Generation
	mu          sync.RWMutex
}

type Strategy struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`    // web, network, cloud, etc.
	Technique   string    `json:"technique"` // sqli, xss, bruteforce, etc.
	Condition   string    `json:"condition"` // When to apply
	Action      string    `json:"action"`    // What to do
	SuccessRate float64   `json:"success_rate"`
	UseCount    int       `json:"use_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Generation struct {
	ID           string    `json:"id"`
	ParentIDs    []string  `json:"parent_ids" gorm:"serializer:json"`
	Payload      string    `json:"payload"`
	Score        int       `json:"score"` // 0-100
	TargetType   string    `json:"target_type"`
	MutationType string    `json:"mutation_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID          string       `json:"id"`
	Target      string       `json:"target"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     *time.Time   `json:"end_time,omitempty"`
	Status      string       `json:"status"` // running, paused, completed, crashed
	Findings    int          `json:"findings"`
	Exploits    int          `json:"exploits"`
	Agents      []string     `json:"agents" gorm:"serializer:json"`
	Checkpoints []Checkpoint `json:"checkpoints" gorm:"serializer:json"`
}

type Checkpoint struct {
	Timestamp time.Time `json:"timestamp"`
	State     string    `json:"state"`
	FilePath  string    `json:"file_path"`
}

// TargetDNA stores evolved knowledge about a specific target
type TargetDNA struct {
	Target             string            `json:"target"`
	TechStack          []string          `json:"tech_stack" gorm:"serializer:json"`
	WAFType            string            `json:"waf_type"`
	WAFSignatures      []string          `json:"waf_signatures" gorm:"serializer:json"`
	SuccessfulPayloads []PayloadDNA      `json:"successful_payloads" gorm:"serializer:json"`
	FailedPayloads     []PayloadDNA      `json:"failed_payloads" gorm:"serializer:json"`
	TimingProfile      map[string]int    `json:"timing_profile" gorm:"serializer:json"`
	ResponsePatterns   map[string]string `json:"response_patterns" gorm:"serializer:json"`
	LastUpdated        time.Time         `json:"last_updated"`
}

type PayloadDNA struct {
	Payload        string    `json:"payload"`
	Technique      string    `json:"technique"`
	InjectionPoint string    `json:"injection_point"`
	Result         string    `json:"result"`
	Timestamp      time.Time `json:"timestamp"`
}

func NewBrain() (*Brain, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	brainDir := filepath.Join(home, BrainDir)
	for _, dir := range []string{
		brainDir,
		filepath.Join(brainDir, SessionsDir),
		filepath.Join(brainDir, TargetDNADir),
		filepath.Join(brainDir, ExploitDNADir),
		filepath.Join(brainDir, AICacheDir),
		filepath.Join(brainDir, StrategiesDir),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	dbPath := filepath.Join(brainDir, KnowledgeDB)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate schema
	if err := db.AutoMigrate(
		&Node{},
		&Edge{},
		&Finding{},
		&Exploit{},
		&Defense{},
		&AgentEvent{},
		&Strategy{},
		&Generation{},
		&Session{},
		&TargetDNA{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	b := &Brain{
		DB:       db,
		homeDir:  brainDir,
		sessions: make(map[string]*Session),
	}

	b.Memory = &MemoryModule{brain: b, cache: make(map[string]interface{})}
	b.Learn = &LearnModule{brain: b, strategies: make(map[string]*Strategy)}
	b.Evolve = &EvolveModule{brain: b, generations: make(map[string][]*Generation)}

	// Load existing strategies
	b.loadStrategies()
	b.loadTargetDNA()

	return b, nil
}

func (b *Brain) Close() error {
	sqlDB, err := b.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (b *Brain) SaveSession(session *Session) error {
	b.mu.Lock()
	b.sessions[session.ID] = session
	b.mu.Unlock()

	result := b.DB.Create(session)
	return result.Error
}

func (b *Brain) GetSession(id string) (*Session, error) {
	var session Session
	result := b.DB.First(&session, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &session, nil
}

func (b *Brain) CreateCheckpoint(sessionID string, state string) error {
	checkpoint := Checkpoint{
		Timestamp: time.Now(),
		State:     state,
		FilePath:  filepath.Join(b.homeDir, SessionsDir, fmt.Sprintf("%s_%d.json", sessionID, time.Now().Unix())),
	}

	// Save state to file
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(checkpoint.FilePath, data, 0644); err != nil {
		return err
	}

	// Update session
	return b.DB.Exec(
		"UPDATE sessions SET checkpoints = json_append(checkpoints, ?) WHERE id = ?",
		checkpoint, sessionID,
	).Error
}

func (b *Brain) loadStrategies() {
	strategiesDir := filepath.Join(b.homeDir, StrategiesDir)
	files, err := os.ReadDir(strategiesDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(strategiesDir, file.Name()))
		if err != nil {
			continue
		}

		var strategy Strategy
		if err := json.Unmarshal(data, &strategy); err != nil {
			continue
		}

		b.Learn.strategies[strategy.ID] = &strategy
	}
}

func (b *Brain) loadTargetDNA() {
	dnaDir := filepath.Join(b.homeDir, TargetDNADir)
	// DNA is loaded on-demand when a target is accessed
	_ = dnaDir
}

func (b *Brain) GetTargetDNA(target string) (*TargetDNA, error) {
	dnaPath := filepath.Join(b.homeDir, TargetDNADir, fmt.Sprintf("%s.json", sanitizeFilename(target)))

	data, err := os.ReadFile(dnaPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty DNA for new target
			return &TargetDNA{
				Target:             target,
				TechStack:          []string{},
				SuccessfulPayloads: []PayloadDNA{},
				FailedPayloads:     []PayloadDNA{},
				TimingProfile:      make(map[string]int),
				ResponsePatterns:   make(map[string]string),
				LastUpdated:        time.Now(),
			}, nil
		}
		return nil, err
	}

	var dna TargetDNA
	if err := json.Unmarshal(data, &dna); err != nil {
		return nil, err
	}

	return &dna, nil
}

func (b *Brain) SaveTargetDNA(dna *TargetDNA) error {
	dna.LastUpdated = time.Now()

	data, err := json.MarshalIndent(dna, "", "  ")
	if err != nil {
		return err
	}

	dnaPath := filepath.Join(b.homeDir, TargetDNADir, fmt.Sprintf("%s.json", sanitizeFilename(dna.Target)))
	return os.WriteFile(dnaPath, data, 0644)
}

func (b *Brain) GetStrategy(domain, technique string) *Strategy {
	b.Learn.mu.RLock()
	defer b.Learn.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", domain, technique)
	if strategy, ok := b.Learn.strategies[key]; ok {
		return strategy
	}

	// Return default strategy
	return &Strategy{
		ID:          key,
		Domain:      domain,
		Technique:   technique,
		Condition:   "default",
		Action:      "standard_approach",
		SuccessRate: 0.5,
	}
}

func (b *Brain) SaveStrategy(strategy *Strategy) error {
	b.Learn.mu.Lock()
	b.Learn.strategies[strategy.ID] = strategy
	b.Learn.mu.Unlock()

	strategy.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(strategy, "", "  ")
	if err != nil {
		return err
	}

	strategiesDir := filepath.Join(b.homeDir, StrategiesDir)
	path := filepath.Join(strategiesDir, fmt.Sprintf("%s.json", sanitizeFilename(strategy.ID)))
	return os.WriteFile(path, data, 0644)
}

// EvolvePayload creates mutated offspring using real genetic algorithm
func (b *Brain) EvolvePayload(targetType string, parentPayloads []string, failureContext string) []string {
	b.Evolve.mu.Lock()
	defer b.Evolve.mu.Unlock()

	if len(parentPayloads) == 0 {
		return []string{}
	}

	// Create population with real genetic algorithm
	pop := NewPopulation(targetType, parentPayloads, 50)

	// Fitness function based on payload complexity and diversity
	fitnessFn := func(payload string) float64 {
		score := 50.0 // Base score

		// Reward encoding diversity
		if strings.Contains(payload, "%") {
			score += 10
		}
		if strings.Contains(payload, "\\u") {
			score += 10
		}
		if strings.Contains(payload, "\x00") {
			score += 15
		}

		// Reward length variation
		if len(payload) > len(parentPayloads[0]) {
			score += float64(len(payload)-len(parentPayloads[0])) * 0.5
		}

		// Reward comment injection (WAF bypass)
		if strings.Contains(payload, "/**/") {
			score += 10
		}

		// Context-specific bonuses
		switch targetType {
		case "sqli":
			if strings.Contains(payload, "UNION") || strings.Contains(payload, "SELECT") {
				score += 15
			}
		case "xss":
			if strings.Contains(payload, "<") && strings.Contains(payload, ">") {
				score += 15
			}
		case "lfi":
			if strings.Contains(payload, "../") || strings.Contains(payload, "..\\") {
				score += 15
			}
		}

		// Failure context bonus
		if failureContext != "" {
			score += 5
		}

		return score
	}

	// Run multiple generations for real evolution
	for gen := 0; gen < MaxGenerations; gen++ {
		pop.Evolve(fitnessFn)
		if pop.BestFitness >= EvolveThreshold*100 {
			break // Converged
		}
	}

	// Store generation in brain
	best := pop.GetBest()
	if best != nil {
		gen := &Generation{
			ID:           fmt.Sprintf("gen_%d_%d", time.Now().Unix(), len(b.Evolve.generations[targetType])),
			ParentIDs:    best.ParentIDs,
			Payload:      string(best.DNA),
			Score:        int(best.Fitness),
			TargetType:   targetType,
			MutationType: best.MutationType,
			CreatedAt:    time.Now(),
		}
		b.Evolve.generations[targetType] = append(b.Evolve.generations[targetType], gen)
	}

	// Return top 20 evolved payloads
	return pop.GetTop(20)
}

// RecordOutcome logs success or failure for learning
func (b *Brain) RecordOutcome(target, technique, payload, result string, score int) error {
	dna, err := b.GetTargetDNA(target)
	if err != nil {
		return err
	}

	payloadDNA := PayloadDNA{
		Payload:        payload,
		Technique:      technique,
		InjectionPoint: "auto-detected",
		Result:         result,
		Timestamp:      time.Now(),
	}

	if score >= 95 {
		dna.SuccessfulPayloads = append(dna.SuccessfulPayloads, payloadDNA)
	} else {
		dna.FailedPayloads = append(dna.FailedPayloads, payloadDNA)
	}

	return b.SaveTargetDNA(dna)
}

func sanitizeFilename(name string) string {
	// Simple sanitization
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}

// crossoverPayloads combines two payloads for genetic evolution
func crossoverPayloads(p1, p2 string) string {
	mid := len(p1) / 2
	if mid > len(p2) {
		mid = len(p2) / 2
	}
	return p1[:mid] + p2[mid:]
}
