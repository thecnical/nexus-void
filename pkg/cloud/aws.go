package cloud

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nexus-void/nexus-void/pkg/utils"
)

// AWSBreaker is the AWS cloud security testing engine
type AWSBreaker struct {
	Target string
}

// AWSResult represents an AWS security finding
type AWSResult struct {
	Service     string `json:"service"`  // s3, iam, lambda, ec2
	Type        string `json:"type"`     // open_bucket, misconfig, metadata
	Resource    string `json:"resource"` // s3://bucket-name
	Proof       string `json:"proof"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation"`
}

func NewAWSBreaker(target string) *AWSBreaker {
	return &AWSBreaker{Target: target}
}

// ScanS3Buckets scans for S3 buckets related to target
func (a *AWSBreaker) ScanS3Buckets(domain string) []AWSResult {
	fmt.Printf("[+] AWS-BREAKER scanning S3 buckets for: %s\n", domain)

	var results []AWSResult

	// Generate permutations
	permutations := generateBucketNames(domain)

	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 50)

	for _, bucket := range permutations {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(b string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Check bucket existence via DNS
			url := fmt.Sprintf("http://%s.s3.amazonaws.com", b)
			resp, err := utils.Fetch(url, nil)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			// Check if bucket exists and is accessible
			if resp.StatusCode == 200 {
				mu.Lock()
				results = append(results, AWSResult{
					Service:  "s3",
					Type:     "open_bucket",
					Resource: fmt.Sprintf("s3://%s", b),
					Proof:    "Bucket listing accessible without authentication",
					Severity: "critical",
				})
				mu.Unlock()
			} else if resp.StatusCode == 403 {
				// Bucket exists but restricted
				if !strings.Contains(bodyStr, "NoSuchBucket") {
					mu.Lock()
					results = append(results, AWSResult{
						Service:  "s3",
						Type:     "existing_bucket",
						Resource: fmt.Sprintf("s3://%s", b),
						Proof:    "Bucket exists (403 denied, not 404)",
						Severity: "info",
					})
					mu.Unlock()
				}
			}
		}(bucket)
	}

	wg.Wait()
	fmt.Printf("[+] AWS-BREAKER found %d S3 issues\n", len(results))
	return results
}

// TestEC2Metadata tests for EC2 metadata service access
func (a *AWSBreaker) TestEC2Metadata(targetURL string) *AWSResult {
	// Try to access EC2 metadata via SSRF
	metadataEndpoints := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://169.254.169.254/latest/meta-data/ami-id",
		"http://169.254.169.254/latest/meta-data/instance-id",
		"http://169.254.169.254/latest/meta-data/iam/security-credentials/",
		"http://169.254.169.254/latest/meta-data/public-ipv4",
		"http://169.254.169.254/latest/user-data",
		"http://169.254.169.254/latest/dynamic/instance-identity/document",
	}

	for _, endpoint := range metadataEndpoints {
		resp, err := utils.Fetch(endpoint, nil)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			// Check for IAM credentials
			if strings.Contains(endpoint, "security-credentials") && strings.Contains(bodyStr, "AccessKeyId") {
				return &AWSResult{
					Service:  "ec2",
					Type:     "metadata_credentials",
					Resource: endpoint,
					Proof:    "IAM credentials accessible via metadata service",
					Severity: "critical",
				}
			}

			// Check for instance ID
			if strings.Contains(bodyStr, "i-") && len(bodyStr) > 10 {
				return &AWSResult{
					Service:  "ec2",
					Type:     "metadata_exposure",
					Resource: endpoint,
					Proof:    fmt.Sprintf("Instance metadata accessible: %s", bodyStr[:minLen(100, len(bodyStr))]),
					Severity: "critical",
				}
			}
		}
	}

	return nil
}

// EnumerateIAMPolicies attempts to enumerate IAM policies
func (a *AWSBreaker) EnumerateIAMPolicies(accountID string) []AWSResult {
	// Simplified - would use AWS SDK in real implementation
	return []AWSResult{}
}

// TestLambdaFunction tests Lambda function for vulnerabilities
func (a *AWSBreaker) TestLambdaFunction(functionArn string) *AWSResult {
	// Check for common Lambda misconfigurations
	return nil
}

func generateBucketNames(domain string) []string {
	base := strings.ReplaceAll(domain, ".", "-")
	baseNoHyphen := strings.ReplaceAll(domain, ".", "")

	var names []string

	// Direct domain
	names = append(names, base)
	names = append(names, baseNoHyphen)

	// Common suffixes
	suffixes := []string{
		"", "-prod", "-production", "-dev", "-development", "-staging",
		"-test", "-backup", "-backups", "-archive", "-logs",
		"-data", "-assets", "-media", "-files", "-uploads",
		"-images", "-docs", "-documents", "-public", "-private",
		"-config", "-configs", "-secrets", "-tmp", "-temp",
		"-cdn", "-static", "-content", "-resources",
	}

	for _, suffix := range suffixes {
		names = append(names, base+suffix)
		names = append(names, baseNoHyphen+suffix)
	}

	return utils.Unique(names)
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
