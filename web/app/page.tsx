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
    <div className="animate-pulse rounded-2xl border-l-4 border-blue-200 bg-white p-5 shadow">
      <div className="h-4 w-3/4 rounded-full bg-gray-200" />
      <div className="mt-4 h-3 w-1/3 rounded-full bg-gray-100" />
      <div className="mt-4 h-3 w-1/4 rounded-full bg-blue-100" />
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
    <main className="mx-auto max-w-5xl px-4 py-10 sm:px-6">
      {/* ページヘッダー */}
      <div className="mb-8 flex items-end justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">記事一覧</h2>
          <p className="mt-1 text-sm text-gray-500">
            X から収集した最新の技術トピック解説
          </p>
        </div>
        {articles !== null && (
          <span className="rounded-full bg-blue-600 px-4 py-1.5 text-xs font-semibold text-white shadow">
            {articles.length} 件
          </span>
        )}
      </div>

      {/* エラー */}
      {error && (
        <div className="mb-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          ⚠️ {error}
        </div>
      )}

      {/* グリッド */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2">
        {!error &&
          articles === null &&
          Array.from({ length: 4 }).map((_, i) => (
            <ArticleCardSkeleton key={i} />
          ))}

        {!error && articles?.length === 0 && (
          <p className="col-span-full py-16 text-center text-sm text-gray-400">
            まだ記事がありません
          </p>
        )}

        {!error &&
          articles?.map((article, index) => (
            <Link
              key={article.id}
              href={`/articles/${article.id}`}
              className="group relative flex flex-col rounded-2xl border-l-4 border-blue-500 bg-white p-5 shadow transition-all duration-200 hover:-translate-y-1 hover:border-indigo-500 hover:shadow-xl"
            >
              {/* 記事番号 */}
              <span className="absolute right-4 top-4 text-xs font-bold text-gray-300">
                #{String(index + 1).padStart(2, "0")}
              </span>

              {/* タイトル */}
              <h3 className="pr-8 text-base font-semibold leading-snug text-gray-900 group-hover:text-indigo-700">
                {article.title}
              </h3>

              {/* 日付 */}
              <p className="mt-4 flex items-center gap-1 text-xs text-gray-400">
                <span>🕐</span>
                {formatDate(article.date)}
              </p>

              {/* リンク */}
              <p className="mt-3 text-xs font-semibold text-blue-600 transition-all group-hover:translate-x-1 group-hover:text-indigo-600">
                詳細を見る →
              </p>
            </Link>
          ))}
      </div>
    </main>
  );
}
