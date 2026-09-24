## Purpose

Lets a developer bring up a fully working local instance of `mqd-client` — API, validation, and
report submission against a stand-in central server — with a single command, without VPN or real
ICP-Brasil credentials.

## ADDED Requirements

### Requirement: One-command local startup
The system SHALL provide a single script that builds and starts `mqd-client` locally, wired to a
local stand-in central server, requiring no manual multi-step setup.

#### Scenario: Fresh local start
- **WHEN** a developer with Docker running executes `scripts/qa-up.sh` from a clean checkout
- **THEN** the mock central server and `mqd-client` are both built and started, and `mqd-client`
  completes startup without the `Initialize()` fatal error caused by an unreachable central server

#### Scenario: Docker not running
- **WHEN** a developer runs `scripts/qa-up.sh` while Docker is not running
- **THEN** the script exits with a clear error message before attempting any build or compose command

### Requirement: Local startup is verified, not just started
The system SHALL confirm the running local instance actually answers API requests, not merely that
containers started.

#### Scenario: Smoke test after startup
- **WHEN** `scripts/qa-up.sh` finishes starting the local stack
- **THEN** it issues a `GET /ValidateResponse` request against the running `mqd-client` API and
  reports success or failure based on the response, before printing "environment ready"

### Requirement: Local startup does not depend on real infrastructure
The system SHALL run entirely against local/mocked dependencies — no real central server, VPN, or
ICP-Brasil certificate is required.

#### Scenario: No network access to the real central server
- **WHEN** a developer runs `scripts/qa-up.sh` with no VPN connection and no real ICP-Brasil
  certificates present
- **THEN** the local stack still starts successfully, because `PROXY_URL` points at the local mock
  instead of the real proxy/central server

### Requirement: Clean teardown without affecting the real deployment path
The system SHALL provide a way to stop the local stack that leaves the real production compose
configuration (`infra/dockerfile/docker-compose.yaml` alone) untouched.

#### Scenario: Stopping the local stack
- **WHEN** a developer runs `scripts/qa-stop.sh`
- **THEN** the local containers (mqd-client + mock) stop, and no changes are made to
  `infra/dockerfile/docker-compose.yaml` or to how the real deployment path behaves
