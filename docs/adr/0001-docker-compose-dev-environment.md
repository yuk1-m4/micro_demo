# 0001. 開発環境を Docker Compose で定義する

- 日付: 2026-11-24
- 状態: 採用
- 関連: Linear WORK-2

## 背景
ローカルとクラウド開発環境の両方で、同じ手順で Go と React+Vite を動かしてテストしたい。

## 決定
`backend/`・`frontend/` それぞれに Dockerfile を置き、`compose.yaml` でまとめて起動する。
Dockerfile は開発用（`dev`）と本番用（`prod`）の2段階にする。
本番用は backend が distroless、frontend が nginx。

## 結果
- 開発・テスト・本番ビルドを同じ Dockerfile から作れる。
- 開発時はソースをマウントするので、編集が即座に反映される。
- Docker が動くことが前提になる（クラウドでの起動は 0004 を参照）。
