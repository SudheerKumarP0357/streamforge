/**
 * Structured logger for StreamForge.
 *
 * Environment detection:
 *   - SERVER vs CLIENT: `typeof window === 'undefined'`
 *   - DEV vs PROD: reads APP_ENV at runtime (server: process.env, client: window.__SF_APP_ENV)
 *
 * Behaviour matrix:
 *   ┌────────────────────┬──────────────────────────────────────────────────┐
 *   │ Environment        │ Output                                          │
 *   ├────────────────────┼──────────────────────────────────────────────────┤
 *   │ Development (any)  │ Pretty-print with colors + emoji prefixes       │
 *   │ Production server  │ JSON lines: { timestamp, level, message, ...ctx}│
 *   │ Production client  │ Suppress debug/info; warn/error only            │
 *   └────────────────────┴──────────────────────────────────────────────────┘
 *
 * Sensitive data:
 *   - Auth tokens, passwords, and full SAS URLs are automatically redacted.
 *   - URLs are truncated to their base path (scheme + host + pathname).
 *
 * Usage:
 *   import { logger } from '@/lib/logger';
 *
 *   logger.debug('Cache miss', { component: 'VideoPlayer', video_id: 'abc' });
 *   logger.info('Page loaded', { page: '/dashboard', user_id: '42' });
 *   logger.warn('Slow response', { action: 'fetchVideos', component: 'API' });
 *   logger.error('Upload failed', new Error('timeout'), { video_id: 'xyz' });
 */

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

/** Structured context attached to every log entry. All fields are optional. */
export interface LogContext {
  page?: string;
  component?: string;
  action?: string;
  user_id?: string;
  video_id?: string;
  request_id?: string;
  /** Catch-all for ad-hoc fields */
  [key: string]: unknown;
}

// ---------------------------------------------------------------------------
// Environment helpers
// ---------------------------------------------------------------------------

const isServer = typeof window === 'undefined';

/**
 * Resolve APP_ENV at *call time* (never baked in at build time).
 *   - Server: process.env.APP_ENV
 *   - Client: window.__SF_APP_ENV (injected by root layout.tsx)
 */
function getAppEnv(): string {
  if (isServer) {
    return process.env.APP_ENV ?? 'development';
  }
  return (window as unknown as Record<string, string>).__SF_APP_ENV ?? 'development';
}

function isProd(): boolean {
  const env = getAppEnv().toLowerCase();
  return env === 'production' || env === 'prod';
}

function isDev(): boolean {
  return !isProd();
}

// ---------------------------------------------------------------------------
// Sensitive data redaction
// ---------------------------------------------------------------------------

/** Keys whose values should always be fully redacted. */
const REDACTED_KEYS = new Set([
  'token',
  'access_token',
  'refresh_token',
  'authorization',
  'password',
  'secret',
  'cookie',
  'set-cookie',
]);

/**
 * Truncate a URL to `scheme://host/pathname`, stripping query strings and
 * fragments (which often contain SAS tokens, auth codes, etc.).
 */
function truncateUrl(url: string): string {
  try {
    const u = new URL(url);
    return `${u.protocol}//${u.host}${u.pathname}`;
  } catch {
    // Not a valid URL — return as-is.
    return url;
  }
}

/** Returns true if the string looks like a full URL with query/fragment. */
function looksLikeFullUrl(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    /^https?:\/\/.+[?#]/.test(value)
  );
}

/**
 * Deep-clone the context object while redacting sensitive fields and
 * truncating URLs that contain query strings / fragments.
 */
function sanitize(obj: unknown): unknown {
  if (obj === null || obj === undefined) return obj;

  if (typeof obj === 'string') {
    return looksLikeFullUrl(obj) ? truncateUrl(obj) : obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(sanitize);
  }

  if (typeof obj === 'object') {
    // Handle Error instances — extract useful fields only.
    if (obj instanceof Error) {
      return {
        name: obj.name,
        message: obj.message,
        stack: isDev() ? obj.stack : undefined,
      };
    }

    const result: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
      if (REDACTED_KEYS.has(key.toLowerCase())) {
        result[key] = '[REDACTED]';
      } else if (looksLikeFullUrl(value)) {
        result[key] = truncateUrl(value as string);
      } else {
        result[key] = sanitize(value);
      }
    }
    return result;
  }

  return obj;
}

// ---------------------------------------------------------------------------
// Level ordering (for suppression logic)
// ---------------------------------------------------------------------------

const LEVEL_PRIORITY: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
};

function shouldLog(level: LogLevel): boolean {
  if (isDev()) return true; // Development → everything

  if (isServer) return true; // Production server → everything (JSON)

  // Production client → warn + error only
  return LEVEL_PRIORITY[level] >= LEVEL_PRIORITY.warn;
}

// ---------------------------------------------------------------------------
// Formatters
// ---------------------------------------------------------------------------

const EMOJI: Record<LogLevel, string> = {
  debug: '🐛',
  info: 'ℹ️ ',
  warn: '⚠️ ',
  error: '🔴',
};

