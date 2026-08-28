# 投稿係チェックリスト

## 前提

- **人が** 最新 `output/*.txt` の【X投稿文案】を承認済み
- `DRY_RUN=false` のみ本番投稿

## 実行

```bash
DRY_RUN=false ./download_post.sh
```

## 完了条件

- [ ] ツイート ID が実行ログに記録
- [ ] `history/entries.jsonl` に URL 追記
- [ ] `logs/topics_history.txt` に1行追記

## やってはいけないこと

- 記事未確認の自動投稿（cron で `DRY_RUN=false` を毎日回さない）
- 承認なしの投稿係単独実行

## 要件追記欄

（投稿タイミング・承認フロー・NGワードなど）
