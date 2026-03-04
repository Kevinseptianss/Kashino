# Kashino Deployment Script
$ErrorActionPreference = "Stop"

Write-Host "--- Starting Kashino Deployment ---" -ForegroundColor Cyan

# 1. Build images first to ensure code is valid before stopping anything
Write-Host "[1/4] Building images (this might take a while)..." -ForegroundColor Yellow
docker-compose build

# 2. If build succeeds, stop and restart
Write-Host "[2/4] Restarting containers with new build..." -ForegroundColor Yellow
docker-compose down
docker-compose up -d

# 3. Wait for services to initialize
Write-Host "[3/4] Waiting for services to initialize (30s)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

# 4. Verification
Write-Host "[4/4] Verifying connectivity..." -ForegroundColor Yellow

# Check Cloudflare Tunnel (Retry loop)
$maxRetries = 5
$retryCount = 0
$connected = $false

while ($retryCount -lt $maxRetries -and -not $connected) {
    $cfLogs = docker-compose logs cloudflared
    if ($cfLogs -match "Connected" -or $cfLogs -match "Registered tunnel connection") {
        $connected = $true
        Write-Host "SUCCESS: Cloudflare Tunnel is connected!" -ForegroundColor Green
    } else {
        $retryCount++
        if ($retryCount -lt $maxRetries) {
            Write-Host "Waiting for tunnel connection (Retry $retryCount/$maxRetries)..." -ForegroundColor Gray
            Start-Sleep -Seconds 10
        }
    }
}

if (-not $connected) {
    Write-Host "WARNING: Cloudflare Tunnel connection not detected in logs yet. Check 'docker-compose logs cloudflared'" -ForegroundColor Yellow
}

# Check Backend Status
try {
    $backendStatus = docker-compose ps --format json | ConvertFrom-Json | Where-Object { $_.Service -eq "backend" }
    if ($backendStatus.State -eq "running") {
        Write-Host "SUCCESS: Backend service is running." -ForegroundColor Green
    } else {
        Write-Host "ERROR: Backend service is not running." -ForegroundColor Red
    }
} catch {
    Write-Host "WARNING: Could not automatically verify backend status." -ForegroundColor Yellow
}

Write-Host "`n--- Deployment Complete ---" -ForegroundColor Cyan
Write-Host "Visit: https://kashino.my.id to view your changes." -ForegroundColor Magenta
