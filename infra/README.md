
# Documentação de Infraestrutura

## Construção da Imagem 

```console
docker build --no-cache -f infra/dockerfile/dockerfile-minimal -t mqd-client .
```

## Execução do container

Nos exemplos abaixo estamos usando um SERVER_ORG_ID de teste e uma pasta relativa /certificates que contém os certificados a serem usados

Por padrão, os arquivos de certificado são carregados usando os nomes:

| Arquivo | Descrição |
|-|-|
| client.crt | Arquivo de certificado |
| client.key | Arquivo da chave |

mas estes podem ser modificados usando variáveis ​​de ambiente:

CLIENT_CRT_FILE e CLIENT_KEY_FILE respectivamente

```console
docker run -p "8081:8080" -e SERVER_ORG_ID=09b20d09-bf30-4497-938e-b0ead8ce9629 -v ./certificates:/usr/ParameterData/certificates mqd-client
```

## Execução usando Compose

Para executar o comando docker-compose, você precisa estar na pasta onde está o arquivo "docker-compose.yaml",
"/infra/dockerfile"

```console
docker-compose up
```