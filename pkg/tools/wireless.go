package tools

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// === 1. WIFI-BREAKER ===
// Runs real WiFi audit: wifite scan or aircrack-ng suite
func runWiFiBreaker(args []string) (string, error) {
	var results []string
	iface := "wlan0"
	if len(args) > 0 && strings.HasPrefix(args[0], "wlan") {
		iface = args[0]
	}

	results = append(results, "[+] WiFi Audit Mode - Interface: "+iface)

	// 1. Show wireless interfaces
	output, _ := exec.Command("iw", "dev").CombinedOutput()
	if len(output) > 0 {
		results = append(results, "[+] Wireless devices:\n"+string(output))
	}

	// 2. Try wifite --check (lists networks, needs root)
	output, err := exec.Command("sudo", "wifite", "--check", "--iface", iface, "--kill").CombinedOutput()
	if err == nil && len(output) > 0 {
		results = append(results, "[+] Wifite scan:\n"+string(output))
	} else {
		// Fallback: iwlist scan
		output, _ = exec.Command("sudo", "iwlist", iface, "scan").CombinedOutput()
		if len(output) > 0 {
			results = append(results, "[+] Scan results:\n"+string(output))
		}
	}

	// 3. Check if aircrack-ng tools are available
	for _, tool := range []string{"aircrack-ng", "airodump-ng", "aireplay-ng", "airbase-ng"} {
		if _, err := exec.LookPath(tool); err == nil {
			results = append(results, fmt.Sprintf("[OK] %s available", tool))
		} else {
			results = append(results, fmt.Sprintf("[!] %s missing - install: nexus-void arsenal install aircrack-ng", tool))
		}
	}

	// 4. Check for wordlists
	output, _ = exec.Command("ls", "/usr/share/wordlists/").CombinedOutput()
	if len(output) > 0 {
		results = append(results, "[+] Available wordlists:\n"+string(output))
	}

	results = append(results, "[*] Note: Cracking requires monitor mode. Run: sudo airmon-ng start "+iface)
	return fmt.Sprintf("[WIFI-BREAKER]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 2. BLUETOOTH-SNIFFER ===
func runBluetoothSniffer(args []string) (string, error) {
	var results []string
	results = append(results, "[+] Bluetooth Analysis Mode")

	if runtime.GOOS == "windows" {
		output, _ := exec.Command("powershell", "-Command", "Get-PnpDevice -Class Bluetooth").CombinedOutput()
		results = append(results, "[+] Bluetooth devices:\n"+string(output))
	} else {
		// hci devices
		output, _ := exec.Command("hciconfig", "-a").CombinedOutput()
		if len(output) > 0 {
			results = append(results, "[+] HCI devices:\n"+string(output))
		}

		// Real scan
		output, _ = exec.Command("hcitool", "scan", "--length", "10").CombinedOutput()
		if len(output) > 0 {
			results = append(results, "[+] BT scan results:\n"+string(output))
		} else {
			results = append(results, "[!] No devices found or bluetooth disabled")
		}

		// BLE scan with bluetoothctl if available
		output, _ = exec.Command("bluetoothctl", "scan", "on").CombinedOutput()
		if len(output) > 0 {
			results = append(results, "[+] BLE scan:\n"+string(output))
		}
	}

	return fmt.Sprintf("[BLUETOOTH-SNIFFER]\n%s", strings.Join(results, "\n---\n")), nil
}

// === 3. RF-SNIFFER ===
func runRFSniffer(args []string) (string, error) {
	var results []string
	results = append(results, "[+] RF Signal Analysis Mode")

	// Check RTL-SDR
	output, _ := exec.Command("rtl_test", "-t").CombinedOutput()
	if len(output) > 0 && !strings.Contains(string(output), "not found") {
		results = append(results, "[+] RTL-SDR test:\n"+string(output))
	} else {
		results = append(results, "[!] RTL-SDR not found. Install: apt install rtl-sdr")
	}

	// Check HackRF
	output, _ = exec.Command("hackrf_info").CombinedOutput()
	if len(output) > 0 && !strings.Contains(string(output), "not found") {
		results = append(results, "[+] HackRF Info:\n"+string(output))
	}

	// Check for SDR tools
	for _, tool := range []string{"gqrx", "urh", "rtl_433", "multimon-ng"} {
		if _, err := exec.LookPath(tool); err == nil {
			results = append(results, fmt.Sprintf("[OK] %s available", tool))
		}
	}

	results = append(results, "[*] Note: Requires SDR hardware (RTL-SDR/HackRF/BladeRF)")
	return fmt.Sprintf("[RF-SNIFFER]\n%s", strings.Join(results, "\n")), nil
}

// === 4. ANDROID-ASSASSIN ===
func runAndroidAssassin(args []string) (string, error) {
	var results []string
	results = append(results, "[+] Android Analysis Mode")

	// Real ADB check
	output, err := exec.Command("adb", "devices").CombinedOutput()
	if err == nil && len(output) > 0 {
		results = append(results, "[+] ADB Status:\n"+string(output))
		// If a device is connected, get basic info
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "device") && !strings.Contains(line, "List") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					device := parts[0]
					info, _ := exec.Command("adb", "-s", device, "shell", "getprop", "ro.product.model").CombinedOutput()
					results = append(results, fmt.Sprintf("    Device: %s | Model: %s", device, strings.TrimSpace(string(info))))
					info, _ = exec.Command("adb", "-s", device, "shell", "getprop", "ro.build.version.release").CombinedOutput()
					results = append(results, fmt.Sprintf("    Android Version: %s", strings.TrimSpace(string(info))))
				}
			}
		}
	} else {
		results = append(results, "[!] No ADB device connected or adb not installed")
		results = append(results, "[*] Connect device: adb connect <ip:5555>")
	}

	return fmt.Sprintf("[ANDROID-ASSASSIN]\n%s", strings.Join(results, "\n")), nil
}

