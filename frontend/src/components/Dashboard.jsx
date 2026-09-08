import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as api from '../api'
import { computeStats } from '../lib/stats'
import { BarChart, ProgressRing } from './charts'

/**
 * The landing page: real stats and charts computed from your actual
 * sessions (not placeholder numbers), a "needs your attention" task list
 * that links straight to the item in question, and the session
 * list/creation form below. Everything here is derived from data you
 * already have — nothing here is stored separately from the sessions
 * themselves.
 */
export default function Dashboard({ you }) {
  const [sessions, setSessions] = useState(null)
  const [stats, setStats] = useState(null)
  const [error, setError] = useState(null)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    refresh()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function refresh() {
    try {
      const sessionsList = await api.listSessions()
      setSessions(sessionsList)
      setStats(await computeStats(sessionsList, you))
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
        <h1>Dashboard</h1>
        <p className="muted">An overview of your review sessions — what's flagged, what's waiting on you.</p>
      </div>

      {error && <p className="error-text">{error}</p>}

      {!stats ? (
        <div className="stat-row">
          <div className="skeleton" style={{ height: 92 }} />
          <div className="skeleton" style={{ height: 92 }} />
          <div className="skeleton" style={{ height: 92 }} />
          <div className="skeleton" style={{ height: 92 }} />
        </div>
      ) : (
        <>
          <div className="stat-row">
            <StatCard
              icon={<IconLayers />}
              value={stats.sessionCount}
              label={stats.sessionCount === 1 ? 'Session' : 'Sessions'}
            />
            <StatCard icon={<IconDocument />} value={stats.totalItems} label="Items total" />
            <StatCard
              icon={<IconFlag />}
              value={stats.totalFlagged}
              label="Needs discussion"
              tone={stats.totalFlagged > 0 ? 'danger' : undefined}
            />
            <StatCard
              icon={<IconClock />}
              value={stats.totalPending}
              label="Awaiting your score"
              tone={stats.totalPending > 0 ? 'warning' : undefined}
            />
          </div>

          <div className="dashboard-grid">
            <div className="panel">
              <h2 className="panel__title">Needs your attention</h2>
              {stats.tasks.length === 0 ? (
                <div className="empty-state">
                  <svg width="32" height="32" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path
                      d="M9 12l2 2 4-4M12 22c5.5-1 8-5 8-10V6l-8-3-8 3v6c0 5 2.5 9 8 10z"
                      stroke="currentColor"
                      strokeWidth="1.6"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <p>Nothing waiting on you — every item is scored and nothing's flagged.</p>
                </div>
              ) : (
                <ul className="task-list">
                  {stats.tasks.map((t) => (
                    <li key={`${t.sessionId}-${t.itemId}-${t.kind}`}>
                      <Link to={`/sessions/${t.sessionId}?item=${t.itemId}`} className="task-item">
                        <span className={`task-item__dot task-item__dot--${t.kind}`} />
                        <span className="task-item__text">
                          <span className="task-item__title">{t.question}</span>
                          <span className="task-item__meta">
                            {t.sessionName} · {t.kind === 'flagged' ? 'Needs discussion' : 'Awaiting your score'}
                          </span>
                        </span>
                      </Link>
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="panel">
              <h2 className="panel__title">Scoring progress</h2>
              <div className="progress-row">
                <ProgressRing percent={stats.scoredPercent} label="Items scored" tone="success" />
                <ProgressRing percent={stats.agreementPercent} label="In agreement" tone="auto" />
              </div>

              <h3 className="panel__subtitle">Disagreement by dimension</h3>
              <BarChart
                data={stats.dimensionBars}
                tone="danger"
                emptyText="No disagreement yet — nothing's been flagged."
              />
            </div>
          </div>
        </>
      )}

      <div className="page-header page-header--tight">
        <h2>Your sessions</h2>
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

      {sessions === null && !error && (
        <div>
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

function StatCard({ icon, value, label, tone }) {
  return (
    <div className={`stat-card${tone ? ` stat-card--${tone}` : ''}`}>
      <span className="stat-card__icon">{icon}</span>
      <span className="stat-card__value">{value}</span>
      <span className="stat-card__label">{label}</span>
    </div>
  )
}

// --- Icons ---

function IconLayers() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M12 2l9 5-9 5-9-5 9-5z" stroke="currentColor" strokeWidth="1.6" strokeLinejoin="round" />
      <path d="M3 12l9 5 9-5M3 17l9 5 9-5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
function IconDocument() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <rect x="4" y="3" width="16" height="18" rx="2" stroke="currentColor" strokeWidth="1.6" />
      <path d="M8 8h8M8 12h8M8 16h5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  )
}
function IconFlag() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M5 3v18M5 4h13l-3 4 3 4H5" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
function IconClock() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.6" />
      <path d="M12 7v5l3.5 2" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}
