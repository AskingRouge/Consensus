import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as api from '../api'

export default function SessionList() {
  const [sessions, setSessions] = useState(null)
  const [error, setError] = useState(null)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    refresh()
  }, [])

  async function refresh() {
    try {
      setSessions(await api.listSessions())
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleCreate(e) {
    e.preventDefault()
    if (!name.trim()) return
    setCreating(true)
    setError(null)
    try {
      await api.createSession(name.trim())
      setName('')
      await refresh()
    } catch (err) {
      setError(err.message)
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>Sessions</h1>
        <p className="muted">Batches you've created or been invited to — private until you invite a teammate.</p>
      </div>

      <form className="create-session-form" onSubmit={handleCreate}>
        <input
          type="text"
          placeholder="e.g. Week 3 Course-Scope Eval"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <button type="submit" className="btn btn--primary" disabled={creating}>
          {creating ? 'Creating…' : 'New session'}
        </button>
      </form>

      {error && <p className="error-text">{error}</p>}

      {sessions === null && !error && (
        <div>
          <div className="skeleton skeleton-row" />
          <div className="skeleton skeleton-row" />
          <div className="skeleton skeleton-row" />
        </div>
      )}

      {sessions && sessions.length === 0 && (
        <div className="empty-state">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <rect x="3" y="4" width="18" height="4" rx="1.5" stroke="currentColor" strokeWidth="1.6" />
            <rect x="3" y="11" width="18" height="4" rx="1.5" stroke="currentColor" strokeWidth="1.6" />
            <rect x="3" y="18" width="10" height="2.5" rx="1.25" stroke="currentColor" strokeWidth="1.6" />
          </svg>
          <p>Nothing here yet — create a session above to get started, or ask a teammate to invite you to theirs.</p>
        </div>
      )}

      <ul className="session-list">
        {sessions?.map((s) => (
          <li key={s.id}>
            <Link to={`/sessions/${s.id}`} className="session-list__item">
              <span className="session-list__name">{s.name}</span>
              <span className="session-list__meta">
                {s.created_by} · {new Date(s.created_at).toLocaleDateString()}
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  )
}
