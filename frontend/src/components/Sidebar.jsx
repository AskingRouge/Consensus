import { Link, useLocation } from 'react-router-dom'
import Logo from './Logo'
import { Avatar } from './PresenceIndicator'

/**
 * Persistent app shell nav: brand mark, the single "Sessions" destination,
 * and the signed-in user at the bottom. Kept deliberately simple — there's
 * one real place to go — but it's what makes the app read as a real
 * product rather than a page.
 */
export default function Sidebar({ you, onLogout }) {
  const location = useLocation()
  const onSessions = location.pathname === '/' || location.pathname.startsWith('/sessions')

  return (
    <aside className="sidebar">
      <Link to="/" className="sidebar__brand">
        <Logo size={30} />
      </Link>

      <nav className="sidebar__nav">
        <Link to="/" className={`sidebar__nav-item${onSessions ? ' is-active' : ''}`}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <rect x="3" y="4" width="18" height="4" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
            <rect x="3" y="11" width="18" height="4" rx="1.5" stroke="currentColor" strokeWidth="1.8" />
            <rect x="3" y="18" width="10" height="2.5" rx="1.25" stroke="currentColor" strokeWidth="1.8" />
          </svg>
          Sessions
        </Link>
      </nav>

      <div className="sidebar__user">
        <Avatar email={you} />
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
      </div>
    </aside>
  )
}
