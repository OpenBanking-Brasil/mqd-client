# 🔐 RELATÓRIO FINAL DE SEGURANÇA - MQD-CLIENT

**Data**: 2026-09-24  
**Projeto**: MQD-Client (OpenBanking Brasil)  
**Branch**: fix/TLM-1915  
**Versão Go**: 1.23.0 → 1.27.1 (`src/go.mod` e imagem de build `golang:1.27.1`)  
**Imagem base de runtime**: `alpine:edge` → **`alpine:3.24.2`** (última estável, versão fixa)  
**Ferramenta**: Trivy v0.51.4, base de vulnerabilidades atualizada em 2026-09-24  
**Imagem analisada**: `mqd-client:latest` (`sha256:3d0ee5584e68`), buildada do zero (`--no-cache --pull`)

**Arquivos de evidência (JSON do Trivy):**

| Arquivo | Alvo |
|---|---|
| `systemfile-24-09-2026.json` | Filesystem do projeto inteiro |
| `dockerfile-24-09-2026.json` | Imagem Docker da aplicação |

Os comandos para reproduzir estão em `../run_trivy.md`.

## 📊 RESUMO EXECUTIVO

### Status
- ✅ **Vulnerabilidades**: **0** no projeto e **0** na imagem, em todas as severidades (CRITICAL a
  UNKNOWN).
- ✅ **CVEs HIGH do relatório anterior (14/09)**: as 3 foram corrigidas.
- ✅ **Dependências Go**: todas as usadas pela aplicação estão na última versão (`go get -u` +
  `go mod tidy`).
- ✅ **Imagem base**: trocada de `alpine:edge` (desenvolvimento) para `alpine:3.24.2` (estável).
  O OpenSSL foi de 3.5.7-r0 (vulnerável) para 3.5.8-r0 (corrigido).
- ⚠️ **Chave privada** presente no disco (`infra/dockerfile/certificates/client.key`), fora do git.
- ⚠️ **Configuração**: falta `HEALTHCHECK` (LOW). O mock de QA roda como root (HIGH, só QA local).

---

## 🔍 VULNERABILIDADES

### Resultado atual
| Alvo | CRITICAL | HIGH | MEDIUM | LOW | UNKNOWN |
|---|---|---|---|---|---|
| Dependências Go (`src/go.mod`, 32 pacotes) | 0 | 0 | 0 | 0 | 0 |
| Binário na imagem (stdlib 1.27.1 + 32 pacotes) | 0 | 0 | 0 | 0 | 0 |
| Sistema da imagem (Alpine 3.24.2, 16 pacotes) | 0 | 0 | 0 | 0 | 0 |

### Por que trocar o `alpine:edge`
- **O que é o `edge`:** a versão de **desenvolvimento** do Alpine (hoje `3.25.0_alpha`). Não tem
  suporte de segurança, e a base de CVE do Trivy **não tem dados para ela**.
- **O problema:** com o `edge`, o scan dava "0 vulnerabilidades" no sistema mesmo com o OpenSSL
  **3.5.7-r0**. Essa mesma versão, na Alpine 3.24.1 estável, tem **2 HIGH (ex.: CVE-2026-14456),
  6 MEDIUM e 12 LOW**. Ou seja, era um falso negativo.
- **A troca:** o `dockerfile-minimal` agora usa `FROM docker.io/alpine:3.24.2` e roda
  `apk upgrade --no-cache` para aplicar correções pendentes.

### Sobre o aviso "This OS version is not on the EOL list"
Esse aviso continua aparecendo para a 3.24. Ele vem de uma **tabela de datas de fim de suporte
gravada dentro do binário** do Trivy v0.51.4, que é anterior à Alpine 3.24. Ele **não indica** falta
de análise. Validação feita: com o mesmo aviso, o Trivy encontra as **20 vulnerabilidades** do
OpenSSL 3.5.7-r0 na `alpine:3.24.1`. Portanto o **0 na 3.24.2 é real**. Atualizar o Trivy para uma
versão recente elimina o aviso.

### Histórico: CVEs do relatório de 14/09
| CVE | Pacote | Antes | Agora | Status |
|---|---|---|---|---|
| CVE-2025-30204 (HIGH) | golang-jwt/jwt/v5 | 5.2.1 | 5.3.1 | ✅ Corrigida |
| CVE-2026-24051 (HIGH) | otel/sdk | 1.32.0 | 1.46.0 | ✅ Corrigida |
| CVE-2026-39883 (HIGH) | otel | 1.32.0 | 1.46.0 | ✅ Corrigida |
| CVE-2026-81870 (LOW) | otel/sdk | 1.44.0 | 1.46.0 | ✅ Corrigida (encontrada e corrigida em 24/09) |
| CVE-2026-14456 e mais 9 (2 HIGH, 6 MEDIUM, 12 LOW) | libssl3/libcrypto3 | 3.5.7-r0 | 3.5.8-r0 | ✅ Corrigidas (troca do `alpine:edge`) |

---

## 📦 DEPENDÊNCIAS GO ATUALIZADAS (24/09)
Comando: `go get -u ./...`, `go get -u all` e `go mod tidy`, com Go 1.27.1. Validado com
`go build`, `go vet` e `go test`.

