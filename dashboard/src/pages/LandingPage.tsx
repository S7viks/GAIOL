import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

type LogLine = {
  kind: 'comment' | 'cmd' | 'out' | 'ok' | 'dim'
  text: string
}

const BOOT_LOG: LogLine[] = [
  { kind: 'comment', text: 'gaiol dispatch — tenant session' },
  { kind: 'out', text: 'scanning provider pool…' },
  { kind: 'ok', text: 'openrouter    linked' },
  { kind: 'ok', text: 'gemini        linked' },
  { kind: 'ok', text: 'huggingface   linked' },
  { kind: 'out', text: 'orchestrator  ready  (beam=4, abtc=on)' },
  { kind: 'dim', text: '—' },
  { kind: 'cmd', text: 'awaiting requests on /v1/chat' },
]

const COMPARE = [
  { before: 'Four API keys in env vars', after: 'One gaiol_sk in env' },
  { before: 'Pick a model and hope', after: 'Beam search across your pool' },
  { before: 'Bill arrives, no breakdown', after: 'Per-request cost + trace id' },
  { before: 'Swap provider = rewrite client', after: 'OpenAI-compatible endpoint' },
] as const

const CAPABILITIES = [
  {
    tag: 'routing',
    title: 'Every prompt gets a committee',
    body: 'Decompose, fan out, score, merge. Not a static model map — a live decision per request.',
  },
  {
    tag: 'abtc',
    title: 'Trust that learns',
    body: 'Posteriors update after every run. Models that fail on your workload lose weight automatically.',
  },
  {
    tag: 'credentials',
    title: 'Your keys stay yours',
    body: 'Provider secrets encrypted at rest. GAIOL routes and logs — it does not train on your data.',
  },
] as const

const DISPATCH_LINES = [
  { ts: '14:02:11', req: 'req_8f2a', msg: 'beam=4  decompose ok', tone: 'dim' as const },
  { ts: '14:02:11', req: 'req_8f2a', msg: 'openrouter/gpt-4o-mini', tone: 'score' as const, val: '0.91' },
  { ts: '14:02:11', req: 'req_8f2a', msg: 'gemini/flash', tone: 'score' as const, val: '0.87' },
  { ts: '14:02:11', req: 'req_8f2a', msg: 'hf/llama-70b', tone: 'score' as const, val: '0.72' },
  { ts: '14:02:12', req: 'req_8f2a', msg: 'winner openrouter/gpt-4o-mini', tone: 'win' as const, val: '842ms  $0.0021' },
  { ts: '14:02:12', req: 'req_8f2a', msg: 'trace a7f3c2e1', tone: 'trace' as const },
] as const

function DispatchFeed() {
  return (
    <div className="lp-dispatch" aria-label="Recent dispatch log">
      <div className="lp-dispatch__head">
        <span>tail -f /var/log/gaiol/dispatch.log</span>
        <span className="lp-dispatch__live">live</span>
      </div>
      <div className="lp-dispatch__body">
        {DISPATCH_LINES.map((line) => (
          <div key={`${line.ts}-${line.msg}`} className={`lp-dispatch__row lp-dispatch__row--${line.tone}`}>
            <span className="lp-dispatch__ts">{line.ts}</span>
            <span className="lp-dispatch__req">{line.req}</span>
            <span className="lp-dispatch__msg">{line.msg}</span>
            {'val' in line && line.val ? <span className="lp-dispatch__val">{line.val}</span> : null}
          </div>
        ))}
      </div>
    </div>
  )
}

