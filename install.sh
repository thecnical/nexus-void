#!/bin/bash
set -e

# ╔═══════════════════════════════════════════════════════════════╗
# ║     NEXUS VOID — Autonomous Swarm Intelligence Weapon        ║
# ║     Linux Installer | Created by Chandan Pandey               ║
# ║     cybermindcli.com — Must Visit                           ║
# ╚═══════════════════════════════════════════════════════════════╝

NEXUS_VERSION="3.0.0"
NEXUS_REPO="https://github.com/thecnical/nexus-void.git"
BACKEND_REPO=${BACKEND_REPO:-"https://github.com/mrgithacks/Nexus-Void-backend.git"}
INSTALL_DIR="/opt/nexus-void"
BIN_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
PURPLE='\033[0;35m'
NC='\033[0m'
BOLD='\033[1m'

print_banner() {
    echo ""
    echo -e "${CYAN}                  ▄▄▄▄▄▄▄▄${NC}"
    echo -e "${CYAN}              ▄▄▀▀${PURPLE}▓▓▓▓▓▓▓▓${CYAN}▀▀▄▄${NC}"
    echo -e "${CYAN}            ▄▀${PURPLE}▓▓▓▓▓▓▓▓▓▓▓▓▓▓${CYAN}▀▄${NC}"
    echo -e "${CYAN}           █   ▄▄${PURPLE}▓▓▓▓${NC}▄▄   █${NC}"
    echo -e "${CYAN}          █   █${GREEN}◉${NC}█  █${GREEN}◉${NC}█   █${NC}"
    echo -e "${CYAN}          █     ▀▀${NC}██${PURPLE}▀▀     █${NC}"
    echo -e "${CYAN}           █  ▄▄${RED}▀▀▀▀${NC}▄▄  █${NC}"
    echo -e "${CYAN}            ▀▄   ▀▀${NC}██${PURPLE}▀▀   ▄▀${NC}"
    echo -e "${CYAN}              ▀▀▄▄${PURPLE}▓▓▓▓${CYAN}▄▄▀▀${NC}"
    echo -e "${CYAN}                  ▀▀▀▀${NC}"
    echo ""
    echo -e "${BOLD}${PURPLE}    ╔════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${PURPLE}    ║${CYAN}            N E X U S   V O I D                   ${PURPLE}║${NC}"
    echo -e "${BOLD}${PURPLE}    ║${GREEN}      AUTONOMOUS SWARM INTELLIGENCE WEAPON          ${PURPLE}║${NC}"
    echo -e "${BOLD}${PURPLE}    ╠════════════════════════════════════════════════════╣${NC}"
    echo -e "${BOLD}${PURPLE}    ║${YELLOW}  189 Tools | 16 Categories | 6 AI Agents | v${NEXUS_VERSION}  ${PURPLE}║${NC}"
    echo -e "${BOLD}${PURPLE}    ╚════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BOLD}${CYAN}         Created by Chandan Pandey — Architect of the Swarm${NC}"
    echo -e "${RED}              cybermindcli.com  —  Must Visit${NC}"
    echo ""
}

check_dependencies() {
    echo -e "${CYAN}[*] Checking dependencies...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${YELLOW}[!] Go not found. Installing Go...${NC}"
        wget -q https://go.dev/dl/go1.24.2.linux-amd64.tar.gz -O /tmp/go.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf /tmp/go.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo -e "${GREEN}[+] Go installed: $(go version)${NC}"
    else
        echo -e "${GREEN}[+] Go found: $(go version)${NC}"
    fi
    
    if ! command -v git &> /dev/null; then
        echo -e "${YELLOW}[!] Git not found. Installing...${NC}"
        sudo apt-get update && sudo apt-get install -y git
    fi
    
    echo -e "${GREEN}[+] All dependencies satisfied.${NC}"
    echo ""
}

install_nexus() {
    echo -e "${CYAN}[*] Installing Nexus Void v${NEXUS_VERSION}...${NC}"
    
    # Create install directory
    sudo mkdir -p ${INSTALL_DIR}
    sudo chown $(whoami):$(whoami) ${INSTALL_DIR}
    
    # Clone repository
    if [ -d "${INSTALL_DIR}/.git" ]; then
        echo -e "${YELLOW}[*] Existing installation found. Updating...${NC}"
        cd ${INSTALL_DIR}
        git pull origin main
    else
        echo -e "${CYAN}[*] Cloning repository...${NC}"
        git clone ${NEXUS_REPO} ${INSTALL_DIR}
    fi
    
    cd ${INSTALL_DIR}
    
    # Build main CLI
    echo -e "${CYAN}[*] Building Nexus Void CLI...${NC}"
    go work sync
    go build -ldflags="-s -w" -o nexus-void ./cmd/nexus-void
    
    # Clone and build backend from separate repo
    echo -e "${CYAN}[*] Cloning backend server repository...${NC}"
    git clone ${BACKEND_REPO} ${INSTALL_DIR}/backend
    
    echo -e "${CYAN}[*] Building Nexus Void Backend Server...${NC}"
    cd ${INSTALL_DIR}/backend
    go mod tidy
    go build -ldflags="-s -w" -o nexus-server ./cmd/server
    
    # Create symlinks
    echo -e "${CYAN}[*] Creating system links...${NC}"
    sudo ln -sf ${INSTALL_DIR}/nexus-void ${BIN_DIR}/nexus-void
    sudo ln -sf ${INSTALL_DIR}/backend/nexus-server ${BIN_DIR}/nexus-server
    
    # Create brain directory
    mkdir -p ~/.nexus-void/brain
    
    echo -e "${GREEN}[+] Nexus Void installed successfully!${NC}"
    echo ""
}

setup_systemd() {
    echo -e "${CYAN}[*] Setting up systemd service for backend...${NC}"
    
    sudo tee /etc/systemd/system/nexus-void-backend.service > /dev/null <<EOF
[Unit]
Description=Nexus Void Backend Server
After=network.target

[Service]
Type=simple
User=$(whoami)
WorkingDirectory=${INSTALL_DIR}/backend
ExecStart=${INSTALL_DIR}/backend/nexus-server -addr :8080 -brain /home/$(whoami)/.nexus-void/brain
Restart=always
RestartSec=5
Environment=GO_ENV=production

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable nexus-void-backend
    echo -e "${GREEN}[+] Systemd service created.${NC}"
    echo ""
}

print_usage() {
    echo -e "${BOLD}${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${GREEN}║              INSTALLATION COMPLETE                            ║${NC}"
    echo -e "${BOLD}${GREEN}╠═══════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${BOLD}${GREEN}║${CYAN}  CLI Commands:                                               ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    nexus-void --help            → Show all commands           ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    nexus-void chat               → Launch AI chat interface    ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    nexus-void scan <target>      → Quick reconnaissance      ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    nexus-void apocalypse <target>→ Full autonomous assault   ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${CYAN}  Backend Server:                                              ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    nexus-server -addr :8080     → Start API + WebSocket      ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    sudo systemctl start nexus-void-backend → Daemon mode     ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${CYAN}  Dashboard:                                                   ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    cd ${INSTALL_DIR}/dashboard && npm install && npm run dev   ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}    Open http://localhost:5173 in your browser                  ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}║${PURPLE}  Created by Chandan Pandey | cybermindcli.com               ${GREEN}║${NC}"
    echo -e "${BOLD}${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Main
print_banner
check_dependencies
install_nexus
setup_systemd
print_usage

echo -e "${GREEN}[+] Ready to breach. The swarm awaits your command.${NC}"
echo ""
