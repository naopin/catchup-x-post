# catchup-x-post

catchup-news の `/api/news` から技術トピックを自動選定し、解説記事と X 投稿文案を生成する CLI（任意で X 投稿）。

[catchup-news](https://github.com/)（ラジオ）とは別リポジトリ。ネタ収集は catchup-news に任せ、投稿と運用はこのリポジトリで行う。

## 前提

- [catchup-news](../catchup-news) が起動して `GET /api/news` が使えること
- `GROK_API_KEY`（文案生成）
- 本番投稿時: X API v2（OAuth 1.0a User context）の4つの認証情報

## セットアップ

```bash
cp .env.example .env
# .env を編集

# ネタ取得 API（:8081。ラジオの :8080 と被らない）
docker compose up -d news
```

## 使い方

```bash
# dry-run（デフォルト）: output/ に記事 .txt を保存、Xには投稿しない
# ネタは catchup-news が X を横断検索し、直近7日の技術トピックを1件自動選定
./download_post.sh

# 本番投稿（output/*.txt を確認したあと）
DRY_RUN=false ./download_post.sh

# 記事を3件まとめて生成（ネタ取得は1件ずつ・最大20件）
ARTICLE_COUNT=3 ./download_post.sh
# または
go run ./cmd/post -dry-run=true -count=3
```

## 環境変数

| 変数 | 説明 |
|------|------|
| `GROK_API_KEY` | xAI API キー（必須） |
| `GROK_MODEL` | デフォルト `grok-4` |
| `CATCHUP_NEWS_BASE_URL` | catchup-news の URL（デフォルト `http://localhost:8081`） |
| （なし） | トピックは API 側で X から自動選定（`TOPIC` は不要） |
| `ARTICLE_COUNT` / `-count` | 1回で生成する記事数（デフォルト `1`、最大 `20`） |
| `OUTPUT_DIR` | 記事 `.txt` 成果物（デフォルト `./output`） |
| `LOGS_DIR` | 実行 `.log`・`topics_history.txt`（デフォルト `./logs`） |
| `HISTORY_DIR` | 類似ネタ判定（`history/entries.jsonl`） |
| `X_API_KEY` 等 | 本番投稿時に必須 |

`docker compose run post` 時は compose 内で `http://news:8080` に自動設定される。

ラジオ用の catchup-news（:8080）を共有する場合のみ `.env` で:

```
CATCHUP_NEWS_BASE_URL=http://localhost:8080
```

## Cursor Skill

`.cursor/skills/catchup-x-post/` に運用手順（収集・解説・投稿の3役割）と要件追記用 `reference/` があります。

## 定期実行

`scripts/cron.example` を参照（既定は dry-run のみ）。

## フロー

1. catchup-news `/api/news` で直近7日の X から技術トピックを1件自動選定・要約
2. `history/entries.jsonl` と `logs/topics_history.txt` で重複を避け、被れば再取得
3. Grok で解説・活用事例付き記事 + 280 字以内の X 投稿文案を生成
4. dry-run なら `output/*.txt` に記事を保存 / 本番なら X には投稿文案のみ投稿
