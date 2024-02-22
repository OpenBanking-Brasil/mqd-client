# Fluxo de Validação

Este fluxo representa o processo de validação executado na aplicação.

![Image 1. ](./desenhos/fluxo_Validation.png)

## Passos

| Passos | Participante | Descrição |
|-|-|-|
| 0. | VALIDATOR | O componente de validação carrega regras de validação com base em arquivos JSON de configuração |
| 1., 2. e 3. | MESSAGE_PROCESS_WORKER | O componente Message Process Worker checa a Fila e em seguida solicita e recebe da Fila a lista de tarefas enfileiradas a serem processadas |
| 4. | MESSAGE_PROCESS_WORKER | Para cada uma das mensagens encontradas, o Message Process Worker envia a mensagem para o componente Validação |
| 5. | VALIDATOR | O componente valida se o Endpoint está na lista de endpoints válidos |
| 6. | VALIDATOR | O componente desserializa as informações |
| 7. | VALIDATOR | O componente valida o objeto usando um esquema JSON carregado anteriormente |
| 8. | VALIDATOR | Retorna a resposta de validação ao Worker |
| 9. | MESSAGE_PROCESS_WORKER | As informações do resultado são salvas na memória |