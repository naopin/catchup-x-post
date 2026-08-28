# catchup-x-post

## プロジェクトの目的

[catchup-news](../catchup-news) の `/api/news` で X から技術トピックを自動選定し、解説記事と X 投稿文案を生成する。必要に応じて X API でテキスト投稿する CLI。

ラジオ（`catchup-news` + `download_radio.sh`）とは別リポジトリ。ネタ収集は catchup-news、投稿運用はこのリポジトリ。

## 主な機能

- `go run ./cmd/post` — 収集 → 解説 →（任意）X 投稿
- `./download_post.sh` — 上記を進捗表示付きで実行（推奨）
- `.cursor/skills/catchup-x-post/` — 運用 Skill（3役割・要件は `reference/`）

## 技術スタック

- **言語**: Go
- **AI**: Grok API（`grok-4` デフォルト、文案生成）
- **ネタ取得**: catchup-news HTTP API（Grok `x_search` は catchup-news 側）
- **投稿**: X API v2 テキストツイート（OAuth 1.0a）
- **インフラ**: Docker / docker-compose（任意）

## 環境変数

| 変数名 | 説明 |
|---|---|
| `GROK_API_KEY` | X.AI の API キー（必須） |
| `GROK_MODEL` | 使用モデル（デフォルト: `grok-4`） |
| `CATCHUP_NEWS_BASE_URL` | catchup-news の URL（デフォルト: `http://localhost:8081`・ラジオの :8080 と分離） |
| （なし） | トピックは catchup-news が X 横断検索で自動選定（直近7日） |
| `ARTICLE_COUNT` / `-count` | 1回の生成件数（デフォルト `1`、最大 `20`） |
| `OUTPUT_DIR` | 記事 `.txt` 成果物（デフォルト: `./output`）。複数件時は `YYYYMMDDHHMM_01.txt` 形式 |
| `LOGS_DIR` | 実行 `.log`・`topics_history.txt`（デフォルト: `./logs`） |
| `HISTORY_DIR` | 類似ネタ判定（`history/entries.jsonl`） |
| `X_API_KEY` | X API（本番投稿時） |
| `X_API_SECRET` | X API（本番投稿時） |
| `X_ACCESS_TOKEN` | X API（本番投稿時） |
| `X_ACCESS_SECRET` | X API（本番投稿時） |
| `DRY_RUN` | `download_post.sh` 用。`true` で X 未投稿（デフォルト） |

## 開発

```bash
# .env を用意
cp .env.example .env

# ネタ取得 API を起動（host :8081・ラジオと並行可）
docker compose up -d news

# dry-run（output/*.txt に記事保存・X未投稿）
./download_post.sh

# 本番投稿（output/*.txt を確認したあと）
DRY_RUN=false ./download_post.sh

# 直接 CLI
go run ./cmd/post -dry-run=true
```

## ディレクトリ

| パス | 内容 |
|---|---|
| `cmd/post/` | CLI エントリ |
| `internal/newsclient/` | catchup-news `/api/news` クライアント |
| `internal/picker/` | 未投稿ネタ 1 件選定 |
| `internal/copywriter/` | 解説記事 + X 投稿文案 |
| `internal/xclient/` | X API 投稿 |
| `internal/history/` | 生成済みネタの類似判定・`entries.jsonl` |
| `output/` | 記事 `.txt` 成果物 |
| `logs/` | 実行 `.log`、`topics_history.txt` |
| `history/` | `entries.jsonl`（URL・タイトル類似の除外用） |

## Docker ポート

| サービス | ホストポート | 備考 |
|---|---|---|
| catchup-news（ラジオ） | **8080** | `../catchup-news/docker-compose.yml` |
| catchup-x-post `news` | **8081** | このリポの `docker compose up -d news` |
| catchup-x-post `api` | **8082** | Go API サーバー（`cmd/web/`）。`Dockerfile.web` |
| catchup-x-post `web` | **3000** | Next.js フロント（`web/`）。`web/Dockerfile` |

## 3役割

| 役割 | 実装 | 出力 |
|------|------|------|
| 収集係 | catchup-news + 重複チェック | ネタ JSON |
| 解説係 | copywriter | `output/*.txt` |
| 投稿係 | xclient | X 投稿（dry-run 時はスキップ） |

## 関連

- ラジオ生成: `../catchup-news/download_radio.sh`
- 定期実行例: `scripts/cron.example`
- 運用 Skill: `.cursor/skills/catchup-x-post/`
