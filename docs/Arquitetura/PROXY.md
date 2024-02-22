### Proxy reverso

Para a versão 2.0 do Motor de Qualidade de Dados (MQD) foi implementada uma camada intermediária encarregada de estabelecer uma conexão segura (mTLS) utilizando os certificados da instituição.

Neste arquivo docker compose, incluímos um exemplo de proxy reverso usando NGINX como tecnologia, mas qualquer ferramenta pode ser usada conforme a conveniência da Instituição.

Será necessário que o proxy tenha a capacidade de rodar em modo "transparente", isso significa que a mensagem a ser enviada ao servidor MQD deve ser a mesma recebida em seus endpoints (cabeçalho e corpo).

Da mesma forma, este proxy terá acesso aos arquivos de certificados (Cert e key) da instituição financeira para permitir o estabelecimento de uma conexão segura com o servidor.

#### Endpoints

| Endpoint | Path | Proxy - Produção | Proxy - Sandbox | 
|-|-|-|-| 
|-|/token|https://mqd.openfinancebrasil.org.br/token|https://mqd.sandbox.openfinancebrasil.org.br/token| 
|-|/report|https://mqd.openfinancebrasil.org.br/report|https://mqd.sandbox.openfinancebrasil.org.br/report| 
|-|/settings|https://mqd.openfinancebrasil.org.br/settings|https://mqd.sandbox.openfinancebrasil.org.br/settings| 