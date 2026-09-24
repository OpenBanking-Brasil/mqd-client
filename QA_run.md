# QA local — mqd-client

Roteiro para rodar o `mqd-client` localmente, em Docker, contra um **mock do servidor central do
MQD** (`tools/mqd-server-mock`). Não precisa de VPN, certificados ICP-Brasil nem acesso ao
sandbox real.

> 💡 Todos os scripts existem em duas versões: `.sh` (WSL2/Linux) e `.ps1` (Windows, PowerShell
> com Docker Desktop). No Windows troque `./scripts/qa-*.sh` por `./scripts/qa-*.ps1` (e
> `--wait-report` por `-WaitReport`).

## Não depende do `mqd-qa-infra`

Diferente das Lambdas `mqd-*`, o `mqd-client` **não usa Postgres, WireMock nem LocalStack**. Ele é
a aplicação instalada na transmissora e só conversa com o servidor central, via `PROXY_URL`. Por
isso o `docker-compose.yml` da raiz sobe só dois serviços, numa rede própria deste projeto:

| Serviço | Container | Porta (host) | O que é |
|---|---|---|---|
| `mqd-server-mock` | `mqd-client-server-mock` | `8082` | Mock do servidor central: `GET /settings/*`, `POST /token`, `POST /report` |
| `app` | `mqd-client-app` | `8080` | O `mqd-client`, buildado com `infra/dockerfile/dockerfile-minimal` e com `PROXY_URL=http://mqd-server-mock:8082` |

Para trocar as portas, use as variáveis `MQD_CLIENT_PORT` / `MQD_SERVER_MOCK_PORT`.

O compose de **instalação** (`infra/dockerfile/docker-compose.yaml`, com proxy nginx e
certificados apontando para o sandbox real) continua separado e não é usado aqui.

---

## Passo 1 — Subir

```bash
./scripts/qa-up.sh               # sobe + smoke test
./scripts/qa-up.sh --wait-report # idem + espera o 1º relatório chegar no mock (~1 min)
```

```powershell
./scripts/qa-up.ps1
./scripts/qa-up.ps1 -WaitReport
```

O smoke test confere três coisas, mais uma opcional:
1. Na inicialização, o `mqd-client` baixou o `configurationSettings.json` do mock.
2. `GET http://localhost:8080/metrics` responde 200.
3. `POST /ValidateResponse` responde. O esperado é **HTTP 400** (`endpointName: Not found`),
   porque a fixture padrão do mock não tem nenhum endpoint configurado (`APIGroupSettings`
   vazio). Esse 400 prova que a API recebeu a mensagem, leu os headers e consultou as
   configurações.
4. Com `--wait-report`, confere também que o `POST /report` chegou no mock. A janela de envio está
   em 1 minuto (`REPORT_EXECUTION_WINDOW=1` no `.env`).

Para validar payloads de verdade, adicione endpoints na fixture do mock (ver
`tools/mqd-server-mock/README.md`, seção "Editando as fixtures") e rode o `qa-reset`.

## Passo 2 — Acompanhar

```bash
docker compose logs -f app mqd-server-mock
```

## Passo 3 — Parar / retomar / resetar

```bash
./scripts/qa-stop.sh    # para os containers deste projeto
./scripts/qa-up.sh      # retoma (rebuilda se o código mudou)
./scripts/qa-reset.sh   # remove os containers deste projeto e sobe tudo de novo
```

```powershell
./scripts/qa-stop.ps1
./scripts/qa-up.ps1
./scripts/qa-reset.ps1
```

Este projeto não tem banco nem volumes, então não há dados para limpar. Nada aqui afeta o
`mqd-qa-infra` nem outros projetos.

---

## Se algo der errado

| O que você viu | O que fazer |
|---|---|
| Docker não está rodando | Abra o Docker Desktop, espere ele iniciar e rode o `qa-up` de novo |
| `port is already allocated` (8080/8082) | Outra coisa usa a porta. Rode com outra, ex.: `MQD_CLIENT_PORT=8090 ./scripts/qa-up.sh` (PowerShell: `$env:MQD_CLIENT_PORT = "8090"`) |
| `mqd-client não pediu GET /settings/...` | Veja `docker compose logs app`: o client não conseguiu falar com o mock na inicialização |
| `POST /ValidateResponse` diferente de 400/200 | Veja `docker compose logs app` |
| Nenhum `POST /report` chegou | Confira `REPORT_EXECUTION_WINDOW` no `.env` (1-60 minutos) e os logs do `app` |
