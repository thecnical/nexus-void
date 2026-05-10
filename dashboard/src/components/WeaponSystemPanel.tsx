import React from 'react';
import { AgentStatus } from '../App';
import { Zap, Shield, Crosshair, Activity, AlertTriangle, CheckCircle, Circle } from 'lucide-react';

interface Props {
  system: string;
  agents: AgentStatus[];
}

const systemConfig: Record<string, { name: string; color: string; icon: any }> = {
  ether: { name: 'ETHERBREACH', color: '#00ff88', icon: Zap },
  mobile: { name: 'MOBILEBREACH', color: '#ff6b6b', icon: Shield },
  osint: { name: 'OSINTBREACH', color: '#00d4ff', icon: Crosshair },
  net: { name: 'NETBREACH', color: '#ff00ff', icon: Activity },
  crypto: { name: 'CRYPTOBREACH', color: '#ffd700', icon: Shield },
  cloud: { name: 'CLOUDBREACH', color: '#00ffff', icon: Zap },
  web: { name: 'WEBBREACH', color: '#ff8800', icon: Activity },
  api: { name: 'APIBREACH', color: '#8800ff', icon: Crosshair },
};

export default function WeaponSystemPanel({ system, agents }: Props) {
  const config = systemConfig[system] || systemConfig.ether;
  const Icon = config.icon;
  const active = agents.filter(a => a.status === 'active').length;
  const idle = agents.filter(a => a.status === 'idle').length;

  return (
    <div style={{ padding: '1rem' }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '1rem',
        marginBottom: '1.5rem',
        padding: '1rem',
        background: '#1a1a2e',
        borderRadius: '8px',
        border: `1px solid ${config.color}33`,
      }}>
        <Icon size={32} color={config.color} />
        <div>
          <h2 style={{ margin: 0, color: config.color, fontSize: '1.5rem' }}>{config.name}</h2>
          <p style={{ margin: 0, color: '#888', fontSize: '0.85rem' }}>
            {agents.length} Agents | {active} Active | {idle} Idle
          </p>
        </div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
        gap: '1rem',
      }}>
        {agents.map(agent => (
          <AgentCard key={agent.name} agent={agent} color={config.color} />
        ))}
      </div>
    </div>
  );
}

function AgentCard({ agent, color }: { agent: AgentStatus; color: string }) {
  const isActive = agent.status === 'active';
  const isOnline = agent.status === 'online';

  return (
    <div style={{
      background: '#12122a',
      borderRadius: '8px',
      padding: '1rem',
      border: `1px solid ${isActive ? color : '#2a2a4e'}`,
      transition: 'all 0.3s',
    }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: '0.5rem',
      }}>
        <span style={{
          fontWeight: 'bold',
          color: isActive ? color : '#e0e0e0',
          fontSize: '0.9rem',
        }}>{agent.name}</span>
        <StatusBadge status={agent.status} />
      </div>

      <div style={{
        height: '4px',
        background: '#1a1a2e',
        borderRadius: '2px',
        overflow: 'hidden',
        marginBottom: '0.5rem',
      }}>
        <div style={{
          width: `${agent.progress}%`,
          height: '100%',
          background: isActive ? color : '#2a2a4e',
          borderRadius: '2px',
          transition: 'width 0.5s',
        }} />
      </div>

      <div style={{
        fontSize: '0.75rem',
        color: '#666',
        maxHeight: '80px',
        overflowY: 'auto',
        fontFamily: "'Roboto Mono', monospace",
      }}>
        {agent.output.slice(-3).map((line, i) => (
          <div key={i}>{line}</div>
        ))}
        {agent.output.length === 0 && <em>No output yet...</em>}
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: '#00ff88',
    online: '#00d4ff',
    idle: '#666',
    error: '#ff6b6b',
  };

  return (
    <span style={{
      display: 'flex',
      alignItems: 'center',
      gap: '0.25rem',
      fontSize: '0.75rem',
      color: colors[status] || '#888',
    }}>
      {status === 'active' && <Activity size={12} />}
      {status === 'online' && <CheckCircle size={12} />}
      {status === 'idle' && <Circle size={12} />}
      {status === 'error' && <AlertTriangle size={12} />}
      {status.toUpperCase()}
    </span>
  );
}
