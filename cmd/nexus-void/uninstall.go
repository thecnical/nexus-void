package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runUninstall() error {
	fmt.Println("[!] NEXUS-VOID UNINSTALL — This will remove ALL data.")
	fmt.Print("[!] Type 'DESTROY' to confirm: ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "DESTROY" {
		fmt.Println("[+] Uninstall cancelled.")
		return nil
	}

	installDir := "/opt/nexus-void"
	binDir := "/usr/local/bin"
	serviceFile := "/etc/systemd/system/nexus-void-backend.service"
	homeDir, _ := os.UserHomeDir()
	brainDir := filepath.Join(homeDir, ".nexus-void")

	steps := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Stop systemd service",
			fn: func() error {
				exec.Command("sudo", "systemctl", "stop", "nexus-void-backend").Run()
				exec.Command("sudo", "systemctl", "disable", "nexus-void-backend").Run()
				return nil
			},
		},
		{
			name: "Remove systemd service file",
			fn: func() error {
				exec.Command("sudo", "rm", "-f", serviceFile).Run()
				exec.Command("sudo", "systemctl", "daemon-reload").Run()
				return nil
			},
		},
		{
			name: "Remove binary symlinks",
			fn: func() error {
				exec.Command("sudo", "rm", "-f", filepath.Join(binDir, "nexus-void")).Run()
				exec.Command("sudo", "rm", "-f", filepath.Join(binDir, "nexus-server")).Run()
				return nil
			},
		},
		{
			name: "Remove installation directory",
			fn: func() error {
				exec.Command("sudo", "rm", "-rf", installDir).Run()
				return nil
			},
		},
		{
			name: "Remove brain/data directory",
			fn: func() error {
				exec.Command("rm", "-rf", brainDir).Run()
				return nil
			},
		},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s... ", step.name)
		if err := step.fn(); err != nil {
			fmt.Printf("WARN (%v)\n", err)
		} else {
			fmt.Println("OK")
		}
	}

	fmt.Println()
	fmt.Println("[+] Nexus Void has been completely removed from this system.")
	fmt.Println("[+] To reinstall: curl -fsSL https://raw.githubusercontent.com/thecnical/nexus-void/main/install.sh | bash")
	return nil
}
