# catchup-x-post web

catchup-x-post の記事一覧・詳細を表示する Next.js フロントエンド。

## 技術スタック

- Next.js 15（App Router）
- TypeScript
- Tailwind CSS v4
- パッケージマネージャ: npm

## セットアップ

```bash
cd web
npm install
cp .env.local.example .env.local
# .env.local を編集（Go API サーバーの URL）
```

## 環境変数

| 変数名 | 説明 |
|---|---|
| `NEXT_PUBLIC_API_URL` | Go API のベース URL（デフォルト: `http://localhost:8082`） |

## 開発

```bash
npm run dev
```

http://localhost:3000 で確認できる。

## ビルド

```bash
npm run build
npm run start
```

## ディレクトリ

| パス | 内容 |
|---|---|
| `app/page.tsx` | 記事一覧 |
| `app/articles/[id]/page.tsx` | 記事詳細 |
| `app/create/page.tsx` | キーワード指定で記事を生成するフォーム |
| `lib/api.ts` | Go API（`/api/articles`, `/api/articles/:id`, `/api/generate`）を叩く型付きクライアント |

## 関連

- Go API サーバー: リポジトリルートの `cmd/`（Issue #2 で実装予定）
