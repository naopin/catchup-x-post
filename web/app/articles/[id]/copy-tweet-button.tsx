"use client";

import { useEffect, useState } from "react";

export function CopyTweetButton({ tweet }: { tweet: string }) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const timer = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timer);
  }, [copied]);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(tweet);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }

  return (
    <div className="relative inline-block">
      <button
        type="button"
        onClick={handleCopy}
        className="flex items-center gap-2 rounded-xl bg-blue-500 px-5 py-2.5 text-sm font-semibold text-white shadow transition hover:bg-blue-400 active:scale-95"
      >
        <span>{copied ? "✅" : "📋"}</span>
        {copied ? "コピーしました" : "コピー"}
      </button>
    </div>
  );
}
