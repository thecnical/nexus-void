package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
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
	Category       string
	InstallMethods []InstallMethod
	CheckCommand   string
}

type InstallMethod struct {
	Type     string
	Command  string
	Priority int
}

var externalTools = []ToolDefinition{
	// --- Network ---
	{Name: "nmap", Description: "Network port scanner", Category: "network", CheckCommand: "nmap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "nmap", Priority: 1}}},
	{Name: "masscan", Description: "Ultra-fast port scanner", Category: "network", CheckCommand: "masscan --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/robertdavidgraham/masscan.git,make,-j", Priority: 1}, {Type: "apt", Command: "masscan", Priority: 2}}},
	{Name: "zmap", Description: "Internet-wide scanner", Category: "network", CheckCommand: "zmap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "zmap", Priority: 1}}},
	{Name: "naabu", Description: "Fast port scanner", Category: "network", CheckCommand: "naabu -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/naabu/v2/cmd/naabu@latest", Priority: 1}}},
	{Name: "rustscan", Description: "Modern port scanner", Category: "network", CheckCommand: "rustscan --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "rustscan", Priority: 1}}},
	{Name: "dnscan", Description: "DNS subdomain scanner", Category: "network", CheckCommand: "dnscan --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/rbsec/dnscan.git", Priority: 1}}},
	{Name: "hydra", Description: "Password brute force", Category: "network", CheckCommand: "hydra -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "hydra", Priority: 1}}},
	{Name: "john", Description: "Password cracker", Category: "network", CheckCommand: "john --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "john", Priority: 1}}},
	{Name: "hashcat", Description: "GPU password cracking", Category: "network", CheckCommand: "hashcat --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "hashcat", Priority: 1}}},
	{Name: "aircrack-ng", Description: "Wireless auditing", Category: "network", CheckCommand: "aircrack-ng --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "aircrack-ng", Priority: 1}}},
	{Name: "reaver", Description: "WPS attack tool", Category: "network", CheckCommand: "reaver -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "reaver", Priority: 1}}},
	{Name: "wifite", Description: "Automated wireless auditor", Category: "network", CheckCommand: "wifite --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "wifite", Priority: 1}}},
	{Name: "netdiscover", Description: "Network discovery", Category: "network", CheckCommand: "netdiscover -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "netdiscover", Priority: 1}}},
	{Name: "arp-scan", Description: "ARP scanner", Category: "network", CheckCommand: "arp-scan --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "arp-scan", Priority: 1}}},
	{Name: "enum4linux", Description: "SMB enumeration", Category: "network", CheckCommand: "enum4linux -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "enum4linux", Priority: 1}}},
	{Name: "nikto", Description: "Web vulnerability scanner", Category: "network", CheckCommand: "nikto -Version", InstallMethods: []InstallMethod{{Type: "apt", Command: "nikto", Priority: 1}}},
	{Name: "whatweb", Description: "Web fingerprinting", Category: "network", CheckCommand: "whatweb --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "whatweb", Priority: 1}}},

	// --- Web ---
	{Name: "sqlmap", Description: "SQL injection automation", Category: "web", CheckCommand: "sqlmap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "sqlmap", Priority: 1}}},
	{Name: "dalfox", Description: "XSS scanner", Category: "web", CheckCommand: "dalfox version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/hahwul/dalfox/v2@latest", Priority: 1}}},
	{Name: "nuclei", Description: "Vulnerability scanner", Category: "web", CheckCommand: "nuclei -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", Priority: 1}}},
	{Name: "katana", Description: "Fast web crawler", Category: "web", CheckCommand: "katana -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/katana/cmd/katana@latest", Priority: 1}}},
	{Name: "ffuf", Description: "Directory fuzzing", Category: "web", CheckCommand: "ffuf -V", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/ffuf/ffuf@latest", Priority: 1}}},
	{Name: "gobuster", Description: "Directory/DNS fuzzer", Category: "web", CheckCommand: "gobuster version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/OJ/gobuster/v3@latest", Priority: 1}}},
	{Name: "gau", Description: "URL extraction", Category: "web", CheckCommand: "gau --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/lc/gau/v2/cmd/gau@latest", Priority: 1}}},
	{Name: "httpx", Description: "Fast HTTP prober", Category: "web", CheckCommand: "httpx -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/httpx/cmd/httpx@latest", Priority: 1}}},
	{Name: "paramspider", Description: "Parameter discovery", Category: "web", CheckCommand: "paramspider --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/devanshbatham/ParamSpider.git", Priority: 1}}},
	{Name: "commix", Description: "Command injection", Category: "web", CheckCommand: "commix --version", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/commixproject/commix.git", Priority: 1}}},
	{Name: "xsstrike", Description: "XSS detection", Category: "web", CheckCommand: "xsstrike --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/s0md3v/XSStrike.git", Priority: 1}}},
	{Name: "wpscan", Description: "WordPress scanner", Category: "web", CheckCommand: "wpscan --version", InstallMethods: []InstallMethod{{Type: "gem", Command: "wpscan", Priority: 1}, {Type: "apt", Command: "wpscan", Priority: 2}}},

	// --- OSINT ---
	{Name: "amass", Description: "Subdomain enumeration", Category: "osint", CheckCommand: "amass -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/owasp-amass/amass/v4/...@master", Priority: 1}}},
	{Name: "subfinder", Description: "Subdomain discovery", Category: "osint", CheckCommand: "subfinder -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest", Priority: 1}}},
	{Name: "uncover", Description: "Search engine queries", Category: "osint", CheckCommand: "uncover -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/uncover/cmd/uncover@latest", Priority: 1}}},
	{Name: "theHarvester", Description: "Email/subdomain harvester", Category: "osint", CheckCommand: "theHarvester --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/laramies/theHarvester.git", Priority: 1}, {Type: "apt", Command: "theharvester", Priority: 2}}},
	{Name: "spiderfoot", Description: "OSINT automation", Category: "osint", CheckCommand: "spiderfoot --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/smicallef/spiderfoot.git", Priority: 1}}},
	{Name: "reconftw", Description: "Full recon pipeline", Category: "osint", CheckCommand: "reconftw --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/six2dez/reconftw.git", Priority: 1}}},
	{Name: "sherlock", Description: "Username OSINT", Category: "osint", CheckCommand: "sherlock --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/sherlock-project/sherlock.git", Priority: 1}}},

	// --- Post-Exploitation ---
	{Name: "metasploit", Description: "Exploit framework", Category: "postexploit", CheckCommand: "msfconsole --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "metasploit-framework", Priority: 1}}},
	{Name: "impacket", Description: "Protocol exploitation", Category: "postexploit", CheckCommand: "python3 -c \"import impacket; print(impacket.__version__)\"", InstallMethods: []InstallMethod{{Type: "apt", Command: "python3-impacket", Priority: 1}, {Type: "pip", Command: "impacket", Priority: 2}}},
	{Name: "bloodhound", Description: "AD visualization", Category: "postexploit", CheckCommand: "bloodhound --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "bloodhound", Priority: 1}}},
	{Name: "mimikatz", Description: "Credential extraction", Category: "postexploit", CheckCommand: "mimikatz --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/gentilkiwi/mimikatz.git", Priority: 1}}},
	{Name: "crackmapexec", Description: "SMB/WinRM pentest", Category: "postexploit", CheckCommand: "crackmapexec --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "crackmapexec", Priority: 1}}},
	{Name: "powershell-empire", Description: "Post-exploitation agent", Category: "postexploit", CheckCommand: "powershell-empire --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "powershell-empire", Priority: 1}}},
	{Name: "responder", Description: "LLMNR/NBT-NS poisoner", Category: "postexploit", CheckCommand: "responder --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "responder", Priority: 1}}},

	// --- Cloud ---
	{Name: "prowler", Description: "AWS security audit", Category: "cloud", CheckCommand: "prowler --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "prowler", Priority: 1}}},
	{Name: "cloudsploit", Description: "Cloud security scanner", Category: "cloud", CheckCommand: "cloudsploit --help", InstallMethods: []InstallMethod{{Type: "git_npm", Command: "https://github.com/aquasecurity/cloudsploit.git", Priority: 1}}},
	{Name: "scoutsuite", Description: "Multi-cloud audit", Category: "cloud", CheckCommand: "scout --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "scoutsuite", Priority: 1}}},

	// --- Active Directory ---
	{Name: "ldapdomaindump", Description: "LDAP domain dump", Category: "ad", CheckCommand: "ldapdomaindump --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "ldapdomaindump", Priority: 1}}},
	{Name: "kerbrute", Description: "Kerberos brute force", Category: "ad", CheckCommand: "kerbrute --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/ropnop/kerbrute@latest", Priority: 1}}},

	// --- Social Engineering ---
	{Name: "setoolkit", Description: "Social engineering toolkit", Category: "social", CheckCommand: "setoolkit --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "set", Priority: 1}, {Type: "git_clone", Command: "https://github.com/trustedsec/social-engineer-toolkit.git", Priority: 2}}},
}

func GetNVHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nexus-void")
}

func nvBin() string {
	return filepath.Join(GetNVHome(), "bin")
}

func ensureDirs() {
	os.MkdirAll(nvBin(), 0755)
	os.MkdirAll(filepath.Join(GetNVHome(), "external_tools"), 0755)
}

func IsInstalled(toolName string) bool {
	// Fast PATH check first
	if _, err := exec.LookPath(toolName); err == nil {
		return true
	}
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
	out, err := cmd.CombinedOutput()
	return err == nil || len(out) > 0
}

func InstallTool(toolName string) error {
	ensureDirs()
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
	ensureDirs()
	fmt.Println("[+] Installing all external tools...")
	fmt.Println("[+] This will use apt, go, pip, and git as needed.")

	// First install common build deps
	installBuildDeps()

	failures := 0
	for _, tool := range externalTools {
		if err := InstallTool(tool.Name); err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", tool.Name, err)
			failures++
		}
	}
	fmt.Println()
	if failures > 0 {
		fmt.Printf("[!] %d/%d tools failed to install\n", failures, len(externalTools))
		fmt.Println("[!] Some tools may need manual setup. Run: nexus-void arsenal install <tool>")
	} else {
		fmt.Println("[+] All external tools installed successfully")
	}
	return nil
}

func installBuildDeps() {
	fmt.Println("[*] Installing common build dependencies...")
	// Update package list first
	exec.Command("sudo", "apt-get", "update").Run()
	deps := []string{
		"build-essential", "git", "curl", "wget", "unzip", "p7zip-full",
		"python3-pip", "python3-dev", "python3-venv",
		"libpcap-dev", "libssl-dev", "libffi-dev", "libxml2-dev", "libxslt1-dev",
		"zlib1g-dev", "ruby", "ruby-dev", "npm", "nodejs",
	}
	args := append([]string{"apt-get", "install", "-y", "--no-install-recommends"}, deps...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  [!] Some deps may be missing: %v\n", err)
	}
}

func executeInstallMethod(toolName string, method InstallMethod) error {
	switch method.Type {
	case "go_install":
		return installViaGo(method.Command)
	case "pip":
		return installViaPip(method.Command)
	case "apt":
		return installViaApt(method.Command)
	case "gem":
		return installViaGem(method.Command)
	case "docker":
		return installViaDocker(toolName, method.Command)
	case "git_clone":
		return installViaGitClone(toolName, method.Command)
	case "git_pip":
		return installViaGitPip(toolName, method.Command)
	case "git_npm":
		return installViaGitNpm(toolName, method.Command)
	case "git_build":
		parts := strings.Split(method.Command, ",")
		if len(parts) < 2 {
			return fmt.Errorf("git_build needs url,buildcmd")
		}
		return installViaGitBuild(toolName, parts[0], parts[1])
	case "binary":
		return installViaBinary(toolName, method.Command)
	default:
		return fmt.Errorf("unknown install method: %s", method.Type)
	}
}

func installViaGo(pkg string) error {
	gobin := nvBin()
	cmd := exec.Command("go", "install", pkg)
	env := os.Environ()
	env = append(env, "GOBIN="+gobin)
	env = append(env, "GO111MODULE=on")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Symlink to /usr/local/bin so it's in PATH
	binName := filepath.Base(strings.Split(pkg, "@")[0])
	if i := strings.LastIndex(binName, "/"); i >= 0 {
		binName = binName[i+1:]
	}
	src := filepath.Join(gobin, binName)
	if runtime.GOOS == "windows" {
		src += ".exe"
	}
	if _, err := os.Stat(src); err == nil {
		exec.Command("sudo", "ln", "-sf", src, "/usr/local/bin/"+binName).Run()
	}
	return nil
}

func installViaPip(pkg string) error {
	// Use a per-package virtualenv to bypass Kali's externally-managed-environment restriction
	venvDir := filepath.Join(GetNVHome(), "venvs", pkg)
	os.MkdirAll(venvDir, 0755)

	// Create venv if not exists
	pythonPath := filepath.Join(venvDir, "bin", "python3")
	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		cmd := exec.Command("python3", "-m", "venv", venvDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("venv creation failed: %v", err)
		}
	}

	// Install package into venv
	cmd := exec.Command(pythonPath, "-m", "pip", "install", "--upgrade", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install failed: %v", err)
	}

	// Symlink all binaries from venv/bin to /usr/local/bin
	venvBinDir := filepath.Join(venvDir, "bin")
	entries, _ := os.ReadDir(venvBinDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip venv internals
		if name == "python" || name == "python3" || name == "pip" || name == "pip3" || name == "activate" || strings.HasSuffix(name, ".pc") {
			continue
		}
		src := filepath.Join(venvBinDir, name)
		info, _ := os.Stat(src)
		if info != nil && info.Mode()&0111 != 0 {
			exec.Command("sudo", "ln", "-sf", src, "/usr/local/bin/"+name).Run()
		}
	}
	return nil
}

func installViaApt(pkg string) error {
	cmd := exec.Command("sudo", "apt-get", "install", "-y", "--no-install-recommends", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Some packages may need apt-get update first
		exec.Command("sudo", "apt-get", "update").Run()
		cmd = exec.Command("sudo", "apt-get", "install", "-y", "--no-install-recommends", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

func installViaGem(pkg string) error {
	cmd := exec.Command("sudo", "gem", "install", pkg)
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
	if _, err := os.Stat(targetDir); err == nil {
		cmd := exec.Command("git", "-C", targetDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("git", "clone", "--depth", "1", url, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaGitPip(toolName, url string) error {
	targetDir := filepath.Join(GetNVHome(), "external_tools", toolName)
	if _, err := os.Stat(targetDir); err != nil {
		cmd := exec.Command("git", "clone", "--depth", "1", url, targetDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		cmd := exec.Command("git", "-C", targetDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Create a dedicated virtualenv for this tool (bypasses Kali pip restriction)
	venvDir := filepath.Join(GetNVHome(), "venvs", toolName)
	os.MkdirAll(venvDir, 0755)
	pythonPath := filepath.Join(venvDir, "bin", "python3")
	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		cmd := exec.Command("python3", "-m", "venv", venvDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("venv creation failed: %v", err)
		}
	}

	// Install requirements into venv
	reqFile := filepath.Join(targetDir, "requirements.txt")
	if _, err := os.Stat(reqFile); err == nil {
		cmd := exec.Command(pythonPath, "-m", "pip", "install", "-r", reqFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Install if setup.py exists
	setupFile := filepath.Join(targetDir, "setup.py")
	if _, err := os.Stat(setupFile); err == nil {
		cmd := exec.Command(pythonPath, "-m", "pip", "install", targetDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Install if pyproject.toml exists
	pyprojectFile := filepath.Join(targetDir, "pyproject.toml")
	if _, err := os.Stat(pyprojectFile); err == nil {
		cmd := exec.Command(pythonPath, "-m", "pip", "install", targetDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Find main executable script and create a wrapper using the venv python
	candidates := []string{"main.py", toolName + ".py", "run.py", "cli.py", toolName, "sherlock.py"}
	// Also check the tool's actual name variations
	for _, c := range candidates {
		p := filepath.Join(targetDir, c)
		if _, err := os.Stat(p); err == nil {
			os.Chmod(p, 0755)
			wrapper := filepath.Join("/usr/local/bin", toolName)
			content := "#!/bin/bash\nexec " + pythonPath + " " + p + " \"$@\"\n"
			os.WriteFile(wrapper, []byte(content), 0755)
			return nil
		}
	}

	// Fallback: check if tool installed a binary in the venv
	venvBin := filepath.Join(venvDir, "bin", toolName)
	if _, err := os.Stat(venvBin); err == nil {
		exec.Command("sudo", "ln", "-sf", venvBin, "/usr/local/bin/"+toolName).Run()
		return nil
	}

	return nil
}

func installViaGitNpm(toolName, url string) error {
	targetDir := filepath.Join(GetNVHome(), "external_tools", toolName)
	if _, err := os.Stat(targetDir); err != nil {
		cmd := exec.Command("git", "clone", "--depth", "1", url, targetDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		cmd := exec.Command("git", "-C", targetDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	cmd := exec.Command("npm", "install")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installViaGitBuild(toolName, url, buildCmd string) error {
	targetDir := filepath.Join(GetNVHome(), "external_tools", toolName)
	if _, err := os.Stat(targetDir); err != nil {
		cmd := exec.Command("git", "clone", "--depth", "1", url, targetDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		cmd := exec.Command("git", "-C", targetDir, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
	// Run build command
	cmd := exec.Command("bash", "-c", buildCmd)
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	// Find built binary and symlink
	entries, _ := os.ReadDir(targetDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == toolName || strings.Contains(name, toolName) {
			p := filepath.Join(targetDir, name)
			info, _ := os.Stat(p)
			if info != nil && info.Mode()&0111 != 0 {
				exec.Command("sudo", "ln", "-sf", p, "/usr/local/bin/"+toolName).Run()
				return nil
			}
		}
	}
	return nil
}

func installViaBinary(toolName, url string) error {
	// For apt-based tools that have no other method
	return installViaApt(toolName)
}

func GetToolPath(toolName string) string {
	if path, err := exec.LookPath(toolName); err == nil {
		return path
	}
	nvBinDir := nvBin()
	path := filepath.Join(nvBinDir, toolName)
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	externalDir := filepath.Join(GetNVHome(), "external_tools", toolName)
	if _, err := os.Stat(externalDir); err == nil {
		return externalDir
	}
	return ""
}

// DownloadFile downloads a file from URL to local path
func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractTarGz extracts a .tar.gz file
func ExtractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, os.FileMode(header.Mode))
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			out, _ := os.Create(target)
			io.Copy(out, tr)
			out.Close()
			os.Chmod(target, os.FileMode(header.Mode))
		}
	}
	return nil
}

// ExtractZip extracts a .zip file
func ExtractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, f.Mode())
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		out, _ := os.Create(target)
		rc, _ := f.Open()
		io.Copy(out, rc)
		out.Close()
		rc.Close()
		os.Chmod(target, f.Mode())
	}
	return nil
}
