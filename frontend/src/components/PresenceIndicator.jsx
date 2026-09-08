// Small colored initials badge for one reviewer, plus a row of them. Color
// is derived deterministically from the email so the same person always
// gets the same color across the app (and across page reloads) without
// storing anything. Chosen to stay visually distinct from the brand's
// violet/indigo (used everywhere for buttons and active states) so an
// avatar never reads as an interactive element.
const COLORS = ['#E1725B', '#3FAE7A', '#C98C2E', '#2E9BB0', '#D6478B', '#3B82C4']

function colorFor(email) {
  let hash = 0
  for (let i = 0; i < email.length; i++) hash = (hash * 31 + email.charCodeAt(i)) >>> 0
  return COLORS[hash % COLORS.length]
}

function initials(email) {
  const name = email.split('@')[0]
  return name.slice(0, 2).toUpperCase()
}

export function Avatar({ email, title }) {
  return (
    <span
      className="avatar"
      style={{ backgroundColor: colorFor(email) }}
      title={title || email}
    >
      {initials(email)}
    </span>
  )
}

/**
 * Renders every reviewer currently present, filtered to those viewing
 * `itemId` (or, if itemId is undefined, everyone in the session).
 * `you` is the current user's email, shown with a "(you)" suffix in the
 * tooltip so it's obvious which badge is yours.
 */
export default function PresenceIndicator({ presence, itemId, you, emptyText = null }) {
  const entries = itemId === undefined ? presence : presence.filter((p) => (p.item_id || '') === itemId)

  if (entries.length === 0) {
    return emptyText ? <span className="presence-empty">{emptyText}</span> : null
  }

  return (
    <div className="presence-row">
      {entries.map((p, i) => (
        // Keyed by index, not email: the same reviewer can legitimately
        // appear more than once (e.g. two open tabs), and this is a
        // stateless decorative badge with nothing per-instance to
        // preserve across reorders.
        <Avatar
          key={i}
          email={p.reviewer_email}
          title={p.reviewer_email === you ? `${p.reviewer_email} (you)` : p.reviewer_email}
        />
      ))}
    </div>
  )
}
