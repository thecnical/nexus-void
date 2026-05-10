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
	// ETHERBREACH WiFi Arsenal
	{Name: "macchanger", Description: "MAC address randomization", Category: "network", CheckCommand: "macchanger --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "macchanger", Priority: 1}}},
	{Name: "mdk4", Description: "Advanced WiFi testing (deauth/beacon/probe)", Category: "network", CheckCommand: "mdk4 --help", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/aircrack-ng/mdk4.git,make", Priority: 1}}},
	{Name: "bettercap", Description: "Modern MITM and packet sniffing", Category: "network", CheckCommand: "bettercap --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/bettercap/bettercap@latest", Priority: 1}}},
	{Name: "hcxdumptool", Description: "PMKID capture tool", Category: "network", CheckCommand: "hcxdumptool --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/ZerBea/hcxdumptool.git,make", Priority: 1}}},
	{Name: "hcxtools", Description: "Hash conversion for hashcat", Category: "network", CheckCommand: "hcxpcapngtool --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/ZerBea/hcxtools.git,make", Priority: 1}}},
	{Name: "bully", Description: "WPS alternative attack", Category: "network", CheckCommand: "bully --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/aanarchyy/bully.git,make", Priority: 1}}},
	{Name: "pixiewps", Description: "WPS pixie dust offline attack", Category: "network", CheckCommand: "pixiewps --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/wiire-a/pixiewps.git,make", Priority: 1}}},
	{Name: "ettercap", Description: "ARP spoofing and MITM", Category: "network", CheckCommand: "ettercap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "ettercap-text-only", Priority: 1}}},
	{Name: "dnsmasq", Description: "DNS/DHCP for evil twin", Category: "network", CheckCommand: "dnsmasq --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "dnsmasq", Priority: 1}}},
	{Name: "hostapd", Description: "Rogue AP hosting", Category: "network", CheckCommand: "hostapd --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "hostapd", Priority: 1}}},
	{Name: "wifiphisher", Description: "Automated evil twin + karma + phishing", Category: "network", CheckCommand: "wifiphisher --version", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/wifiphisher/wifiphisher.git", Priority: 1}}},
	{Name: "fluxion", Description: "Evil twin + captive portal automation", Category: "network", CheckCommand: "fluxion --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/FluxionNetwork/fluxion.git", Priority: 1}}},
	{Name: "airgeddon", Description: "Multi-attack WiFi framework", Category: "network", CheckCommand: "airgeddon --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/v1s1t0r1sh3r3/airgeddon.git", Priority: 1}}},
	{Name: "mitmproxy", Description: "Transparent HTTP/HTTPS proxy", Category: "network", CheckCommand: "mitmproxy --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "mitmproxy", Priority: 1}}},
	{Name: "scapy", Description: "Python packet crafting", Category: "network", CheckCommand: "python3 -c 'import scapy'", InstallMethods: []InstallMethod{{Type: "pip", Command: "scapy", Priority: 1}}},
	{Name: "netdiscover", Description: "Network discovery", Category: "network", CheckCommand: "netdiscover -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "netdiscover", Priority: 1}}},
	{Name: "arp-scan", Description: "ARP scanner", Category: "network", CheckCommand: "arp-scan --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "arp-scan", Priority: 1}}},
	{Name: "enum4linux", Description: "SMB enumeration", Category: "network", CheckCommand: "enum4linux -h", InstallMethods: []InstallMethod{{Type: "apt", Command: "enum4linux", Priority: 1}}},
	{Name: "nikto", Description: "Web vulnerability scanner", Category: "network", CheckCommand: "nikto -Version", InstallMethods: []InstallMethod{{Type: "apt", Command: "nikto", Priority: 1}}},
	{Name: "whatweb", Description: "Web fingerprinting", Category: "network", CheckCommand: "whatweb --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "whatweb", Priority: 1}}},

	// --- MOBILE PENTEST (MOBILEBREACH) ---
	{Name: "adb", Description: "Android debug bridge", Category: "network", CheckCommand: "adb --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "adb", Priority: 1}}},
	{Name: "scrcpy", Description: "Android screen mirror and control", Category: "network", CheckCommand: "scrcpy --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "scrcpy", Priority: 1}}},
	{Name: "apktool", Description: "APK decompile and rebuild", Category: "network", CheckCommand: "apktool --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "apktool", Priority: 1}}},
	{Name: "jadx", Description: "Java decompiler for APK", Category: "network", CheckCommand: "jadx --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "jadx", Priority: 1}}},
	{Name: "libimobiledevice-utils", Description: "iOS device communication suite", Category: "network", CheckCommand: "ideviceinfo --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "libimobiledevice-utils", Priority: 1}}},
	{Name: "class-dump", Description: "Objective-C header extraction", Category: "network", CheckCommand: "class-dump --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/nygard/class-dump.git,make", Priority: 1}}},
	{Name: "frida-tools", Description: "Dynamic instrumentation for mobile", Category: "network", CheckCommand: "frida --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "frida-tools", Priority: 1}}},
	{Name: "objection", Description: "Frida mobile runtime wrapper", Category: "network", CheckCommand: "objection --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "objection", Priority: 1}}},
	{Name: "apkleaks", Description: "Hardcoded secret scanner for APK", Category: "network", CheckCommand: "apkleaks -h", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/dwisiswant0/apkleaks.git", Priority: 1}}},
	{Name: "quark-engine", Description: "Android malware scoring engine", Category: "network", CheckCommand: "quark --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "quark-engine", Priority: 1}}},
	{Name: "apk-mitm", Description: "Auto-patch APK certificate pinning", Category: "network", CheckCommand: "apk-mitm --version", InstallMethods: []InstallMethod{{Type: "npm", Command: "apk-mitm", Priority: 1}}},
	{Name: "house", Description: "Web GUI runtime mobile analysis (Frida)", Category: "network", CheckCommand: "house --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/nccgroup/house.git", Priority: 1}}},
	{Name: "rms", Description: "Runtime Mobile Security GUI", Category: "network", CheckCommand: "rms --help", InstallMethods: []InstallMethod{{Type: "git_npm", Command: "https://github.com/m0bilesecurity/RMS-Runtime-Mobile-Security.git", Priority: 1}}},
	{Name: "graphqlmap", Description: "GraphQL attack and introspection tool", Category: "network", CheckCommand: "graphqlmap --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/swisskyrepo/GraphQLmap.git", Priority: 1}}},
	{Name: "arjun", Description: "HTTP parameter discovery", Category: "network", CheckCommand: "arjun --version", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/s0md3v/Arjun.git", Priority: 1}}},
	{Name: "jwt_tool", Description: "JWT token security testing", Category: "network", CheckCommand: "jwt_tool --version", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/ticarpi/jwt_tool.git", Priority: 1}}},
	{Name: "PacketRusher", Description: "5G UE/gNodeB simulator", Category: "network", CheckCommand: "PacketRusher --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/HewlettPackard/PacketRusher.git,make", Priority: 1}}},
	{Name: "ueransim", Description: "5G UE and RAN simulator", Category: "network", CheckCommand: "nr-gnb --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/aligungr/UERANSIM.git,make", Priority: 1}}},
	{Name: "pySim", Description: "SIM and eSIM read/write", Category: "network", CheckCommand: "pySim-read --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/osmocom/pysim.git", Priority: 1}}},
	{Name: "gr-gsm", Description: "GSM SDR decoder", Category: "network", CheckCommand: "grgsm_scanner --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "gr-gsm", Priority: 1}}},
	{Name: "kalibrate-rtl", Description: "GSM frequency calibration for SDR", Category: "network", CheckCommand: "kal --help", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/steve-m/kalibrate-rtl.git,make", Priority: 1}}},
	{Name: "srsran", Description: "4G SDR stack", Category: "network", CheckCommand: "srsenb --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "srsran", Priority: 1}}},
	{Name: "rtl-sdr", Description: "RTL-SDR driver tools", Category: "network", CheckCommand: "rtl_test --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "rtl-sdr", Priority: 1}}},
	{Name: "hackrf", Description: "HackRF SDR tools", Category: "network", CheckCommand: "hackrf_info --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "hackrf", Priority: 1}}},
	{Name: "pcsc-tools", Description: "Smart card reader tools", Category: "network", CheckCommand: "pcsc_scan --help", InstallMethods: []InstallMethod{{Type: "apt", Command: "pcsc-tools", Priority: 1}}},

	// --- OSINT RECONNAISSANCE (OSINTBREACH) ---
	{Name: "amass", Description: "DNS enumeration and attack surface mapping", Category: "network", CheckCommand: "amass --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/owasp-amass/amass/v4/...@master", Priority: 1}}},
	{Name: "subfinder", Description: "Subdomain discovery via passive sources", Category: "network", CheckCommand: "subfinder --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest", Priority: 1}}},
	{Name: "assetfinder", Description: "Find domains and subdomains related to target", Category: "network", CheckCommand: "assetfinder --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/tomnomnom/assetfinder@latest", Priority: 1}}},
	{Name: "findomain", Description: "Fast cross-platform subdomain enumerator", Category: "network", CheckCommand: "findomain --version", InstallMethods: []InstallMethod{{Type: "git_build", Command: "https://github.com/findomain/findomain.git,cargo build --release", Priority: 1}}},
	{Name: "dnsx", Description: "Fast DNS resolver and brute forcer", Category: "network", CheckCommand: "dnsx --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/dnsx/cmd/dnsx@latest", Priority: 1}}},
	{Name: "naabu", Description: "Fast port scanner", Category: "network", CheckCommand: "naabu --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/naabu/v2/cmd/naabu@latest", Priority: 1}}},
	{Name: "waybackurls", Description: "Fetch known URLs from Wayback Machine", Category: "network", CheckCommand: "waybackurls --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/tomnomnom/waybackurls@latest", Priority: 1}}},
	{Name: "gau", Description: "GetAllUrls - fetch URLs from multiple sources", Category: "network", CheckCommand: "gau --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/lc/gau/v2/cmd/gau@latest", Priority: 1}}},
	{Name: "theHarvester", Description: "Email, subdomain and people OSINT gatherer", Category: "network", CheckCommand: "theHarvester --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/laramies/theHarvester.git", Priority: 1}}},
	{Name: "sherlock", Description: "Hunt social media accounts by username", Category: "network", CheckCommand: "sherlock --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/sherlock-project/sherlock.git", Priority: 1}}},
	{Name: "holehe", Description: "Check if email is registered on sites", Category: "network", CheckCommand: "holehe --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "holehe", Priority: 1}}},
	{Name: "h8mail", Description: "Email OSINT and password breach hunting", Category: "network", CheckCommand: "h8mail --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "h8mail", Priority: 1}}},
	{Name: "phoneinfoga", Description: "Advanced phone number OSINT scanner", Category: "network", CheckCommand: "phoneinfoga --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/sundowndev/phoneinfoga/v2@latest", Priority: 1}}},
	{Name: "gitleaks", Description: "Detect secrets in git repos", Category: "network", CheckCommand: "gitleaks version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/gitleaks/gitleaks@latest", Priority: 1}}},
	{Name: "trufflehog", Description: "Find secrets in code, binaries, containers", Category: "network", CheckCommand: "trufflehog --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/trufflesecurity/trufflehog@latest", Priority: 1}}},
	{Name: "osv-scanner", Description: "OSV dependency vulnerability scanner", Category: "network", CheckCommand: "osv-scanner --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/google/osv-scanner/cmd/osv-scanner@latest", Priority: 1}}},
	{Name: "trivy", Description: "Container, filesystem, repo vulnerability scanner", Category: "network", CheckCommand: "trivy --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "trivy", Priority: 1}}},
	{Name: "paramspider", Description: "Parameter discovery from web archives", Category: "network", CheckCommand: "paramspider --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/devanshbatham/ParamSpider.git", Priority: 1}}},

	// --- Web ---
	{Name: "sqlmap", Description: "SQL injection automation", Category: "web", CheckCommand: "sqlmap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "sqlmap", Priority: 1}}},
	{Name: "dalfox", Description: "XSS scanner", Category: "web", CheckCommand: "dalfox version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/hahwul/dalfox/v2@latest", Priority: 1}}},
	{Name: "nuclei", Description: "Vulnerability scanner", Category: "web", CheckCommand: "nuclei -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", Priority: 1}}},
	{Name: "katana", Description: "Fast web crawler", Category: "web", CheckCommand: "katana -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/katana/cmd/katana@latest", Priority: 1}}},
	{Name: "ffuf", Description: "Directory fuzzing", Category: "web", CheckCommand: "ffuf -V", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/ffuf/ffuf@latest", Priority: 1}}},
	{Name: "gobuster", Description: "Directory/DNS fuzzer", Category: "web", CheckCommand: "gobuster version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/OJ/gobuster/v3@latest", Priority: 1}}},
	{Name: "httpx", Description: "Fast HTTP prober", Category: "web", CheckCommand: "httpx -version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/projectdiscovery/httpx/cmd/httpx@latest", Priority: 1}}},
	{Name: "commix", Description: "Command injection", Category: "web", CheckCommand: "commix --version", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/commixproject/commix.git", Priority: 1}}},
	{Name: "xsstrike", Description: "XSS detection", Category: "web", CheckCommand: "xsstrike --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/s0md3v/XSStrike.git", Priority: 1}}},
	{Name: "wpscan", Description: "WordPress scanner", Category: "web", CheckCommand: "wpscan --version", InstallMethods: []InstallMethod{{Type: "gem", Command: "wpscan", Priority: 1}, {Type: "apt", Command: "wpscan", Priority: 2}}},

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

	// --- NETBREACH Tools ---
	{Name: "netexec", Description: "Network execution toolkit (NXC)", Category: "network", CheckCommand: "nxc --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "netexec", Priority: 1}}},
	{Name: "chisel", Description: "Fast TCP/UDP tunnel over HTTP", Category: "network", CheckCommand: "chisel --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/jpillora/chisel@latest", Priority: 1}}},
	{Name: "ligolo-ng", Description: "Advanced tunneling/pivoting tool", Category: "network", CheckCommand: "ligolo-ng --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/nicocha30/ligolo-ng@latest", Priority: 1}}},
	{Name: "sliver", Description: "Adversary simulation framework", Category: "network", CheckCommand: "sliver-client --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/BishopFox/sliver/client@latest", Priority: 1}}},
	{Name: "laZagne", Description: "Credentials recovery tool", Category: "network", CheckCommand: "laZagne --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/AlessandroZ/LaZagne.git", Priority: 1}}},
	{Name: "certipy", Description: "ADCS exploitation", Category: "network", CheckCommand: "certipy --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "certipy-ad", Priority: 1}}},
	{Name: "coercer", Description: "Coercion attack tester", Category: "network", CheckCommand: "coercer --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "coercer", Priority: 1}}},
	{Name: "bloodhound-python", Description: "BloodHound Python collector", Category: "network", CheckCommand: "bloodhound-python --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "bloodhound", Priority: 1}}},
	{Name: "dnscat2", Description: "DNS tunnel C2", Category: "network", CheckCommand: "dnscat2 --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/iagox86/dnscat2.git", Priority: 1}}},
	{Name: "icmpsh", Description: "ICMP shell reverse shell", Category: "network", CheckCommand: "icmpsh --help", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/bdamele/icmpsh.git", Priority: 1}}},

	// --- CRYPTOBREACH Tools ---
	{Name: "hashcat", Description: "World's fastest password cracker", Category: "crypto", CheckCommand: "hashcat --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "hashcat", Priority: 1}}},
	{Name: "john", Description: "John the Ripper password cracker", Category: "crypto", CheckCommand: "john --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "john", Priority: 1}}},
	{Name: "hashid", Description: "Hash type identifier", Category: "crypto", CheckCommand: "hashid --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "hashid", Priority: 1}}},
	{Name: "testssl.sh", Description: "TLS/SSL test tool", Category: "crypto", CheckCommand: "testssl.sh --version", InstallMethods: []InstallMethod{{Type: "git_clone", Command: "https://github.com/drwetter/testssl.sh.git", Priority: 1}}},
	{Name: "sslyze", Description: "SSL/TLS scanner", Category: "crypto", CheckCommand: "sslyze --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "sslyze", Priority: 1}}},
	{Name: "rsactftool", Description: "RSACTFTool for RSA attacks", Category: "crypto", CheckCommand: "RsaCtfTool --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/RsaCtfTool/RsaCtfTool.git", Priority: 1}}},
	{Name: "bettercap", Description: "Network attack and monitoring", Category: "crypto", CheckCommand: "bettercap --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "bettercap", Priority: 1}}},

	// --- CLOUDBREACH Tools ---
	{Name: "pacu", Description: "AWS exploitation framework", Category: "cloud", CheckCommand: "pacu --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "pacu", Priority: 1}}},
	{Name: "steampipe", Description: "SQL queries for cloud APIs", Category: "cloud", CheckCommand: "steampipe --version", InstallMethods: []InstallMethod{{Type: "apt", Command: "steampipe", Priority: 1}}},
	{Name: "cloudsplaining", Description: "IAM privilege escalation scanner", Category: "cloud", CheckCommand: "cloudsplaining --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "cloudsplaining", Priority: 1}}},
	{Name: "s3scanner", Description: "S3 bucket enumeration", Category: "cloud", CheckCommand: "s3scanner --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/sa7mon/S3Scanner@latest", Priority: 1}}},
	{Name: "peirates", Description: "K8s pentest tool", Category: "cloud", CheckCommand: "peirates --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/inguardians/peirates@latest", Priority: 1}}},
	{Name: "kube-hunter", Description: "K8s vulnerability scanner", Category: "cloud", CheckCommand: "kube-hunter --version", InstallMethods: []InstallMethod{{Type: "pip", Command: "kube-hunter", Priority: 1}}},
	{Name: "amicontained", Description: "Container capability checker", Category: "cloud", CheckCommand: "amicontained --help", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/genuinetools/amicontained@latest", Priority: 1}}},
	{Name: "cdk", Description: "Container escape toolkit", Category: "cloud", CheckCommand: "cdk --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/cdk-team/CDK@latest", Priority: 1}}},
	{Name: "enumerate-iam", Description: "IAM permission enumerator", Category: "cloud", CheckCommand: "enumerate-iam --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/andresriancho/enumerate-iam.git", Priority: 1}}},
	{Name: "cloud-nuke", Description: "Cloud resource destruction", Category: "cloud", CheckCommand: "cloud-nuke --version", InstallMethods: []InstallMethod{{Type: "go_install", Command: "github.com/gruntwork-io/cloud-nuke@latest", Priority: 1}}},
	{Name: "cloudmapper", Description: "AWS infrastructure mapper", Category: "cloud", CheckCommand: "cloudmapper --help", InstallMethods: []InstallMethod{{Type: "git_pip", Command: "https://github.com/duo-labs/cloudmapper.git", Priority: 1}}},
	{Name: "cartography", Description: "Cloud infrastructure mapper", Category: "cloud", CheckCommand: "cartography --help", InstallMethods: []InstallMethod{{Type: "pip", Command: "cartography", Priority: 1}}},
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
	// Create a temp module directory to avoid "current directory is not in a module" error
	tmpDir, err := os.MkdirTemp("", "nv-go-install-*")
	if err != nil {
		return fmt.Errorf("temp dir failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize a dummy go.mod so go install works from this directory
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module nvinstall\ngo 1.21\n"), 0644)

	cmd := exec.Command("go", "install", pkg)
	cmd.Dir = tmpDir
	env := os.Environ()
	env = append(env, "GOBIN="+gobin)
	env = append(env, "GO111MODULE=on")
	env = append(env, "GOPROXY=https://proxy.golang.org,direct")
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
	venvDir := filepath.Join(GetNVHome(), "venvs", pkg)
	os.MkdirAll(venvDir, 0755)

	pythonPath := filepath.Join(venvDir, "bin", "python3")
	pipPath := filepath.Join(venvDir, "bin", "pip")

	// Create venv if not exists (always recreate to avoid corruption)
	if _, err := os.Stat(pythonPath); err == nil {
		os.RemoveAll(venvDir)
	}
	cmd := exec.Command("python3", "-m", "venv", venvDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback: try with --without-pip and bootstrap
		cmd = exec.Command("python3", "-m", "venv", "--without-pip", venvDir)
		cmd.Run()
		// Bootstrap pip
		bootstrapPip(pythonPath)
	}

	// Ensure pip exists
	if _, err := os.Stat(pipPath); os.IsNotExist(err) {
		bootstrapPip(pythonPath)
	}

	// Install package into venv
	cmd = exec.Command(pythonPath, "-m", "pip", "install", "--upgrade", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback for Kali/Debian: use --break-system-packages
		cmd = exec.Command(pythonPath, "-m", "pip", "install", "--upgrade", "--break-system-packages", pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip install failed: %v", err)
		}
	}

	// Symlink all binaries from venv/bin to /usr/local/bin
	venvBinDir := filepath.Join(venvDir, "bin")
	entries, _ := os.ReadDir(venvBinDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
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

func bootstrapPip(pythonPath string) {
	// Download and run get-pip.py to bootstrap pip in the venv
	getPipURL := "https://bootstrap.pypa.io/get-pip.py"
	getPipFile := filepath.Join(os.TempDir(), "get-pip.py")
	resp, err := http.Get(getPipURL)
	if err == nil {
		defer resp.Body.Close()
		f, _ := os.Create(getPipFile)
		if f != nil {
			io.Copy(f, resp.Body)
			f.Close()
			exec.Command(pythonPath, getPipFile).Run()
		}
	}
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
