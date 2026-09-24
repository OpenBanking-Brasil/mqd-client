# Análise do código mqd-client — Race Condition e Confiabilidade no Envio de Relatórios ao Servidor Central

**Status:** Diagnóstico apenas. Nenhuma linha de código de produção foi alterada. As correções propostas foram formalizadas como especificação em `openspec/changes/harden-report-reliability-and-fix-race/` (proposal, specs, design e tasks), aguardando aprovação antes de qualquer implementação.

**Escopo analisado:** `src/` (módulo Go do mqd-client), ~3.450 linhas.

**Correção de escopo (revisão desta análise):** a investigação original partiu de um caso de suporte em que um cliente teria enviado um relatório que não apareceu no S3. Ao ler o código, confirmamos que **o mqd-client não possui nenhuma integração com S3** — não há SDK da AWS, nem referência a bucket, nem qualquer chamada relacionada a S3 em lugar nenhum do repositório. O mqd-client só faz uma chamada HTTP (`POST /report`) para um servidor central; o que acontece depois disso (incluindo qualquer gravação em S3) é responsabilidade de outro sistema, fora deste repositório e fora do escopo desta análise. **Não é necessária uma nova análise** — os problemas técnicos encontrados no código do mqd-client (abaixo) continuam válidos e reais; o que muda é apenas a motivação/contexto: eles não devem ser apresentados como "a causa do relatório sumir no S3", e sim como bugs de confiabilidade do próprio mqd-client no trecho que ele de fato controla (do recebimento da mensagem até o envio HTTP ao servidor central).

---

## 1. Resumo executivo

Este documento responde a duas perguntas:
1. Existe race condition no código Go do mqd-client?
2. O envio do relatório agregado ao servidor central é confiável (o mqd-client trata corretamente falhas de entrega)?

**Respostas curtas:**
1. **Sim, existe uma data race real e verificável** no contador que decide quando disparar o envio de relatório (`totalResults`), além de um padrão de inicialização de singleton frágil (não é uma race hoje, mas é uma bomba-relógio).
2. **Não.** Existe um bug concreto que faz o mqd-client **acreditar que o envio ao servidor central deu certo mesmo quando a resposta foi de erro**, e o relatório agregado é apagado da memória antes de confirmar que o envio funcionou — sem retry, sem persistência local (sem padrão outbox). Uma falha de rede ou um erro do servidor central faz o mqd-client perder o relatório daquele ciclo sem perceber e sem alertar ninguém.

### Tabela de prioridade

| # | Problema | Onde | Severidade | Por quê |
|---|----------|------|------------|---------|
| 1 | `postReport` trata erro HTTP do servidor central como sucesso | `report_server_mqd.go` | **CRÍTICO** | O mqd-client segue achando que entregou o relatório mesmo quando o servidor central respondeu com erro |
| 2 | Resultados agregados são apagados da memória antes de confirmar o envio (sem outbox/retry) | `result_processor.go` | **CRÍTICO** | Qualquer falha de rede/timeout descarta o relatório inteiro do ciclo, para sempre |
| 3 | Falha em 1 transmissor aborta o envio dos demais do mesmo ciclo | `result_processor.go` | **ALTO** | Um cliente com problema "derruba" os relatórios de outros clientes no mesmo lote |
| 4 | Data race real em `totalResults` | `result_processor.go` | **ALTO** | Comportamento indefinido pela memory model do Go; `go test -race` acusa (evidência já capturada) |
| 5 | `Metrics.Values` acumula duplicado entre transmissores no mesmo ciclo | `result_processor.go` | **MÉDIO** | Relatório enviado ao servidor central fica com dados de métricas incorretos/duplicados a partir do 2º transmissor do lote |
| 6 | "OK" da API HTTP do mqd-client só significa "enfileirado", não "processado/entregue" | `api_server.go` | **MÉDIO** | Gera falsa sensação de garantia de entrega ponta a ponta para quem integra com o mqd-client |
| 7 | Padrão de inicialização de singleton inseguro (check-then-lock incompleto) | 4 arquivos (ver seção 4) | **BAIXO** | Não é race hoje (tudo instanciado sequencialmente no `main()`), mas é frágil a mudanças futuras |
| 8 | Fila em memória sem persistência (`messageQueue`, buffer 1000) | `queue_manager.go` | **BAIXO/INFORMATIVO** | Se o processo cair, mensagens em trânsito no buffer se perdem; comportamento esperado de fila em memória, mas vale documentar o limite |

---

## 2. Confiabilidade do envio do relatório agregado ao servidor central

### 2.1 O caminho completo de um relatório dentro do mqd-client

