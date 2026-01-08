#!/bin/bash
# @description Install pingmonke as a launchd user agent
# Also installs tailmonke for log viewing
# Installs to ~/.local/bin and uses user-level launchd service
# No sudo required

set -e

BIN_PATH="$HOME/.local/bin/pingmonke"
TAIL_BIN_PATH="$HOME/.local/bin/tailmonke"
PLIST="$HOME/Library/LaunchAgents/com.pingmonke.service.plist"

# Colors
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}╔═══════════════════════════════════════╗${NC}"
echo -e "${BOLD}${BLUE}║  Pingmonke Service Installer (macOS)  ║${NC}"
echo -e "${BOLD}${BLUE}╚═══════════════════════════════════════╝${NC}"
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
echo -e "  • Service Name: ${GREEN}com.pingmonke.service${NC}"
echo -e "  • Install Path: ${GREEN}$HOME/.local/bin${NC}"
echo -e "  • Plist File: ${GREEN}$PLIST${NC}"
echo -e "  • Service Type: ${GREEN}User-level LaunchAgent${NC}"
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

# Create ~/.local/bin if it doesn't exist
echo "Creating directories..."
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/Library/LaunchAgents"

# Ensure binary is executable
chmod +x "$BIN_PATH"
echo -e "  ${GREEN}✓${NC} $HOME/.local/bin/pingmonke"

if [ -f "$TAIL_BIN_PATH" ]; then
    chmod +x "$TAIL_BIN_PATH"
    echo -e "  ${GREEN}✓${NC} $HOME/.local/bin/tailmonke"
fi

echo "Creating launchd plist file..."

cat > "$PLIST" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.pingmonke.service</string>
    <key>Program</key>
    <string>$HOME/.local/bin/pingmonke</string>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$HOME/Library/Logs/pingmonke.log</string>
    <key>StandardErrorPath</key>
    <string>$HOME/Library/Logs/pingmonke-error.log</string>
</dict>
</plist>
EOF

# Substitute HOME in the plist file
sed -i '' "s|\$HOME|$HOME|g" "$PLIST"
echo -e "  ${GREEN}✓${NC} $PLIST"

echo "Loading launchd service..."
launchctl load "$PLIST"
echo -e "  ${GREEN}✓${NC} Service loaded and started"

echo ""
echo -e "${BOLD}${GREEN}✓ Installation complete!${NC}"
echo ""
echo -e "${BOLD}Service Management:${NC}"
echo -e "  View status:   ${BLUE}launchctl list | grep pingmonke${NC}"
echo -e "  View logs:     ${BLUE}tail -f ~/Library/Logs/pingmonke.log${NC}"
echo -e "  Unload service: ${BLUE}launchctl unload $PLIST${NC}"
echo -e "  Load service:  ${BLUE}launchctl load $PLIST${NC}"
echo ""
echo -e "${BOLD}Next Steps:${NC}"
echo -e "  • Verify logs are being created: ${BLUE}tail -f ~/ping-logs/*.csv${NC}"
echo -e "  • View logs in TUI: ${BLUE}tailmonke${NC}"
echo -e "  • Edit config: ${BLUE}nano ~/ping-logs/config.yaml${NC}"
echo ""