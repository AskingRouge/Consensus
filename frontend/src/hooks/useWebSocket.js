import { useEffect, useRef, useState, useCallback } from 'react'
import { API_BASE } from '../api'

const WS_BASE = API_BASE.replace(/^http/, 'ws')

const INITIAL_RECONNECT_DELAY_MS = 1000
const MAX_RECONNECT_DELAY_MS = 15000

/**
 * Manages the live connection to one review session: presence, and (from
 * Phase 5/6 on) score and comment updates. Reconnects automatically with
 * exponential backoff, and — since the hub broadcasts full snapshots, not
 * deltas — a client that just reconnected renders correctly from the very
 * next presence message without needing any special "catch up" logic.
 *
 * `viewItemId` is the item the caller is currently looking at (or null).
 * Whenever it changes, or the socket (re)connects, the current value is
 * sent to the server so presence reflects it — this is also how a dropped
 * connection resyncs its state automatically on reconnect, per the
 * project's non-functional requirements.
 */
export function useWebSocket(sessionId, viewItemId) {
  const [presence, setPresence] = useState([])
  const [scoreUpdates, setScoreUpdates] = useState([]) // Phase 5 populates these
  const [comments, setComments] = useState([]) // Phase 6 populates these
  const [status, setStatus] = useState('connecting') // 'connecting' | 'open' | 'closed'

  const socketRef = useRef(null)
  const viewItemIdRef = useRef(viewItemId)
  const reconnectDelayRef = useRef(INITIAL_RECONNECT_DELAY_MS)
  const reconnectTimerRef = useRef(null)
  const closedByCallerRef = useRef(false)

  useEffect(() => {
    viewItemIdRef.current = viewItemId
    sendViewItem(viewItemId)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [viewItemId])

  const sendViewItem = useCallback((itemId) => {
    const socket = socketRef.current
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: 'view_item', item_id: itemId || '' }))
    }
  }, [])

  useEffect(() => {
    if (!sessionId) return undefined

    closedByCallerRef.current = false

    function connect() {
      setStatus('connecting')
      const socket = new WebSocket(`${WS_BASE}/ws/sessions/${sessionId}`)
      socketRef.current = socket

      socket.onopen = () => {
        setStatus('open')
        reconnectDelayRef.current = INITIAL_RECONNECT_DELAY_MS
        // Resync: tell the server what we're viewing right now, in case
        // this is a reconnect after a dropped connection.
        sendViewItem(viewItemIdRef.current)
      }

      socket.onmessage = (event) => {
        let msg
        try {
          msg = JSON.parse(event.data)
        } catch {
          return
        }
        switch (msg.type) {
          case 'presence':
            setPresence(msg.payload || [])
            break
          case 'score_update':
            setScoreUpdates((prev) => [...prev, msg.payload])
            break
          case 'comment':
            setComments((prev) => [...prev, msg.payload])
            break
          default:
            break
        }
      }

      socket.onclose = () => {
        // If a newer socket has already replaced this one in socketRef
        // (e.g. this effect's cleanup ran and a fresh connect() happened
        // — as React StrictMode's dev-mode double-invoke does, or a fast
        // unmount/remount of the same session), this close event is for a
        // stale generation. Reconnecting here would create a duplicate
        // live connection alongside the newer one, which is exactly what
        // caused two presence entries for one browser tab in testing.
        if (socketRef.current !== socket) return
        setStatus('closed')
        if (closedByCallerRef.current) return
        reconnectTimerRef.current = setTimeout(() => {
          reconnectDelayRef.current = Math.min(reconnectDelayRef.current * 2, MAX_RECONNECT_DELAY_MS)
          connect()
        }, reconnectDelayRef.current)
      }

      socket.onerror = () => {
        socket.close()
      }
    }

    connect()

    return () => {
      closedByCallerRef.current = true
      clearTimeout(reconnectTimerRef.current)
      socketRef.current?.close()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessionId])

  return { presence, scoreUpdates, comments, status, sendViewItem }
}
