# Fluxo Atualização de COnfiguração

Este fluxo representa o processo de atualização de configuração.

![Image 1. ](./desenhos/fluxo_atualizacao_configuracao.png)

## Passos

| Passos | Participante | Descrição |
|-|-|-|
| 0. | CONFIGURATION_MANAGER | CONFIGURATION_MANAGER solicita o proxy para o token usando o OrganizationID configurado |
| 1. | PROXY | O proxy solicita o token do servidor usando o OrganizationID enviado no cabeçalho usando uma conexão segura |
| 2. | GATEWAY | O componente valida o ID do cliente e retorna o token se for válido |
| 3. | PROXY | PROXY retorna o token recebido ao cliente |
| 4. | CONFIGURATION_MANAGER | O componente solicita a configuração através do proxy |
| 5. | PROXY | O proxy estabelece uma conexão segura com o servidor e solicita a configuração |
| 6. | GATEWAY | O componente retorna a configuração solicitada |
| 7. | PROXY | O proxy retorna a configuração solicitada ao cliente |
| 8. | CONFIGURATION_MANAGER | O componente valida as versões da configuração obtidas |
| 9. | CONFIGURATION_MANAGER | Se for uma versão diferente da mais recente, o componente solicita a configuração dos diferentes grupos de APIs através do proxy |
| 10. | PROXY | O componente estabelece uma conexão segura e solicita os arquivos de configuração |
| 11. | GATEWAY | O componente valida a existência dos arquivos e se for encontrado, os retorna |
| 12. | PROXY | O proxy retorna as informações recebidas ao cliente |
| 13. | CONFIGURATION_MANAGER | O componente atualiza a configuração localmente |


## Esquema de Definições de Configuração

O objeto recebido pelo cliente no momento da atualização da configuração está definido abaixo.

![Image 2. ](./desenhos/configuration_settings_schema.png)