// Server Components — reads K8s env var at runtime, never baked into bundle
export const serverApiUrl =
  process.env.API_URL ||
  'http://localhost:8080'

// Client Components — relative path, proxied through Next.js server
// No hardcoded URL, works in every environment without rebuild
export const clientApiUrl = '/api/proxy'