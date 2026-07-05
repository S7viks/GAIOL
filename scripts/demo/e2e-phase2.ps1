# Phase 2 E2E checks (no-auth / local stack).
# Prerequisites: Go API on :8080 (.\scripts\dev\start-relay.ps1 or go run ./cmd/web-server).
# Run from repo root: .\scripts\demo\e2e-phase2.ps1

param(
    [string]$GoBase = "http://127.0.0.1:8080",
    [switch]$Help
)

if ($Help) {
    Write-Host @"
Phase 2 automated checks (local no-auth).

  .\scripts\demo\e2e-phase2.ps1

Requires Go API at $GoBase (in-process orchestrator).
"@
    exit 0
}

$ErrorActionPreference = "Stop"
$failed = 0

function Assert-True($cond, [string]$msg) {
    if (-not $cond) {
        Write-Host "FAIL: $msg" -ForegroundColor Red
        $script:failed++
    } else {
        Write-Host "OK:   $msg" -ForegroundColor Green
    }
}

Write-Host "Phase 2 E2E (automated)" -ForegroundColor Cyan
Write-Host ""

try {
    $health = Invoke-RestMethod -Uri "$GoBase/health" -Method Get -TimeoutSec 10
    Assert-True ($health.inference_mode -eq "orchestrator_only") "inference_mode is orchestrator_only"
    Assert-True ($health.orchestrator_configured -eq $true) "orchestrator_configured is true"
    Assert-True ($health.orchestrator_reachable -eq $true) "orchestrator_reachable is true"
} catch {
    Assert-True $false "GET /health failed: $($_.Exception.Message)"
}

try {
    $setup = Invoke-RestMethod -Uri "$GoBase/api/setup/status" -Method Get -TimeoutSec 10
    Assert-True ($setup.inference_mode -eq "orchestrator_only") "setup status inference_mode"
    Assert-True ($null -ne $setup.setup_complete) "setup_complete field present"
} catch {
    Assert-True $false "GET /api/setup/status failed: $($_.Exception.Message)"
}

$traceId = $null
try {
    $body = @{
        prompt     = "Phase 2 E2E ping"
        task       = "qa"
        strategy   = "beam"
        max_tokens = 64
    } | ConvertTo-Json
    $chat = Invoke-RestMethod -Uri "$GoBase/api/query/smart" -Method Post -Body $body -ContentType "application/json" -TimeoutSec 120
    $traceId = $chat.metadata.trace_id
    Assert-True ([bool]$traceId) "POST /api/query/smart returned metadata.trace_id"
    Assert-True ($chat.strategy -eq "go_orchestrator") "strategy is go_orchestrator"
} catch {
    Assert-True $false "POST /api/query/smart failed: $($_.Exception.Message)"
}

$appPath = Join-Path (Split-Path (Split-Path $PSScriptRoot -Parent) -Parent) "dashboard\src\App.tsx"
if (Test-Path $appPath) {
    $appText = Get-Content $appPath -Raw
    Assert-True ($appText -notmatch 'path="reasoning"') "No /reasoning route in App.tsx"
} else {
    Write-Host "SKIP: App.tsx not found" -ForegroundColor Yellow
}

Write-Host ""
if ($failed -gt 0) {
    Write-Host "$failed check(s) failed." -ForegroundColor Red
    exit 1
}
Write-Host "All Phase 2 automated checks passed." -ForegroundColor Green
if ($traceId) { Write-Host "Sample trace_id: $traceId" -ForegroundColor Gray }
exit 0
