.PHONY: build build-pingmonke build-tailmonke install uninstall clean help lint test install-windows uninstall-windows update

# Variables
BINARY_DIR := ./bin
INSTALL_DIR := ~/.local/bin
PINGMONKE_BIN := pingmonke
TAILMONKE_BIN := tailmonke

# Detect OS for platform-specific commands
UNAME_S := $(shell uname -s)
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
else
    ifeq ($(UNAME_S),Linux)
        DETECTED_OS := Linux
    endif
    ifeq ($(UNAME_S),Darwin)
        DETECTED_OS := Darwin
    endif
endif

# Default target
help:
	@echo "Pingmonke Build System"
	@echo ""
	@echo "Detected OS: $(DETECTED_OS)"
	@echo ""
	@echo "Available targets:"
	@echo "  make build              - Build both binaries to $(BINARY_DIR)/"
	@echo "  make build-pingmonke    - Build only pingmonke"
	@echo "  make build-tailmonke    - Build only tailmonke"
	@echo "  make install            - Install binaries to $(INSTALL_DIR)/"
	@echo "  make uninstall          - Remove installed binaries"
	@echo "  make clean              - Remove built binaries from $(BINARY_DIR)/"
	@echo "  make lint               - Run go linter"
	@echo "  make test               - Run tests"
	@echo "  make update             - Stop service, rebuild, and restart (Linux systemd)"
	@echo ""
	@echo "Windows-specific targets:"
	@echo "  make install-windows    - Install to %USERPROFILE%\\.local\\bin"
	@echo "  make uninstall-windows  - Remove from %USERPROFILE%\\.local\\bin"
	@echo ""

# Build both binaries to local bin directory
build: build-pingmonke build-tailmonke
	@echo "✓ Build complete: $(BINARY_DIR)/$(PINGMONKE_BIN) $(BINARY_DIR)/$(TAILMONKE_BIN)"

# Build pingmonke
build-pingmonke:
	@echo "Building pingmonke..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(PINGMONKE_BIN) ./cmd/pingmonke

# Build tailmonke
build-tailmonke:
	@echo "Building tailmonke..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(TAILMONKE_BIN) ./cmd/tailmonke

# Install to ~/.local/bin (Unix-like systems)
install: build
	@echo "Installing to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_DIR)/$(PINGMONKE_BIN) $(INSTALL_DIR)/$(PINGMONKE_BIN)
	@cp $(BINARY_DIR)/$(TAILMONKE_BIN) $(INSTALL_DIR)/$(TAILMONKE_BIN)
	@chmod +x $(INSTALL_DIR)/$(PINGMONKE_BIN) $(INSTALL_DIR)/$(TAILMONKE_BIN)
	@echo "✓ Installation complete"
	@echo "  $(INSTALL_DIR)/$(PINGMONKE_BIN)"
	@echo "  $(INSTALL_DIR)/$(TAILMONKE_BIN)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH!"

# Install to %USERPROFILE%\.local\bin (Windows)
install-windows: build
	@powershell -Command "if (!(Test-Path \"$$env:USERPROFILE\.local\bin\")) { mkdir \"$$env:USERPROFILE\.local\bin\" | Out-Null }; Copy-Item -Path '$(BINARY_DIR)\$(PINGMONKE_BIN).exe' -Destination \"$$env:USERPROFILE\.local\bin\"; Copy-Item -Path '$(BINARY_DIR)\$(TAILMONKE_BIN).exe' -Destination \"$$env:USERPROFILE\.local\bin\""
	@echo ✓ Installation complete to %USERPROFILE%\.local\bin
	@echo Add %USERPROFILE%\.local\bin to your PATH!

# Uninstall from ~/.local/bin (Unix-like systems)
uninstall:
	@echo "Uninstalling from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(PINGMONKE_BIN) $(INSTALL_DIR)/$(TAILMONKE_BIN)
	@echo "✓ Uninstall complete"

# Uninstall from %USERPROFILE%\.local\bin (Windows)
uninstall-windows:
	@powershell -Command "if (Test-Path \"$$env:USERPROFILE\.local\bin\$(PINGMONKE_BIN).exe\") { Remove-Item \"$$env:USERPROFILE\.local\bin\$(PINGMONKE_BIN).exe\" }; if (Test-Path \"$$env:USERPROFILE\.local\bin\$(TAILMONKE_BIN).exe\") { Remove-Item \"$$env:USERPROFILE\.local\bin\$(TAILMONKE_BIN).exe\" }"
	@echo ✓ Uninstall complete

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_DIR)/$(PINGMONKE_BIN) $(BINARY_DIR)/$(TAILMONKE_BIN)
	@echo "✓ Clean complete"

# Lint code
lint:
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...
	@echo "✓ Lint complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ Tests complete"

# Development build with verbose output
dev: clean build
	@echo "✓ Development build complete"

# Update running service (Linux systemd only)
update: build
	@echo "Updating pingmonke service..."
	@systemctl --user stop pingmonke || true
	@cp $(BINARY_DIR)/$(PINGMONKE_BIN) $(INSTALL_DIR)/$(PINGMONKE_BIN)
	@cp $(BINARY_DIR)/$(TAILMONKE_BIN) $(INSTALL_DIR)/$(TAILMONKE_BIN)
	@chmod +x $(INSTALL_DIR)/$(PINGMONKE_BIN) $(INSTALL_DIR)/$(TAILMONKE_BIN)
	@systemctl --user start pingmonke
	@echo "✓ Service updated and restarted"
	@systemctl --user status pingmonke
