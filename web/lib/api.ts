export interface Article {
  id: string;
  title: string;
  summary: string;
  content: string;
  tweet: string;
  url: string;
  created_at: string;
}

export interface ArticleListResponse {
  articles: Article[];
  total: number;
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

export function fetchArticles(): Promise<ArticleListResponse> {
  return apiFetch<ArticleListResponse>("/api/articles");
}

export function fetchArticle(id: string): Promise<Article> {
  return apiFetch<Article>(`/api/articles/${encodeURIComponent(id)}`);
}
