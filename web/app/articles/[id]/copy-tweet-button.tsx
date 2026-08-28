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
        className="rounded-md border border-current px-4 py-2 text-sm font-medium hover:opacity-80"
      >
        コピー
      </button>
      {copied && (
        <span
          role="status"
          className="absolute left-1/2 top-full mt-2 -translate-x-1/2 whitespace-nowrap rounded-md bg-black px-3 py-1 text-xs text-white shadow"
        >
          コピーしました
        </span>
      )}
    </div>
  );
}
