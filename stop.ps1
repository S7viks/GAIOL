# GAIOL Web Server Stop Script
# This script stops the GAIOL web server by finding and terminating processes on port 8080

param(
    [switch]$Help
)

if ($Help) {
    Write-Host "GAIOL Web Server Stop Script"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\stop.ps1           - Stop the server running on port 8080"
    Write-Host "  .\stop.ps1 -Help     - Show this help message"
    Write-Host ""
    exit 0
}

# Change to script directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

Write-Host "Stopping GAIOL Web Server..." -ForegroundColor Cyan
Write-Host ""

$stopped = $false

# Method 1: Find processes using port 8080
try {
    $connections = netstat -ano | Select-String ":8080"
    $pids = @()
    
    foreach ($line in $connections) {
        if ($line -match '\s+(\d+)\s*$') {
            $pid = [int]$matches[1]
            if ($pid -gt 0) {
                $pids += $pid
            }
        }
    }
    
    $uniquePids = $pids | Sort-Object -Unique
    
    if ($uniquePids.Count -gt 0) {
        Write-Host "Found processes using port 8080:" -ForegroundColor Yellow
        foreach ($pid in $uniquePids) {
            try {
                $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                if ($process) {
                    Write-Host "  Stopping process: $($process.ProcessName) (PID: $pid)" -ForegroundColor Yellow
                    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
                    $stopped = $true
                }
            } catch {
                # Process might have already exited
            }
        }
    }
} catch {
    # netstat might not be available or no processes found
}

# Method 2: Find web-server.exe processes
try {
    $webServerProcesses = Get-Process -Name "web-server" -ErrorAction SilentlyContinue
    if ($webServerProcesses) {
        Write-Host "Found web-server.exe processes:" -ForegroundColor Yellow
        foreach ($proc in $webServerProcesses) {
            Write-Host "  Stopping process: $($proc.ProcessName) (PID: $proc.Id)" -ForegroundColor Yellow
            Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
            $stopped = $true
        }
    }
} catch {
    # No web-server processes found
}

# Method 3: Find Go processes running web-server
try {
    $goProcesses = Get-Process -Name "go" -ErrorAction SilentlyContinue | Where-Object {
        $_.CommandLine -like "*web-server*" -or $_.Path -like "*web-server*"
    }
    
    # Alternative: Check command line arguments
    $allGoProcesses = Get-Process -Name "go" -ErrorAction SilentlyContinue
    foreach ($proc in $allGoProcesses) {
        try {
            $cmdLine = (Get-CimInstance Win32_Process -Filter "ProcessId = $($proc.Id)").CommandLine
            if ($cmdLine -and ($cmdLine -like "*web-server*" -or $cmdLine -like "*cmd/web-server*")) {
                Write-Host "  Stopping Go process running web-server (PID: $($proc.Id))" -ForegroundColor Yellow
                Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
                $stopped = $true
            }
        } catch {
            # Can't access command line, skip
        }
    }
} catch {
    # No matching Go processes
}

# Wait a moment for processes to terminate
if ($stopped) {
    Start-Sleep -Seconds 2
    
    # Verify server is stopped
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
        Write-Host ""
        Write-Host "Warning: Server might still be running. Health check succeeded." -ForegroundColor Yellow
    } catch {
        Write-Host ""
        Write-Host "Server stopped successfully!" -ForegroundColor Green
        Write-Host "Port 8080 is now free." -ForegroundColor Green
    }
} else {
    Write-Host ""
    Write-Host "No GAIOL server processes found running on port 8080." -ForegroundColor Yellow
    Write-Host "The server may already be stopped." -ForegroundColor Yellow
}

Write-Host ""
