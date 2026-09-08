// Brand mark, redrawn to match the actual logo: an isometric hexagon
// "cube" — three rhombus facets (light violet upper-left, Royal Violet
// upper-right, Deep Plum below, lit from the upper-left) — with a white
// squircle badge and checkmark floating on top, plus the wordmark.
export default function Logo({ size = 32, wordmark = true, className = '' }) {
  return (
    <span className={`logo ${className}`.trim()}>
      <svg width={size} height={size} viewBox="0 0 200 200" className="logo__mark" aria-hidden="true">
        <defs>
          <linearGradient id="logoFaceLeft" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#C7B4FB" />
            <stop offset="100%" stopColor="#9F7EF0" />
          </linearGradient>
          <linearGradient id="logoFaceRight" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#9B70EE" />
            <stop offset="100%" stopColor="#7C3AED" />
          </linearGradient>
          <linearGradient id="logoFaceBottom" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#6428C9" />
            <stop offset="100%" stopColor="#3B0F6F" />
          </linearGradient>
          <filter id="logoBadgeShadow" x="-40%" y="-40%" width="180%" height="180%">
            <feDropShadow dx="0" dy="3" stdDeviation="5" floodColor="#3B0F6F" floodOpacity="0.35" />
          </filter>
        </defs>

        {/* three facets of the isometric hexagon */}
        <path d="M100,100 L100,15 L173.6,57.5 L173.6,142.5 Z" fill="url(#logoFaceRight)" />
        <path d="M100,100 L173.6,142.5 L100,185 L26.4,142.5 Z" fill="url(#logoFaceBottom)" />
        <path d="M100,100 L26.4,142.5 L26.4,57.5 L100,15 Z" fill="url(#logoFaceLeft)" />

        {/* checkmark badge */}
        <g filter="url(#logoBadgeShadow)">
          <rect x="46" y="56" width="102" height="102" rx="27" fill="white" />
        </g>
        <path
          d="M69 109l23 22 47-49"
          stroke="#3B0F6F"
          strokeWidth="15"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      {wordmark && <span className="logo__wordmark">Consensus</span>}
    </span>
  )
}
