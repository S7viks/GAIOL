import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader, PageSection } from '../components/layout/PageShell'
import type { ProviderKeyRow } from '../types/api'
import { apiGet, apiPost, ApiError, fetchHealthBody } from '../lib/api'
import { fetchAuthSession, getAccessToken, loginHref } from '../lib/auth'
import { useToast } from '../components/ui/Toast'
import { isLocalProvider, providerMeta, PROVIDER_OPTIONS } from '../lib/providers'

type GaiolKeyRow = {
  id?: string
  name?: string
  created_at?: string
  last_used_at?: string | null
}

export function OnboardingPage() {
  const toast = useToast()
  const [authDisabled, setAuthDisabled] = useState(false)
  const [authLoading, setAuthLoading] = useState(true)
  const [authenticated, setAuthenticated] = useState(false)

  const [providerKeys, setProviderKeys] = useState<ProviderKeyRow[]>([])
  const [gaiolKeys, setGaiolKeys] = useState<GaiolKeyRow[]>([])

  const [providerLoading, setProviderLoading] = useState(false)
  const [providerSaveError, setProviderSaveError] = useState<string | null>(null)
  const [encryptionKeyMissing, setEncryptionKeyMissing] = useState(false)
  const [newProvider, setNewProvider] = useState<string>('openrouter')
  const [newApiKey, setNewApiKey] = useState<string>('')
  const [newBaseUrl, setNewBaseUrl] = useState<string>('http://localhost:11434')

  const [ensureLoading, setEnsureLoading] = useState(false)
  const [gaiolKeyCreated, setGaiolKeyCreated] = useState<boolean | null>(null)
  const [gaiolKeySecret, setGaiolKeySecret] = useState<string | null>(null)

  const [step, setStep] = useState<number>(0)

  const loadProviderKeys = useCallback(async () => {
    const raw = await apiGet('/api/settings/provider-keys')
    const list = Array.isArray(raw) ? (raw as ProviderKeyRow[]) : []
    setProviderKeys(list)
    return list
  }, [])

  const loadGaiolKeys = useCallback(async () => {
    const raw = await apiGet('/api/gaiol-keys')
    const list = Array.isArray(raw) ? (raw as GaiolKeyRow[]) : []
    setGaiolKeys(list)
    return list
  }, [])

  const refreshAll = useCallback(async () => {
    await loadProviderKeys()
    await loadGaiolKeys()
  }, [loadGaiolKeys, loadProviderKeys])

  const refreshAuthState = useCallback(async () => {
    const h = await fetchHealthBody()
    setAuthDisabled(!!h.authDisabled)
    setEncryptionKeyMissing(h.encryptionKeyConfigured === false)

    if (h.authDisabled) {
      setAuthenticated(false)
      setProviderKeys([])
      setGaiolKeys([])
      setStep(0)
      return false
    }

    const s = await fetchAuthSession()
    const signedIn = !!s.authenticated && !!getAccessToken()?.trim()
    setAuthenticated(signedIn)
    return signedIn
  }, [])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setAuthLoading(true)
      setAuthDisabled(false)
      setAuthenticated(false)
      try {
        const signedIn = await refreshAuthState()
        if (cancelled) return
        if (signedIn) await refreshAll()
      } catch (e) {
        if (!cancelled) {
          setAuthenticated(false)
          toast.error(e instanceof ApiError ? e.message : String(e))
        }
      } finally {
        if (!cancelled) setAuthLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [refreshAll, refreshAuthState, toast])

  useEffect(() => {
    const onFocus = () => {
      void refreshAuthState().then((signedIn) => {
        if (signedIn) void refreshAll()
      })
    }
    window.addEventListener('focus', onFocus)
    return () => window.removeEventListener('focus', onFocus)
  }, [refreshAll, refreshAuthState])

  useEffect(() => {
    if (authDisabled || !authenticated) return
    if (providerKeys.length === 0) return
    if (gaiolKeyCreated !== null) return

    let cancelled = false
    ;(async () => {
      setEnsureLoading(true)
      setProviderSaveError(null)
      try {
        const data = (await apiPost('/api/settings/gaiol-key/ensure', {})) as {
          gaiol_api_key_created?: boolean
          gaiol_api_key?: string
        }
        if (cancelled) return
        setGaiolKeyCreated(!!data.gaiol_api_key_created)
        if (typeof data.gaiol_api_key === 'string' && data.gaiol_api_key) {
          setGaiolKeySecret(data.gaiol_api_key)
        } else {
          setGaiolKeySecret(null)
        }
      } catch (e) {
        if (!cancelled) {
          const msg = e instanceof ApiError ? e.message : String(e)
          setProviderSaveError(msg)
          toast.error(msg)
        }
      } finally {
        if (!cancelled) setEnsureLoading(false)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [authenticated, authDisabled, gaiolKeyCreated, providerKeys.length, toast])

  useEffect(() => {
    if (authDisabled || !authenticated) {
      setStep(0)
      return
    }
    setStep(providerKeys.length > 0 ? 1 : 0)
  }, [authDisabled, authenticated, providerKeys.length])

  async function addProviderKey() {
    setProviderSaveError(null)
    if (!getAccessToken()?.trim()) {
      setAuthenticated(false)
      setProviderSaveError('Sign in to save provider keys.')
      return
    }
    const meta = providerMeta(newProvider)
    const trimmed = newApiKey.trim()
    if (meta?.requiresApiKey !== false && !trimmed) {
      toast.error('Paste an API key first')
      return
    }
    setProviderLoading(true)
    try {
      const payload: Record<string, string> = { provider: newProvider }
      if (trimmed) payload.api_key = trimmed
      if (isLocalProvider(newProvider)) {
        payload.base_url = newBaseUrl.trim() || meta?.defaultBaseUrl || 'http://localhost:11434'
      }
      const data = (await apiPost('/api/settings/provider-keys', payload)) as Record<string, unknown>

      setNewApiKey('')
      await refreshAll()

      if (typeof data.gaiol_api_key_created === 'boolean' && data.gaiol_api_key_created) {
        if (typeof data.gaiol_api_key === 'string' && data.gaiol_api_key) {
          setGaiolKeyCreated(true)
          setGaiolKeySecret(data.gaiol_api_key)
        }
      }

      toast.success('Provider key saved')
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : String(e)
      const code = e instanceof ApiError ? e.code : undefined
      const status = e instanceof ApiError ? e.status : 0
      if (status === 401 || /authorization required/i.test(msg)) {
        setAuthenticated(false)
        setProviderSaveError('Your session expired. Sign in again, then save your provider key.')
      } else if (code === 'encryption_key_missing') {
        setEncryptionKeyMissing(true)
        setProviderSaveError(msg)
      } else {
        setProviderSaveError(msg)
      }
      toast.error(msg)
    } finally {
      setProviderLoading(false)
    }
  }

  if (authLoading) return <div className="page"><div className="skeleton skeleton--block" /></div>

  if (authDisabled) {
    return (
      <div className="page">
        <PageHeader
          title="Setup"
          description="Server is in auth_disabled mode. Configure providers via environment variables."
        />
      </div>
    )
  }

  if (!authenticated) {
    return (
      <div className="page">
        <PageHeader title="Setup" description="Sign in to connect your model providers." />
        <PageSection>
          <p className="page-shell__desc" style={{ marginBottom: 12 }}>
            Provider keys are stored per account. Create an account or sign in, then return here to connect OpenRouter,
            Gemini, or other providers.
          </p>
          <a className="btn" href={loginHref()}>
            Sign in
          </a>{' '}
          <Link className="btn btn--ghost" to="/signup">
            Create account
          </Link>
        </PageSection>
      </div>
    )
  }

  return (
    <div className="page">
      <PageHeader
        title="Setup"
        description="Connect at least one provider. Dashboard chat uses your login — no API key required to get started."
      />

      <PageSection>
        <div className="onboarding-steps">
          <span className="badge">1 · Provider keys</span>
          <span className="badge">2 · Ready</span>
        </div>

        {encryptionKeyMissing && (
          <div className="alert alert--err" style={{ marginBottom: 16 }} role="alert">
            <strong>Server configuration required.</strong> The API host is missing{' '}
            <code>GAIOL_ENCRYPTION_KEY</code> (needed to encrypt provider keys). On Render: open your GAIOL service →
            Environment → add <code>GAIOL_ENCRYPTION_KEY</code> with a 64-character hex value from{' '}
            <code>openssl rand -hex 32</code>, then redeploy. Saving provider keys will not work until this is set.
          </div>
        )}

        {providerSaveError && (
          <div className="alert alert--err" style={{ marginBottom: 16 }}>
            {providerSaveError}
          </div>
        )}

        {step === 0 && (
          <section>
            <h2 className="page-section__title">Connect a provider</h2>
            <p className="page-shell__desc" style={{ marginBottom: 12 }}>
              Add your OpenRouter, Gemini, Groq, Hugging Face, or other provider key. Keys are encrypted per account.
              A default model is registered automatically when you save.
            </p>

            <div className="form-field">
              <label htmlFor="pk-provider">Provider</label>
              <select
                id="pk-provider"
                value={newProvider}
                onChange={(e) => {
                  const v = e.target.value
                  setNewProvider(v)
                  const m = providerMeta(v)
                  if (m?.defaultBaseUrl) setNewBaseUrl(m.defaultBaseUrl)
                }}
                style={{ width: '100%', maxWidth: 320 }}
                disabled={providerLoading}
              >
                {PROVIDER_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>

            {isLocalProvider(newProvider) ? (
              <div className="form-field">
                <label htmlFor="pk-base">Ollama base URL</label>
                <input
                  id="pk-base"
                  value={newBaseUrl}
                  onChange={(e) => setNewBaseUrl(e.target.value)}
                  placeholder="http://localhost:11434"
                  disabled={providerLoading}
                />
              </div>
            ) : (
              <div className="form-field">
                <label htmlFor="pk-key">API key</label>
                <input
                  id="pk-key"
                  type="password"
                  autoComplete="off"
                  value={newApiKey}
                  onChange={(e) => setNewApiKey(e.target.value)}
                  placeholder={providerMeta(newProvider)?.placeholderKeyHint ?? 'Paste key'}
                  disabled={providerLoading}
                />
              </div>
            )}

            <button
              type="button"
              className="btn"
              disabled={providerLoading || encryptionKeyMissing}
              onClick={() => void addProviderKey()}
            >
              {providerLoading ? 'Saving…' : isLocalProvider(newProvider) ? 'Connect Ollama' : 'Save provider key'}
            </button>

            <div style={{ marginTop: 20 }}>
              <h3 style={{ fontSize: '0.9rem', marginBottom: 8 }}>Saved keys</h3>
              {providerKeys.length === 0 ? (
                <p className="page-shell__desc" style={{ color: 'var(--text-secondary)' }}>
                  No provider keys saved yet.
                </p>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>Provider</th>
                      <th>Hint</th>
                    </tr>
                  </thead>
                  <tbody>
                    {providerKeys.map((row) => (
                      <tr key={row.id ?? row.provider}>
                        <td>{row.provider}</td>
                        <td>{row.key_hint ?? '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </section>
        )}

        {step === 1 && (
          <section>
            <h2 className="page-section__title">You&apos;re ready</h2>
            <p className="page-shell__desc" style={{ marginBottom: 16 }}>
              {providerKeys.length} provider{providerKeys.length === 1 ? '' : 's'} connected. Open Chat to run prompts
              through orchestration using your saved keys.
            </p>

            <Link to="/chat" className="btn">
              Open Chat
            </Link>
            <Link to="/home" className="btn btn--secondary" style={{ marginLeft: 12 }}>
              Go to Home
            </Link>

            {(ensureLoading || gaiolKeySecret || gaiolKeys.length > 0) && (
              <div className="panel" style={{ marginTop: 24 }}>
                <h3 style={{ fontSize: '0.9rem', marginBottom: 8 }}>For apps and scripts (optional)</h3>
                <p className="page-shell__desc" style={{ marginBottom: 12, fontSize: '0.9rem' }}>
                  External integrations use a GAIOL API key with <code>POST /v1/chat</code>. Manage keys in{' '}
                  <Link to="/settings">Settings</Link>.
                </p>
                {ensureLoading ? (
                  <div className="skeleton skeleton--block" />
                ) : gaiolKeySecret ? (
                  <div className="alert alert--warn" style={{ display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'center' }}>
                    <div style={{ flex: '1 1 260px' }}>
                      <strong>New GAIOL API key</strong> — copy now (shown once): <code>{gaiolKeySecret}</code>
                    </div>
                    <button
                      type="button"
                      className="btn btn--secondary"
                      onClick={async () => {
                        try {
                          await navigator.clipboard.writeText(gaiolKeySecret)
                          toast.success('Copied')
                        } catch {
                          toast.error('Could not copy')
                        }
                      }}
                    >
                      Copy
                    </button>
                  </div>
                ) : gaiolKeys.length > 0 ? (
                  <p className="page-shell__desc" style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
                    Your tenant has {gaiolKeys.length} GAIOL API key{gaiolKeys.length === 1 ? '' : 's'}. Create or
                    revoke keys in Settings.
                  </p>
                ) : null}
              </div>
            )}
          </section>
        )}
      </PageSection>
    </div>
  )
}
