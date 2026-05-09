package network

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// PortResult represents a single port scan result
type PortResult struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Open         bool          `json:"open"`
	Service      string        `json:"service"`
	Banner       string        `json:"banner"`
	Protocol     string        `json:"protocol"`
	TTL          int           `json:"ttl"`
	ResponseTime time.Duration `json:"response_time"`
}

// PortBreacher is the high-speed port scanner
type PortBreacher struct {
	Timeout     time.Duration
	Workers     int
	ScanType    string // syn, connect, udp
	OSDetection bool
}

// Common ports to scan
var TopPorts = []int{
	21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445,
	993, 995, 1723, 3306, 3389, 5900, 8080, 8443, 8888, 9200,
	27017, 6379, 5432, 1521, 1433, 3306, 5000, 8000, 9000,
	10000, 27017, 27018, 27019, 28017, 50000, 49152, 49153,
}

// Service map for common ports
var ServiceMap = map[int]string{
	21:    "ftp",
	22:    "ssh",
	23:    "telnet",
	25:    "smtp",
	53:    "dns",
	80:    "http",
	110:   "pop3",
	111:   "rpcbind",
	135:   "msrpc",
	139:   "netbios-ssn",
	143:   "imap",
	443:   "https",
	445:   "microsoft-ds",
	993:   "imaps",
	995:   "pop3s",
	1723:  "pptp",
	3306:  "mysql",
	3389:  "ms-wbt-server",
	5432:  "postgresql",
	5900:  "vnc",
	6379:  "redis",
	8080:  "http-proxy",
	8443:  "https-alt",
	9200:  "elasticsearch",
	27017: "mongodb",
}

func NewPortBreacher() *PortBreacher {
	return &PortBreacher{
		Timeout:     3 * time.Second,
		Workers:     1000,
		ScanType:    "connect",
		OSDetection: true,
	}
}

// ScanHost scans a single host for open ports
func (p *PortBreacher) ScanHost(host string, ports []int) []PortResult {
	fmt.Printf("[+] PORT-BREACHER scanning: %s (%d ports, %d workers)\n", host, len(ports), p.Workers)

	var results []PortResult
	var mu sync.Mutex

	// Use semaphore for concurrency control
	semaphore := make(chan struct{}, p.Workers)
	var wg sync.WaitGroup

	for _, portNum := range ports {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(port int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			result := p.scanPort(host, port)
			if result.Open {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}(portNum)
	}

	wg.Wait()

	// Sort by port number
	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	fmt.Printf("[+] PORT-BREACHER found %d open ports on %s\n", len(results), host)
	return results
}

// ScanRange scans a CIDR range
func (p *PortBreacher) ScanRange(cidr string, ports []int) map[string][]PortResult {
	fmt.Printf("[+] PORT-BREACHER scanning range: %s\n", cidr)

	hosts := parseCIDR(cidr)
	allResults := make(map[string][]PortResult)

	for _, host := range hosts {
		results := p.ScanHost(host, ports)
		if len(results) > 0 {
			allResults[host] = results
		}
	}

	return allResults
}

func (p *PortBreacher) scanPort(host string, port int) PortResult {
	result := PortResult{
		Host:    host,
		Port:    port,
		Open:    false,
		Service: ServiceMap[port],
	}

	target := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	start := time.Now()
	conn, err := net.DialTimeout("tcp", target, p.Timeout)
	result.ResponseTime = time.Since(start)

	if err != nil {
		return result
	}

	defer conn.Close()
	result.Open = true

	// Try to grab banner
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	banner := make([]byte, 1024)
	n, _ := conn.Read(banner)
	if n > 0 {
		result.Banner = string(banner[:n])
	}

	return result
}

// OSFingerprint attempts OS fingerprinting via TCP/IP stack quirks
func (p *PortBreacher) OSFingerprint(host string) map[string]string {
	fingerprint := make(map[string]string)

	// TTL-based OS detection (simplified)
	target := net.JoinHostPort(host, "80")
	conn, err := net.DialTimeout("tcp", target, 2*time.Second)
	if err != nil {
		return fingerprint
	}
	defer conn.Close()

	fingerprint["method"] = "ttl_estimation"
	fingerprint["note"] = "Full fingerprinting requires raw socket access"

	return fingerprint
}

func parseCIDR(cidr string) []string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []string{cidr}
	}

	var hosts []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		hosts = append(hosts, ip.String())
	}

	// Remove network and broadcast addresses for IPv4
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}

	return hosts
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ScanUDP performs UDP port scanning
func (p *PortBreacher) ScanUDP(host string, ports []int) []PortResult {
	fmt.Printf("[+] PORT-BREACHER UDP scanning: %s (%d ports)\n", host, len(ports))

	var results []PortResult
	var mu sync.Mutex

	semaphore := make(chan struct{}, p.Workers)
	var wg sync.WaitGroup

	for _, portNum := range ports {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(port int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			target := net.JoinHostPort(host, fmt.Sprintf("%d", port))
			conn, err := net.DialTimeout("udp", target, p.Timeout)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send probe
			conn.Write([]byte("\n"))

			// Set read deadline
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)

			// UDP is open if we get a response or ICMP unreachable is not received
			// This is simplified - proper UDP scanning requires raw sockets
			if err == nil && n > 0 {
				mu.Lock()
				results = append(results, PortResult{
					Host:     host,
					Port:     port,
					Open:     true,
					Service:  ServiceMap[port],
					Protocol: "udp",
					Banner:   string(buf[:n]),
				})
				mu.Unlock()
			}
		}(portNum)
	}

	wg.Wait()
	return results
}

// ScanSYN performs SYN stealth scan (uses fast connect as fallback on Windows)
func (p *PortBreacher) ScanSYN(host string, ports []int) []PortResult {
	fmt.Printf("[+] PORT-BREACHER SYN scan: %s\n", host)

	var results []PortResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// On Windows, raw sockets require admin and are complex
	// Use aggressive fast-connect as stealthy alternative
	for _, port := range ports {
		wg.Add(1)
		go func(p_ int) {
			defer wg.Done()

			start := time.Now()
			addr := net.JoinHostPort(host, fmt.Sprintf("%d", p_))

			// Fast connect with very short timeout (stealthy)
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			elapsed := time.Since(start)

			mu.Lock()
			if err == nil {
				conn.Close()
				// Grab banner if possible
				banner := grabBanner(host, p_, 1*time.Second)
				results = append(results, PortResult{
					Host:         host,
					Port:         p_,
					Open:         true,
					Service:      ServiceMap[p_],
					Banner:       banner,
					Protocol:     "tcp",
					ResponseTime: elapsed,
				})
			} else {
				results = append(results, PortResult{
					Host:     host,
					Port:     p_,
					Open:     false,
					Service:  ServiceMap[p_],
					Protocol: "tcp",
				})
			}
			mu.Unlock()
		}(port)
	}

	wg.Wait()

	// Sort by port number
	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	fmt.Printf("[+] SYN scan complete: %d/%d ports open\n",
		countOpen(results), len(results))
	return results
}

func grabBanner(host string, port int, timeout time.Duration) string {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(timeout))

	// Try to read banner
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	if n > 0 {
		return string(buf[:n])
	}
	return ""
}

func countOpen(results []PortResult) int {
	count := 0
	for _, r := range results {
		if r.Open {
			count++
		}
	}
	return count
}
