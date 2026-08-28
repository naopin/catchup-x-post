"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { generateArticles, type GeneratedArticle } from "@/lib/api";

const COUNT_OPTIONS = Array.from({ length: 20 }, (_, i) => i + 1);
const MAX_KEYWORD_LENGTH = 100;

export default function CreatePage() {
  const [keyword, setKeyword] = useState("");
  const [count, setCount] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [articles, setArticles] = useState<GeneratedArticle[] | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = keyword.trim();
    if (!trimmed) {
      setError("キーワードを入力してください");
      return;
    }

    setLoading(true);
    setError(null);
    setArticles(null);
    try {
      const res = await generateArticles({ keyword: trimmed, count });
      setArticles(res.articles);
    } catch {
      setError(
        "記事の生成に失敗しました。catchup-news (news API) が起動しているか確認してください。",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto max-w-2xl px-4 py-10 sm:px-6">
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-800">✨ 記事を生成する</h2>
        <p className="mt-1 text-sm text-gray-500">
          キーワードを指定して X から技術トピックを収集し、解説記事を生成します
        </p>
      </div>

      <form
        onSubmit={handleSubmit}
        className="rounded-2xl border border-gray-200 bg-white p-6 shadow"
      >
        <div>
          <label
            htmlFor="keyword"
            className="block text-sm font-semibold text-gray-700"
          >
            キーワード
          </label>
          <input
            id="keyword"
            type="text"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            placeholder="例: Claude MCP, Rust async"
            maxLength={MAX_KEYWORD_LENGTH}
            disabled={loading}
            className="mt-2 w-full rounded-xl border border-gray-300 px-4 py-2.5 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200 disabled:bg-gray-100"
          />
        </div>

        <div className="mt-5">
          <label
            htmlFor="count"
            className="block text-sm font-semibold text-gray-700"
          >
            生成件数
          </label>
          <select
            id="count"
            value={count}
            onChange={(e) => setCount(Number(e.target.value))}
            disabled={loading}
            className="mt-2 w-32 rounded-xl border border-gray-300 px-4 py-2.5 text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-200 disabled:bg-gray-100"
          >
            {COUNT_OPTIONS.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="mt-6 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-700 px-4 py-3 text-sm font-semibold text-white shadow transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading ? (
            <>
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white" />
              Grok API で調査中...
            </>
          ) : (
            "生成する"
          )}
        </button>
      </form>

      {error && (
        <div className="mt-6 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          ⚠️ {error}
        </div>
      )}

      {articles && (
        <div className="mt-8">
          <h3 className="mb-4 text-sm font-semibold text-gray-600">
            生成された記事（{articles.length} 件）
          </h3>
          {articles.length === 0 ? (
            <p className="py-8 text-center text-sm text-gray-400">
              記事は生成されませんでした
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-4">
              {articles.map((article) => (
                <Link
                  key={article.id}
                  href={`/articles/${article.id}`}
                  className="group flex flex-col rounded-2xl border-l-4 border-blue-500 bg-white p-5 shadow transition-all duration-200 hover:-translate-y-1 hover:border-indigo-500 hover:shadow-xl"
                >
                  <h4 className="text-base font-semibold leading-snug text-gray-900 group-hover:text-indigo-700">
                    {article.title}
                  </h4>
                  <p className="mt-3 text-xs font-semibold text-blue-600 transition-all group-hover:translate-x-1 group-hover:text-indigo-600">
                    詳細を見る →
                  </p>
                </Link>
              ))}
            </div>
          )}
        </div>
      )}
    </main>
  );
}
