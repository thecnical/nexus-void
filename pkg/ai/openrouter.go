package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	OpenRouterEndpoint = "https://openrouter.ai/api/v1/chat/completions"
	DefaultModel       = "openai/gpt-oss-120b:free"
)

// OpenRouterClient connects to OpenRouter for free AI inference
type OpenRouterClient struct {
	APIKey string
	Model  string
	Client *http.Client
}

// NewOpenRouter creates a client using the OPENROUTER_API_KEY env var
func NewOpenRouter() *OpenRouterClient {
	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		key = os.Getenv("OPENROUTER_KEY")
	}
	return &OpenRouterClient{
		APIKey: key,
		Model:  DefaultModel,
		Client: &http.Client{Timeout: 120 * time.Second},
	}
}

// Message is a chat message
 type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the payload sent to OpenRouter
 type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse is the OpenRouter response
 type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a prompt to the OpenRouter model and returns the response
func (c *OpenRouterClient) Chat(systemPrompt, userPrompt string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY not set")
	}

	reqBody := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", OpenRouterEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("HTTP-Referer", "https://nexus-void.ai")
	req.Header.Set("X-Title", "Nexus Void")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openrouter HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("openrouter error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in openrouter response")
	}

	return result.Choices[0].Message.Content, nil
}

// AutonomousDecision asks the AI what action to take next
func (c *OpenRouterClient) AutonomousDecision(context, availableActions string) (string, error) {
	system := `You are OMEGA-BRAIN, the autonomous AI pentest orchestrator for Nexus Void. 
You analyze recon data, vulnerability findings, and tool outputs.
Respond ONLY with the exact action name and parameters. No explanations.`
	user := fmt.Sprintf("Context:\n%s\n\nAvailable actions:\n%s\n\nWhat action should be taken next?", context, availableActions)
	return c.Chat(system, user)
}

// AnalyzeToolOutput asks AI to interpret raw tool output
func (c *OpenRouterClient) AnalyzeToolOutput(toolName, output string) (string, error) {
	system := "You are a cybersecurity expert. Analyze tool output and extract critical findings, vulnerabilities, and next steps. Be concise."
	user := fmt.Sprintf("Tool: %s\nOutput:\n%s\n\nSummarize critical findings and recommend next action.", toolName, output)
	return c.Chat(system, user)
}

// GeneratePayload asks AI to generate or mutate an attack payload
func (c *OpenRouterClient) GeneratePayload(targetTech, vulnType string) (string, error) {
	system := "You are a security researcher generating proof-of-concept payloads for authorized penetration testing. Output ONLY the payload, no markdown."
	user := fmt.Sprintf("Target technology: %s\nVulnerability type: %s\nGenerate a working PoC payload.", targetTech, vulnType)
	return c.Chat(system, user)
}
