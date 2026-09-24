## Purpose

Garante que o mqd-client só considera um relatório de qualidade de dados "entregue" quando ele realmente foi aceito pelo servidor central, e que resultados agregados nunca são descartados silenciosamente quando a entrega falha.

## ADDED Requirements

### Requirement: Detect delivery failures accurately
O sistema SHALL tratar como falha de entrega qualquer resposta HTTP do endpoint de submissão de relatório que não seja um status de sucesso (2xx), assim como qualquer erro de transporte (timeout, conexão recusada, DNS, etc.). O sistema SHALL NOT considerar a entrega bem-sucedida com base apenas no fato de a requisição ter sido enviada sem erro de transporte.

#### Scenario: Servidor central responde com status de erro
- **WHEN** o servidor central responde à submissão de um relatório com um status HTTP diferente de 2xx (ex.: 401, 500, 503)
- **THEN** o mqd-client SHALL registrar essa tentativa como falha de entrega, não como sucesso

#### Scenario: Falha de transporte durante o envio
- **WHEN** a requisição de submissão do relatório falha por erro de rede/timeout antes de receber qualquer resposta do servidor
- **THEN** o mqd-client SHALL registrar essa tentativa como falha de entrega

#### Scenario: Servidor central responde com sucesso
- **WHEN** o servidor central responde à submissão de um relatório com status HTTP 2xx
- **THEN** o mqd-client SHALL considerar a entrega bem-sucedida

### Requirement: Preserve aggregated results until delivery is confirmed
O sistema SHALL manter os resultados de validação agregados de um ciclo de relatório disponíveis (em memória e/ou em armazenamento local durável) até que a entrega ao servidor central seja confirmada como bem-sucedida. Em caso de falha de entrega, o sistema SHALL tentar reenviar os mesmos resultados agregados (com uma política de retry/backoff) e/ou persisti-los localmente antes de removê-los do estado em memória, de forma que nenhum resultado de validação já coletado seja perdido silenciosamente por uma única falha de envio.

#### Scenario: Falha de entrega aciona nova tentativa
- **WHEN** uma tentativa de envio de relatório falha (conforme a Requirement "Detect delivery failures accurately")
- **THEN** os resultados agregados daquele ciclo SHALL permanecer disponíveis para reenvio, não sendo descartados do estado em memória apenas por causa dessa falha

#### Scenario: Falhas de entrega esgotam a política de retry
- **WHEN** as tentativas de reenvio de um relatório se esgotam sem sucesso, de acordo com a política de retry configurada
- **THEN** o sistema SHALL persistir os resultados agregados em um armazenamento local durável (ao invés de descartá-los) e SHALL registrar a falha em nível de severidade que permita alerta operacional

#### Scenario: Entrega bem-sucedida libera o estado agregado
- **WHEN** uma tentativa de envio de relatório é confirmada como bem-sucedida
- **THEN** os resultados agregados correspondentes àquele envio SHALL ser removidos do estado em memória (e do armazenamento local durável, se houver)

### Requirement: Isolate per-transmitter delivery failures
O sistema SHALL processar e tentar enviar o relatório de cada `transmitterID` de um ciclo de forma independente. Uma falha ao entregar o relatório de um `transmitterID` SHALL NOT impedir que os relatórios dos demais `transmitterID`s do mesmo ciclo sejam processados e enviados.

#### Scenario: Um transmitter falha, outros têm sucesso
- **WHEN** um ciclo de processamento contém resultados agregados para múltiplos `transmitterID`s e o envio do relatório de um deles falha
- **THEN** o sistema SHALL continuar tentando enviar os relatórios dos demais `transmitterID`s do mesmo ciclo

### Requirement: Observable failure signaling
O sistema SHALL registrar toda falha de entrega de relatório (HTTP ou transporte) em um nível de log que suporte alerta operacional (equivalente a erro, não a aviso silencioso), incluindo identificação do `transmitterID`/relatório afetado, de forma distinguível de um envio bem-sucedido.

#### Scenario: Falha de entrega gera log de erro identificável
- **WHEN** uma tentativa de envio de relatório falha por qualquer motivo coberto pela Requirement "Detect delivery failures accurately"
- **THEN** o sistema SHALL emitir um registro de log em nível de erro contendo informação suficiente para identificar o `transmitterID` e o ciclo de relatório afetados
