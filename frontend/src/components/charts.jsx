import { useEffect, useId, useState } from 'react'

// Semantic color pairs (gradient start -> end) for chart tones. "auto" on
// ProgressRing resolves to one of these based on the value itself, so a
// low agreement percentage actually *reads* as bad (red) rather than
// looking identical to a high one just because both used the brand color.
const TONE_COLORS = {
  success: ['#4ade80', '#16a34a'],
  warning: ['#fbbf24', '#d97706'],
  danger: ['#f87171', '#dc2626'],
  brand: ['#8b6ae8', '#6366f1'],
}

function toneFor(tone, percent) {
  if (tone !== 'auto') return TONE_COLORS[tone] || TONE_COLORS.brand
  if (percent >= 80) return TONE_COLORS.success
  if (percent >= 50) return TONE_COLORS.warning
  return TONE_COLORS.danger
}

/**
 * Simple horizontal bar chart, built from real aggregated data (no
 * charting library — a handful of divs is all this needs). Bars animate
 * in from zero width on mount rather than appearing at full length
 * instantly. `tone` picks the fill color's meaning (e.g. "danger" for a
 * disagreement count, where more should visually read as worse).
 */
export function BarChart({ data, tone = 'brand', emptyText = 'Nothing to show yet.' }) {
  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true))
    return () => cancelAnimationFrame(id)
  }, [])

  if (data.length === 0) {
    return <p className="muted small">{emptyText}</p>
  }

  const max = Math.max(1, ...data.map((d) => d.value))
  const [c1, c2] = TONE_COLORS[tone] || TONE_COLORS.brand

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
              style={{
                width: mounted ? `${(d.value / max) * 100}%` : 0,
                background: `linear-gradient(90deg, ${c1}, ${c2})`,
              }}
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
 * scored" style completion stats. `tone` is "success" | "warning" |
 * "danger" | "brand" (a fixed color) or "auto" (color follows the value:
 * high percent reads as good/green, low as bad/red — see toneFor above).
 * Animates in on mount the same way the bar chart does.
 */
export function ProgressRing({ percent, size = 92, stroke = 10, label, tone = 'brand' }) {
  const [mounted, setMounted] = useState(false)
  const gradientId = useId() // each ring needs its own <linearGradient> id, or every ring on the page would render whichever gradient def happens to come first in the DOM
  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true))
    return () => cancelAnimationFrame(id)
  }, [])

  const r = (size - stroke) / 2
  const c = 2 * Math.PI * r
  const pct = mounted ? percent : 0
  const offset = c - (pct / 100) * c
  const [c1, c2] = toneFor(tone, percent)

  return (
    <div className="progress-ring">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="var(--color-surface-sunken)" strokeWidth={stroke} />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          stroke={`url(#${gradientId})`}
          strokeWidth={stroke}
          strokeDasharray={c}
          strokeDashoffset={offset}
          strokeLinecap="round"
          transform={`rotate(-90 ${size / 2} ${size / 2})`}
          className="progress-ring__fill"
        />
        <defs>
          <linearGradient id={gradientId} x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor={c1} />
            <stop offset="100%" stopColor={c2} />
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
