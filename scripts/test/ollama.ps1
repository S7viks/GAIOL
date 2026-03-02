# Test Ollama and GAIOL Integration
Write-Host "=== OLLAMA + GAIOL DEBUG TEST ===" -ForegroundColor Cyan

# Test 1: Check if Ollama is running
Write-Host "`n1. Testing Ollama availability..." -ForegroundColor Yellow
try {
    $tags = Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -TimeoutSec 5
    Write-Host "✅ Ollama is running" -ForegroundColor Green
    Write-Host "Models: $($tags.models.name -join ', ')" -ForegroundColor White
}
catch {
    Write-Host "❌ Ollama not responding: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Test Ollama generation
Write-Host "`n2. Testing Ollama generation..." -ForegroundColor Yellow
$body = @{
    model  = "llama2:latest"
    prompt = "Say hello in one sentence"
    stream = $false
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:11434/api/generate" -Method POST -Body $body -ContentType "application/json" -TimeoutSec 30
    Write-Host "✅ Ollama generated response" -ForegroundColor Green
    Write-Host "Response: $($response.response)" -ForegroundColor White
}
catch {
    Write-Host "❌ Ollama generation failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 3: Check if GAIOL server is running
Write-Host "`n3. Testing GAIOL server..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 3
    Write-Host "✅ GAIOL server is running" -ForegroundColor Green
    Write-Host "Status: $($health.status)" -ForegroundColor White
}
catch {
    Write-Host "❌ GAIOL server not responding: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Start the server first!" -ForegroundColor Yellow
    exit 1
}

# Test 4: Submit a query to GAIOL
Write-Host "`n4. Testing GAIOL query..." -ForegroundColor Yellow
$query = @{
    prompt   = "Write hello world in Python"
    strategy = "free_only"
    task     = "code"
} | ConvertTo-Json

try {
    Write-Host "Submitting query (this may take 30-60 seconds)..." -ForegroundColor Gray
    $result = Invoke-RestMethod -Uri "http://localhost:8080/api/query/smart" -Method POST -Body $query -ContentType "application/json" -TimeoutSec 120
    
    Write-Host "`n✅ Query completed!" -ForegroundColor Green
    Write-Host "Steps executed: $($result.metadata.steps_executed)" -ForegroundColor Cyan
    Write-Host "`nFinal Response:" -ForegroundColor Cyan
    Write-Host $result.response -ForegroundColor White
    
}
catch {
    Write-Host "❌ Query failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== TEST COMPLETE ===" -ForegroundColor Cyan
