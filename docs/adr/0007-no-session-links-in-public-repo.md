# 0007. Claude のセッションへのリンクを公開しない

- 日付: 2026-11-25
- 状態: 採用

## 背景
Claude Code がコミットメッセージの末尾に `Claude-Session:` 行を付け、個人のセッションへのリンクがパブリックリポジトリの履歴に載った。PR の説明文にも同じリンクが入った。Claude Code は「コミットや PR の末尾に付ける」というシステムの指示に従っていたため、ルールを書くだけでは再発を防げない。

## 決定
- セッションへのリンクと `Claude-Session:` 行は、コミット・PR・コメント・ファイルのどこにも書かない。ルールは `CLAUDE.md` に置く。
- 仕組みでも止める。1か所が抜けても次で止まるよう、4段にする。
  1. `CLAUDE.md`：毎回のセッションで読ませる
  2. `.claude/hooks/block-session-links.sh`：Claude Code がコミットや GitHub への書き込みをする前に止める
  3. `.githooks/commit-msg`・`pre-push`：git のコミットとプッシュで止める。クラウドではセッション開始時のフックが `core.hooksPath` を設定する
  4. `.github/workflows/no-session-links.yml`：PR のコミットメッセージと説明文を CI で検査する
- すでに載っていたリンクは、`main` の履歴を書き換えて消した（ファイルの中身は変えていない）。

## 結果
- `main` のコミットのハッシュが変わった。書き換える前にクローンした手元は、`git fetch` してから `git reset --hard origin/main` で合わせる。
- GitHub は PR ごとに元のコミットを `refs/pull/N/head` として残す。これは利用者には消せないので、PR #1・#3 の画面から古いコミットのメッセージは見えたままになる。完全に消すには GitHub サポートへの依頼が要る。
- ローカルで git フックを使うには、一度だけ `git config core.hooksPath .githooks` を実行する。
- CI の検査でマージを止めるには、`main` のブランチ保護で `no-session-links` を必須にする。
