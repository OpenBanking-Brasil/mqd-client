## 1. Mock server containerization

- [ ] 1.1 Add `tools/mqd-server-mock/Dockerfile` (build against the module's own `go.mod`, matching
  `mqd-client`'s Go version) and verify `docker build -f tools/mqd-server-mock/Dockerfile
  tools/mqd-server-mock` succeeds and produces a runnable image
- [ ] 1.2 Verify the built image serves `GET /settings/configurationSettings.json`, `POST /token`,
  `POST /report` correctly when run standalone (`docker run -p 8082:8082 <image>` + `curl`)

## 2. Compose override

- [ ] 2.1 Add `infra/dockerfile/docker-compose.local.yml`: a `mqd-server-mock` service (built from
  `tools/mqd-server-mock/Dockerfile`) bound to `8082`, and an `mqd-client` override that depends on
  `mqd-server-mock` instead of `proxy`
- [ ] 2.2 Verify `docker compose -f infra/dockerfile/docker-compose.yaml -f
  infra/dockerfile/docker-compose.local.yml config` renders a valid merged config with `proxy`
  excluded from `mqd-client`'s `depends_on` and the mock present

## 3. Scripts

- [ ] 3.1 Add `scripts/qa-up.sh`: checks Docker is running (clear error + exit 1 if not), builds
  both images, brings the local stack up via the compose override, waits for `mqd-client` to be
  ready, and exits non-zero with a clear message on any failed step
- [ ] 3.2 Add the smoke test to the end of `qa-up.sh`: `curl` the `GET /ValidateResponse` endpoint
  (headers per `docs/install/INSTALL.md`) and report pass/fail explicitly
- [ ] 3.3 Add `scripts/qa-stop.sh`: stops the local stack (compose override) without touching the
  real `infra/dockerfile/docker-compose.yaml` path
- [ ] 3.4 chmod +x both scripts and verify they run from a fresh WSL2 shell (`bash scripts/qa-up.sh`)

## 4. Docs & end-to-end verification

- [ ] 4.1 Add `QA_run.md` documenting the one-command flow, the WSL2 `network_mode: host` note, and
  how this relates to the existing `docs/install/INSTALL.md` (real deployment) and
  `tools/mqd-server-mock/README.md` (mock details)
- [ ] 4.2 Full manual run on WSL2: `scripts/qa-up.sh` from a clean checkout completes with the smoke
  test passing; confirm via `docker compose ... logs mqd-client` that startup did not hit the
  `Initialize()` fatal error
- [ ] 4.3 Full manual run of `scripts/qa-stop.sh`, confirm containers stop and `docker compose ... up`
  again brings the same environment back cleanly
