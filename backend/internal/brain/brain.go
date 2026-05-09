package brain

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	BrainDir        = ".nexus-void/brain"
	KnowledgeDB     = "knowledge_graph.db"
	MaxGenerations  = 50
	EvolveThreshold = 0.92
)

// Brain is the central autonomous reasoning engine
type Brain struct {
	DB       *gorm.DB
	homeDir  string
	mu       sync.RWMutex
	Memory   *MemoryModule
	Learn    *LearnModule
	Evolve   *EvolveModule
	Reason   *ReasonModule
	Predict  *PredictModule
	sessions map[string]*Session
}

// MemoryModule stores all historical data with semantic indexing
type MemoryModule struct {
	brain     *Brain
	shortTerm map[string]interface{}
	longTerm  map[string]interface{}
	mu        sync.RWMutex
}

// LearnModule recognizes patterns and adapts strategies
type LearnModule struct {
	brain      *Brain
	strategies map[string]*Strategy
	patterns   map[string]*Pattern
	mu         sync.RWMutex
}

// EvolveModule uses genetic algorithms to mutate payloads
type EvolveModule struct {
	brain       *Brain
	generations map[string][]*Generation
	dnaArchive  map[string]*TargetDNA
	mu          sync.RWMutex
}

// ReasonModule performs autonomous decision making
type ReasonModule struct {
	brain     *Brain
	knowledge map[string]*Fact
	rules     []Rule
	mu        sync.RWMutex
}

// PredictModule predicts target behavior and vulnerability likelihood
type PredictModule struct {
	brain  *Brain
	models map[string]*PredictiveModel
	mu     sync.RWMutex
}

