/**
 * Debug logger utility for StreamForge.
 *
 * Controlled by the APP_ENV environment variable (runtime, NOT build-time):
 *
 *   APP_ENV=development (default):
 *     ✅ SF DEBUG, SF INFO, SF API, SF SERVER-FETCH, SF WARN, SF ERROR — all enabled
 *
 *   APP_ENV=production / prod:
 *     ❌ SF DEBUG, SF INFO, SF API, SF SERVER-FETCH — silenced
 *     ✅ SF WARN, SF ERROR — always shown (necessary for troubleshooting)
 *
 * Runtime configuration:
 *   - SERVER: reads process.env.APP_ENV at runtime
 *   - CLIENT: reads window.__SF_APP_ENV, injected by root layout.tsx via <script> tag
 *   - No rebuild needed — set APP_ENV at container start
 *
 * Usage:
 *   import { logger } from '@/lib/logger';
 *   logger.info('[LoginPage]', 'Submitting login', { email });   // dev only
 *   logger.api('POST', '/auth/login', { status: 200 });          // dev only
 *   logger.warn('[proxy]', 'Backend slow', { elapsed: 5000 });   // always shown
 *   logger.error('[UploadForm]', 'Upload failed', err);          // always shown
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/**
 * Resolve the current APP_ENV value at call time.
 * - Server: process.env.APP_ENV (read at runtime, never baked in)
 * - Client: window.__SF_APP_ENV (injected by layout.tsx <script>)
 */
function getAppEnv(): string {
  // Server-side
  if (typeof window === 'undefined') {
    return process.env.APP_ENV ?? 'development';
  }
  // Client-side — read from the injected global
  return (window as unknown as Record<string, string>).__SF_APP_ENV ?? 'development';
}

function isProd(): boolean {
  const env = getAppEnv().toLowerCase();
  return env === 'production' || env === 'prod';
}

/**
 * In production, only WARN and ERROR are shown.
 * In development, everything is shown.
 */
function shouldLog(level: LogLevel): boolean {
  if (!isProd()) return true;               // dev → log everything
  return level === 'warn' || level === 'error';  // prod → only warnings & errors
}

function makeLog(level: LogLevel) {
  const method = level === 'debug' ? console.debug
    : level === 'info' ? console.info
      : level === 'warn' ? console.warn
        : console.error;

  return (...args: unknown[]) => {
    if (!shouldLog(level)) return;
    const timestamp = new Date().toISOString();
    method(`[SF ${level.toUpperCase()}] [${timestamp}]`, ...args);
  };
}

/**
 * Log an API request/response with a consistent format.
 * Only shown in development — too noisy for production.
 */
function apiLog(
  method: string,
  url: string,
  details?: Record<string, unknown>,
): void {
  if (isProd()) return;
  const timestamp = new Date().toISOString();
  console.info(
    `[SF API] [${timestamp}] ${method} ${url}`,
    details ? details : '',
  );
}

/**
 * Log server-side fetch calls (from Server Components / Route Handlers / Server Actions).
 * Only shown in development — too noisy for production.
 */
function serverFetchLog(
  method: string,
  backendUrl: string,
  details?: Record<string, unknown>,
): void {
  if (isProd()) return;
  const timestamp = new Date().toISOString();
  console.info(
    `[SF SERVER-FETCH] [${timestamp}] ${method} ${backendUrl}`,
    details ? details : '',
  );
}

export const logger = {
  /** Dev only — verbose debugging */
  debug: makeLog('debug'),
  /** Dev only — informational */
  info: makeLog('info'),
  /** Always shown — something unexpected but recoverable */
  warn: makeLog('warn'),
  /** Always shown — something failed */
  error: makeLog('error'),
  /** Dev only — API request/response tracing */
  api: apiLog,
  /** Dev only — server-side fetch tracing */
  serverFetch: serverFetchLog,
  /** Whether ALL logging is enabled (dev mode). Use for dev-only response headers etc. */
  get isEnabled() { return !isProd(); },
  /** Current environment value. Evaluated at call time. */
  get env() { return getAppEnv(); },
};
