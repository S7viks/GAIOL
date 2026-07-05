import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader, PageSection } from '../components/layout/PageShell'
import { apiGet, ApiError } from '../lib/api'
import { useToast } from '../components/ui/Toast'
import type { TraceBundle, TraceIdsResponse } from '../types/api'

export function MetricsPage() {
  const toast = useToast()
  const [traceId, setTraceId] = useState('')
  const [summary, setSummary] = useState<Record<string, unknown> | null>(null)
  const [ids, setIds] = useState<string[]>([])
  const [loading, setLoading] = useState(false)

  const loadIds = useCallback(async () => {
    try {
      const data = (await apiGet('/api/orchestration/trace-ids?limit=30')) as TraceIdsResponse
      setIds(data.trace_ids ?? [])
    } catch {
      setIds([])
    }
  }, [])

  useEffect(() => {
    void loadIds()
  }, [loadIds])

  async function loadSummary() {
    const id = traceId.trim()
    if (!id) {
      toast.error('Enter trace id')
      return
    }
    setLoading(true)
    setSummary(null)
    try {
      const bundle = (await apiGet(`/api/orchestration/traces/${encodeURIComponent(id)}`)) as TraceBundle
      setSummary((bundle.metrics_summary as Record<string, unknown>) ?? {})
      toast.success('Loaded metrics_summary')
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  const numericRows = summary
    ? Object.entries(summary).filter(
        ([, v]) => typeof v === 'number' || (typeof v === 'object' && v !== null && !Array.isArray(v)),
      )
    : []

  return (
    <div className="page">
      <PageHeader
        title="Metrics"
        description={
          <>
            Inspect <code>metrics_summary</code> for a trace. Recent ids from{' '}
            <code>/api/orchestration/trace-ids</code>.
          </>
        }
      />

      <div className="page-grid page-grid--sidebar">
        <PageSection title="Trace lookup">
          <div className="form-field">
            <label htmlFor="tid">Trace id</label>
            <input id="tid" value={traceId} onChange={(e) => setTraceId(e.target.value)} placeholder="uuid" />
          </div>
          <div className="btn-row">
            <button type="button" className="btn" onClick={() => void loadSummary()} disabled={loading}>
              {loading ? 'Loading…' : 'Load metrics'}
            </button>
            <button type="button" className="btn btn--secondary" onClick={() => void loadIds()}>
              Refresh ids
            </button>
            {traceId.trim() && (
              <Link to={`/trace/${encodeURIComponent(traceId.trim())}`} className="btn btn--secondary">
                Open trace
              </Link>
            )}
          </div>
        </PageSection>

        <PageSection title="Recent trace ids">
          {ids.length === 0 ? (
            <p className="empty-state">No traces yet.</p>
          ) : (
            <ul className="list-plain">
              {ids.map((id) => (
                <li key={id} className="list-plain__item">
                  <Link to={`/trace/${encodeURIComponent(id)}`} className="mono">
                    {id}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </PageSection>
      </div>

      {summary && (
        <PageSection title="metrics_summary">
          {numericRows.length === 0 ? (
            <pre className="mono-block">{JSON.stringify(summary, null, 2)}</pre>
          ) : (
            <div className="table-wrap table-wrap--flush">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Key</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  {numericRows.map(([k, v]) => (
                    <tr key={k}>
                      <td className="mono">{k}</td>
                      <td className="mono">{typeof v === 'object' ? JSON.stringify(v) : String(v)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </PageSection>
      )}
    </div>
  )
}

