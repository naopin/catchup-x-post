import Link from "next/link";

export default function ArticleNotFound() {
  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col items-center justify-center gap-4 p-8 text-center">
      <h1 className="text-2xl font-bold">記事が見つかりません</h1>
      <p className="text-sm opacity-70">
        指定された記事は存在しないか、削除された可能性があります。
      </p>
      <Link href="/" className="text-sm underline hover:opacity-80">
        ← 一覧に戻る
      </Link>
    </main>
  );
}