```
Cliente do mqd-client (sistema do banco/instituição)
        │  POST /ValidateResponse
        ▼
handleValidateResponseMessage (api_server.go)
        │  responde "Message enqueued for processing!" AQUI — antes de qualquer validação real
        ▼
messageQueue (channel em memória, buffer 1000)  ← queue_manager.go
        │
        ▼
worker() → processMessage() (message_process_Worker.go)
        │  valida o payload, gera MessageResult
        ▼
ResultProcessor.AppendResult (result_processor.go)
        │  agrega o resultado em memória (txGroupedResults)
        ▼
   [a cada N minutos OU quando atinge X resultados]
        ▼
ResultProcessor.processAndSendResults()
        │  1. getAndClearResults() → JÁ APAGA o estado agregado da memória
        │  2. monta o "report" (resumo agregado)
        │  3. SendReport(report) → HTTP POST /report para o servidor central
        ▼
ReportServerMQD.postReport (report_server_mqd.go)
        │  se status == 200 → OK de verdade
        │  se status != 200 → **loga Warning e retorna nil (sucesso!)**
        ▼
Servidor central (fora deste repositório — o que acontece a partir daqui está fora do escopo do mqd-client e desta análise)
```

### 2.2 O "OK" da API HTTP só garante "recebi e enfileirei"

Em `src/application/api_server.go`, função `handleValidateResponseMessage`:

```go
as.qm.EnqueueMessage(&msg)
...
_, err = fmt.Fprintf(w, "Message enqueued for processing!")
```

Esse "OK" acontece **antes** de: validar o payload de verdade, agregar no relatório, e — principalmente — antes de qualquer tentativa de enviar esse dado ao servidor central. Se o processo cair entre esse ponto e o envio do relatório periódico, o dado nunca chega a lugar nenhum, e quem chamou já recebeu "OK" há muito tempo.

**Isso não é necessariamente um bug** — é assim que filas assíncronas funcionam. O problema é que não existe nenhuma camada de confirmação posterior ("seu relatório X foi de fato entregue ao servidor central às HH:MM") nem alerta em caso de falha — o que nos leva ao próximo ponto, que é o mais grave.

### 2.3 [CRÍTICO] `postReport` finge sucesso quando o servidor central responde com erro

Arquivo: `src/domain/services/report_server_mqd.go`, linhas 116–129:

```go
// Check the response status code
if resp.StatusCode != http.StatusOK {
    rs.Logger.Warning("Error sending report, Status code: "+fmt.Sprint(resp.StatusCode), rs.Pack, "postReport")
} else {
    // Read the body of the message
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    rs.Logger.Info(string(body), rs.Pack, "postReport")
}

return nil   // <-- SEMPRE retorna nil aqui, mesmo quando caiu no "if" de erro acima
```

Se o servidor central responder `500`, `401` (token expirado), `400` (payload malformado por algum motivo), ou qualquer coisa diferente de `200`:
- A função só loga em nível **Warning** (não Error — normalmente não dispara alerta).
- Retorna `nil`, ou seja, **sem erro**.
- Quem chamou (`SendReport` → `processAndSendResults`) entende que o envio foi um sucesso completo.
- O relatório já tinha sido apagado da memória (ver próximo item) — resultado: **perda silenciosa e sem qualquer sinalização de erro** dentro do mqd-client.

Esse é o bug mais importante encontrado nesta análise: o mqd-client não tem como saber (nem avisar ninguém) que um envio de relatório falhou de verdade.

### 2.4 [CRÍTICO] Não existe outbox: o estado é limpo antes de confirmar a entrega

Arquivo: `src/application/result_processor.go`, função `processAndSendResults` (linhas 174–197):

```go
results := rp.getAndClearResults()   // já ZERA txGroupedResults, ANTES de saber se o envio vai funcionar
...
for _, transmitterResult := range results {
    ...
    err := rp.mqdServer.SendReport(report)
    if err != nil {
        rp.Logger.Error(err, "Error sending report", rp.Pack, "processAndSendResults")
        return   // aborta tudo, ver item 2.5
    }
    ...
}
```

**O que é o problema de "dual write" / por que outbox resolveria isso:**
Dual write é quando um sistema precisa fazer duas coisas que deveriam ser atômicas (ex.: "marcar como processado" + "entregar/persistir de verdade externamente"), mas as faz como duas operações separadas e não transacionais. Se a segunda falhar depois que a primeira já aconteceu, você fica com um estado inconsistente: o sistema *acha* que terminou, mas o efeito externo nunca aconteceu.

É exatamente isso que acontece aqui:
1. `getAndClearResults()` já "marca como processado" (limpa a memória) — ação 1.
2. `SendReport()` tenta a entrega de verdade ao servidor central — ação 2.

