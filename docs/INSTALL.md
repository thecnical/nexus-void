# NEXUS-VOID OMEGA Installation Guide

## System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| OS | Windows 10 / Ubuntu 20.04 / macOS 12 | Windows 11 / Ubuntu 22.04 |
| CPU | 4 cores | 8+ cores |
| RAM | 4 GB | 16 GB |
| Disk | 10 GB | 50 GB (with all tools) |
| Go | 1.21 | 1.22+ |
| Node.js | 18 | 20+ (for dashboard) |

## Quick Install

### Windows (PowerShell)
```powershell
# One-liner install
irm https://raw.githubusercontent.com/nexus-void/nexus-void/main/install.ps1 | iex

# Or manual
git clone https://github.com/nexus-void/nexus-void.git
cd nexus-void
.\install.ps1
```

### Linux / macOS
```bash
# Clone repository
git clone https://github.com/nexus-void/nexus-void.git
cd nexus-void

# Build binary
go build -o nexus-void ./cmd/nexus-void

# Initialize brain
./nexus-void init

# Install external tools
./nexus-void install
```

## Manual Step-by-Step

### 1. Install Go
```bash
# Linux
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# macOS
brew install go

# Windows
# Download from https://go.dev/dl/
```

### 2. Install Node.js (for Dashboard)
```bash
# Linux
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# macOS
brew install node

# Windows
# Download from https://nodejs.org/
```

### 3. Clone and Build
```bash
git clone https://github.com/nexus-void/nexus-void.git
cd nexus-void

# Download dependencies
go mod tidy

# Build main binary
go build -o nexus-void ./cmd/nexus-void

# Build backend (optional, for cloud deployment)
cd backend
go mod tidy
go build -o server ./cmd/server
cd ..
```

### 4. Initialize Brain
```bash
./nexus-void init
```

This creates:
- `~/.nexus-void/brain/knowledge_graph.db` — SQLite brain database
- `~/.nexus-void/sessions/` — Session storage
- `~/.nexus-void/target_dna/` — Target profiles

### 5. Install External Tools
```bash
./nexus-void install
```

This auto-installs 47 external tools:
- Nmap, SQLMap, Metasploit
- Nuclei, Subfinder, Amass
- BloodHound, Impacket
- And more...

### 6. Verify Installation
```bash
./nexus-void doctor
```

Expected output:
```
[+] NEXUS-VOID Doctor - Self-Test System
[+] Checking core systems...
[+] Brain Database: OK
[+] Network Scanner: OK
[+] Web Tools: OK
[+] Crypto Module: OK
[+] AI Engine: OK
[+] 19/20 systems operational
```

## Dashboard Setup

### 1. Install Dashboard Dependencies
```bash
cd dashboard
npm install --legacy-peer-deps
```

### 2. Start Backend Server
```bash
# Terminal 1
cd backend
./server -addr localhost:8080
```

### 3. Start Dashboard
```bash
# Terminal 2
cd dashboard
npm start
```

### 4. Access Dashboard
Open `http://localhost:3000` in your browser.

## Docker Deployment

### Build Image
```bash
docker build -t nexus-void .
```

### Run Container
```bash
docker run -p 8080:8080 -v ~/.nexus-void:/root/.nexus-void nexus-void
```

### Docker Compose (Full Stack)
```yaml
version: '3.8'
services:
  nexus-void:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ~/.nexus-void:/root/.nexus-void
    environment:
      - NEXUS_AI_OPENROUTER_KEY=${OPENROUTER_KEY}
  
  dashboard:
    build: ./dashboard
    ports:
      - "3000:3000"
    environment:
      - REACT_APP_WS_URL=ws://localhost:8080/ws
```

## Cloud Deployment (Render)

### 1. Push Backend to Private Repo
```bash
cd backend
git init
git remote add origin https://github.com/YOURNAME/nexus-void-backend.git
git add .
git commit -m "Initial backend deployment"
git push -u origin main
```

### 2. Deploy on Render
- Create new Web Service
- Connect your private GitHub repo
- Build Command: `go build -o server ./cmd/server`
- Start Command: `./server -addr :10000`
- Environment: `GO_ENV=production`

### 3. Update Dashboard
```bash
# In dashboard/.env
REACT_APP_WS_URL=wss://your-render-app.onrender.com/ws
REACT_APP_API_URL=https://your-render-app.onrender.com/api
```

## Troubleshooting

### Build Errors
```bash
# Missing dependencies
go mod tidy

# Outdated Go version
go version  # Should be 1.22+

# Permission denied (Linux/macOS)
chmod +x nexus-void
```

### Dashboard Errors
```bash
# Node modules corrupted
rm -rf dashboard/node_modules dashboard/package-lock.json
cd dashboard && npm install --legacy-peer-deps

# Port already in use
# Change port in dashboard/package.json or backend start command
```

### Brain Database Locked
```bash
# Kill any running nexus-void processes
pkill -f nexus-void

# Remove lock file
rm ~/.nexus-void/brain/*.db-journal
```

### External Tool Installation Failed
```bash
# Manual install for missing tools
sudo apt-get install nmap sqlmap metasploit-framework

# Or check individual tool docs
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `NEXUS_AI_OPENROUTER_KEY` | OpenRouter API key | No |
| `NEXUS_AI_GROQ_KEY` | Groq API key | No |
| `NEXUS_HOME` | Custom brain directory | No |
| `PORT` | Server port | No (default: 8080) |
| `GO_ENV` | Environment mode | No |

## Updating

```bash
# Pull latest code
git pull origin main

# Rebuild
go build -o nexus-void ./cmd/nexus-void

# Update external tools
./nexus-void install

# Verify
./nexus-void doctor
```

## Uninstall

```bash
# Remove binary
rm nexus-void

# Remove brain data
rm -rf ~/.nexus-void

# Remove external tools (optional)
# See install.ps1 for tool locations
```
