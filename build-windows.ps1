#!/usr/bin/env pwsh
# Build script for Windows

$ErrorActionPreference = "Stop"

Write-Host "=== Building Shotgun Code for Windows ===" -ForegroundColor Cyan
Write-Host ""

# Find wails (same logic as dev.ps1)
$wailsPath = $null
if (Test-Path "$env:USERPROFILE\go\bin\wails.exe") {
    $wailsPath = "$env:USERPROFILE\go\bin\wails.exe"
} elseif (Get-Command wails -ErrorAction SilentlyContinue) {
    $wailsPath = "wails"
} else {
    Write-Host "❌ Wails not found. Install with: go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Red
    exit 1
}

# Ensure build directory exists
New-Item -ItemType Directory -Force -Path "build/bin" | Out-Null

Push-Location backend

try {
    # Copy frontend dist to backend/frontend/dist for embedding
    Write-Host "Preparing frontend for embedding..." -ForegroundColor Yellow
    if (Test-Path "frontend/dist") {
        Remove-Item "frontend/dist" -Recurse -Force -ErrorAction SilentlyContinue
    }
    Copy-Item "../frontend/dist" "frontend/dist" -Recurse -Force
    
    Write-Host "Building Windows executable..." -ForegroundColor Green
    & $wailsPath build -clean -skipbindings
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✓ Build successful!" -ForegroundColor Green
        
        # Move from backend/build/bin to root build/bin
        if (Test-Path "build/bin/ShotgunWB.exe") {
            Write-Host "Moving executable to build/bin/..." -ForegroundColor Yellow
            Copy-Item "build/bin/ShotgunWB.exe" "../build/bin/ShotgunWB.exe" -Force
            Write-Host ""
            
            $exePath = Get-Item "../build/bin/ShotgunWB.exe"
            $size = [math]::Round($exePath.Length / 1MB, 2)
            Write-Host "Output: $($exePath.FullName)" -ForegroundColor White
            Write-Host "Size: ${size}MB" -ForegroundColor White
        } else {
            Write-Host "⚠️  Warning: Could not find built executable" -ForegroundColor Yellow
        }
    } else {
        Write-Host ""
        Write-Host "✗ Build failed!" -ForegroundColor Red
        exit 1
    }
} finally {
    Pop-Location
}
