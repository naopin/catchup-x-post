import Link from "next/link";
import { notFound } from "next/navigation";

import { ArticleNotFoundError, fetchArticle } from "@/lib/api";
import { CopyTweetButton } from "./copy-tweet-button";

function formatDate(date: string): string {
  if (!date) return "";
  const parsed = new Date(date);
  if (Number.isNaN(parsed.getTime())) return date;
  return new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(parsed);
}

export default async function ArticleDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  const article = await fetchArticle(id).catch((err) => {
    if (err instanceof ArticleNotFoundError) {
      return null;
    }
    throw err;
  });

  if (!article) {
    notFound();
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-2xl flex-col gap-8 p-8">
      <Link href="/" className="text-sm underline hover:opacity-80">
        ← 一覧に戻る
      </Link>

      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold">{article.title}</h1>
        {article.date && (
          <time dateTime={article.date} className="text-sm opacity-60">
            {formatDate(article.date)}
          </time>
        )}
      </header>

      <article className="whitespace-pre-wrap text-sm leading-relaxed">
        {article.body}
      </article>

      {article.source_url && (
        <p className="text-sm">
          参考:{" "}
          <a
            href={article.source_url}
            target="_blank"
            rel="noopener noreferrer"
            className="underline hover:opacity-80"
          >
            {article.source_url}
          </a>
        </p>
      )}

      <section className="flex flex-col gap-3 border-t pt-6">
        <h2 className="text-sm font-bold">📣 X 投稿文案</h2>
        <p className="whitespace-pre-wrap rounded-md border p-4 text-sm leading-relaxed">
          {article.tweet}
        </p>
        <CopyTweetButton tweet={article.tweet} />
      </section>
    </main>
  );
}
