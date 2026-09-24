#!/bin/bash
# Recria o QA local do mqd-client do zero: remove os containers deste projeto
# (app + mqd-server-mock) e roda o qa-up de novo. Este projeto não tem banco nem
# volumes, então não há dados pra limpar — e nada aqui afeta o mqd-qa-infra ou
# outros projetos.
#
# Equivalente bash de qa-reset.ps1.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$REPO_ROOT"
echo "Removendo containers deste projeto..."
docker compose down --remove-orphans

exec "$SCRIPT_DIR/qa-up.sh" "$@"
