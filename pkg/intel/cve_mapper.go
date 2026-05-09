package intel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CVERecord represents a CVE entry
type CVERecord struct {
	ID           string   `json:"id"`
	Severity     string   `json:"severity"`
	Score        float64  `json:"score"`
	Description  string   `json:"description"`
	Product      string   `json:"product"`
	Version      string   `json:"version"`
	AttackVector string   `json:"attack_vector"`
	Exploitable  bool     `json:"exploitable"`
	References   []string `json:"references"`
	Mitigation   string   `json:"mitigation"`
}

// CVEResponse from NVD API
type CVEResponse struct {
	ResultsPerPage  int `json:"resultsPerPage"`
	Vulnerabilities []struct {
		CVE struct {
			ID           string `json:"id"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
			Metrics struct {
				CVSSV31 struct {
					CVSSData struct {
						BaseScore    float64 `json:"baseScore"`
						AttackVector string  `json:"attackVector"`
					} `json:"cvssData"`
				} `json:"cvssMetricV31"`
			} `json:"metrics"`
			References []struct {
				URL string `json:"url"`
			} `json:"references"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

// CVEFetcher queries NVD for CVEs matching tech stack
type CVEFetcher struct {
	client *http.Client
	apiKey string
}

// NewCVEFetcher creates CVE mapper
func NewCVEFetcher(apiKey string) *CVEFetcher {
	return &CVEFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
		apiKey: apiKey,
	}
}

// SearchCVEs finds CVEs for a product
func (c *CVEFetcher) SearchCVEs(product string, limit int) ([]CVERecord, error) {
	url := fmt.Sprintf("https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=%s&resultsPerPage=%d",
		strings.ReplaceAll(product, " ", "%20"), limit)

	req, _ := http.NewRequest("GET", url, nil)
	if c.apiKey != "" {
		req.Header.Set("apiKey", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("NVD API returned %d", resp.StatusCode)
	}

	var nvdResp CVEResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, err
	}

	var records []CVERecord
	for _, vuln := range nvdResp.Vulnerabilities {
		cve := vuln.CVE
		desc := ""
		for _, d := range cve.Descriptions {
			if d.Lang == "en" {
				desc = d.Value
				break
			}
		}

		score := cve.Metrics.CVSSV31.CVSSData.BaseScore
		vector := cve.Metrics.CVSSV31.CVSSData.AttackVector
		if score == 0 {
			score = 0.0
			vector = ""
		}

		var refs []string
		for _, r := range cve.References {
			refs = append(refs, r.URL)
		}

		severity := "unknown"
		if score >= 9.0 {
			severity = "critical"
		} else if score >= 7.0 {
			severity = "high"
		} else if score >= 4.0 {
			severity = "medium"
		} else if score > 0 {
			severity = "low"
		}

		records = append(records, CVERecord{
			ID:           cve.ID,
			Severity:     severity,
			Score:        score,
			Description:  desc,
			Product:      product,
			AttackVector: vector,
			References:   refs,
			Exploitable:  strings.Contains(desc, "remote") || strings.Contains(desc, "unauthenticated"),
		})
	}

	return records, nil
}

// MapTechStackToCVEs takes tech stack and returns all CVEs
func MapTechStackToCVEs(techStack []string, apiKey string) []CVERecord {
	fetcher := NewCVEFetcher(apiKey)
	var allCVEs []CVERecord
	seen := make(map[string]bool)

	for _, tech := range techStack {
		cveList, err := fetcher.SearchCVEs(tech, 5)
		if err != nil {
			continue
		}
		for _, cve := range cveList {
			if !seen[cve.ID] {
				seen[cve.ID] = true
				allCVEs = append(allCVEs, cve)
			}
		}
	}

	return allCVEs
}
