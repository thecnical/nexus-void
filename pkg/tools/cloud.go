package tools

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

// === 1. AWS-BREAKER ===
func runAWSBreaker(args []string) (string, error) {
	// Check for AWS CLI and credentials
	output, err := RunExternalSafe("aws", "sts", "get-caller-identity")
	if err != nil {
		return "[AWS-BREAKER] AWS CLI not configured or no credentials", nil
	}
	return fmt.Sprintf("[AWS-BREAKER] Identity:\n%s", output), nil
}

// === 2. AWS-PRIVESC ===
func runAWSPrivesc(args []string) (string, error) {
	// Enumerate IAM policies
	var results []string
	output, err := RunExternalSafe("aws", "iam", "list-attached-user-policies", "--user-name", "$(aws sts get-caller-identity --query Arn --output text)")
	if err == nil {
		results = append(results, "Attached Policies:\n"+output)
	}
	output, err = RunExternalSafe("aws", "iam", "list-user-policies")
	if err == nil {
		results = append(results, "Inline Policies:\n"+output)
	}
	output, err = RunExternalSafe("aws", "iam", "list-roles")
	if err == nil {
		results = append(results, "Roles:\n"+output)
	}
	return fmt.Sprintf("[AWS-PRIVESC]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 3. AZURE-PHANTOM ===
func runAzurePhantom(args []string) (string, error) {
	output, err := RunExternalSafe("az", "account", "show")
	if err != nil {
		return "[AZURE-PHANTOM] Azure CLI not configured", nil
	}
	return fmt.Sprintf("[AZURE-PHANTOM] Account:\n%s", output), nil
}

// === 4. AZURE-PRIVESC ===
func runAzurePrivesc(args []string) (string, error) {
	var results []string
	output, _ := RunExternalSafe("az", "role", "assignment", "list", "--assignee")
	if output != "" {
		results = append(results, "Role Assignments:\n"+output)
	}
	output, _ = RunExternalSafe("az", "ad", "user", "list")
	if output != "" {
		results = append(results, "AD Users:\n"+output)
	}
	return fmt.Sprintf("[AZURE-PRIVESC]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 5. GCP-RAIDER ===
func runGCPRaider(args []string) (string, error) {
	output, err := RunExternalSafe("gcloud", "config", "list")
	if err != nil {
		return "[GCP-RAIDER] gcloud not configured", nil
	}
	return fmt.Sprintf("[GCP-RAIDER] Config:\n%s", output), nil
}

// === 6. GCP-PRIVESC ===
func runGCPPrivesc(args []string) (string, error) {
	var results []string
	output, _ := RunExternalSafe("gcloud", "projects", "get-iam-policy", "$(gcloud config get-value project)")
	if output != "" {
		results = append(results, "IAM Policy:\n"+output)
	}
	return fmt.Sprintf("[GCP-PRIVESC]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 7. K8S-ASSASSIN ===
func runK8SAssassin(args []string) (string, error) {
	output, err := RunExternalSafe("kubectl", "get", "pods", "--all-namespaces")
	if err != nil {
		return "[K8S-ASSASSIN] kubectl not configured or no cluster access", nil
	}
	return fmt.Sprintf("[K8S-ASSASSIN] All Pods:\n%s", output), nil
}

// === 8. K8S-SECRET-HARVESTER ===
func runK8SSecretHarvester(args []string) (string, error) {
	output, err := RunExternalSafe("kubectl", "get", "secrets", "--all-namespaces", "-o", "yaml")
	if err != nil {
		return "[K8S-SECRET-HARVESTER] kubectl not configured", nil
	}
	return fmt.Sprintf("[K8S-SECRET-HARVESTER] Secrets:\n%s", output), nil
}

// === 9. DOCKER-BREAKER ===
func runDockerBreaker(args []string) (string, error) {
	output, err := RunExternalSafe("docker", "ps")
	if err != nil {
		return "[DOCKER-BREAKER] Docker not available or no permissions", nil
	}
	return fmt.Sprintf("[DOCKER-BREAKER] Running containers:\n%s", output), nil
}

// === 10. DOCKER-REGISTRY-ABUSER ===
func runDockerRegistryAbuser(args []string) (string, error) {
	registry := "https://index.docker.io"
	if len(args) > 0 {
		registry = args[0]
	}
	resp, err := http.Get(registry + "/v2/_catalog")
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return fmt.Sprintf("[DOCKER-REGISTRY-ABUSER] Registry: %s\nStatus: %s\nBody: %s",
		registry, resp.Status, string(body)), nil
}

// === 11. TERRAFORM-HARVESTER ===
func runTerraformHarvester(args []string) (string, error) {
	var results []string
	// Check for terraform state files
	output, _ := RunExternalSafe("find", ".", "-name", "terraform.tfstate", "-o", "-name", "*.tfstate")
	if output != "" {
		results = append(results, "State files found:\n"+output)
	}
	output, _ = RunExternalSafe("find", ".", "-name", "*.tfvars")
	if output != "" {
		results = append(results, "TFVars files found:\n"+output)
	}
	return fmt.Sprintf("[TERRAFORM-HARVESTER]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 12. CLOUD-SHADOW ===
func runCloudShadow(args []string) (string, error) {
	var results []string
	results = append(results, "Checking cloud persistence mechanisms...")
	output, _ := RunExternalSafe("aws", "lambda", "list-functions")
	if output != "" {
		results = append(results, "Lambda functions:\n"+output)
	}
	output, _ = RunExternalSafe("aws", "events", "list-rules")
	if output != "" {
		results = append(results, "EventBridge rules:\n"+output)
	}
	return fmt.Sprintf("[CLOUD-SHADOW]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 13. SERVERLESS-PWN ===
func runServerlessPwn(args []string) (string, error) {
	var results []string
	results = append(results, "Enumerating serverless resources...")
	output, _ := RunExternalSafe("aws", "lambda", "list-functions", "--query", "Functions[*].FunctionName")
	if output != "" {
		results = append(results, "Functions: "+output)
	}
	output, _ = RunExternalSafe("aws", "lambda", "get-function-url-config", "--function-name")
	if output != "" {
		results = append(results, "Function URLs: "+output)
	}
	return fmt.Sprintf("[SERVERLESS-PWN]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 14. CLOUDFRONT-BYPASS ===
func runCloudFrontBypass(args []string) (string, error) {
	target := "https://example.com"
	if len(args) > 0 {
		target = args[0]
	}
	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("Host", "origin.example.com")
	req.Header.Set("X-Forwarded-Host", "origin.example.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return fmt.Sprintf("[CLOUDFRONT-BYPASS] Target: %s\nStatus: %s\nHeaders:\n%s\nBody: %s",
		target, resp.Status, headersToString(resp.Header), string(body)[:min(len(body), 500)]), nil
}

// === 15. IAM-PRIVESC ===
func runIAMPrivesc(args []string) (string, error) {
	var results []string
	output, _ := RunExternalSafe("aws", "iam", "list-users")
	if output != "" {
		results = append(results, "Users:\n"+output)
	}
	output, _ = RunExternalSafe("aws", "iam", "list-groups")
	if output != "" {
		results = append(results, "Groups:\n"+output)
	}
	output, _ = RunExternalSafe("aws", "iam", "list-policies", "--scope", "Local")
	if output != "" {
		results = append(results, "Custom Policies:\n"+output)
	}
	return fmt.Sprintf("[IAM-PRIVESC]\n%s", strings.Join(results, "\n---\n")), nil
}

func headersToString(h http.Header) string {
	var lines []string
	for k, v := range h {
		lines = append(lines, k+": "+strings.Join(v, ", "))
	}
	return strings.Join(lines, "\n")
}

// runExternalSafe already defined in registry.go
func runExternalSafe(command string, args ...string) (string, error) {
	if _, err := exec.LookPath(command); err != nil {
		return "", fmt.Errorf("tool not installed: %s", command)
	}
	return RunExternal(command, args...)
}
