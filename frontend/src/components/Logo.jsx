// Brand mark, redrawn to match the actual logo: a folded-ribbon "C" (for
// Consensus) — two pointed triangular arms meeting at a deep-plum spine,
// lightest violet at the top tip through to indigo at the bottom tip —
// with a bold checkmark filling the opening, plus the wordmark.
export default function Logo({ size = 32, wordmark = true, className = '' }) {
  return (
    <span className={`logo ${className}`.trim()}>
      <svg width={size} height={size * 1.18} viewBox="0 0 220 260" className="logo__mark" aria-hidden="true">
        <defs>
          <linearGradient id="logoTopArm" x1="100%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#CCB2FB" />
            <stop offset="100%" stopColor="#7C4FE0" />
          </linearGradient>
          <linearGradient id="logoBotArm" x1="0%" y1="100%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#3A2E9A" />
            <stop offset="100%" stopColor="#5F5FE8" />
          </linearGradient>
          <linearGradient id="logoSpine" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#2A0B52" />
            <stop offset="100%" stopColor="#3B0F6F" />
          </linearGradient>
        </defs>
        <path d="M150,18 L26,108 L58,120 Z" fill="url(#logoTopArm)" />
        <path d="M150,242 L26,152 L58,140 Z" fill="url(#logoBotArm)" />
        <path d="M26,108 L58,120 L58,140 L26,152 Z" fill="url(#logoSpine)" />
        <path
          d="M112 138 L142 168 L200 108"
          stroke="#3F2A8C"
          strokeWidth="20"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      {wordmark && <span className="logo__wordmark">Consensus</span>}
    </span>
  )
}
