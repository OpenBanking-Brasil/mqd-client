## Why

`mqd-client` cannot start locally without a reachable central MQD server: on boot it calls `GET {PROXY_URL}/settings/configurationSettings.json` (`configuration_manager.go` -> `report_server_mqd.go`), and a failure there makes `main.go` call `logger.Fatal`, killing the process. Any developer without VPN/network access to the real central server (and its ICP-Brasil-backed proxy) is blocked from running or testing `mqd-client` at all. We need a lightweight local stand-in for that server's API surface.

## What Changes

- Add a standalone Go HTTP mock server that emulates the subset of the central MQD server API that `mqd-client` depends on:
  - `GET /settings/{file}.json` (specifically `configurationSettings.json`) — required at startup, returning a minimal valid `models.ConfigurationSettings` payload.
  - `POST /token` — issues a fake JWT so the report flow (`getJWKToken`) can complete without a real auth server.
  - `POST /report` — accepts and logs/stores submitted reports, returning `200 OK`, so `RESULT`/`REPORT` flows can be exercised end to end locally.
- Ship it in its own directory as an independent Go module/binary (no changes to `src/`), runnable standalone (`go run ./...`) on a configurable port, defaulting to `8082` to match the existing `PROXY_URL=http://127.0.0.1:8082` default in `settings.yml` / `docker-compose.yaml`.
- Commit this mock to the repository (it is a shared dev tool, not personal tooling) so any developer on the team can run `mqd-client` locally without depending on the real central server.
- Document how to run it and point `PROXY_URL` at it (README update), separate from the OpenSpec planning artifacts themselves, which stay local-only.

## Capabilities

### New Capabilities
- `mqd-server-mock`: a lightweight, standalone Go HTTP server that simulates the central MQD server endpoints (`/settings/*.json`, `/token`, `/report`) consumed by `mqd-client`, for local development and manual testing only.

### Modified Capabilities
(none — `mqd-client` itself is not modified by this change)

## Impact

- New code: a new top-level folder (e.g. `tools/mqd-server-mock/`) containing its own `go.mod`, `main.go`, and static/handler logic. Does not touch `src/` (the real `mqd-client` module).
- Docs: README note (root or `infra/`) explaining how to start the mock and configure `PROXY_URL` to point at it for local runs.
- No production impact: this is a dev-only utility, never built into the `mqd-client` Docker image or deployed artifact.
- No new runtime dependency for `mqd-client` itself — the mock is a separate, independently buildable program.
