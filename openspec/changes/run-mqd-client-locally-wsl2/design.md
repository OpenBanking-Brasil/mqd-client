## Context

See proposal.md - Why. Relevant existing pieces:

- `infra/dockerfile/docker-compose.yaml` already defines `mqd-client` (`network_mode: host`,
  `PROXY_URL=http://127.0.0.1:8082` by default) and a `proxy` (nginx) sidecar that terminates real
  ICP-Brasil mTLS and forwards to the real central server — required in production, irrelevant (and
  in the way, since it has no valid upstream/certs locally) for local dev.
- `tools/mqd-server-mock/` (from the already-complete `mqd-server-mock` change) is a standalone Go
  module/binary serving `GET /settings/{file}.json`, `POST /token`, `POST /report` — the exact
  surface `mqd-client` calls at startup and during report submission. It was already manually
  verified to work with `mqd-client` when `PROXY_URL` points straight at it, bypassing `proxy`
  entirely (see that change's tasks 4.2/4.3).
- `network_mode: host` is Linux-only Docker behavior. WSL2 runs a real Linux kernel, so this works
  the same as on a native Linux host — unlike Docker Desktop's non-WSL2 backends, where
  `network_mode: host` is unsupported/partial. No adaptation needed for WSL2 specifically.

## Goals / Non-Goals

**Goals:**
- One command to build and run `mqd-client` locally end-to-end on WSL2, using the existing mock.
- Reuse `tools/mqd-server-mock/` unchanged — this change is orchestration only.
- Match the script/doc conventions already established across the sibling mqd-* repos.

**Non-Goals:**
- Not touching the real production compose path (`docker-compose.yaml` + `proxy`) — it stays exactly
  as-is for real deployments.
- Not adding HTTPS/mTLS/cert handling for local dev (mirrors `mqd-server-mock`'s own non-goals).
- Not building a shared-Postgres-style QA infra like the other repos — `mqd-client` has no database
  dependency, so there's nothing to share/orchestrate beyond the mock.

## Decisions

- **Docker Compose override file (`docker-compose.local.yml`) layered on top of the real
  `docker-compose.yaml`, not a full separate compose file.**
  Alternative considered: a fully standalone local compose file duplicating `mqd-client`'s service
  definition. Rejected — duplicating the service definition would drift from the real one over time
  (ports, resource limits, volumes); an override (`docker compose -f docker-compose.yaml -f
  docker-compose.local.yml up`) only needs to redefine what actually changes: replace `proxy` with
  the mock, and adjust `mqd-client`'s `depends_on`.

- **The mock replaces `proxy` on the same port (`8082`) rather than running alongside it.**
  Alternative considered: run the mock on a different port and repoint `PROXY_URL` via an env
  override. Rejected — reusing the existing default port means zero env-var changes are needed
  anywhere (`docker-compose.yaml`, `settings.yml` already default to `8082`), and it avoids ever
  running both `proxy` and the mock at once, which would just waste resources since nothing uses
  `proxy` in local/mock mode.

- **`scripts/qa-up.sh` builds `tools/mqd-server-mock` as its own Docker image on the fly** (a small
  inline `Dockerfile` step or `docker build` against `tools/mqd-server-mock/`, since that module has
  no Dockerfile of its own today — it was designed to be run via `go run .`, not containerized).
  Alternative considered: run the mock as a bare host process (`go run .` in the background) instead
  of a container. Rejected — keeping everything inside `docker compose up` means one process
  supervises the whole stack, one `qa-stop.sh` cleanly tears it down, and it matches how every
  sibling repo's `qa-up.sh` works (everything containerized, nothing left running loose on the host).
  A `Dockerfile` for `tools/mqd-server-mock/` is added as part of this change (dev-only, lives next
  to the mock, never referenced by the real deployment path).

- **Smoke test**: `qa-up.sh` finishes with the same `curl` from `docs/install/INSTALL.md`
  (`GET /ValidateResponse` with `serverOrgId`/`endpointName` headers) against `localhost:8080`,
  confirming the whole chain (compose up → mqd-client boots against the mock → API responds) works,
  not just "containers started."

## Risks / Trade-offs

- [Compose override syntax (`-f a -f b`) is easy to forget/type wrong by hand] → `scripts/qa-up.sh`
  always passes both files explicitly; devs never need to remember the flag combination themselves.
- [A Dockerfile added for the mock could bit-rot if `mqd-server-mock`'s own go.mod changes] →
  it's a two-line build (`golang:1.27-alpine` build stage matching `mqd-client`'s own Go version,
  copy, `go build`), kept next to `tools/mqd-server-mock/README.md` so both are updated together.
- [Someone runs the override in a context that looks like production] → `QA_run.md` and the override
  filename itself (`docker-compose.local.yml`) make the dev-only intent explicit; mirrors the
  existing `mqd-server-mock` README's "local dev only" disclaimer.

## Migration Plan

Purely additive — new files only, nothing removed or changed in the real deployment path. Nothing to
roll back beyond deleting the new files.
