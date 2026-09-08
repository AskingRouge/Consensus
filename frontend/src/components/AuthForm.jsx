import { useState } from 'react'
import { flushSync } from 'react-dom'
import { useNavigate, useSearchParams } from 'react-router-dom'
import * as api from '../api'

/**
 * Combined login/signup screen. There's no Google OAuth yet (see the
 * project README) — this is real username/password auth against the
 * backend, gated by its email allowlist.
 */
export default function AuthForm({ onAuthenticated }) {
  const [mode, setMode] = useState('login') // 'login' | 'signup'
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState(null)
  const [submitting, setSubmitting] = useState(false)

  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  async function handleSubmit(e) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const result = mode === 'login' ? await api.login(email, password) : await api.signup(email, password)
      // flushSync forces the parent's `you` state update to commit before
      // we navigate. Without it, navigate()'s own (synchronous) history
      // update wins the race against the still-pending setYou, so
      // RequireAuth briefly re-evaluates with the *old* (unauthenticated)
      // value, bounces back to /login, and only then sees `you` update —
      // landing on "/" instead of the originally requested page.
      flushSync(() => onAuthenticated(result.email))
      navigate(searchParams.get('next') || '/', { replace: true })
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="auth-screen">
      <form className="auth-form" onSubmit={handleSubmit}>
        <h1>Consensus</h1>
        <p className="auth-form__subtitle">Collaborative LLM output review</p>

        <div className="auth-form__tabs">
          <button
            type="button"
            className={mode === 'login' ? 'is-active' : ''}
            onClick={() => setMode('login')}
          >
            Log in
          </button>
          <button
            type="button"
            className={mode === 'signup' ? 'is-active' : ''}
            onClick={() => setMode('signup')}
          >
            Sign up
          </button>
        </div>

        <label className="field">
          <span>Email</span>
          <input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>

        <label className="field">
          <span>Password</span>
          <input
            type="password"
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={mode === 'signup' ? 8 : undefined}
            required
          />
        </label>

        {mode === 'signup' && (
          <p className="field-hint">
            8+ characters. Your email must already be on the team's approved list.
          </p>
        )}

        {error && <p className="error-text">{error}</p>}

        <button type="submit" className="btn btn--primary" disabled={submitting}>
          {submitting ? 'Please wait…' : mode === 'login' ? 'Log in' : 'Create account'}
        </button>
      </form>
    </div>
  )
}
