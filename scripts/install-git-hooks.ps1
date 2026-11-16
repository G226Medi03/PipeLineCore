$ErrorActionPreference = "Stop"

Set-Location (Resolve-Path (Join-Path $PSScriptRoot ".."))

if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
    throw "golangci-lint is not installed. Install it before enabling git hooks."
}

git config core.hooksPath .githooks
if ($LASTEXITCODE -ne 0) {
    throw "failed to configure core.hooksPath"
}

Write-Host "git hooks enabled: core.hooksPath=.githooks"
Write-Host "pre-push runs: scripts/lint.ps1 (builds and runs custom golangci-lint)"
