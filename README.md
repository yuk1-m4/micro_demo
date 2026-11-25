本番のアプリケーション開発前に、技術検証をするためのリポジトリです。

いつもと違うワークフローと違うアプリケーション開発管理を実践中。

# 作業フロー

- LinearでIssue管理。整理してGithubへ転送
  - クラウド開発環境からは claude.ai のコネクタで Linear を読む（ローカルの MCP 設定は引き継がれない）
- 設計上の決定は [docs/adr](docs/adr/) に残す
- パブリックリポジトリなので、Claude へのリンクは載せない（[ADR 0007](docs/adr/0007-no-session-links-in-public-repo.md)）。ローカルでは一度だけ `git config core.hooksPath .githooks` を実行する
- Claude Code, Codexのクラウド開発をメインで使う
- できる限りややこしい動線を使わず、シンプルな開発ワークフローを目指す。オッカムの髭剃り
- IDEはDevin。私はCursorより先にWindsurfつかってたので……

# 開発対象
- [ ] 厳しいテストの実施。ミューテーションテストとか、普段やらないことも厳しく実装
- [ ] CI/CD を組んで、機械的にヘボいコードを検出して落とす
- [ ] Go言語でバックエンド処理。金銭の計算をそつなくできるかを知る
- [ ] フロントエンドはReact＋Vite。いつもの
- [ ] DBはAWS DynamoDB

# 開発環境（Docker）

| ディレクトリ | 内容 |
|---|---|
| `backend/` | Go の API サーバー（`:8080`、`GET /api/health`） |
| `frontend/` | React + Vite（`:5173`、`/api` を backend へプロキシ） |

```bash
docker compose up --build                         # 起動
docker compose run --rm backend go test ./...     # バックエンドのテスト
docker compose run --rm frontend npm test         # フロントエンドのテスト
```

Claude Code などのクラウド開発環境では、コンテナからの外部通信がプロキシ経由になるため、上書き設定を併用する。

```bash
docker compose -f compose.yaml -f compose.cloud.yaml build
```

各 Dockerfile の `prod` ターゲットは本番用イメージ（backend: distroless、frontend: nginx）。
frontend の nginx は `/api` を環境変数 `API_UPSTREAM`（既定 `http://backend:8080`）へ転送する。
