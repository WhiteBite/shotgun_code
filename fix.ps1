# PowerShell скрипт для реструктуризации UI
Write-Host "Starting UI refactoring to single-pane view..." -ForegroundColor Yellow

$baseDir = "frontend\src"

# --- 1. Удаление старых файлов и директорий ---
Write-Host "Step 1: Deleting old workflow components..."
$dirsToDelete = @(
    "$baseDir\components\workflow"
)
$filesToDelete = @(
    "$baseDir\views\WorkflowView.vue",
    "$baseDir\stores\workflow.store.ts",
    "$baseDir\components\shared\FilePreview.vue", # Будет заменено
    "$baseDir\components\shared\ProgressOverlay.vue" # Будет заменено
)

foreach ($dir in $dirsToDelete) {
    if (Test-Path $dir) {
        Write-Host "Deleting directory: $dir" -ForegroundColor Magenta
        Remove-Item -Recurse -Force $dir
    }
}
foreach ($file in $filesToDelete) {
    if (Test-Path $file) {
        Write-Host "Deleting file: $file" -ForegroundColor Magenta
        Remove-Item -Force $file
    }
}

# --- 2. Создание новых директорий и файлов ---
Write-Host "Step 2: Creating new UI structure..."
$dirsToCreate = @(
    "$baseDir\components\panels",
    "$baseDir\components\workspace" # Пересоздаем, если была удалена
)
$filesToCreate = @(
    "$baseDir\views\ProjectSelectionView.vue",
    "$baseDir\views\WorkspaceView.vue",
    "$baseDir\components\panels\FilePanel.vue",
    "$baseDir\components\panels\MainPanel.vue",
    "$baseDir\components\panels\ActionsPanel.vue",
    "$baseDir\components\workspace\FilePreview.vue", # Новая версия
    "$baseDir\components\workspace\TaskComposer.vue",
    "$baseDir\components\workspace\ContextSummary.vue"
)

foreach ($dir in $dirsToCreate) {
    if (-not (Test-Path $dir)) {
        Write-Host "Creating directory: $dir" -ForegroundColor Cyan
        New-Item -ItemType Directory -Path $dir | Out-Null
    }
}
foreach ($file in $filesToCreate) {
    if (-not (Test-Path $file)) {
        Write-Host "Creating file placeholder: $file" -ForegroundColor Cyan
        New-Item -ItemType File -Path $file | Out-Null
    }
}

Write-Host "UI refactoring script finished. Please provide the new file contents." -ForegroundColor Green