# Pingmonke Service Installers

Installation scripts for running pingmonke as a background service on different operating systems. Also includes tailmonke for convenient log viewing.

## Quick Start

All platforms: First build both binaries from the **repo root**, then run the installer for your OS.

```bash
# Navigate to the repo root
cd ~/Programming/pingmonke-go

# Build both tools to user-local directory
go build -o ~/.local/bin/pingmonke ./cmd/pingmonke
go build -o ~/.local/bin/tailmonke ./cmd/tailmonke

# Or use the Makefile (easier)
make install

# Then run your platform's installer (Linux example)
bash ./installers/linux-systemd.sh
```

## Platform Support

### Windows (PowerShell)

**File:** `windows-service.ps1`

Install pingmonke as a Windows Service with automatic startup and failure recovery. Uses user-local AppData directory.

#### Requirements

- Windows 7 or later
- Administrator privileges (for service creation only)
- PowerShell 3.0 or later
- `pingmonke.exe` in the script's directory or specify via `-InstallPath`

#### Installation

1. Build both binaries to your preferred location:

```powershell
# Option A: Install to user AppData (recommended)
mkdir -p "$env:LOCALAPPDATA\Pingmonke"
go build -o "$env:LOCALAPPDATA\Pingmonke\pingmonke.exe" ./cmd/pingmonke
go build -o "$env:LOCALAPPDATA\Pingmonke\tailmonke.exe" ./cmd/tailmonke

# Option B: Add to PATH via user .local\bin (Unix-compatible)
mkdir -p "$env:USERPROFILE\.local\bin"
go build -o "$env:USERPROFILE\.local\bin\pingmonke.exe" ./cmd/pingmonke
go build -o "$env:USERPROFILE\.local\bin\tailmonke.exe" ./cmd/tailmonke
```

1. Run installer:

```powershell
powershell -ExecutionPolicy Bypass -File installers\windows-service.ps1 -Install
```

The installer will verify both binaries exist and create the Windows service.

#### Usage

```powershell
# Install service (prompts for admin if needed)
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Install

# View service status
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Status

# Start the service
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Start

# Stop the service
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Stop

# Restart the service
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Restart

# Uninstall service
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Uninstall
```

#### Install Location

- **Binary:** `%LOCALAPPDATA%\Pingmonke\pingmonke.exe`
- **Config:** `%LOCALAPPDATA%\Pingmonke\config.yaml` (optional)
- **Example:** `C:\Users\YourName\AppData\Local\Pingmonke\`

#### Features

- User-local installation (no system-wide changes)
- Automatic startup on boot
- Automatic restart on failure (5-second delay)
- Color-coded output for success/error/info messages
- Service status verification after installation
- Graceful uninstall with proper cleanup

#### Logs

Service logs are available in Windows Event Viewer:

- Open Event Viewer
- Navigate to Windows Logs > Application
- Filter by service name "Pingmonke"

---

### Linux (systemd)

**File:** `linux-systemd.sh`

Install pingmonke as a systemd user service for automatic startup and management. Uses XDG Base Directory spec.

#### Linux Requirements

- Linux with systemd (most modern distributions)
- No sudo required (uses user-level systemd)
- `pingmonke` binary available or built beforehand

#### Linux Installation

1. Build both binaries to `~/.local/bin`:

```bash
go build -o ~/.local/bin/pingmonke ./cmd/pingmonke
go build -o ~/.local/bin/tailmonke ./cmd/tailmonke
```

1. Run installer:

```bash
bash ./installers/linux-systemd.sh
```

The installer will verify both binaries exist and enable the systemd service.

#### Linux Usage

```bash
# View service status
systemctl --user status pingmonke

# Start the service
systemctl --user start pingmonke

# Stop the service
systemctl --user stop pingmonke

# Restart the service
systemctl --user restart pingmonke

# View service logs
systemctl --user status pingmonke
journalctl --user -u pingmonke -f

# Uninstall
systemctl --user disable pingmonke
systemctl --user stop pingmonke
rm ~/.config/systemd/user/pingmonke.service
systemctl --user daemon-reload
```

#### Linux Install Locations

- **Binary:** `~/.local/bin/pingmonke`
- **Config:** `~/.config/pingmonke/config.yaml` or `~/ping-logs/config.yaml`
- **Service:** `~/.config/systemd/user/pingmonke.service`

#### Linux Features

- User-local installation (no sudo required)
- Automatic startup at user login
- Proper logging to journalctl
- Service restart on failure (10-second delay)
- Standard systemd user service

#### Linux Logs

View logs with journalctl:

```bash
# Last 50 lines
journalctl --user -u pingmonke -n 50

# Follow new logs
journalctl --user -u pingmonke -f

