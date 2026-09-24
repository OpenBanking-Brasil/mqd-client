<#
.SYNOPSIS
    Sobe o ambiente de QA local do mqd-client: builda e sobe o mqd-server-mock
    (substituto local do servidor central, tools/mqd-server-mock) e o mqd-client
    apontando pra ele (docker-compose.yml da raiz), e faz um smoke test.

    NÃO depende do mqd-qa-infra: o mqd-client não usa Postgres, WireMock nem
    LocalStack — só conversa com o servidor central do MQD (via PROXY_URL).

.PARAMETER WaitReport
    Espera o mqd-client enviar o primeiro relatório (POST /report) pro mock —
    leva até REPORT_EXECUTION_WINDOW minutos (1 no .env).
#>

param(
    [switch]$WaitReport
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$clientPort = if ($env:MQD_CLIENT_PORT) { $env:MQD_CLIENT_PORT } else { "8080" }
$mockPort = if ($env:MQD_SERVER_MOCK_PORT) { $env:MQD_SERVER_MOCK_PORT } else { "8082" }

function Get-HttpStatus {
    param([string]$Uri, [string]$Method = "GET", [hashtable]$Headers = @{}, [string]$Body)
    try {
        $params = @{ Uri = $Uri; Method = $Method; Headers = $Headers; UseBasicParsing = $true; TimeoutSec = 10 }
        if ($Body) { $params.Body = $Body; $params.ContentType = "application/json" }
        return [int](Invoke-WebRequest @params).StatusCode
    } catch {
        # Invoke-WebRequest lança exceção pra status 4xx/5xx (PowerShell 5.1 e 7+).
        if ($_.Exception.Response) { return [int]$_.Exception.Response.StatusCode }
        return 0
    }
}

function Get-MockLogs {
    return (docker compose logs mqd-server-mock 2>&1 | Out-String)
}

Write-Host "Verificando Docker..." -ForegroundColor Cyan
docker info *> $null
if ($LASTEXITCODE -ne 0) {
    Write-Error "Docker não está rodando. Abra o Docker Desktop e tente novamente."
    exit 1
}

Push-Location $repoRoot
try {
    $envFile = Join-Path $repoRoot ".env"
    if (-not (Test-Path $envFile)) {
        Write-Host "Criando .env a partir de .env.example..." -ForegroundColor Cyan
        Copy-Item (Join-Path $repoRoot ".env.example") $envFile
    }

    Write-Host "Buildando e subindo mqd-server-mock + mqd-client..." -ForegroundColor Cyan
    docker compose up -d --build --wait
    if ($LASTEXITCODE -ne 0) {
        docker compose logs --tail 50
        throw "Os containers não ficaram healthy. Veja: docker compose logs"
    }

    Write-Host ""
    Write-Host "Smoke test..." -ForegroundColor Cyan
    $failed = $false

    # 1) Na inicialização o mqd-client baixa as configurações do servidor central.
    if ((Get-MockLogs) -match [regex]::Escape("GET /settings/configurationSettings.json -> 200")) {
        Write-Host "  OK    mqd-client baixou configurationSettings.json do mock" -ForegroundColor Green
    } else {
        Write-Host "  FALHA mqd-client não pediu GET /settings/configurationSettings.json ao mock" -ForegroundColor Red
        $failed = $true
    }

    # 2) API de métricas no ar.
    $status = Get-HttpStatus -Uri "http://localhost:$clientPort/metrics"
    if ($status -eq 200) {
        Write-Host "  OK    GET http://localhost:$clientPort/metrics" -ForegroundColor Green
    } else {
        Write-Host "  FALHA GET http://localhost:$clientPort/metrics (HTTP $status)" -ForegroundColor Red
        $failed = $true
    }

    # 3) API de validação respondendo. A fixture padrão do mock não tem nenhum
    #    endpoint configurado (APIGroupSettings vazio, ver tools/mqd-server-mock/README.md),
    #    então o esperado aqui é 400 "endpointName: Not found" — prova que a API
    #    recebeu, leu os headers e consultou as configurações.
    $status = Get-HttpStatus -Uri "http://localhost:$clientPort/ValidateResponse" -Method "POST" `
        -Headers @{ "serverOrgId" = "d7384bd0-842f-43c5-be02-9d2b2d5efc2c"; "endpointName" = "/accounts/v2/accounts" } `
        -Body '{"data":{}}'
    if ($status -eq 400 -or $status -eq 200) {
        Write-Host "  OK    POST /ValidateResponse respondeu HTTP $status" -ForegroundColor Green
    } else {
        Write-Host "  FALHA POST /ValidateResponse respondeu HTTP $status" -ForegroundColor Red
        $failed = $true
    }

    if ($WaitReport) {
        Write-Host ""
        Write-Host "Esperando o primeiro POST /report no mock (até ~2 min)..." -ForegroundColor Cyan
        $received = $false
        for ($i = 0; $i -lt 40; $i++) {
            $logs = Get-MockLogs
            if ($logs -match [regex]::Escape("POST /report -> 200")) {
                $received = $true
                Write-Host "  OK    relatório recebido pelo mock:" -ForegroundColor Green
                ($logs -split "`r?`n" | Where-Object { $_ -match "received report" } | Select-Object -Last 1) | ForEach-Object { Write-Host "        $_" }
                break
            }
            Start-Sleep -Seconds 3
        }
        if (-not $received) {
            Write-Host "  FALHA nenhum POST /report chegou no mock" -ForegroundColor Red
            $failed = $true
        }
    }

    Write-Host ""
    Write-Host "mqd-client:        http://localhost:$clientPort (POST /ValidateResponse, GET /metrics)" -ForegroundColor DarkGray
    Write-Host "mqd-server-mock:   http://localhost:$mockPort" -ForegroundColor DarkGray
    Write-Host "Logs:              docker compose logs -f app mqd-server-mock" -ForegroundColor DarkGray

    if ($failed) { throw "Smoke test falhou (ver acima)." }
    Write-Host "Ambiente pronto." -ForegroundColor Green
}
finally {
    Pop-Location
}
