# syntax=docker/dockerfile:1
# 開発・テスト用: docker compose から dev サーバーやテストを実行する
FROM node:26-alpine AS dev
WORKDIR /app
COPY package.json package-lock.json ./
# proxy_ca はプロキシ環境（クラウド開発環境など）でのみ渡す。イメージには残らない
RUN --mount=type=secret,id=proxy_ca,required=false \
    if [ -f /run/secrets/proxy_ca ]; then export NODE_EXTRA_CA_CERTS=/run/secrets/proxy_ca; fi; \
    npm ci
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev"]

# 本番用ビルド
FROM dev AS build
RUN npm run build

FROM nginx:1.31-alpine AS prod
COPY --from=build /app/dist /usr/share/nginx/html
# 起動時に API_UPSTREAM を埋め込んだ設定を生成する（nginx 公式イメージの templates 機能）
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
ENV API_UPSTREAM=http://backend:8080
# コンテナの DNS 設定を NGINX_LOCAL_RESOLVERS として使えるようにする
ENV NGINX_ENTRYPOINT_LOCAL_RESOLVERS=1
EXPOSE 80
