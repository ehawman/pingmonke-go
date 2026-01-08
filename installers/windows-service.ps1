# Windows Service Installer for Pingmonke
# Installs pingmonke as a Windows Service with automatic startup
# Also installs tailmonke for log viewing
# Uses user-local AppData directory (no admin required for binaries)
# 
# Usage:
#   powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Install
#   powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Uninstall
#   powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Status
#
# Note: First run may require Administrator privileges to create the service

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("Install", "Uninstall", "Status", "Start", "Stop", "Restart")]
    [string]$Action = "Install"
)

# Configuration
$serviceName = "Pingmonke"
$serviceDisplayName = "Pingmonke Network Monitor"
$serviceDescription = "Monitors network connectivity via ICMP ping with event detection"
$localAppData = $env:LOCALAPPDATA
$installPath = Join-Path $localAppData "Pingmonke"
$exeName = "pingmonke.exe"
$exePath = Join-Path $installPath $exeName
$tailExeName = "tailmonke.exe"
$tailExePath = Join-Path $installPath $tailExeName
$configPath = Join-Path $installPath "config.yaml"

# Colors for output
$errorColor = "Red"
$successColor = "Green"
$infoColor = "Cyan"

function Write-Error-Message {
    param([string]$message)
    Write-Host "ERROR: $message" -ForegroundColor $errorColor
}

function Write-Success-Message {
    param([string]$message)
    Write-Host "✓ $message" -ForegroundColor $successColor
}

function Write-Info-Message {
    param([string]$message)
    Write-Host "ℹ $message" -ForegroundColor $infoColor
}

# Check for Administrator privileges
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-Administrator)) {
    Write-Error-Message "This script must be run as Administrator"
    Write-Info-Message "Please re-run this script with Administrator privileges"
    exit 1
}

# Service status check
function Get-ServiceStatus {
    $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
    if ($service) {
        return $service.Status
    }
    return "NotFound"
}

# Install service
function Install-Service {
    # Check admin for service creation
    if (-not (Test-Administrator)) {
        Write-Error-Message "Service installation requires Administrator privileges"
        Write-Info-Message "Please run: powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Install"
        Write-Info-Message "And accept the UAC prompt"
        return $false
    }

    $status = Get-ServiceStatus
    
    if ($status -ne "NotFound") {
        Write-Error-Message "Service '$serviceName' already exists"
        Write-Info-Message "Current status: $status"
        Write-Info-Message "To reinstall, run: powershell -ExecutionPolicy Bypass -File windows-service.ps1 -Uninstall"
        return $false
    }

    Write-Info-Message "Installation Details:"
    Write-Host "  • Service Name: $serviceName"
    Write-Host "  • Install Path: $installPath"
    Write-Host "  • Startup Type: Automatic"
    Write-Host "  • Service Type: Windows Service"
    Write-Host "  • Permissions Required: Administrator (for service creation only)"
    Write-Host ""

    Write-Info-Message "Checking permissions..."
    
    # Try to create install directory to verify write access
    if (-not (Test-Path $installPath)) {
        try {
            New-Item -ItemType Directory -Path $installPath -Force -ErrorAction Stop | Out-Null
            Write-Success-Message "Created directory: $installPath"
        }
        catch {
            Write-Error-Message "Cannot create directory: $installPath"
            Write-Info-Message "Ensure %LOCALAPPDATA% is accessible"
            Write-Info-Message "Current AppData: $installPath"
            return $false
        }
    } else {
        # Directory exists, check if writable
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
    }

    Write-Host ""

    if (-not (Test-Path $exePath)) {
        Write-Error-Message "Executable not found at: $exePath"
        Write-Info-Message "Please build pingmonke first"
        Write-Info-Message "Build command: go build -o \"$exePath\" .\cmd\pingmonke"
        return $false
    }

    # Check for tailmonke and warn if missing
    if (-not (Test-Path $tailExePath)) {
        Write-Info-Message "tailmonke not found at: $tailExePath"
        Write-Info-Message "To install tailmonke: go build -o \"$tailExePath\" .\cmd\tailmonke"
    } else {
        Write-Success-Message "tailmonke found and ready"
    }

    Write-Host ""
    $confirm = Read-Host "Proceed with installation? [Y/n]"
    if ($confirm -eq 'n' -or $confirm -eq 'N') {
        Write-Info-Message "Installation cancelled."
        return $false
    }

    Write-Host ""
    Write-Info-Message "Installing..."
    
    try {
        $serviceParams = @{
            Name = $serviceName
            BinaryPathName = $exePath
            DisplayName = $serviceDisplayName
            Description = $serviceDescription
            StartupType = "Automatic"
        }
        
        New-Service @serviceParams | Out-Null
        Write-Success-Message "Service created successfully"
        
        # Configure service to restart on failure
        sc.exe failure $serviceName reset= 60 actions= restart/5000
        Write-Success-Message "Configured automatic restart on failure"
        
        # Start the service
        Start-Service -Name $serviceName
        Write-Success-Message "Service started successfully"
        
        # Verify it's running
        Start-Sleep -Seconds 2
        $service = Get-Service -Name $serviceName
        if ($service.Status -eq "Running") {
            Write-Success-Message "Service is running"
        } else {
            Write-Error-Message "Service failed to start. Status: $($service.Status)"
            return $false
        }
        
        return $true
    }
    catch {
        Write-Error-Message "Failed to install service: $_"
        return $false
    }
}

