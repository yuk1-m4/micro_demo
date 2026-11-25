#!/bin/bash
# クラウドセッション開始時の準備（ローカルでは何もしない）
#  1. git のコミット作成者を yuk1-m4 にする（このリポジトリのみ）
#  2. git フック（.githooks）を有効にする
#  3. Docker デーモンが止まっていれば起動する
set -euo pipefail

if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

cd "${CLAUDE_PROJECT_DIR:-.}"

git config user.name "yuk1-m4"
git config user.email "264142626+yuk1-m4@users.noreply.github.com"
# Claude へのリンクを含むコミットを止める（ADR 0007）
git config core.hooksPath .githooks

if ! docker info >/dev/null 2>&1; then
  # 起動途中の dockerd があれば待つだけにする（二重起動するとすぐ落ちる）
  if ! pgrep -x dockerd >/dev/null; then
    # フックが終わってもデーモンが残るよう、セッションから切り離して起動する
    setsid nohup dockerd >/tmp/dockerd.log 2>&1 </dev/null &
  fi
  for _ in $(seq 1 30); do
    docker info >/dev/null 2>&1 && break
    sleep 1
  done
  if ! docker info >/dev/null 2>&1; then
    echo "dockerd failed to start; see /tmp/dockerd.log" >&2
    exit 1
  fi
fi
