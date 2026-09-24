#!/bin/bash
# Sobe o ambiente de QA local do mqd-client: builda e sobe o mqd-server-mock
# (substituto local do servidor central, tools/mqd-server-mock) e o mqd-client
# apontando pra ele (docker-compose.yml da raiz), e faz um smoke test.
#
# NÃO depende do mqd-qa-infra: o mqd-client não usa Postgres, WireMock nem
# LocalStack — só conversa com o servidor central do MQD (via PROXY_URL).
#
# Uso: ./scripts/qa-up.sh [--wait-report]
#   --wait-report  espera o mqd-client enviar o primeiro relatório (POST /report)
#                  pro mock — leva até REPORT_EXECUTION_WINDOW minutos (1 no .env).
#
# Equivalente bash de qa-up.ps1 (mesma lógica), pra uso dentro do WSL2/Linux.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
CLIENT_PORT="${MQD_CLIENT_PORT:-8080}"

WAIT_REPORT=false
if [ "${1:-}" = "--wait-report" ]; then
    WAIT_REPORT=true
fi

echo "Verificando Docker..."
if ! docker info > /dev/null 2>&1; then
    echo "ERRO: Docker não está rodando." >&2
    exit 1
fi

cd "$REPO_ROOT"

if [ ! -f "$REPO_ROOT/.env" ]; then
    echo "Criando .env a partir de .env.example..."
    cp "$REPO_ROOT/.env.example" "$REPO_ROOT/.env"
fi

echo "Buildando e subindo mqd-server-mock + mqd-client..."
if ! docker compose up -d --build --wait; then
    echo "ERRO: os containers não ficaram healthy. Veja: docker compose logs" >&2
    docker compose logs --tail 50 >&2 || true
    exit 1
fi

echo ""
echo "Smoke test..."
FAILED=false

# 1) Na inicialização o mqd-client baixa as configurações do servidor central.
if docker compose logs mqd-server-mock | grep -q "GET /settings/configurationSettings.json -> 200"; then
    echo "  OK   mqd-client baixou configurationSettings.json do mock"
else
    echo "  FALHA mqd-client não pediu GET /settings/configurationSettings.json ao mock" >&2
    FAILED=true
fi

# 2) API de métricas no ar.
if curl -fsS -o /dev/null "http://localhost:${CLIENT_PORT}/metrics"; then
    echo "  OK   GET http://localhost:${CLIENT_PORT}/metrics"
else
    echo "  FALHA GET http://localhost:${CLIENT_PORT}/metrics" >&2
    FAILED=true
fi

# 3) API de validação respondendo. A fixture padrão do mock não tem nenhum
#    endpoint configurado (APIGroupSettings vazio, ver tools/mqd-server-mock/README.md),
#    então o esperado aqui é 400 "endpointName: Not found" — prova que a API
#    recebeu, leu os headers e consultou as configurações.
STATUS="$(curl -s -o /dev/null -w '%{http_code}' -X POST "http://localhost:${CLIENT_PORT}/ValidateResponse" \
    -H 'Content-Type: application/json' \
    -H 'serverOrgId: d7384bd0-842f-43c5-be02-9d2b2d5efc2c' \
    -H 'endpointName: /accounts/v2/accounts' \
    -d '{"data":{}}' || true)"
if [ "$STATUS" = "400" ] || [ "$STATUS" = "200" ]; then
    echo "  OK   POST /ValidateResponse respondeu HTTP $STATUS"
else
    echo "  FALHA POST /ValidateResponse respondeu HTTP $STATUS" >&2
    FAILED=true
fi

if [ "$WAIT_REPORT" = true ]; then
    echo ""
    echo "Esperando o primeiro POST /report no mock (até ~2 min)..."
    for _ in $(seq 1 40); do
        if docker compose logs mqd-server-mock | grep -q "POST /report -> 200"; then
            echo "  OK   relatório recebido pelo mock:"
            docker compose logs mqd-server-mock | grep "received report" | tail -1 | sed 's/^/       /'
            break
        fi
        sleep 3
    done
    if ! docker compose logs mqd-server-mock | grep -q "POST /report -> 200"; then
        echo "  FALHA nenhum POST /report chegou no mock" >&2
        FAILED=true
    fi
fi

echo ""
echo "mqd-client:        http://localhost:${CLIENT_PORT} (POST /ValidateResponse, GET /metrics)"
echo "mqd-server-mock:   http://localhost:${MQD_SERVER_MOCK_PORT:-8082}"
echo "Logs:              docker compose logs -f app mqd-server-mock"

if [ "$FAILED" = true ]; then
    echo "" >&2
    echo "ERRO: smoke test falhou (ver acima)." >&2
    exit 1
fi
echo "Ambiente pronto."
