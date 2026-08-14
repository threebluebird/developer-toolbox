$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path -Parent $PSScriptRoot
$outputPath = Join-Path $projectRoot 'build\bin\developer-toolbox.exe'
$env:GOCACHE = Join-Path $projectRoot '.codex-go-cache'

Push-Location $projectRoot
try {
    wails build -clean -platform windows/amd64 -o developer-toolbox.exe
    if (-not (Test-Path -LiteralPath $outputPath)) { throw "Package was not created: $outputPath" }
    $file = Get-Item -LiteralPath $outputPath
    if ($file.Length -lt 1MB) { throw "Packaged executable is unexpectedly small: $($file.Length) bytes" }
    Write-Output "Packaged $($file.FullName) ($($file.Length) bytes)"
} finally {
    Pop-Location
}
