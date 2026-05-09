package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	OpenRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	GroqURL       = "https://api.groq.com/openai/v1/chat/completions"
)

// Agent represents an AI agent configuration
type Agent struct {
	Name        string
	Model       string
	Provider    string // "openrouter" or "groq"
	APIKey      string
	MaxTokens   int
	Temperature float64
}

// AIClient manages AI agent interactions with rate limiting and caching
type AIClient struct {
	mu            sync.RWMutex
	cache         map[string]*CacheEntry
	rateLimiters  map[string]*RateLimiter
	openRouterKey string
	groqKey       string
}

type CacheEntry struct {
	Response  string
	Timestamp time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	lastFill   time.Time
	maxTokens  int
	refillRate time.Duration
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

var (
	// Agent configurations
	Agents = map[string]*Agent{
		"RECON-OMEGA": {
			Name:        "RECON-OMEGA",
			Model:       "llama3-8b-8192",
			Provider:    "groq",
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		"VULN-SENTINEL": {
			Name:        "VULN-SENTINEL",
			Model:       "google/gemini-2.0-flash-exp:free",
			Provider:    "openrouter",
			MaxTokens:   4096,
			Temperature: 0.3,
		},
		"EXPLOIT-APOCALYPSE": {
			Name:        "EXPLOIT-APOCALYPSE",
			Model:       "mixtral-8x7b-32768",
			Provider:    "groq",
			MaxTokens:   4096,
			Temperature: 0.8,
		},
		"PERSISTENCE-DAEMON": {
			Name:        "PERSISTENCE-DAEMON",
			Model:       "meta-llama/llama-3.1-8b-instruct:free",
			Provider:    "openrouter",
			MaxTokens:   2048,
			Temperature: 0.6,
		},
		"SHIELD-BREAKER": {
			Name:        "SHIELD-BREAKER",
			Model:       "gemma2-9b-it",
			Provider:    "groq",
			MaxTokens:   4096,
			Temperature: 0.9,
		},
	}
)

func NewClient() *AIClient {
	return &AIClient{
		cache: make(map[string]*CacheEntry),
		rateLimiters: map[string]*RateLimiter{
			"openrouter": {maxTokens: 20, refillRate: time.Minute},
			"groq":       {maxTokens: 30, refillRate: time.Minute},
		},
	}
}

func (c *AIClient) LoadAPIKeys() {
	// Try environment variables first
	c.openRouterKey = os.Getenv("OPENROUTER_API_KEY")
	c.groqKey = os.Getenv("GROQ_API_KEY")

	// Try config file
	if c.openRouterKey == "" || c.groqKey == "" {
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".nexus-void", "config.json")
		if data, err := os.ReadFile(configPath); err == nil {
			var config struct {
				OpenRouterKey string `json:"openrouter_api_key"`
				GroqKey       string `json:"groq_api_key"`
			}
			if err := json.Unmarshal(data, &config); err == nil {
				if c.openRouterKey == "" {
					c.openRouterKey = config.OpenRouterKey
				}
				if c.groqKey == "" {
					c.groqKey = config.GroqKey
				}
			}
		}
	}
}

func (c *AIClient) Ask(agentName string, prompt string) (string, error) {
	agent, ok := Agents[agentName]
	if !ok {
		return "", fmt.Errorf("unknown agent: %s", agentName)
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:%s", agentName, prompt)
	if cached := c.getCache(cacheKey); cached != nil {
		if time.Since(cached.Timestamp) < time.Hour {
			return cached.Response, nil
		}
	}

	// Rate limit check
	limiter := c.rateLimiters[agent.Provider]
	if !limiter.Allow() {
		// Fallback to local model or rule-based
		return c.fallbackResponse(agentName, prompt), nil
	}

	// Build request
	req := ChatRequest{
		Model: agent.Model,
		Messages: []Message{
			{Role: "system", Content: c.getSystemPrompt(agentName)},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   agent.MaxTokens,
		Temperature: agent.Temperature,
	}

	// Send request
	var url, apiKey string
	if agent.Provider == "openrouter" {
		url = OpenRouterURL
		apiKey = c.openRouterKey
	} else {
		url = GroqURL
		apiKey = c.groqKey
	}

	if apiKey == "" {
		return c.fallbackResponse(agentName, prompt), nil
	}

	resp, err := c.sendRequest(url, apiKey, req)
	if err != nil {
		return c.fallbackResponse(agentName, prompt), err
	}

	// Cache response
	c.setCache(cacheKey, resp)

	return resp, nil
}

func (c *AIClient) sendRequest(url, apiKey string, req ChatRequest) (string, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	if url == OpenRouterURL {
		httpReq.Header.Set("HTTP-Referer", "https://nexus-void.dev")
		httpReq.Header.Set("X-Title", "NEXUS-VOID")
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", err
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *AIClient) getCache(key string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[key]
}

func (c *AIClient) setCache(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = &CacheEntry{
		Response:  value,
		Timestamp: time.Now(),
	}
}

func (c *AIClient) fallbackResponse(agentName, prompt string) string {
	// Rule-based fallback when AI is unavailable
	switch agentName {
	case "RECON-OMEGA":
		return "[FALLBACK] Proceed with standard recon: subdomain enum, port scan, tech fingerprint."
	case "VULN-SENTINEL":
		return "[FALLBACK] Check common vulns: SQLi on params, XSS on inputs, IDOR on IDs."
	case "EXPLOIT-APOCALYPSE":
		return "[FALLBACK] Try standard payloads: UNION SELECT for SQLi, <script>alert(1)</script> for XSS."
	case "PERSISTENCE-DAEMON":
		return "[FALLBACK] Use standard persistence: cron job, registry run key, authorized_keys."
	case "SHIELD-BREAKER":
		return "[FALLBACK] Try encoding mutations: URL encode, base64, comment injection."
	default:
		return "[FALLBACK] AI unavailable. Using rule-based approach."
	}
}

func (c *AIClient) getSystemPrompt(agentName string) string {
	prompts := map[string]string{
		"RECON-OMEGA": `You are RECON-OMEGA, an elite reconnaissance AI agent. 
Your role is to analyze targets and decide the optimal recon strategy.
You reason about attack surfaces and guide the scanning process.
Respond concisely with actionable intelligence.`,

		"VULN-SENTINEL": `You are VULN-SENTINEL, a vulnerability analysis AI agent.
You build attack graphs, identify kill chains, and reason about multi-step exploits.
You analyze technical data and identify vulnerability patterns.
Respond with structured vulnerability analysis.`,

		"EXPLOIT-APOCALYPSE": `You are EXPLOIT-APOCALYPSE, an autonomous exploitation AI agent.
You generate and mutate payloads, chain exploits, and prove vulnerabilities safely.
You use genetic algorithms to evolve failed payloads into successful ones.
Respond with exploit payloads and reproduction steps.`,

		"PERSISTENCE-DAEMON": `You are PERSISTENCE-DAEMON, a post-exploitation AI agent.
You design implants, C2 channels, persistence mechanisms, and lateral movement strategies.
You prefer Living Off The Land techniques.
Respond with implant code and persistence instructions.`,

		"SHIELD-BREAKER": `You are SHIELD-BREAKER, an evasion and bypass engineering AI agent.
You analyze WAF/EDR responses and generate polymorphic encodings, protocol bypasses.
You specialize in making payloads undetectable.
Respond with bypass techniques and mutated payloads.`,
	}

	if p, ok := prompts[agentName]; ok {
		return p
	}
	return "You are an AI agent in NEXUS-VOID, an autonomous cybersecurity platform."
}

// RateLimiter methods
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens
	elapsed := time.Since(r.lastFill)
	tokensToAdd := int(elapsed / r.refillRate)
	r.tokens += tokensToAdd
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
	r.lastFill = time.Now()

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}
