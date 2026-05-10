// ETHERBREACH - Core Types & Shared Data Structures
// All 15 features share these types across 5 agents

package etherbreach

import (
	"sync"
	"time"
)

// NetworkTarget represents a discovered wireless AP
type NetworkTarget struct {
	SSID      string
	BSSID     string
	Channel   int
	Security  string   // OPEN/WEP/WPA/WPA2/WPA3
	Clients   []ClientInfo
	Brand     string   // From OUI lookup
	Model     string
	Firmware  string
	Signal    int      // dBm
	FirstSeen time.Time
	LastSeen  time.Time
}

// ClientInfo represents a connected station
type ClientInfo struct {
	MAC      string
	Signal   int
	Packets  int
	Probing  []string // Probe-requested SSIDs (for Karma)
}

// AttackPlan is AI-generated optimal attack sequence
type AttackPlan struct {
	Target        *NetworkTarget
	Steps         []AttackStep
	Reasoning     string
	EstimatedTime time.Duration
	AgentVotes    map[string]string // agent -> vote (WPS/PMKID/HANDSHAKE/EVIL_TWIN/KARMA/WPA3_DOWNGRADE)
}

// AttackStep is a single executable action
type AttackStep struct {
	Name      string
	Tool      string
	Command   []string
	NeedsRoot bool
	Timeout   time.Duration
	Agent     string // ALPHA/BETA/GAMMA/DELTA/OMEGA
}

// AttackResults holds all findings
type AttackResults struct {
	NetworksScanned int
	Handshakes      int
	PMKIDs          int
	Cracked         int
	EvilTwins       int
	KarmaTraps      int
	Passwords       []Credential
	Findings        []string
	StartTime       time.Time
	mu              sync.Mutex
}

// Credential stores discovered password/token
type Credential struct {
	Target   string
	Username string
	Password string
	Type     string // wpa/wps/web/admin
	Source   string // reaver/pixiewps/evil_twin/etc
	Time     time.Time
}

// AgentMessage is the event bus message between agents
type AgentMessage struct {
	From     string            // ALPHA/BETA/GAMMA/DELTA/OMEGA
	To       string            // ALL or specific agent
	Type     string            // SCAN_RESULT/ATTACK_STEP/DECISION/SUCCESS/FAILURE/TELEMETRY
	Target   *NetworkTarget
	Data     string
	Payload  map[string]interface{}
	Priority int
	Time     time.Time
}

// EventBus for inter-agent communication
type EventBus struct {
	subscribers map[string]chan AgentMessage
	mu          sync.RWMutex
}

// NewEventBus creates the agent communication bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]chan AgentMessage),
	}
}

// Subscribe registers an agent to receive messages
func (eb *EventBus) Subscribe(agentName string) chan AgentMessage {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan AgentMessage, 100)
	eb.subscribers[agentName] = ch
	return ch
}

// Broadcast sends a message to all or specific agent
func (eb *EventBus) Broadcast(msg AgentMessage) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	msg.Time = time.Now()
	if msg.To == "ALL" || msg.To == "" {
		for _, ch := range eb.subscribers {
			select {
			case ch <- msg:
			default:
			}
		}
	} else if ch, ok := eb.subscribers[msg.To]; ok {
		select {
		case ch <- msg:
		default:
		}
	}
}

// SharedState holds the current session state visible to all agents
type SharedState struct {
	Adapter        string
	MonitorIface   string
	Targets        []*NetworkTarget
	CurrentTarget  *NetworkTarget
	GhostMode      bool
	LastMAC        string
	HandshakesDir  string
	WordlistsDir   string
	CaptiveDir     string
	Results        *AttackResults
	Connected      bool
	InternalHosts  []string
	mu             sync.RWMutex
}

// NewSharedState creates shared state with dirs
func NewSharedState() *SharedState {
	home := "/tmp/nexus-void"
	return &SharedState{
		Adapter:       "wlan0",
		MonitorIface:  "wlan0mon",
		HandshakesDir: home + "/handshakes",
		WordlistsDir:  home + "/wordlists",
		CaptiveDir:    home + "/captive",
		Results:       &AttackResults{StartTime: time.Now()},
	}
}

