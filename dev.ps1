# Shotgun Code - Development Script
# Node.js память ограничена по умолчанию из-за утечки в Vite dev server
param(
    [switch]$Verbose,
    [int]$NodeMemory = 1024  # MB, по умолчанию 1GB
)

Write-Host "🚀 Запуск Shotgun Code..." -ForegroundColor Green

# Найти wails
$wails = Get-Command wails -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source
if (-not $wails) {
    $gopath = $env:GOPATH
    if (-not $gopath) { $gopath = "$env:USERPROFILE\go" }
    $wailsPath = "$gopath\bin\wails.exe"
    if (Test-Path $wailsPath) { $wails = $wailsPath }
}

# Проверка зависимостей
$missing = @()
if (-not $wails) { $missing += "wails (go install github.com/wailsapp/wails/v2/cmd/wails@latest)" }
if (-not (Get-Command node -ErrorAction SilentlyContinue)) { $missing += "node" }
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { $missing += "go" }

if ($missing.Count -gt 0) {
    Write-Host "❌ Не установлено: $($missing -join ', ')" -ForegroundColor Red
    exit 1
}

Push-Location backend
$env:GOGC = "50"

# Ограничиваем память Node.js (Vite dev server имеет утечку)
$env:NODE_OPTIONS = "--max-old-space-size=$NodeMemory"
Write-Host "📦 Node.js heap limit: ${NodeMemory}MB" -ForegroundColor Cyan

try {
    $wailsArgs = @("dev", "-loglevel", "error")
    
    if ($Verbose) {
        & $wails dev
    } else {
        Write-Host "ℹ️  Флаги: -Verbose, -NodeMemory <MB> (default 1024)" -ForegroundColor Gray
        & $wails @wailsArgs 2>&1 | Where-Object { 
            $_ -and $_ -notmatch "KnownStructs:|Not found: time\.Time|^\s*$" 
        }
    }
} finally {
    Pop-Location
    Remove-Item Env:GOGC -ErrorAction SilentlyContinue
    Remove-Item Env:NODE_OPTIONS -ErrorAction SilentlyContinue
}
