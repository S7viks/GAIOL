export function ChatTypingIndicator() {
  return (
    <div className="chatbot-msg chatbot-msg--assistant chatbot-msg--typing" aria-live="polite" aria-label="Assistant is thinking">
      <div className="chatbot-msg__avatar chatbot-msg__avatar--bot" aria-hidden="true">
        G
      </div>
      <div className="chatbot-msg__bubble">
        <span className="chatbot-typing">
          <span />
          <span />
          <span />
        </span>
      </div>
    </div>
  )
}
