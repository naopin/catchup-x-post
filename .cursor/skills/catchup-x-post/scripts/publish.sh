#!/bin/bash
# 投稿係のみ（要・事前に output/*.txt を人が確認）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
cd "$ROOT"
export LOGS_DIR="${LOGS_DIR:-./logs}"
export OUTPUT_DIR="${OUTPUT_DIR:-./output}"
LATEST="$(ls -t "$OUTPUT_DIR"/*.txt 2>/dev/null | head -1 || true)"
if [ -z "$LATEST" ]; then
  echo "記事がありません。先に dry-run.sh を実行してください。" >&2
  exit 1
fi
echo "投稿対象: $LATEST"
echo "内容を確認しましたか? [y/N]"
read -r ans
case "$ans" in
  y|Y|yes|YES) ;;
  *) echo "中止"; exit 1 ;;
esac
export DRY_RUN=false
exec ./download_post.sh
