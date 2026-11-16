import { useState } from 'react'

const COVER_W = 48
const COVER_H = 72

// Cover thumbnail with a same-sized placeholder when missing or failing to load.
export function Cover({ title, url }) {
  const [failedUrl, setFailedUrl] = useState(null)
  const box = { width: COVER_W, height: COVER_H, flex: 'none', verticalAlign: 'middle', marginRight: 8 }
  if (!url || failedUrl === url) {
    return (
      <span
        className="book-cover-placeholder"
        data-testid="cover-placeholder"
        aria-hidden="true"
        style={{ ...box, display: 'inline-block', background: '#ddd', borderRadius: 2 }}
      />
    )
  }
  return (
    <img
      src={url}
      alt={`Cover of ${title}`}
      loading="lazy"
      width={COVER_W}
      height={COVER_H}
      style={{ ...box, objectFit: 'cover' }}
      onError={() => setFailedUrl(url)}
    />
  )
}

