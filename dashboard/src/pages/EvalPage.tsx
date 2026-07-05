import { useState } from 'react'
import { PageHeader, PageSection } from '../components/layout/PageShell'
import { apiPost, ApiError } from '../lib/api'
import { useToast } from '../components/ui/Toast'

const DEFAULT_EXAMPLES = `[
  { "objective": "Greet the user", "expectedContains": ["hello"] }
]`

export function EvalPage() {
  const toast = useToast()
  const [answerText, setAnswerText] = useState('Hello there!')
  const [examplesJson, setExamplesJson] = useState(DEFAULT_EXAMPLES)
  const [resultJson, setResultJson] = useState('')
  const [loading, setLoading] = useState(false)

  async function runEval() {
    let examples: unknown
    try {
      examples = JSON.parse(examplesJson) as unknown
    } catch {
      toast.error('Examples must be valid JSON array')
      return
    }
    if (!Array.isArray(examples)) {
      toast.error('Examples must be a JSON array')
      return
    }
    setLoading(true)
    setResultJson('')
    try {
      const out = await apiPost('/api/orchestration/eval/contains', {
        examples,
        answerText,
      })
      setResultJson(JSON.stringify(out, null, 2))
      toast.success('Eval complete')
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="page">
      <PageHeader
        title="Eval"
        description={
          <>
            Offline QA harness: checks whether answer text contains expected substrings (case-insensitive).
            Does not call a model — paste an answer from Chat or elsewhere, then run eval.
          </>
        }
      />

      <div className="page-grid page-grid--2">
        <PageSection title="Input">
          <div className="form-field">
            <label htmlFor="ex">Examples JSON</label>
            <textarea id="ex" value={examplesJson} onChange={(e) => setExamplesJson(e.target.value)} spellCheck={false} />
          </div>
          <div className="form-field">
            <label htmlFor="ans">Answer text</label>
            <textarea id="ans" value={answerText} onChange={(e) => setAnswerText(e.target.value)} rows={4} />
            <p className="form-hint">
              Each string in <code>expectedContains</code> must appear literally in the answer (ignoring case).
              &quot;Hello there!&quot; matches <code>hello</code> but not <code>hi</code>.
            </p>
          </div>
          <button type="button" className="btn" onClick={() => void runEval()} disabled={loading}>
            {loading ? 'Running…' : 'Run eval'}
          </button>
        </PageSection>

        <PageSection title="Result">
          {resultJson ? (
            <pre className="mono-block">{resultJson}</pre>
          ) : (
            <p className="empty-state">Run eval to see output.</p>
          )}
        </PageSection>
      </div>
    </div>
  )
}