Se a ação 2 falhar (rede caiu, timeout, servidor central fora do ar, ou o bug do item 2.3 mascarando um erro), a ação 1 já aconteceu e **não tem volta**: não há uma cópia em disco do relatório, não há fila de retry, não há nada. O padrão **outbox** resolveria isso assim: gravar o relatório em um local durável (arquivo local, tabela, fila) **antes** de tentar o envio, e só apagar essa cópia **depois** de confirmar que o envio deu certo — com um processo de repescagem (sweeper) tentando reenviar o que ficou pendente, inclusive depois de um crash do processo. Hoje isso **não existe em nenhum lugar do mqd-client**.

### 2.5 [ALTO] Falha em um cliente derruba os relatórios de todos os outros do mesmo ciclo

Continuando o mesmo trecho acima: o `return` dentro do `for` faz o método sair **completamente** assim que o envio de **um** `transmitterID` falha. Se naquele ciclo havia resultados agregados de 3 clientes diferentes e o envio do primeiro falhar, os outros 2 — que poderiam ter sido enviados com sucesso — são simplesmente abandonados (e já foram removidos da memória pelo `getAndClearResults`). Ou seja, um problema pontual com um cliente pode causar perda de dados de relatório de clientes completamente não relacionados.

---

## 3. Race condition confirmada

### 3.1 O que é uma race condition, rapidamente

É quando duas ou mais goroutines acessam a mesma variável ao mesmo tempo, pelo menos uma delas escrevendo, sem nenhuma sincronização (mutex, canal, atomic) garantindo uma ordem. O resultado é "comportamento indefinido": pode funcionar 99% das vezes e falhar de forma imprevisível sob carga — e o compilador/runtime do Go tem liberdade para otimizar de formas que pioram isso (não é só "ler um valor levemente desatualizado").

### 3.2 O caso concreto: `totalResults`

Arquivo: `src/application/result_processor.go`.

**Escrita (com lock)** — dentro de `AppendResult`, chamada pela goroutine do *worker* (a que processa mensagens da fila):
```go
func (rp *ResultProcessor) AppendResult(result *MessageResult) {
    resultProcessorMutex.Lock()
    totalResults++
    ...
    resultProcessorMutex.Unlock()
}
```

**Leitura (SEM lock)** — dentro de `StartResultsProcessor`, que roda em outra goroutine (o próprio ciclo de geração de relatório):
```go
case <-time.After(5 * time.Second):
    if totalResults >= rp.cm.GetSendOnReportNumber() {   // <-- lê sem qualquer mutex
        rp.processAndSendResults()
        ...
    }
```

Duas goroutines diferentes (`go mp.StartWorker()` e `go rp.StartResultsProcessor()`, ambas iniciadas em `main.go`) tocam a mesma variável de pacote, uma protegida por mutex e outra não. Isso é uma data race real e demonstrável — rodando com `go test -race` (ou `go run -race`) sobre um cenário que exercite os dois caminhos concorrentemente, o Go acusa o problema.

**Evidência já capturada:** foi criado um teste isolado (`src/application/result_processor_race_test.go`, não altera nenhum código de produção) que reproduz exatamente esse acesso concorrente. Rodando:
```
go test -race -run TestResultProcessorTotalResultsRace ./application/... -v
```
o Go acusa `WARNING: DATA RACE`, apontando a escrita em `result_processor.go:95` (dentro de `AppendResult`) colidindo com a leitura desprotegida.

**Efeito prático hoje:** provavelmente baixo (na pior das hipóteses, o disparo do relatório acontece um pouco cedo ou atrasado por causa de um valor "meio atualizado"), mas é undefined behavior — não há garantia de que vai se manter "só isso" para sempre, especialmente se o código mudar.

**Correção sugerida (documentada em `design.md` do change):** trocar `totalResults` por `atomic.Int64` (mais simples e sem custo de lock para um contador escalar) ou garantir que toda leitura também passe pelo mesmo mutex.

### 3.3 [BAIXO, latente] Inicialização de singleton sem dupla checagem

Padrão repetido em 4 lugares: `crosscutting/log/logger_factory.go` (`GetLogger`), `domain/services/report_server_factory.go` (`GetReportServer`), `application/message_process_Worker.go` (`GetMessageProcessorWorker`), `application/configuration_manager.go` (`NewConfigurationManager`), e também `application/result_processor.go` (`GetResultProcessor`):

```go
func GetLogger(loggingLevel string) Logger {
    if singleton == nil {          // checagem SEM lock
        lock.Lock()
        defer lock.Unlock()
        singleton = GetNewJSONLogger()   // sem checar de novo se outra goroutine já criou
        ...
    }
    return singleton
}
```

