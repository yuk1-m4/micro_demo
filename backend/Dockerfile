# syntax=docker/dockerfile:1
# 開発・テスト用: docker compose から go run / go test を実行する
FROM golang:1.27-alpine AS dev
WORKDIR /src
COPY go.mod ./
# proxy_ca はプロキシ環境（クラウド開発環境など）でのみ渡す。イメージには残らない
RUN --mount=type=secret,id=proxy_ca,required=false \
    if [ -f /run/secrets/proxy_ca ]; then export SSL_CERT_FILE=/run/secrets/proxy_ca; fi; \
    go mod download
COPY . .
EXPOSE 8080
CMD ["go", "run", "."]

# 本番用ビルド
FROM dev AS build
RUN CGO_ENABLED=0 go build -o /out/server .

FROM gcr.io/distroless/static-debian13 AS prod
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
