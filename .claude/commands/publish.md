---
description: catchup-x-post の X 投稿を本番実行する。事前に output/*.txt の【X投稿文案】を人が確認・承認した後のみ使う。
---

# 投稿係（本番実行）

**前提: 人が `output/*.txt` の【X投稿文案】を確認・承認済みであること。**

## 実行前チェック

```bash
# 最新記事を確認
ls -t output/*.txt | head -1 | xargs cat

# X API 認証情報の確認
grep -E "X_API_KEY|X_API_SECRET|X_ACCESS_TOKEN|X_ACCESS_SECRET" .env
```

## 実行

```bash
DRY_RUN=false ./download_post.sh
```

## 投稿係チェックリスト

- [ ] 【X投稿文案】を人が読んで承認した
- [ ] X API 4変数が `.env` に揃っている
- [ ] `DRY_RUN=false` を明示した

## 完了確認

- [ ] ツイート ID が実行ログに記録される
- [ ] `history/entries.jsonl` に URL が追記される
- [ ] `logs/topics_history.txt` に1行追記される

```bash
tail -1 logs/topics_history.txt
tail -1 history/entries.jsonl
```

## やってはいけないこと

- 記事未確認の自動投稿
- 承認なしの単独実行
- cron で `DRY_RUN=false` を毎日回すこと
