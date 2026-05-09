import React from 'react';
import { Brain, Activity } from 'lucide-react';
import { AgentStatus } from '../App';

interface AgentPanelProps {
  agents: AgentStatus[];
}

const AgentPanel: React.FC<AgentPanelProps> = ({ agents }) => {
  const statusColors: Record<string, string> = {
    idle: '#888',
    scanning: '#4ecdc4',
    analyzing: '#a855f7',
    exploiting: '#ff6b6b',
    evolving: '#ffd93d',
    complete: '#00ff88',
  };

  return (
    <div>
      <h2 style={{ color: '#00ff88', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <Brain size={20} />
        AI Agents
      </h2>
      
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))',
        gap: '1rem',
      }}>
        {agents.map((agent, i) => (
          <div 
            key={i}
            style={{
              background: '#1a1a2e',
              border: '1px solid #2a2a4e',
              borderRadius: '8px',
              padding: '1rem',
            }}
          >
            <div style={{ 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'space-between',
              marginBottom: '0.75rem',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Activity size={16} color={statusColors[agent.status] || '#888'} />
                <span style={{ fontWeight: 'bold', color: '#e0e0e0' }}>{agent.name}</span>
              </div>
              <span style={{ 
                color: statusColors[agent.status] || '#888',
                fontSize: '0.8rem',
                textTransform: 'uppercase',
              }}>
                {agent.status}
              </span>
            </div>
            
            <div style={{
              background: '#0f0f1e',
              borderRadius: '4px',
              height: '4px',
              marginBottom: '0.75rem',
              overflow: 'hidden',
            }}>
              <div style={{
                width: `${agent.progress}%`,
                height: '100%',
                background: statusColors[agent.status] || '#888',
                transition: 'width 0.3s',
              }} />
            </div>
            
            <div style={{
              background: '#0f0f1e',
              borderRadius: '4px',
              padding: '0.5rem',
              height: '120px',
              overflowY: 'auto',
              fontSize: '0.75rem',
              fontFamily: "'Roboto Mono', monospace",
            }}>
              {agent.output.length === 0 ? (
                <span style={{ color: '#555' }}>Waiting for activity...</span>
              ) : (
                agent.output.slice(-5).map((line, j) => (
                  <div key={j} style={{ color: '#aaa', padding: '0.1rem 0' }}>
                    {line}
                  </div>
                ))
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default AgentPanel;
