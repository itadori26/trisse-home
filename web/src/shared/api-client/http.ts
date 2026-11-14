import { serverMessage, t } from '../i18n';
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
    /** Key of the backend message template (translated if known). */
    readonly messageKey?: string,
    readonly params?: Record<string, string>,
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
  /**
   * POSTs multipart form data (e.g. a file). A 401 here means a wrong
   * password re-entered in the form, so it does not end the session.
   */
  postForm<T>(path: string, form: FormData): Promise<T>;
  /**
   * POSTs JSON and returns the response body as a stream of text chunks,
   * e.g. an AI answer (#162); aborting the signal cancels the request.
   */
  postStream?(path: string, body: unknown, signal: AbortSignal): Promise<AsyncIterable<string>>;
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
  postForm<T>(path: string, form: FormData): Promise<T> {
    return this.request<T>('POST', path, form);
  }
  async postStream(path: string, body: unknown, signal: AbortSignal): Promise<AsyncIterable<string>> {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (this.csrfToken) headers[CSRF_HEADER] = this.csrfToken;
    const res = await fetch(path, {
      method: 'POST',
      headers,
      credentials: 'same-origin',
      body: JSON.stringify(body),
      signal,
    });
    if (!res.ok) {
      const err = ((await res.json().catch(() => undefined)) as ErrorBody | undefined)?.error;
      if (res.status === 401) this.onUnauthorized();
      throw new ApiError(
        err?.message ?? res.statusText,
        res.status,
        err?.code ?? 'unknown',
        err?.field,
        err?.messageKey,
        err?.params,
      );
    }
    const reader = res.body?.getReader();
    const decoder = new TextDecoder();
    return {
      async *[Symbol.asyncIterator]() {
        if (!reader) return;
        for (;;) {
          const { done, value } = await reader.read();
          if (done) return;
          yield decoder.decode(value, { stream: true });
        }
      },
    };
  }

  private async request<T>(method: string, path: string, body?: unknown, rawType?: string): Promise<T> {
    const isForm = typeof FormData !== 'undefined' && body instanceof FormData;
    const headers: Record<string, string> = { Accept: 'application/json' };
    // The browser sets the multipart boundary itself.
    if (body !== undefined && !isForm) headers['Content-Type'] = rawType ?? 'application/json';
    if (method !== 'GET' && this.csrfToken) headers[CSRF_HEADER] = this.csrfToken;
    const res = await fetch(path, {
      method,
      headers,
      credentials: 'same-origin',
      body: body === undefined ? undefined : isForm ? body : rawType ? String(body) : JSON.stringify(body),
    });
    if (res.status === 204) return undefined as T;
    const payload: unknown = await res.json().catch(() => undefined);
    if (!res.ok) {
      const err = (payload as ErrorBody | undefined)?.error;
      if (res.status === 401 && !path.startsWith('/api/auth/') && !isForm) this.onUnauthorized();
      throw new ApiError(
        err?.message ?? res.statusText,
        res.status,
        err?.code ?? 'unknown',
        err?.field,
        err?.messageKey,
        err?.params,
      );
    }
    return payload as T;
  }
}

/** Returns a user-facing message for any thrown value. */
export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return serverMessage(err.messageKey, err.params) ?? err.message;
  if (err instanceof Error) return err.message;
  return t('common.something_went_wrong');
}
