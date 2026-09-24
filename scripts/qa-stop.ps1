<#
.SYNOPSIS
    Para os containers do QA local do mqd-client (app + mqd-server-mock), sem
    removê-los. Use scripts/qa-up.ps1 pra retomar.
#>

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot

Push-Location $repoRoot
try {
    Write-Host "Parando containers..." -ForegroundColor Cyan
    docker compose stop
    if ($LASTEXITCODE -ne 0) { throw "Falha ao parar os containers." }
    Write-Host "Parado. Pra retomar: scripts/qa-up.ps1" -ForegroundColor Green
}
finally {
    Pop-Location
}
