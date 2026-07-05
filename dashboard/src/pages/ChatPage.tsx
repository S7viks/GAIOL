import { useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGet, apiPost, ApiError } from '../lib/api'
import { type SetupStatus } from '../lib/setupStatus'
import { useToast } from '../components/ui/Toast'
import type { SmartQueryResponse } from '../types/api'
import { ChatMessageContent } from '../components/chat/ChatMessageContent'
import { ChatTypingIndicator } from '../components/chat/ChatTypingIndicator'
import { ChatWelcome } from '../components/chat/ChatWelcome'
import { useAppStore } from '../store'

function chatErrorMessage(err: unknown): string {
  if (!(err instanceof ApiError)) return String(err)
  switch (err.code) {
    case 'no_provider_keys':
      return 'No provider connected. Add a provider API key in Settings or complete onboarding first.'
    case 'orchestrator_unavailable':
    case 'orchestrator_unreachable':
      return 'Orchestrator is not available. Ensure the Go API is running (orchestration runs in-process).'
    case 'orchestrator_timeout':
      return 'Orchestrator timed out. Try a shorter prompt or check provider connectivity.'
    case 'orchestrator_upstream_error':
      return `Orchestrator error: ${err.message}`
    default:
      return err.message
  }
}

const STRATEGIES = [
  'balanced',
  'lowest_cost',
  'highest_quality',
  'free_only',
  'beam',
] as const

const TASKS = ['qa', 'code', 'summarization', 'reasoning', 'creative', 'tool_use', 'unknown'] as const

type ChatTurn = {
  id: string
  role: 'user' | 'assistant'
  content: string
  traceId?: string
}

function newTurnId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

