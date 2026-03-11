import type { ApiError, Video, WatchHistory, AdminVideo, AdminStats } from './types';

const BASE_URL = '/api/proxy';
const ACCESS_TOKEN_NAME = 'sf_access_token';
const REFRESH_TOKEN_NAME = 'sf_refresh_token';

/**
 * 3. Cookie helpers
 * Work in both browser (document.cookie) and can be read server-side
 */
export function getTokenCookie(name: string): string | null {
  if (typeof window !== 'undefined') {
    // Client-side
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) {
      const match = parts.pop()?.split(';').shift();
      return match ? decodeURIComponent(match) : null;
    }
    return null;
  } else {
    // Server-side
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { cookies } = require('next/headers');
      const cookieStore = cookies();
      return cookieStore.get(name)?.value ?? null;
    } catch {
      return null;
    }
  }
}

export function setTokenCookie(name: string, token: string, expiresIn?: number): void {
  if (typeof window !== 'undefined') {
    let cookieStr = `${name}=${encodeURIComponent(token)}; path=/; samesite=lax`;
    if (expiresIn !== undefined) {
      cookieStr += `; max-age=${expiresIn}`;
    }
    document.cookie = cookieStr;
  } else {
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { cookies } = require('next/headers');
      const cookieStore = cookies();
      // Server components cannot call .set(), but Server Actions/Route handlers can.
      cookieStore.set({
        name,
        value: token,
        path: '/',
        maxAge: expiresIn,
        sameSite: 'lax',
      });
    } catch {
      // Ignore if called in a context where cookies cannot be set (e.g. Server Component)
    }
  }
}

export function clearAuthCookies(): void {
  if (typeof window !== 'undefined') {
    document.cookie = `${ACCESS_TOKEN_NAME}=; path=/; max-age=0; samesite=lax`;
    document.cookie = `${REFRESH_TOKEN_NAME}=; path=/; max-age=0; samesite=lax`;
  } else {
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { cookies } = require('next/headers');
      const cookieStore = cookies();
      cookieStore.delete(ACCESS_TOKEN_NAME);
      cookieStore.delete(REFRESH_TOKEN_NAME);
    } catch {
      // Ignore
    }
  }
}

/**
 * 2. attemptRefresh()
 * POSTs to /auth/refresh, reads 'sf_refresh_token' cookie
 * On success: writes new access token to cookie
 * On failure: throws (caller will redirect to login)
 */
async function attemptRefresh(): Promise<void> {
  const refreshToken = getTokenCookie(REFRESH_TOKEN_NAME);
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }

  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!res.ok) {
    throw new Error('Refresh failed');
  }

  const data = (await res.json()) as { access_token: string; expires_in: number };
  setTokenCookie(ACCESS_TOKEN_NAME, data.access_token, data.expires_in);
}

/**
 * 1. Base fetch wrapper fetchWithAuth
 * Reads JWT, attaches Authorization
 * On 401 response: calls attemptRefresh(), retries once
 * On second 401: clears cookies, redirects to /login
 * Returns typed response or throws ApiError
 */