// Strategy represents a learned attack strategy
type Strategy struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	Technique   string    `json:"technique"`
	Condition   string    `json:"condition"`
	Action      string    `json:"action"`
	SuccessRate float64   `json:"success_rate"`
	UseCount    int       `json:"use_count"`
	AvgTime     float64   `json:"avg_time_ms"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Pattern represents a discovered vulnerability pattern
type Pattern struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Signature   string   `json:"signature"`
	Confidence  float64  `json:"confidence"`
	Occurrences int      `json:"occurrences"`
	Targets     []string `json:"targets" gorm:"serializer:json"`
}

// Generation represents an evolved payload generation
type Generation struct {
	ID           string    `json:"id"`
	ParentIDs    []string  `json:"parent_ids" gorm:"serializer:json"`
	Payload      string    `json:"payload"`
	Score        int       `json:"score"`
	TargetType   string    `json:"target_type"`
	MutationType string    `json:"mutation_type"`
	CreatedAt    time.Time `json:"created_at"`
}

// Fact represents a known fact about targets
type Fact struct {
	ID         string  `json:"id"`
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
}

// Rule represents a reasoning rule
type Rule struct {
	ID       string `json:"id"`
	If       string `json:"if"`
	Then     string `json:"then"`
	Priority int    `json:"priority"`
}

// PredictiveModel stores trained prediction models
type PredictiveModel struct {
	ID             string             `json:"id"`
	FeatureWeights map[string]float64 `json:"feature_weights" gorm:"serializer:json"`
	Accuracy       float64            `json:"accuracy"`
	TrainedOn      int                `json:"trained_on"`
	LastUpdated    time.Time          `json:"last_updated"`
}

// TargetDNA stores evolved knowledge about a specific target
type TargetDNA struct {
	Target          string             `json:"target"`
	Fingerprint     string             `json:"fingerprint"`
	Technologies    []string           `json:"technologies" gorm:"serializer:json"`
	Behaviors       []string           `json:"behaviors" gorm:"serializer:json"`
	Weaknesses      []string           `json:"weaknesses" gorm:"serializer:json"`
	SuccessfulPaths []string           `json:"successful_paths" gorm:"serializer:json"`
	FailedPaths     []string           `json:"failed_paths" gorm:"serializer:json"`
	ConfidenceMap   map[string]float64 `json:"confidence_map" gorm:"serializer:json"`
	LastUpdated     time.Time          `json:"last_updated"`
}

// Session represents an operation session
type Session struct {
	ID          string       `json:"id"`
	Target      string       `json:"target"`
	StartTime   time.Time    `json:"start_time"`
	EndTime     *time.Time   `json:"end_time,omitempty"`
	Status      string       `json:"status"`
	Findings    int          `json:"findings"`
	Exploits    int          `json:"exploits"`
	Agents      []string     `json:"agents" gorm:"serializer:json"`
	Checkpoints []Checkpoint `json:"checkpoints" gorm:"serializer:json"`
	Strategy    string       `json:"strategy"`
}

type Checkpoint struct {
	Timestamp time.Time `json:"timestamp"`
	State     string    `json:"state"`
	FilePath  string    `json:"file_path"`
}

// Initialize creates and initializes the brain
func Initialize(homeDir string) (*Brain, error) {
	brainDir := filepath.Join(homeDir, BrainDir)
	if err := os.MkdirAll(brainDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(brainDir, KnowledgeDB)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("brain db: %w", err)
	}

	b := &Brain{
		DB:       db,
		homeDir:  homeDir,
		sessions: make(map[string]*Session),
	}

	b.Memory = &MemoryModule{brain: b, shortTerm: make(map[string]interface{}), longTerm: make(map[string]interface{})}
	b.Learn = &LearnModule{brain: b, strategies: make(map[string]*Strategy), patterns: make(map[string]*Pattern)}
	b.Evolve = &EvolveModule{brain: b, generations: make(map[string][]*Generation), dnaArchive: make(map[string]*TargetDNA)}
	b.Reason = &ReasonModule{brain: b, knowledge: make(map[string]*Fact), rules: defaultRules()}
	b.Predict = &PredictModule{brain: b, models: make(map[string]*PredictiveModel)}

	if err := db.AutoMigrate(&Strategy{}, &Pattern{}, &Generation{}, &Fact{}, &Session{}); err != nil {
		return nil, err
	}

	b.loadInitialKnowledge()
	fmt.Println("[+] OMEGA Brain v3.0 initialized - Autonomous reasoning online")
	return b, nil
}

// Close shuts down the brain
func (b *Brain) Close() error {
	sqlDB, err := b.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// --- Memory Module ---

func (m *MemoryModule) StoreShort(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shortTerm[key] = value
}

func (m *MemoryModule) StoreLong(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.longTerm[key] = value

	// Persist to DB
	data, _ := json.Marshal(value)
	m.brain.DB.Exec("INSERT OR REPLACE INTO memories (key, value, created_at) VALUES (?, ?, ?)",
		key, string(data), time.Now())
}

func (m *MemoryModule) Recall(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m.shortTerm[key]; ok {
		return v, true
	}
	if v, ok := m.longTerm[key]; ok {
		return v, true
	}
	return nil, false
}

// --- Learn Module ---

func (l *LearnModule) RecordSuccess(technique, domain string, durationMs float64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := fmt.Sprintf("%s:%s", domain, technique)
	strat, exists := l.strategies[key]
	if !exists {
		strat = &Strategy{
			ID:        key,
			Domain:    domain,
			Technique: technique,
			CreatedAt: time.Now(),
		}
		l.strategies[key] = strat
	}

	strat.UseCount++
	strat.SuccessRate = (strat.SuccessRate*float64(strat.UseCount-1) + 1.0) / float64(strat.UseCount)
	strat.AvgTime = (strat.AvgTime*float64(strat.UseCount-1) + durationMs) / float64(strat.UseCount)
	strat.UpdatedAt = time.Now()

	l.brain.DB.Save(strat)
}

func (l *LearnModule) RecordFailure(technique, domain string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := fmt.Sprintf("%s:%s", domain, technique)
	strat, exists := l.strategies[key]
	if !exists {
		strat = &Strategy{
			ID:        key,
			Domain:    domain,
			Technique: technique,
			CreatedAt: time.Now(),
		}
		l.strategies[key] = strat
	}

	strat.UseCount++
	strat.SuccessRate = (strat.SuccessRate * float64(strat.UseCount-1)) / float64(strat.UseCount)
	strat.UpdatedAt = time.Now()
	l.brain.DB.Save(strat)
}

func (l *LearnModule) GetBestStrategy(domain string) *Strategy {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var best *Strategy
	for _, s := range l.strategies {
		if s.Domain == domain && (best == nil || s.SuccessRate > best.SuccessRate) {
			best = s
		}
	}
	return best
}

// --- Reason Module ---

func (r *ReasonModule) Infer(facts []Fact) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var conclusions []string
	for _, rule := range r.rules {
		if r.matches(rule.If, facts) {
			conclusions = append(conclusions, rule.Then)
		}
	}
	return conclusions
}

func (r *ReasonModule) matches(condition string, facts []Fact) bool {
	for _, f := range facts {
		if fmt.Sprintf("%s %s %s", f.Subject, f.Predicate, f.Object) == condition {
			return true
		}
	}
	return false
}

// --- Predict Module ---

func (p *PredictModule) PredictVulnerability(targetDNA *TargetDNA, vulnType string) float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	model, ok := p.models[vulnType]
	if !ok {
		// Default heuristic prediction
		return p.heuristicPredict(targetDNA, vulnType)
	}

	score := 0.0
	for tech, weight := range model.FeatureWeights {
		for _, t := range targetDNA.Technologies {
			if t == tech {
				score += weight
			}
		}
	}
	return min(score, 1.0)
}

func (p *PredictModule) heuristicPredict(dna *TargetDNA, vulnType string) float64 {
	score := 0.0
	techSet := make(map[string]bool)
	for _, t := range dna.Technologies {
		techSet[t] = true
	}

	switch vulnType {
	case "sqli":
		if techSet["php"] || techSet["wordpress"] || techSet["drupal"] {
			score += 0.4
		}
		if techSet["mysql"] || techSet["mssql"] || techSet["postgresql"] {
			score += 0.3
		}
		if techSet["asp"] || techSet["asp.net"] {
			score += 0.2
		}
	case "xss":
		if techSet["javascript"] || techSet["react"] || techSet["angular"] {
			score += 0.3
		}
		if techSet["php"] || techSet["jsp"] {
			score += 0.2
		}
		score += float64(len(dna.Behaviors)) * 0.05
	case "rce":
		if techSet["java"] || techSet["struts"] || techSet["spring"] {
			score += 0.4
		}
		if techSet["python"] || techSet["django"] || techSet["flask"] {
			score += 0.2
		}
		if techSet["nodejs"] || techSet["express"] {
			score += 0.15
		}
	case "ssrf":
		if techSet["java"] || techSet["python"] || techSet["nodejs"] {
			score += 0.3
		}
		if techSet["cloud"] || techSet["aws"] || techSet["gcp"] {
			score += 0.25
		}
	}

	// Adjust based on past success/failure
	for _, path := range dna.SuccessfulPaths {
		if path == vulnType {
			score += 0.2
		}
	}
	for _, path := range dna.FailedPaths {
		if path == vulnType {
			score -= 0.15
		}
	}

	return min(max(score, 0.0), 1.0)
}

// --- Evolve Module ---

func (e *EvolveModule) EvolvePayload(seed string, targetType string, generations int) *Generation {
	e.mu.Lock()
	defer e.mu.Unlock()

	if generations > MaxGenerations {
		generations = MaxGenerations
	}

	best := &Generation{
		ID:         hash(seed),
		Payload:    seed,
		Score:      e.scorePayload(seed, targetType),
		TargetType: targetType,
	}

	for i := 0; i < generations; i++ {
		mutated := e.mutate(best.Payload)
		score := e.scorePayload(mutated, targetType)

		gen := &Generation{
			ID:           hash(mutated),
			ParentIDs:    []string{best.ID},
			Payload:      mutated,
			Score:        score,
			TargetType:   targetType,
			MutationType: e.mutationType(mutated, best.Payload),
			CreatedAt:    time.Now(),
		}

		e.generations[targetType] = append(e.generations[targetType], gen)

		if score > best.Score {
			best = gen
		}

		if float64(score)/100.0 >= EvolveThreshold {
			break
		}
	}

	return best
}

func (e *EvolveModule) mutate(payload string) string {
	mutations := []func(string) string{
		func(s string) string { return s + "/*" },
		func(s string) string { return "%2527" + s },
		func(s string) string { return s + "-- -" },
		func(s string) string { return "/*!50000" + s + "*/" },
		func(s string) string { return s + "#" },
		func(s string) string { return "0x" + hexEncode(s) },
		func(s string) string { return "BASE64:" + base64Encode(s) },
	}

	idx := time.Now().UnixNano() % int64(len(mutations))
	return mutations[idx](payload)
}

func (e *EvolveModule) scorePayload(payload, targetType string) int {
	score := 50

	// Length complexity bonus
	if len(payload) > 20 {
		score += 5
	}
	if len(payload) > 50 {
		score += 5
	}

	// Encoding diversity bonus
	if containsAny(payload, []string{"%", "0x", "/*!", "BASE64", "/*"}) {
		score += 10
	}

	// Target-specific scoring
	switch targetType {
	case "sqli":
		if containsAny(payload, []string{"SELECT", "UNION", "sleep", "benchmark"}) {
			score += 10
		}
		if containsAny(payload, []string{"information_schema", "sysobjects"}) {
			score += 5
		}
	case "xss":
		if containsAny(payload, []string{"<script", "javascript:", "onerror", "onload"}) {
			score += 10
		}
		if containsAny(payload, []string{"svg", "iframe", "object", "embed"}) {
			score += 5
		}
	case "cmdi":
		if containsAny(payload, []string{";", "&&", "||", "|", "$", "`"}) {
			score += 10
		}
		if containsAny(payload, []string{"nc", "bash", "python", "curl"}) {
			score += 5
		}
	}

	// Penalize simple payloads
	if len(payload) < 5 {
		score -= 20
	}

	return min(score, 100)
}