const COLORS: Record<LogLevel, string> = {
  debug: '\x1b[90m',   // gray
  info: '\x1b[36m',    // cyan
  warn: '\x1b[33m',    // yellow
  error: '\x1b[31m',   // red
};
const RESET = '\x1b[0m';

/** Pretty format for development (both server and client). */
function prettyLog(
  level: LogLevel,
  message: string,
  context?: LogContext,
  error?: Error,
): void {
  const consoleFn =
    level === 'debug' ? console.debug
    : level === 'info' ? console.info
    : level === 'warn' ? console.warn
    : console.error;

  const timestamp = new Date().toISOString();
  const emoji = EMOJI[level];
  const envTag = isServer ? 'server' : 'client';

  if (isServer) {
    // Server dev — use ANSI colors
    const color = COLORS[level];
    const prefix = `${color}${emoji} [SF:${level.toUpperCase()}]${RESET}`;
    const ts = `\x1b[90m${timestamp}\x1b[0m`;
    const tag = `\x1b[35m[${envTag}]\x1b[0m`;
    const parts: unknown[] = [prefix, ts, tag, message];
    if (context && Object.keys(context).length > 0) parts.push(sanitize(context));
    if (error) parts.push('\n', error);
    consoleFn(...parts);
  } else {
    // Browser dev — use CSS (console supports %c)
    const cssMap: Record<LogLevel, string> = {
      debug: 'color:#999;font-weight:bold',
      info: 'color:#0ea5e9;font-weight:bold',
      warn: 'color:#f59e0b;font-weight:bold',
      error: 'color:#ef4444;font-weight:bold',
    };
    const prefix = `%c${emoji} [SF:${level.toUpperCase()}]`;
    const parts: unknown[] = [
      `${prefix} %c${timestamp} %c[${envTag}]%c ${message}`,
      cssMap[level],
      'color:#999',
      'color:#a855f7',
      'color:inherit',
    ];
    if (context && Object.keys(context).length > 0) parts.push(sanitize(context));
    if (error) parts.push(error);
    consoleFn(...parts);
  }
}

/** JSON structured output for production server. */
function jsonLog(
  level: LogLevel,
  message: string,
  context?: LogContext,
  error?: Error,
): void {
  const entry: Record<string, unknown> = {
    timestamp: new Date().toISOString(),
    level,
    message,
    environment: 'server',
  };

  if (context && Object.keys(context).length > 0) {
    Object.assign(entry, sanitize(context));
  }

  if (error) {
    entry.error = sanitize(error);
  }

  const consoleFn = level === 'error' ? console.error : level === 'warn' ? console.warn : console.log;
  consoleFn(JSON.stringify(entry));
}

/** Production client — minimal console output for warn/error only. */
function minimalLog(
  level: LogLevel,
  message: string,
  context?: LogContext,
  error?: Error,
): void {
  const consoleFn = level === 'error' ? console.error : console.warn;
  const parts: unknown[] = [`[SF:${level.toUpperCase()}] ${message}`];
  if (context && Object.keys(context).length > 0) parts.push(sanitize(context));
  if (error) parts.push(error);
  consoleFn(...parts);
}

// ---------------------------------------------------------------------------
// Core dispatch
// ---------------------------------------------------------------------------

function dispatch(
  level: LogLevel,
  message: string,
  context?: LogContext,
  error?: Error,
): void {
  if (!shouldLog(level)) return;

  if (isDev()) {
    prettyLog(level, message, context, error);
  } else if (isServer) {
    jsonLog(level, message, context, error);
  } else {
    minimalLog(level, message, context, error);
  }
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

function debug(message: string, context?: LogContext): void {
  dispatch('debug', message, context);
}

function info(message: string, context?: LogContext): void {
  dispatch('info', message, context);
}

function warn(message: string, context?: LogContext): void {
  dispatch('warn', message, context);
}

/**
 * Log an error. The second argument can be an Error object, a context object,
 * or omitted entirely:
 *
 *   logger.error('Boom');
 *   logger.error('Boom', new Error('details'));
 *   logger.error('Boom', { component: 'Upload' });
 *   logger.error('Boom', new Error('details'), { component: 'Upload' });
 */
function error(message: string, errorOrContext?: Error | LogContext, context?: LogContext): void {
  let err: Error | undefined;
  let ctx: LogContext | undefined;

  if (errorOrContext instanceof Error) {
    err = errorOrContext;
    ctx = context;
  } else if (errorOrContext && typeof errorOrContext === 'object') {
    ctx = errorOrContext as LogContext;
  }

  dispatch('error', message, ctx, err);
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

export const logger = {
  debug,
  info,
  warn,
  error,

  /** Whether verbose logging is enabled (development mode). */
  get isEnabled(): boolean {
    return isDev();
  },

  /** Current APP_ENV value, evaluated at call time. */
  get env(): string {
    return getAppEnv();
  },
};
