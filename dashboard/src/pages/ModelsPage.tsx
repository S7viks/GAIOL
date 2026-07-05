import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  FRONTIER_MODELS,
  FRONTIER_VENDORS,
  type FrontierTier,
  type FrontierVendor,
  frontierModelsForVendor,
} from '../lib/frontier-models'
import { PageHeader, PageSection } from '../components/layout/PageShell'
import { apiGet, ApiError } from '../lib/api'
import { useToast } from '../components/ui/Toast'
import type { ModelRow, ModelsListResponse } from '../types/api'

type TierFilter = FrontierTier | 'all'

function tierLabel(tier: FrontierTier): string {
  if (tier === 'frontier') return 'Frontier'
  if (tier === 'standard') return 'Standard'
  return 'Free'
}

export function ModelsPage() {
  const toast = useToast()
  const [searchParams] = useSearchParams()
  const initialQ = searchParams.get('q') ?? ''

  const [serverModels, setServerModels] = useState<ModelRow[]>([])
  const [filter, setFilter] = useState(initialQ)
  const [vendor, setVendor] = useState<FrontierVendor>('All')
  const [tier, setTier] = useState<TierFilter>('all')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setFilter(initialQ)
  }, [initialQ])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      try {
        const data = (await apiGet('/api/models')) as ModelsListResponse
        if (!cancelled) setServerModels(Array.isArray(data.models) ? data.models : [])
      } catch (e) {
        if (!cancelled) {
          toast.error(e instanceof ApiError ? e.message : String(e))
          setServerModels([])
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [toast])

  const catalog = useMemo(() => {
    const q = filter.trim().toLowerCase()
    let list = frontierModelsForVendor(vendor, tier)
    if (q) {
      list = list.filter((m) => {
        const hay = [m.model_id, m.display_name, m.vendor, m.provider_key, ...m.tags].join(' ').toLowerCase()
        return hay.includes(q)
      })
    }
    return list
  }, [filter, tier, vendor])

  const serverFiltered = useMemo(() => {
    const q = filter.trim().toLowerCase()
    if (!q) return serverModels
    return serverModels.filter((m) => {
      const hay = [m.id, m.display_name, m.model_name, m.provider, ...(m.tags ?? [])]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return hay.includes(q)
    })
  }, [filter, serverModels])

  return (
    <div className="page">
      <PageHeader
        title="Models"
        description={
          <>
            {FRONTIER_MODELS.length} frontier and standard models in the GAIOL catalog. Connect a provider in{' '}
            <Link to="/settings">Settings</Link> to route to any of them. Live server registry:{' '}
            {loading ? '…' : serverModels.length} models.
          </>
        }
      />

      <PageSection title="Catalog">
        <div className="model-toolbar">
          <div className="form-field">
            <label htmlFor="filter">Search</label>
            <input
              id="filter"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              placeholder="Claude, GPT-4.1, gemini, deepseek…"
            />
          </div>
        </div>

        <div className="model-filter-row" style={{ marginBottom: 12 }}>
          {FRONTIER_VENDORS.map((v) => (
            <button
              key={v}
              type="button"
              className={`filter-chip${vendor === v ? ' filter-chip--active' : ''}`}
              onClick={() => setVendor(v)}
            >
              {v}
            </button>
          ))}
        </div>

        <div className="model-filter-row" style={{ marginBottom: 20 }}>
          {(['all', 'frontier', 'standard', 'free'] as const).map((t) => (
            <button
              key={t}
              type="button"
              className={`filter-chip${tier === t ? ' filter-chip--active' : ''}`}
              onClick={() => setTier(t)}
            >
              {t === 'all' ? 'All tiers' : tierLabel(t)}
            </button>
          ))}
        </div>

        <p className="table-meta" style={{ marginBottom: 16 }}>
          Showing {catalog.length} catalog models
        </p>

        <div className="model-grid">
          {catalog.map((m) => (
            <article key={`${m.provider_key}:${m.model_id}`} className="model-card">
              <div className="model-card__head">
                <div>
                  <div className="model-card__name">{m.display_name}</div>
                  <div className="model-card__vendor">
                    {m.vendor} · via {m.provider_key}
                  </div>
                </div>
                <span className={`tier-badge tier-badge--${m.tier}`}>{tierLabel(m.tier)}</span>
              </div>
              <code className="model-card__id">{m.model_id}</code>
              <div className="model-card__tags">
                {m.tags.map((tag) => (
                  <span key={tag} className="badge">
                    {tag}
                  </span>
                ))}
              </div>
              {m.context && (
                <div className="model-card__meta">
                  <span>Context {m.context}</span>
                </div>
              )}
              <div className="model-card__actions">
                <Link to="/settings" className="btn btn--secondary btn--sm">
                  Add in Settings
                </Link>
              </div>
            </article>
          ))}
        </div>
      </PageSection>

      {!loading && serverModels.length > 0 && (
        <PageSection title="Live server registry" subtitle="Models currently registered on this API instance.">
          <div className="table-wrap table-wrap--flush">
            <p className="table-meta">
              {serverFiltered.length} of {serverModels.length} models
            </p>
            <table className="data-table">
              <thead>
                <tr>
                  <th>Id</th>
                  <th>Display</th>
                  <th>Provider</th>
                  <th>Quality</th>
                  <th>Tags</th>
                </tr>
              </thead>
              <tbody>
                {serverFiltered.map((m) => (
                  <tr key={m.id}>
                    <td className="mono">{m.id}</td>
                    <td>{m.display_name ?? m.model_name ?? '—'}</td>
                    <td>{m.provider ?? '—'}</td>
                    <td>{m.quality_score != null ? m.quality_score.toFixed(2) : '—'}</td>
                    <td>{(m.tags ?? []).join(', ') || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </PageSection>
      )}
    </div>
  )
}
