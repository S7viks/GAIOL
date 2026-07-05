import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

type ChatMessageContentProps = {
  content: string
  role: 'user' | 'assistant'
}

function shouldRenderMarkdown(role: 'user' | 'assistant', content: string): boolean {
  if (role !== 'assistant') return false
  const t = content.trimStart()
  if (t.startsWith('Error:')) return false
  return true
}

export function ChatMessageContent({ content, role }: ChatMessageContentProps) {
  if (!shouldRenderMarkdown(role, content)) {
    return <div className="chat-bubble__body chat-bubble__body--plain chatbot-msg__text">{content}</div>
  }

  return (
    <div className="chat-bubble__body chat-markdown chatbot-msg__text">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  )
}
