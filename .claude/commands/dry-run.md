---
description: catchup-x-post を dry-run で実行する（収集 + 解説のみ、X未投稿）。output/*.txt に記事を保存する。
---

# dry-run 実行

収集係 + 解説係のみ実行（X 未投稿）。

## 実行

```bash
./download_post.sh
```

または複数件:

```bash
ARTICLE_COUNT=3 ./download_post.sh
```

## 前提確認

```bash
docker compose up -d --build news
curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:8081/api/news?health=1"
# 200 であること
```

## 収集係チェックリスト

- [ ] `health=1` が 200
- [ ] 返却 JSON に `url`, `title`, `summary`, `explanation`, `use_cases` がある
- [ ] `history/entries.jsonl` と類似していない
- [ ] `logs/topics_history.txt` とトピック重複していない

失敗時:
- 類似 → 同一 URL を `exclude_url` に含めて再取得（最大5回）
- フィールド欠落 → `docker compose up -d --build news`

## 解説係チェックリスト

- [ ] ファイル先頭が `トピック:`
- [ ] 【概要】【メリット・デメリット】【解説】【活用事例】がある
- [ ] メリット・デメリットは採用判断に役立つ具体（コスト・リスク・互換性など）
- [ ] `参考:` 行に URL
- [ ] 【X投稿文案】が JSON 生文字列ではない（280字以内のプレーンテキスト）

## 成果物確認

```bash
ls -t output/*.txt | head -1 | xargs cat
```
