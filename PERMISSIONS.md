# Permissions Requirements

## Summary

| Platform  | Admin Required | Location                            | Notes                                       |
|-----------|----------------|------------------------------------|---------------------------------------------|
| Linux     | ❌ No          | ~/.local/bin, ~/.config/systemd/user | User-level systemd service                  |
| macOS     | ❌ No          | ~/.local/bin, ~/Library/LaunchAgents | User-level LaunchAgent                      |
| Windows   | ✅ Yes         | %LOCALAPPDATA%\Pingmonke            | Admin needed for Windows Service creation   |


## Details

### Linux (systemd user service)

**Permissions Required:** None (runs as current user)

**Directories Checked:**

- `~/.local/bin` - Must have write permission for binary installation
- `~/.config/systemd/user` - Must have write permission for service file

**What the installer does:**

- Creates directories if they don't exist
- Verifies write access before proceeding
- Returns clear error messages if permissions are insufficient

**If you get permission errors:**

```bash
# Create directories manually
mkdir -p ~/.local/bin
mkdir -p ~/.config/systemd/user

# Fix permissions if needed
chmod u+w ~/.local/bin
chmod u+w ~/.config/systemd/user
```

### macOS (launchd user agent)

**Permissions Required:** None (runs as current user)

**Directories Checked:**

- `~/.local/bin` - Must have write permission for binary installation
- `~/Library/LaunchAgents` - Must have write permission for plist file

**What the installer does:**

- Creates directories if they don't exist
- Verifies write access before proceeding
- Returns clear error messages if permissions are insufficient

**If you get permission errors:**

```bash
# Create directories manually
mkdir -p ~/.local/bin
mkdir -p ~/Library/LaunchAgents

# Fix permissions if needed
chmod u+w ~/.local/bin
chmod u+w ~/Library/LaunchAgents
```

### Windows (Windows Service)

**Permissions Required:** Administrator (for service creation only)

**Directory Checked:**

- `%LOCALAPPDATA%\Pingmonke` - User's AppData directory (always writable)

**What the installer does:**

1. Checks for Administrator privileges before proceeding
2. Creates install directory if needed
3. Verifies write access to AppData
4. Creates Windows Service with automatic startup
5. Returns helpful error messages with next steps

**If you get permission errors:**

```powershell
# Run PowerShell as Administrator
# Then execute the installer script
powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Install

# Or right-click PowerShell and select "Run as Administrator"
```

**Why Admin is needed:**

- Creating a Windows Service requires administrative privileges
- This is a Windows OS requirement, not specific to pingmonke
- Admin is only needed during installation, not for running the service


## Permission Checks in Installers

All three installers now include explicit permission checks:

### Linux & macOS

```bash
check_directory_permissions() {
    local dir=$1
    local dir_display=$2
    
    # Try to create directory if it doesn't exist
    if [ ! -d "$dir" ]; then
        if ! mkdir -p "$dir" 2>/dev/null; then
            echo -e "${RED}✗ Cannot create directory: $dir_display${NC}"
            return 1
        fi
    fi
    
    # Check if directory is writable
    if [ ! -w "$dir" ]; then
        echo -e "${RED}✗ No write permission for: $dir_display${NC}"
        return 1
    fi
    
    return 0
}
```

### Windows

```powershell
# Check for Administrator privileges
if (-not (Test-Administrator)) {
    Write-Error-Message "Service installation requires Administrator privileges"
    return $false
}

# Verify write access to AppData
try {
    $testFile = Join-Path $installPath ".permtest"
    "test" | Out-File -FilePath $testFile -ErrorAction Stop
    Remove-Item -Path $testFile -ErrorAction Stop
    Write-Success-Message "Directory is writable"
}
catch {
    Write-Error-Message "No write permission for: $installPath"
    return $false
}
```

## Troubleshooting

### Linux/macOS: "Cannot create directory"

- Check if parent directories exist and are accessible
- Verify you're not in read-only filesystem
- Try creating the directory manually: `mkdir -p ~/.local/bin`

### Linux/macOS: "No write permission"

- Fix with: `chmod u+w ~/.local/bin`
- Check parent directory permissions: `ls -la ~/`

### Windows: "Must run as Administrator"

- Right-click PowerShell icon and select "Run as Administrator"
- Re-run the installer script
- Accept any UAC (User Account Control) prompts

### Windows: "Cannot create directory"

- Ensure %LOCALAPPDATA% exists (it always does on Windows)
- Check disk space
- Verify antivirus isn't blocking directory creation


## Service Permissions After Installation

Once installed, the services run with the permissions of the user who installed them:

- **Linux:** Service runs as `$USER` (the user who ran the installer)
- **macOS:** Service runs as `$USER` (the user who ran the installer)
- **Windows:** Service runs as `Local System` (required for network monitoring)

The services have only the minimum permissions needed to:

- Send ICMP ping requests
- Read/write log files in their configured directory
- Read the configuration file

