## Context

Ver `proposal.md` para a motivação completa. Pontos de código relevantes para as decisões abaixo:

- `ResultProcessor.processAndSendResults` (`src/application/result_processor.go:174-197`) chama `getAndClearResults()` — que **já limpa** o mapa agregado em memória — e só depois tenta `mqdServer.SendReport(report)`. Em erro, faz `return` no meio do `for`, abandonando os demais `transmitterID`s do lote.
- `ReportServerMQD.postReport` (`src/domain/services/report_server_mqd.go:83-130`) só trata como erro falhas de transporte (`httpClient.Do` retornando `err != nil`); um status HTTP não-200 vira `Logger.Warning` e a função retorna `nil`.
- `RestAPI.executeGet` (`src/domain/services/api_dao.go:131-187`) já implementa um padrão de retry síncrono com `retryTimes` e `time.Sleep(1s)` para as chamadas de configuração (`GET /settings/...`). Esse mesmo padrão hoje não é usado no envio de relatório (`POST /report`).
- `LocalResultManager` (`src/application/local_result_manager.go`) já persiste amostras de mensagens (não os relatórios agregados) em disco, em `./data_logs/<YYYY-MM-DD>/<appID>/<hora>-<api>.json`, com limpeza automática por `DaysToStore`. Esse mecanismo é independente do envio de relatório e continua funcionando mesmo se o relatório agregado falhar — ou seja, os dados brutos de validação já não se perdem hoje; o que se perde é o **relatório-resumo agregado** enviado ao servidor central.
- `totalResults` é uma variável de pacote (`int`) em `result_processor.go`, escrita sob `resultProcessorMutex` em `AppendResult`, mas lida sem lock em `StartResultsProcessor` (`if totalResults >= ...`, linha 160).
- Os singletons (`log.GetLogger`, `services.GetReportServer`, `application.NewConfigurationManager`, `application.GetMessageProcessorWorker`) usam o padrão "check-then-lock" sem segundo check dentro do lock. Hoje são todos instanciados sequencialmente em `main()` antes de qualquer `go func()`, então não há corrida observável na prática — mas o padrão em si não é seguro para concorrência.

## Goals / Non-Goals

**Goals:**
- Definir a abordagem técnica para eliminar a perda silenciosa de relatórios agregados quando o envio ao servidor central falha (HTTP ou transporte), incluindo isolamento por `transmitterID` e um mecanismo de persistência local (outbox) para falhas persistentes.
- Definir a abordagem para eliminar a data race em `totalResults` e endurecer a inicialização dos singletons.
- Estabelecer decisões suficientes para que `tasks.md` possa quebrar o trabalho em passos implementáveis, sem ambiguidade sobre a estratégia escolhida.

**Non-Goals:**
- Não é objetivo desta mudança alterar o comportamento do servidor central MQD nem o que ele faz com o relatório após aceitá-lo — isso está fora do repositório `mqd-client`. O mqd-client não possui nenhuma integração direta com S3.
- Não é objetivo implementar um sistema de outbox genérico e reutilizável para toda a aplicação; o escopo é o fluxo específico de envio do relatório agregado (`POST /report`).
- Não é objetivo mudar o contrato do endpoint `POST /ValidateResponse` consumido pelos clientes do mqd-client.
- Não é objetivo nesta etapa escrever ou alterar código — apenas produzir os artefatos de planejamento (`proposal.md`, `specs/`, `design.md`, `tasks.md`).

## Decisions

### 1. Tratar status HTTP não-2xx como erro real
`postReport` passa a retornar `error` (com o status code e, se disponível, um trecho do corpo da resposta) sempre que `resp.StatusCode` não for 2xx, em vez de apenas logar `Warning` e retornar `nil`. `SendReport` continua propagando esse erro como já faz hoje para erros de transporte.
**Alternativa considerada:** manter o log em `Warning` e adicionar apenas uma métrica separada. Rejeitada porque não resolve o problema central — o chamador (`processAndSendResults`) continua achando que o envio teve sucesso e descarta os dados.

### 2. Outbox real: persistir o relatório em disco antes de tentar o envio, remover só após confirmação
Ao montar o relatório agregado de um `transmitterID` (antes de chamar `SendReport`), o relatório é serializado e gravado em um diretório de pendências local (ex.: `./data_logs/pending_reports/<transmitterID>-<timestampCiclo>.json`). O arquivo só é removido **depois** que `SendReport` confirmar sucesso (2xx). Um processo em background (sweeper), executado periodicamente, varre esse diretório e tenta reenviar qualquer arquivo pendente (inclusive os deixados por uma queda do processo no meio do ciclo), aplicando a mesma política de retry/backoff.
**Alternativa considerada (persistir só depois de esgotar os retries em memória):** mais barata em I/O no caminho feliz, mas não protege contra um crash do processo *durante* a janela de retry em memória — o relatório ainda pode se perder nesse intervalo. Rejeitada porque o objetivo explícito é fechar o gap "sistema disse OK, mas não persistiu", e isso inclui o caso de crash do processo, não só de erro HTTP.
**Alternativa considerada (fila de retry só em memória, sem disco):** mais simples, mas reintroduz perda de dados em caso de reinício/deploy do mqd-client antes do retry ter sucesso. Rejeitada pelo mesmo motivo acima.

