# 収集係チェックリスト

## 実行内容

- `GET /api/news`（catchup-news、host 既定 :8081）
- 1トピック・`explanation` / `use_cases` 付き JSON

## 完了条件

- [ ] `health=1` が 200
- [ ] 返却 JSON に `url`, `title`, `summary`, `explanation`, `use_cases` がある
- [ ] `history/entries.jsonl` と類似していない
- [ ] `logs/topics_history.txt` とトピック重複していない

## 失敗時

- 類似 → 同一 URL を `exclude_url` に含めて再取得（最大5回）
- フィールド欠落 → `docker compose up -d --build news`

## 要件追記欄

（ここに収集ポリシーの追加を書く。例: 避けたいベンダー、優先する分野）
