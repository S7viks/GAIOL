import { useCallback, useEffect, useMemo, useState } from 'react'
import { PageAlert, PageHeader, PageSection, PageStack } from '../components/layout/PageShell'
import { apiDelete, apiGet, apiPost, apiPut, ApiError, fetchHealthBody } from '../lib/api'
import { apiUrl } from '../lib/apiBase'
import { fetchSetupStatus, type SetupStatus } from '../lib/setupStatus'
import { useToast, useToastStore } from '../components/ui/Toast'
import type {
  GaiolKeyRow,
  PreferencesResponse,
  ProviderKeyRow,
  TenantModelRow,
  TenantModelsResponse,
} from '../types/api'
import { fetchAuthSession, loginHref } from '../lib/auth'
import { isLocalProvider, MODEL_SUGGESTIONS, providerMeta, PROVIDER_OPTIONS } from '../lib/providers'

export function SettingsPage() {
  const toast = useToast()
  const [strategy, setStrategy] = useState('balanced')
  const [budgetLimit, setBudgetLimit] = useState('')
  const [beamWidth, setBeamWidth] = useState('2')
  const [consensusMode, setConsensusMode] = useState('abtc')
  const [domain, setDomain] = useState('general')
  const [explorePaths, setExplorePaths] = useState(false)
  const [bootstrapped, setBootstrapped] = useState(false)
  const [dataLoading, setDataLoading] = useState(true)
  const [authDisabled, setAuthDisabled] = useState(false)
  const [authenticated, setAuthenticated] = useState(false)
  const [healthUnreachable, setHealthUnreachable] = useState(false)
  const [dbPingHint, setDbPingHint] = useState('')
  const [setup, setSetup] = useState<SetupStatus | null>(null)
  const [providerKeys, setProviderKeys] = useState<ProviderKeyRow[]>([])
  const [newProvider, setNewProvider] = useState<string>('openrouter')
  const [newApiKey, setNewApiKey] = useState('')
  const [newBaseUrl, setNewBaseUrl] = useState('http://localhost:11434')
  const [savingKey, setSavingKey] = useState(false)
  const [oneTimeGaiolKey, setOneTimeGaiolKey] = useState<string | null>(null)
  const [gaiolKeys, setGaiolKeys] = useState<GaiolKeyRow[]>([])
  const [newGaiolKeyName, setNewGaiolKeyName] = useState('default')
  const [creatingGaiolKey, setCreatingGaiolKey] = useState(false)
  const [revokingGaiolId, setRevokingGaiolId] = useState<string | null>(null)
  const [tenantModels, setTenantModels] = useState<TenantModelRow[]>([])
  const [modelProvider, setModelProvider] = useState('openrouter')
  const [modelIdInput, setModelIdInput] = useState('')
  const [modelDisplayName, setModelDisplayName] = useState('')
  const [modelSaving, setModelSaving] = useState(false)
  const [removingModel, setRemovingModel] = useState<string | null>(null)

  const savedProviderValues = useMemo(
    () => providerKeys.map((k) => k.provider).filter((p): p is string => !!p),
    [providerKeys],
  )

  const loadProviderKeys = useCallback(async () => {
    const raw = await apiGet('/api/settings/provider-keys')
    setProviderKeys(Array.isArray(raw) ? (raw as ProviderKeyRow[]) : [])
  }, [])

  const loadGaiolKeys = useCallback(async () => {
    const raw = await apiGet('/api/gaiol-keys')
    setGaiolKeys(Array.isArray(raw) ? (raw as GaiolKeyRow[]) : [])
  }, [])

  const loadTenantModels = useCallback(async () => {
    try {
      const raw = (await apiGet('/api/settings/models')) as TenantModelsResponse
      setTenantModels(Array.isArray(raw.models) ? raw.models : [])
    } catch {
      setTenantModels([])
    }
  }, [])

  useEffect(() => {
    if (savedProviderValues.length === 0) return
    if (!savedProviderValues.includes(modelProvider)) {
      setModelProvider(savedProviderValues[0]!)
    }
  }, [savedProviderValues, modelProvider])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const health = await fetchHealthBody()
        if (!cancelled) {
          setAuthDisabled(health.authDisabled)
          setHealthUnreachable(!health.authDisabled && health.ok && !health.databaseReachable)
          setDbPingHint(health.databasePingError ?? 'check SUPABASE_URL in .env')
        }

        if (!health.authDisabled) {
          const session = await fetchAuthSession()
          if (!cancelled) setAuthenticated(session.authenticated)
        } else if (!cancelled) {
          setAuthenticated(false)
        }
      } catch (e) {
        if (!cancelled) {
          useToastStore.getState().add('error', e instanceof ApiError ? e.message : String(e))
        }
      } finally {
        if (!cancelled) setBootstrapped(true)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!bootstrapped) return
    let cancelled = false
    ;(async () => {
      setDataLoading(true)
      try {
        const [setupResult, prefsResult] = await Promise.allSettled([
          fetchSetupStatus(apiGet),
          apiGet('/api/settings/preferences'),
        ])
        if (!cancelled && setupResult.status === 'fulfilled') {
          setSetup(setupResult.value)
        }
        if (!cancelled && prefsResult.status === 'fulfilled') {
          const p = prefsResult.value as PreferencesResponse
          setStrategy(p.strategy ?? 'balanced')
          setBudgetLimit(p.budget_limit != null ? String(p.budget_limit) : '')
          setBeamWidth(p.beam_width != null ? String(p.beam_width) : '2')
          setConsensusMode(p.consensus_mode ?? 'abtc')
          setDomain(p.domain ?? 'general')
          setExplorePaths(!!p.explore_paths)
        } else if (!cancelled && prefsResult.status === 'rejected') {
          const err = prefsResult.reason
          if (!(err instanceof ApiError && err.status === 401)) {
            useToastStore.getState().add(
              'error',
              err instanceof ApiError ? err.message : String(err),
            )
          }
        }

        const keyResults = await Promise.allSettled([
          loadProviderKeys(),
          loadGaiolKeys(),
          loadTenantModels(),
        ])
        for (const result of keyResults) {
          if (result.status === 'rejected' && !cancelled) {
            const err = result.reason
            if (!(err instanceof ApiError && err.status === 401)) {
              useToastStore.getState().add(
                'error',
                err instanceof ApiError ? err.message : String(err),
              )
            }
          }
        }
      } finally {
        if (!cancelled) setDataLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [bootstrapped, loadProviderKeys, loadGaiolKeys, loadTenantModels])

  function parseBudgetLimit(): number | null {
    const t = budgetLimit.trim()
    if (!t) return null
    const n = Number(t)
    if (Number.isNaN(n) || n < 0) return null
    return n
  }

  async function savePrefs() {
    const budget = parseBudgetLimit()
    if (budgetLimit.trim() && budget === null) {
      toast.error('Budget must be a non-negative number or empty')
      return
    }
    const bw = Number(beamWidth)
    if (!Number.isInteger(bw) || bw < 1) {
      toast.error('Beam width must be a whole number of at least 1')
      return
    }
    const dom = domain.trim()
    if (!dom) {
      toast.error('Domain is required')
      return
    }
    try {
      await apiPut('/api/settings/preferences', {
        strategy,
        default_model_id: '',
        budget_limit: budget,
        beam_width: bw,
        consensus_mode: consensusMode,
        domain: dom,
        explore_paths: explorePaths,
      })
      toast.success('Preferences saved')
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    }
  }

  async function addProviderKey() {
    const meta = providerMeta(newProvider)
    const trimmed = newApiKey.trim()
    if (meta?.requiresApiKey !== false && !trimmed) {
      toast.error('Paste an API key first')
      return
    }
    setSavingKey(true)
    setOneTimeGaiolKey(null)
    try {
      const payload: Record<string, string> = { provider: newProvider }
      if (trimmed) payload.api_key = trimmed
      if (isLocalProvider(newProvider)) {
        payload.base_url = newBaseUrl.trim() || meta?.defaultBaseUrl || 'http://localhost:11434'
      }
      const data = (await apiPost('/api/settings/provider-keys', payload)) as Record<string, unknown>
      setNewApiKey('')
      toast.success('Provider key saved')
      if (typeof data.gaiol_api_key === 'string' && data.gaiol_api_key) {
        setOneTimeGaiolKey(data.gaiol_api_key)
        toast.info('A GAIOL API key was created — copy it from the yellow box below (shown once).')
      }
      await loadProviderKeys()
      await loadGaiolKeys()
      await loadTenantModels()
      setSetup(await fetchSetupStatus(apiGet))
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        toast.error('Sign in required — open Login from the top bar.')
        setAuthenticated(false)
      } else {
        toast.error(e instanceof ApiError ? e.message : String(e))
      }
    } finally {
      setSavingKey(false)
    }
  }

  async function removeProviderKey(provider: string) {
    try {
      await apiDelete(`/api/settings/provider-keys?provider=${encodeURIComponent(provider)}`)
      toast.success(`Removed ${provider}`)
      await loadProviderKeys()
      await loadTenantModels()
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    }
  }

  async function upsertTenantModel(provider_key: string, model_id: string, display_name?: string) {
    setModelSaving(true)
    try {
      await apiPost('/api/settings/models', {
        provider_key,
        model_id: model_id.trim(),
        display_name: (display_name ?? model_id).trim(),
      })
      toast.success('Model saved')
      setModelIdInput('')
      setModelDisplayName('')
      await loadTenantModels()
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setModelSaving(false)
    }
  }

  async function removeTenantModel(provider_key: string, model_id: string) {
    const key = `${provider_key}:${model_id}`
    setRemovingModel(key)
    try {
      await apiDelete(
        `/api/settings/models?provider_key=${encodeURIComponent(provider_key)}&model_id=${encodeURIComponent(model_id)}`,
      )
      toast.success('Model removed')
      await loadTenantModels()
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setRemovingModel(null)
    }
  }

  async function createGaiolKey() {
    setCreatingGaiolKey(true)
    setOneTimeGaiolKey(null)
    try {
      const data = (await apiPost('/api/gaiol-keys', { name: newGaiolKeyName.trim() || 'default' })) as {
        api_key?: string
      }
      if (data.api_key) {
        setOneTimeGaiolKey(data.api_key)
        toast.success('GAIOL API key created — copy it now (shown once)')
      } else {
        toast.success('GAIOL API key created')
      }
      await loadGaiolKeys()
      setSetup(await fetchSetupStatus(apiGet))
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setCreatingGaiolKey(false)
    }
  }

  async function revokeGaiolKey(id: string) {
    setRevokingGaiolId(id)
    try {
      await apiDelete(`/api/gaiol-keys/${encodeURIComponent(id)}`)
      toast.success('GAIOL key revoked')
      await loadGaiolKeys()
      setSetup(await fetchSetupStatus(apiGet))
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setRevokingGaiolId(null)
    }
  }

  const curlSnippet = `curl -X POST ${apiUrl('/v1/chat')} \\
  -H "Authorization: Bearer YOUR_GAIOL_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"prompt":"Hello from curl","max_tokens":200}'`

  async function copyText(text: string) {
    try {
      await navigator.clipboard.writeText(text)
      toast.success('Copied')
    } catch {
      toast.error('Could not copy — select and copy manually')
    }
  }

  const canManageKeys = authDisabled ? false : authenticated

  const setupIncomplete =
    setup &&
    setup.setup_complete === false &&
    ((setup.providers_connected ?? 0) === 0 || (setup.gaiol_keys_count ?? 0) === 0)

  return (
    <div className="page">
      <PageHeader
        title="Settings"
        description="Preferences, provider keys, tenant models, and GAIOL API keys."
      />

      {authDisabled && (
        <PageAlert variant="warn" title="Local no-auth mode">
          Provider keys, GAIOL keys, and tenant models are stored in the database and require Supabase auth.
          Set provider keys in <code>.env</code> (e.g. <code>OPENROUTER_API_KEY</code>) and restart the server,
          or unset <code>GAIOL_DISABLE_AUTH</code> and configure Supabase per QUICKSTART.md.
        </PageAlert>
      )}

      {!authDisabled && bootstrapped && !authenticated && (
        <PageAlert variant="warn" title="Sign in required">
          Provider keys and GAIOL keys are saved per account.{' '}
          <a href={loginHref()}>Sign in</a> or create an account, then return here.
        </PageAlert>
      )}

      {!authDisabled && bootstrapped && authenticated && healthUnreachable && (
        <PageAlert variant="warn" title="Database unreachable">
          The API cannot reach Supabase ({dbPingHint}). Saving keys will fail until the Project URL in{' '}
          <code>.env</code> is correct and the server is restarted.
        </PageAlert>
      )}

      {setupIncomplete && (
        <PageAlert variant="warn" title="Setup incomplete" actionTo="/onboarding" actionLabel="Complete setup">
          {(setup?.providers_connected ?? 0) === 0 && 'No provider keys connected. '}
          {(setup?.gaiol_keys_count ?? 0) === 0 && 'No GAIOL API key yet.'}
        </PageAlert>
      )}

      {dataLoading && bootstrapped && <div className="skeleton skeleton--block" aria-hidden />}

      {bootstrapped && !dataLoading && (
        <PageStack>
          <div className="page-grid page-grid--settings">
            <PageSection
              title="Preferences"
              subtitle="Routing and orchestration for your account. Saved here and applied to Chat and API requests."
            >
              <div className="form-field">
                <label htmlFor="strategy">Routing strategy</label>
                <select
                  id="strategy"
                  value={strategy}
                  onChange={(e) => setStrategy(e.target.value)}
                  style={{ width: '100%', maxWidth: 320 }}
                >
                  <option value="balanced">balanced</option>
                  <option value="lowest_cost">lowest_cost</option>
                  <option value="highest_quality">highest_quality</option>
                  <option value="free_only">free_only</option>
                  <option value="beam">beam</option>
                </select>
              </div>
              <div className="form-field">
                <label htmlFor="beam-width">Beam width</label>
                <input
                  id="beam-width"
                  type="number"
                  min={1}
                  max={16}
                  value={beamWidth}
                  onChange={(e) => setBeamWidth(e.target.value)}
                />
              </div>
              <div className="form-field">
                <label htmlFor="consensus">Consensus mode</label>
                <select
                  id="consensus"
                  value={consensusMode}
                  onChange={(e) => setConsensusMode(e.target.value)}
                  style={{ width: '100%', maxWidth: 320 }}
                >
                  <option value="abtc">abtc</option>
                  <option value="uniform">uniform</option>
                  <option value="static">static</option>
                </select>
              </div>
              <div className="form-field">
                <label htmlFor="domain">Domain tag</label>
                <input
                  id="domain"
                  value={domain}
                  onChange={(e) => setDomain(e.target.value)}
                  placeholder="general"
                />
              </div>
              <div className="form-field">
                <label htmlFor="explore-paths" style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <input
                    id="explore-paths"
                    type="checkbox"
                    checked={explorePaths}
                    onChange={(e) => setExplorePaths(e.target.checked)}
                  />
                  Explore paths (parallel candidate paths in orchestration)
                </label>
              </div>
              <div className="form-field">
                <label htmlFor="budget">Monthly budget (USD)</label>
                <input
                  id="budget"
                  type="number"
                  min={0}
                  step="0.01"
                  value={budgetLimit}
                  onChange={(e) => setBudgetLimit(e.target.value)}
                  placeholder="optional"
                />
              </div>
              <button type="button" className="btn" onClick={() => void savePrefs()}>
                Save preferences
              </button>
            </PageSection>

            <PageSection title="GAIOL API keys" subtitle="Bearer token for POST /v1/chat">
              {authDisabled ? (
                <p className="empty-state">Not available in local no-auth mode.</p>
              ) : !authenticated ? (
                <p className="empty-state">
                  <a href={loginHref()}>Sign in</a> to create GAIOL API keys.
                </p>
              ) : (
                <>
              {(oneTimeGaiolKey || gaiolKeys.length === 0) && (
                <div className="form-field">
                  <label htmlFor="gk-name">Key name</label>
                  <input
                    id="gk-name"
                    value={newGaiolKeyName}
                    onChange={(e) => setNewGaiolKeyName(e.target.value)}
                    placeholder="default"
                  />
                </div>
              )}
              <button
                type="button"
                className="btn btn--secondary"
                disabled={creatingGaiolKey}
                onClick={() => void createGaiolKey()}
              >
                {creatingGaiolKey ? 'Creating…' : gaiolKeys.length === 0 ? 'Create key' : 'Create another'}
              </button>
              {gaiolKeys.length > 0 && (
                <div className="table-wrap">
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>Name</th>
                        <th>Created</th>
                        <th>Last used</th>
                        <th />
                      </tr>
                    </thead>
                    <tbody>
                      {gaiolKeys.map((row) => (
                        <tr key={row.id ?? row.name}>
                          <td>{row.name ?? '—'}</td>
                          <td className="mono">{row.created_at ? row.created_at.slice(0, 10) : '—'}</td>
                          <td className="mono">{row.last_used_at ? row.last_used_at.slice(0, 10) : 'never'}</td>
                          <td>
                            {row.id && (
                              <button
                                type="button"
                                className="btn btn--secondary btn--sm"
                                disabled={revokingGaiolId === row.id}
                                onClick={() => void revokeGaiolKey(row.id!)}
                              >
                                {revokingGaiolId === row.id ? 'Revoking…' : 'Revoke'}
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
              <details className="ui-details" style={{ marginTop: 10 }}>
                <summary>curl example</summary>
                <div className="ui-details__body">
                  <pre className="mono-block">{curlSnippet}</pre>
                  <button type="button" className="btn btn--secondary btn--sm" onClick={() => void copyText(curlSnippet)}>
                    Copy
                  </button>
                </div>
              </details>
                </>
              )}
            </PageSection>
          </div>

          <PageSection title="Provider API keys">
            {authDisabled ? (
              <p className="empty-state">
                Configure keys via environment variables and restart the Go server.
              </p>
            ) : !authenticated ? (
              <p className="empty-state">
                <a href={loginHref()}>Sign in</a> to save provider keys to your account.
              </p>
            ) : (
              <>
            {oneTimeGaiolKey && (
              <div className="alert alert--warn page-alert" style={{ marginBottom: 10 }}>
                <div className="page-alert__body">
                  <strong>GAIOL key (show once)</strong>
                  <div className="page-alert__content">
                    <code className="inline-code">{oneTimeGaiolKey}</code>
                    <div className="btn-row">
                      <button type="button" className="btn btn--sm" onClick={() => void copyText(oneTimeGaiolKey)}>
                        Copy
                      </button>
                      <button type="button" className="btn btn--secondary btn--sm" onClick={() => setOneTimeGaiolKey(null)}>
                        Dismiss
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            <div className="page-grid page-grid--sidebar">
              <div>
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
                    />
                    {providerMeta(newProvider)?.helpUrl && (
                      <p className="table-meta">
                        <a href={providerMeta(newProvider)!.helpUrl} target="_blank" rel="noreferrer">
                          Get a {providerMeta(newProvider)!.label} key
                        </a>
                      </p>
                    )}
                  </div>
                )}
                <button type="button" className="btn" disabled={savingKey || !canManageKeys} onClick={() => void addProviderKey()}>
                  {savingKey ? 'Saving…' : isLocalProvider(newProvider) ? 'Connect Ollama' : 'Save key'}
                </button>
              </div>

              {providerKeys.length > 0 && (
                <div className="table-wrap table-wrap--flush">
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>Provider</th>
                        <th>Hint</th>
                        <th />
                      </tr>
                    </thead>
                    <tbody>
                      {providerKeys.map((row) => (
                        <tr key={row.id ?? row.provider}>
                          <td>{row.provider}</td>
                          <td className="mono">{row.key_hint ?? '—'}</td>
                          <td>
                            <button
                              type="button"
                              className="btn btn--secondary btn--sm"
                              onClick={() => row.provider && void removeProviderKey(row.provider)}
                            >
                              Remove
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
              </>
            )}
          </PageSection>

          <PageSection title="Tenant models">
            {authDisabled ? (
              <p className="empty-state">Models are loaded from env provider keys in local no-auth mode.</p>
            ) : savedProviderValues.length === 0 ? (
              <p className="empty-state">Save a provider key first.</p>
            ) : (
              <>
                <div className="page-grid page-grid--sidebar">
                  <div>
                    <div className="form-field">
                      <label htmlFor="tm-provider">Provider</label>
                      <select
                        id="tm-provider"
                        value={modelProvider}
                        onChange={(e) => setModelProvider(e.target.value)}
                        disabled={modelSaving}
                      >
                        {savedProviderValues.map((pv) => (
                          <option key={pv} value={pv}>
                            {pv}
                          </option>
                        ))}
                      </select>
                    </div>
                    {(MODEL_SUGGESTIONS[modelProvider] ?? []).length > 0 && (
                      <div className="chip-row">
                        {(MODEL_SUGGESTIONS[modelProvider] ?? []).map((s) => (
                          <button
                            key={s.model_id}
                            type="button"
                            className="btn btn--secondary btn--sm"
                            disabled={modelSaving}
                            onClick={() => void upsertTenantModel(modelProvider, s.model_id, s.display_name)}
                          >
                            {s.display_name}
                          </button>
                        ))}
                      </div>
                    )}
                    <div className="form-field">
                      <label htmlFor="tm-mid">Model id</label>
                      <input
                        id="tm-mid"
                        value={modelIdInput}
                        onChange={(e) => setModelIdInput(e.target.value)}
                        placeholder="anthropic/claude-3.5-sonnet"
                        disabled={modelSaving}
                      />
                    </div>
                    <div className="form-field">
                      <label htmlFor="tm-name">Display name</label>
                      <input
                        id="tm-name"
                        value={modelDisplayName}
                        onChange={(e) => setModelDisplayName(e.target.value)}
                        disabled={modelSaving}
                      />
                    </div>
                    <button
                      type="button"
                      className="btn btn--secondary"
                      disabled={modelSaving || !modelIdInput.trim()}
                      onClick={() => void upsertTenantModel(modelProvider, modelIdInput, modelDisplayName || undefined)}
                    >
                      {modelSaving ? 'Saving…' : 'Add model'}
                    </button>
                  </div>

                  {tenantModels.length > 0 ? (
                    <div className="table-wrap table-wrap--flush">
                      <table className="data-table">
                        <thead>
                          <tr>
                            <th>Provider</th>
                            <th>Model id</th>
                            <th>Name</th>
                            <th />
                          </tr>
                        </thead>
                        <tbody>
                          {tenantModels.map((row) => {
                            const pk = row.provider_key ?? ''
                            const mid = row.model_id ?? ''
                            const rmKey = `${pk}:${mid}`
                            return (
                              <tr key={row.id ?? rmKey}>
                                <td>{pk}</td>
                                <td className="mono">{mid}</td>
                                <td>{row.display_name ?? '—'}</td>
                                <td>
                                  {pk && mid && (
                                    <button
                                      type="button"
                                      className="btn btn--secondary btn--sm"
                                      disabled={removingModel === rmKey}
                                      onClick={() => void removeTenantModel(pk, mid)}
                                    >
                                      {removingModel === rmKey ? '…' : 'Remove'}
                                    </button>
                                  )}
                                </td>
                              </tr>
                            )
                          })}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <p className="empty-state">No tenant models yet.</p>
                  )}
                </div>
              </>
            )}
          </PageSection>
        </PageStack>
      )}
    </div>
  )
}
