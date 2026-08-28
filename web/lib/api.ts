export interface ArticleSummary {
  id: string;
  title: string;
  date: string; // ISO8601
  filename: string;
}

export interface ArticleDetail {
  id: string;
  title: string;
  body: string;
  tweet: string;
  source_url: string;
  date: string; // ISO8601
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8082";

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers,
    },
  });

  if (!res.ok) {
    throw new Error(`API request failed: ${res.status} ${res.statusText} (${path})`);
  }

  return res.json() as Promise<T>;
}

export function fetchArticles(): Promise<ArticleSummary[]> {
  return apiFetch<ArticleSummary[]>("/api/articles");
}

export function fetchArticle(id: string): Promise<ArticleDetail> {
  return apiFetch<ArticleDetail>(`/api/articles/${encodeURIComponent(id)}`);
}
