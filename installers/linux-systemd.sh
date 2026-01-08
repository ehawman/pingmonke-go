#!/bin/bash
# @description Install pingmonke as a systemd user service
# Also installs tailmonke for log viewing
# Installs to ~/.local/bin and uses user-level systemd service
# No sudo required

set -e

SERVICE_NAME="pingmonke"
BIN_PATH="$HOME/.local/bin/pingmonke"
TAIL_BIN_PATH="$HOME/.local/bin/tailmonke"
SERVICE_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SERVICE_DIR/$SERVICE_NAME.service"

# Colors
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${BLUE}║   Pingmonke Service Installer (Linux)  ║${NC}"
echo -e "${BOLD}${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Helper function for prompts
prompt_confirm() {
    local prompt="$1"
    local response
    while true; do
        read -p "$(echo -e "${BOLD}${prompt}${NC}") [Y/n] " -r response
        case "$response" in
            [yY][eE][sS]|[yY]|'')
                return 0
                ;;
            [nN][oO]|[nN])
                return 1
                ;;
            *)
                echo "Please answer yes or no."
                ;;
        esac
    done
}

echo -e "${BOLD}Installation Details:${NC}"
echo -e "  • Service Name: ${GREEN}$SERVICE_NAME${NC}"
echo -e "  • Install Path: ${GREEN}$HOME/.local/bin${NC}"
echo -e "  • Service File: ${GREEN}$SERVICE_FILE${NC}"
echo -e "  • Service Type: ${GREEN}User-level systemd${NC}"
echo ""

# Check if binaries exist
echo -e "${BOLD}Checking for binaries...${NC}"
if [ ! -f "$BIN_PATH" ]; then
    echo -e "${RED}✗ pingmonke not found at $BIN_PATH${NC}"
    echo -e "${YELLOW}Please build pingmonke first:${NC}"
    echo -e "  ${BLUE}go build -o ~/.local/bin/pingmonke ./cmd/pingmonke${NC}"
    exit 1
else
    echo -e "${GREEN}✓ pingmonke found${NC}"
fi

if [ ! -f "$TAIL_BIN_PATH" ]; then
    echo -e "${YELLOW}⚠ tailmonke not found at $TAIL_BIN_PATH${NC}"
    echo -e "  (optional but recommended)"
    echo -e "  ${BLUE}go build -o ~/.local/bin/tailmonke ./cmd/tailmonke${NC}"
else
    echo -e "${GREEN}✓ tailmonke found${NC}"
fi

echo ""

# Confirm installation
if ! prompt_confirm "${YELLOW}Proceed with installation?${NC}"; then
    echo -e "${YELLOW}Installation cancelled.${NC}"
    exit 0
fi

echo ""
echo -e "${BOLD}Installing...${NC}"

# Create directories if they don't exist
echo "Creating directories..."
mkdir -p "$HOME/.local/bin"
mkdir -p "$SERVICE_DIR"

# Ensure binary is executable
chmod +x "$BIN_PATH"
echo -e "  ${GREEN}✓${NC} $HOME/.local/bin/pingmonke"

if [ -f "$TAIL_BIN_PATH" ]; then
    chmod +x "$TAIL_BIN_PATH"
    echo -e "  ${GREEN}✓${NC} $HOME/.local/bin/tailmonke"
fi

# Create systemd service file
echo "Creating systemd service file..."

cat > "$SERVICE_FILE" << 'EOF'
[Unit]
Description=Pingmonke Network Monitor
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$HOME/.local/bin/pingmonke
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
EOF

# Expand HOME variable in the file
sed -i "s|\$HOME|$HOME|g" "$SERVICE_FILE"
echo -e "  ${GREEN}✓${NC} $SERVICE_FILE"

# Enable and start service
echo "Configuring systemd..."
systemctl --user daemon-reload
echo -e "  ${GREEN}✓${NC} Daemon reloaded"

systemctl --user enable $SERVICE_NAME.service
echo -e "  ${GREEN}✓${NC} Service enabled for auto-start"

systemctl --user start $SERVICE_NAME.service
echo -e "  ${GREEN}✓${NC} Service started"

echo ""
echo -e "${BOLD}${GREEN}✓ Installation complete!${NC}"
echo ""
echo -e "${BOLD}Service Management:${NC}"
echo -e "  View status:   ${BLUE}systemctl --user status pingmonke${NC}"
echo -e "  View logs:     ${BLUE}journalctl --user -u pingmonke -f${NC}"
echo -e "  Stop service:  ${BLUE}systemctl --user stop pingmonke${NC}"
echo -e "  Start service: ${BLUE}systemctl --user start pingmonke${NC}"
echo ""
echo -e "${BOLD}Next Steps:${NC}"
echo -e "  • Verify logs are being created: ${BLUE}tail -f ~/ping-logs/*.csv${NC}"
echo -e "  • View logs in TUI: ${BLUE}tailmonke${NC}"
echo -e "  • Edit config: ${BLUE}nano ~/ping-logs/config.yaml${NC}"
echo ""