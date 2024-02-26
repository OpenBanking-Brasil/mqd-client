# Details

Date : 2024-02-05 17:59:12

Directory c:\\Code\\OFB\\MQD\\mqd-client

Total : 42 files,  8630 codes, 674 comments, 513 blanks, all 9817 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [mqd-client/README.md](/mqd-client/README.md) | Markdown | 22 | 0 | 11 | 33 |
| [mqd-client/docs/Arquitetura/FLUXO_API.md](/mqd-client/docs/Arquitetura/FLUXO_API.md) | Markdown | 69 | 0 | 9 | 78 |
| [mqd-client/docs/Arquitetura/FLUXO_RESULTS.md](/mqd-client/docs/Arquitetura/FLUXO_RESULTS.md) | Markdown | 18 | 0 | 4 | 22 |
| [mqd-client/docs/Arquitetura/FLUXO_VALIDATION.md](/mqd-client/docs/Arquitetura/FLUXO_VALIDATION.md) | Markdown | 13 | 0 | 4 | 17 |
| [mqd-client/docs/ENDPOINTS.md](/mqd-client/docs/ENDPOINTS.md) | Markdown | 29 | 0 | 8 | 37 |
| [mqd-client/docs/README.md](/mqd-client/docs/README.md) | Markdown | 12 | 0 | 5 | 17 |
| [mqd-client/docs/specification/mqd-client-openapi.yml](/mqd-client/docs/specification/mqd-client-openapi.yml) | YAML | 101 | 0 | 3 | 104 |
| [mqd-client/infra/README.md](/mqd-client/infra/README.md) | Markdown | 23 | 0 | 13 | 36 |
| [mqd-client/infra/dockerfile/docker-compose.yaml](/mqd-client/infra/dockerfile/docker-compose.yaml) | YAML | 25 | 0 | 1 | 26 |
| [mqd-client/src/Makefile](/mqd-client/src/Makefile) | Makefile | 13 | 4 | 4 | 21 |
| [mqd-client/src/application/api_server.go](/mqd-client/src/application/api_server.go) | Go | 180 | 77 | 41 | 298 |
| [mqd-client/src/application/configuration_manager.go](/mqd-client/src/application/configuration_manager.go) | Go | 203 | 28 | 44 | 275 |
| [mqd-client/src/application/message_process_Worker.go](/mqd-client/src/application/message_process_Worker.go) | Go | 116 | 43 | 26 | 185 |
| [mqd-client/src/application/queue_manager.go](/mqd-client/src/application/queue_manager.go) | Go | 21 | 7 | 7 | 35 |
| [mqd-client/src/application/result_processor.go](/mqd-client/src/application/result_processor.go) | Go | 224 | 67 | 38 | 329 |
| [mqd-client/src/crosscutting/configuration/configuration.go](/mqd-client/src/crosscutting/configuration/configuration.go) | Go | 47 | 14 | 10 | 71 |
| [mqd-client/src/crosscutting/errorhandling/error.go](/mqd-client/src/crosscutting/errorhandling/error.go) | Go | 6 | 1 | 2 | 9 |
| [mqd-client/src/crosscutting/general.go](/mqd-client/src/crosscutting/general.go) | Go | 8 | 1 | 3 | 12 |
| [mqd-client/src/crosscutting/log/json_log.go](/mqd-client/src/crosscutting/log/json_log.go) | Go | 68 | 82 | 17 | 167 |
| [mqd-client/src/crosscutting/log/logger.go](/mqd-client/src/crosscutting/log/logger.go) | Go | 28 | 11 | 6 | 45 |
| [mqd-client/src/crosscutting/log/logger_factory.go](/mqd-client/src/crosscutting/log/logger_factory.go) | Go | 14 | 5 | 5 | 24 |
| [mqd-client/src/crosscutting/monitoring/metrics.go](/mqd-client/src/crosscutting/monitoring/metrics.go) | Go | 191 | 94 | 40 | 325 |
| [mqd-client/src/crosscutting/security/jwt/jwt.go](/mqd-client/src/crosscutting/security/jwt/jwt.go) | Go | 70 | 24 | 18 | 112 |
| [mqd-client/src/crosscutting/tools.go](/mqd-client/src/crosscutting/tools.go) | Go | 21 | 11 | 7 | 39 |
| [mqd-client/src/domain/models/api_configuration_settings.go](/mqd-client/src/domain/models/api_configuration_settings.go) | Go | 54 | 1 | 10 | 65 |
| [mqd-client/src/domain/models/configuration_settings.go](/mqd-client/src/domain/models/configuration_settings.go) | Go | 10 | 2 | 3 | 15 |
| [mqd-client/src/domain/models/endpoint_settings.go](/mqd-client/src/domain/models/endpoint_settings.go) | Go | 76 | 23 | 18 | 117 |
| [mqd-client/src/domain/models/report.go](/mqd-client/src/domain/models/report.go) | Go | 60 | 9 | 13 | 82 |
| [mqd-client/src/domain/services/api_dao.go](/mqd-client/src/domain/services/api_dao.go) | Go | 248 | 65 | 50 | 363 |
| [mqd-client/src/domain/services/report_server.go](/mqd-client/src/domain/services/report_server.go) | Go | 7 | 1 | 3 | 11 |
| [mqd-client/src/domain/services/report_server_factory.go](/mqd-client/src/domain/services/report_server_factory.go) | Go | 17 | 6 | 6 | 29 |
| [mqd-client/src/domain/services/report_server_mqd.go](/mqd-client/src/domain/services/report_server_mqd.go) | Go | 94 | 25 | 26 | 145 |
| [mqd-client/src/go.mod](/mqd-client/src/go.mod) | Go Module File | 43 | 0 | 5 | 48 |
| [mqd-client/src/go.sum](/mqd-client/src/go.sum) | Go Checksum File | 221 | 0 | 1 | 222 |
| [mqd-client/src/main.go](/mqd-client/src/main.go) | Go | 68 | 11 | 17 | 96 |
| [mqd-client/src/validation/rule_validator.go](/mqd-client/src/validation/rule_validator.go) | Go | 64 | 40 | 17 | 121 |
| [mqd-client/src/validation/schema_validator.go](/mqd-client/src/validation/schema_validator.go) | Go | 45 | 21 | 11 | 77 |
| [mqd-client/src/validation/validator.go](/mqd-client/src/validation/validator.go) | Go | 8 | 1 | 3 | 12 |
| [mqd-client/tests/Functional-Responses.jmx](/mqd-client/tests/Functional-Responses.jmx) | XML | 5,430 | 0 | 1 | 5,431 |
| [mqd-client/tests/PoC Tests.jmx](/mqd-client/tests/PoC%20Tests.jmx) | XML | 148 | 0 | 1 | 149 |
| [mqd-client/tests/performance.jmx](/mqd-client/tests/performance.jmx) | XML | 253 | 0 | 1 | 254 |
| [mqd-client/tests/performance_limited.jmx](/mqd-client/tests/performance_limited.jmx) | XML | 262 | 0 | 1 | 263 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)