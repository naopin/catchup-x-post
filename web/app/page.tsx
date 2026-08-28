"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { fetchArticles, type ArticleSummary } from "@/lib/api";

function formatDate(iso: string): string {
  const d = new Date(iso);
  if (!iso || Number.isNaN(d.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function ArticleCardSkeleton() {
  return (
    <div className="animate-pulse rounded-lg border border-gray-200 p-4">
      <div className="h-5 w-3/4 rounded bg-gray-200" />
      <div className="mt-3 h-4 w-1/3 rounded bg-gray-200" />
    </div>
  );
}

export default function Home() {
  const [articles, setArticles] = useState<ArticleSummary[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchArticles()
      .then((data) => {
        if (!cancelled) setArticles(data);
      })
      .catch(() => {
        if (!cancelled) setError("記事の取得に失敗しました");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6">
      <h1 className="text-2xl font-bold text-gray-900">記事一覧</h1>

      <div className="mt-6 flex flex-col gap-4">
        {error && <p className="text-sm text-red-600">{error}</p>}

        {!error &&
          articles === null &&
          Array.from({ length: 3 }).map((_, i) => <ArticleCardSkeleton key={i} />)}

        {!error && articles !== null && articles.length === 0 && (
          <p className="text-sm text-gray-500">まだ記事がありません</p>
        )}

        {!error &&
          articles?.map((article) => (
            <div
              key={article.id}
              className="rounded-lg border border-gray-200 p-4 shadow-sm"
            >
              <h2 className="text-lg font-semibold text-gray-900">
                {article.title}
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                {formatDate(article.date)}
              </p>
              <Link
                href={`/articles/${article.id}`}
                className="mt-3 inline-block text-sm font-medium text-blue-600 hover:underline"
              >
                詳細を見る
              </Link>
            </div>
          ))}
      </div>
    </main>
  );
}