export function ChatPage() {
  const toast = useToast()
  const setSessionId = useAppStore((s) => s.setSessionId)
  const addToHistory = useAppStore((s) => s.addToHistory)
  const threadEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const [prompt, setPrompt] = useState('')
  const [strategy, setStrategy] = useState<string>('balanced')
  const [task, setTask] = useState<string>('qa')
  const [maxTokens, setMaxTokens] = useState(2048)
  const [temperature, setTemperature] = useState(0.7)
  const [loading, setLoading] = useState(false)
  const [turns, setTurns] = useState<ChatTurn[]>([])
  const [lastMeta, setLastMeta] = useState<string>('')
  const [setup, setSetup] = useState<SetupStatus | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const p = (await apiGet('/api/settings/preferences')) as { strategy?: string }
        if (cancelled || !p?.strategy?.trim()) return
        setStrategy(p.strategy.trim())
      } catch {
        /* keep defaults */
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const s = (await apiGet('/api/setup/status')) as SetupStatus
        if (!cancelled) setSetup(s)
      } catch {
        if (!cancelled) setSetup(null)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    threadEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [turns, loading])

  const resizeComposer = useCallback(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 160)}px`
  }, [])

  useEffect(() => {
    resizeComposer()
  }, [prompt, resizeComposer])

  async function sendMessage(text: string) {
    const trimmed = text.trim()
    if (!trimmed || loading) return

    const userTurn: ChatTurn = { id: newTurnId(), role: 'user', content: trimmed }
    setTurns((prev) => [...prev, userTurn])
    setPrompt('')
    setLoading(true)
    setLastMeta('')

    try {
      const data = (await apiPost('/api/query/smart', {
        prompt: trimmed,
        strategy,
        task,
        max_tokens: maxTokens,
        temperature,
      })) as SmartQueryResponse

      const answer = data.response ?? data.result?.data ?? ''
      const tid =
        data.metadata?.trace_id ??
        data.orchestration?.trace_id ??
        data.metadata?.session_id ??
        undefined

      const displayAnswer =
        answer.startsWith('provider error') || answer.includes(': provider error')
          ? `${answer}\n\nTip: open Settings → Provider keys, delete and re-add your key. Keys are validated on save.`
          : answer || '(empty response)'

      setTurns((prev) => [
        ...prev,
        { id: newTurnId(), role: 'assistant', content: displayAnswer, traceId: tid },
      ])

      if (tid) setSessionId(tid)

      addToHistory({
        id: newTurnId(),
        query: trimmed,
        timestamp: Date.now(),
      })

      setLastMeta(
        JSON.stringify(
          {
            metadata: data.metadata,
            orchestration: data.orchestration,
            cost: data.cost,
            latency_ms: data.latency_ms,
            strategy: data.strategy,
          },
          null,
          2,
        ),
      )
    } catch (err) {
      setTurns((prev) => [
        ...prev,
        { id: newTurnId(), role: 'assistant', content: `Error: ${chatErrorMessage(err)}` },
      ])
      toast.error(chatErrorMessage(err))
    } finally {
      setLoading(false)
      textareaRef.current?.focus()
    }
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    void sendMessage(prompt)
  }

  function onComposerKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      void sendMessage(prompt)
    }
  }

  function clearThread() {
    setTurns([])
    setLastMeta('')
    setSettingsOpen(false)
    textareaRef.current?.focus()
  }

  const showSetupWarnings =
    setup && (!setup.orchestrator_reachable || setup.tenant_ready === false)

  return (
    <div className="chatbot">
      <header className="chatbot__header">
        <div className="chatbot__header-main">
          <h1 className="chatbot__title">Chat</h1>
          <span className="chatbot__subtitle">{strategy} · {task}</span>
        </div>
        <div className="chatbot__header-actions">
          {turns.length > 0 && (
            <button type="button" className="chatbot__icon-btn" onClick={clearThread} title="New chat">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
                <path d="M12 5v14M5 12h14" strokeWidth="2" strokeLinecap="round" />
              </svg>
              <span className="chatbot__icon-btn-label">New</span>
            </button>
          )}
          <button
            type="button"
            className={`chatbot__icon-btn ${settingsOpen ? 'chatbot__icon-btn--active' : ''}`}
            onClick={() => setSettingsOpen((v) => !v)}
            title="Chat settings"
            aria-expanded={settingsOpen}
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
              <path
                d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z"
                strokeWidth="2"
              />
              <path
                d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9c.26.604.852.997 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"
                strokeWidth="2"
              />
            </svg>
          </button>
        </div>
      </header>

      {showSetupWarnings && (
        <div className="chatbot__banner chatbot__banner--warn">
          {!setup?.orchestrator_reachable && (
            <span>Orchestrator unreachable — start the Go API. </span>
          )}
          {setup?.tenant_ready === false && (
            <span>
              No provider connected.{' '}
              <Link to="/onboarding">Setup</Link> or <Link to="/settings">Settings</Link>.
            </span>
          )}
        </div>
      )}

      {settingsOpen && (
        <div className="chatbot__settings">
          <div className="chatbot__settings-grid">
            <div className="form-field">
              <label htmlFor="strategy">Strategy</label>
              <select
                id="strategy"
                value={strategy}
                onChange={(e) => setStrategy(e.target.value)}
                disabled={loading}
              >
                {!STRATEGIES.includes(strategy as (typeof STRATEGIES)[number]) && strategy ? (
                  <option value={strategy}>{strategy} (from settings)</option>
                ) : null}
                {STRATEGIES.map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-field">
              <label htmlFor="task">Task</label>
              <select id="task" value={task} onChange={(e) => setTask(e.target.value)} disabled={loading}>
                {TASKS.map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-field">
              <label htmlFor="max_tokens">Max tokens</label>
              <input
                id="max_tokens"
                type="number"
                min={16}
                max={8192}
                value={maxTokens}
                onChange={(e) => setMaxTokens(Number(e.target.value) || 2048)}
                disabled={loading}
              />
            </div>
            <div className="form-field">
              <label htmlFor="temp">Temperature</label>
              <input
                id="temp"
                type="number"
                step="0.1"
                min={0}
                max={2}
                value={temperature}
                onChange={(e) => setTemperature(Number(e.target.value))}
                disabled={loading}
              />
            </div>
          </div>
          <p className="chatbot__settings-note">
            Defaults from <Link to="/settings">Settings</Link>. Changes apply to the next message only.
          </p>
        </div>
      )}

      <div className="chatbot__scroll">
        <div className="chatbot__thread">
          {turns.length === 0 && !loading ? (
            <ChatWelcome onPick={(s) => setPrompt(s)} />
          ) : (
            turns.map((turn) => (
              <article
                key={turn.id}
                className={`chatbot-msg chatbot-msg--${turn.role}`}
              >
                <div
                  className={`chatbot-msg__avatar chatbot-msg__avatar--${turn.role === 'user' ? 'user' : 'bot'}`}
                  aria-hidden="true"
                >
                  {turn.role === 'user' ? 'U' : 'G'}
                </div>
                <div className="chatbot-msg__content">
                  <div className="chatbot-msg__meta">
                    <span>{turn.role === 'user' ? 'You' : 'GAIOL'}</span>
                    {turn.traceId && (
                      <Link to={`/trace/${encodeURIComponent(turn.traceId)}`} className="chatbot-msg__trace">
                        View trace
                      </Link>
                    )}
                  </div>
                  <div className="chatbot-msg__bubble">
                    <ChatMessageContent content={turn.content} role={turn.role} />
                  </div>
                </div>
              </article>
            ))
          )}
          {loading && <ChatTypingIndicator />}
          <div ref={threadEndRef} className="chatbot__anchor" />
        </div>
      </div>

      <footer className="chatbot__footer">
        {lastMeta && (
          <details className="chatbot__meta-details">
            <summary>Last response details</summary>
            <pre className="mono-block">{lastMeta}</pre>
          </details>
        )}
        <form className="chatbot__composer" onSubmit={onSubmit}>
          <div className="chatbot__composer-inner">
            <textarea
              ref={textareaRef}
              id="prompt"
              className="chatbot__input"
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              onKeyDown={onComposerKeyDown}
              disabled={loading}
              placeholder="Message GAIOL…"
              rows={1}
              aria-label="Message"
            />
            <button
              type="submit"
              className="chatbot__send"
              disabled={loading || !prompt.trim()}
              title="Send message"
              aria-label="Send message"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true">
                <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" strokeWidth="2" strokeLinejoin="round" />
              </svg>
            </button>
          </div>
          <p className="chatbot__hint">Enter to send · Shift+Enter for new line</p>
        </form>
      </footer>
    </div>
  )
}
