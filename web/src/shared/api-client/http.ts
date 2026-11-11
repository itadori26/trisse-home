import type { ErrorBody } from './generated/httpapi';

/** Header that carries the per-session CSRF token on state-changing requests. */
export const CSRF_HEADER = 'X-CSRF-Token';

/** Error returned by the backend, carrying the HTTP status and error code. */
export class ApiError extends Error {
  /**
   * @param status - HTTP status code
   * @param code - machine-readable error code from the error envelope
   * @param field - offending field for validation errors
   */
  constructor(
    message: string,
    readonly status: number,
    readonly code: string,
    readonly field?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/** Minimal typed HTTP client used by all API adapters. */
export interface HttpClient {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, body?: unknown): Promise<T>;
  put<T>(path: string, body?: unknown): Promise<T>;
  patch<T>(path: string, body?: unknown): Promise<T>;
  delete<T = void>(path: string): Promise<T>;
  /** POSTs a raw text body such as an uploaded HTML file. */
  postText<T>(path: string, text: string, contentType: string): Promise<T>;
}

/** fetch-based HttpClient that adds the CSRF header and parses error envelopes. */
export class FetchHttpClient implements HttpClient {
  private csrfToken = '';

  /** @param onUnauthorized - called when a request fails because the session expired */
  constructor(private readonly onUnauthorized: () => void = () => {}) {}

  /** Stores the CSRF token received after login or from the status endpoint. */
  setCsrfToken(token: string): void {
    this.csrfToken = token;
  }

  get<T>(path: string): Promise<T> {
    return this.request<T>('GET', path);
  }
  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }
  put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PUT', path, body);
  }
  patch<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('PATCH', path, body);
  }
  delete<T = void>(path: string): Promise<T> {
    return this.request<T>('DELETE', path);
  }
  postText<T>(path: string, text: string, contentType: string): Promise<T> {
    return this.request<T>('POST', path, text, contentType);
  }

  private async request<T>(method: string, path: string, body?: unknown, rawType?: string): Promise<T> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (body !== undefined) headers['Content-Type'] = rawType ?? 'application/json';
    if (method !== 'GET' && this.csrfToken) headers[CSRF_HEADER] = this.csrfToken;
    const res = await fetch(path, {
      method,
      headers,
      credentials: 'same-origin',
      body: body === undefined ? undefined : rawType ? String(body) : JSON.stringify(body),
    });
    if (res.status === 204) return undefined as T;
    const payload: unknown = await res.json().catch(() => undefined);
    if (!res.ok) {
      const err = (payload as ErrorBody | undefined)?.error;
      if (res.status === 401 && !path.startsWith('/api/auth/')) this.onUnauthorized();
      throw new ApiError(err?.message ?? res.statusText, res.status, err?.code ?? 'unknown', err?.field);
    }
    return payload as T;
  }
}

/** Returns a user-facing message for any thrown value. */
export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message;
  return 'Something went wrong.';
}
