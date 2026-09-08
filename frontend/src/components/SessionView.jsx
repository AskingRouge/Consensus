import { useEffect, useState, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import * as api from '../api'
import { useWebSocket } from '../hooks/useWebSocket'
import CommentThread from './CommentThread'
import ItemCard from './ItemCard'
import PresenceIndicator from './PresenceIndicator'
import RubricScorer from './RubricScorer'

const STATUS_LABEL = {
  connecting: 'Connecting…',
  open: 'Live',
  closed: 'Reconnecting…',
}

export default function SessionView({ you }) {
  const { sessionId } = useParams()
  const [session, setSession] = useState(null)
  const [items, setItems] = useState(null)
  const [scoresByItem, setScoresByItem] = useState({}) // item id -> Score[]
  const [commentsByItem, setCommentsByItem] = useState({}) // item id -> Comment[]
  const [error, setError] = useState(null)
  const [selectedItemId, setSelectedItemId] = useState(null)
  const [flaggedOnly, setFlaggedOnly] = useState(false)
  const fileInputRef = useRef(null)
  const [uploading, setUploading] = useState(false)

  // A score update — whether it arrives live over the socket or comes
  // straight back from our own submit call — always means "replace this
  // reviewer's score on this item, and refresh the item's flag". Sharing
  // one function for both keeps a reviewer's own view correct even if
  // their WebSocket happens to be disconnected right when they submit.
  function applyScoreUpdate(payload) {
    const { score, flagged, disagreeing_dimensions } = payload
    setScoresByItem((prev) => {
      const existing = prev[score.item_id] || []
      const idx = existing.findIndex((s) => s.reviewer_email === score.reviewer_email)
      const next = idx === -1 ? [...existing, score] : existing.map((s, i) => (i === idx ? score : s))
      return { ...prev, [score.item_id]: next }
    })
    setItems((prev) =>
      prev
        ? prev.map((i) => (i.id === score.item_id ? { ...i, flagged, disagreeing_dimensions } : i))
        : prev,
    )
  }

  // A comment can arrive twice for the poster — once as the direct REST
  // response, once as the live broadcast echoed back to them — so this
  // is a dedupe-by-id append, not a plain append.
  function applyComment(comment) {
    setCommentsByItem((prev) => {
      const existing = prev[comment.item_id] || []
      if (existing.some((c) => c.id === comment.id)) return prev
      return { ...prev, [comment.item_id]: [...existing, comment] }
    })
  }

  const { presence, status } = useWebSocket(sessionId, selectedItemId, {
    onScoreUpdate: applyScoreUpdate,
    onComment: applyComment,
  })

  useEffect(() => {
    setSelectedItemId(null)
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId])

  async function load() {
    try {
      const [sessionData, itemsData, scoresData, commentsData] = await Promise.all([
        api.getSession(sessionId),
        api.listItems(sessionId),
        api.listScores(sessionId),
        api.listComments(sessionId),
      ])
      setSession(sessionData)
      setItems(itemsData)

      const groupedScores = {}
      for (const s of scoresData) {
        ;(groupedScores[s.item_id] ||= []).push(s)
      }
      setScoresByItem(groupedScores)

      const groupedComments = {}
      for (const c of commentsData) {
        ;(groupedComments[c.item_id] ||= []).push(c)
      }
      setCommentsByItem(groupedComments)
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleUpload(e) {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    setError(null)
    try {
      await api.uploadItems(sessionId, file)
      setItems(await api.listItems(sessionId))
    } catch (err) {
      setError(err.message)
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  async function handleSubmitScore(itemId, rubricScores) {
    const result = await api.submitScore(sessionId, itemId, rubricScores)
    applyScoreUpdate(result)
  }

  async function handlePostComment(itemId, text) {
    const comment = await api.createComment(sessionId, itemId, text)
    applyComment(comment)
  }

  if (error) {
    return (
      <div className="page">
        <p className="error-text">{error}</p>
        <Link to="/">← Back to sessions</Link>
      </div>
    )
  }

  if (!session || !items) {
    return (
      <div className="page">
        <p className="muted">Loading…</p>
      </div>
    )
  }

  const selectedItem = items.find((i) => i.id === selectedItemId) || null
  const viewingListEntries = presence.filter((p) => !p.item_id)
  const flaggedCount = items.filter((i) => i.flagged).length
  const visibleItems = flaggedOnly ? items.filter((i) => i.flagged) : items

  return (
    <div className="page page--session">
      <div className="session-header">
        <div>
          <Link to="/" className="back-link">← All sessions</Link>
          <h1>{session.name}</h1>
          <p className="muted">
            {items.length} item{items.length === 1 ? '' : 's'} · created by {session.created_by}
          </p>
        </div>
        <div className="session-header__right">
          <span className={`ws-status ws-status--${status}`}>{STATUS_LABEL[status]}</span>
          <PresenceIndicator presence={presence} itemId={undefined} you={you} />
          <div className="export-links">
            <a className="btn btn--secondary" href={`${api.API_BASE}/api/sessions/${sessionId}/export?format=json`}>
              Export JSON
            </a>
            <a className="btn btn--secondary" href={`${api.API_BASE}/api/sessions/${sessionId}/export?format=csv`}>
              Export CSV
            </a>
          </div>
        </div>
      </div>

      <div className="session-body">
        <div className="item-list-pane">
          <div className="item-list-pane__toolbar">
            <label className="btn btn--secondary">
              {uploading ? 'Uploading…' : 'Upload items (JSON/CSV)'}
              <input
                ref={fileInputRef}
                type="file"
                accept=".json,.csv"
                onChange={handleUpload}
                disabled={uploading}
                hidden
              />
            </label>
            {flaggedCount > 0 && (
              <label className="flagged-filter">
                <input
                  type="checkbox"
                  checked={flaggedOnly}
                  onChange={(e) => setFlaggedOnly(e.target.checked)}
                />
                Needs discussion only ({flaggedCount})
              </label>
            )}
            {viewingListEntries.length > 0 && (
              <span className="muted small">{viewingListEntries.length} browsing the list</span>
            )}
          </div>

          {items.length === 0 ? (
            <p className="muted">No items yet — upload a batch to get started.</p>
          ) : visibleItems.length === 0 ? (
            <p className="muted">No flagged items.</p>
          ) : (
            <div className="item-list">
              {visibleItems.map((item) => (
                <ItemCard
                  key={item.id}
                  item={item}
                  presence={presence}
                  you={you}
                  isSelected={item.id === selectedItemId}
                  onSelect={(id) => setSelectedItemId(id === selectedItemId ? null : id)}
                />
              ))}
            </div>
          )}
        </div>

        <div className="item-detail-pane">
          {selectedItem ? (
            <ItemDetail
              item={selectedItem}
              session={session}
              presence={presence}
              you={you}
              scores={scoresByItem[selectedItem.id] || []}
              onSubmitScore={handleSubmitScore}
              comments={commentsByItem[selectedItem.id] || []}
              onPostComment={handlePostComment}
            />
          ) : (
            <p className="muted item-detail-pane__placeholder">
              Select an item to view it in full and score it.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

function ItemDetail({ item, session, presence, you, scores, onSubmitScore, comments, onPostComment }) {
  const myScore = scores.find((s) => s.reviewer_email === you)
  const [draft, setDraft] = useState(myScore?.rubric_scores || {})
  const [submitting, setSubmitting] = useState(false)
  const [scoreError, setScoreError] = useState(null)

  // Reset the draft when switching to a different item — but NOT when
  // `scores` changes for the current item (e.g. another reviewer's score
  // arrives live), which would otherwise wipe out an in-progress edit.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    setDraft(myScore?.rubric_scores || {})
    setScoreError(null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [item.id])

  const viewers = presence.filter((p) => p.item_id === item.id && p.reviewer_email !== you)
  const isComplete = session.rubric.every((dim) => draft[dim.name] !== undefined)

  async function handleSubmit() {
    setSubmitting(true)
    setScoreError(null)
    try {
      await onSubmitScore(item.id, draft)
    } catch (err) {
      setScoreError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="item-detail">
      <div className="item-detail__header">
        <h2>Item #{item.order_index + 1}</h2>
        <PresenceIndicator presence={presence} itemId={item.id} you={you} />
      </div>

      {item.flagged && (
        <p className="flagged-badge">
          Needs discussion — reviewers disagree on {item.disagreeing_dimensions?.join(', ')}
        </p>
      )}

      {viewers.length > 0 && (
        <p className="muted small">Also viewing: {viewers.map((v) => v.reviewer_email).join(', ')}</p>
      )}

      <section className="item-detail__field">
        <h3>Question</h3>
        <p>{item.question}</p>
      </section>

      <section className="item-detail__field">
        <h3>Retrieved context</h3>
        <p className="item-detail__context">{item.retrieved_context}</p>
      </section>

      <section className="item-detail__field">
        <h3>Generated answer</h3>
        <p>{item.generated_answer}</p>
      </section>

      <section className="item-detail__field">
        <h3>Your score</h3>
        <RubricScorer
          rubric={session.rubric}
          draft={draft}
          disabled={submitting}
          onChange={(name, value) => setDraft((d) => ({ ...d, [name]: value }))}
        />
        {scoreError && <p className="error-text">{scoreError}</p>}
        <button type="button" className="btn btn--primary" disabled={!isComplete || submitting} onClick={handleSubmit}>
          {submitting ? 'Submitting…' : myScore ? 'Update score' : 'Submit score'}
        </button>
      </section>

      <section className="item-detail__field">
        <h3>All scores ({scores.length})</h3>
        {scores.length === 0 ? (
          <p className="muted small">No one has scored this item yet.</p>
        ) : (
          <table className="scores-table">
            <thead>
              <tr>
                <th>Reviewer</th>
                {session.rubric.map((dim) => (
                  <th key={dim.name}>{dim.name}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {scores.map((s) => (
                <tr key={s.reviewer_email} className={s.reviewer_email === you ? 'is-you' : ''}>
                  <td>{s.reviewer_email === you ? `${s.reviewer_email} (you)` : s.reviewer_email}</td>
                  {session.rubric.map((dim) => (
                    <td key={dim.name}>{formatScoreValue(dim, s.rubric_scores[dim.name])}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="item-detail__field">
        <h3>Discussion ({comments.length})</h3>
        {/* Keyed by item.id so switching items remounts the thread fresh
            — a draft comment shouldn't follow you to a different item. */}
        <CommentThread
          key={item.id}
          comments={comments}
          you={you}
          onPost={(text) => onPostComment(item.id, text)}
        />
      </section>
    </div>
  )
}

function formatScoreValue(dim, value) {
  if (value === undefined) return '—'
  return dim.type === 'pass_fail' ? (value === 1 ? 'Pass' : 'Fail') : value
}
