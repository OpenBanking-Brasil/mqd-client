## Purpose

Garante que o estado compartilhado usado para agregar resultados de validação e decidir quando disparar o envio de relatórios é acessado de forma segura por múltiplas goroutines, e que a inicialização de componentes singleton da aplicação não pode produzir instâncias inconsistentes sob concorrência.

## ADDED Requirements

### Requirement: Thread-safe aggregation state access
O sistema SHALL sincronizar toda leitura e toda escrita do estado compartilhado usado para agregar resultados de validação e decidir o disparo do envio de relatórios (incluindo contadores de total de resultados e mapas de resultados agrupados por transmissor/servidor), de forma que nenhuma goroutine leia ou modifique esse estado fora de sincronização. Não deve haver nenhuma condição de corrida (data race) detectável pelo detector de race do Go nesse acesso.

#### Scenario: Contagem de resultados usada para decidir o envio
- **WHEN** uma goroutine está adicionando um novo resultado de validação ao estado agregado ao mesmo tempo em que outra goroutine verifica se o número de resultados atingiu o limiar configurado para disparar o envio do relatório
- **THEN** ambas as operações SHALL ocorrer sob sincronização adequada, sem que a leitura do limiar ocorra sem proteção contra a escrita concorrente

#### Scenario: Execução sob detector de race
- **WHEN** o sistema é executado (em teste ou localmente) com o detector de race do Go habilitado, exercitando o fluxo de ingestão de mensagens concorrente ao ciclo periódico de geração de relatório
- **THEN** nenhuma condição de corrida SHALL ser reportada relacionada ao estado agregado de resultados

### Requirement: Safe singleton initialization
Componentes singleton compartilhados da aplicação (logger, cliente do servidor de relatório, gerenciador de configuração, processador de mensagens) SHALL ser inicializados exatamente uma vez, mesmo que seus pontos de acesso sejam chamados concorrentemente por múltiplas goroutines, e nenhum chamador SHALL observar uma instância parcialmente construída.

#### Scenario: Acesso concorrente ao ponto de obtenção do singleton
- **WHEN** múltiplas goroutines chamam concorrentemente o ponto de acesso de um componente singleton pela primeira vez
- **THEN** apenas uma instância SHALL ser criada e todas as goroutines SHALL receber a mesma instância totalmente inicializada, sem condição de corrida reportada pelo detector de race do Go