// === 5. IOS-PHANTOM ===
func runIOSPhantom(args []string) (string, error) {
	var results []string
	results = append(results, "[+] iOS Analysis Mode")

	// Check for frida
	output, _ := exec.Command("frida", "--version").CombinedOutput()
	if len(output) > 0 && !strings.Contains(string(output), "not found") {
		results = append(results, "[+] Frida version: "+strings.TrimSpace(string(output)))
	} else {
		results = append(results, "[!] Frida not installed. pip install frida-tools")
	}

	// Check for libimobiledevice
	for _, tool := range []string{"ideviceinfo", "ideviceinstaller", "idevicedebug"} {
		if _, err := exec.LookPath(tool); err == nil {
			results = append(results, fmt.Sprintf("[OK] %s available", tool))
		} else {
			results = append(results, fmt.Sprintf("[!] %s missing - apt install libimobiledevice-utils", tool))
		}
	}

	results = append(results, "[*] Connect iOS device via USB for analysis")
	return fmt.Sprintf("[IOS-PHANTOM]\n%s", strings.Join(results, "\n")), nil
}

// === 6. APK-REVERSE ===
func runAPKReverse(args []string) (string, error) {
	apk := "app.apk"
	if len(args) > 0 {
		apk = args[0]
	}

	// Try jadx first, then apktool
	output, err := RunExternalSafe("jadx", "-d", apk+"_jadx", apk)
	if err == nil {
		return fmt.Sprintf("[APK-REVERSE] %s\nDecompiled with jadx to: %s_jadx/\n%s", apk, apk, output), nil
	}

	output, err = RunExternalSafe("apktool", "d", apk, "-o", apk+"_apktool")
	if err == nil {
		return fmt.Sprintf("[APK-REVERSE] %s\nDecompiled with apktool to: %s_apktool/\n%s", apk, apk, output), nil
	}

	// Try unzip for quick manifest look
	output, _ = RunExternal("unzip", "-p", apk, "AndroidManifest.xml")
	if len(output) > 0 {
		return fmt.Sprintf("[APK-REVERSE] %s\n[!] jadx/apktool not found. Raw AndroidManifest.xml (first 2KB):\n%s", apk, output[:min(len(output), 2048)]), nil
	}

	return fmt.Sprintf("[APK-REVERSE] %s\n[!] jadx/apktool not available. Install: nexus-void arsenal install jadx", apk), nil
}

// === 7. MOBILE-API-PWN ===
func runMobileAPIPwn(args []string) (string, error) {
	api := "http://api.example.com"
	if len(args) > 0 {
		api = args[0]
	}

	var results []string
	results = append(results, fmt.Sprintf("[+] Target API: %s", api))

	// Real probe with httpx
	output, _ := RunExternalSafe("httpx", "-u", api, "-tech-detect", "-title", "-status-code", "-silent")
	if len(output) > 0 {
		results = append(results, "[+] httpx probe:\n"+string(output))
	}

	// Check SSL/TLS with nuclei
	output, _ = RunExternalSafe("nuclei", "-u", api, "-t", "ssl/", "-silent")
	if len(output) > 0 {
		results = append(results, "[+] SSL/TLS checks:\n"+string(output))
	}

	// Check for common mobile API endpoints
	endpoints := []string{"/api/v1/", "/graphql", "/swagger.json", "/openapi.json", "/api/docs"}
	for _, ep := range endpoints {
		output, _ = RunExternal("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", api+ep)
		code := strings.TrimSpace(string(output))
		if code == "200" || code == "401" || code == "403" {
			results = append(results, fmt.Sprintf("[+] Found: %s (HTTP %s)", api+ep, code))
		}
	}

	results = append(results, "[*] For deeper testing, proxy through Burp/ZAP")
	return fmt.Sprintf("[MOBILE-API-PWN]\n%s", strings.Join(results, "\n")), nil
}

// === 8. BASEBAND-HUNTER ===
func runBasebandHunter(args []string) (string, error) {
	var results []string
	results = append(results, "[+] Cellular Baseband Analysis Mode")

	// Check for srsRAN
	for _, tool := range []string{"srsenb", "srsepc", "srsue"} {
		if _, err := exec.LookPath(tool); err == nil {
			results = append(results, fmt.Sprintf("[OK] %s available", tool))
		}
	}

	// Check for Osmocom
	for _, tool := range []string{"osmo-trx", "osmo-bsc", "osmo-nitb"} {
		if _, err := exec.LookPath(tool); err == nil {
			results = append(results, fmt.Sprintf("[OK] %s available", tool))
		}
	}

	// Check SDR hardware
	output, _ := exec.Command("hackrf_info").CombinedOutput()
	if len(output) > 0 && !strings.Contains(string(output), "not found") {
		results = append(results, "[+] HackRF detected:\n"+string(output))
	}

	output, _ = exec.Command("rtl_test", "-t").CombinedOutput()
	if len(output) > 0 && !strings.Contains(string(output), "not found") {
		results = append(results, "[+] RTL-SDR detected:\n"+string(output))
	}

	results = append(results, "[*] Requires: SDR hardware + srsRAN/Osmocom stack")
	results = append(results, "[*] Warning: Cellular testing requires licenses in most countries")
	return fmt.Sprintf("[BASEBAND-HUNTER]\n%s", strings.Join(results, "\n")), nil
}
