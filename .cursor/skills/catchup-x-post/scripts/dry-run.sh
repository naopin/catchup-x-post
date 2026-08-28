#!/bin/bash
# 収集係 + 解説係まで（X未投稿）
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../../../.." && pwd)"
cd "$ROOT"
export LOGS_DIR="${LOGS_DIR:-./logs}"
export OUTPUT_DIR="${OUTPUT_DIR:-./output}"
export DRY_RUN=true
export ARTICLE_COUNT="${ARTICLE_COUNT:-1}"
exec ./download_post.sh
