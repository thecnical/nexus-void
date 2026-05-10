// ALPHA-SCANNER Agent
// Hardware detection, monitor mode, network scan, OUI/firmware fingerprint, packet profiling

package etherbreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// AlphaScanner handles reconnaissance and hardware
type AlphaScanner struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	status string
}

// NewAlphaScanner creates the scanner agent
func NewAlphaScanner() *AlphaScanner {
	return &AlphaScanner{
		stopCh: make(chan struct{}),
		status: "idle",
	}
}

func (a *AlphaScanner) Name() string   { return "ALPHA" }
func (a *AlphaScanner) Status() string { return a.status }

// Start begins listening for events
func (a *AlphaScanner) Start(bus *EventBus, state *SharedState) {
	a.bus = bus
	a.state = state
	ch := bus.Subscribe(a.Name())

	for {
		select {
		case msg := <-ch:
			a.handleMessage(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *AlphaScanner) Stop() {
	close(a.stopCh)
}

func (a *AlphaScanner) handleMessage(msg AgentMessage) {
	switch msg.Type {
	case "SCAN_START":
		a.status = "scanning"
		a.runFullScan()
		a.status = "idle"
	case "FIRMWARE_CHECK":
		if msg.Target != nil {
			a.fingerprintFirmware(msg.Target)
		}
	case "PACKET_PROFILE":
		if msg.Target != nil {
			a.packetProfile(msg.Target)
		}
	}
}

// runFullScan does a real network discovery
func (a *AlphaScanner) runFullScan() {
	a.broadcast("Starting full wireless scan...")
	iface := a.getMonitorIface()

	// Write CSV output for parsing
	csvFile := filepath.Join("/tmp/nexus-void", "scan.csv")

	// Run airodump-ng for 15 seconds
	cmd := exec.Command("sudo", "timeout", "15", "airodump-ng",
		"--write-interval", "1",
		"--output-format", "csv",
		"--write", csvFile,
		iface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start and let it run
	if err := cmd.Start(); err != nil {
		a.broadcast(fmt.Sprintf("[!] airodump-ng failed: %v", err))
		// Fallback to iwlist
		a.scanWithIwlist()
		return
	}

	// Wait for completion
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
		// airodump finished
	case <-time.After(16 * time.Second):
		cmd.Process.Kill()
	}

	// Parse CSV results
	targets := a.parseAirodumpCSV(csvFile + "-01.csv")

	a.state.mu.Lock()
	a.state.Targets = targets
	a.state.mu.Unlock()

	a.broadcast(fmt.Sprintf("Scan complete. Found %d networks.", len(targets)))

	// Send results to OMEGA for AI decision
	a.bus.Broadcast(AgentMessage{
		From:   a.Name(),
		To:     "OMEGA",
		Type:   "SCAN_RESULT",
		Target: nil,
		Data:   fmt.Sprintf("found %d networks", len(targets)),
		Payload: map[string]interface{}{
			"targets": targets,
		},
	})
}

// scanWithIwlist is the fallback scanner
func (a *AlphaScanner) scanWithIwlist() {
	iface := a.state.Adapter
	out, err := exec.Command("sudo", "iwlist", iface, "scan").CombinedOutput()
	if err != nil {
		a.broadcast(fmt.Sprintf("[!] iwlist scan failed: %v", err))
		return
	}

	var targets []*NetworkTarget
	lines := strings.Split(string(out), "\n")
	var current *NetworkTarget

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Cell ") {
			if current != nil {
				targets = append(targets, current)
			}
			current = &NetworkTarget{}
		}
		if strings.Contains(line, "ESSID:") {
			if current != nil {
				current.SSID = strings.Trim(strings.Split(line, "ESSID:")[1], "\"")
			}
		}
		if strings.Contains(line, "Address:") {
			if current != nil {
				parts := strings.Fields(line)
				if len(parts) >= 5 {
					current.BSSID = parts[4]
				}
			}
		}
		if strings.Contains(line, "Channel:") {
			if current != nil {
				chStr := strings.TrimPrefix(line, "Channel:")
				chStr = strings.TrimSpace(chStr)
				if ch, err := strconv.Atoi(chStr); err == nil {
					current.Channel = ch
				}
			}
		}
		if strings.Contains(line, "Encryption key:") {
			if current != nil && strings.Contains(line, "off") {
				current.Security = "OPEN"
			}
		}
		if strings.Contains(line, "WPA3") {
			if current != nil {
				current.Security = "WPA3"
			}
		} else if strings.Contains(line, "WPA2") {
			if current != nil && current.Security == "" {
				current.Security = "WPA2"
			}
		} else if strings.Contains(line, "WPA") {
			if current != nil && current.Security == "" {
				current.Security = "WPA"
			}
		} else if strings.Contains(line, "WEP") {
			if current != nil && current.Security == "" {
				current.Security = "WEP"
			}
		}
	}
	if current != nil {
		targets = append(targets, current)
	}

	a.state.mu.Lock()
	a.state.Targets = targets
	a.state.mu.Unlock()

	a.broadcast(fmt.Sprintf("iwlist scan found %d networks", len(targets)))
	a.bus.Broadcast(AgentMessage{
		From: a.Name(), To: "OMEGA", Type: "SCAN_RESULT",
		Payload: map[string]interface{}{"targets": targets},
	})
}

// parseAirodumpCSV reads the airodump-ng CSV output
func (a *AlphaScanner) parseAirodumpCSV(filename string) []*NetworkTarget {
	var targets []*NetworkTarget
	data, err := os.ReadFile(filename)
	if err != nil {
		return targets
	}

	lines := strings.Split(string(data), "\n")
	inAccessPoints := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "BSSID") {
			inAccessPoints = true
			continue
		}
		if strings.HasPrefix(line, "Station MAC") {
			break // End of AP section
		}
		if inAccessPoints {
			parts := strings.Split(line, ",")
			if len(parts) >= 14 {
				t := &NetworkTarget{
					BSSID:     strings.TrimSpace(parts[0]),
					FirstSeen: time.Now(),
					LastSeen:  time.Now(),
				}
				if ch, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil {
					t.Channel = ch
				}
				t.Security = strings.TrimSpace(parts[5])
				t.SSID = strings.Trim(strings.TrimSpace(parts[13]), "\r")
				t.Brand = NewOUIDatabase().Lookup(t.BSSID)
				targets = append(targets, t)
			}
		}
	}

	return targets
}

