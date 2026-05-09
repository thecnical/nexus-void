# NEXUS-VOID Makefile
# One command to install everything

BINARY_NAME=nexus-void
INSTALL_DIR=$(HOME)/.nexus-void
GO_VERSION=1.23

.PHONY: all install build clean test doctor check-go check-docker

all: check-go install

install: check-go check-docker build
	@echo "[+] NEXUS-VOID installation starting..."
	@echo "[+] Creating directories..."
	@mkdir -p $(INSTALL_DIR)/bin
	@mkdir -p $(INSTALL_DIR)/brain
	@mkdir -p $(INSTALL_DIR)/brain/sessions
	@mkdir -p $(INSTALL_DIR)/brain/target_dna
	@mkdir -p $(INSTALL_DIR)/brain/exploit_dna
	@mkdir -p $(INSTALL_DIR)/brain/ai_cache
	@mkdir -p $(INSTALL_DIR)/brain/learned_strategies
	@mkdir -p $(INSTALL_DIR)/external_tools
	@mkdir -p $(INSTALL_DIR)/cache
	@mkdir -p $(INSTALL_DIR)/logs
	@echo "[+] Installing binary..."
	@cp bin/$(BINARY_NAME) $(INSTALL_DIR)/bin/
	@chmod +x $(INSTALL_DIR)/bin/$(BINARY_NAME)
	@echo "[+] Setting up PATH..."
	@echo "export PATH=\"$$HOME/.nexus-void/bin:$$PATH\"" >> $(HOME)/.bashrc 2>/dev/null || true
	@echo "[+] Running self-test..."
	@$(INSTALL_DIR)/bin/$(BINARY_NAME) doctor || true
	@echo "[+] Installation complete!"
	@echo "[+] Run: nexus-void --help"
	@echo "[+] Run: nexus-void apocalypse https://target.com"

build: check-go
	@echo "[+] Building NEXUS-VOID..."
	@mkdir -p bin
	@go build -ldflags="-s -w -X main.Version=1.0.0-OMEGA" -o bin/$(BINARY_NAME) ./cmd/nexus-void
	@echo "[+] Build complete: bin/$(BINARY_NAME)"

build-all: check-go
	@echo "[+] Building for all platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/nexus-void
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/nexus-void
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/nexus-void
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/nexus-void
	@echo "[+] All builds complete"

check-go:
	@which go > /dev/null 2>&1 || (echo "[-] Go not found. Installing..." && ./scripts/install-go.sh)
	@go version | grep "go$(GO_VERSION)" > /dev/null 2>&1 || (echo "[-] Go version mismatch. Installing..." && ./scripts/install-go.sh)
	@echo "[+] Go is ready"

check-docker:
	@which docker > /dev/null 2>&1 || echo "[!] Docker not found. Some external tools require Docker."

test:
	@go test -v ./pkg/... ./internal/... ./cmd/...

clean:
	@rm -rf bin/
	@echo "[+] Cleaned build artifacts"

doctor:
	@go run ./cmd/nexus-void doctor

.DEFAULT_GOAL := install
