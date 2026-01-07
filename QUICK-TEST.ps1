# ULTIMATE OLLAMA DEBUG TEST  
Write-Host "=== TESTING OLLAMA FALLBACK ===" -ForegroundColor Cyan

$query = @{
    prompt = 'Write hello world in Python'
} | ConvertTo-Json

Write-Host "`nSubmitting query..." -ForegroundColor Yellow
Write-Host "(This will take 30-60 seconds - watch the server console!)" -ForegroundColor Gray

try {
    $result = Invoke-RestMethod -Uri "http://localhost:8080/api/reasoning/start" -Method POST -Body $query -ContentType "application/json" -TimeoutSec 180
    
    Write-Host "`n*** QUERY COMPLETED! ***" -ForegroundColor Green
    Write-Host "Session ID: $($result.session_id)" -ForegroundColor Cyan
    
    # Poll for status
    Start-Sleep -Seconds 2
    $status = Invoke-RestMethod -Uri "http://localhost:8080/api/reasoning/status/$($result.session_id)" -TimeoutSec 30
    
    Write-Host "`nSteps completed: $($status.steps_completed) / $($status.total_steps)" -ForegroundColor Cyan
    Write-Host "Status: $($status.status)" -ForegroundColor White
    
    if ($status.steps -and $status.steps.Count -gt 0) {
        foreach ($step in $status.steps) {
            Write-Host "`n--- Step: $($step.title) ---" -ForegroundColor Yellow
            if ($step.top_output -and $step.top_output.response) {
                $preview = $step.top_output.response.Substring(0, [Math]::Min(150, $step.top_output.response.Length))
                Write-Host "Model: $($step.top_output.model_id)" -ForegroundColor Cyan
                Write-Host "Response: $preview..." -ForegroundColor White
            }
            else {
                Write-Host "No output yet" -ForegroundColor Gray
            }
        }
    }
    
    Write-Host "`n=== SUCCESS! ===" -ForegroundColor Green
    
}
catch {
    Write-Host "`n*** FAILED ***" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Gray
}
