Write-Host "Simple test - raw output" -ForegroundColor Cyan

$query = @{ prompt = 'Say hello' } | ConvertTo-Json

$result = Invoke-RestMethod -Uri "http://localhost:8080/api/reasoning/start" -Method POST -Body $query -ContentType "application/json" -TimeoutSec 180

Write-Host "`nRAW RESULT:" -ForegroundColor Yellow
$result | ConvertTo-Json -Depth 10 | Write-Host

Write-Host "`n`nNow checking status..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

$status = Invoke-RestMethod -Uri "http://localhost:8080/api/reasoning/status/$($result.session_id)" -TimeoutSec 30

Write-Host "`nRAW STATUS:" -ForegroundColor Yellow  
$status | ConvertTo-Json -Depth 10 | Write-Host
