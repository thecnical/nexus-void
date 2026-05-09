package intel

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EmployeeRecord represents a discovered employee
type EmployeeRecord struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	LinkedIn string   `json:"linkedin"`
	Emails   []string `json:"emails"`
	Skills   []string `json:"skills"`
}

// LinkedInOSINT performs LinkedIn reconnaissance
type LinkedInOSINT struct {
	client *http.Client
	apiKey string // If using a scraper API
}

// NewLinkedInOSINT creates client
func NewLinkedInOSINT() *LinkedInOSINT {
	return &LinkedInOSINT{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// SearchEmployees finds employees by company name
func (l *LinkedInOSINT) SearchEmployees(companyName string, limit int) ([]EmployeeRecord, error) {
	// Use Google dorking to find LinkedIn profiles
	_ = fmt.Sprintf("site:linkedin.com/in/ \"%s\" intitle:engineer OR intitle:developer OR intitle:admin",
		companyName)

	// In real implementation, use a search API or scraper
	// For now, return structured placeholder based on company

	records := []EmployeeRecord{
		{
			Name:     fmt.Sprintf("%s Admin", companyName),
			Title:    "System Administrator",
			LinkedIn: fmt.Sprintf("https://linkedin.com/in/%s-admin", strings.ToLower(companyName)),
			Emails:   []string{fmt.Sprintf("admin@%s.com", strings.ToLower(companyName))},
			Skills:   []string{"Linux", "Windows", "AWS"},
		},
		{
			Name:     fmt.Sprintf("%s Dev", companyName),
			Title:    "Software Developer",
			LinkedIn: fmt.Sprintf("https://linkedin.com/in/%s-dev", strings.ToLower(companyName)),
			Emails:   []string{fmt.Sprintf("dev@%s.com", strings.ToLower(companyName))},
			Skills:   []string{"Python", "Go", "JavaScript"},
		},
	}

	return records, nil
}

// EnumerateRoles maps employee roles to attack surface
func EnumerateRoles(employees []EmployeeRecord) map[string][]string {
	roleMap := make(map[string][]string)

	for _, emp := range employees {
		role := strings.ToLower(emp.Title)

		switch {
		case strings.Contains(role, "admin") || strings.Contains(role, "ops"):
			roleMap["privileged"] = append(roleMap["privileged"], emp.Name)
		case strings.Contains(role, "dev") || strings.Contains(role, "engineer"):
			roleMap["technical"] = append(roleMap["technical"], emp.Name)
		case strings.Contains(role, "security"):
			roleMap["security"] = append(roleMap["security"], emp.Name)
		default:
			roleMap["general"] = append(roleMap["general"], emp.Name)
		}
	}

	return roleMap
}

// BuildOrgChart structures employees by hierarchy
func BuildOrgChart(employees []EmployeeRecord) string {
	roles := EnumerateRoles(employees)

	var output []string
	output = append(output, "=== ORGANIZATIONAL STRUCTURE ===")
	output = append(output, fmt.Sprintf("Privileged Accounts: %v", roles["privileged"]))
	output = append(output, fmt.Sprintf("Technical Staff: %v", roles["technical"]))
	output = append(output, fmt.Sprintf("Security Team: %v", roles["security"]))
	output = append(output, fmt.Sprintf("General Staff: %v", roles["general"]))

	return strings.Join(output, "\n")
}
