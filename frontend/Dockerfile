# syntax=docker/dockerfile:1
# 開発・テスト用: docker compose から dev サーバーやテストを実行する
FROM node:22-alpine AS dev
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

FROM nginx:1.27-alpine AS prod
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