// fingerprintFirmware does OUI lookup and CVE queries
func (a *AlphaScanner) fingerprintFirmware(target *NetworkTarget) {
	brand := NewOUIDatabase().Lookup(target.BSSID)
	a.broadcast(fmt.Sprintf("[FIRMWARE] %s → Brand: %s", target.BSSID, brand))
	target.Brand = brand

	// Check for known default credentials by brand
	creds := a.getDefaultCreds(brand)
	if creds != "" {
		a.broadcast(fmt.Sprintf("[FIRMWARE] Default creds for %s: %s", brand, creds))
	}

	// Send to OMEGA
	a.bus.Broadcast(AgentMessage{
		From:   a.Name(),
		To:     "OMEGA",
		Type:   "FIRMWARE_RESULT",
		Target: target,
		Data:   fmt.Sprintf("brand=%s default_creds=%s", brand, creds),
	})
}

// getDefaultCreds returns known default credentials for common brands
func (a *AlphaScanner) getDefaultCreds(brand string) string {
	creds := map[string]string{
		"TP-Link":  "admin/admin",
		"NetGear":  "admin/password",
		"D-Link":   "admin/\"\"",
		"Linksys":  "admin/admin",
		"Asus":     "admin/admin",
		"Huawei":   "admin/admin",
		"Cisco":    "admin/cisco",
		"Belkin":   "admin/\"\"",
		"Xiaomi":   "admin/admin",
		"Ubiquiti": "ubnt/ubnt",
	}
	if c, ok := creds[brand]; ok {
		return c
	}
	return ""
}

// packetProfile analyzes encrypted traffic metadata
func (a *AlphaScanner) packetProfile(target *NetworkTarget) {
	a.broadcast(fmt.Sprintf("[PACKET-PROFILE] Analyzing traffic for %s...", target.SSID))

	// Capture for 30 seconds on target channel
	iface := a.getMonitorIface()
	capFile := filepath.Join("/tmp/nexus-void", fmt.Sprintf("profile_%s.pcap", target.BSSID))

	cmd := exec.Command("sudo", "timeout", "30", "tcpdump",
		"-i", iface,
		"-w", capFile,
		"ether", "host", target.BSSID)
	cmd.Run()

	// Analyze with tcpdump -r
	out, err := exec.Command("tcpdump", "-r", capFile, "-nn", "-q").CombinedOutput()
	if err != nil {
		return
	}

	// Infer devices from packet patterns
	var devices []string
	if strings.Contains(string(out), "1900") {
		devices = append(devices, "UPnP/IoT Device")
	}
	if strings.Contains(string(out), "5353") {
		devices = append(devices, "Apple Device (mDNS)")
	}
	if strings.Contains(string(out), "5355") {
		devices = append(devices, "Windows Device (LLMNR)")
	}

	a.broadcast(fmt.Sprintf("[PACKET-PROFILE] Detected: %v", devices))
	a.bus.Broadcast(AgentMessage{
		From:   a.Name(),
		To:     "OMEGA",
		Type:   "PROFILE_RESULT",
		Target: target,
		Data:   fmt.Sprintf("devices=%v", devices),
	})
}

// getMonitorIface returns the monitor interface name
func (a *AlphaScanner) getMonitorIface() string {
	a.state.mu.RLock()
	defer a.state.mu.RUnlock()
	return a.state.MonitorIface
}

func (a *AlphaScanner) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{
		From: a.Name(), To: "ALL", Type: "LOG", Data: msg,
	})
}
