---
name: x-collect
description: 指定キーワードで X（Twitter）を検索し、最新投稿10件と簡潔な概要を .txt ファイルに保存する軽量スキル。/x-research（詳細解説記事）より速くて軽い。Docker 不要・Grok API 直接呼び出し。「キーワードで X を集めて」「投稿をまとめて」「x-collect」「X ダイジェスト」「〇〇の最新ポストを調べて」と言ったら積極的に使う。
---

# X ポスト収集（x-collect）

キーワードで X をリアルタイム検索し、投稿10件＋概要を `output/` に保存する。
`/x-research` が詳細解説記事を1件生成するのに対し、こちらは投稿ダイジェストを素早く10件まとめる軽量版。

## ステップ1 — キーワード確認

まだ聞いていなければユーザーに1つだけ確認:

> **どのキーワードで X を検索しますか？**

## ステップ2 — GROK_API_KEY 読み込み

```bash
export $(grep -E '^GROK_API_KEY' .env | xargs)
export GROK_MODEL=$(grep -E '^GROK_MODEL' .env | cut -d= -f2)
GROK_MODEL=${GROK_MODEL:-grok-4}
echo "KEY: ${GROK_API_KEY:0:8}... MODEL: $GROK_MODEL"
```

`GROK_API_KEY` が空なら `.env` に `GROK_API_KEY=xai-...` を追加するようユーザーに伝えて終了。

## ステップ3 — Grok API で X 検索

`<キーワード>` を実際の値に置き換えて実行:

```bash
KEYWORD="<キーワード>"
TMPFILE=$(mktemp)

curl -s https://api.x.ai/v1/responses \
  -H "Authorization: Bearer $GROK_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$GROK_MODEL\",
    \"input\": [{
      \"role\": \"user\",
      \"content\": \"X(Twitter)で '${KEYWORD}' に関する最新投稿を10件検索してください。\\n\\n各投稿を以下の形式で出力してください:\\n\\n[番号]. @[アカウント名] — [投稿のひとこと要約]\\n概要:\\n[何を言っているか]\\n[なぜ注目されるか・背景]\\n[エンジニアへの実務的な影響や活用ポイント]\\nURL: [投稿URL]\\n\\n出力は日本語で。URLは必ず含めること。\"
    }],
    \"tools\": [{\"type\": \"x_search\"}]
  }" > "$TMPFILE"

CONTENT=$(python3 -c "
import json
with open('$TMPFILE') as f:
    d = json.load(f)
for item in d.get('output', []):
    if item.get('type') == 'message':
        for c in item.get('content', []):
            if c.get('type') == 'output_text':
                print(c.get('text', ''))
")
rm "$TMPFILE"
echo "$CONTENT"
```

レスポンス取得に失敗した場合は `echo "$RESPONSE"` でエラー内容を確認してユーザーに伝える。

## ステップ4 — ファイルに保存

取得したコンテンツを Write ツールで `output/YYYYMMDDHHMM_collect.txt` に書き込む。

ファイル名の日時は `date +%Y%m%d%H%M` で取得。

ファイル形式:

```
キーワード: <キーワード>
収集日時: YYYY-MM-DD HH:MM
---

1. @account — ひとこと要約
概要:
何を言っているか。
なぜ注目されているか・背景。
エンジニアへの実務的影響や活用ポイント。
URL: https://x.com/...

---

2. @account — ひとこと要約
概要:
...

（10件まで）
```

## ステップ5 — 完了報告

- 保存先ファイルパスをユーザーに伝える
- 先頭2〜3件を会話にも表示して内容をチラ見せする

## エラー対応

| 症状 | 対処 |
|------|------|
| `GROK_API_KEY` が空 | `.env` に `GROK_API_KEY=xai-...` を追加 |
| curl が 401 | API キーの値を確認 |
| 投稿がほとんど取れない | キーワードを英語 or 別表現で再試行 |
| JSON パース失敗 | `echo "$RESPONSE"` でレスポンス全体を確認 |
