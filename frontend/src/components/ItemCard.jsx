import PresenceIndicator from './PresenceIndicator'

/**
 * One row in a session's item list: the question, a preview of the
 * retrieved context and generated answer, and who (if anyone) is
 * currently viewing this item. Click to open it in the detail panel.
 *
 * Rubric scoring (Phase 5) and the comment thread (Phase 6) live in the
 * detail panel in SessionView, not here — this card is just the list
 * entry.
 */
export default function ItemCard({ item, presence, you, isSelected, onSelect }) {
  return (
    <button
      type="button"
      className={`item-card${isSelected ? ' item-card--selected' : ''}`}
      onClick={() => onSelect(item.id)}
    >
      <div className="item-card__header">
        <span className="item-card__index">#{item.order_index + 1}</span>
        <PresenceIndicator presence={presence} itemId={item.id} you={you} />
      </div>
      <p className="item-card__question">{item.question}</p>
      <p className="item-card__answer">{item.generated_answer}</p>
    </button>
  )
}
