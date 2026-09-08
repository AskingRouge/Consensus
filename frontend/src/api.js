// Thin wrapper around the backend REST API. Every call sends cookies
// (credentials: 'include') so the session cookie set by /api/auth/login
// or /api/auth/signup rides along automatically — there's no token to
// manage on the frontend at all.

export const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.status = status
  }
}

async function request(path, options = {}) {
  // FormData sets its own multipart Content-Type (with boundary) — don't
  // override it. Computed after the ...options spread so it always wins
  // regardless of what options contains.
  const isFormData = options.body instanceof FormData
  const headers = isFormData ? options.headers : { 'Content-Type': 'application/json', ...options.headers }

  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...options,
    headers,
  })

  if (res.status === 204) return null

  const isJson = res.headers.get('content-type')?.includes('application/json')
  const data = isJson ? await res.json() : null

  if (!res.ok) {
    throw new ApiError(data?.error || `Request failed (${res.status})`, res.status)
  }
  return data
}

const get = (path) => request(path)
const post = (path, body) => request(path, { method: 'POST', body: body != null ? JSON.stringify(body) : undefined })

// --- Auth ---
export const signup = (email, password) => post('/api/auth/signup', { email, password })
export const login = (email, password) => post('/api/auth/login', { email, password })
export const logout = () => post('/api/auth/logout')
export const me = () => get('/api/auth/me')

// --- Sessions ---
export const listSessions = () => get('/api/sessions')
export const createSession = (name, options = {}) => post('/api/sessions', { name, ...options })
export const getSession = (sessionId) => get(`/api/sessions/${sessionId}`)

// --- Items ---
export const listItems = (sessionId) => get(`/api/sessions/${sessionId}/items`)

export function uploadItems(sessionId, file) {
  const form = new FormData()
  form.append('file', file)
  return request(`/api/sessions/${sessionId}/items/upload`, { method: 'POST', body: form })
}

// --- Scores ---
// listScores returns every score in the session (all items, all
// reviewers) — the caller groups them by item itself.
export const listScores = (sessionId) => get(`/api/sessions/${sessionId}/scores`)

// submitScore creates or overwrites the caller's score for one item.
// rubricScores is {[dimensionName]: number}, covering every rubric
// dimension. Returns { score, flagged, disagreeing_dimensions }.
export const submitScore = (sessionId, itemId, rubricScores) =>
  post(`/api/sessions/${sessionId}/items/${itemId}/scores`, { rubric_scores: rubricScores })

export { ApiError }
