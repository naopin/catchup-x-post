---
description: エンジニア向け X 投稿パイプラインを実行する（収集 → 解説 → 任意で X 投稿）。dry-run、バッチ生成、output/*.txt 確認、X 投稿承認後の本番実行に使う。
---

# catchup-x-post 運用

エンジニア向け X 投稿パイプライン。実行は **shell + Go**（変更しない）。

## 3役割（実行順）

| 役割 | 担当 | 成果物 |
|------|------|--------|
| 収集係 | catchup-news `/api/news` + 重複除外 | ネタ JSON（コード内） |
| 解説係 | copywriter | `output/YYYYMMDDHHMM.txt` |
| 投稿係 | xclient（承認後のみ） | X 投稿 + 履歴更新 |

詳細チェックリスト: `collect.md`（収集係）、`write.md`（解説係）、`publish.md`（投稿係）

## プリフライト（毎回）

リポジトリルートで:

```bash
docker compose up -d --build news
curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:8081/api/news?health=1"
# 200 であること
test -f .env && grep -q GROK_API_KEY .env
export LOGS_DIR=./logs
export OUTPUT_DIR=./output
```

- トピックは **ユーザーから渡さない**（API が直近7日から自動選定）
- `TOPIC` 環境変数は不要

## 収集 + 解説（定期の既定・X未投稿）

```bash
./download_post.sh
```

- デフォルト `DRY_RUN=true`（投稿係は動かない）
- 複数件: `ARTICLE_COUNT=3 ./download_post.sh`（最大20・1件あたりネタ取得+文案）
- 所要: おおよそ 30〜90 秒 × 件数（Grok API）

## 成果物の確認

```bash
ls -t output/*.txt 2>/dev/null | head -1
```

記事フォーマット（`requirements.md` 参照）:
```text
トピック: {トピック名}

【概要】
【メリット・デメリット】
【解説】
【活用事例】
参考: {URL}
【X投稿文案】（280字以内）
```

## 投稿係（本番のみ・人の承認後）

1. 最新 `output/*.txt` を読み、記事と【X投稿文案】を確認
2. 問題なければ:

```bash
DRY_RUN=false ./download_post.sh
```

- X API 4変数が `.env` に揃っていること（`X_API_KEY`, `X_API_SECRET`, `X_ACCESS_TOKEN`, `X_ACCESS_SECRET`）
- **承認なしに `DRY_RUN=false` を実行しない**

## 重複回避

| ファイル | 用途 |
|----------|------|
| `history/entries.jsonl` | URL・タイトル類似の記事除外 |
| `logs/topics_history.txt` | 選定トピック名の重複除外（1行1トピック） |

類似時は最大5回まで再取得（`exclude_url` はその都度1件のみ）。

## 障害対応

| 症状 | 対処 |
|------|------|
| connection refused :8081 | `docker compose up -d news` |
| `explanation` 欠落・pick 失敗 | `docker compose up -d --build news` |
| 5回枯渇で fresh topic なし | `topics_history` / `history` を確認。必要なら除外行を見直し |
| 記事が `output/` にない | `.env` の `OUTPUT_DIR=./output` と `export OUTPUT_DIR` を確認 |
| 実行ログがない | `LOGS_DIR=./logs` を確認 |

## 定期実行

- **推奨**: OS cron で `DRY_RUN=true ./download_post.sh`（`scripts/cron.example` 参照）
- Loop は補助。毎ステップを Agent 推論にしない
