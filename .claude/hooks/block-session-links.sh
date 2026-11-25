#!/bin/bash
# Claude Code の PreToolUse フック。
# セッションへのリンクを含む git のコミットや GitHub への書き込みを、実行前に止める（ADR 0007）。
set -euo pipefail

input=$(cat)
tool=$(jq -r '.tool_name // ""' <<<"$input")
payload=$(jq -c '.tool_input // {}' <<<"$input")

pattern='claude\.ai/code/session_[A-Za-z0-9]|Claude-Session:'

case "$tool" in
  Bash)
    # 調べるための grep などは通し、コミット・タグ・ノートを作るコマンドだけ検査する
    cmd=$(jq -r '.command // ""' <<<"$payload")
    grep -qE 'git[[:space:]].*(commit|tag|notes)' <<<"$cmd" || exit 0
    text=$cmd
    ;;
  mcp__github__*)
    text=$payload
    ;;
  *)
    exit 0
    ;;
esac

if grep -qE "$pattern" <<<"$text"; then
  echo "ブロック: Claude のセッションへのリンク（または Claude-Session: 行）が含まれています。パブリックリポジトリなので削除してからやり直してください（CLAUDE.md・ADR 0007）。" >&2
  exit 2
fi
exit 0
