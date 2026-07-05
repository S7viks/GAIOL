# Relay / GAIOL - one command, one inference engine (Go in-process orchestrator).
# Run from repo root: .\scripts\dev\start-relay.ps1
# Optional: -Dashboard also starts Vite on :5173

param(
    [switch]$Dashboard,
    [switch]$NoBuild,
    [switch]$Help
)

if ($Help) {
    Write-Host "Relay local stack - Go API with in-process orchestrator."
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\scripts\dev\start-relay.ps1              Go API only"
    Write-Host "  .\scripts\dev\start-relay.ps1 -Dashboard   Also start Vite dashboard"
    Write-Host "  .\scripts\dev\start-relay.ps1 -NoBuild     Skip Go rebuild"
    Write-Host ""
    Write-Host "Docs: docs/LOCAL-DEV-STACK.md"
    exit 0
}

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Set-Location $root

if (-not (Test-Path "go.mod")) {
    Write-Host "Run from the GAIOL repo root (go.mod not found)." -ForegroundColor Red
    exit 1
}

$goPort = "8080"
if ($env:PORT -and $env:PORT.Trim()) {
    $goPort = $env:PORT.Trim()
}

if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            if ($key -and $null -ne $value) {
                [Environment]::SetEnvironmentVariable($key, $value, "Process")
            }
        }
    }
    if ($env:PORT -and $env:PORT.Trim()) {
        $goPort = $env:PORT.Trim()
    }
}

$script:dashProc = $null

function Stop-ChildStack {
    if ($script:dashProc -and -not $script:dashProc.HasExited) {
        Write-Host "Stopping dashboard..." -ForegroundColor Gray
        Stop-Process -Id $script:dashProc.Id -Force -ErrorAction SilentlyContinue
    }
}

try {
    Write-Host ""
    Write-Host "Relay - Go API with in-process orchestrator" -ForegroundColor Cyan
    Write-Host "  Go API: http://127.0.0.1:$goPort" -ForegroundColor Gray
    Write-Host ""

    if ($Dashboard) {
        $dashDir = Join-Path $root "dashboard"
        if (-not (Test-Path (Join-Path $dashDir "node_modules"))) {
            Write-Host "Installing dashboard dependencies..." -ForegroundColor Yellow
            Push-Location $dashDir
            npm install --no-fund --no-audit
            if ($LASTEXITCODE -ne 0) { exit 1 }
            Pop-Location
        }
        Write-Host "Starting Vite dashboard on http://localhost:5173 ..." -ForegroundColor Green
        $script:dashProc = Start-Process -FilePath "npm.cmd" -ArgumentList "run", "dev" `
            -WorkingDirectory $dashDir -PassThru -WindowStyle Minimized
    }

    $needsBuild = $false
    if (-not $NoBuild) {
        if (-not (Test-Path "web-server.exe")) {
            $needsBuild = $true
        } else {
            $exeTime = (Get-Item "web-server.exe").LastWriteTimeUtc
            foreach ($dir in @("cmd", "internal")) {
                if (-not (Test-Path $dir)) { continue }
                Get-ChildItem -Path $dir -Recurse -Filter "*.go" -File -ErrorAction SilentlyContinue | ForEach-Object {
                    if ($_.LastWriteTimeUtc -gt $exeTime) { $needsBuild = $true }
                }
            }
        }
    }
    if ($needsBuild) {
        Write-Host "Building Go API..." -ForegroundColor Cyan
        go build -o web-server.exe ./cmd/web-server/
        if ($LASTEXITCODE -ne 0) { exit 1 }
    }

    Write-Host ""
    Write-Host "Starting Go API..." -ForegroundColor Cyan
    if ($Dashboard) {
        Write-Host "Dashboard UI: http://localhost:5173" -ForegroundColor Green
    }
    Write-Host "Health: http://localhost:$goPort/health" -ForegroundColor Green
    Write-Host "Press Ctrl+C to stop" -ForegroundColor Yellow
    Write-Host ""

    & ".\web-server.exe"
} finally {
    Stop-ChildStack
}
