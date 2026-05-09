package intel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MISPClient connects to MISP threat intel platform
type MISPClient struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// MISPEvent represents a threat intel event
type MISPEvent struct {
	ID          string          `json:"id,omitempty"`
	Info        string          `json:"info"`
	ThreatLevel string          `json:"threat_level_id"`
	Analysis    string          `json:"analysis"`
	Date        string          `json:"date"`
	Attributes  []MISPAttribute `json:"Attribute"`
	Tags        []string        `json:"Tag"`
}

// MISPAttribute represents an IOC
type MISPAttribute struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Comment  string `json:"comment"`
	Category string `json:"category"`
}

// NewMISPClient creates MISP connector
func NewMISPClient(baseURL, apiKey string) *MISPClient {
	return &MISPClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// SearchIOC queries MISP for an indicator
func (m *MISPClient) SearchIOC(iocType, value string) ([]MISPAttribute, error) {
	if m.APIKey == "" {
		return nil, fmt.Errorf("MISP API key not configured")
	}

	url := fmt.Sprintf("%s/attributes/restSearch", m.BaseURL)
	body := map[string]interface{}{
		"returnFormat": "json",
		"type":         iocType,
		"value":        value,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", m.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Response struct {
			Attribute []MISPAttribute `json:"Attribute"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Response.Attribute, nil
}

// CreateEvent pushes findings to MISP as a new event
func (m *MISPClient) CreateEvent(target string, findings []string) (string, error) {
	if m.APIKey == "" {
		return "", fmt.Errorf("MISP API key not configured")
	}

	event := MISPEvent{
		Info:        fmt.Sprintf("Nexus-Void Scan: %s", target),
		ThreatLevel: "3",
		Analysis:    "2",
		Date:        time.Now().Format("2006-01-02"),
		Tags:        []string{"nexus-void", "automated-scan", "penetration-test"},
	}

	for _, f := range findings {
		event.Attributes = append(event.Attributes, MISPAttribute{
			Type:     "text",
			Value:    f,
			Category: "External analysis",
			Comment:  "Discovered by Nexus-Void AI",
		})
	}

	url := fmt.Sprintf("%s/events/add", m.BaseURL)
	jsonBody, _ := json.Marshal(event)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", m.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Event struct {
			ID string `json:"id"`
		} `json:"Event"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Event.ID, nil
}

// EnrichFindings queries MISP for each finding and returns enriched data
func EnrichFindings(target string, findings []string, mispURL, mispKey string) map[string][]MISPAttribute {
	client := NewMISPClient(mispURL, mispKey)
	enriched := make(map[string][]MISPAttribute)

	for _, f := range findings {
		attrs, _ := client.SearchIOC("text", f)
		if len(attrs) > 0 {
			enriched[f] = attrs
		}
	}

	return enriched
}
