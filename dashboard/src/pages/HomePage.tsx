import { useEffect, useState } from 'react'

import { Link } from 'react-router-dom'

import { apiGet, ApiError, fetchHealthBody } from '../lib/api'

import { fetchSetupStatus, type SetupStatus } from '../lib/setupStatus'

import type {

  BillingSummaryResponse,

  PreferencesResponse,

  UsageResponse,

} from '../types/api'



type QuickItem = {

  to: string

  title: string

  body: string

  group: 'work' | 'inspect' | 'configure'

}



const QUICK: QuickItem[] = [

  {

    to: '/chat',

    title: 'Chat',

    body: 'Run a prompt through orchestration. Each response includes a trace id.',

    group: 'work',

  },

  {

    to: '/trace',

    title: 'Trace Viewer',

    body: 'Metrics, timeline, and raw payload for any run.',

    group: 'inspect',

  },

  {

    to: '/trust',

    title: 'Trust heatmap',

    body: 'ABTC posteriors after smart-query runs.',

    group: 'inspect',

  },

  {

    to: '/history',

    title: 'History',

    body: 'Usage, billing, and recent activity.',

    group: 'inspect',

  },

  {

    to: '/models',

    title: 'Models',

    body: 'Registered model catalog from the server.',

    group: 'configure',

  },

  {

    to: '/settings',

    title: 'Settings',

    body: 'Strategy, budget, provider keys, and GAIOL API keys.',

    group: 'configure',

  },

  {

    to: '/onboarding',

    title: 'Onboarding',

    body: 'Connect provider keys and start chatting.',

    group: 'configure',

  },

]



const QUICK_GROUPS: { id: QuickItem['group']; label: string }[] = [

  { id: 'work', label: 'Work' },

  { id: 'inspect', label: 'Inspect' },

  { id: 'configure', label: 'Configure' },

]



function formatUsd(n: number | undefined): string {

  if (n == null || Number.isNaN(n)) return '—'

  return `$${n.toFixed(4)}`

}



