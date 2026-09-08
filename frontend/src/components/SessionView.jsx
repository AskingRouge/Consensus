import { useEffect, useState, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import * as api from '../api'
import { useWebSocket } from '../hooks/useWebSocket'
import ItemCard from './ItemCard'
import PresenceIndicator from './PresenceIndicator'

const STATUS_LABEL = {
  connecting: 'Connecting…',
  open: 'Live',
  closed: 'Reconnecting…',
}

export default function SessionView({ you }) {
  const { sessionId } = useParams()
  const [session, setSession] = useState(null)
  const [items, setItems] = useState(null)
  const [error, setError] = useState(null)
  const [selectedItemId, setSelectedItemId] = useState(null)
  const fileInputRef = useRef(null)
  const [uploading, setUploading] = useState(false)

  const { presence, status } = useWebSocket(sessionId, selectedItemId)

  useEffect(() => {
    setSelectedItemId(null)
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId])

  async function load() {
    try {
      const [sessionData, itemsData] = await Promise.all([
        api.getSession(sessionId),
        api.listItems(sessionId),
      ])
      setSession(sessionData)
      setItems(itemsData)
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
          <PresenceIndicator
            presence={presence}
            itemId={undefined}
            you={you}
            emptyText={null}
          />
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
            {viewingListEntries.length > 0 && (
              <span className="muted small">
                {viewingListEntries.length} browsing the list
              </span>
            )}
          </div>

          {items.length === 0 ? (
            <p className="muted">No items yet — upload a batch to get started.</p>
          ) : (
            <div className="item-list">
              {items.map((item) => (
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
            <ItemDetail item={selectedItem} session={session} presence={presence} you={you} />
          ) : (
            <p className="muted item-detail-pane__placeholder">
              Select an item to view it in full and (soon) score it.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

function ItemDetail({ item, session, presence, you }) {
  const viewers = presence.filter((p) => p.item_id === item.id && p.reviewer_email !== you)

  return (
    <div className="item-detail">
      <div className="item-detail__header">
        <h2>Item #{item.order_index + 1}</h2>
        <PresenceIndicator presence={presence} itemId={item.id} you={you} />
      </div>

      {viewers.length > 0 && (
        <p className="muted small">
          Also viewing: {viewers.map((v) => v.reviewer_email).join(', ')}
        </p>
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

      <section className="item-detail__placeholder">
        <h3>Rubric scoring</h3>
        <p className="muted">
          {session.rubric.map((r) => r.name).join(' · ')} — scoring UI arrives in Phase 5.
        </p>
      </section>

      <section className="item-detail__placeholder">
        <h3>Discussion</h3>
        <p className="muted">Comment thread arrives in Phase 6.</p>
      </section>
    </div>
  )
}
