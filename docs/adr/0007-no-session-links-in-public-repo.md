# 0007. Claude へのリンクを公開しない

- 日付: 2026-11-25
- 状態: 採用

## 背景
Claude Code がコミットメッセージの末尾に `Claude-Session:` 行を付け、個人のセッションへのリンクがパブリックリポジトリの履歴に載った。PR の説明文にも Claude へのリンクが入った。原因は2つある。Claude Code が「末尾に付ける」というシステムの指示に従ったことと、`gh` での書き込みや PR の作成時にリンクが自動で付くことである。ルールを書くだけでは防げない。

## 決定
- Claude へのリンク（`claude.ai`・`claude.com`）、`Claude-Session:` 行、「Generated with Claude Code」などのフッターは、コミット・PR・コメントのどこにも書かない。ルールは `CLAUDE.md` に置く。
- GitHub への書き込みは MCP ツールで行い、作成後に本文を読み直してリンクを消す。
- 仕組みでも止める。1か所が抜けても次で止まるよう、4段にする。
  1. `CLAUDE.md`：毎回のセッションで読ませる
  2. `.claude/hooks/block-session-links.sh`：Claude Code がコミットや GitHub への書き込みをする前に止める。`gh` での書き込みも止める
  3. `.githooks/commit-msg`・`pre-push`：git のコミットとプッシュで止める。クラウドではセッション開始時のフックが `core.hooksPath` を設定する
  4. `.github/workflows/no-session-links.yml`：PR のコミットメッセージ・タイトル・説明文を CI で検査する
- すでに載っていたリンクは消した。`main` の履歴を書き換え（ファイルの中身は変えていない）、PR #1〜#4 の説明文を書き直した。

## 結果
- `main` のコミットのハッシュが変わった。書き換える前にクローンした手元は、`git fetch` してから `git reset --hard origin/main` で合わせる。
- GitHub は PR ごとに元のコミットを `refs/pull/N/head` として残し、説明文の編集履歴も残す。利用者には消せないものがあるので、GitHub サポートに削除を依頼する。
- ローカルで git フックを使うには、一度だけ `git config core.hooksPath .githooks` を実行する。
- CI の検査でマージを止めるには、`main` のブランチ保護で `no-session-links` を必須にする。
