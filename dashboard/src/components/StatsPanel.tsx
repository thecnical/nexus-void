import React from 'react';
import { Target, Clock, AlertTriangle, Zap, Dna } from 'lucide-react';
import { NexusState } from '../App';

interface StatsPanelProps {
  state: NexusState;
}

const StatsPanel: React.FC<StatsPanelProps> = ({ state }) => {
  const stats = [
    { icon: Target, label: 'Target', value: state.target, color: '#4ecdc4' },
    { icon: Clock, label: 'Elapsed', value: state.elapsed, color: '#ffd93d' },
    { icon: AlertTriangle, label: 'Findings', value: state.findings.toString(), color: '#ff6b6b' },
    { icon: Zap, label: 'Exploits', value: state.exploits.toString(), color: '#ff6600' },
    { icon: Dna, label: 'Evolutions', value: state.evolutions.toString(), color: '#00ff88' },
  ];

  return (
    <div style={{ 
      display: 'grid', 
      gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', 
      gap: '1rem',
      marginBottom: '1.5rem',
    }}>
      {stats.map((stat, i) => (
        <div 
          key={i}
          style={{
            background: '#1a1a2e',
            border: '1px solid #2a2a4e',
            borderRadius: '8px',
            padding: '1rem',
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
          }}
        >
          <stat.icon size={24} color={stat.color} />
          <div>
            <div style={{ fontSize: '0.75rem', color: '#888', textTransform: 'uppercase' }}>
              {stat.label}
            </div>
            <div style={{ fontSize: '1.25rem', fontWeight: 'bold', color: stat.color }}>
              {stat.value}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

export default StatsPanel;
