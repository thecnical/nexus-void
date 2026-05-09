package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ToolCategory organizes tools by domain
type ToolCategory string

const (
	CatWebCrawling ToolCategory = "Web Crawling"
	CatWebApp      ToolCategory = "Web Application"
	CatNetwork     ToolCategory = "Network"
	CatCloud       ToolCategory = "Cloud"
	CatWireless    ToolCategory = "Wireless & Mobile"
	CatCrypto      ToolCategory = "Cryptography"
	CatOSINT       ToolCategory = "OSINT"
	CatPostExploit ToolCategory = "Post-Exploitation"
	CatSocialEng   ToolCategory = "Social Engineering"
	CatSupplyChain ToolCategory = "Supply Chain"
	CatHardware    ToolCategory = "Hardware"
	CatTelecom     ToolCategory = "Telecom"
	CatAD          ToolCategory = "Active Directory"
	CatPurple      ToolCategory = "Purple Team"
	CatML          ToolCategory = "ML & AI"
	CatC2          ToolCategory = "C2 & Implant"
)

// Tool represents a single tool in the arsenal
type Tool struct {
	Name        string                              `json:"name"`
	Category    ToolCategory                        `json:"category"`
	Description string                              `json:"description"`
	Internal    bool                                `json:"internal"` // true = Go code, false = external binary
	Command     string                              `json:"command"`  // External command name
	GoFunc      func(args []string) (string, error) `json:"-"`
	Installed   bool                                `json:"installed"`
}

// Registry holds all tools
type Registry struct {
	tools map[string]*Tool
}

// NewRegistry creates the full 189-tool registry
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]*Tool)}
	r.registerAll()
	return r
}

