import React, { useState, useEffect } from 'react';
import { Terminal, Shield, Activity, Zap, Brain, Radio, ScanLine, Target, Globe, Cpu, Bug, Sword, Eye, Lock, Server, AlertTriangle, Crosshair } from 'lucide-react';
import Dashboard from './components/Dashboard';
import LiveFeed from './components/LiveFeed';
import StatsPanel from './components/StatsPanel';
import AgentPanel from './components/AgentPanel';
import { useWebSocket } from './hooks/useWebSocket';

export interface AgentStatus {
  name: string;
  status: string;
  output: string[];
  progress: number;
}

export interface Finding {
  id: string;
  type: string;
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info';
  url: string;
  description: string;
  timestamp: string;
}

export interface NexusState {
  target: string;
  elapsed: string;
  findings: number;
  exploits: number;
  evolutions: number;
  agents: AgentStatus[];
  messages: string[];
  clients: number;
  brainStats: any;
  sessions: any[];
  strategies: any[];
  activeFindings: Finding[];
  scanning: boolean;
}

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';
const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';

const initialState: NexusState = {
  target: 'target.com',
  elapsed: '00:00:00',
  findings: 0,
  exploits: 0,
  evolutions: 0,
  agents: [
    { name: 'RECON-OMEGA', status: 'idle', output: [], progress: 0 },
    { name: 'VULN-SENTINEL', status: 'idle', output: [], progress: 0 },
    { name: 'EXPLOIT-APOCALYPSE', status: 'idle', output: [], progress: 0 },
    { name: 'PERSISTENCE-DAEMON', status: 'idle', output: [], progress: 0 },
    { name: 'SHIELD-BREAKER', status: 'idle', output: [], progress: 0 },
    { name: 'C2-NEXUS', status: 'idle', output: [], progress: 0 },
  ],
  messages: [],
  clients: 0,
  brainStats: {},
  sessions: [],
  strategies: [],
  activeFindings: [],
  scanning: false,
};

