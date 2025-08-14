function Show-Tree {
    param(
        [string]$Path = ".",
        [string[]]$Exclude = @('node_modules','.git','.idea','dist','build','.cache','.vite','.wails','out','target','bin','obj','coverage')
    )

    $excludePattern = '^(?:' + ($Exclude -join '|').Replace('.', '\.') + ')$'

    function Draw-Tree($dir, $prefix = '') {
        $items = Get-ChildItem -LiteralPath $dir -Force |
                 Where-Object { $_.Name -notmatch $excludePattern }

        for ($i = 0; $i -lt $items.Count; $i++) {
            $item = $items[$i]
            $connector = if ($i -eq $items.Count - 1) { '└── ' } else { '├── ' }
            Write-Output "$prefix$connector$($item.Name)"
            if ($item.PSIsContainer) {
                $newPrefix = if ($i -eq $items.Count - 1) { "$prefix    " } else { "$prefix│   " }
                Draw-Tree $item.FullName $newPrefix
            }
        }
    }

    Draw-Tree (Resolve-Path $Path)
}

Show-Tree
