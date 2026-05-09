package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolDefinition describes an external tool
type ToolDefinition struct {
	Name           string
	Description    string
	Category       string // web, network, cloud, osint, postexploit
	InstallMethods []InstallMethod
	CheckCommand   string // Command to verify installation
}

type InstallMethod struct {
	Type     string // go_install, binary, docker, pip, package_manager, git_clone
	Command  string
	Priority int // Lower = tried first
}

var externalTools = []ToolDefinition{
	{
		Name:         "nmap",
		Description:  "Network port scanner",
		Category:     "network",
		CheckCommand: "nmap --version",
		InstallMethods: []InstallMethod{
			{Type: "package_manager", Command: "nmap", Priority: 1},
			{Type: "binary", Command: "https://nmap.org/dist/nmap-7.95-setup.exe", Priority: 2},
		},
	},
	{
		Name:         "masscan",
		Description:  "Ultra-fast port scanner",
		Category:     "network",
		CheckCommand: "masscan --version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/robertdavidgraham/masscan", Priority: 1},
			{Type: "git_clone", Command: "https://github.com/robertdavidgraham/masscan.git", Priority: 2},
		},
	},
	{
		Name:         "amass",
		Description:  "Subdomain enumeration",
		Category:     "osint",
		CheckCommand: "amass -version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/owasp-amass/amass/v4/...@master", Priority: 1},
		},
	},
	{
		Name:         "subfinder",
		Description:  "Subdomain discovery",
		Category:     "osint",
		CheckCommand: "subfinder -version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/projectdiscovery/subfinder/v2/cmd/subfinder", Priority: 1},
		},
	},
	{
		Name:         "katana",
		Description:  "Fast web crawler",
		Category:     "web",
		CheckCommand: "katana -version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/projectdiscovery/katana/cmd/katana", Priority: 1},
		},
	},
	{
		Name:         "paramspider",
		Description:  "Hidden parameter discovery",
		Category:     "web",
		CheckCommand: "paramspider --help",
		InstallMethods: []InstallMethod{
			{Type: "pip", Command: "git+https://github.com/devanshbatham/ParamSpider.git", Priority: 1},
		},
	},
	{
		Name:         "gau",
		Description:  "URL extraction from archives",
		Category:     "web",
		CheckCommand: "gau --version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/lc/gau/v2/cmd/gau", Priority: 1},
		},
	},
	{
		Name:         "sqlmap",
		Description:  "SQL injection automation",
		Category:     "web",
		CheckCommand: "sqlmap --version",
		InstallMethods: []InstallMethod{
			{Type: "pip", Command: "sqlmap", Priority: 1},
			{Type: "git_clone", Command: "https://github.com/sqlmapproject/sqlmap.git", Priority: 2},
		},
	},
	{
		Name:         "dalfox",
		Description:  "XSS detection and exploitation",
		Category:     "web",
		CheckCommand: "dalfox version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/hahwul/dalfox/v2", Priority: 1},
		},
	},
	{
		Name:         "nuclei",
		Description:  "Vulnerability scanner",
		Category:     "web",
		CheckCommand: "nuclei -version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/projectdiscovery/nuclei/v3/cmd/nuclei", Priority: 1},
		},
	},
	{
		Name:         "ffuf",
		Description:  "Directory fuzzing",
		Category:     "web",
		CheckCommand: "ffuf -V",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/ffuf/ffuf", Priority: 1},
		},
	},
	{
		Name:         "gobuster",
		Description:  "Directory/DNS fuzzer",
		Category:     "web",
		CheckCommand: "gobuster version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/OJ/gobuster/v3", Priority: 1},
		},
	},
	{
		Name:         "wpscan",
		Description:  "WordPress scanner",
		Category:     "web",
		CheckCommand: "wpscan --version",
		InstallMethods: []InstallMethod{
			{Type: "docker", Command: "wpscanteam/wpscan", Priority: 1},
			{Type: "gem", Command: "wpscan", Priority: 2},
		},
	},
	{
		Name:         "metasploit",
		Description:  "Exploit framework",
		Category:     "postexploit",
		CheckCommand: "msfconsole --version",
		InstallMethods: []InstallMethod{
			{Type: "package_manager", Command: "metasploit-framework", Priority: 1},
			{Type: "binary", Command: "https://windows.metasploit.com/metasploitframework-latest.msi", Priority: 2},
		},
	},
	{
		Name:         "hydra",
		Description:  "Password brute force",
		Category:     "network",
		CheckCommand: "hydra -h",
		InstallMethods: []InstallMethod{
			{Type: "package_manager", Command: "hydra", Priority: 1},
		},
	},
	{
		Name:         "impacket",
		Description:  "Protocol exploitation",
		Category:     "network",
		CheckCommand: "python3 -c \"import impacket; print(impacket.__version__)\"",
		InstallMethods: []InstallMethod{
			{Type: "pip", Command: "impacket", Priority: 1},
		},
	},
	{
		Name:         "hashcat",
		Description:  "GPU password cracking",
		Category:     "network",
		CheckCommand: "hashcat --version",
		InstallMethods: []InstallMethod{
			{Type: "binary", Command: "https://hashcat.net/files/hashcat-6.2.6.7z", Priority: 1},
		},
	},
	{
		Name:         "spiderfoot",
		Description:  "OSINT automation",
		Category:     "osint",
		CheckCommand: "python3 -c \"import spiderfoot\"",
		InstallMethods: []InstallMethod{
			{Type: "git_clone", Command: "https://github.com/smicallef/spiderfoot.git", Priority: 1},
			{Type: "pip", Command: "spiderfoot", Priority: 2},
		},
	},
	{
		Name:         "reconftw",
		Description:  "Full recon pipeline",
		Category:     "osint",
		CheckCommand: "reconftw --help",
		InstallMethods: []InstallMethod{
			{Type: "git_clone", Command: "https://github.com/six2dez/reconftw.git", Priority: 1},
		},
	},
	{
		Name:         "uncover",
		Description:  "Search engine queries",
		Category:     "osint",
		CheckCommand: "uncover -version",
		InstallMethods: []InstallMethod{
			{Type: "go_install", Command: "github.com/projectdiscovery/uncover/cmd/uncover", Priority: 1},
		},
	},
}

func GetNVHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nexus-void")
}

func IsInstalled(toolName string) bool {
	for _, tool := range externalTools {
		if tool.Name == toolName {
			return checkTool(tool.CheckCommand)
		}
	}
	return false
}

func checkTool(command string) bool {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = os.Environ()
	err := cmd.Run()
	return err == nil
}

func InstallTool(toolName string) error {
	var toolDef *ToolDefinition
	for i := range externalTools {
		if externalTools[i].Name == toolName {
			toolDef = &externalTools[i]
			break
		}
	}
	if toolDef == nil {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	if IsInstalled(toolName) {
		fmt.Printf("  [OK] %s already installed\n", toolName)
		return nil
	}

	fmt.Printf("  [+] Installing %s...\n", toolName)

	for _, method := range toolDef.InstallMethods {
		if err := executeInstallMethod(toolDef.Name, method); err == nil {
			fmt.Printf("  [OK] %s installed via %s\n", toolName, method.Type)
			return nil
		} else {
			fmt.Printf("  [!] %s failed: %v\n", method.Type, err)
		}
	}

	return fmt.Errorf("all install methods failed for %s", toolName)
}

func InstallAll() error {
	fmt.Println("[+] Installing all external tools...")

	failures := 0
	for _, tool := range externalTools {
		if err := InstallTool(tool.Name); err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", tool.Name, err)
			failures++
		}
	}

	if failures > 0 {
		fmt.Printf("[!] %d/%d tools failed to install\n", failures, len(externalTools))
	} else {
		fmt.Println("[+] All external tools installed successfully")
	}
	return nil
}

func executeInstallMethod(toolName string, method InstallMethod) error {
	switch method.Type {
	case "go_install":
		return installViaGo(method.Command)
	case "pip":
		return installViaPip(method.Command)
	case "package_manager":
		return installViaPackageManager(method.Command)
	case "docker":
		return installViaDocker(toolName, method.Command)
	case "git_clone":
		return installViaGitClone(toolName, method.Command)
	case "binary":
		return installViaBinary(toolName, method.Command)
	default:
		return fmt.Errorf("unknown install method: %s", method.Type)
	}
}

func installViaGo(pkg string) error {
	cmd := exec.Command("go", "install", pkg)
	cmd.Env = append(os.Environ(), "GOBIN="+filepath.Join(GetNVHome(), "bin"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaPip(pkg string) error {
	var cmd *exec.Cmd
	if strings.HasPrefix(pkg, "git+") {
		cmd = exec.Command("pip3", "install", pkg)
	} else {
		cmd = exec.Command("pip3", "install", pkg)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaPackageManager(pkg string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Try winget, then chocolatey
		cmd = exec.Command("winget", "install", "--silent", pkg)
	case "darwin":
		cmd = exec.Command("brew", "install", pkg)
	default:
		// Debian/Ubuntu
		cmd = exec.Command("sudo", "apt-get", "install", "-y", pkg)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaDocker(toolName, image string) error {
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaGitClone(toolName, url string) error {
	targetDir := filepath.Join(GetNVHome(), "external_tools", toolName)

	// Check if already cloned
	if _, err := os.Stat(targetDir); err == nil {
		// Pull latest
		cmd := exec.Command("git", "-C", targetDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("git", "clone", url, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaBinary(toolName, url string) error {
	// Download binary from URL
	// This is simplified - real implementation would parse URL and extract
	fmt.Printf("    [>] Download binary from: %s\n", url)
	return fmt.Errorf("binary download requires manual implementation for %s", toolName)
}

func GetToolPath(toolName string) string {
	// Check in PATH first
	if path, err := exec.LookPath(toolName); err == nil {
		return path
	}

	// Check in our bin directory
	nvBin := filepath.Join(GetNVHome(), "bin")
	path := filepath.Join(nvBin, toolName)
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Check in external_tools directory
	externalDir := filepath.Join(GetNVHome(), "external_tools", toolName)
	if _, err := os.Stat(externalDir); err == nil {
		return externalDir
	}

	return ""
}
