# Shotgun Code - Development Script
# Быстрый запуск в режиме разработки

Write-Host "🚀 Запуск Shotgun Code в режиме разработки..." -ForegroundColor Green

# Проверяем наличие Wails
$wailsPath = $null
if (Get-Command wails -ErrorAction SilentlyContinue) {
    $wailsPath = "wails"
} elseif (Test-Path "$env:USERPROFILE\go\bin\wails.exe") {
    $wailsPath = "$env:USERPROFILE\go\bin\wails.exe"
} else {
    Write-Host "❌ Wails не установлен. Установите: go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Red
    exit 1
}

# Проверяем наличие Node.js
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Node.js не установлен" -ForegroundColor Red
    exit 1
}

# Проверяем наличие Go
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go не установлен" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Все зависимости установлены" -ForegroundColor Green

# Переходим в директорию backend и запускаем wails dev
Write-Host "📁 Переходим в backend/..." -ForegroundColor Yellow
cd backend

Write-Host "🔥 Запускаем wails dev..." -ForegroundColor Cyan
& $wailsPath dev
