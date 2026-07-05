/**
 * API origin for same-host or split frontend (VITE_API_BASE).
 * Shared by api.ts and auth.ts to avoid circular imports.
 */
const apiOrigin = (import.meta.env.VITE_API_BASE ?? '').replace(/\/+$/, '')

export function apiUrl(path: string): string {
  if (!path.startsWith('/')) return path
  // Collapse duplicate slashes (e.g. //api/...) and trim trailing slash on /api/auth/* so POST hits Go routes, not SPA 405.
  let p = path.replace(/\/{2,}/g, '/')
  if (p.startsWith('/api/auth/') && p.length > '/api/auth/'.length + 1 && p.endsWith('/')) {
    p = p.replace(/\/+$/, '')
  }
  return apiOrigin ? `${apiOrigin}${p}` : p
}