func (r *Registry) registerAll() {
	// === WEB CRAWLING (12 tools) ===
	r.add(&Tool{Name: "CRAWLER-X", Category: CatWebCrawling, Description: "Deep web crawler with JavaScript rendering", Internal: true})
	r.add(&Tool{Name: "JS-LINK-FINDER", Category: CatWebCrawling, Description: "Extract endpoints from JavaScript files", Internal: true})
	r.add(&Tool{Name: "FORM-HUNTER", Category: CatWebCrawling, Description: "Discover and map all forms", Internal: true})
	r.add(&Tool{Name: "PARAMETER-DISCOVERER", Category: CatWebCrawling, Description: "Find hidden parameters", Internal: true})
	r.add(&Tool{Name: "ARCHIVE-CRAWLER", Category: CatWebCrawling, Description: "Crawl Wayback Machine for endpoints", Internal: true, Command: "curl"})
	r.add(&Tool{Name: "SPA-CRAWLER", Category: CatWebCrawling, Description: "Single Page Application crawler", Internal: true})
	r.add(&Tool{Name: "RENDER-SPIDER", Category: CatWebCrawling, Description: "Headless browser spider", Internal: true})
	r.add(&Tool{Name: "COMMENT-HARVESTER", Category: CatWebCrawling, Description: "Extract comments from source code", Internal: true})
	r.add(&Tool{Name: "SOURCE-LEAK-FINDER", Category: CatWebCrawling, Description: "Find source code leaks (.git, .env, backup files)", Internal: true})
	r.add(&Tool{Name: "SITEMAP-ABUSER", Category: CatWebCrawling, Description: "Parse and abuse sitemap.xml", Internal: true})
	r.add(&Tool{Name: "ROBOTS-ABUSER", Category: CatWebCrawling, Description: "Extract hidden paths from robots.txt", Internal: true})
	r.add(&Tool{Name: "URL-PARAMETER-FUZZER", Category: CatWebCrawling, Description: "Fuzz parameters with wordlists", Internal: true})

	// === WEB APPLICATION (25 tools) ===
	r.add(&Tool{Name: "WEB-SPIDER", Category: CatWebApp, Description: "Smart web spidering", Internal: true})
	r.add(&Tool{Name: "API-DISCOVERER", Category: CatWebApp, Description: "Discover REST/GraphQL APIs", Internal: true})
	r.add(&Tool{Name: "SQLI-REAPER", Category: CatWebApp, Description: "SQL injection engine", Internal: true})
	r.add(&Tool{Name: "SQLI-SCHEMA-EXTRACTOR", Category: CatWebApp, Description: "Extract DB schema via SQLi", Internal: true})
	r.add(&Tool{Name: "SQLI-DATA-DUMPER", Category: CatWebApp, Description: "Dump data via SQLi", Internal: true})
	r.add(&Tool{Name: "XSS-HUNTER", Category: CatWebApp, Description: "Cross-site scripting finder", Internal: true})
	r.add(&Tool{Name: "XSS-POLYGLOT-GEN", Category: CatWebApp, Description: "Generate XSS polyglots", Internal: true})
	r.add(&Tool{Name: "CMD-INJECTOR", Category: CatWebApp, Description: "Command injection tester", Internal: true})
	r.add(&Tool{Name: "LFI-RAIDER", Category: CatWebApp, Description: "Local file inclusion", Internal: true})
	r.add(&Tool{Name: "LFI-TO-RCE", Category: CatWebApp, Description: "LFI to RCE exploitation", Internal: true})
	r.add(&Tool{Name: "RFI-PHANTOM", Category: CatWebApp, Description: "Remote file inclusion", Internal: true})
	r.add(&Tool{Name: "XXE-HARVESTER", Category: CatWebApp, Description: "XML external entity", Internal: true})
	r.add(&Tool{Name: "SSRF-LEECH", Category: CatWebApp, Description: "Server-side request forgery", Internal: true})
	r.add(&Tool{Name: "SSRF-BYPASS", Category: CatWebApp, Description: "SSRF bypass techniques", Internal: true})
	r.add(&Tool{Name: "JWT-BREAKER", Category: CatWebApp, Description: "JWT vulnerability tester", Internal: true})
	r.add(&Tool{Name: "JWT-FORGER", Category: CatWebApp, Description: "Forge JWT tokens", Internal: true})
	r.add(&Tool{Name: "GRAPHQL-PWN", Category: CatWebApp, Description: "GraphQL introspection and injection", Internal: true})
	r.add(&Tool{Name: "API-CHAOS", Category: CatWebApp, Description: "API fuzzing and abuse", Internal: true})
	r.add(&Tool{Name: "API-CHAIN-ATTACKER", Category: CatWebApp, Description: "Chain API vulnerabilities", Internal: true})
	r.add(&Tool{Name: "UPLOAD-ASSASSIN", Category: CatWebApp, Description: "File upload vulnerability", Internal: true})
	r.add(&Tool{Name: "CACHE-POISONER", Category: CatWebApp, Description: "Web cache poisoning", Internal: true})
	r.add(&Tool{Name: "DESERIAL-KILLER", Category: CatWebApp, Description: "Deserialization attacks", Internal: true})
	r.add(&Tool{Name: "TEMPLATE-INJECTOR", Category: CatWebApp, Description: "Server-side template injection", Internal: true})
	r.add(&Tool{Name: "WEBSOCKET-PWN", Category: CatWebApp, Description: "WebSocket exploitation", Internal: true})
	r.add(&Tool{Name: "CORS-BREAKER", Category: CatWebApp, Description: "CORS misconfiguration", Internal: true})

	// === NETWORK (15 tools) ===
	r.add(&Tool{Name: "PORT-BREACHER", Category: CatNetwork, Description: "Advanced port scanner", Internal: true})
	r.add(&Tool{Name: "PORT-SWEEPER", Category: CatNetwork, Description: "Fast mass port sweep", Internal: true})
	r.add(&Tool{Name: "SERVICE-PROBER", Category: CatNetwork, Description: "Service fingerprinting", Internal: true})
	r.add(&Tool{Name: "OS-FINGERPRINTER", Category: CatNetwork, Description: "OS detection", Internal: true, Command: "nmap"})
	r.add(&Tool{Name: "SMB-RAIDER", Category: CatNetwork, Description: "SMB exploitation", Internal: true, Command: "enum4linux"})
	r.add(&Tool{Name: "SMB-RELAYER", Category: CatNetwork, Description: "NTLM relay attacks", Internal: true})
	r.add(&Tool{Name: "LDAP-INJECTOR", Category: CatNetwork, Description: "LDAP injection", Internal: true})
	r.add(&Tool{Name: "DNS-PHANTOM", Category: CatNetwork, Description: "DNS enumeration", Internal: true})
	r.add(&Tool{Name: "DNS-BRUTEFORCER", Category: CatNetwork, Description: "Subdomain brute force", Internal: true})
	r.add(&Tool{Name: "SNMP-HARVESTER", Category: CatNetwork, Description: "SNMP data extraction", Internal: true, Command: "snmpwalk"})
	r.add(&Tool{Name: "VULN-SCANNER", Category: CatNetwork, Description: "Network vulnerability scan", Internal: true, Command: "nmap"})
	r.add(&Tool{Name: "BRUTE-FORGE", Category: CatNetwork, Description: "Protocol brute forcer", Internal: true})
	r.add(&Tool{Name: "BRUTE-OPTIMIZER", Category: CatNetwork, Description: "Smart credential spraying", Internal: true})
	r.add(&Tool{Name: "MITM-SHADOW", Category: CatNetwork, Description: "Man-in-the-middle", Internal: true})
	r.add(&Tool{Name: "PACKET-FORGE", Category: CatNetwork, Description: "Packet crafting", Internal: true})

	// === CLOUD (15 tools) ===
	r.add(&Tool{Name: "AWS-BREAKER", Category: CatCloud, Description: "AWS exploitation", Internal: true, Command: "aws"})
	r.add(&Tool{Name: "AWS-PRIVESC", Category: CatCloud, Description: "AWS privilege escalation", Internal: true})
	r.add(&Tool{Name: "AZURE-PHANTOM", Category: CatCloud, Description: "Azure testing", Internal: true})
	r.add(&Tool{Name: "AZURE-PRIVESC", Category: CatCloud, Description: "Azure privilege escalation", Internal: true})
	r.add(&Tool{Name: "GCP-RAIDER", Category: CatCloud, Description: "GCP exploitation", Internal: true})
	r.add(&Tool{Name: "GCP-PRIVESC", Category: CatCloud, Description: "GCP privilege escalation", Internal: true})
	r.add(&Tool{Name: "K8S-ASSASSIN", Category: CatCloud, Description: "Kubernetes attacks", Internal: true, Command: "kubectl"})
	r.add(&Tool{Name: "K8S-SECRET-HARVESTER", Category: CatCloud, Description: "Extract K8s secrets", Internal: true})
	r.add(&Tool{Name: "DOCKER-BREAKER", Category: CatCloud, Description: "Docker escape", Internal: true})
	r.add(&Tool{Name: "DOCKER-REGISTRY-ABUSER", Category: CatCloud, Description: "Docker registry abuse", Internal: true})
	r.add(&Tool{Name: "TERRAFORM-HARVESTER", Category: CatCloud, Description: "Terraform state extraction", Internal: true})
	r.add(&Tool{Name: "CLOUD-SHADOW", Category: CatCloud, Description: "Cloud persistence", Internal: true})
	r.add(&Tool{Name: "SERVERLESS-PWN", Category: CatCloud, Description: "Serverless exploitation", Internal: true})
	r.add(&Tool{Name: "CLOUDFRONT-BYPASS", Category: CatCloud, Description: "CDN bypass", Internal: true})
	r.add(&Tool{Name: "IAM-PRIVESC", Category: CatCloud, Description: "IAM privilege escalation", Internal: true})

	// === CRYPTO (8 tools) ===
	r.add(&Tool{Name: "HASH-CRACKER", Category: CatCrypto, Description: "Password hash cracking", Internal: true, Command: "hashcat"})
	r.add(&Tool{Name: "CERT-ABUSER", Category: CatCrypto, Description: "SSL/TLS certificate abuse", Internal: true})
	r.add(&Tool{Name: "KEY-HARVESTER", Category: CatCrypto, Description: "Extract cryptographic keys", Internal: true})
	r.add(&Tool{Name: "JWT-BREAKER-ADV", Category: CatCrypto, Description: "Advanced JWT attacks", Internal: true})
	r.add(&Tool{Name: "OPENSSL-SCANNER", Category: CatCrypto, Description: "OpenSSL vulnerability scan", Internal: true})
	r.add(&Tool{Name: "ENCRYPTION-BYPASS", Category: CatCrypto, Description: "Weak encryption detection", Internal: true})
	r.add(&Tool{Name: "RANDOM-FAILURE", Category: CatCrypto, Description: "PRNG weakness detection", Internal: true})
	r.add(&Tool{Name: "SIDE-CHANNEL-HUNTER", Category: CatCrypto, Description: "Side-channel analysis", Internal: true})

	// === OSINT (10 tools) ===
	r.add(&Tool{Name: "RECON-OMEGA", Category: CatOSINT, Description: "Full reconnaissance engine", Internal: true})
	r.add(&Tool{Name: "EMAIL-HUNTER", Category: CatOSINT, Description: "Email enumeration", Internal: true})
	r.add(&Tool{Name: "DOMAIN-MAPPER", Category: CatOSINT, Description: "Full domain mapping", Internal: true})
	r.add(&Tool{Name: "SOCIAL-SCANNER", Category: CatOSINT, Description: "Social media recon", Internal: true})
	r.add(&Tool{Name: "BREACH-CHECKER", Category: CatOSINT, Description: "Check breach databases", Internal: true})
	r.add(&Tool{Name: "GITHUB-HARVESTER", Category: CatOSINT, Description: "GitHub secret scanning", Internal: true})
	r.add(&Tool{Name: "SHODAN-SCANNER", Category: CatOSINT, Description: "Shodan integration", Internal: true})
	r.add(&Tool{Name: "PASTEBIN-HUNTER", Category: CatOSINT, Description: "Pastebin leak search", Internal: true})
	r.add(&Tool{Name: "WHOIS-ABUSER", Category: CatOSINT, Description: "WHOIS data extraction", Internal: true})
	r.add(&Tool{Name: "METADATA-EXTRACTOR", Category: CatOSINT, Description: "File metadata analysis", Internal: true})

	// === POST-EXPLOITATION (12 tools) ===
	r.add(&Tool{Name: "PRIVESC-FINDER", Category: CatPostExploit, Description: "Local privilege escalation", Internal: true})
	r.add(&Tool{Name: "LATERAL-MOVER", Category: CatPostExploit, Description: "Lateral movement", Internal: true})
	r.add(&Tool{Name: "PERSISTENCE-DAEMON", Category: CatPostExploit, Description: "Establish persistence", Internal: true})
	r.add(&Tool{Name: "CREDENTIAL-HARVESTER", Category: CatPostExploit, Description: "Dump credentials", Internal: true})
	r.add(&Tool{Name: "KEYLOGGER-PHANTOM", Category: CatPostExploit, Description: "Keylogging", Internal: true})
	r.add(&Tool{Name: "SCREEN-CAPTOR", Category: CatPostExploit, Description: "Screenshot capture", Internal: true})
	r.add(&Tool{Name: "AUDIO-CAPTOR", Category: CatPostExploit, Description: "Audio capture", Internal: true})
	r.add(&Tool{Name: "FILE-STEALER", Category: CatPostExploit, Description: "File exfiltration", Internal: true})
	r.add(&Tool{Name: "NETWORK-SNIFFER", Category: CatPostExploit, Description: "Network traffic capture", Internal: true})
	r.add(&Tool{Name: "PROCESS-INJECTOR", Category: CatPostExploit, Description: "Process injection", Internal: true})
	r.add(&Tool{Name: "ROOTKIT-PHANTOM", Category: CatPostExploit, Description: "Rootkit functionality", Internal: true})
	r.add(&Tool{Name: "C2-BEACON", Category: CatPostExploit, Description: "C2 communication", Internal: true})

	// === SOCIAL ENGINEERING (8 tools) ===
	r.add(&Tool{Name: "PHISHING-ENGINE", Category: CatSocialEng, Description: "Phishing campaign builder", Internal: true})
	r.add(&Tool{Name: "SPEAR-PHISHER", Category: CatSocialEng, Description: "Targeted spear phishing", Internal: true})
	r.add(&Tool{Name: "PRETEXT-MAKER", Category: CatSocialEng, Description: "Social pretext generation", Internal: true})
	r.add(&Tool{Name: "SMISHING-TOOL", Category: CatSocialEng, Description: "SMS phishing", Internal: true})
	r.add(&Tool{Name: "VISHING-TOOL", Category: CatSocialEng, Description: "Voice phishing", Internal: true})
	r.add(&Tool{Name: "QR-CODE-TRAP", Category: CatSocialEng, Description: "Malicious QR codes", Internal: true})
	r.add(&Tool{Name: "USB-DROP-ATTACK", Category: CatSocialEng, Description: "USB drop attack", Internal: true})
	r.add(&Tool{Name: "WATER-HOLE", Category: CatSocialEng, Description: "Watering hole attack", Internal: true})

	// === WIRELESS & MOBILE (8 tools) ===
	r.add(&Tool{Name: "WIFI-BREAKER", Category: CatWireless, Description: "WiFi cracking", Internal: true, Command: "aircrack-ng"})
	r.add(&Tool{Name: "BLUETOOTH-SNIFFER", Category: CatWireless, Description: "Bluetooth scanning", Internal: true})
	r.add(&Tool{Name: "RF-SNIFFER", Category: CatWireless, Description: "RF signal analysis", Internal: true})
	r.add(&Tool{Name: "ANDROID-ASSASSIN", Category: CatWireless, Description: "Android exploitation", Internal: true})
	r.add(&Tool{Name: "IOS-PHANTOM", Category: CatWireless, Description: "iOS exploitation", Internal: true})
	r.add(&Tool{Name: "APK-REVERSE", Category: CatWireless, Description: "APK reverse engineering", Internal: true, Command: "jadx"})
	r.add(&Tool{Name: "MOBILE-API-PWN", Category: CatWireless, Description: "Mobile API attacks", Internal: true})
	r.add(&Tool{Name: "BASEBAND-HUNTER", Category: CatWireless, Description: "Cellular baseband testing", Internal: true})

	// === SUPPLY CHAIN (6 tools) ===
	r.add(&Tool{Name: "DEPENDENCY-SCANNER", Category: CatSupplyChain, Description: "Vulnerable dependency finder", Internal: true})
	r.add(&Tool{Name: "TYPOSQUAT-HUNTER", Category: CatSupplyChain, Description: "Typosquatting detection", Internal: true})
	r.add(&Tool{Name: "REPO-POISONER", Category: CatSupplyChain, Description: "Repository poisoning", Internal: true})
	r.add(&Tool{Name: "CI-CD-BREAKER", Category: CatSupplyChain, Description: "CI/CD pipeline attacks", Internal: true})
	r.add(&Tool{Name: "SIGNATURE-FORGER", Category: CatSupplyChain, Description: "Code signature forgery", Internal: true})
	r.add(&Tool{Name: "PACKAGE-HIJACKER", Category: CatSupplyChain, Description: "Package manager hijacking", Internal: true})

	// === HARDWARE (5 tools) ===
	r.add(&Tool{Name: "JTAG-DEBUGGER", Category: CatHardware, Description: "JTAG debugging", Internal: true})
	r.add(&Tool{Name: "SPI-FLASHER", Category: CatHardware, Description: "SPI flash extraction", Internal: true})
	r.add(&Tool{Name: "I2C-SNIFFER", Category: CatHardware, Description: "I2C bus sniffing", Internal: true})
	r.add(&Tool{Name: "FIRMWARE-DUMPER", Category: CatHardware, Description: "Firmware extraction", Internal: true})
	r.add(&Tool{Name: "CHIP-DECAP", Category: CatHardware, Description: "Chip decapsulation guide", Internal: true})

	// === TELECOM (5 tools) ===
	r.add(&Tool{Name: "SS7-PHANTOM", Category: CatTelecom, Description: "SS7 protocol attacks", Internal: true})
	r.add(&Tool{Name: "DIAMETER-HUNTER", Category: CatTelecom, Description: "Diameter protocol attacks", Internal: true})
	r.add(&Tool{Name: "GTP-BREAKER", Category: CatTelecom, Description: "GTP tunnel exploitation", Internal: true})
	r.add(&Tool{Name: "SIP-CRACKER", Category: CatTelecom, Description: "SIP/VoIP cracking", Internal: true})
	r.add(&Tool{Name: "SMS-INTERCEPTOR", Category: CatTelecom, Description: "SMS interception", Internal: true})

	// === ACTIVE DIRECTORY (10 tools) ===
	r.add(&Tool{Name: "DOMAIN-DUMP", Category: CatAD, Description: "AD domain enumeration", Internal: true, Command: "ldapdomaindump"})
	r.add(&Tool{Name: "BLOODHOUND-PATH", Category: CatAD, Description: "BloodHound path analysis", Internal: true})
	r.add(&Tool{Name: "KERBEROAST-HARVESTER", Category: CatAD, Description: "Kerberoasting", Internal: true})
	r.add(&Tool{Name: "ASREPROAST-CRACKER", Category: CatAD, Description: "AS-REP Roasting", Internal: true})
	r.add(&Tool{Name: "DC-SYNC-PHANTOM", Category: CatAD, Description: "DCSync attack", Internal: true})
	r.add(&Tool{Name: "PASS-THE-HASH", Category: CatAD, Description: "Pass-the-Hash", Internal: true})
	r.add(&Tool{Name: "GOLDEN-TICKET", Category: CatAD, Description: "Golden Ticket creation", Internal: true})
	r.add(&Tool{Name: "SILVER-TICKET", Category: CatAD, Description: "Silver Ticket creation", Internal: true})
	r.add(&Tool{Name: "ACL-ABUSER", Category: CatAD, Description: "ACL abuse", Internal: true})
	r.add(&Tool{Name: "RBCD-ATTACKER", Category: CatAD, Description: "RBCD exploitation", Internal: true})

	// === PURPLE TEAM (8 tools) ===
	r.add(&Tool{Name: "DETECTION-TESTER", Category: CatPurple, Description: "Test detection capabilities", Internal: true})
	r.add(&Tool{Name: "SIGMA-RULE-MAKER", Category: CatPurple, Description: "Generate Sigma rules", Internal: true})
	r.add(&Tool{Name: "MITRE-MAPPER", Category: CatPurple, Description: "Map to MITRE ATT&CK", Internal: true})
	r.add(&Tool{Name: "BLUETEAM-TRAINER", Category: CatPurple, Description: "Blue team training", Internal: true})
	r.add(&Tool{Name: "HONEYPOT-DEPLOYER", Category: CatPurple, Description: "Deploy honeypots", Internal: true})
	r.add(&Tool{Name: "LOG-POISONER", Category: CatPurple, Description: "Test log analysis", Internal: true})
	r.add(&Tool{Name: "EDR-EVADER", Category: CatPurple, Description: "EDR evasion testing", Internal: true})
	r.add(&Tool{Name: "SIEM-STRESSER", Category: CatPurple, Description: "SIEM stress testing", Internal: true})

	// === ML & AI (8 tools) ===
	r.add(&Tool{Name: "PAYLOAD-EVOLVER", Category: CatML, Description: "Genetic payload evolution", Internal: true})
	r.add(&Tool{Name: "VULN-PREDICTOR", Category: CatML, Description: "AI vulnerability prediction", Internal: true})
	r.add(&Tool{Name: "WAF-BYPASS-AI", Category: CatML, Description: "AI WAF bypass", Internal: true})
	r.add(&Tool{Name: "THREAT-INTEL-AI", Category: CatML, Description: "AI threat intelligence", Internal: true})
	r.add(&Tool{Name: "TARGET-PROFILER", Category: CatML, Description: "AI target profiling", Internal: true})
	r.add(&Tool{Name: "AUTOMATION-ENGINE", Category: CatML, Description: "Workflow automation", Internal: true})
	r.add(&Tool{Name: "NLP-PHISHING", Category: CatML, Description: "AI phishing text", Internal: true})
	r.add(&Tool{Name: "ANOMALY-DETECTOR", Category: CatML, Description: "Behavioral anomaly detection", Internal: true})

	// === C2 & IMPLANT (10 tools) ===
	r.add(&Tool{Name: "BEACON-ENGINE", Category: CatC2, Description: "C2 beacon", Internal: true})
	r.add(&Tool{Name: "IMPLANT-MAKER", Category: CatC2, Description: "Implant builder", Internal: true})
	r.add(&Tool{Name: "STEGANO-ENCODER", Category: CatC2, Description: "Steganography encoder", Internal: true})
	r.add(&Tool{Name: "DNS-TUNNEL", Category: CatC2, Description: "DNS tunneling", Internal: true})
	r.add(&Tool{Name: "ICMP-SHELL", Category: CatC2, Description: "ICMP backdoor", Internal: true})
	r.add(&Tool{Name: "HTTPS-BEACON", Category: CatC2, Description: "HTTPS C2 channel", Internal: true})
	r.add(&Tool{Name: "WEBSOCKET-C2", Category: CatC2, Description: "WebSocket C2", Internal: true})
	r.add(&Tool{Name: "GRPC-TUNNEL", Category: CatC2, Description: "gRPC tunneling", Internal: true})
	r.add(&Tool{Name: "SMB-PIPE", Category: CatC2, Description: "SMB named pipe C2", Internal: true})
	r.add(&Tool{Name: "ALIVE-CHECKER", Category: CatC2, Description: "Implant heartbeat", Internal: true})

	// Wire all internal tools to their real Go functions
	r.wireGoFuncs()
}

