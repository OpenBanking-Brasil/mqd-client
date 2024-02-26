# Documentação de Arquitetura MQD

1. [Arquitetura](#arquitetura)
1. [Proxy](./PROXY.md)
3. Diagramas de Sequência
    1. [Fluxo da Receptora](./FLUXO_RECEPTORA.md)
    2. [Fluxo da Transmissora](./FLUXO_TRANSMISSORA.md)
    3. [Fluxo de Atualização de Configuração](./FLUXO_ATUALIZACAO_DE_CONFIGURACAO.md)
    4. [Fluxo de Validação](./FLUXO_VALIDACAO.md)
    5. [Fluxo de Envio de Resultados](./FLUXO_RESULTADOS.md)


## Arquitetura 

Apesar de ter sido criada como um monolito, a aplicação MQD – Client foi desenvolvida com um desacoplamento em componentes, o que permitirá futuramente dividi-la em seções menores (microserviços) se necessário.

![Imagem 1. Arquitetura](./desenhos/Architecture.png)

## Componentes

| Serviço | Descrição | Tecnologia | Versão |
|---|---|---|---|
| SERVICE | Serviço executado na Instituição financeira que envia a resposta obtida ao MQD | N.A | N.A |
| API | API REST que expõe os métodos necessários para validar as mensagens da RECEPTORA | Go | 1.20 |
| CONFIGURATION MANAGER | Componente responsável por ler os arquivos de configuração e estabelecer os valores encontrados para que a aplicação possa utilizá-los | Go | 1.20 |
| MONITORING | Componente responsável pela criação das métricas da aplicação, tanto de desempenho quanto de negócio | Go <br /> OpenTelemetry | 1.20 <br /> |
| QUEUE MANAGER | Linha que armazena as mensagens recebidas pela API | Go | 1.20 |
| MESSAGE PROCESS WORKER | Componente que lê a fila de tarefas e processa cada uma das mensagens | Go | 1.20 |
| VALIDATOR | Componente que valida mensagens convertendo-as em objetos e executando uma validação baseada em um esquema JSON para cada um dos endpoints | Go | 1.20 |
| QUEUE RESULTS |  Linha que salva os resultados das mensagens já validadas | Go | 1.20 |
| RESULT PROCESSOR | Componente responsável por processar os resultados, criando um resumo de cada janela de tempo (definida na configuração) e enviando-os ao servidor MQD | Go | 1.20 |
| MQD SERVER PROXY | Servidor proxy reverso responsável por estabelecer uma conexão segura com o servidor | NGINX | 1.25 |
| GATEWAY | Camada responsável pelo controle e administração das APIs do servidor MQD | AWS Gateway | N.A. |
