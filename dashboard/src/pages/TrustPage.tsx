import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader, PageSection } from '../components/layout/PageShell'
import { apiGet, ApiError } from '../lib/api'
import { useToast } from '../components/ui/Toast'
import type { TrustListResponse, TrustRecord } from '../types/api'

function trustMean(d: TrustRecord['distribution'] | undefined): number {
  if (!d) return 0.5
  const s = d.alpha + d.beta
  if (s <= 0) return 0.5
  return d.alpha / s
}

export function TrustPage() {
  const toast = useToast()
  const [domain, setDomain] = useState('')
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<TrustListResponse | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const q = domain.trim() ? `?domain=${encodeURIComponent(domain.trim())}` : ''
      const res = (await apiGet(`/api/orchestration/trust${q}`)) as TrustListResponse
      setData(res)
    } catch (e) {
      setData(null)
      const msg = e instanceof ApiError ? e.message : String(e)
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }, [domain, toast])

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="page">
      <PageHeader
        title="Trust heatmap"
        description={
          <>
            ABTC posteriors from <code>/api/orchestration/trust</code> after smart-query runs.
          </>
        }
      />

      <PageSection title="Filter">
        <div className="page-grid page-grid--sidebar">
          <div className="form-field">
            <label htmlFor="domain">Domain (optional)</label>
            <input
              id="domain"
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="e.g. general"
            />
          </div>
          <div className="btn-row" style={{ alignSelf: 'end' }}>
            <button type="button" className="btn" onClick={() => void load()} disabled={loading}>
              {loading ? 'Loading…' : 'Refresh'}
            </button>
          </div>
        </div>
      </PageSection>

      {data && (
        <PageSection title="Records">
          <div className="badge-row">
            <span className="badge">{data.count} records</span>
            {data.domain ? <span className="badge">domain={data.domain}</span> : null}
          </div>
          <div className="table-wrap table-wrap--flush">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Model</th>
                  <th>Domain</th>
                  <th>Mean trust</th>
                  <th>α / β</th>
                  <th>Updated</th>
                </tr>
              </thead>
              <tbody>
                {data.records?.length ? (
                  data.records.map((r) => (
                    <tr key={`${r.model_id}-${r.domain}`}>
                      <td>
                        <Link to={`/models?q=${encodeURIComponent(r.model_id)}`}>{r.model_id}</Link>
                      </td>
                      <td>{r.domain}</td>
                      <td>{trustMean(r.distribution).toFixed(3)}</td>
                      <td className="mono">
                        {(r.distribution?.alpha ?? 0).toFixed(2)} / {(r.distribution?.beta ?? 0).toFixed(2)}
                      </td>
                      <td className="mono">{r.updated_at}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={5} className="empty-state">
                      No trust rows yet. Run a smart query to populate ABTC state.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </PageSection>
      )}
    </div>
  )
}

