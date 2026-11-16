param(
    [switch]$ChangedOnly
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
    throw "golangci-lint is not installed. Install it before running lint."
}

$binary = Join-Path $root "bin/custom-gcl.exe"
if (-not (Test-Path $binary)) {
    $binary = Join-Path $root "bin/custom-gcl"
}

$needsBuild = -not (Test-Path $binary)
if (-not $needsBuild) {
    $binaryTime = (Get-Item $binary).LastWriteTimeUtc
    if ((Get-Item (Join-Path $root ".custom-gcl.yml")).LastWriteTimeUtc -gt $binaryTime) {
        $needsBuild = $true
    }
    if (Get-ChildItem (Join-Path $root "tools/nodeconstructor") -Recurse -File |
        Where-Object { $_.LastWriteTimeUtc -gt $binaryTime }) {
        $needsBuild = $true
    }
}

if ($needsBuild) {
    Write-Host "building custom golangci-lint..."
    golangci-lint custom
    if ($LASTEXITCODE -ne 0) {
        throw "failed to build custom golangci-lint"
    }
}

if ($ChangedOnly) {
    $base = git merge-base origin/main HEAD 2>$null
    if ($LASTEXITCODE -eq 0 -and $base) {
        & $binary run --new-from-rev=$base ./...
        exit $LASTEXITCODE
    }
}

& $binary run ./...
exit $LASTEXITCODE
