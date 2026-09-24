## 1. Corrigir a data race em `totalResults` e os singletons

- [ ] 1.1 Converter `totalResults` (`src/application/result_processor.go`) de `int` para `atomic.Int64`, atualizando todos os pontos de leitura e escrita (`AppendResult`, `getAndClearResults`, e a checagem no `select` de `StartResultsProcessor`); verificar com `go build ./...`
- [ ] 1.2 Adicionar/ajustar teste que dispara `AppendResult` concorrentemente a partir de uma goroutine enquanto outra faz a checagem de limiar (simulando o cenário real worker vs. ticker); verificar rodando `go test -race ./application/...` sem nenhuma race reportada
- [ ] 1.3 Substituir o padrão "check-then-lock" por `sync.Once` em `log.GetLogger`, `services.GetReportServer`, `application.NewConfigurationManager`, `application.GetMessageProcessorWorker` e `application.GetResultProcessor`; verificar que `main.go` continua compilando sem alteração de assinatura pública (`go build ./...`)
- [ ] 1.4 Adicionar teste que chama cada `Get*`/`New*` do item 1.3 concorrentemente (múltiplas goroutines) e confirma que todas recebem o mesmo ponteiro; verificar com `go test -race ./...` sem race reportada

## 2. Tratar falhas de entrega do relatório como falhas reais

- [ ] 2.1 Alterar `ReportServerMQD.postReport` (`src/domain/services/report_server_mqd.go`) para retornar `error` quando `resp.StatusCode` não for 2xx (incluindo o status code na mensagem de erro), em vez de apenas logar `Warning` e retornar `nil`; verificar com teste usando `httptest.Server` retornando 500/401 e asserindo que `SendReport` retorna erro não nulo
- [ ] 2.2 Refatorar o laço em `ResultProcessor.processAndSendResults` para usar `continue` em vez de `return` quando o envio de um `transmitterID` falha, registrando o erro e seguindo para o próximo; verificar com teste que usa um `ReportServer` mockado que falha para um `transmitterID` específico e confirma que `SendReport` é chamado para os demais
- [ ] 2.3 Adicionar retry síncrono com backoff limitado ao redor da chamada de `SendReport` dentro do ciclo de envio (reaproveitando o padrão já usado em `RestAPI.executeGet`); verificar com teste que simula falha transitória seguida de sucesso e confirma que o relatório é enviado dentro do número máximo de tentativas configurado

## 3. Implementar outbox local para relatórios não confirmados

- [ ] 3.1 Antes de cada tentativa de envio, serializar o relatório agregado em um arquivo no diretório de pendências local (ex.: `./data_logs/pending_reports/<transmitterID>-<timestamp>.json`); verificar com teste que confirma a criação do arquivo antes da chamada de envio
- [ ] 3.2 Remover o arquivo de pendência somente após confirmação de sucesso (2xx) do envio; verificar com dois testes — um em que o envio funciona (arquivo é removido) e outro em que falha (arquivo permanece)
- [ ] 3.3 Implementar um processo de varredura periódica (sweeper) que lê o diretório de pendências, tenta reenviar cada relatório encontrado e remove o arquivo em caso de sucesso; verificar com teste que cria um arquivo de pendência "antigo", executa um ciclo do sweeper contra um servidor mockado e confirma envio + remoção
- [ ] 3.4 Adicionar limpeza/retenção para arquivos de pendência muito antigos, reaproveitando o padrão de `DaysToStore`/`cleanupFiles` já existente em `LocalResultManager`; verificar com teste que um arquivo além do período de retenção é removido sem ser reenviado
- [ ] 3.5 Iniciar o sweeper no bootstrap da aplicação (`main.go`, seguindo o padrão dos demais `go xxx.StartXxx()`); verificar manualmente que o log de início do sweeper aparece na inicialização

## 4. Observabilidade das falhas de entrega

- [ ] 4.1 Garantir que toda falha de entrega (status não-2xx, erro de transporte, ou esgotamento de retries) seja logada em nível de erro com identificação do `transmitterID`/ciclo do relatório; verificar com teste que confirma a chamada de `Logger.Error` (ou o conteúdo do log) nesses três caminhos de falha

## 5. Verificação de regressão

- [ ] 5.1 Rodar a suíte de testes completa com `go build -race ./...` e `go test -race ./...` a partir de `src/`, confirmando ausência de races e nenhuma regressão nos testes existentes
- [ ] 5.2 Verificar manualmente o fluxo ponta a ponta usando `tools/mqd-server-mock`: configurar o mock para responder `500` em `POST /report`, confirmar que o mqd-client tenta novamente, grava o relatório em `pending_reports`, que os demais `transmitterID`s do lote continuam sendo entregues, e que o relatório pendente é reenviado com sucesso assim que o mock voltar a responder `200`
