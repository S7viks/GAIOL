# Requires: Go 1.21+
$ErrorActionPreference = "Stop"
$root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
Set-Location $root

go test ./internal/integration -tags=integration -count=1 -v
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Optional benchmark (5 iterations):"
go test ./internal/integration -tags=integration -bench=BenchmarkGoOrchestrator_SmartQuery -benchtime=5x -run=^$
