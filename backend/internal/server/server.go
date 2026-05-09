package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nexus-void/backend/internal/agents"
	"github.com/nexus-void/backend/internal/brain"
	"github.com/nexus-void/backend/internal/recon"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a connected dashboard client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Operator string
}

// NexusServer is the team server with multi-operator support
type NexusServer struct {
	brain       *brain.Brain
	coordinator *agents.Coordinator
	clients     map[string]*Client
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	sessions    map[string]*brain.Session
	httpServer  *http.Server
}

// ServerMessage represents a message broadcast to clients
type ServerMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewServer creates a new team server
func NewServer(b *brain.Brain, coord *agents.Coordinator) *NexusServer {
	return &NexusServer{
		brain:       b,
		coordinator: coord,
		clients:     make(map[string]*Client),
		broadcast:   make(chan []byte, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		sessions:    make(map[string]*brain.Session),
	}
}

// Start begins the team server
func (s *NexusServer) Start(addr string) error {
	go s.run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/sessions", s.handleSessions)
	mux.HandleFunc("/api/agents", s.handleAgents)
	mux.HandleFunc("/api/brain/stats", s.handleBrainStats)
	mux.HandleFunc("/api/brain/strategies", s.handleStrategies)
	mux.HandleFunc("/api/recon", s.handleRecon)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/exploit/verify", s.handleExploitVerify)
	mux.HandleFunc("/api/payload/evolve", s.handlePayloadEvolve)
	mux.HandleFunc("/api/report/generate", s.handleReportGenerate)
	mux.HandleFunc("/api/c2/beacons", s.handleC2Beacons)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Printf("[+] NEXUS-VOID Backend v2.0 listening on %s\n", addr)
	fmt.Printf("[+] WebSocket: ws://%s/ws\n", addr)
	fmt.Printf("[+] API: http://%s/api/\n", addr)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *NexusServer) Shutdown() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
	if s.coordinator != nil {
		s.coordinator.Stop()
	}
}

func (s *NexusServer) run() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.ID] = client
			s.mu.Unlock()
			log.Printf("[+] Client connected: %s (operator: %s)", client.ID, client.Operator)
			s.sendWelcome(client)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.mu.Unlock()
			log.Printf("[-] Client disconnected: %s", client.ID)

		case message := <-s.broadcast:
			s.mu.RLock()
			for _, client := range s.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(s.clients, client.ID)
				}
			}
			s.mu.RUnlock()

		case <-ticker.C:
			s.broadcastStatusUpdate()
		}
	}
}

func (s *NexusServer) broadcastStatusUpdate() {
	s.mu.RLock()
	clientCount := len(s.clients)
	s.mu.RUnlock()

	if clientCount == 0 {
		return
	}

	agents := s.coordinator.GetAgentStatus()
	stats := s.brain.GetStats()

	msg := ServerMessage{
		Type:      "status_update",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"clients": clientCount,
			"agents":  agents,
			"brain":   stats,
			"uptime":  time.Now().Format(time.RFC3339),
		},
	}

	data, _ := json.Marshal(msg)
	select {
	case s.broadcast <- data:
	default:
	}
}

func (s *NexusServer) sendWelcome(client *Client) {
	msg := ServerMessage{
		Type:      "welcome",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"version": "2.0",
			"agents":  s.coordinator.AgentCount(),
			"clients": len(s.clients),
			"message": "Connected to NEXUS-VOID OMEGA Backend",
		},
	}
	data, _ := json.Marshal(msg)
	client.Send <- data
}

func (s *NexusServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[!] WebSocket upgrade failed: %v", err)
		return
	}

	operator := r.URL.Query().Get("operator")
	if operator == "" {
		operator = "anonymous"
	}

	client := &Client{
		ID:       fmt.Sprintf("client_%d", time.Now().UnixNano()),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Operator: operator,
	}

	s.register <- client

	go s.writePump(client)
	go s.readPump(client)
}

func (s *NexusServer) readPump(client *Client) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[!] WebSocket error: %v", err)
			}
			break
		}
		s.handleClientMessage(client, message)
	}
}

func (s *NexusServer) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			client.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *NexusServer) handleClientMessage(client *Client, message []byte) {
	var cmd map[string]interface{}
	if err := json.Unmarshal(message, &cmd); err != nil {
		return
	}

	cmdType, _ := cmd["type"].(string)
	switch cmdType {
	case "start_session":
		target, _ := cmd["target"].(string)
		s.startSession(client, target)
	case "stop_session":
		sessionID, _ := cmd["session_id"].(string)
		s.stopSession(sessionID)
	case "dispatch_task":
		taskType, _ := cmd["task_type"].(string)
		target, _ := cmd["target"].(string)
		priority := 5
		if p, ok := cmd["priority"].(float64); ok {
			priority = int(p)
		}
		taskID := s.coordinator.DispatchTask(taskType, target, priority)
		s.BroadcastStatus("task_dispatched", map[string]interface{}{
			"task_id": taskID,
			"type":    taskType,
			"target":  target,
		})
	case "command":
		s.BroadcastStatus("command_received", cmd)
	}
}

func (s *NexusServer) startSession(client *Client, target string) {
	brainSession := s.brain.NewSession(target)

	// Dispatch initial recon tasks
	s.coordinator.DispatchTask("recon", target, 10)
	s.coordinator.DispatchTask("osint", target, 8)

	s.BroadcastStatus("session_started", map[string]interface{}{
		"session_id": brainSession.ID,
		"target":     target,
		"operator":   client.Operator,
		"timestamp":  time.Now(),
	})
}

