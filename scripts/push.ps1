param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$PushArgs
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

& (Join-Path $PSScriptRoot "lint.ps1")
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

git push @PushArgs
exit $LASTEXITCODE
