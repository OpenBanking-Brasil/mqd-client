# Análise do MQD Client

## O que é e que problema resolve

O **MQD Client** é o cliente do **Motor de Qualidade de Dados (MQD)** do ecossistema **Open Finance Brasil**. Ele roda dentro da infraestrutura de cada instituição participante (banco/fintech) e valida em tempo real os payloads trocados nas APIs OpenFinance, sem que dados sensíveis precisem sair da instituição — apenas estatísticas agregadas são enviadas a um servidor central (`mqd.openfinancebrasil.org.br`).

**Problema resolvido:** o Open Finance Brasil precisa garantir que todos os participantes estejam enviando dados em conformidade com os schemas oficiais das APIs, mas não pode (nem deve) inspecionar dados pessoais/sensíveis dos clientes finais em um servidor central. O MQD Client resolve isso descentralizando a validação: cada instituição valida localmente e reporta apenas contagens de sucesso/erro por endpoint, deixando o servidor central com uma visão agregada de qualidade sem acesso ao conteúdo real.

## Como resolve — arquitetura e fluxo

Referência: [main.go](../src/main.go)

1. **Stack:** Go 1.23, sem banco de dados nem fila externa — usa um `chan` interno como fila e arquivos JSON locais para persistência de resultados.
2. **Entrada:** [api_server.go](../src/application/api_server.go) expõe `POST /ValidateResponse` — a instituição envia ali a resposta que sua API OpenFinance gerou, com headers como `endpointName`, `version`, `consentID`.
3. **Amostragem:** por taxa de amostragem configurável por criticidade do endpoint (`EXTREMELY_HIGH` → `VERY_LOW`), decide (via `crypto/rand`) se aquela mensagem específica será validada — evita validar 100% do tráfego.
4. **Validação:** as mensagens selecionadas vão para uma fila interna, consumida por um worker que valida o JSON contra o **JSON Schema** oficial do endpoint ([validation/schema_validator.go](../src/validation/schema_validator.go)).
5. **Agregação e envio:** os resultados (sucesso/erro) são agregados em janelas de tempo (padrão configurável, ex. 2h) e enviados via HTTP/REST — autenticado por OAuth2 client_credentials/JWT — ao servidor central MQD, junto com métricas de sistema (CPU, memória, latência).
6. **Config dinâmica:** a cada intervalo, o cliente baixa do servidor central os schemas de validação e regras de amostragem atualizadas, sem precisar reiniciar.
7. **Segurança de transporte:** um proxy NGINX local (mTLS com certificados ICP-Brasil) intermedeia a comunicação com o servidor central — o serviço Go em si não implementa TLS mútuo.

**Implantação:** via Docker (`infra/dockerfile/dockerfile-minimal` + `docker-compose.yaml`), rodando o serviço `mqd-client` junto com o proxy NGINX, configurado via `settings.yml` + variáveis de ambiente (modo TRANSMITTER/RECEIVER, org ID, URLs, políticas de retenção de logs locais).

## Onde os dados são salvos/enviados

- **Resultados locais**: gravados em **arquivos JSON no disco local** (`data_logs/AAAA-MM-DD/<appID>/HHMM-<api>.json`), com limpeza automática por política de retenção configurável — fica dentro da própria infraestrutura da instituição.
- **Resultados agregados**: enviados via **HTTP/REST** (autenticado com OAuth2/JWT) diretamente para o **servidor central do MQD** (`mqd.openfinancebrasil.org.br`), passando por um proxy NGINX local com mTLS.

## Existe integração com S3/AWS?

**Não.** Nem o código nem a documentação do mqd-client mencionam S3 ou qualquer serviço da AWS relacionado a storage.

A única ocorrência de "AWS" em toda a documentação está em [ARQUITETURA.MD:33](Arquitetura/ARQUITETURA.MD), numa tabela que descreve os componentes do **desenho geral** do ecossistema (incluindo o lado do servidor central do MQD, que é externo a este repositório):

> `GATEWAY | Camada responsável pelo controle e administração das APIs do servidor MQD | AWS Gateway`

Ou seja, é o **API Gateway** (AWS API Gateway) usado pelo **servidor central** do MQD para expor suas APIs — não é o mqd-client armazenando nada em S3, e não é um serviço de storage. É apenas infraestrutura de borda do lado do servidor central, citada de forma genérica no diagrama de arquitetura de alto nível.

**Resumo:** o mqd-client (este repositório) não usa AWS/S3 em nenhum lugar — nem para persistência, nem para envio de dados. Ele salva localmente em arquivos JSON (`data_logs/`) e envia os resultados agregados via HTTP/REST para a API do servidor MQD central. O que esse servidor central faz internamente com os dados (se usa S3, banco relacional, etc.) está fora do escopo deste repositório.
