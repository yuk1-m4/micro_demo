#!/bin/bash
# Claude Code の PreToolUse フック。
# Claude へのリンクを含む git のコミットや GitHub への書き込みを、実行前に止める（ADR 0007）。
set -euo pipefail

input=$(cat)
tool=$(jq -r '.tool_name // ""' <<<"$input")
payload=$(jq -c '.tool_input // {}' <<<"$input")

pattern='claude\.ai/|claude\.com/|Claude-Session:'

case "$tool" in
  Bash)
    # 調べるための grep などは通し、コミット・タグ・ノートを作るコマンドだけ検査する
    cmd=$(jq -r '.command // ""' <<<"$payload")
    # gh で PR・Issue・コメントに書き込むと、プロキシが末尾に Claude へのリンクを自動で付ける。
    # GitHub への書き込みは MCP ツール（mcp__github__*）を使わせる。
    if grep -qE '(^|[;&|[:space:]])gh[[:space:]]+(pr|issue)[[:space:]]+(create|edit|comment|review)' <<<"$cmd" ||
      { grep -qE '(^|[;&|[:space:]])gh[[:space:]]+api' <<<"$cmd" &&
        grep -qE '(pulls|issues|comments|reviews)' <<<"$cmd" &&
        ! grep -qE '(-X|--method)[[:space:]]*GET' <<<"$cmd" &&
        grep -qE '(-X|--method)[[:space:]]*(POST|PATCH|PUT)|(^|[[:space:]])(-f|-F|--field|--raw-field|--input)[[:space:]=]' <<<"$cmd"; }; then
      echo "ブロック: gh で PR・Issue・コメントに書き込むと、Claude へのリンクが自動で付きます。mcp__github__* のツールを使ってください（CLAUDE.md・ADR 0007）。" >&2
      exit 2
    fi
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
  echo "ブロック: Claude へのリンク（claude.ai・claude.com）または Claude-Session: 行が含まれています。パブリックリポジトリなので削除してからやり直してください（CLAUDE.md・ADR 0007）。" >&2
  exit 2
fi
exit 0
