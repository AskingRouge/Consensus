// Brand mark — the actual logo file (public/logo-mark.png), not a
// redrawn approximation. Two hand-traced SVG attempts didn't match
// closely enough, and once the real asset was available there was no
// reason to keep guessing at it.
export default function Logo({ size = 32, wordmark = true, className = '' }) {
  return (
    <span className={`logo ${className}`.trim()}>
      <img
        src="/logo-mark.png"
        width={size}
        height={size * 0.953} /* actual file is 400x381, not square */
        alt=""
        className="logo__mark"
      />
      {wordmark && <span className="logo__wordmark">Consensus</span>}
    </span>
  )
}
