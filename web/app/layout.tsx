import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "catchup-x-post",
  description: "エンジニア向け技術トピック解説記事",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className="min-h-screen bg-slate-100">
        <header className="bg-gradient-to-r from-blue-600 to-indigo-700 shadow-lg">
          <div className="mx-auto flex max-w-5xl items-center justify-between gap-3 px-4 py-5 sm:px-6">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/20 text-xl shadow-inner">
                📡
              </div>
              <div>
                <p className="text-lg font-bold leading-tight text-white">
                  catchup-x-post
                </p>
                <p className="text-xs text-blue-200">
                  エンジニア向け技術トピック解説
                </p>
              </div>
            </div>
            <nav className="flex items-center gap-1 text-sm font-medium">
              <Link
                href="/"
                className="rounded-full px-3 py-1.5 text-blue-100 transition hover:bg-white/10 hover:text-white"
              >
                一覧
              </Link>
              <Link
                href="/create"
                className="rounded-full px-3 py-1.5 text-blue-100 transition hover:bg-white/10 hover:text-white"
              >
                記事を生成
              </Link>
            </nav>
          </div>
        </header>
        <div className="pb-16">{children}</div>
      </body>
    </html>
  );
}
