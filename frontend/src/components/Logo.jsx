// Brand mark: a squircle badge in the Royal Violet → Indigo gradient with
// a checkmark (the "consensus" idea — agreement, confirmed) plus the
// wordmark. Used in the sidebar and the auth screen; `mark` alone is used
// wherever space is tight.
export default function Logo({ size = 30, wordmark = true, className = '' }) {
  return (
    <span className={`logo ${className}`.trim()}>
      <svg width={size} height={size} viewBox="0 0 40 40" className="logo__mark" aria-hidden="true">
        <defs>
          <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="var(--color-primary)" />
            <stop offset="100%" stopColor="var(--color-accent)" />
          </linearGradient>
        </defs>
        <rect width="40" height="40" rx="12" fill="url(#logoGradient)" />
        <path
          d="M12 21.5l5.5 5.5L28 14"
          stroke="white"
          strokeWidth="3.6"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      {wordmark && <span className="logo__wordmark">Consensus</span>}
    </span>
  )
}