async function fetchWithAuth<T>(path: string, method: string = 'GET', body?: unknown, retried: boolean = false): Promise<T> {
  const token = getTokenCookie(ACCESS_TOKEN_NAME);
  const headers = new Headers();

  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  // Generate random UUID for request tracing
  headers.set('X-Request-ID', crypto.randomUUID());

  if (body && !(body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  // 30 second timeout abort controller
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 30000);

  const options: RequestInit = {
    method,
    headers,
    signal: controller.signal,
  };

  if (body) {
    options.body = body instanceof FormData ? body : JSON.stringify(body);
  }

  let res: Response;
  try {
    res = await fetch(`${BASE_URL}${path}`, options);
  } catch (err: any) {
    if (err.name === 'AbortError') {
      if (process.env.NODE_ENV !== 'production') console.error(`[API] Timeout 30s: ${method} ${path}`);
      throw { error: 'Request timed out after 30 seconds' };
    }
    if (process.env.NODE_ENV !== 'production') console.error(`[API] Fetch Error: ${method} ${path}`, err);
    throw { error: err.message || 'Network error' };
  } finally {
    clearTimeout(timeoutId);
  }

  if (!res.ok) {
    if (res.status === 401) {
      if (!retried) {
        try {
          await attemptRefresh();
          return fetchWithAuth<T>(path, method, body, true); // retry once
        } catch (refreshErr) {
          clearAuthCookies();
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          throw refreshErr;
        }
      } else {
        clearAuthCookies();
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        throw new Error('Unauthorized');
      }
    }

    let errorData: ApiError | { error: string };
    try {
      errorData = await res.json();
    } catch {
      errorData = { error: await res.text() || res.statusText };
    }
    if (process.env.NODE_ENV !== 'production') {
      console.error(`[API] Error ${res.status} on ${method} ${path}:`, errorData);
    }
    throw errorData;
  }

  if (res.status === 204) {
    return undefined as unknown as T;
  }

  return res.json() as Promise<T>;
}

/**
 * 4. Typed API methods
 */
export const auth = {
  register: (email: string, password: string): Promise<{ user_id: string; email: string }> =>
    fetchWithAuth('/auth/register', 'POST', { email, password }),

  login: async (email: string, password: string): Promise<{ access_token: string; refresh_token: string; expires_in: number }> => {
    // Standard fetch wrapper is used here, credentials are provided directly.
    const data = await fetchWithAuth<{ access_token: string; refresh_token: string; expires_in: number }>('/auth/login', 'POST', { email, password });
    setTokenCookie(ACCESS_TOKEN_NAME, data.access_token, data.expires_in);
    setTokenCookie(REFRESH_TOKEN_NAME, data.refresh_token, 60 * 60 * 24 * 7); // 7 days expiry
    return data;
  },

  logout: async (): Promise<void> => {
    try {
      await fetchWithAuth('/auth/logout', 'POST');
    } finally {
      clearAuthCookies();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
  },
};

export const videos = {
  list: (page?: number, status?: string): Promise<{ videos: Video[]; total: number; page: number }> => {
    const params = new URLSearchParams();
    if (page !== undefined) params.append('page', page.toString());
    if (status) params.append('status', status);

    const qs = params.toString() ? `?${params.toString()}` : '';
    return fetchWithAuth(`/videos${qs}`);
  },

  search: (query: string, page?: number, limit?: number): Promise<{ videos: Video[]; total: number; page: number }> => {
    const params = new URLSearchParams();
    params.append('q', query);
    if (page !== undefined) params.append('page', page.toString());
    if (limit !== undefined) params.append('limit', limit.toString());

    const qs = params.toString() ? `?${params.toString()}` : '';
    return fetchWithAuth(`/videos/search${qs}`);
  },

  getById: (id: string): Promise<Video> =>
    fetchWithAuth(`/videos/${id}`),

  getStreamUrl: (id: string): Promise<{ master_playlist_url: string }> =>
    fetchWithAuth(`/videos/${id}/stream`),

  delete: (id: string): Promise<void> =>
    fetchWithAuth(`/videos/${id}`, 'DELETE'),
};

export const watch = {
  saveEvent: (videoId: string, eventType: 'play' | 'pause' | 'seek' | 'end', positionSeconds: number): Promise<void> =>
    fetchWithAuth(`/watch/${videoId}/event`, 'POST', { event_type: eventType, position_seconds: positionSeconds }),

  getHistory: (): Promise<WatchHistory[]> =>
    fetchWithAuth<{ history: WatchHistory[] }>('/watch/history').then(res => res.history),
};

export const admin = {
  getVideos: (page?: number, limit?: number): Promise<{ videos: AdminVideo[]; total: number; page: number; limit: number }> => {
    const params = new URLSearchParams();
    if (page !== undefined) params.append('page', page.toString());
    if (limit !== undefined) params.append('limit', limit.toString());
    
    const qs = params.toString() ? `?${params.toString()}` : '';
    return fetchWithAuth(`/admin/videos${qs}`);
  },

  getStats: (): Promise<AdminStats> =>
    fetchWithAuth('/admin/stats'),
};