export function HomePage() {

  const [prefs, setPrefs] = useState<PreferencesResponse | null>(null)

  const [prefsError, setPrefsError] = useState<string | null>(null)

  const [setup, setSetup] = useState<SetupStatus | null>(null)

  const [authDisabled, setAuthDisabled] = useState(false)

  const [billing, setBilling] = useState<BillingSummaryResponse | null>(null)

  const [usage, setUsage] = useState<UsageResponse | null>(null)



  useEffect(() => {

    let cancelled = false

    ;(async () => {

      try {

        const health = await fetchHealthBody()

        if (cancelled) return

        setAuthDisabled(health.authDisabled)



        const status = await fetchSetupStatus(apiGet)

        if (!cancelled) setSetup(status)



        try {

          const p = (await apiGet('/api/settings/preferences')) as PreferencesResponse

          if (!cancelled) {

            setPrefs(p)

            setPrefsError(null)

          }

        } catch (e) {

          if (!cancelled) {

            setPrefs(null)

            setPrefsError(e instanceof ApiError ? e.message : String(e))

          }

        }



        if (!health.authDisabled) {

          try {

            const [bill, use] = await Promise.all([

              apiGet('/api/billing/summary') as Promise<BillingSummaryResponse>,

              apiGet('/api/usage') as Promise<UsageResponse>,

            ])

            if (!cancelled) {

              setBilling(bill)

              setUsage(use)

            }

          } catch {

            if (!cancelled) {

              setBilling(null)

              setUsage(null)

            }

          }

        }

      } catch {

        /* non-fatal */

      }

    })()

    return () => {

      cancelled = true

    }

  }, [])



  const mtdCost = billing?.total_cost ?? 0

  const budgetLimit = prefs?.budget_limit

  const budgetPct =

    budgetLimit != null && budgetLimit > 0 ? Math.min(100, (mtdCost / budgetLimit) * 100) : null

  const budgetOver = (budgetPct ?? 0) >= 90



  const showSetupAlert = setup?.setup_complete === false

  const showOrchestratorAlert = setup?.orchestrator_reachable === false



  return (

    <div className="page home-page">

      <header className="home-hero panel">

        <div className="home-hero__copy">

          <h1>Home</h1>

          <p className="page-shell__desc">

            One orchestration engine across your connected providers — with a trace for every run.

          </p>

        </div>

        <div className="home-hero__actions">

          <Link to="/chat" className="btn home-hero__cta">

            Open Chat

          </Link>

          <Link to="/trace" className="btn btn--secondary">

            Trace Viewer

          </Link>

        </div>

      </header>



      {(showSetupAlert || showOrchestratorAlert || authDisabled) && (

        <div className="home-status-stack">

          {showSetupAlert && (

            <div className="alert alert--warn home-status">

              <div className="home-status__body">

                <strong>Setup incomplete</strong>

                <span>Connect at least one provider and create a GAIOL API key.</span>

              </div>

              <Link to="/onboarding" className="btn btn--secondary home-status__action">

                Complete setup

              </Link>

            </div>

          )}

          {showOrchestratorAlert && (

            <div className="alert alert--warn home-status">

              <div className="home-status__body">

                <strong>Orchestrator unreachable</strong>

                <span>

                  Run <code>.\scripts\dev\start-relay.ps1</code> or <code>go run ./cmd/web-server</code>.

                </span>

              </div>

            </div>

          )}

          {authDisabled && (

            <div className="alert alert--warn home-status">

              <div className="home-status__body">

                <strong>Local no-auth mode</strong>

                <span>Usage and billing stats require Supabase auth and database.</span>

              </div>

            </div>

          )}

        </div>

      )}



      {!authDisabled && (

        <section className="home-stat-grid" aria-label="Usage overview">

          <div className="panel home-stat-card">

            <span className="home-stat-card__label">MTD spend</span>

            <span className="home-stat-card__value">{formatUsd(mtdCost)}</span>

            {billing?.period && <span className="home-stat-card__hint">{billing.period}</span>}

          </div>

          <div className="panel home-stat-card">

            <span className="home-stat-card__label">Requests (30d)</span>

            <span className="home-stat-card__value">{usage?.summary?.requests ?? '—'}</span>

            <Link to="/history" className="home-stat-card__link">

              View history

            </Link>

          </div>

          <div className="panel home-stat-card">

            <span className="home-stat-card__label">Providers</span>

            <span className="home-stat-card__value">{setup?.providers_connected ?? '—'}</span>

          </div>

          <div className="panel home-stat-card">

            <span className="home-stat-card__label">Monthly budget</span>

            {budgetLimit != null && budgetLimit > 0 ? (

              <>

                <span className="home-stat-card__value">

                  {formatUsd(mtdCost)} / {formatUsd(budgetLimit)}

                </span>

                <div className="home-budget-bar" role="presentation">

                  <div

                    className={`home-budget-bar__fill${budgetOver ? ' home-budget-bar__fill--warn' : ''}`}

                    style={{ width: `${budgetPct ?? 0}%` }}

                  />

                </div>

              </>

            ) : (

              <>

                <span className="home-stat-card__value home-stat-card__value--muted">Not set</span>

                <Link to="/settings" className="home-stat-card__link">

                  Set in Settings

                </Link>

              </>

            )}

          </div>

        </section>

      )}



      <div className="home-main-grid">

        <section className="panel home-defaults">

          <h2 className="home-section-title">Saved defaults</h2>

          <p className="page-shell__desc home-defaults__intro">

            From <Link to="/settings">Settings</Link>. Chat pre-fills strategy when you open it; each request still

            sends whatever you choose in the form.

          </p>

          {prefsError && (

            <p className="home-defaults__error">Could not load preferences ({prefsError}).</p>

          )}

          {!prefsError && (

            <dl className="home-kv">

              <div className="home-kv__row">

                <dt>Strategy</dt>

                <dd>

                  <code className="inline-code">{prefs?.strategy ?? 'balanced'}</code>

                </dd>

              </div>

              <div className="home-kv__row">

                <dt>Routing</dt>

                <dd>

                  <span className="home-kv__empty">All connected providers</span>

                  <p className="home-kv__note">

                    Chat uses every provider key you save and models listed under Settings → Tenant models.
                    Strategy controls how the orchestrator picks among them.

                  </p>

                </dd>

              </div>

            </dl>

          )}

        </section>



        <section className="home-quick">

          <h2 className="home-section-title">Go to</h2>

          {QUICK_GROUPS.map((group) => {

            const items = QUICK.filter((q) => q.group === group.id)

            if (items.length === 0) return null

            return (

              <div key={group.id} className="home-quick-group">

                <h3 className="home-quick-group__label">{group.label}</h3>

                <div className="home-quick-grid">

                  {items.map((q) => (

                    <Link key={q.to} to={q.to} className="panel home-quick-card">

                      <span className="home-quick-card__title">{q.title}</span>

                      <span className="home-quick-card__body">{q.body}</span>

                    </Link>

                  ))}

                </div>

              </div>

            )

          })}

        </section>

      </div>

    </div>

  )

}