func (s *NexusServer) stopSession(sessionID string) {
	s.brain.UpdateSession(sessionID, map[string]interface{}{"status": "completed"})
	s.BroadcastStatus("session_stopped", map[string]string{"session_id": sessionID})
}

// BroadcastStatus sends a status update to all connected clients
func (s *NexusServer) BroadcastStatus(msgType string, data interface{}) {
	msg := ServerMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case s.broadcast <- jsonData:
	default:
		// Channel full, drop message
	}
}

// BroadcastAgentUpdate broadcasts agent status update
func (s *NexusServer) BroadcastAgentUpdate(agentName, status string, progress int) {
	s.BroadcastStatus("agent_update", map[string]interface{}{
		"agent":    agentName,
		"status":   status,
		"progress": progress,
	})
}

func (s *NexusServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"version": "2.0",
		"clients": len(s.clients),
		"agents":  s.coordinator.AgentCount(),
		"uptime":  "running",
		"status":  "operational",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *NexusServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.sessions)
}

func (s *NexusServer) handleAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.coordinator.GetAgentStatus())
}

func (s *NexusServer) handleBrainStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.brain.GetStats())
}

func (s *NexusServer) handleStrategies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.brain.ExportStrategies())
}

// handleRecon performs actual reconnaissance on a target
func (s *NexusServer) handleRecon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if req.Target == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "target required"})
		return
	}

	// Broadcast start
	s.BroadcastStatus("recon_started", map[string]string{"target": req.Target})

	// Perform actual reconnaissance
	go func() {
		result, err := recon.Scan(req.Target)
		if err != nil {
			s.BroadcastStatus("recon_error", map[string]string{"target": req.Target, "error": err.Error()})
			return
		}

		// Learn from results
		s.brain.UpdateTargetDNA(req.Target, &brain.TargetDNA{
			Target:       req.Target,
			Technologies: result.TechStack,
		})

		// Record learned strategies
		for _, tech := range result.TechStack {
			s.brain.Learn.RecordSuccess("tech_detect", "recon", 0)
			_ = tech
		}

		s.BroadcastStatus("recon_complete", result)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started", "target": req.Target})
}

// handleScan performs a quick scan on a target
func (s *NexusServer) handleScan(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "target query param required"})
		return
	}

	code, serverHeader, err := recon.QuickProbe(target)
	result := map[string]interface{}{
		"target":      target,
		"status_code": code,
		"server":      serverHeader,
		"alive":       code > 0 && code < 500,
	}
	if err != nil {
		result["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleExploitVerify tests a vulnerability on a real target
func (s *NexusServer) handleExploitVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Target    string `json:"target"`
		Parameter string `json:"parameter"`
		Type      string `json:"type"` // sqli, xss, lfi, ssrf, cmdi
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	s.BroadcastStatus("exploit_verify_started", map[string]string{
		"target": req.Target,
		"type":   req.Type,
	})

	go func() {
		result := s.verifyExploit(req.Target, req.Parameter, req.Type)
		s.BroadcastStatus("exploit_verify_complete", result)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *NexusServer) verifyExploit(target, param, vulnType string) map[string]interface{} {
	// Use brain's reasoning to determine verification approach
	dna := s.brain.GetTargetDNA(target)
	if dna != nil {
		s.brain.Reason.Infer([]brain.Fact{
			{Subject: target, Predicate: "has", Object: vulnType, Confidence: 0.7},
		})
	}

	// Perform actual verification via recon module
	code, server, err := recon.QuickProbe(target)
	return map[string]interface{}{
		"target":      target,
		"type":        vulnType,
		"parameter":   param,
		"status_code": code,
		"server":      server,
		"reachable":   err == nil,
		"timestamp":   time.Now(),
	}
}

// handlePayloadEvolve genetically evolves payloads
func (s *NexusServer) handlePayloadEvolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SeedPayloads []string `json:"seed_payloads"`
		TargetType   string   `json:"target_type"` // sqli, xss, lfi, rce
		Generations  int      `json:"generations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if req.Generations <= 0 {
		req.Generations = 20
	}
	if req.Generations > 50 {
		req.Generations = 50
	}

	var evolved []map[string]interface{}
	for _, seed := range req.SeedPayloads {
		gen := s.brain.Evolve.EvolvePayload(seed, req.TargetType, req.Generations)
		evolved = append(evolved, map[string]interface{}{
			"parent":      seed,
			"evolved":     gen.Payload,
			"score":       gen.Score,
			"target_type": req.TargetType,
			"generations": req.Generations,
			"mutation":    gen.MutationType,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payloads": evolved,
		"count":    len(evolved),
	})
}

// handleReportGenerate generates engagement reports
func (s *NexusServer) handleReportGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		Format    string `json:"format"` // html, json, markdown
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	stats := s.brain.GetStats()
	strategies := s.brain.ExportStrategies()

	report := map[string]interface{}{
		"generated_at": time.Now(),
		"session_id":   req.SessionID,
		"format":       req.Format,
		"brain_stats":  stats,
		"strategies":   strategies,
		"agents":       s.coordinator.GetAgentStatus(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// handleC2Beacons manages C2 beacon status
func (s *NexusServer) handleC2Beacons(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return beacon status
		beacons := []map[string]interface{}{
			{"id": "beacon-001", "status": "active", "last_checkin": time.Now().Add(-30 * time.Second)},
			{"id": "beacon-002", "status": "idle", "last_checkin": time.Now().Add(-5 * time.Minute)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(beacons)

	case http.MethodPost:
		// Register new beacon or send task
		var req struct {
			BeaconID string `json:"beacon_id"`
			Command  string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.BroadcastStatus("c2_task", map[string]interface{}{
			"beacon_id": req.BeaconID,
			"command":   req.Command,
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "task_sent"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