func (e *EvolveModule) mutationType(a, b string) string {
	if len(a) > len(b) {
		return "expansion"
	}
	if len(a) < len(b) {
		return "contraction"
	}
	return "substitution"
}

// --- Target DNA ---

func (b *Brain) UpdateTargetDNA(target string, dna *TargetDNA) {
	b.mu.Lock()
	defer b.mu.Unlock()

	dna.LastUpdated = time.Now()
	b.Evolve.dnaArchive[target] = dna

	data, _ := json.Marshal(dna)
	dnaPath := filepath.Join(b.homeDir, BrainDir, "target_dna", hash(target)+".json")
	os.MkdirAll(filepath.Dir(dnaPath), 0755)
	os.WriteFile(dnaPath, data, 0644)
}

func (b *Brain) GetTargetDNA(target string) *TargetDNA {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Evolve.dnaArchive[target]
}

// --- Session Management ---

func (b *Brain) NewSession(target string) *Session {
	session := &Session{
		ID:        hash(target + time.Now().String()),
		Target:    target,
		StartTime: time.Now(),
		Status:    "running",
		Agents:    []string{},
	}

	b.mu.Lock()
	b.sessions[session.ID] = session
	b.mu.Unlock()

	b.DB.Create(session)
	return session
}

func (b *Brain) UpdateSession(sessionID string, updates map[string]interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if session, ok := b.sessions[sessionID]; ok {
		for key, val := range updates {
			switch key {
			case "findings":
				if v, ok := val.(int); ok {
					session.Findings += v
				}
			case "exploits":
				if v, ok := val.(int); ok {
					session.Exploits += v
				}
			case "status":
				if v, ok := val.(string); ok {
					session.Status = v
				}
			}
		}
		b.DB.Save(session)
	}
}

