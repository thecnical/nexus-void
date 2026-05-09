import React, { useRef, useEffect } from 'react';
import { Terminal } from 'lucide-react';

interface LiveFeedProps {
  messages: string[];
}

const LiveFeed: React.FC<LiveFeedProps> = ({ messages }) => {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const getMessageColor = (msg: string): string => {
    if (msg.includes('CRITICAL') || msg.includes('PROVEN')) return '#ff0000';
    if (msg.includes('HIGH')) return '#ff6600';
    if (msg.includes('EXPLOIT') || msg.includes('bypass')) return '#ffd93d';
    if (msg.includes('RECON')) return '#4ecdc4';
    if (msg.includes('SHIELD')) return '#00ff88';
    if (msg.includes('VULN')) return '#a855f7';
    return '#888';
  };

  return (
    <div>
      <h2 style={{ color: '#00ff88', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <Terminal size={20} />
        Live Feed
      </h2>
      
      <div 
        ref={scrollRef}
        style={{
          background: '#0f0f1e',
          border: '1px solid #2a2a4e',
          borderRadius: '8px',
          padding: '1rem',
          height: '500px',
          overflowY: 'auto',
          fontFamily: "'Roboto Mono', monospace",
          fontSize: '0.85rem',
        }}
      >
        {messages.length === 0 && (
          <div style={{ color: '#555', textAlign: 'center', marginTop: '200px' }}>
            Waiting for activity...
          </div>
        )}
        
        {messages.map((msg, i) => (
          <div 
            key={i}
            style={{
              padding: '0.3rem 0',
              color: getMessageColor(msg),
              borderBottom: '1px solid #1a1a2e',
            }}
          >
            <span style={{ color: '#555', marginRight: '0.5rem' }}>
              {new Date().toLocaleTimeString()}
            </span>
            {msg}
          </div>
        ))}
      </div>
    </div>
  );
};

export default LiveFeed;