// OUIDatabase maps first 3 octets to brand
type OUIDatabase struct {
	entries map[string]string
	mu      sync.RWMutex
}

// NewOUIDatabase creates a basic OUI lookup table
func NewOUIDatabase() *OUIDatabase {
	return &OUIDatabase{
		entries: map[string]string{
			"00:11:22": "Cisco",
			"00:14:bf": "NetGear",
			"00:1a:2b": "TP-Link",
			"00:1c:c4": "D-Link",
			"00:22:6b": "Huawei",
			"00:24:d4": "Asus",
			"00:25:00": "Apple",
			"00:26:bb": "Linksys",
			"00:50:56": "VMware",
			"00:e0:4c": "Realtek",
			"00:18:e7": "Intel",
			"08:00:27": "VirtualBox",
			"20:28:18": "Arris",
			"20:aa:4b": "Cisco",
			"24:a4:3c": "Ubiquiti",
			"2c:30:33": "Belkin",
			"30:23:03": "Xiaomi",
			"30:b5:c2": "TP-Link",
			"3c:5a:b4": "Google",
			"40:16:7e": "Asus",
			"44:d9:e7": "NetGear",
			"4c:17:44": "Amazon",
			"4c:ed:fb": "Intel",
			"50:c7:bf": "TP-Link",
			"54:af:97": "MikroTik",
			"58:8d:09": "Cisco",
			"5c:e2:8c": "Huawei",
			"60:38:e0": "Belkin",
			"60:a4:4c": "Asus",
			"68:ff:7b": "Apple",
			"70:4d:7b": "NetGear",
			"70:8b:cd": "Linksys",
			"74:44:01": "Huawei",
			"78:8a:20": "Ubiquiti",
			"80:2a:a8": "Ubiquiti",
			"84:16:f9": "Google",
			"88:28:b1": "Samsung",
			"90:9a:4a": "Xiaomi",
			"94:44:52": "Realtek",
			"9c:99:a0": "Xiaomi",
			"a0:04:60": "Apple",
			"a4:2b:8c": "NetGear",
			"a4:91:b1": "Google",
			"ac:22:05": "Samsung",
			"ac:9e:17": "Apple",
			"b0:95:75": "D-Link",
			"b0:be:76": "NetGear",
			"b0:c5:54": "D-Link",
			"b4:74:9f": "Asus",
			"b8:27:eb": "Raspberry Pi",
			"bc:30:7e": "Sony",
			"c0:49:ef": "Huawei",
			"c0:a0:bb": "Cisco",
			"c4:6e:1f": "TP-Link",
			"c4:e9:84": "Apple",
			"cc:46:d6": "Cisco",
			"d0:17:c2": "Asus",
			"d0:52:a8": "Nest/Google",
			"d4:6d:6d": "D-Link",
			"d8:07:b6": "Belkin",
			"d8:24:bd": "Huawei",
			"dc:a4:ca": "Raspberry Pi",
			"e0:3f:49": "Apple",
			"e0:91:f5": "NetGear",
			"e0:cb:bc": "Asus",
			"e4:8d:8c": "TP-Link",
			"e8:de:27": "TP-Link",
			"ec:08:6b": "D-Link",
			"ec:26:ca": "Cisco",
			"f0:7d:68": "D-Link",
			"f0:b4:29": "Samsung",
			"f0:9f:c2": "Ubiquiti",
			"f4:f2:6d": "NetGear",
			"f8:1a:67": "TP-Link",
			"f8:8e:85": "Intel",
			"fc:0f:4b": "NetGear",
			"fc:25:3f": "NetGear",
			"fc:4d:8c": "Google",
		},
	}
}

// Lookup returns brand for a BSSID
func (db *OUIDatabase) Lookup(bssid string) string {
	if len(bssid) < 8 {
		return "Unknown"
	}
	oui := bssid[:8]
	db.mu.RLock()
	defer db.mu.RUnlock()
	if brand, ok := db.entries[oui]; ok {
		return brand
	}
	return "Unknown"
}