function App() {
  const [state, setState] = useState<NexusState>(initialState);
  const [activeTab, setActiveTab] = useState<'dashboard' | 'agents' | 'feed'>('dashboard');
  const [scanTarget, setScanTarget] = useState('');
  const [scanning, setScanning] = useState(false);

  const ws = useWebSocket(WS_URL);

  useEffect(() => {
    if (ws.lastMessage) {
      try {
        const msg = JSON.parse(ws.lastMessage.data);
        if (msg.type === 'status_update') {
          const data = msg.data || {};
          setState(prev => ({
            ...prev,
            clients: data.clients || prev.clients,
            brainStats: data.brain || prev.brainStats,
            agents: data.agents?.map((a: any) => ({
              name: a.Type || a.name,
              status: a.Status || a.status || 'idle',
              output: a.Output || a.output || [],
              progress: a.Progress || a.progress || 0,
            })) || prev.agents,
          }));
        } else if (msg.type === 'recon_complete') {
          const data = msg.data || {};
          setState(prev => ({
            ...prev,
            findings: prev.findings + 1,
            messages: [...prev.messages, `Recon complete: ${data.target || 'unknown'} (${data.status_code || '?'})`].slice(-100),
          }));
        } else {
          setState(prev => ({
            ...prev,
            messages: [...prev.messages, JSON.stringify(msg)].slice(-100),
          }));
        }
      } catch {
        setState(prev => ({
          ...prev,
          messages: [...prev.messages, String(ws.lastMessage?.data)].slice(-100),
        }));
      }
    }
  }, [ws.lastMessage]);

  const startScan = async () => {
    if (!scanTarget) return;
    setScanning(true);
    try {
      const resp = await fetch(`${API_URL}/recon`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: scanTarget }),
      });
      const data = await resp.json();
      setState(prev => ({
        ...prev,
        target: scanTarget,
        messages: [...prev.messages, `Scan started: ${scanTarget} (${data.status})`].slice(-100),
      }));
    } catch (e) {
      setState(prev => ({
        ...prev,
        messages: [...prev.messages, `Scan error: ${String(e)}`].slice(-100),
      }));
    } finally {
      setScanning(false);
    }
  };

  const styles = {
    app: {
      minHeight: '100vh',
      background: '#0a0a1a',
      color: '#e0e0e0',
      fontFamily: "'Segoe UI', 'Roboto Mono', monospace",
    },
    header: {
      background: '#1a1a2e',
      borderBottom: '1px solid #2a2a4e',
      padding: '1rem 2rem',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
    },
    title: {
      fontSize: '1.5rem',
      fontWeight: 'bold' as const,
      color: '#ff6b6b',
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
    },
    subtitle: {
      fontSize: '0.8rem',
      color: '#00ff88',
    },
    nav: {
      display: 'flex',
      gap: '1rem',
    },
    navButton: (active: boolean) => ({
      background: active ? '#2a2a4e' : 'transparent',
      border: '1px solid #2a2a4e',
      color: active ? '#00ff88' : '#888',
      padding: '0.5rem 1rem',
      borderRadius: '4px',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      fontSize: '0.9rem',
      transition: 'all 0.2s',
    }),
    main: {
      padding: '1.5rem',
      maxWidth: '1400px',
      margin: '0 auto',
    },
    connectionStatus: {
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      fontSize: '0.8rem',
      color: ws.readyState === 1 ? '#00ff88' : '#ff6b6b',
    },
    scanBar: {
      display: 'flex',
      gap: '0.5rem',
      padding: '1rem 2rem',
      background: '#12122a',
      borderBottom: '1px solid #2a2a4e',
      alignItems: 'center',
    },
    scanInput: {
      flex: 1,
      background: '#1a1a2e',
      border: '1px solid #2a2a4e',
      color: '#e0e0e0',
      padding: '0.5rem 1rem',
      borderRadius: '4px',
      fontSize: '0.9rem',
      fontFamily: "'Roboto Mono', monospace",
      outline: 'none',
    },
    scanButton: {
      background: '#ff6b6b',
      border: 'none',
      color: '#fff',
      padding: '0.5rem 1.5rem',
      borderRadius: '4px',
      cursor: 'pointer',
      display: 'flex',
      alignItems: 'center',
      gap: '0.5rem',
      fontWeight: 'bold',
      fontSize: '0.85rem',
      opacity: scanning || !scanTarget ? 0.5 : 1,
    },
  };

  return (
    <div style={styles.app}>
      <header style={styles.header}>
        <div>
          <div style={styles.title}>
            <Zap size={24} />
            NEXUS VOID
          </div>
          <div style={styles.subtitle}>Autonomous Swarm Intelligence Weapon | Created by Chandan Pandey | cybermindcli.com</div>
        </div>
        
        <nav style={styles.nav}>
          <button
            style={styles.navButton(activeTab === 'dashboard')}
            onClick={() => setActiveTab('dashboard')}
          >
            <Activity size={16} /> Dashboard
          </button>
          <button
            style={styles.navButton(activeTab === 'agents')}
            onClick={() => setActiveTab('agents')}
          >
            <Brain size={16} /> Agents
          </button>
          <button
            style={styles.navButton(activeTab === 'feed')}
            onClick={() => setActiveTab('feed')}
          >
            <Terminal size={16} /> Live Feed
          </button>
        </nav>

        <div style={styles.connectionStatus}>
          <Radio size={14} />
          {ws.readyState === 1 ? 'CONNECTED' : 'DISCONNECTED'}
          <span style={{ marginLeft: '1rem', color: '#888', fontSize: '0.75rem' }}>
            Clients: {state.clients}
          </span>
        </div>
      </header>

      <div style={styles.scanBar}>
        <Target size={16} style={{ color: '#ff6b6b' }} />
        <input
          type="text"
          placeholder="Enter target (e.g. example.com)"
          value={scanTarget}
          onChange={e => setScanTarget(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && startScan()}
          style={styles.scanInput}
        />
        <button onClick={startScan} disabled={scanning || !scanTarget} style={styles.scanButton}>
          <ScanLine size={16} />
          {scanning ? 'SCANNING...' : 'SCAN'}
        </button>
      </div>

      <main style={styles.main}>
        <StatsPanel state={state} />
        
        {activeTab === 'dashboard' && <Dashboard state={state} />}
        {activeTab === 'agents' && <AgentPanel agents={state.agents} />}
        {activeTab === 'feed' && <LiveFeed messages={state.messages} />}
      </main>
    </div>
  );
}

export default App;
