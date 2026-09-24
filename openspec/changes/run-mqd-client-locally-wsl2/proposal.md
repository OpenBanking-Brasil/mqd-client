## Why

`mqd-client` cannot be brought up with a single command today. `openspec/changes/mqd-server-mock`
(complete, all tasks done, not yet archived) already built a standalone local stand-in for the
central MQD server (`tools/mqd-server-mock/`), which unblocks running `mqd-client` without VPN/ICP-
Brasil access — but wiring it in is still a manual, undocumented dance (start the mock separately,
work out that the shipped `docker-compose.yaml`'s `proxy` (nginx) service is for real ICP-Brasil
mTLS passthrough and gets in the way of the mock, edit env vars by hand). Charles has now moved his
dev environment to WSL2, and every other lambda repo in this squad (`mqd-iqd_generator`,
`mqd-report_process`, `mqd-aggregator_day`, `mqd-ticket_generator`) already has a scripted
one-command `scripts/qa-up.sh` local-run experience — `mqd-client` should have the same for WSL2.

## What Changes

- Add a local-only Docker Compose override (`infra/dockerfile/docker-compose.local.yml`) that
  starts `tools/mqd-server-mock/` as a container on port `8082` in place of the real `proxy` (nginx)
  service, so `mqd-client`'s existing `PROXY_URL=http://127.0.0.1:8082` default reaches the mock
  with zero config changes.
- Add `scripts/qa-up.sh` / `scripts/qa-stop.sh` at the repo root (same conventions as the sibling
  mqd-* repos: `set -euo pipefail`, Docker checks, `docker compose -f ... -f ...` orchestration,
  a smoke-test `curl` against `/ValidateResponse` at the end).
- Add `QA_run.md` documenting the one-command flow and the WSL2-specific note that
  `network_mode: host` (already used by `mqd-client`'s service) works natively on WSL2's real Linux
  kernel, unlike Docker Desktop's Windows/Hyper-V backend.
- No changes to `src/` (the shipped app) and no changes to `tools/mqd-server-mock/`'s already-
  implemented handlers — purely additive local orchestration/tooling.

## Capabilities

### New Capabilities
- `local-dev-environment`: one-command local build+run of `mqd-client` against the existing mock
  central server, verified via Docker Compose on WSL2, with a smoke-test request confirming the app
  actually started and answers requests.

### Modified Capabilities
(none)

## Impact

- New files: `infra/dockerfile/docker-compose.local.yml`, `scripts/qa-up.sh`, `scripts/qa-stop.sh`,
  `QA_run.md`.
- No production impact: the override compose file and scripts are dev-only, never used in the real
  deployment path (`infra/dockerfile/docker-compose.yaml` alone, with the real `proxy`).
- Depends on `tools/mqd-server-mock/` (already implemented, from the `mqd-server-mock` change).
