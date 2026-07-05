type ChatWelcomeProps = {
  onPick: (text: string) => void
}

const SUGGESTIONS = [
  'What is Einstein\'s equation?',
  'Explain how GAIOL routes prompts across providers',
  'Write a short Python function to reverse a string',
]

export function ChatWelcome({ onPick }: ChatWelcomeProps) {
  return (
    <div className="chatbot-welcome">
      <div className="chatbot-welcome__icon" aria-hidden="true">
        G
      </div>
      <h2 className="chatbot-welcome__title">How can I help?</h2>
      <p className="chatbot-welcome__subtitle">
        Ask anything — GAIOL orchestrates your connected models and picks the best answer.
      </p>
      <div className="chatbot-welcome__chips">
        {SUGGESTIONS.map((s) => (
          <button key={s} type="button" className="chatbot-chip" onClick={() => onPick(s)}>
            {s}
          </button>
        ))}
      </div>
    </div>
  )
}