func (r *Registry) wireGoFuncs() {
	// Web Crawling
	r.wire("CRAWLER-X", runCrawlerX)
	r.wire("JS-LINK-FINDER", runJSLinkFinder)
	r.wire("FORM-HUNTER", runFormHunter)
	r.wire("PARAMETER-DISCOVERER", runParameterDiscoverer)
	r.wire("ARCHIVE-CRAWLER", runArchiveCrawler)
	r.wire("SPA-CRAWLER", runSPACrawler)
	r.wire("RENDER-SPIDER", runRenderSpider)
	r.wire("COMMENT-HARVESTER", runCommentHarvester)
	r.wire("SOURCE-LEAK-FINDER", runSourceLeakFinder)
	r.wire("SITEMAP-ABUSER", runSitemapAbuser)
	r.wire("ROBOTS-ABUSER", runRobotsAbuser)
	r.wire("URL-PARAMETER-FUZZER", runURLParameterFuzzer)

	// Web Application
	r.wire("WEB-SPIDER", runWebSpider)
	r.wire("API-DISCOVERER", runAPIDiscoverer)
	r.wire("SQLI-REAPER", runSQLIReaper)
	r.wire("SQLI-SCHEMA-EXTRACTOR", runSQLISchemaExtractor)
	r.wire("SQLI-DATA-DUMPER", runSQLIDataDumper)
	r.wire("XSS-HUNTER", runXSSHunter)
	r.wire("XSS-POLYGLOT-GEN", runXSSPolyglotGen)
	r.wire("CMD-INJECTOR", runCmdInjector)
	r.wire("LFI-RAIDER", runLFIRaider)
	r.wire("LFI-TO-RCE", runLFIToRCE)
	r.wire("RFI-PHANTOM", runRFIPhantom)
	r.wire("XXE-HARVESTER", runXXEHarvester)
	r.wire("SSRF-LEECH", runSSRFLeeche)
	r.wire("SSRF-BYPASS", runSSRFBypass)
	r.wire("JWT-BREAKER", runJWTBreaker)
	r.wire("JWT-FORGER", runJWTForger)
	r.wire("GRAPHQL-PWN", runGraphQLPwn)
	r.wire("API-CHAOS", runAPIChaos)
	r.wire("API-CHAIN-ATTACKER", runAPIChainAttacker)
	r.wire("UPLOAD-ASSASSIN", runUploadAssassin)
	r.wire("CACHE-POISONER", runCachePoisoner)
	r.wire("DESERIAL-KILLER", runDeserialKiller)
	r.wire("TEMPLATE-INJECTOR", runTemplateInjector)
	r.wire("WEBSOCKET-PWN", runWebSocketPwn)
	r.wire("CORS-BREAKER", runCORSBreaker)

	// Network
	r.wire("PORT-BREACHER", runPortBreacher)
	r.wire("PORT-SWEEPER", runPortSweeper)
	r.wire("SERVICE-PROBER", runServiceProber)
	r.wire("OS-FINGERPRINTER", runOSFingerprinter)
	r.wire("SMB-RAIDER", runSMBRaider)
	r.wire("SMB-RELAYER", runSMBRelayer)
	r.wire("LDAP-INJECTOR", runLDAPInjector)
	r.wire("DNS-PHANTOM", runDNSPhantom)
	r.wire("DNS-BRUTEFORCER", runDNSBruteforcer)
	r.wire("SNMP-HARVESTER", runSNMPHarvester)
	r.wire("VULN-SCANNER", runVulnScanner)
	r.wire("BRUTE-FORGE", runBruteForge)
	r.wire("BRUTE-OPTIMIZER", runBruteOptimizer)
	r.wire("MITM-SHADOW", runMITMShadow)
	r.wire("PACKET-FORGE", runPacketForge)

	// Cloud
	r.wire("AWS-BREAKER", runAWSBreaker)
	r.wire("AWS-PRIVESC", runAWSPrivesc)
	r.wire("AZURE-PHANTOM", runAzurePhantom)
	r.wire("AZURE-PRIVESC", runAzurePrivesc)
	r.wire("GCP-RAIDER", runGCPRaider)
	r.wire("GCP-PRIVESC", runGCPPrivesc)
	r.wire("K8S-ASSASSIN", runK8SAssassin)
	r.wire("K8S-SECRET-HARVESTER", runK8SSecretHarvester)
	r.wire("DOCKER-BREAKER", runDockerBreaker)
	r.wire("DOCKER-REGISTRY-ABUSER", runDockerRegistryAbuser)
	r.wire("TERRAFORM-HARVESTER", runTerraformHarvester)
	r.wire("CLOUD-SHADOW", runCloudShadow)
	r.wire("SERVERLESS-PWN", runServerlessPwn)
	r.wire("CLOUDFRONT-BYPASS", runCloudFrontBypass)
	r.wire("IAM-PRIVESC", runIAMPrivesc)

	// Crypto
	r.wire("HASH-CRACKER", runHashCracker)
	r.wire("CERT-ABUSER", runCertAbuser)
	r.wire("KEY-HARVESTER", runKeyHarvester)
	r.wire("JWT-BREAKER-ADV", runJWTBreakerAdv)
	r.wire("OPENSSL-SCANNER", runOpenSSLScanner)
	r.wire("ENCRYPTION-BYPASS", runEncryptionBypass)
	r.wire("RANDOM-FAILURE", runRandomFailure)
	r.wire("SIDE-CHANNEL-HUNTER", runSideChannelHunter)

	// OSINT
	r.wire("RECON-OMEGA", runReconOmega)
	r.wire("EMAIL-HUNTER", runEmailHunter)
	r.wire("DOMAIN-MAPPER", runDomainMapper)
	r.wire("SOCIAL-SCANNER", runSocialScanner)
	r.wire("BREACH-CHECKER", runBreachChecker)
	r.wire("GITHUB-HARVESTER", runGitHubHarvester)
	r.wire("SHODAN-SCANNER", runShodanScanner)
	r.wire("PASTEBIN-HUNTER", runPastebinHunter)
	r.wire("WHOIS-ABUSER", runWHOISAbuser)
	r.wire("METADATA-EXTRACTOR", runMetadataExtractor)

	// Post-Exploit
	r.wire("PRIVESC-FINDER", runPrivescFinder)
	r.wire("LATERAL-MOVER", runLateralMover)
	r.wire("PERSISTENCE-DAEMON", runPersistenceDaemon)
	r.wire("CREDENTIAL-HARVESTER", runCredentialHarvester)
	r.wire("KEYLOGGER-PHANTOM", runKeyloggerPhantom)
	r.wire("SCREEN-CAPTOR", runScreenCaptor)
	r.wire("AUDIO-CAPTOR", runAudioCaptor)
	r.wire("FILE-STEALER", runFileStealer)
	r.wire("NETWORK-SNIFFER", runNetworkSniffer)
	r.wire("PROCESS-INJECTOR", runProcessInjector)
	r.wire("ROOTKIT-PHANTOM", runRootkitPhantom)
	r.wire("C2-BEACON", runC2Beacon)

	// Social Engineering
	r.wire("PHISHING-ENGINE", runPhishingEngine)
	r.wire("SPEAR-PHISHER", runSpearPhisher)
	r.wire("PRETEXT-MAKER", runPretextMaker)
	r.wire("SMISHING-TOOL", runSmishingTool)
	r.wire("VISHING-TOOL", runVishingTool)
	r.wire("QR-CODE-TRAP", runQRCodeTrap)
	r.wire("USB-DROP-ATTACK", runUSBDropAttack)
	r.wire("WATER-HOLE", runWaterHole)

	// Wireless
	r.wire("WIFI-BREAKER", runWiFiBreaker)
	r.wire("BLUETOOTH-SNIFFER", runBluetoothSniffer)
	r.wire("RF-SNIFFER", runRFSniffer)
	r.wire("ANDROID-ASSASSIN", runAndroidAssassin)
	r.wire("IOS-PHANTOM", runIOSPhantom)
	r.wire("APK-REVERSE", runAPKReverse)
	r.wire("MOBILE-API-PWN", runMobileAPIPwn)
	r.wire("BASEBAND-HUNTER", runBasebandHunter)

	// Supply Chain
	r.wire("DEPENDENCY-SCANNER", runDependencyScanner)
	r.wire("TYPOSQUAT-HUNTER", runTyposquatHunter)
	r.wire("REPO-POISONER", runRepoPoisoner)
	r.wire("CI-CD-BREAKER", runCICDBreaker)
	r.wire("SIGNATURE-FORGER", runSignatureForger)
	r.wire("PACKAGE-HIJACKER", runPackageHijacker)

	// Hardware
	r.wire("JTAG-DEBUGGER", runJTAGDebugger)
	r.wire("SPI-FLASHER", runSPIFlasher)
	r.wire("I2C-SNIFFER", runI2CSniffer)
	r.wire("FIRMWARE-DUMPER", runFirmwareDumper)
	r.wire("CHIP-DECAP", runChipDecap)

	// Telecom
	r.wire("SS7-PHANTOM", runSS7Phantom)
	r.wire("DIAMETER-HUNTER", runDiameterHunter)
	r.wire("GTP-BREAKER", runGTPBreaker)
	r.wire("SIP-CRACKER", runSIPCracker)
	r.wire("SMS-INTERCEPTOR", runSMSInterceptor)

	// Active Directory
	r.wire("DOMAIN-DUMP", runDomainDump)
	r.wire("BLOODHOUND-PATH", runBloodhoundPath)
	r.wire("KERBEROAST-HARVESTER", runKerberoastHarvester)
	r.wire("ASREPROAST-CRACKER", runASREPRoastCracker)
	r.wire("DC-SYNC-PHANTOM", runDCSyncPhantom)
	r.wire("PASS-THE-HASH", runPassTheHash)
	r.wire("GOLDEN-TICKET", runGoldenTicket)
	r.wire("SILVER-TICKET", runSilverTicket)
	r.wire("ACL-ABUSER", runACLAbuser)
	r.wire("RBCD-ATTACKER", runRBCDAttacker)

	// Purple Team
	r.wire("DETECTION-TESTER", runDetectionTester)
	r.wire("SIGMA-RULE-MAKER", runSigmaRuleMaker)
	r.wire("MITRE-MAPPER", runMITREMapper)
	r.wire("BLUETEAM-TRAINER", runBlueTeamTrainer)
	r.wire("HONEYPOT-DEPLOYER", runHoneypotDeployer)
	r.wire("LOG-POISONER", runLogPoisoner)
	r.wire("EDR-EVADER", runEDREvader)
	r.wire("SIEM-STRESSER", runSIEMStresser)

	// ML & AI
	r.wire("PAYLOAD-EVOLVER", runPayloadEvolver)
	r.wire("VULN-PREDICTOR", runVulnPredictor)
	r.wire("WAF-BYPASS-AI", runWAFBypassAI)
	r.wire("THREAT-INTEL-AI", runThreatIntelAI)
	r.wire("TARGET-PROFILER", runTargetProfiler)
	r.wire("AUTOMATION-ENGINE", runAutomationEngine)
	r.wire("NLP-PHISHING", runNLPPhishing)
	r.wire("ANOMALY-DETECTOR", runAnomalyDetector)

	// C2 & Implant
	r.wire("BEACON-ENGINE", runBeaconEngine)
	r.wire("IMPLANT-MAKER", runImplantMaker)
	r.wire("STEGANO-ENCODER", runSteganoEncoder)
	r.wire("DNS-TUNNEL", runDNSTunnel)
	r.wire("ICMP-SHELL", runICMPShell)
	r.wire("HTTPS-BEACON", runHTTPSBeacon)
	r.wire("WEBSOCKET-C2", runWebSocketC2)
	r.wire("GRPC-TUNNEL", runGRPCTunnel)
	r.wire("SMB-PIPE", runSMBPipe)
	r.wire("ALIVE-CHECKER", runAliveChecker)
}

