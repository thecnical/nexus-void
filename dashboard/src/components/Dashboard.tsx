import React from 'react';
import { AlertTriangle, Bug, Globe, Server } from 'lucide-react';
import { NexusState } from '../App';

interface DashboardProps {
  state: NexusState;
}

const Dashboard: React.FC<DashboardProps> = ({ state }) => {
  const cards = [
    {
      title: 'Web Attack',
      icon: Globe,
      color: '#ff6b6b',
      items: ['SQLiReaper', 'XSSHunter', 'LFIRaider', 'SSRFLeech', 'JWTBreaker'],
      count: 5,
    },
    {
      title: 'Network',
      icon: Server,
      color: '#4ecdc4',
      items: ['PortBreacher', 'UDPFlooder', 'SYN-Blitz', 'Service-ID', 'OS-Fingerprint'],
      count: 5,
    },
    {
      title: 'Cloud',
      icon: Server,
      color: '#ffd93d',
      items: ['AWSBreaker', 'S3Scanner', 'EC2-Metadata', 'IAM-Enum', 'Lambda-Test'],
      count: 5,
    },
    {
      title: 'Post-Exploit',
      icon: Bug,
      color: '#ff6600',
      items: ['Persistence', 'Lateral-Move', 'Cred-Dump', 'User-Enum', 'Exfiltrate'],
      count: 5,
    },
    {
      title: 'Active Directory',
      icon: Server,
      color: '#00ff88',
      items: ['User-Enum', 'Kerberoast', 'ASREPRoast', 'ACL-Abuse', 'DCSync'],
      count: 5,
    },
    {
      title: 'OSINT',
      icon: Globe,
      color: '#a855f7',
      items: ['Email-Harvest', 'Subdomain-Disco', 'GitHub-Dorks', 'Social-Media', 'Metadata-Extract'],
      count: 5,
    },
  ];

  return (
    <div>
      <h2 style={{ color: '#00ff88', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
        <AlertTriangle size={20} />
        Attack Surface Overview
      </h2>
      
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
        gap: '1rem',
      }}>
        {cards.map((card, i) => (
          <div 
            key={i}
            style={{
              background: '#1a1a2e',
              border: `1px solid ${card.color}33`,
              borderRadius: '8px',
              padding: '1rem',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.75rem' }}>
              <card.icon size={18} color={card.color} />
              <span style={{ fontWeight: 'bold', color: card.color }}>{card.title}</span>
              <span style={{ 
                marginLeft: 'auto', 
                background: `${card.color}22`, 
                color: card.color,
                padding: '0.15rem 0.5rem',
                borderRadius: '4px',
                fontSize: '0.75rem',
              }}>
                {card.count} tools
              </span>
            </div>
            <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
              {card.items.map((item, j) => (
                <li key={j} style={{ 
                  padding: '0.3rem 0', 
                  color: '#aaa',
                  fontSize: '0.85rem',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '0.5rem',
                }}>
                  <span style={{ color: card.color, fontSize: '0.6rem' }}>●</span>
                  {item}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Dashboard;
