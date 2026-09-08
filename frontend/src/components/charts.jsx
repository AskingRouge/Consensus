import { useEffect, useState } from 'react'

/**
 * Simple horizontal bar chart, built from real aggregated data (no
 * charting library — a handful of divs is all this needs). Bars animate
 * in from zero width on mount rather than appearing at full length
 * instantly.
 */
export function BarChart({ data, emptyText = 'Nothing to show yet.' }) {
  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true))
    return () => cancelAnimationFrame(id)
  }, [])

  if (data.length === 0) {
    return <p className="muted small">{emptyText}</p>
  }

  const max = Math.max(1, ...data.map((d) => d.value))

  return (
    <div className="bar-chart">
      {data.map((d) => (
        <div className="bar-chart__row" key={d.label}>
          <span className="bar-chart__label" title={d.label}>
            {d.label}
          </span>
          <div className="bar-chart__track">
            <div
              className="bar-chart__fill"
              style={{ width: mounted ? `${(d.value / max) * 100}%` : 0 }}
            />
          </div>
          <span className="bar-chart__value">{d.value}</span>
        </div>
      ))}
    </div>
  )
}

/**
 * Circular progress ring (SVG stroke-dashoffset) — used for "% of items
 * scored" style completion stats. Animates in on mount the same way.
 */
export function ProgressRing({ percent, size = 92, stroke = 10, label }) {
  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true))
    return () => cancelAnimationFrame(id)
  }, [])

  const r = (size - stroke) / 2
  const c = 2 * Math.PI * r
  const pct = mounted ? percent : 0
  const offset = c - (pct / 100) * c

  return (
    <div className="progress-ring">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="var(--color-surface-sunken)" strokeWidth={stroke} />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          stroke="url(#ringGradient)"
          strokeWidth={stroke}
          strokeDasharray={c}
          strokeDashoffset={offset}
          strokeLinecap="round"
          transform={`rotate(-90 ${size / 2} ${size / 2})`}
          className="progress-ring__fill"
        />
        <defs>
          <linearGradient id="ringGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="var(--color-primary)" />
            <stop offset="100%" stopColor="var(--color-accent)" />
          </linearGradient>
        </defs>
        <text x="50%" y="50%" textAnchor="middle" dy="0.35em" className="progress-ring__text">
          {Math.round(percent)}%
        </text>
      </svg>
      {label && <span className="progress-ring__label">{label}</span>}
    </div>
  )
}
