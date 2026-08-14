$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$goFiles = Get-ChildItem -Path $projectRoot -Recurse -Filter '*.go' |
    Where-Object { $_.FullName -notmatch '\\frontend\\node_modules\\' }

$unformatted = @($goFiles | ForEach-Object { gofmt -l $_.FullName })
if ($unformatted.Count -gt 0) {
    throw "Go files require formatting: $($unformatted -join ', ')"
}

Push-Location $projectRoot
try {
	$env:GOCACHE = Join-Path $projectRoot '.codex-go-cache'
    go test ./...
    go vet ./...
	go test -race ./backend/services ./backend/storage ./backend/worker
	go test -run '^$' -bench '100KB' -benchtime 200ms ./backend/tools/json ./backend/tools/base64tool
} finally {
    Pop-Location
}

Push-Location (Join-Path $projectRoot 'frontend')
try {
	$node = Get-Command node -ErrorAction SilentlyContinue
	if (-not $node) { throw 'Node.js is required for frontend verification.' }
	& $node.Source --check src/main.js
	& $node.Source --test src/components/editor/editor.test.js
	& $node.Source ./node_modules/vite/bin/vite.js build
} finally {
    Pop-Location
}
