package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nexus-void/backend/internal/agents"
	"github.com/nexus-void/backend/internal/brain"
	"github.com/nexus-void/backend/internal/server"
)

func main() {
	var addr string
	var brainDir string
	var verbose bool

	flag.StringVar(&addr, "addr", ":8080", "Server listen address")
	flag.StringVar(&brainDir, "brain", "./.nexus-void/brain", "Brain data directory")
	flag.BoolVar(&verbose, "v", false, "Verbose logging")
	flag.Parse()

	fmt.Println(`
                  ▄▄▄▄▄▄▄▄
              ▄▄▀▀▓▓▓▓▓▓▓▓▀▀▄▄
            ▄▀▓▓▓▓▓▓▓▓▓▓▓▓▓▓▀▄
           █   ▄▄▓▓▓▓▄▄   █
          █   █◉█  █◉█   █
          █     ▀▀██▀▀     █
           █  ▄▄▀▀▀▀▄▄  █
            ▀▄   ▀▀██▀▀   ▄▀
              ▀▀▄▄▓▓▓▓▄▄▀▀
                  ▀▀▀▀

    ╔════════════════════════════════════════════════════╗
    ║              N E X U S   V O I D                   ║
    ║      AUTONOMOUS SWARM INTELLIGENCE WEAPON          ║
    ╠════════════════════════════════════════════════════╣
    ║         BACKEND SERVER v2.0 — AI Brain API         ║
    ╚════════════════════════════════════════════════════╝

         Created by Chandan Pandey — Architect of the Swarm
  `)

	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Initialize the brain
	b, err := brain.Initialize(brainDir)
	if err != nil {
		log.Fatalf("[!] Brain initialization failed: %v", err)
	}
	defer b.Close()
	log.Printf("[+] Brain initialized at %s", brainDir)

	// Initialize multi-agent coordinator
	coordinator := agents.NewCoordinator(b)
	log.Printf("[+] Agent coordinator initialized with %d agents", coordinator.AgentCount())

	// Start the team server
	srv := server.NewServer(b, coordinator)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[+] Shutting down server...")
		srv.Shutdown()
	}()

	log.Printf("[+] Server listening on %s", addr)
	if err := srv.Start(addr); err != nil {
		log.Fatalf("[!] Server failed: %v", err)
	}
}
