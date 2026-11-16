import { useEffect, useState } from 'react'
import { fetchJSON } from './api.js'

export const DEBOUNCE_MS = 300
const PAGE_SIZE = 10

// The API serializes a nullable language as {String, Valid}; accept a plain string too.
function languageOf(book) {
  const l = book.language
  if (l && typeof l === 'object') return l.Valid ? l.String : ''
  return l || ''
}

const COVER_W = 48
const COVER_H = 72

// Cover thumbnail with a same-sized placeholder when missing or failing to load.
function Cover({ title, url }) {
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

export function BookSearch() {
  const [input, setInput] = useState('')
  const [search, setSearch] = useState({ q: '', offset: 0 })
  const [result, setResult] = useState(null)

  // Debounce typing; a new query starts again from the first page.
  useEffect(() => {
    const q = input.trim()
    const timer = setTimeout(() => setSearch({ q, offset: 0 }), DEBOUNCE_MS)
    return () => clearTimeout(timer)
  }, [input])

  const { q, offset } = search
  useEffect(() => {
    if (!q) return
    const controller = new AbortController()
    const params = new URLSearchParams({ q, limit: PAGE_SIZE, offset })
    fetchJSON(`/api/books/search?${params}`, controller.signal)
      .then((data) => setResult({ q, offset, status: 'ok', data }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setResult({ q, offset, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [q, offset])

  const current = q && result && result.q === q && result.offset === offset ? result : null
  const data = current?.status === 'ok' ? current.data : null
  const total = data ? data.total : 0
  const books = data?.books ?? []

  return (
    <section className="book-search">
      <h2>Search books</h2>
      <input
        type="search"
        aria-label="Search books"
        placeholder="Title, author…"
        value={input}
        onChange={(e) => setInput(e.target.value)}
      />
      <p className="hint">
        Tip: narrow results with <code>author:</code> and <code>lang:</code>, e.g.{' '}
        <code>dune author:herbert lang:en</code>.
      </p>
      {q && !current && <p>Searching…</p>}
      {current?.status === 'error' && (
        <p role="alert">Could not search books: {current.message}</p>
      )}
      {data && total === 0 && <p>No books found.</p>}
      {data && total > 0 && (
        <>
          <p>
            {total} {total === 1 ? 'result' : 'results'}
          </p>
          <ul>
            {books.map((b) => (
              <li key={b.id}>
                <Cover title={b.title} url={b.cover_url} />
                <strong>{b.title}</strong> by {b.author}
                {languageOf(b) && <> ({languageOf(b)})</>}
              </li>
            ))}
          </ul>
          <button
            type="button"
            disabled={offset === 0}
            onClick={() => setSearch({ q, offset: Math.max(0, offset - PAGE_SIZE) })}
          >
            Previous
          </button>{' '}
          <button
            type="button"
            disabled={offset + PAGE_SIZE >= total}
            onClick={() => setSearch({ q, offset: offset + PAGE_SIZE })}
          >
            Next
          </button>
        </>
      )}
    </section>
  )
}