function SessionLog() {
  const [visible, setVisible] = useState(0)

  useEffect(() => {
    if (visible >= BOOT_LOG.length) return
    const line = BOOT_LOG[visible]
    const ms = line?.kind === 'cmd' ? 420 : 300
    const t = window.setTimeout(() => setVisible((v) => v + 1), ms)
    return () => window.clearTimeout(t)
  }, [visible])

  return (
    <div className="lp-log" aria-label="Session boot log">
      {BOOT_LOG.slice(0, visible).map((line, i) => (
        <div key={i} className={`lp-log__line lp-log__line--${line.kind}`}>
          {line.kind === 'comment' && <span className="lp-log__sig"># </span>}
          {line.kind === 'cmd' && <span className="lp-log__sig">$ </span>}
          {line.kind === 'ok' && <span className="lp-log__sig">  ✓ </span>}
          {line.text}
        </div>
      ))}
      {visible < BOOT_LOG.length && (
        <div className="lp-log__line lp-log__line--cursor">
          <span className="blink">▊</span>
        </div>
      )}
    </div>
  )
}

export function LandingPage() {
  return (
    <div className="lp">
      <header className="lp-hero">
        <div className="lp-hero__grid">
          <div className="lp-hero__left">
            <p className="lp-kicker">model multiplexor</p>
            <h1 className="lp-headline">
              Your providers.
              <br />
              One wire.
            </h1>
            <p className="lp-lede">
              GAIOL sits between your app and the models you already pay for. One key in, orchestrated
              routing out — with traces you can actually read.
            </p>
            <div className="lp-actions">
              <Link to="/signup" className="lp-cmd lp-cmd--go">
                gaiol init
              </Link>
              <Link to="/login" className="lp-cmd lp-cmd--ghost">
                login
              </Link>
            </div>
          </div>

          <div className="lp-hero__right">
            <div className="lp-panel">
              <div className="lp-panel__bar">
                <span className="term-dot" />
                <span className="term-dot" />
                <span className="term-dot" />
                <span className="lp-panel__title">~/dispatch</span>
              </div>
              <SessionLog />
            </div>
          </div>
        </div>
      </header>

      <section id="dispatch" className="lp-block">
        <p className="lp-block__sig"># dispatch log</p>
        <div className="lp-bento">
          <div className="lp-bento__feed">
            <DispatchFeed />
          </div>
          <div className="lp-bento__trace">
            <p className="lp-bento__tag">last run</p>
            <pre className="lp-trace">{`{
  "trace_id": "a7f3c2e1",
  "winner": "openrouter/gpt-4o-mini",
  "candidates": 4,
  "latency_ms": 842,
  "cost_usd": 0.0021,
  "abtc_trust": 0.94
}`}</pre>
          </div>
          {CAPABILITIES.map((item) => (
            <article key={item.tag} className="lp-bento__card">
              <p className="lp-bento__tag">{item.tag}</p>
              <h3 className="lp-bento__title">{item.title}</h3>
              <p className="lp-bento__body">{item.body}</p>
            </article>
          ))}
        </div>
      </section>

      <section id="compare" className="lp-block">
        <p className="lp-block__sig"># diff</p>
        <div className="lp-diff">
          <div className="lp-diff__head">
            <span>without gaiol</span>
            <span aria-hidden="true" className="lp-diff__spacer" />
            <span>with gaiol</span>
          </div>
          {COMPARE.map((row) => (
            <div key={row.before} className="lp-diff__row">
              <span className="lp-diff__before">{row.before}</span>
              <span className="lp-diff__arrow" aria-hidden="true">
                →
              </span>
              <span className="lp-diff__after">{row.after}</span>
            </div>
          ))}
        </div>
      </section>

      <section id="setup" className="lp-block">
        <p className="lp-block__sig"># setup</p>
        <div className="lp-setup">
          <ol className="lp-setup__steps">
            <li>
              Paste provider keys in{' '}
              <Link to="/settings" className="lp-setup__link">
                settings
              </Link>{' '}
              (after signup).
            </li>
            <li>
              Create your <code className="lp-setup__kw">gaiol_sk_*</code> API key.
            </li>
            <li>
              Point your client at <code className="lp-setup__kw">POST /v1/chat</code>.
            </li>
          </ol>
          <pre className="lp-setup__pre">
            <code>{`$ curl -X POST "$GAIOL_API/v1/chat" \\
  -H "Authorization: Bearer $GAIOL_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"messages":[{"role":"user","content":"hi"}]}'`}</code>
          </pre>
          <Link to="/signup" className="lp-cmd lp-cmd--go lp-cmd--block">
            run step 1
          </Link>
        </div>
      </section>

      <footer className="lp-foot">
        <span>
          gaiol<span className="public-nav__brand-accent">_</span>
        </span>
        <nav className="lp-foot__nav" aria-label="Footer">
          <Link to="/login">login</Link>
          <Link to="/home">dashboard</Link>
          <Link to="/terms">terms</Link>
        </nav>
        <span className="lp-foot__hint">exit 0</span>
      </footer>
    </div>
  )
}
