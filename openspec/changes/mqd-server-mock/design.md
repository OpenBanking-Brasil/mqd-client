## Context

`mqd-client`'s `Initialize()` path (`src/application/configuration_manager.go`) blocks on `ReportServerMQD.LoadConfigurationSettings()`, a `GET {PROXY_URL}/settings/configurationSettings.json` call that must return a JSON body matching `models.ConfigurationSettings` (`src/domain/models/configuration_settings.go`, `api_configuration_settings.go`). Later, `ReportServerMQD.SendReport()` does `POST /token` (form-encoded `grant_type`/`client_id`, response matching `jwt.JWKToken`) followed by `POST /report` with a `models.Report` body (`src/domain/services/report_server_mqd.go`, `api_dao.go`). Without a real central server reachable via `PROXY_URL`, none of this works locally.

## Goals / Non-Goals

**Goals:**
- Provide a runnable local stand-in for exactly these three endpoints, shaped to satisfy `mqd-client`'s Go structs.
- Make the fixture data (config settings, validation rules) easy to edit without recompiling.
- Keep it trivially simple to start (`go run`) and safe to commit (no secrets, no real business logic).

**Non-Goals:**
- No real authentication/JWT signing or verification — `mqd-client` never verifies the token signature, it only reads `access_token`/`expires_in` off the JSON body.
- No mTLS / certificate handling, no S3-compatible semantics, no persistence layer for received reports beyond stdout logging.
- Not a source of truth for real OpenFinance validation rules — fixtures are illustrative only.

## Decisions

- **Separate Go module under `tools/mqd-server-mock/`, not a package inside `src/`.**
  Alternative considered: a `cmd/mock` package inside the existing `src` module. Rejected because it would pull a dev-only tool into `mqd-client`'s own `go.mod`/build graph and Docker image; a separate module keeps `src/` (the shipped artifact) untouched, per the proposal's Impact section.

- **Fixtures as JSON files on disk (default set embedded via `go:embed`, overridable via a `-fixtures` dir flag / `FIXTURES_DIR` env var), not hardcoded Go structs.**
  Alternative considered: hardcode the response in Go. Rejected because devs will want to tweak `APIGroupSettings`/validation rates per scenario without recompiling; a JSON file matching the real shape (see `models.ConfigurationSettings`) is easier to diff and reason about.

- **Generic `GET /settings/{file}` handler**, serving any file from the fixtures dir (mirroring the real server's `settingsPath + "/" + filePath` pattern used by both `LoadConfigurationSettings` and `LoadAPIConfigurationFile`). This lets the same mock also serve JSON-schema files referenced by validation rules, not just `configurationSettings.json`, with no extra endpoint code.

- **`POST /token` returns a static, non-expiring-in-practice fake JWT** (`access_token: "mock-token"`, large `expires_in`), no real crypto. `mqd-client` only checks `expires_in` client-side (`jwt.ValidateExpiration`), so no real signing is needed.

- **`POST /report` just logs the decoded body to stdout and returns `200 OK`** with a small JSON ack. No storage/DB — non-goal per above. If someone needs to inspect reports later, stdout redirection is enough for a dev tool.

- **Default port `8082`**, matching the existing `PROXY_URL=http://127.0.0.1:8082` default already present in `src/settings/settings.yml` and `infra/dockerfile/docker-compose.yaml`, so zero config is needed for the common case; overridable via `-port` / `PORT`.

## Risks / Trade-offs

- [Fixture JSON drifts from the real `models.ConfigurationSettings` shape over time] → keep fixtures next to a short README noting which Go structs they must match; mismatches surface immediately as `mqd-client` logs a JSON unmarshal error at startup.
- [Someone mistakes mock validation-rate/rule fixtures for real production rules] → README banner: "local dev only, not representative of production validation rules."
- [Mock and real server behavior diverge silently (e.g. real server adds a new required field)] → out of scope to auto-sync; acceptable since this is a manual dev convenience tool, not a contract test.

## Migration Plan

Purely additive — new folder, no changes to `src/`, no deployment. Document usage (README snippet: build/run command, how to point `PROXY_URL` at it) as part of `tasks.md`. Nothing to roll back beyond deleting the folder.

## Open Questions

- None blocking. Optional follow-up (not in this change): add a `mqd-server-mock` service to `infra/dockerfile/docker-compose.yaml` for one-command local startup — can be proposed later if useful.
