/**
 * Relative paths: Vite dev server proxies /api, /v1, /health to Go (vite.config.ts).
 * Production (e.g. Vercel static): set VITE_API_BASE=https://your-api.example.com at build time.
 */
import { getAccessToken } from './auth'
import { apiUrl } from './apiBase'

export { apiUrl } from './apiBase'

const jsonHeaders = { 'Content-Type': 'application/json' } as const

function authHeaders(base: Record<string, string>): Record<string, string> {
  try {
    const t = getAccessToken()?.trim()
    if (t) return { ...base, Authorization: `Bearer ${t}` }
  } catch {
    /* private mode / SSR */
  }
  return base
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function errorMessageFromBody(data: unknown, statusText: string): string {
  if (typeof data === 'string') {
    const trimmed = data.trim()
    if (trimmed) return trimmed
  }
  const o = data as Record<string, unknown> | null
  if (o && typeof o.error === 'string' && o.error) return o.error
  if (o && typeof o.message === 'string' && o.message) return o.message
  return statusText
}

function errorCodeFromBody(data: unknown): string | undefined {
  if (typeof data === 'string') return undefined
  const o = data as Record<string, unknown> | null
  return o && typeof o.code === 'string' ? o.code : undefined
}

export async function apiGet(path: string): Promise<unknown> {
  const res = await fetch(apiUrl(path), {
    credentials: 'include',
    headers: authHeaders({}),
  })
  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text) as unknown
    } catch {
      data = text
    }
  }
  if (!res.ok) {
    throw new ApiError(errorMessageFromBody(data, res.statusText), res.status, errorCodeFromBody(data))
  }
  return data
}

export async function apiPut(path: string, body: unknown): Promise<unknown> {
  const res = await fetch(apiUrl(path), {
    method: 'PUT',
    credentials: 'include',
    headers: authHeaders({ ...jsonHeaders }),
    body: JSON.stringify(body),
  })
  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text) as unknown
    } catch {
      data = text
    }
  }
  if (!res.ok) {
    throw new ApiError(errorMessageFromBody(data, res.statusText), res.status, errorCodeFromBody(data))
  }
  return data
}

export async function apiPost(path: string, body: unknown): Promise<unknown> {
  const res = await fetch(apiUrl(path), {
    method: 'POST',
    credentials: 'include',
    headers: authHeaders({ ...jsonHeaders }),
    body: JSON.stringify(body),
  })
  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text) as unknown
    } catch {
      data = text
    }
  }
  if (!res.ok) {
    throw new ApiError(errorMessageFromBody(data, res.statusText), res.status, errorCodeFromBody(data))
  }
  return data
}

export async function apiDelete(path: string): Promise<void> {
  const res = await fetch(apiUrl(path), {
    method: 'DELETE',
    credentials: 'include',
    headers: authHeaders({}),
  })
  if (res.status === 204 || res.status === 200) return
  const text = await res.text()
  let data: unknown = null
  if (text) {
    try {
      data = JSON.parse(text) as unknown
    } catch {
      data = text
    }
  }
  throw new ApiError(errorMessageFromBody(data, text || res.statusText), res.status, errorCodeFromBody(data))
}

export async function fetchHealth(): Promise<boolean> {
  try {
    const res = await fetch(apiUrl('/health'), { credentials: 'include' })
    return res.ok
  } catch {
    return false
  }
}

const healthFail = { ok: false as const, authDisabled: false, databaseReachable: false }

async function fetchHealthOnce(): Promise<{
  ok: boolean
  authDisabled: boolean
  databaseReachable: boolean
  databasePingError?: string
  encryptionKeyConfigured?: boolean
  orchestration?: {
    beam_width?: number
    consensus_mode?: string
    domain?: string
    explore_paths?: boolean
  }
}> {
  const res = await fetch(apiUrl('/health'), { credentials: 'include' })
  if (!res.ok) {
    // Render/Fly cold start can return 502/503 before the process is ready.
    if (res.status === 502 || res.status === 503 || res.status === 504) {
      throw new Error(`health ${res.status}`)
    }
    return healthFail
  }
  const data = (await res.json()) as {
    auth_disabled?: boolean
    encryption_key_configured?: boolean
    database?: { reachable?: boolean; ping_error?: string }
    orchestration?: {
      beam_width?: number
      consensus_mode?: string
      domain?: string
      explore_paths?: boolean
    }
  }
  return {
    ok: true,
    authDisabled: !!data.auth_disabled,
    databaseReachable: data.database?.reachable !== false,
    databasePingError: data.database?.ping_error,
    encryptionKeyConfigured: data.encryption_key_configured,
    orchestration: data.orchestration,
  }
}

export async function fetchHealthBody(): Promise<{
  ok: boolean
  authDisabled: boolean
  databaseReachable: boolean
  databasePingError?: string
  encryptionKeyConfigured?: boolean
  orchestration?: {
    beam_width?: number
    consensus_mode?: string
    domain?: string
    explore_paths?: boolean
  }
}> {
  // Render free tier can take 30–60s to wake; retry with backoff before showing unreachable.
  const delays = [0, 3000, 8000, 15000, 25000]
  for (let i = 0; i < delays.length; i++) {
    if (delays[i] > 0) await new Promise((r) => setTimeout(r, delays[i]))
    try {
      return await fetchHealthOnce()
    } catch {
      /* retry — Render free tier cold start */
    }
  }
  return healthFail
}

/** Download a binary/text export with auth headers (e.g. usage CSV). */
export async function apiDownload(path: string, filename: string): Promise<void> {
  const res = await fetch(apiUrl(path), {
    credentials: 'include',
    headers: authHeaders({}),
  })
  if (!res.ok) {
    const text = await res.text()
    let msg = res.statusText
    try {
      const o = JSON.parse(text) as Record<string, unknown>
      if (typeof o.error === 'string') msg = o.error
      else if (typeof o.message === 'string') msg = o.message
    } catch {
      if (text) msg = text
    }
    throw new ApiError(msg, res.status)
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
