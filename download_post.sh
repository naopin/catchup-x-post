#!/bin/bash

# catchup-news（ネタ取得）の API（ラジオ :8080 と被らない :8081）
NEWS_API="${CATCHUP_NEWS_BASE_URL:-http://localhost:8081}"
# 実行ログ・topics_history / 記事成果物
LOGS_DIR="${LOGS_DIR:-./logs}"
OUTPUT_DIR="${OUTPUT_DIR:-./output}"
# 本番投稿: DRY_RUN=false ./download_post.sh
# 記事件数: ARTICLE_COUNT=3 ./download_post.sh  （最大20・既定1）
# トピック指定: TOPIC="AI駆動開発" ./download_post.sh  （未指定なら自動選定）
DRY_RUN="${DRY_RUN:-true}"
ARTICLE_COUNT="${ARTICLE_COUNT:-1}"
TOPIC="${TOPIC:-}"

# --- スピナー ---
spin_pid=""

start_spinner() {
  local msg="$1"
  (
    i=0
    chars=('-' '\' '|' '/')
    while true; do
      printf "\r  [%s] %s" "${chars[$((i % 4))]}" "$msg"
      sleep 0.1
      i=$((i + 1))
    done
  ) &
  spin_pid=$!
}

stop_spinner() {
  [ -n "$spin_pid" ] && kill "$spin_pid" 2>/dev/null && wait "$spin_pid" 2>/dev/null
  spin_pid=""
  printf "\r\033[K"
}

elapsed() {
  echo $(( $(date +%s) - $1 ))
}

news_api_up() {
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 --max-time 5 \
    "$NEWS_API/api/news?health=1" 2>/dev/null || echo "000")
  [ "$code" = "200" ]
}

wait_for_news_api() {
  local i=0
  while [ "$i" -lt 30 ]; do
    if news_api_up; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  return 1
}

# --- メイン ---
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

NEWS_API="${CATCHUP_NEWS_BASE_URL:-http://localhost:8081}"
LOGS_DIR="${LOGS_DIR:-./logs}"
OUTPUT_DIR="${OUTPUT_DIR:-./output}"
DRY_RUN="${DRY_RUN:-true}"
ARTICLE_COUNT="${ARTICLE_COUNT:-1}"
TOPIC="${TOPIC:-}"

export CATCHUP_NEWS_BASE_URL="$NEWS_API"
export LOGS_DIR
export OUTPUT_DIR
export ARTICLE_COUNT
export TOPIC

echo ""
if [ "$DRY_RUN" = "false" ] || [ "$DRY_RUN" = "0" ]; then
  echo "=== Engineer News X投稿 生成・投稿開始 ==="
else
  echo "=== Engineer News X投稿 生成開始 (dry-run) ==="
fi
if [ -n "$TOPIC" ]; then
  echo "       NEWS_API=$NEWS_API (topic=$TOPIC・直近7日)"
else
  echo "       NEWS_API=$NEWS_API (topic=X自動選定・直近7日)"
fi
echo "       ARTICLE_COUNT=$ARTICLE_COUNT"
echo ""

mkdir -p "$LOGS_DIR" "$OUTPUT_DIR" history

# Step 1: catchup-news 起動 (0% -> 25%)
printf "[ 0%%] [1/3] catchup-news 起動確認\n"
T0=$(date +%s)

if ! news_api_up; then
  if command -v docker >/dev/null 2>&1 && [ -f docker-compose.yml ]; then
    start_spinner "catchup-news を Docker で起動中 (port 8081)..."
    if docker compose up -d --build news; then
      stop_spinner
      echo "      起動待機..."
      sleep 3
    else
      stop_spinner
      echo "[WARN] docker compose news の起動に失敗しました"
    fi
  fi
fi

if ! wait_for_news_api; then
  echo "[FAIL] catchup-news に接続できません ($NEWS_API)"
  echo "       docker compose up -d news   # host :8081"
  exit 1
fi

printf "[25%%] [1/3] catchup-news 準備完了 (%d秒)\n" "$(elapsed $T0)"
echo ""

# Step 2: ネタ取得・文案・画像 (25% -> 95%)
printf "[25%%] [2/3] ネタ取得・文案生成中\n"
start_spinner "ネタ取得・文案生成中..."
T1=$(date +%s)

POST_ARGS=(-count="$ARTICLE_COUNT")
if [ "$DRY_RUN" = "false" ] || [ "$DRY_RUN" = "0" ]; then
  POST_ARGS+=(-dry-run=false)
else
  POST_ARGS+=(-dry-run=true)
fi

RUN_LOG=$(mktemp /tmp/catchup_x_post_XXXXXX.log)
if command -v go >/dev/null 2>&1; then
  go run ./cmd/post "${POST_ARGS[@]}" >"$RUN_LOG" 2>&1
  STATUS=$?
else
  if command -v docker >/dev/null 2>&1 && [ -f docker-compose.yml ]; then
    docker compose build post >/dev/null 2>&1
    docker compose run --rm post "${POST_ARGS[@]}" >"$RUN_LOG" 2>&1
    STATUS=$?
  else
    stop_spinner
    echo "[FAIL] go も docker も利用できません"
    exit 1
  fi
fi
stop_spinner

if [ $STATUS -ne 0 ]; then
  echo "[FAIL] 投稿素材の生成に失敗しました"
  cat "$RUN_LOG"
  rm -f "$RUN_LOG"
  exit 1
fi

printf "[95%%] [2/3] 生成完了 (%d秒)\n" "$(elapsed $T1)"
echo ""
cat "$RUN_LOG" | sed 's/^/       /'
rm -f "$RUN_LOG"
echo ""

# Step 3: 出力確認 (95% -> 100%)
printf "[95%%] [3/3] 出力確認中..."
T2=$(date +%s)

SAVED_TXT=$(ls -t "$OUTPUT_DIR"/*.txt 2>/dev/null | head -"$ARTICLE_COUNT")
if [ -z "$SAVED_TXT" ]; then
  printf "\r[FAIL] output/ に記事 .txt が見つかりません\n"
  exit 1
fi

printf "\r[100%%] [3/3] 保存完了 (%d秒)\n" "$(elapsed $T2)"
while IFS= read -r f; do
  [ -z "$f" ] && continue
  TXT_SIZE=$(wc -c <"$f" | tr -d ' ')
  printf "        記事: %s (%s bytes)\n" "$f" "$TXT_SIZE"
done <<<"$SAVED_TXT"
printf "        実行ログ: %s/*.log\n" "$LOGS_DIR"
printf "        トピック履歴: %s/topics_history.txt\n" "$LOGS_DIR"
printf "        ネタ履歴: history/entries.jsonl\n"
echo ""

if [ "$DRY_RUN" = "false" ] || [ "$DRY_RUN" = "0" ]; then
  echo "=== 完了（Xに投稿済み） ==="
else
  echo "=== 完了（dry-run・X未投稿） ==="
  echo "       内容を確認後: DRY_RUN=false ./download_post.sh"
fi
echo ""
