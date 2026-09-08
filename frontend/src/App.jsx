import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation, Link } from 'react-router-dom'
import * as api from './api'
import AuthForm from './components/AuthForm'
import SessionList from './components/SessionList'
import SessionView from './components/SessionView'

export default function App() {
  // undefined = still checking; null = logged out; string = the user's email
  const [you, setYou] = useState(undefined)

  useEffect(() => {
    api.me().then(
      (res) => setYou(res.email),
      () => setYou(null),
    )
  }, [])

  if (you === undefined) {
    return (
      <div className="page">
        <p className="muted">Loading…</p>
      </div>
    )
  }

  return (
    <BrowserRouter>
      <TopBar you={you} onLogout={() => setYou(null)} />
      <Routes>
        <Route
          path="/login"
          element={you ? <Navigate to="/" replace /> : <AuthForm onAuthenticated={setYou} />}
        />
        <Route
          path="/"
          element={
            <RequireAuth you={you}>
              <SessionList />
            </RequireAuth>
          }
        />
        <Route
          path="/sessions/:sessionId"
          element={
            <RequireAuth you={you}>
              <SessionView you={you} />
            </RequireAuth>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}

function RequireAuth({ you, children }) {
  const location = useLocation()
  if (!you) {
    const next = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/login?next=${next}`} replace />
  }
  return children
}

function TopBar({ you, onLogout }) {
  if (!you) return null

  async function handleLogout() {
    await api.logout()
    onLogout()
  }

  return (
    <header className="top-bar">
      <Link to="/" className="top-bar__brand">
        Consensus
      </Link>
      <div className="top-bar__right">
        <span className="muted small">{you}</span>
        <button type="button" className="btn btn--text" onClick={handleLogout}>
          Log out
        </button>
      </div>
    </header>
  )
}
