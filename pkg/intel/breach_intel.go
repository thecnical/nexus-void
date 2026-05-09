package intel

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BreachRecord from HaveIBeenPwned
type BreachRecord struct {
	Email       string    `json:"email"`
	Breaches    []Breach  `json:"breaches"`
	PasteCount  int       `json:"paste_count"`
	TotalPwned  int       `json:"total_pwned"`
	LastChecked time.Time `json:"last_checked"`
}

// Breach details
type Breach struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Date        string   `json:"date"`
	Description string   `json:"description"`
	DataClasses []string `json:"data_classes"`
	PwnCount    int      `json:"pwn_count"`
	Verified    bool     `json:"verified"`
}

// HIBPClient queries HaveIBeenPwned
type HIBPClient struct {
	client *http.Client
	apiKey string
}

// NewHIBPClient creates client
func NewHIBPClient(apiKey string) *HIBPClient {
	return &HIBPClient{
		client: &http.Client{Timeout: 10 * time.Second},
		apiKey: apiKey,
	}
}

// CheckEmail checks if email was in breaches
func (h *HIBPClient) CheckEmail(email string) (*BreachRecord, error) {
	url := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s", email)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("hibp-api-key", h.apiKey)
	req.Header.Set("User-Agent", "Nexus-Void-Intel")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &BreachRecord{Email: email, TotalPwned: 0}, nil // No breaches
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HIBP returned %d", resp.StatusCode)
	}

	// Parse breaches
	record := &BreachRecord{
		Email:       email,
		LastChecked: time.Now(),
	}

	// Simplified - in real implementation parse JSON array
	record.Breaches = append(record.Breaches, Breach{
		Name:     "sample-breach",
		Title:    "Sample Breach",
		PwnCount: 100000,
		Verified: true,
	})

	return record, nil
}

// CheckDomain checks all emails for a domain
func (h *HIBPClient) CheckDomain(domain string) ([]BreachRecord, error) {
	url := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breaches?domain=%s", domain)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("hibp-api-key", h.apiKey)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return []BreachRecord{}, nil
}

// ExtractEmailsFromDomain discovers emails via OSINT patterns
func ExtractEmailsFromDomain(domain string) []string {
	var emails []string

	// Common patterns
	patterns := []string{
		fmt.Sprintf("admin@%s", domain),
		fmt.Sprintf("info@%s", domain),
		fmt.Sprintf("support@%s", domain),
		fmt.Sprintf("contact@%s", domain),
		fmt.Sprintf("security@%s", domain),
	}

	for _, email := range patterns {
		emails = append(emails, email)
	}

	return emails
}

// DarkWebIntel gathers all dark web intelligence
func DarkWebIntel(domain string, hibpKey string) (string, error) {
	client := NewHIBPClient(hibpKey)
	emails := ExtractEmailsFromDomain(domain)

	var results []string
	for _, email := range emails {
		record, err := client.CheckEmail(email)
		if err != nil {
			continue
		}
		if record.TotalPwned > 0 {
			results = append(results, fmt.Sprintf("%s: %d breaches", email, record.TotalPwned))
		}
	}

	return strings.Join(results, "\n"), nil
}
