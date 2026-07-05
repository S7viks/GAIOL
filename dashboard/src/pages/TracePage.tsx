import { useEffect, useState } from 'react'

import { Link, useNavigate, useParams } from 'react-router-dom'

import { PageAlert, PageHeader, PageSection } from '../components/layout/PageShell'

import { apiGet, ApiError } from '../lib/api'

import { useToast } from '../components/ui/Toast'

import type { TraceBundle } from '../types/api'



export function TracePage() {

  const { id } = useParams<{ id: string }>()

  const navigate = useNavigate()

  const toast = useToast()

  const [manualId, setManualId] = useState('')

  const [loading, setLoading] = useState(false)

  const [bundle, setBundle] = useState<TraceBundle | null>(null)

  const [error, setError] = useState<string | null>(null)

  const [expanded, setExpanded] = useState<'timeline' | 'trace' | 'metrics' | null>('metrics')



  useEffect(() => {

    if (!id) {

      setBundle(null)

      setError(null)

      return

    }

    let cancelled = false

    ;(async () => {

      setLoading(true)

      setError(null)

      setBundle(null)

      try {

        const data = (await apiGet(`/api/orchestration/traces/${encodeURIComponent(id)}`)) as TraceBundle

        if (!cancelled) setBundle(data)

      } catch (e) {

        if (cancelled) return

        const msg = e instanceof ApiError ? `${e.message} (${e.status})` : String(e)

        setError(msg)

        if (e instanceof ApiError && (e.code === 'orchestrator_disabled' || e.code === 'ts_orchestrator_disabled')) {

          toast.error('Orchestrator not configured on Go server')

        }

      } finally {

        if (!cancelled) setLoading(false)

      }

    })()

    return () => {

      cancelled = true

    }

  }, [id, toast])



  function goTrace() {

    const t = manualId.trim()

    if (!t) {

      toast.error('Enter a trace id')

      return

    }

    navigate(`/trace/${encodeURIComponent(t)}`)

  }



  if (!id) {

    return (

      <div className="page">

        <PageHeader

          title="Trace viewer"

          description="Load orchestration traces by id — from Chat links or your logs."

        />

        <PageSection title="Lookup">

          <div className="form-field">

            <label htmlFor="tid">Trace id</label>

            <input

              id="tid"

              value={manualId}

              onChange={(e) => setManualId(e.target.value)}

              placeholder="paste uuid"

            />

          </div>

          <div className="btn-row">

            <button type="button" className="btn" onClick={goTrace}>

              Open trace

            </button>

            <Link to="/chat" className="btn btn--secondary">

              Chat

            </Link>

          </div>

        </PageSection>

      </div>

    )

  }



  return (

    <div className="page">

      <PageHeader

        title="Trace viewer"

        description={<span className="trace-id">{id}</span>}

        actions={

          <>

            <Link to="/trace" className="btn btn--secondary btn--sm">

              New lookup

            </Link>

            <Link to="/chat" className="btn btn--secondary btn--sm">

              Chat

            </Link>

          </>

        }

      />



      {loading && <div className="skeleton skeleton--block" />}



      {error && <PageAlert variant="err">{error}</PageAlert>}



      {bundle && (

        <div className="page-stack">

          {(['metrics', 'timeline', 'trace'] as const).map((key) => {

            const labels = {

              metrics: 'metrics_summary',

              timeline: 'timeline_rebuilt',

              trace: 'trace (raw)',

            }

            const payload =

              key === 'metrics'

                ? bundle.metrics_summary ?? {}

                : key === 'timeline'

                  ? bundle.timeline_rebuilt ?? []

                  : bundle.trace ?? {}

            return (

              <section key={key} className="panel accordion-panel">

                <button

                  type="button"

                  className="accordion-panel__toggle"

                  onClick={() => setExpanded(expanded === key ? null : key)}

                >

                  <span>{labels[key]}</span>

                  <span aria-hidden>{expanded === key ? '▼' : '▶'}</span>

                </button>

                {expanded === key && (

                  <div className="accordion-panel__body">

                    <pre className="mono-block">{JSON.stringify(payload, null, 2)}</pre>

                  </div>

                )}

              </section>

            )

          })}

        </div>

      )}

    </div>

  )

}


