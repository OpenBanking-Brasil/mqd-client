# mqd-server-mock

Substituto local para o servidor central do MQD com o qual o `mqd-client` (veja [`../../src`](../../src)) se comunica via `PROXY_URL`. Ele implementa apenas o suficiente da API real para que o `mqd-client` consiga iniciar e rodar localmente, sem precisar de acesso de rede/VPN ao servidor central real ou de certificados ICP-Brasil.

**Apenas para desenvolvimento local.** Os dados de fixture (regras de validação, taxas) são ilustrativos e não representam as regras de validação reais do OpenFinance em produção.

## Endpoints

| Método | Caminho | Finalidade |
|---|---|---|
| `GET` | `/settings/{fileName}` | Serve um arquivo do diretório de fixtures como JSON. O `mqd-client` solicita `configurationSettings.json` na inicialização. |
| `POST` | `/token` | Retorna um JWT falso estático (`access_token`, `expires_in`, ...). Sem assinatura/validação real. |
| `POST` | `/report` | Aceita um corpo de relatório, registra um resumo de uma linha no stdout e retorna `200`. |

## Como executar

```bash
cd tools/mqd-server-mock
go run .
```

Por padrão, escuta na porta `:8082`, correspondendo ao `PROXY_URL=http://127.0.0.1:8082` já configurado por padrão em [`src/settings/settings.yml`](../../src/settings/settings.yml).

### Sobrescrever porta / fixtures

```bash
go run . -port 9000 -fixtures ./my-fixtures
```

ou via variáveis de ambiente: `PORT`, `FIXTURES_DIR`. As flags têm precedência caso ambas sejam definidas.

## Apontar o mqd-client para o mock

A partir de `src/`, com o mock já em execução em `:8082`:

**Linux / macOS (bash/zsh):**

```bash
cd ../../src
PROXY_URL=http://127.0.0.1:8082 SERVER_ORG_ID=d7384bd0-842f-43c5-be02-9d2b2d5efc2c APPLICATION_MODE=TRANSMITTER go run .
```

**Windows (PowerShell):**

```powershell
cd ..\..\src
$env:PROXY_URL = "http://127.0.0.1:8082"
$env:SERVER_ORG_ID = "d7384bd0-842f-43c5-be02-9d2b2d5efc2c"
$env:APPLICATION_MODE = "TRANSMITTER"
go run .
```

No PowerShell, a sintaxe `VAR=valor comando` do bash não funciona — cada variável precisa ser definida separadamente com `$env:VAR = "valor"` antes de rodar `go run .`. Essas variáveis valem apenas para a sessão atual do terminal. Alternativamente, use `docker-compose` ou um arquivo `.env` em vez de variáveis inline.

## Editando as fixtures

`fixtures/configurationSettings.json` corresponde ao formato esperado pelo `mqd-client` (`src/domain/models.ConfigurationSettings`). Edite-o diretamente para alterar taxas de validação, ou inicie o mock com `-fixtures <dir>` apontando para sua própria cópia.

A fixture padrão vem propositalmente com `ValidationSettings.APIGroupSettings` vazio: o `mqd-client` busca um arquivo extra por entrada de API ali (`GET /settings/{basePath}/{apiBasePath}/{apiVersion}/response/endpoints.json`), então mantê-lo vazio é o que permite que o `mqd-client` inicie corretamente sem nenhuma configuração adicional. Para exercitar a validação real de endpoints localmente, adicione uma entrada em `APIGroupSettings` e coloque um `endpoints.json` correspondente (um array de `models.APIEndpointSetting`) no caminho derivado dentro do seu diretório de fixtures.
