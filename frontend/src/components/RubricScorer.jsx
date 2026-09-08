// One clickable control per rubric dimension — a 1..scale_max button row
// for "scale" dimensions, Pass/Fail buttons for "pass_fail" ones. `draft`
// is {[dimensionName]: number|undefined}; the parent owns that state and
// decides when to submit it, so this component has no state of its own.
export default function RubricScorer({ rubric, draft, onChange, disabled }) {
  return (
    <div className="rubric-scorer">
      {rubric.map((dim) => (
        <div className="rubric-scorer__dimension" key={dim.name}>
          <span className="rubric-scorer__label">{dim.name}</span>
          {dim.type === 'scale' ? (
            <div className="scale-buttons">
              {Array.from({ length: dim.scale_max }, (_, i) => i + 1).map((n) => (
                <button
                  key={n}
                  type="button"
                  disabled={disabled}
                  className={`scale-btn${draft[dim.name] === n ? ' is-selected' : ''}`}
                  onClick={() => onChange(dim.name, n)}
                >
                  {n}
                </button>
              ))}
            </div>
          ) : (
            <div className="pass-fail-buttons">
              <button
                type="button"
                disabled={disabled}
                className={`pf-btn pf-btn--pass${draft[dim.name] === 1 ? ' is-selected' : ''}`}
                onClick={() => onChange(dim.name, 1)}
              >
                Pass
              </button>
              <button
                type="button"
                disabled={disabled}
                className={`pf-btn pf-btn--fail${draft[dim.name] === 0 ? ' is-selected' : ''}`}
                onClick={() => onChange(dim.name, 0)}
              >
                Fail
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
