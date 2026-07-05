import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { PageHeader, PageSection, PageStack } from '../components/layout/PageShell'
import { apiDownload, apiGet, ApiError } from '../lib/api'
import { useToast } from '../components/ui/Toast'
import { useAppStore } from '../store'
import type {
  ActivityResponse,
  BillingHistoryResponse,
  UsageResponse,
} from '../types/api'

function formatUsd(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `$${n.toFixed(4)}`
}

export function HistoryPage() {
  const toast = useToast()
  const localHistory = useAppStore((s) => s.queryHistory)
  const [activity, setActivity] = useState<ActivityResponse['activity']>([])
  const [usage, setUsage] = useState<UsageResponse | null>(null)
  const [billing, setBilling] = useState<BillingHistoryResponse['history']>([])
  const [loading, setLoading] = useState(true)
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setLoading(true)
      try {
        const [act, use, bill] = await Promise.all([
          apiGet('/api/activity?limit=50').catch(() => ({ activity: [] })),
          apiGet('/api/usage').catch(() => null),
          apiGet('/api/billing/history').catch(() => ({ history: [] })),
        ])
        if (!cancelled) {
          setActivity((act as ActivityResponse).activity ?? [])
          setUsage(use as UsageResponse | null)
          setBilling((bill as BillingHistoryResponse).history ?? [])
        }
      } catch (e) {
        if (!cancelled) toast.error(e instanceof ApiError ? e.message : String(e))
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [toast])

  async function exportCsv() {
    setExporting(true)
    try {
      await apiDownload('/api/usage/export', 'usage.csv')
      toast.success('Export started')
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="page">
      <PageHeader title="History" description="Session prompts, tenant usage, billing, and activity." />

      <PageStack>
        <PageSection
          title="This session"
          subtitle={
            <>
              Browser-only for this tab · <Link to="/chat">Open Chat</Link>
            </>
          }
        >
          {localHistory.length === 0 ? (
            <p className="empty-state">No prompts yet.</p>
          ) : (
            <ul className="list-plain">
              {localHistory.map((h) => (
                <li key={h.id} className="list-plain__item">
                  <div className="list-plain__meta">{new Date(h.timestamp).toLocaleString()}</div>
                  <div>{h.query}</div>
                </li>
              ))}
            </ul>
          )}
        </PageSection>

        {loading && <div className="skeleton skeleton--line" />}

        {!loading && (
          <>
            <PageSection
              title="Usage (30 days)"
              headerActions={
                <button type="button" className="btn btn--secondary btn--sm" disabled={exporting} onClick={() => void exportCsv()}>
                  {exporting ? 'Exporting…' : 'Export CSV'}
                </button>
              }
            >
              {usage?.summary ? (
                <p className="table-meta">
                  {usage.summary.requests ?? 0} requests · {usage.summary.tokens ?? 0} tokens ·{' '}
                  {formatUsd(usage.summary.cost)}
                </p>
              ) : (
                <p className="empty-state">No usage data (requires auth + database).</p>
              )}

              {(usage?.by_day?.length ?? 0) > 0 && (
                <>
                  <h3 className="page-section__title" style={{ marginTop: 10 }}>
                    By day
                  </h3>
                  <div className="table-wrap">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>Date</th>
                          <th>Requests</th>
                          <th>Tokens</th>
                          <th>Cost</th>
                        </tr>
                      </thead>
                      <tbody>
                        {(usage?.by_day ?? []).map((row) => (
                          <tr key={row.date}>
                            <td>{row.date}</td>
                            <td>{row.requests ?? 0}</td>
                            <td>{row.tokens ?? 0}</td>
                            <td>{formatUsd(row.cost)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </>
              )}

              {(usage?.by_provider?.length ?? 0) > 0 && (
                <>
                  <h3 className="page-section__title" style={{ marginTop: 10 }}>
                    By provider
                  </h3>
                  <div className="table-wrap">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>Provider</th>
                          <th>Requests</th>
                          <th>Cost</th>
                        </tr>
                      </thead>
                      <tbody>
                        {(usage?.by_provider ?? []).map((row) => (
                          <tr key={row.provider}>
                            <td>{row.provider}</td>
                            <td>{row.requests ?? 0}</td>
                            <td>{formatUsd(row.cost)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </>
              )}
            </PageSection>

            <div className="page-grid page-grid--2">
              <PageSection title="Billing history">
                {(billing?.length ?? 0) === 0 ? (
                  <p className="empty-state">No billing history yet.</p>
                ) : (
                  <div className="table-wrap table-wrap--flush">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>Month</th>
                          <th>Total cost</th>
                        </tr>
                      </thead>
                      <tbody>
                        {(billing ?? []).map((row) => (
                          <tr key={row.month}>
                            <td>{row.month}</td>
                            <td>{formatUsd(row.total_cost)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </PageSection>

              <PageSection title="Activity">
                {(activity?.length ?? 0) === 0 ? (
                  <p className="empty-state">No activity logged.</p>
                ) : (
                  <div className="table-wrap table-wrap--flush">
                    <table className="data-table">
                      <thead>
                        <tr>
                          <th>When</th>
                          <th>Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {(activity ?? []).map((row) => (
                          <tr key={row.id ?? `${row.action}-${row.created_at}`}>
                            <td className="mono">{row.created_at?.slice(0, 19) ?? '—'}</td>
                            <td>{row.action ?? '—'}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </PageSection>
            </div>
          </>
        )}
      </PageStack>
    </div>
  )
}
