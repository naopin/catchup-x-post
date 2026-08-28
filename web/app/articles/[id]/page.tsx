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

type Section = { title: string; content: string };

function parseSections(body: string): Section[] {
  const parts = body.split(/(【[^】]+】)/);
  const sections: Section[] = [];
  for (let i = 1; i < parts.length; i += 2) {
    const title = parts[i].replace(/[【】]/g, "");
    const content = (parts[i + 1] ?? "").trim();
    if (content) sections.push({ title, content });
  }
  return sections;
}

const SECTION_CONFIG: Record<
  string,
  { icon: string; badge: string; border: string; bg: string; title: string }
> = {
  概要: {
    icon: "📝",
    badge: "bg-blue-100 text-blue-700",
    border: "border-blue-200",
    bg: "bg-blue-50",
    title: "text-blue-800",
  },
  "メリット・デメリット": {
    icon: "⚖️",
    badge: "bg-amber-100 text-amber-700",
    border: "border-amber-200",
    bg: "bg-amber-50",
    title: "text-amber-800",
  },
  解説: {
    icon: "🔍",
    badge: "bg-green-100 text-green-700",
    border: "border-green-200",
    bg: "bg-green-50",
    title: "text-green-800",
  },
  活用事例: {
    icon: "💡",
    badge: "bg-purple-100 text-purple-700",
    border: "border-purple-200",
    bg: "bg-purple-50",
    title: "text-purple-800",
  },
};

const DEFAULT_CONFIG = {
  icon: "📄",
  badge: "bg-gray-100 text-gray-700",
  border: "border-gray-200",
  bg: "bg-gray-50",
  title: "text-gray-800",
};

function SectionBlock({ section }: { section: Section }) {
  const cfg = SECTION_CONFIG[section.title] ?? DEFAULT_CONFIG;
  return (
    <div className={`rounded-2xl border p-6 ${cfg.bg} ${cfg.border}`}>
      <div className="mb-4 flex items-center gap-2">
        <span className="text-lg">{cfg.icon}</span>
        <span
          className={`rounded-full px-3 py-0.5 text-xs font-bold ${cfg.badge}`}
        >
          {section.title}
        </span>
      </div>
      <p
        className={`whitespace-pre-wrap text-sm leading-relaxed ${cfg.title}`}
      >
        {section.content}
      </p>
    </div>
  );
}

export default async function ArticleDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  const article = await fetchArticle(id).catch((err) => {
    if (err instanceof ArticleNotFoundError) return null;
    throw err;
  });

  if (!article) notFound();

  const sections = parseSections(article.body);

  return (
    <main className="mx-auto max-w-2xl px-4 pb-16 sm:px-6">
      {/* 戻るボタン */}
      <div className="py-6">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-sm font-medium text-gray-600 shadow-sm transition hover:bg-gray-50 hover:text-gray-900"
        >
          ← 一覧に戻る
        </Link>
      </div>

      {/* ヒーローヘッダー */}
      <div className="mb-8 overflow-hidden rounded-2xl bg-gradient-to-br from-blue-600 to-indigo-700 p-8 shadow-lg">
        <h1 className="text-2xl font-bold leading-snug text-white">
          {article.title}
        </h1>
        {article.date && (
          <time
            dateTime={article.date}
            className="mt-3 flex items-center gap-1.5 text-sm text-blue-200"
          >
            <span>🗓</span>
            {formatDate(article.date)}
          </time>
        )}
      </div>

      {/* 本文セクション */}
      <div className="flex flex-col gap-4">
        {sections.length > 0 ? (
          sections.map((s) => <SectionBlock key={s.title} section={s} />)
        ) : (
          <div className="rounded-2xl border border-gray-200 bg-white p-6">
            <p className="whitespace-pre-wrap text-sm leading-relaxed text-gray-800">
              {article.body}
            </p>
          </div>
        )}
      </div>

      {/* 参考リンク */}
      {article.source_url && (
        <div className="mt-6 rounded-xl border border-gray-200 bg-white px-5 py-3">
          <p className="text-xs text-gray-500">
            参考:{" "}
            <a
              href={article.source_url}
              target="_blank"
              rel="noopener noreferrer"
              className="break-all text-blue-600 hover:underline"
            >
              {article.source_url}
            </a>
          </p>
        </div>
      )}

      {/* X 投稿文案 */}
      <section className="mt-6 overflow-hidden rounded-2xl bg-gray-900 shadow-lg">
        <div className="border-b border-white/10 px-6 py-4">
          <h2 className="flex items-center gap-2 text-sm font-bold text-white">
            <span>🐦</span> X 投稿文案
          </h2>
        </div>
        <div className="px-6 py-5">
          <p className="whitespace-pre-wrap rounded-xl bg-white/10 p-4 text-sm leading-relaxed text-gray-100">
            {article.tweet}
          </p>
          <div className="mt-4">
            <CopyTweetButton tweet={article.tweet} />
          </div>
        </div>
      </section>
    </main>
  );
}
