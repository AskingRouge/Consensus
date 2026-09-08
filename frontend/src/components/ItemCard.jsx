import PresenceIndicator from './PresenceIndicator'

/**
 * One row in a session's item list: the question, a preview of the
 * retrieved context and generated answer, whether it's flagged for
 * disagreement, and who (if anyone) is currently viewing it. Click to
 * open it in the detail panel, where it's scored (RubricScorer) and
 * discussed (CommentThread, Phase 6).
 */
export default function ItemCard({ item, presence, you, isSelected, onSelect }) {
  return (
    <button
      type="button"
      className={`item-card${isSelected ? ' item-card--selected' : ''}`}
      onClick={() => onSelect(item.id)}
    >
      <div className="item-card__header">
        <span className="item-card__index">
          #{item.order_index + 1}
          {item.flagged && <span className="item-card__flag" title="Needs discussion">Needs discussion</span>}
        </span>
        <PresenceIndicator presence={presence} itemId={item.id} you={you} />
      </div>
      <p className="item-card__question">{item.question}</p>
      <p className="item-card__answer">{item.generated_answer}</p>
    </button>
  )
}
