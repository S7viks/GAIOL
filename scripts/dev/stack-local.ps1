# Local stack helper — use start-relay.ps1 for one-command startup.
$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Write-Host "GAIOL local stack (Go API + in-process orchestrator)"
Write-Host "  .\scripts\dev\start-relay.ps1"
Write-Host "  .\scripts\dev\start-relay.ps1 -Dashboard   # + Vite on :5173"
Write-Host ""
Write-Host "Manual: docs/LOCAL-DEV-STACK.md"
Write-Host "Repo: $root"
