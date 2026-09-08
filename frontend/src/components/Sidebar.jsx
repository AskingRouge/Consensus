import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import * as api from '../api'
import { computeStats } from '../lib/stats'
import Logo from './Logo'
import { Avatar } from './PresenceIndicator'

const COLLAPSE_KEY = 'consensus:sidebar-collapsed'

/**
 * Persistent app shell nav: brand mark, a collapsible rail, quick-glance
 * stats pulled from the same numbers the dashboard shows (so you can see
 * what needs attention without leaving whatever session you're in), a
 * live "Recent" list of your sessions, and the signed-in user at the
 * bottom.
 */
export default function Sidebar({ you, onLogout }) {
  const location = useLocation()
  const onSessions = location.pathname === '/' || location.pathname.startsWith('/sessions')

  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem(COLLAPSE_KEY) === '1'
    } catch {
      return false
    }
  })
  const [recent, setRecent] = useState(null)
  const [stats, setStats] = useState(null)

  useEffect(() => {
    try {
      localStorage.setItem(COLLAPSE_KEY, collapsed ? '1' : '0')
    } catch {
      // private browsing / storage disabled — collapse state just won't persist
    }
  }, [collapsed])

  // Refetch on every navigation — cheap, and it's how a session you were
  // just invited to (or just scored something in) shows up here without a
  // reload. Same sessions list feeds both the "Recent" list and the
  // quick-stat counts, so this is one fetch, not two.
  useEffect(() => {
    let cancelled = false
    api.listSessions().then((sessionsList) => {
      if (cancelled) return
      setRecent(sessionsList)
      computeStats(sessionsList, you).then((s) => {
        if (!cancelled) setStats(s)
      }, () => {})
    }, () => {})
    return () => {
      cancelled = true
    }
  }, [location.pathname, you])

  const badgeCount = stats ? stats.totalFlagged + stats.totalPending : 0

  return (
    <aside className={`sidebar${collapsed ? ' is-collapsed' : ''}`}>
      <div className="sidebar__top">
        <Link to="/" className="sidebar__brand">
          <Logo size={28} wordmark={!collapsed} />
        </Link>
        <button
          type="button"
          className="sidebar__collapse"
          onClick={() => setCollapsed((c) => !c)}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d={collapsed ? 'M9 6l6 6-6 6' : 'M15 6l-6 6 6 6'}
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
      </div>

      <nav className="sidebar__nav">
        <Link to="/" className={`sidebar__nav-item${onSessions ? ' is-active' : ''}`} title="Dashboard">
          <span className="sidebar__nav-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <rect x="3" y="3" width="8" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
              <rect x="13" y="3" width="8" height="5" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
              <rect x="13" y="10" width="8" height="11" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
              <rect x="3" y="13" width="8" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
            </svg>
            {collapsed && badgeCount > 0 && <span className="sidebar__nav-badge">{badgeCount}</span>}
          </span>
          <span className="sidebar__nav-label">Dashboard</span>
        </Link>
      </nav>

      {!collapsed && stats && (stats.totalFlagged > 0 || stats.totalPending > 0) && (
        <div className="sidebar__quickstats">
          {stats.totalFlagged > 0 && (
            <Link to="/" className="sidebar__quickstat sidebar__quickstat--danger">
              <span className="sidebar__quickstat-dot" />
              Needs discussion
              <span className="sidebar__quickstat-count">{stats.totalFlagged}</span>
            </Link>
          )}
          {stats.totalPending > 0 && (
            <Link to="/" className="sidebar__quickstat sidebar__quickstat--warning">
              <span className="sidebar__quickstat-dot" />
              Awaiting your score
              <span className="sidebar__quickstat-count">{stats.totalPending}</span>
            </Link>
          )}
        </div>
      )}

      {!collapsed && recent && recent.length > 0 && (
        <div className="sidebar__recent">
          <span className="sidebar__section-label">Recent</span>
          <div className="sidebar__recent-list">
            {recent.slice(0, 5).map((s) => (
              <Link
                key={s.id}
                to={`/sessions/${s.id}`}
                className={`sidebar__recent-item${location.pathname === `/sessions/${s.id}` ? ' is-active' : ''}`}
                title={s.name}
              >
                {s.name}
              </Link>
            ))}
          </div>
        </div>
      )}

      <div className="sidebar__user">
        <Avatar email={you} />
        {!collapsed && (
          <>
            <span className="sidebar__user-email" title={you}>
              {you}
            </span>
            <button type="button" className="sidebar__logout" onClick={onLogout} title="Log out" aria-label="Log out">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path
                  d="M15 17l5-5-5-5M20 12H9M12 19H6a2 2 0 01-2-2V7a2 2 0 012-2h6"
                  stroke="currentColor"
                  strokeWidth="1.8"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>
          </>
        )}
      </div>
    </aside>
  )
}
