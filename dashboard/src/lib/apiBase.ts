/**
 * API origin for same-host or split frontend (VITE_API_BASE).
 * Shared by api.ts and auth.ts to avoid circular imports.
 */
const apiOrigin = (import.meta.env.VITE_API_BASE ?? '').replace(/\/+$/, '')

/** Configured API origin from VITE_API_BASE (empty when using same-host relative paths). */
export function configuredApiOrigin(): string {
  return apiOrigin
}

/** True when the browser will call the API on a different origin than the dashboard. */
export function isCrossOriginApi(): boolean {
  if (!apiOrigin || typeof window === 'undefined') return false
  try {
    return new URL(apiOrigin).origin !== window.location.origin
  } catch {
    return false
  }
}

export function apiUrl(path: string): string {
  if (!path.startsWith('/')) return path
  // Collapse duplicate slashes (e.g. //api/...) and trim trailing slash on /api/auth/* so POST hits Go routes, not SPA 405.
  let p = path.replace(/\/{2,}/g, '/')
  if (p.startsWith('/api/auth/') && p.length > '/api/auth/'.length + 1 && p.endsWith('/')) {
    p = p.replace(/\/+$/, '')
  }
  return apiOrigin ? `${apiOrigin}${p}` : p
}