# Uninstall service
function Uninstall-Service {
    $status = Get-ServiceStatus
    
    if ($status -eq "NotFound") {
        Write-Error-Message "Service '$serviceName' not found"
        return $false
    }

    Write-Info-Message "Uninstalling service '$serviceName'..."
    
    try {
        # Stop the service if it's running
        $service = Get-Service -Name $serviceName
        if ($service.Status -eq "Running") {
            Stop-Service -Name $serviceName -Force
            Write-Info-Message "Service stopped"
        }
        
        # Remove the service
        Remove-Service -Name $serviceName -Force
        Write-Success-Message "Service uninstalled successfully"
        return $true
    }
    catch {
        Write-Error-Message "Failed to uninstall service: $_"
        return $false
    }
}

# Show service status
function Show-ServiceStatus {
    $status = Get-ServiceStatus
    
    if ($status -eq "NotFound") {
        Write-Error-Message "Service '$serviceName' not found"
        return $false
    }

    $service = Get-Service -Name $serviceName
    Write-Info-Message "Service Information:"
    Write-Host "  Name: $($service.Name)"
    Write-Host "  Display Name: $($service.DisplayName)"
    Write-Host "  Status: $($service.Status)"
    Write-Host "  Start Type: $($service.StartType)"
    
    return $true
}

# Start service
function Start-ServiceNow {
    $status = Get-ServiceStatus
    
    if ($status -eq "NotFound") {
        Write-Error-Message "Service '$serviceName' not found"
        return $false
    }

    try {
        Start-Service -Name $serviceName
        Write-Success-Message "Service started"
        return $true
    }
    catch {
        Write-Error-Message "Failed to start service: $_"
        return $false
    }
}

# Stop service
function Stop-ServiceNow {
    $status = Get-ServiceStatus
    
    if ($status -eq "NotFound") {
        Write-Error-Message "Service '$serviceName' not found"
        return $false
    }

    try {
        Stop-Service -Name $serviceName
        Write-Success-Message "Service stopped"
        return $true
    }
    catch {
        Write-Error-Message "Failed to stop service: $_"
        return $false
    }
}

# Restart service
function Restart-ServiceNow {
    $status = Get-ServiceStatus
    
    if ($status -eq "NotFound") {
        Write-Error-Message "Service '$serviceName' not found"
        return $false
    }

    try {
        Restart-Service -Name $serviceName
        Write-Success-Message "Service restarted"
        return $true
    }
    catch {
        Write-Error-Message "Failed to restart service: $_"
        return $false
    }
}

# Main script logic
Write-Host ""
Write-Host "Pingmonke Windows Service Installer" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan
Write-Host "Install directory: $installPath" -ForegroundColor Gray
Write-Host ""

switch ($Action) {
    "Install" {
        Install-Service
    }
    "Uninstall" {
        Uninstall-Service
    }
    "Status" {
        Show-ServiceStatus
    }
    "Start" {
        Start-ServiceNow
    }
    "Stop" {
        Stop-ServiceNow
    }
    "Restart" {
        Restart-ServiceNow
    }
}

Write-Host ""