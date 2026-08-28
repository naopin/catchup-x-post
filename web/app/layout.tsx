import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "catchup-x-post",
  description: "エンジニア向け技術トピック解説記事の一覧・詳細",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