# Since last boot
journalctl --user -b
```

---

### macOS (launchd)

**File:** `macos-launchd.sh`

Install pingmonke as a launchd user agent for automatic startup and management. Uses user-level LaunchAgent instead of system-wide LaunchDaemon.

#### macOS Requirements

- macOS 10.5 or later
- No sudo required (uses user-level LaunchAgent)
- `pingmonke` binary available or built beforehand

#### macOS Installation

1. Build both binaries to `~/.local/bin`:

```bash
go build -o ~/.local/bin/pingmonke ./cmd/pingmonke
go build -o ~/.local/bin/tailmonke ./cmd/tailmonke
```

1. Run installer:

```bash
bash ./installers/macos-launchd.sh
```

The installer will verify both binaries exist and enable the launchd service.

#### macOS Usage

```bash
# View service status
launchctl list | grep pingmonke

# View service logs
log stream --predicate 'process == "pingmonke"'

# Stop the service (unload)
launchctl unload ~/Library/LaunchAgents/com.pingmonke.service.plist

# Start the service (load)
launchctl load ~/Library/LaunchAgents/com.pingmonke.service.plist

# Uninstall
launchctl unload ~/Library/LaunchAgents/com.pingmonke.service.plist
rm ~/Library/LaunchAgents/com.pingmonke.service.plist
```

#### macOS Install Locations

- **Binary:** `~/.local/bin/pingmonke`
- **Config:** `~/.config/pingmonke/config.yaml` or `~/ping-logs/config.yaml`
- **Plist:** `~/Library/LaunchAgents/com.pingmonke.service.plist`
- **Logs:** `~/Library/Logs/pingmonke.log` and `~/Library/Logs/pingmonke-error.log`

#### macOS Features

- User-local installation (no sudo required)
- Automatic startup at user login (not system boot)
- Proper logging to ~/Library/Logs
- Standard launchd user agent
- Loads at user login time

#### macOS Logs

View logs with the unified log system:

```bash
# Stream pingmonke logs
log stream --predicate 'process == "pingmonke"'

# View log files directly
tail -f ~/Library/Logs/pingmonke.log
tail -f ~/Library/Logs/pingmonke-error.log
```

---

## Configuration

All platforms use the same `config.yaml` file. Place it in your user-local configuration directory:

- **Windows:** `%LOCALAPPDATA%\Pingmonke\config.yaml`
  - Example: `C:\Users\YourName\AppData\Local\Pingmonke\config.yaml`
- **Linux:** `~/.config/pingmonke/config.yaml` or `~/ping-logs/config.yaml`
- **macOS:** `~/.config/pingmonke/config.yaml` or `~/ping-logs/config.yaml`

See the main README for configuration options.

---

## Installation Comparison

| Feature        | Windows                         | Linux                 | macOS                        |
|----------------|---------------------------------|-----------------------|------------------------------|
| Requires Sudo  | No (except service creation)    | No                    | No                           |
| Install Path   | `%LOCALAPPDATA%\Pingmonke`      | `~/.local/bin`        | `~/.local/bin`               |
| Service Type   | Windows Service                 | systemd user service  | launchd user agent           |
| Auto-start     | On boot                         | At user login         | At user login                |
| Logs           | Event Viewer                    | journalctl            | ~/Library/Logs & log stream  |


---

## Troubleshooting

### Service fails to start

1. **Windows:**
   - Check Event Viewer: Applications > Pingmonke service
   - Verify binary exists: `%LOCALAPPDATA%\Pingmonke\pingmonke.exe`
   - Run: `powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Status`

2. **Linux:**

   ```bash
   systemctl --user status pingmonke
   journalctl --user -u pingmonke -n 20
   ```

3. **macOS:**

   ```bash
   launchctl list | grep pingmonke
   log stream --predicate 'process == "pingmonke"'
   ```

### Service stops unexpectedly

Check the logs for error messages. Common issues:

- Configuration file not found or invalid YAML syntax
- Network connectivity issues preventing ping
- Permission problems with log directory (`~/ping-logs`)
- Insufficient disk space

### Reinstall service

**Windows:**

```powershell
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Uninstall
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Install
```

**Linux:**

```bash
systemctl --user disable pingmonke
systemctl --user stop pingmonke
rm ~/.config/systemd/user/pingmonke.service
systemctl --user daemon-reload
bash ./linux-systemd.sh
```

**macOS:**

```bash
launchctl unload ~/Library/LaunchAgents/com.pingmonke.service.plist
rm ~/Library/LaunchAgents/com.pingmonke.service.plist
bash ./macos-launchd.sh
```

### Binary not found errors

Ensure the binary is in the correct location for your platform:

**Windows:**

```powershell
# Check if file exists
Test-Path "$env:LOCALAPPDATA\Pingmonke\pingmonke.exe"

# Build to correct location
go build -o "$env:LOCALAPPDATA\Pingmonke\pingmonke.exe" ./cmd/pingmonke
```

**Linux/macOS:**

```bash
# Check if file exists
ls -la ~/.local/bin/pingmonke

# Build to correct location
go build -o ~/.local/bin/pingmonke ./cmd/pingmonke
```

---

## Cross-Platform Notes

- All services run with the user that installed them
- Ensure the log directory (`~/ping-logs` by default) is writable
- Configuration file is loaded from the same directory as the executable or config path
- Services automatically restart on failure
- Use `--debug-rollover` flag in `config.yaml` to enable verbose logging

---

## License

These installers are part of the Pingmonke project.
