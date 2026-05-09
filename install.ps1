# NEXUS-VOID Auto-Installer for Windows
# Run: .\install.ps1
# This installs EVERYTHING automatically

$ErrorActionPreference = "Continue"
$NVHome = "$env:USERPROFILE\.nexus-void"
$NVBin = "$NVHome\bin"
$NVBrain = "$NVHome\brain"
$NVExternal = "$NVHome\external_tools"
$NVCache = "$NVHome\cache"
$NVLogs = "$NVHome\logs"

function Write-Status($msg) { Write-Host "[+] $msg" -ForegroundColor Green }
function Write-Warning($msg) { Write-Host "[!] $msg" -ForegroundColor Yellow }
function Write-Error($msg) { Write-Host "[-] $msg" -ForegroundColor Red }

Write-Host @"
   _   _                 __     __        _     _
  | \ | |_   ___   _____  \ \   / /__  ___| | __| |
  |  \| | | | \ \ / / _ \  \ \ / / _ \/ _ \ |/ _' |
  | |\  | |_| |\ V / (_) |  \ V /  __/  __/ | (_| |
  |_| \_|\__, | \_/ \___/    \_/ \___|\___|_|\__,_|
         |___/
   OMEGA - The Autonomous Cyber-Weapon
"@ -ForegroundColor Cyan

Write-Status "NEXUS-VOID Auto-Installer Starting..."

# Check Go
$goVersion = go version 2>$null
if (-not $goVersion) {
    Write-Status "Go not found. Downloading Go 1.23..."
    $goUrl = "https://go.dev/dl/go1.23.4.windows-amd64.msi"
    $goInstaller = "$env:TEMP\go-installer.msi"
    Invoke-WebRequest -Uri $goUrl -OutFile $goInstaller
    Start-Process -FilePath "msiexec.exe" -ArgumentList "/i", $goInstaller, "/quiet", "/norestart" -Wait
    $env:PATH = "$env:PATH;C:\Program Files\Go\bin"
    Write-Status "Go installed. Restart PowerShell if needed."
} else {
    Write-Status "Go found: $goVersion"
}

# Check Docker
$dockerVersion = docker --version 2>$null
if (-not $dockerVersion) {
    Write-Warning "Docker not found. Some external tools will use binary downloads instead."
} else {
    Write-Status "Docker found: $dockerVersion"
}

# Create directories
Write-Status "Creating directories..."
@($NVHome, $NVBin, $NVBrain, "$NVBrain\sessions", "$NVBrain\target_dna", "$NVBrain\exploit_dna", 
  "$NVBrain\ai_cache", "$NVBrain\learned_strategies", $NVExternal, $NVCache, $NVLogs) | ForEach-Object {
    New-Item -ItemType Directory -Path $_ -Force | Out-Null
}

# Build the project
Write-Status "Building NEXUS-VOID..."
go build -ldflags="-s -w -X main.Version=1.0.0-OMEGA" -o "$NVBin\nexus-void.exe" .\cmd\nexus-void
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed. Check Go installation."
    exit 1
}
Write-Status "Binary built: $NVBin\nexus-void.exe"

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if (-not $currentPath.Contains($NVBin)) {
    [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$NVBin", "User")
    Write-Status "Added $NVBin to PATH"
}

# Initialize brain
Write-Status "Initializing Knowledge Graph..."
& "$NVBin\nexus-void.exe" init --brain-only

# Self-test
Write-Status "Running self-test..."
& "$NVBin\nexus-void.exe" doctor

Write-Status "Installation Complete!"
Write-Host ""
Write-Host "Run these commands:" -ForegroundColor Cyan
Write-Host "  nexus-void --help"
Write-Host "  nexus-void apocalypse https://target.com"
Write-Host "  nexus-void doctor"
Write-Host ""
Write-Host "Brain location: $NVBrain" -ForegroundColor Gray
Write-Host "Tools location: $NVExternal" -ForegroundColor Gray