Isso é "double-checked locking" **incompleto** — falta o segundo `if singleton == nil` **dentro** do lock. Hoje, na prática, **não é uma race ativa**, porque todos esses `Get*`/`New*` são chamados sequencialmente dentro de `main()`, antes de qualquer `go func()` ser disparada. Mas é um padrão frágil: se um dia esse código passar a ser chamado a partir de mais de uma goroutine (por exemplo, em testes paralelos, ou se a inicialização for movida para dentro de um handler), vira uma race de verdade, com o agravante de poder criar **duas instâncias diferentes** do "singleton" sendo usadas por partes diferentes do sistema. Correção recomendada: usar `sync.Once`.

---

## 4. Achados secundários (menor urgência, mas vale registrar)

### 4.1 [MÉDIO] `Metrics.Values` duplicado entre transmissores do mesmo ciclo

Em `processAndSendResults`, o struct `report` é criado **uma vez**, fora do laço, e reutilizado para todos os `transmitterID`s do ciclo:

```go
report := models.Report{DataOwnerID: rp.cm.settings.ApplicationSettings.OrganisationID}
rp.updateMetrics(&report)
...
for _, transmitterResult := range results {
    report.ClientID = transmitterResult.TransmitterID
    report.ServerSummary = rp.getSummary(transmitterResult.GroupedResults)   // sobrescrito a cada iteração (ok)
    report.Metrics.Values = append(report.Metrics.Values, models.MetricObject{...})  // ACUMULA a cada iteração (não ok)
    err := rp.mqdServer.SendReport(report)
    ...
}
```

`ServerSummary` é reatribuído a cada iteração (correto). Mas `Metrics.Values` é um `append` — a cada `transmitterID` do mesmo ciclo, o relatório enviado carrega as métricas de **todos** os transmissores já processados naquele ciclo, duplicadas. Se houver 3 clientes no mesmo lote, o relatório do 3º cliente vai com 3x a entrada `runtime.ReportGenerationTime` (uma delas do relatório de outro cliente). É um bug de qualidade de dados enviado ao servidor central, não de perda de dados.

### 4.2 [MÉDIO] O "OK" da API não é uma garantia de ponta a ponta

Já comentado na seção 2.2 — reforçando aqui como item de checklist: o contrato do endpoint `POST /ValidateResponse` hoje é "recebi e enfileirei", não "processei e confirmei entrega ao servidor central". Se algum consumidor do mqd-client (ou algum time interno) está interpretando esse "OK" como uma garantia de entrega ponta a ponta, esse é um desalinhamento de expectativa que vale esclarecer independentemente das correções de código.

### 4.3 [BAIXO/INFORMATIVO] Fila em memória sem persistência

`queue_manager.go`: `messageQueue = make(chan *Message, 1000)`. É um buffer em memória. Se o processo for reiniciado/atualizado (deploy) ou crashar com mensagens ainda no buffer, elas se perdem. Isso é esperado para uma fila em memória simples, mas é outra manifestação do mesmo tema geral (falta de durabilidade em pontos do pipeline) — vale ter em mente ao dimensionar o impacto de reinícios/deploys frequentes.

---

## 5. O que fazer a seguir

1. **Não foi alterado nenhum código de produção nesta etapa.** Toda a investigação foi formalizada como um change do OpenSpec:
   `openspec/changes/harden-report-reliability-and-fix-race/`
   - `proposal.md` — por que mudar
   - `specs/report-delivery-reliability/spec.md` — requisitos de confiabilidade de entrega (itens 2.3, 2.4, 2.5)
   - `specs/result-aggregation-concurrency/spec.md` — requisitos de thread-safety (itens 3.2, 3.3)
   - `design.md` — decisões técnicas de como resolver cada ponto (inclui a proposta de outbox real: gravar o relatório em disco antes de tentar enviar, remover só após confirmação, com um sweeper para reprocessar pendências)
   - `tasks.md` — checklist de implementação, ainda **não executado** (nenhum item marcado)
   Validado com `openspec validate harden-report-reliability-and-fix-race --strict` → OK.

2. **Ordem de prioridade recomendada para quando formos implementar** (da `tasks.md`):
   1. Item 2.3 (tratar status != 200 como erro) — menor esforço, maior impacto na confiabilidade.
   2. Item 2.4/2.5 (outbox + isolar falha por transmissor) — resolve a perda de dados de forma estrutural.
   3. Item 3.2 (fix da race em `totalResults`) — baixo risco, alto valor de correção.
   4. Itens 4.1 e 3.3 — cleanup de qualidade e robustez, sem urgência de incidente.

3. **Sobre o S3:** não é necessário investigar mais nada relacionado a S3 dentro deste repositório — confirmado que o mqd-client não possui nenhuma integração com S3. Qualquer investigação de um caso envolvendo S3 deve ser conduzida no sistema que efetivamente grava lá, fora deste escopo.
