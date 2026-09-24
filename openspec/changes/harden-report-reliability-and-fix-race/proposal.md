## Why

Uma análise do caminho que o mqd-client percorre até entregar o relatório agregado ao servidor central encontrou problemas de confiabilidade reais, que merecem correção por mérito próprio, independente de qualquer investigação externa a este repositório: `ReportServerMQD.postReport` (`src/domain/services/report_server_mqd.go`) trata qualquer status HTTP diferente de `200` como um simples `Warning` e retorna `nil` (sucesso) mesmo assim; e `ResultProcessor.processAndSendResults` (`src/application/result_processor.go`) já apaga os resultados agregados da memória (`getAndClearResults`) *antes* de confirmar que o envio foi bem-sucedido, sem retry nem persistência local (sem padrão outbox). Se o envio ao servidor central falhar de qualquer forma — timeout, erro 5xx, erro de rede — os dados daquele ciclo são perdidos silenciosamente dentro do próprio mqd-client, e ele nunca sabe (nem alerta) que isso aconteceu. Além disso, o mesmo laço aborta o processamento de **todos** os demais `transmitterID`s do lote assim que um envio falha, ampliando a perda de dados.

Em paralelo, a análise de concorrência do código Go encontrou uma data race real: a variável de pacote `totalResults` é escrita sob `resultProcessorMutex` em `AppendResult` (chamada pela goroutine worker), mas é lida sem nenhum lock dentro do `select` de `StartResultsProcessor` (goroutine separada), violando o memory model do Go.

Antes de alterar qualquer código, precisamos formalizar essas descobertas como requisitos verificáveis (specs) e um plano técnico (design), para orientar a correção e os testes de regressão de forma consistente. Nenhuma implementação é feita nesta etapa.

## What Changes

- Especificar que respostas HTTP diferentes de `200` no `POST /report` (e falhas de transporte) devem ser tratadas como erro real pelo mqd-client, nunca como sucesso silencioso.
- Especificar que os resultados agregados de um ciclo de relatório não podem ser descartados da memória antes de uma confirmação de entrega bem-sucedida; em caso de falha, deve haver retry com backoff e/ou persistência local (outbox) para não perder dados de qualidade já coletados.
- Especificar que a falha ao enviar o relatório de um `transmitterID` não deve interromper o processamento/envio dos demais `transmitterID`s do mesmo ciclo.
- Especificar que todo acesso (leitura e escrita) ao estado compartilhado usado para decidir a janela/gatilho de envio de relatórios (`totalResults` e mapas de agregação) deve ser thread-safe, eliminando a data race identificada.
- Especificar (prioridade menor) que a inicialização de singletons compartilhados (`logger`, `report server`, `configuration manager`, `message processor worker`) deve seguir um padrão seguro para concorrência (ex.: `sync.Once`), mesmo que hoje só sejam instanciados sequencialmente em `main()`, para não se tornar uma race latente caso esse padrão de chamada mude no futuro.
- Este change cobre apenas `proposal.md` + `specs/` + `design.md` + `tasks.md`. Nenhuma task será executada nesta etapa — a lista de tasks fica pronta para revisão e aprovação antes de qualquer implementação.

## Capabilities

### New Capabilities
- `report-delivery-reliability`: garante que o envio de relatórios do mqd-client ao servidor central detecta corretamente falhas de entrega (HTTP e transporte), preserva os resultados agregados até confirmar entrega bem-sucedida (evitando perda silenciosa de dados), e isola falhas por `transmitterID` sem abortar o restante do lote. Escopo limitado ao trecho mqd-client → servidor central; o mqd-client não possui nenhuma integração direta com S3.
- `result-aggregation-concurrency`: garante acesso thread-safe ao estado compartilhado de agregação de resultados (contadores e mapas) usado para decidir quando disparar o envio de relatórios, e um padrão seguro de inicialização para os singletons da aplicação.

### Modified Capabilities
(nenhuma — não há specs arquivadas anteriormente em `openspec/specs/` para o mqd-client; este é o primeiro conjunto de requisitos formalizados)

## Impact

- Código afetado: `src/application/result_processor.go`, `src/domain/services/report_server_mqd.go`, `src/domain/services/api_dao.go`, e os pontos de inicialização de singleton em `src/application/message_process_Worker.go`, `src/application/configuration_manager.go`, `src/domain/services/report_server_factory.go`, `src/crosscutting/log/logger_factory.go`.
- Sem mudança no contrato HTTP exposto aos clientes do mqd-client (`POST /ValidateResponse` continua respondendo da mesma forma).
- Pode exigir nova configuração (ex.: número de tentativas de retry, backoff, diretório/arquivo para persistência local do outbox) — a ser detalhado em `design.md`.
- Sem impacto em produção nesta etapa: apenas artefatos de planejamento OpenSpec. A implementação real é um change futuro, após aprovação das tasks aqui descritas.
