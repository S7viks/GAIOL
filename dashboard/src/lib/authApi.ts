import { apiUrl } from './apiBase'
import { ACCESS_TOKEN_KEY, REFRESH_TOKEN_KEY, clearAuthStorage } from './auth'

function applyAuthTokensFromPayload(data: unknown): void {
  if (!data || typeof data !== 'object') return
  const d = data as Record<string, unknown>
  const session = d.session as Record<string, unknown> | undefined
  const access =
    (typeof d.access_token === 'string' && d.access_token) ||
    (session &&
      (typeof session.access_token === 'string'
        ? session.access_token
        : typeof session.accessToken === 'string'
          ? session.accessToken
          : '')) ||
    ''
  const refresh =
    (typeof d.refresh_token === 'string' && d.refresh_token) ||
    (session &&
      (typeof session.refresh_token === 'string'
        ? session.refresh_token
        : typeof session.refreshToken === 'string'
          ? session.refreshToken
          : '')) ||
    ''
  try {
    if (access) localStorage.setItem(ACCESS_TOKEN_KEY, access)
    if (refresh) localStorage.setItem(REFRESH_TOKEN_KEY, refresh)
  } catch {
    /* ignore */
  }
}

/** Map Supabase / backend auth errors to short, actionable UI copy. */
export function normalizeAuthError(raw: string): string {
  let msg = raw.replace(/^(sign(?:up|in)|recover|update password) failed:\s*/i, '').trim()
  const lower = msg.toLowerCase()
  if (lower.includes('user already registered') || lower.includes('already been registered')) {
    return 'An account with this email already exists. Sign in instead, or use a different email.'
  }
  if (lower.includes('invalid login credentials')) {
    return 'Invalid email or password. If you just signed up, confirm your email first, then try again.'
  }
  if (lower.includes('email not confirmed')) {
    return 'Confirm your email using the link we sent, then sign in.'
  }
  if (lower.includes('password should be at least') || lower.includes('weak password')) {
    return 'Password is too weak. Use at least 6 characters.'
  }
  if (lower.includes('unable to validate email') || lower.includes('invalid email')) {
    return 'Enter a valid email address.'
  }
  if (lower.includes('cannot reach supabase') || lower.includes('supabase')) {
    return msg
  }
  return msg || 'Request failed'
}

async function postJson(endpoint: string, body: unknown): Promise<{ ok: boolean; data: unknown; error?: string }> {
  const res = await fetch(apiUrl(endpoint), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(body),
  })
  const raw = await res.text()
  let data: unknown = null
  if (raw) {
    try {
      data = JSON.parse(raw)
    } catch {
      data = { error: raw.trim() }
    }
  }
  const errObj = data && typeof data === 'object' ? (data as Record<string, unknown>) : null
  const msg =
    (errObj?.error as string) ||
    (errObj?.message as string) ||
    (raw.trim() || res.statusText || `HTTP ${res.status}`)
  if (!res.ok) {
    return { ok: false, data, error: msg }
  }
  return { ok: true, data }
}

export type AuthUser = { id?: string; email?: string }

export async function signUp(email: string, password: string, metadata: Record<string, unknown> = {}): Promise<{
  user?: AuthUser
}> {
  clearAuthStorage()
  const { ok, data, error } = await postJson('/api/auth/signup', {
    email,
    password,
    data: metadata,
  })
  if (!ok) throw new Error(normalizeAuthError(error || 'Sign up failed'))
  applyAuthTokensFromPayload(data)
  const d = data as Record<string, unknown>
  return { user: (d.user as AuthUser) || undefined }
}

export async function signIn(email: string, password: string): Promise<{ user?: AuthUser }> {
  clearAuthStorage()
  const { ok, data, error } = await postJson('/api/auth/signin', { email, password })
  if (!ok) throw new Error(normalizeAuthError(error || 'Sign in failed'))
  applyAuthTokensFromPayload(data)
  const token = typeof localStorage !== 'undefined' ? localStorage.getItem(ACCESS_TOKEN_KEY) : null
  if (!token) {
    throw new Error('Sign-in succeeded but no access token was returned. Confirm your email if required, then try again.')
  }
  const d = data as Record<string, unknown>
  return { user: (d.user as AuthUser) || undefined }
}

export async function requestPasswordRecovery(email: string, redirectTo: string): Promise<void> {
  const { ok, data, error } = await postJson('/api/auth/recover', { email, redirect_to: redirectTo })
  if (!ok) throw new Error(error || 'Recovery request failed')
  const d = data as Record<string, unknown>
  if (d.success === false) throw new Error((d.error as string) || (d.message as string) || 'Request failed')
}

export async function updatePasswordWithRecoveryToken(accessToken: string, newPassword: string): Promise<void> {
  const res = await fetch(apiUrl('/api/auth/update-password'), {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`,
    },
    credentials: 'include',
    body: JSON.stringify({ password: newPassword }),
  })
  let data: unknown = null
  try {
    data = await res.json()
  } catch {
    data = null
  }
  if (!res.ok) {
    const errObj = data && typeof data === 'object' ? (data as Record<string, unknown>) : null
    throw new Error(
      (errObj?.error as string) || (errObj?.message as string) || res.statusText || `HTTP ${res.status}`,
    )
  }
}
