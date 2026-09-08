import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import * as api from './api'
import AuthForm from './components/AuthForm'
import Sidebar from './components/Sidebar'
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

  const routes = (
    <Routes>
      <Route path="/login" element={you ? <Navigate to="/" replace /> : <AuthForm onAuthenticated={setYou} />} />
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
  )

  async function handleLogout() {
    await api.logout()
    setYou(null)
  }

  return (
    <BrowserRouter>
      {you ? (
        <div className="app-shell">
          <Sidebar you={you} onLogout={handleLogout} />
          <main className="app-main">{routes}</main>
        </div>
      ) : (
        routes
      )}
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
