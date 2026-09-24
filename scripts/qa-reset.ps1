<#
.SYNOPSIS
    Recria o QA local do mqd-client do zero: remove os containers deste projeto
    (app + mqd-server-mock) e roda o qa-up de novo. Este projeto não tem banco
    nem volumes, então não há dados pra limpar — e nada aqui afeta o
    mqd-qa-infra ou outros projetos.

.PARAMETER WaitReport
    Repassado pro scripts/qa-up.ps1.
#>

param(
    [switch]$WaitReport
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot

Push-Location $repoRoot
try {
    Write-Host "Removendo containers deste projeto..." -ForegroundColor Yellow
    docker compose down --remove-orphans
    if ($LASTEXITCODE -ne 0) { throw "Falha ao remover os containers." }
}
finally {
    Pop-Location
}

& (Join-Path $PSScriptRoot "qa-up.ps1") -WaitReport:$WaitReport
exit $LASTEXITCODE