func (r *Registry) wire(name string, fn func([]string) (string, error)) {
	if t, ok := r.tools[name]; ok {
		t.GoFunc = fn
	}
}

func (r *Registry) add(t *Tool) {
	// Check if external tool is installed
	if t.Command != "" && !t.Internal {
		_, err := exec.LookPath(t.Command)
		t.Installed = err == nil
	} else if t.Command != "" {
		_, err := exec.LookPath(t.Command)
		t.Installed = err == nil
	} else {
		t.Installed = true
	}
	r.tools[t.Name] = t
}

// List returns all tools
func (r *Registry) List() []*Tool {
	var list []*Tool
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

// ListByCategory returns tools in a category
func (r *Registry) ListByCategory(cat ToolCategory) []*Tool {
	var list []*Tool
	for _, t := range r.tools {
		if t.Category == cat {
			list = append(list, t)
		}
	}
	return list
}

// Get returns a single tool
func (r *Registry) Get(name string) (*Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// Count returns total tool count
func (r *Registry) Count() int {
	return len(r.tools)
}

// Categories returns all category names with counts
func (r *Registry) Categories() map[ToolCategory]int {
	counts := make(map[ToolCategory]int)
	for _, t := range r.tools {
		counts[t.Category]++
	}
	return counts
}

// RunExternal executes an external tool with 5 minute timeout
func RunExternal(command string, args ...string) (string, error) {
	cctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cctx, "cmd", append([]string{"/c", command}, args...)...)
	} else {
		cmd = exec.CommandContext(cctx, command, args...)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunExternalSafe checks if tool exists before running
func RunExternalSafe(command string, args ...string) (string, error) {
	if _, err := exec.LookPath(command); err != nil {
		return "", fmt.Errorf("tool not installed: %s", command)
	}
	return RunExternal(command, args...)
}

// RunNmap executes nmap with given targets and options
func RunNmap(targets string, opts ...string) (string, error) {
	args := append([]string{"-sV", "-T4"}, opts...)
	args = append(args, strings.Split(targets, " ")...)
	return RunExternalSafe("nmap", args...)
}

// RunSQLMap executes sqlmap against a target
func RunSQLMap(target string, opts ...string) (string, error) {
	args := append([]string{"-u", target, "--batch", "--random-agent"}, opts...)
	return RunExternalSafe("sqlmap", args...)
}

// RunNuclei executes nuclei against targets
func RunNuclei(targets string, opts ...string) (string, error) {
	args := append([]string{"-u", targets, "-stats"}, opts...)
	return RunExternalSafe("nuclei", args...)
}

// RunGobuster executes gobuster directory scan
func RunGobuster(url, wordlist string, opts ...string) (string, error) {
	args := append([]string{"dir", "-u", url, "-w", wordlist}, opts...)
	return RunExternalSafe("gobuster", args...)
}

// RunSubfinder executes subfinder for subdomain enumeration
func RunSubfinder(domain string, opts ...string) (string, error) {
	args := append([]string{"-d", domain}, opts...)
	return RunExternalSafe("subfinder", args...)
}

// RunAmass executes amass for subdomain enumeration
func RunAmass(domain string, opts ...string) (string, error) {
	args := append([]string{"enum", "-d", domain}, opts...)
	return RunExternalSafe("amass", args...)
}

// RunFFUF executes ffuf for fuzzing
func RunFFUF(target, wordlist string, opts ...string) (string, error) {
	args := append([]string{"-u", target, "-w", wordlist}, opts...)
	return RunExternalSafe("ffuf", args...)
}

// RunGitLeaks executes gitleaks for secret scanning
func RunGitLeaks(repo string, opts ...string) (string, error) {
	args := append([]string{"detect", "-s", repo}, opts...)
	return RunExternalSafe("gitleaks", args...)
}

// RunNikto executes nikto web scanner
func RunNikto(target string, opts ...string) (string, error) {
	args := append([]string{"-h", target}, opts...)
	return RunExternalSafe("nikto", args...)
}

// RunWPScan executes wpscan for WordPress
func RunWPScan(url string, opts ...string) (string, error) {
	args := append([]string{"--url", url, "--no-update"}, opts...)
	return RunExternalSafe("wpscan", args...)
}

// RunMasscan executes masscan for fast port scanning
func RunMasscan(targets string, opts ...string) (string, error) {
	args := append([]string{"-p1-65535", "--rate", "10000"}, opts...)
	args = append(args, strings.Split(targets, " ")...)
	return RunExternalSafe("masscan", args...)
}

// RunJohn executes john the ripper
func RunJohn(hashFile string, opts ...string) (string, error) {
	args := append([]string{hashFile}, opts...)
	return RunExternalSafe("john", args...)
}

// RunHashcat executes hashcat
func RunHashcat(hashFile string, opts ...string) (string, error) {
	args := append([]string{hashFile}, opts...)
	return RunExternalSafe("hashcat", args...)
}
