---
description: Xから指定したキーワードで技術情報を調査・収集する（X未投稿）。エンジニア向けtips・トレンドを日本語で解説。キーワードと件数（デフォルト3）をユーザーに確認してから実行。
---

# X キーワード調査

指定キーワードで X を検索し、技術解説記事を生成する（投稿なし）。

## 実行前にユーザーに確認すること

1. **キーワード**（必須）: どんなキーワードで検索しますか？
   例: `Rust async`, `k8s gateway API`, `Claude MCP`
2. **件数**（デフォルト 3、最大 20）: 何件生成しますか？

確認が取れたら以下を実行。

## 前提確認

```bash
docker compose up -d --build news
curl -s -o /dev/null -w "%{http_code}\n" "http://localhost:8081/api/news?health=1"
# 200 であること
```

## 実行

```bash
TOPIC="<キーワード>" ARTICLE_COUNT=<件数> DRY_RUN=true ./download_post.sh
```

例:
```bash
TOPIC="Claude MCP" ARTICLE_COUNT=3 DRY_RUN=true ./download_post.sh
```

- `DRY_RUN=true` は固定（X投稿は絶対にしない）
- `TOPIC` に指定したキーワードで catchup-news が X 横断検索する
- 海外ツイート・公式ツイートも検索対象

## 出力確認

実行後に最新ファイルを表示:

```bash
ls -t output/*.txt | head -<件数> | xargs -I{} sh -c 'echo "=== {} ===" && cat {}'
```

## 解説記事フォーマット（チェックリスト）

- [ ] ファイル先頭が `トピック:`
- [ ] 【概要】【メリット・デメリット】【解説】【活用事例】がある
- [ ] メリット・デメリットは実務判断に役立つ具体情報（コスト・リスク・互換性など）
- [ ] 比較・ユースケースが含まれている
- [ ] `参考:` 行に URL
- [ ] 出力言語は日本語

## 障害対応

| 症状 | 対処 |
|------|------|
| connection refused :8081 | `docker compose up -d news` |
| fresh topic なし（5回枯渇） | キーワードを変えて再実行 |
| `explanation` 欠落 | `docker compose up -d --build news` |