// --- Helpers ---

func defaultRules() []Rule {
	return []Rule{
		{ID: "r1", If: "technology contains WordPress", Then: "test_sqli_plugin", Priority: 10},
		{ID: "r2", If: "technology contains Apache", Then: "test_path_traversal", Priority: 8},
		{ID: "r3", If: "technology contains IIS", Then: "test_iis_shortname", Priority: 9},
		{ID: "r4", If: "technology contains Java", Then: "test_spring_rce", Priority: 10},
		{ID: "r5", If: "technology contains Node.js", Then: "test_prototype_pollution", Priority: 8},
		{ID: "r6", If: "technology contains PHP", Then: "test_lfi_wrapper", Priority: 9},
		{ID: "r7", If: "technology contains .NET", Then: "test_deserialization", Priority: 10},
		{ID: "r8", If: "technology contains Cloud", Then: "test_ssrf_metadata", Priority: 9},
	}
}

func (b *Brain) loadInitialKnowledge() {
	initialFacts := []Fact{
		{ID: "f1", Subject: "WordPress", Predicate: "has", Object: "plugin_sqli", Confidence: 0.85},
		{ID: "f2", Subject: "Apache", Predicate: "has", Object: "path_traversal", Confidence: 0.75},
		{ID: "f3", Subject: "Java", Predicate: "has", Object: "deserialization_rce", Confidence: 0.90},
		{ID: "f4", Subject: "Node.js", Predicate: "has", Object: "prototype_pollution", Confidence: 0.70},
		{ID: "f5", Subject: "PHP", Predicate: "has", Object: "lfi_wrapper", Confidence: 0.80},
		{ID: "f6", Subject: "Cloud", Predicate: "has", Object: "ssrf_metadata", Confidence: 0.85},
		{ID: "f7", Subject: "Active Directory", Predicate: "has", Object: "kerberoasting", Confidence: 0.90},
		{ID: "f8", Subject: "Windows", Predicate: "has", Object: "pass_the_hash", Confidence: 0.80},
	}

	for _, f := range initialFacts {
		b.Reason.knowledge[f.ID] = &f
	}
}

func (b *Brain) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"strategies_learned":  len(b.Learn.strategies),
		"patterns_discovered": len(b.Learn.patterns),
		"generations_created": func() int {
			count := 0
			for _, gens := range b.Evolve.generations {
				count += len(gens)
			}
			return count
		}(),
		"targets_profiled": len(b.Evolve.dnaArchive),
		"active_sessions":  len(b.sessions),
		"facts_known":      len(b.Reason.knowledge),
		"models_trained":   len(b.Predict.models),
	}
}

// ExportStrategies exports learned strategies
func (b *Brain) ExportStrategies() []*Strategy {
	b.Learn.mu.RLock()
	defer b.Learn.mu.RUnlock()

	var list []*Strategy
	for _, s := range b.Learn.strategies {
		list = append(list, s)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].SuccessRate > list[j].SuccessRate
	})

	return list
}

// --- Utility functions ---

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}

func hexEncode(s string) string {
	return hex.EncodeToString([]byte(s))
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) > 0 {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