### 3. Retry com backoff limitado antes de depender do sweeper
`SendReport` passa a ser chamado através de uma pequena política de retry (número máximo de tentativas + backoff, seguindo o padrão já existente em `executeGet`), para absorver falhas transitórias (timeout pontual, 5xx passageiro) sem depender do ciclo do sweeper. Se todas as tentativas falharem, o arquivo permanece no outbox para o sweeper tentar depois.
**Alternativa considerada:** delegar 100% do retry ao sweeper (sem retry síncrono no ciclo principal). Rejeitada porque atrasaria desnecessariamente a recuperação de falhas transitórias curtas, que hoje já são tratadas com sucesso em `executeGet` para outras chamadas.

### 4. Isolar falha por `transmitterID`
O laço em `processAndSendResults` passa a usar `continue` (não `return`) quando o envio de um `transmitterID` falha, registrando o erro e seguindo para o próximo. Um erro agregado (lista de `transmitterID`s que falharam) é logado ao final do ciclo para visibilidade operacional.

### 5. `totalResults` como contador atômico
`totalResults` passa a ser um `atomic.Int64` (pacote `sync/atomic`), com `Add`/`Load`/`Store` substituindo os acessos diretos. O mapa `txGroupedResults` continua protegido por `resultProcessorMutex`, já que operações em mapas não são atômicas por si só.
**Alternativa considerada:** manter `int` e garantir que toda leitura também passe pelo mutex (incluir a checagem do `select` dentro de uma função helper com lock). Funciona, mas adiciona contenção de lock em um caminho de verificação frequente (a cada 5s); `atomic.Int64` é mais simples e idiomático para um contador escalar isolado.

### 6. `sync.Once` para os singletons
`log.GetLogger`, `services.GetReportServer`, `application.NewConfigurationManager` e `application.GetMessageProcessorWorker` passam a usar `sync.Once` para garantir inicialização única e segura sob concorrência, substituindo o padrão atual de "check-then-lock" sem segundo check.
**Alternativa considerada:** manter o padrão atual, já que hoje não há chamada concorrente real (tudo instanciado sequencialmente em `main()`). Rejeitada como decisão de design porque o custo de migrar para `sync.Once` é baixo e remove uma classe inteira de bug latente caso o padrão de chamada mude no futuro (ex.: em testes que chamam esses `Get*` em paralelo).

## Risks / Trade-offs

- [Passar a tratar status não-2xx como erro pode gerar um volume de logs/alertas de erro que hoje não existia, já que esses casos eram só `Warning`] → Coordenar com quem observa os logs/alertas antes do rollout; considerar isso uma correção esperada (o silêncio anterior é o próprio bug), não um efeito colateral a esconder.
- [Escrever o relatório em disco antes de cada tentativa de envio adiciona I/O e um novo diretório para gerenciar (`pending_reports`)] → Reaproveitar o padrão de retenção/limpeza já existente em `LocalResultManager` (`cleanupFiles`, baseado em `DaysToStore`) para evitar crescimento ilimitado do outbox.
- [Retry síncrono com backoff no ciclo principal pode atrasar o próximo ciclo de agregação se o servidor central estiver lento/instável por muito tempo] → Limitar o número de tentativas síncronas (ex.: as mesmas 3 tentativas já usadas em `executeGet`) e delegar o restante ao sweeper assíncrono, que roda em sua própria goroutine/ticker.
- [Migrar `totalResults` para `atomic.Int64` exige revisar todos os pontos que hoje leem/escrevem essa variável como `int` simples, incluindo o reset em `getAndClearResults`] → Mapear explicitamente todos os usos como uma task própria antes de tocar no tipo, para não deixar um acesso não convertido (o que reintroduziria a race).
- [`sync.Once` para singletons é uma mudança mecânica em 4 arquivos de bootstrap] → Baixo risco, mas tocar em código de inicialização compartilhado merece verificação manual de que a assinatura pública de cada `Get*`/`New*` não muda (para não quebrar `main.go` nem eventuais testes existentes).

## Migration Plan

Mudança de comportamento dentro do binário do `mqd-client`, sem migração de dados/schema. Passos de rollout (a detalhar em `tasks.md`, sem execução nesta etapa):
1. Implementar as correções de forma incremental (uma decisão por vez), cada uma coberta por teste (incluindo execução com `-race` para as mudanças de concorrência).
2. Adicionar o novo diretório de outbox (`./data_logs/pending_reports/` ou equivalente) à documentação/infra de deploy, se ele precisar de um volume persistente (hoje `./data_logs` já é usado por `LocalResultManager`, então herda a mesma configuração de volume, se houver).
3. Deploy como uma nova versão do `mqd-client`; rollback é reverter para a imagem/versão anterior — não há estado persistente novo que exija migração reversa (arquivos de outbox remanescentes podem ser simplesmente descartados ou reprocessados manualmente).

## Open Questions

- Números exatos da política de retry (quantidade de tentativas síncronas, tempos de backoff) e do intervalo do sweeper do outbox — não bloqueiam a especificação nem a abordagem escolhida; podem ser definidos durante a implementação (task correspondente), reaproveitando por padrão os mesmos valores já usados em `executeGet` (3 tentativas, 1s) como ponto de partida.
- Caminho/nome definitivo do diretório de outbox e se ele deve ser configurável via `settings.yml` (como `basePath` de `LocalResultManager` hoje é fixo em `./data_logs`) — decisão de implementação, não afeta os requisitos definidos em `specs/`.
