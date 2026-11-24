# 0001. 開発環境を Docker Compose で定義する

- 日付: 2026-11-24
- 状態: 採用
- 関連: Linear WORK-2

## 背景
ローカルとクラウド開発環境の両方で、同じ手順で Go と React+Vite を動かしてテストしたい。

## 決定
`backend/`・`frontend/` それぞれに Dockerfile を置き、`compose.yaml` でまとめて起動する。
Dockerfile は開発用（`dev`）と本番用（`prod`）の2段階にする。
本番用は backend が distroless、frontend が nginx。nginx は `/api` を `API_UPSTREAM`（既定 `http://backend:8080`）へ転送する。
開発時の変更の反映は、backend が air（再ビルド・再起動）、frontend が Vite が担う。

## 結果
- 開発・テスト・本番ビルドを同じ Dockerfile から作れる。
- 開発時はソースをマウントし、編集が自動で反映される。
- Docker が動くことが前提になる（クラウドでの起動は 0004 を参照）。
