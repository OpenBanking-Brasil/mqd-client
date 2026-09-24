## 1. Module setup

- [x] 1.1 Create `tools/mqd-server-mock/` with its own `go.mod` (module `mqd-server-mock`, Go 1.23+), independent from `src/go.mod`
- [x] 1.2 Add default fixtures under `tools/mqd-server-mock/fixtures/` (`configurationSettings.json` matching `models.ConfigurationSettings`) and embed them via `go:embed` as the built-in default set

## 2. HTTP handlers

- [x] 2.1 Implement `GET /settings/{fileName}` reading from the fixtures directory (embedded default, or `-fixtures`/`FIXTURES_DIR` override if set), returning `200`+JSON or `404`+JSON error
- [x] 2.2 Implement `POST /token` parsing the form-encoded body and returning a static `jwt.JWKToken`-shaped JSON response with a large `expires_in`
- [x] 2.3 Implement `POST /report` decoding the body as JSON, logging a one-line summary (`ClientID`, `DataOwnerID`) to stdout, returning `200`; returning `400`+JSON error on invalid JSON

## 3. Server wiring & config

- [x] 3.1 Wire handlers into an `http.ServeMux` in `main.go`, with `-port`/`PORT` (default `8082`) and `-fixtures`/`FIXTURES_DIR` (default: embedded fixtures) flags
- [x] 3.2 Add request logging middleware (method, path, status) to stdout for visibility during local dev

## 4. Docs & verification

- [x] 4.1 Add `tools/mqd-server-mock/README.md`: how to run it (`go run .`), how to point `mqd-client`'s `PROXY_URL` at it, and the "local dev only, not representative of production rules" disclaimer
- [x] 4.2 Manually verify end-to-end: start the mock, run `mqd-client` from `src/` with `PROXY_URL=http://127.0.0.1:8082`, confirm it starts without the `Initialize()` fatal error
- [x] 4.3 Manually verify `POST /report` path: trigger a report send from `mqd-client` (or `curl`) and confirm the mock logs it and returns `200`
