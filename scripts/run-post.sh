#!/bin/bash
# ルートの download_post.sh へ委譲（ラジオの scripts 相当）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$SCRIPT_DIR/../download_post.sh" "$@"
