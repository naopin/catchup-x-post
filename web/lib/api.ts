export interface ArticleSummary {
  id: string;
  title: string;
  date: string; // ISO8601
  filename: string;
}

// ArticleDetail は Go API（internal/webarticle.Detail）の GET /api/articles/:id レスポンス。
export interface ArticleDetail {
  id: string;
  title: string;
  body: string;
  tweet: string;
  source_url: string;
  date: string; // ISO8601
}

export class ArticleNotFoundError extends Error {}

// GenerateRequest は POST /api/generate のリクエストボディ。
export interface GenerateRequest {
  keyword: string;
  count: number;
}

export interface GeneratedArticle {
  id: string;
  title: string;
}

// GenerateResponse は POST /api/generate（Go API）のレスポンス。
export interface GenerateResponse {
  articles: GeneratedArticle[];
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8082";
// Grok API 呼び出しを含むため生成には時間がかかる。バックエンド側の上限（5分）に合わせる。
const GENERATE_TIMEOUT_MS = 5 * 60 * 1000;

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers,
    },
  });

  if (res.status === 404) {
    throw new ArticleNotFoundError(`article not found (${path})`);
  }
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    const message =
      body && typeof body.error === "string" ? body.error : res.statusText;
    throw new Error(`API request failed: ${res.status} ${message} (${path})`);
  }

  return res.json() as Promise<T>;
}

export function fetchArticles(): Promise<ArticleSummary[]> {
  return apiFetch<ArticleSummary[]>("/api/articles");
}

export function fetchArticle(id: string): Promise<ArticleDetail> {
  return apiFetch<ArticleDetail>(`/api/articles/${encodeURIComponent(id)}`);
}

export function generateArticles(req: GenerateRequest): Promise<GenerateResponse> {
  return apiFetch<GenerateResponse>("/api/generate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
    signal: AbortSignal.timeout(GENERATE_TIMEOUT_MS),
  });
}
