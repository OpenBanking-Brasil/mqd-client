#!/bin/bash
# Para os containers do QA local do mqd-client (app + mqd-server-mock), sem
# removê-los. Use scripts/qa-up.sh pra retomar.
#
# Equivalente bash de qa-stop.ps1.

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$REPO_ROOT"
echo "Parando containers..."
docker compose stop
echo "Parado. Pra retomar: scripts/qa-up.sh"
