# mqd-client
Aplicação que será instalada no cliente (Receptora) encarregada de realizar validações aos payloads recebidos nas diferentes APIs OpenFinance, e depois enviar os resultados para um servidor central localizado no perímetro central

Nesta primeira versão, você terá as seguintes limitações:
1. Escopo reduzido para APIs da Fase 2
2. As validações são limitadas aos seguintes endpoints:
3. As informações são enviadas como um resumo em um intervalo de tempo definido (2 horas)


## O aplicativo possui os seguintes componentes

# API
Este módulo é responsável por receber as solicitações e encaminhá-las para a lista (queue) para trabalhos posteriores.

# Queue
Este componente se encarrega de armazenar as mensagens permitindo um processamento paralelo posterior

# Worker
Componente encarregado de processar em paralelo (múltiplos threads) os diferentes encontrados na fila
Executa uma validação inicial do formulário e, em seguida, encaminha as mensagens para o "Validador"

# Validator
Componente que executará as validações específicas para cada um dos payloads da mensagem

Também é responsável por carregar as regras de validação de um arquivo JSON

# Results
Encarregado de processar os resultados das validações em cada tempo definido.

Este processo agrupa os resultados, envia para o servidor central e limpa a lista de resultados localmente, reduzindo a necessidade de espaço.

# Monitoring
Componente encarregado de monitorar a qualidade do serviço, observando os valores de desempenho do sistema (CPU, memória) bem como valores específicos do aplicativo (número de solicitações).