| Módulo | Antes | Depois |
|---|---|---|
| github.com/prometheus/client_golang | 1.23.2 | 1.24.1 |
| github.com/prometheus/client_model | 0.6.2 | 0.6.3 |
| github.com/prometheus/common | 0.67.5 | 0.71.0 |
| github.com/prometheus/procfs | 0.20.1 | 0.22.0 |
| github.com/rs/zerolog | 1.33.0 | 1.35.1 |
| github.com/sethvargo/go-envconfig | 1.1.0 | 1.4.3 |
| go.opentelemetry.io/otel (+ metric, sdk, sdk/metric, trace) | 1.44.0 | 1.46.0 |
| go.opentelemetry.io/otel/exporters/prometheus | 0.66.0 | 0.68.0 |
| github.com/go-logr/logr | 1.4.3 | 1.4.4 |
| github.com/mattn/go-colorable | 0.1.13 | 0.1.15 |
| github.com/mattn/go-isatty | 0.0.20 | 0.0.24 |
| golang.org/x/sys | 0.45.0 | 0.48.0 |
| google.golang.org/protobuf | 1.36.11 | 1.36.12 |

**Módulos que ainda mostram versão mais nova no `go list -m -u all`:** `x/net`, `x/text`,
`x/sync`, `x/oauth2`, `klauspost/compress`, `golang/protobuf`, `creack/pty`, `rogpeppe/go-internal`,
`stretchr/objx` e `xhit/go-str2duration`. Eles **não estão no `go.mod` e não entram no binário**
(conferido com `go list -deps`). Só aparecem porque outras bibliotecas citam essas versões nos
próprios `go.mod`, e o `go mod tidy` os remove. Não há o que atualizar.

**Testes:**
- `go test ./...`: ✅.
- `go test -race ./...`: ❌ no `TestResultProcessorTotalResultsRace`. É uma data race **conhecida e
  anterior** em `result_processor.go` (`totalResults`), documentada no openspec
  `harden-report-reliability-and-fix-race`. Não foi causada pela atualização.

---

## 🔑 SEGREDOS
| Local | Achado | Severidade |
|---|---|---|
| `infra/dockerfile/certificates/client.key` | Chave privada | HIGH |
| Imagem Docker | Nenhum segredo | ✅ |

A chave está no `.gitignore`, então **não é versionada**. Ela existe no disco da máquina de
desenvolvimento, pertence ao root e tem permissão 600. O scan de filesystem é rodado **como root**
justamente para ler esse arquivo: com o usuário comum, o Trivy pula o arquivo sem avisar.

---

## ⚙️ CONFIGURAÇÃO (Dockerfile / imagem)
| Arquivo | Regra | Severidade | Achado |
|---|---|---|---|
| `infra/dockerfile/dockerfile-minimal` | DS026 | LOW | Sem `HEALTHCHECK` |
| imagem `mqd-client:latest` | DS026 | LOW | Sem `HEALTHCHECK` |
| `tools/mqd-server-mock/Dockerfile` (só QA local) | DS002 | HIGH | Container roda como `root` |
| `tools/mqd-server-mock/Dockerfile` (só QA local) | DS026 | LOW | Sem `HEALTHCHECK` |

Pontos positivos do `dockerfile-minimal`:
- ✅ multi-stage;
- ✅ usuário não-root (`mqd_user`);
- ✅ binário estático (`CGO_ENABLED=0`);
- ✅ imagem base estável com versão fixa.

---

## 📜 LICENÇAS
| Origem | Licenças | Observação |
|---|---|---|
| Dependências Go (38) | Apache-2.0 (19), BSD-3-Clause (11), MIT (8) | Todas permissivas |
| Pacotes Alpine | GPL-2.0 (9, "restricted"), MPL-2.0 (1), MIT/Apache/BSD/Zlib | GPL-2.0 vem de busybox, apk-tools, musl-utils etc., o padrão de qualquer imagem Alpine |

---

## 🚀 PRÓXIMOS PASSOS (recomendações)
1. **HEALTHCHECK** no `dockerfile-minimal`:
   ```dockerfile
   HEALTHCHECK --interval=30s --timeout=5s CMD wget -q -O /dev/null http://localhost:8080/metrics || exit 1
   ```
2. **Mock de QA**: adicionar `USER` não-root e `HEALTHCHECK` em `tools/mqd-server-mock/Dockerfile`.
3. **Atualizar o Trivy** instalado (v0.51.4) para uma versão recente. Isso elimina o aviso de EOL e
   habilita `--pkg-types` e `--detection-priority comprehensive`.
4. **Binário duplicado** na imagem (`/usr/server/mqd-client` e `/usr/mqd-client`): remover a cópia
   de `/usr/server`.
5. **Corrigir a data race** do `result_processor.go` (openspec
   `harden-report-reliability-and-fix-race`).

---

**Status**: 🟢 OK: 0 vulnerabilidades no projeto e na imagem; restam só recomendações de configuração  
**Gerado por**: Claude Opus 5.5 + Trivy v0.51.4
