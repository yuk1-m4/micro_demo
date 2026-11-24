# ADR（Architecture Decision Records）

設計上の決定を1件1ファイルで残す。書き方は Nygard 形式（背景・決定・結果）。

- 1本は20行程度まで。手順はコミットや README に任せ、「なぜそうしたか」だけを書く。
- 「結果」には良い点と引き換えにした点を両方書く。
- 決定を覆すときは古い ADR を書き換えず、新しい ADR を追加して古い方の状態を「置き換え」にする。

| No. | タイトル | 状態 |
|---|---|---|
| [0001](0001-docker-compose-dev-environment.md) | 開発環境を Docker Compose で定義する | 採用 |
| [0002](0002-use-latest-stable-versions.md) | 言語・ツールは最新の安定版を使う | 採用 |
| [0003](0003-cloud-proxy-override.md) | クラウド環境用のプロキシ設定は上書きファイルに分ける | 採用 |
| [0004](0004-session-start-hook.md) | セッション開始時の準備は SessionStart フックで行う | 採用 |
