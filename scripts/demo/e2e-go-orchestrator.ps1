# E2E: Go API with in-process orchestrator
# Prerequisites: Go API on :8080 (.\scripts\dev\start-relay.ps1 or go run ./cmd/web-server)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Set-Location $root

$apiBase = if ($env:GAIOL_API_BASE) { $env:GAIOL_API_BASE.TrimEnd("/") } else { "http://127.0.0.1:8080" }

Write-Host "Health check: $apiBase/health"
$r = Invoke-RestMethod -Uri "$apiBase/health" -Method Get
if (-not $r.orchestrator_reachable) {
    Write-Host "FAIL: orchestrator not reachable" -ForegroundColor Red
    exit 1
}
Write-Host "OK: orchestrator reachable (in-process)" -ForegroundColor Green

$body = @{
    prompt   = "Say hello in one short sentence."
    task     = "qa"
    strategy = "beam"
} | ConvertTo-Json

Write-Host "POST /api/query/smart"
$resp = Invoke-RestMethod -Uri "$apiBase/api/query/smart" -Method Post -Body $body -ContentType "application/json"
if ($resp.strategy -ne "go_orchestrator") {
    Write-Host "Unexpected strategy: $($resp.strategy)" -ForegroundColor Yellow
}
Write-Host "Answer: $($resp.result.data.Substring(0, [Math]::Min(120, $resp.result.data.Length)))..." -ForegroundColor Green
Write-Host "Trace ID: $($resp.metadata.trace_id)"
