# Imagem do mqd-server-mock (substituto local do servidor central do MQD) usada
# pelo ambiente de QA local (docker-compose.yml na raiz do repo). As fixtures
# padrão já vão embutidas no binário (go:embed).
FROM golang:1.23 AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mqd-server-mock .

FROM alpine:3.20
COPY --from=builder /out/mqd-server-mock /usr/mqd-server-mock
EXPOSE 8082
ENTRYPOINT ["/usr/mqd-server-mock"]
