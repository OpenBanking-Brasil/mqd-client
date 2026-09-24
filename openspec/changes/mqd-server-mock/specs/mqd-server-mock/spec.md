## ADDED Requirements

### Requirement: Serve settings files
The mock server SHALL expose `GET /settings/{fileName}` and respond with the raw contents of `{fileName}` read from a configured fixtures directory, with `Content-Type: application/json` and HTTP `200`.

#### Scenario: Configuration settings file requested
- **WHEN** a client sends `GET /settings/configurationSettings.json`
- **THEN** the server responds `200 OK` with the JSON contents of `configurationSettings.json` from the fixtures directory, matching the shape of `models.ConfigurationSettings` (fields `Version`, `ValidationSettings`, `ReportSettings`)

#### Scenario: Requested file does not exist
- **WHEN** a client sends `GET /settings/{fileName}` for a file not present in the fixtures directory
- **THEN** the server responds with HTTP `404` and a JSON error body

### Requirement: Issue mock access token
The mock server SHALL expose `POST /token`, accept a form-encoded body containing `grant_type` and `client_id`, and respond `200 OK` with a JSON body matching the `jwt.JWKToken` shape (`access_token`, `token_type`, `expires_in`, `refresh_expires_in`, `not-before-policy`, `scope`), with `expires_in` large enough that a single local dev session never needs to re-request a token.

#### Scenario: Token requested with client credentials
- **WHEN** a client sends `POST /token` with `Content-Type: application/x-www-form-urlencoded` and body `grant_type=client_credentials&client_id=<any-uuid>`
- **THEN** the server responds `200 OK` with a JSON body containing a non-empty `access_token` and an `expires_in` greater than zero

### Requirement: Accept submitted reports
The mock server SHALL expose `POST /report`, accept a JSON body matching `models.Report`, log a summary of the received report to stdout, and respond `200 OK`.

#### Scenario: Valid report submitted
- **WHEN** a client sends `POST /report` with a JSON body decodable as `models.Report` and header `Authorization: Bearer <token>`
- **THEN** the server responds `200 OK` and writes a log line containing the report's `ClientID` and `DataOwnerID` to stdout

#### Scenario: Malformed report body
- **WHEN** a client sends `POST /report` with a body that is not valid JSON
- **THEN** the server responds with HTTP `400` and a JSON error body

### Requirement: Configurable listen port and fixtures directory
The mock server SHALL listen on port `8082` by default (matching `mqd-client`'s default `PROXY_URL`) and SHALL allow overriding both the port and the fixtures directory via command-line flags and equivalent environment variables, without requiring a source change or rebuild.

#### Scenario: Default startup
- **WHEN** the mock server is started with no flags or environment variables set
- **THEN** it listens on `:8082` and serves fixtures from its built-in default set

#### Scenario: Overridden port and fixtures directory
- **WHEN** the mock server is started with `-port 9000 -fixtures ./my-fixtures` (or equivalent `PORT`/`FIXTURES_DIR` environment variables)
- **THEN** it listens on `:9000` and serves files from `./my-fixtures` instead of the built-in defaults
