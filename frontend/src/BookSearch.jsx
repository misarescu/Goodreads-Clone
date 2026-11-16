import { useEffect, useState } from 'react'
import { fetchBooksByTag, fetchJSON } from './api.js'
import { Cover } from './Cover.jsx'

export const DEBOUNCE_MS = 300
const PAGE_SIZE = 10
const TAG_PAGE_SIZE = 20

// The API serializes a nullable language as {String, Valid}; accept a plain string too.
function languageOf(book) {
  const l = book.language
  if (l && typeof l === 'object') return l.Valid ? l.String : ''
  return l || ''
}

function TagChips({ tags, onSelect }) {
  if (!tags || tags.length === 0) return null
  return (
    <>
      {' '}
      {tags.map((t) => (
        <button
          key={t}
          type="button"
          className="tag-chip"
          style={{ marginRight: 4, padding: '0 8px', borderRadius: 12, fontSize: '0.8em' }}
          onClick={() => onSelect(t)}
        >
          {t}
        </button>
      ))}
    </>
  )
}

// Books carrying one tag, 20 at a time.
function TagBooks({ tag, onBack, onSelectTag }) {
  const [state, setState] = useState({ books: [], status: 'loading', more: false })
  const [offset, setOffset] = useState(0)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    fetchBooksByTag(tag, TAG_PAGE_SIZE, offset, controller.signal)
      .then((page) => {
        const list = Array.isArray(page) ? page : []
        setState((s) => ({
          books: offset === 0 ? list : [...s.books, ...list],
          status: 'ok',
          more: list.length === TAG_PAGE_SIZE,
        }))
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState((s) => ({ ...s, status: 'error', message: err.message }))
        }
      })
    return () => controller.abort()
  }, [tag, offset, attempt])

  // Fetches the next page; after a failed fetch this retries the same page.
  const loadMore = () => {
    setState((s) => ({ ...s, status: 'loading' }))
    if (state.status === 'error') setAttempt((a) => a + 1)
    else setOffset(state.books.length)
  }

  return (
    <section className="book-search">
      <button type="button" onClick={onBack}>
        Back to search results
      </button>
      <h2>Books tagged “{tag}”</h2>
      {state.books.length > 0 && (
        <ul>
          {state.books.map((b) => (
            <li key={b.id}>
              <Cover title={b.title} url={b.cover_url} />
              <strong>{b.title}</strong> by {b.author}
              <TagChips tags={b.tags} onSelect={onSelectTag} />
            </li>
          ))}
        </ul>
      )}
      {state.status === 'loading' && <p>Loading…</p>}
      {state.status === 'error' && <p role="alert">Could not load books: {state.message}</p>}
      {state.status === 'ok' && state.books.length === 0 && <p>No books with this tag.</p>}
      {state.status !== 'loading' && (state.more || (state.status === 'error' && offset > 0)) && (
        <button type="button" onClick={loadMore}>
          Load more
        </button>
      )}
    </section>
  )
}

// "4.2 (12 reviews)", or "No reviews yet" when there are none.
function ratingSummary(book) {
  const n = book.review_count
  if (book.average_rating == null || !n) return 'No reviews yet'
  return `${book.average_rating.toFixed(1)} (${n} ${n === 1 ? 'review' : 'reviews'})`
}

export function BookSearch() {
  const [tag, setTag] = useState(null)
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

  if (tag !== null) {
    return <TagBooks key={tag} tag={tag} onBack={() => setTag(null)} onSelectTag={setTag} />
  }

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
                {languageOf(b) && <> ({languageOf(b)})</>}{' '}
                <span className="book-rating">{ratingSummary(b)}</span>
                <TagChips tags={b.tags} onSelect={setTag} />
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
