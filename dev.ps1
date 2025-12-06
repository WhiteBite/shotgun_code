# Shotgun Code - Development Script
# Быстрый запуск в режиме разработки

param(
    [switch]$Verbose,
    [switch]$NoHMR
)

Write-Host "🚀 Запуск Shotgun Code в режиме разработки..." -ForegroundColor Green

# Проверяем наличие Wails
$wailsPath = $null
if (Test-Path "$env:USERPROFILE\go\bin\wails.exe") {
    $wailsPath = "$env:USERPROFILE\go\bin\wails.exe"
} elseif (Get-Command wails -ErrorAction SilentlyContinue) {
    $wailsPath = "wails"
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

# Запускаем wails dev из директории backend
Write-Host "📁 Временно переходим в backend/ и запускаем wails dev..." -ForegroundColor Yellow

# Сохраняем текущую директорию
$currentDir = Get-Location

# Set environment to reduce memory usage
$env:GOGC = "50"  # More aggressive GC

# Переходим в backend, запускаем wails dev и возвращаемся обратно
try {
    Set-Location -Path "backend"
    Write-Host "🔥 Запускаем wails dev..." -ForegroundColor Cyan
    
    if (-not $Verbose) {
        Write-Host "ℹ️  Отладочный вывод Wails фильтруется. Используйте -Verbose для полного вывода." -ForegroundColor Gray
        if ($NoHMR) {
            Write-Host "⚠️  HMR отключен для экономии памяти" -ForegroundColor Yellow
        } else {
            Write-Host "ℹ️  Используйте -NoHMR если память растёт слишком быстро." -ForegroundColor Gray
        }
        # MEMORY FIX: Use -loglevel error to disable DEB logs that accumulate in memory
        # FIX: Redirect stderr properly to avoid RemoteException spam
        $ErrorActionPreference = "Continue"
        if ($NoHMR) {
            & $wailsPath dev -loglevel error -skipbindings 2>&1 | ForEach-Object {
                $line = if ($_ -is [System.Management.Automation.ErrorRecord]) { $_.Exception.Message } else { $_.ToString() }
                if ($line -and -not ($line -match "KnownStructs:" -or $line -match "Not found: time\.Time" -or $line -match "^\s*$")) {
                    Write-Host $line
                }
            }
        } else {
            & $wailsPath dev -loglevel error 2>&1 | ForEach-Object {
                $line = if ($_ -is [System.Management.Automation.ErrorRecord]) { $_.Exception.Message } else { $_.ToString() }
                if ($line -and -not ($line -match "KnownStructs:" -or $line -match "Not found: time\.Time" -or $line -match "^\s*$")) {
                    Write-Host $line
                }
            }
        }
    } else {
        & $wailsPath dev
    }
} finally {
    # Возвращаемся в исходную директорию
    Set-Location -Path $currentDir
    # Reset GOGC
    Remove-Item Env:GOGC -ErrorAction SilentlyContinue
}
