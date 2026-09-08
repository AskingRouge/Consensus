import { useState } from 'react'

/**
 * One item's discussion thread: existing comments (oldest first) plus a
 * form to post a new one. Posting goes through the parent's onPost (which
 * calls the API and applies the result the same way a live WebSocket
 * comment would be applied), so this component only owns its own draft
 * text — not the comment list itself.
 */
export default function CommentThread({ comments, you, onPost }) {
  const [text, setText] = useState('')
  const [posting, setPosting] = useState(false)
  const [error, setError] = useState(null)

  async function handleSubmit(e) {
    e.preventDefault()
    const trimmed = text.trim()
    if (!trimmed) return

    setPosting(true)
    setError(null)
    try {
      await onPost(trimmed)
      setText('')
    } catch (err) {
      setError(err.message)
    } finally {
      setPosting(false)
    }
  }

  return (
    <div className="comment-thread">
      {comments.length === 0 ? (
        <p className="muted small">No comments yet — start the discussion below.</p>
      ) : (
        <ul className="comment-list">
          {comments.map((c) => (
            <li key={c.id} className={`comment${c.author_email === you ? ' is-you' : ''}`}>
              <div className="comment__meta">
                <span className="comment__author">{c.author_email === you ? 'You' : c.author_email}</span>
                <span className="comment__time">{formatTime(c.created_at)}</span>
              </div>
              <p className="comment__text">{c.text}</p>
            </li>
          ))}
        </ul>
      )}

      <form className="comment-form" onSubmit={handleSubmit}>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Add a comment…"
          rows={2}
          disabled={posting}
          maxLength={4000}
        />
        {error && <p className="error-text">{error}</p>}
        <button type="submit" className="btn btn--primary" disabled={posting || !text.trim()}>
          {posting ? 'Posting…' : 'Post comment'}
        </button>
      </form>
    </div>
  )
}

function formatTime(iso) {
  return new Date(iso).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
