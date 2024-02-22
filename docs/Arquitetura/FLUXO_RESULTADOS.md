# Fluxo de Envio de Resultados

Este fluxo representa o processo de envio de resultados executado na aplicação.

![Image 1. ](./desenhos/fluxo_Envio_Results.png)

## Passos

| Passos | Participante | Descrição |
|-|-|-|
| 0. | RESULT_PROCESSOR | A cada intervalo definido o Componente inicia seu processo |
| 1. | RESULT_PROCESSOR | O componente solicita ao Validador a lista de resultados a serem processados |
| 2. e 3. | VALIDADOR | Após retornar os resultados, o validador limpa a lista para um novo processo |
| 4. | RESULT_PROCESSOR | O componente gera um novo resumo com a lista de resultados obtidos |
| 5. | RESULT_PROCESSOR | O componente solicita um token de acesso ao proxy do servidor |
| 6. | PROXY | Solicita o token ao servidor utilizando certificados para estabelecer mTLS |
| 7. e 8. | GATEWAY | O Gateway usa as credenciais para validar a ID do Cliente e obtem um token JWT, enviando ao Proxy |
| 9. | PROXY | Proxy retorna as informações recebidas ao cliente |
| 10. | RESULT_PROCESSOR | O componente envia o resumo criado para o servidor usando o token recebido como autorização |
| 11. | PROXY | O proxy estabelece uma conexão segura e envia as informações do relatório ao servidor |
| 12. | GATEWAY | O Gateway valida o acesso ao recurso (MQD_SERVER) com base no token recebido |
| 13. | GATEWAY | Se tudo estiver correto, o componente armazena as informações do relatório com a data de recebimento |
| 14. e 15. | GATEWAY | Gateway retorna a resposta do servidor para o componente Result Processor |

## Esquema de Report Enviado



![Image 2. ](./desenhos/report_schema.png)