import { useEffect, useState } from 'react'
import { fetchJSON } from './api.js'

const PAGE_SIZE = 20
const JSON_HEADERS = { 'Content-Type': 'application/json' }

function ReviewForm({ bookId, userId, existing, onChanged }) {
  const [rating, setRating] = useState(existing ? String(existing.rating) : '5')
  const [body, setBody] = useState(existing ? existing.body : '')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  function run(url, options) {
    setBusy(true)
    setError('')
    return fetchJSON(url, undefined, options)
      .then(onChanged)
      .catch((err) => setError(err.message))
      .finally(() => setBusy(false))
  }

  function submit(e) {
    e.preventDefault()
    run(`/api/books/${bookId}/reviews`, {
      method: 'POST',
      headers: JSON_HEADERS,
      body: JSON.stringify({ user_id: Number(userId), rating: Number(rating), body }),
    })
  }

  function remove() {
    run(`/api/books/${bookId}/reviews/${userId}`, { method: 'DELETE' })
  }

  return (
    <form onSubmit={submit}>
      <label>
        Rating{' '}
        <select value={rating} onChange={(e) => setRating(e.target.value)}>
          {[1, 2, 3, 4, 5].map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
      </label>{' '}
      <label>
        Review{' '}
        <textarea
          required
          maxLength={5000}
          value={body}
          onChange={(e) => setBody(e.target.value)}
        />
      </label>{' '}
      <button type="submit" disabled={busy}>
        {busy ? 'Saving…' : 'Save review'}
      </button>{' '}
      <button type="button" disabled={busy} onClick={remove}>
        Delete my review
      </button>
      {error && <p role="alert">Could not save review: {error}</p>}
    </form>
  )
}

function ReviewList({ bookId, userId }) {
  const [version, setVersion] = useState(0)
  const [state, setState] = useState({ key: '', status: 'loading' })
  const [extra, setExtra] = useState({ key: '', items: [], done: false, error: '' })
  const key = `${bookId}:${version}`

  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      fetchJSON(`/api/books/${bookId}`, controller.signal),
      fetchJSON(`/api/books/${bookId}/reviews?limit=${PAGE_SIZE}&offset=0`, controller.signal),
    ])
      .then(([book, reviews]) => setState({ key, status: 'ok', book, reviews }))
      .catch((err) => {
        if (err.name !== 'AbortError') setState({ key, status: 'error', message: err.message })
      })
    return () => controller.abort()
  }, [bookId, key])

  // Stale state from a previous book or version counts as loading.
  if (state.key !== key || state.status === 'loading') return <p>Loading reviews…</p>
  if (state.status === 'error') {
    return <p role="alert">Could not load reviews: {state.message}</p>
  }

  const more = extra.key === key ? extra : { items: [], done: false, error: '' }
  const reviews = [...state.reviews, ...more.items]
  const { book } = state
  const canLoadMore = !more.done && reviews.length > 0 && reviews.length % PAGE_SIZE === 0
  const mine = reviews.find((r) => String(r.user_id) === String(userId))

  function loadMore() {
    fetchJSON(`/api/books/${bookId}/reviews?limit=${PAGE_SIZE}&offset=${reviews.length}`)
      .then((page) =>
        setExtra({
          key,
          items: [...more.items, ...page],
          done: page.length < PAGE_SIZE,
          error: '',
        }),
      )
      .catch((err) => setExtra({ ...more, key, error: err.message }))
  }

  return (
    <>
      <p>
        Average rating:{' '}
        {book.average_rating == null ? 'No ratings yet' : book.average_rating.toFixed(1)}
        {' · '}
        {book.review_count} {book.review_count === 1 ? 'review' : 'reviews'}
      </p>
      {reviews.length === 0 ? (
        <p>No reviews yet.</p>
      ) : (
        <ul>
          {reviews.map((r) => (
            <li key={r.id}>
              <strong>{r.username}</strong> — {r.rating}/5
              <p>{r.body}</p>
            </li>
          ))}
        </ul>
      )}
      {more.error && <p role="alert">Could not load more reviews: {more.error}</p>}
      {canLoadMore && (
        <button type="button" onClick={loadMore}>
          Load more
        </button>
      )}
      <ReviewForm
        key={`${key}:${userId}:${mine ? mine.id : ''}`}
        bookId={bookId}
        userId={userId}
        existing={mine}
        onChanged={() => setVersion((v) => v + 1)}
      />
    </>
  )
}

export function BookReviews({ userId }) {
  const [books, setBooks] = useState({ status: 'loading', list: [] })
  const [bookId, setBookId] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    // Simple select: only the first 100 books (server cap) are selectable.
    fetchJSON('/api/books?limit=100', controller.signal)
      .then((list) => {
        setBooks({ status: 'ok', list })
        if (list.length > 0) setBookId(String(list[0].id))
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setBooks({ status: 'error', list: [], message: err.message })
        }
      })
    return () => controller.abort()
  }, [])

  return (
    <section className="book-reviews">
      <h2>Reviews</h2>
      {books.status === 'loading' && <p>Loading books…</p>}
      {books.status === 'error' && (
        <p role="alert">Could not load books: {books.message}</p>
      )}
      {books.status === 'ok' && books.list.length === 0 && <p>No books yet.</p>}
      {books.list.length > 0 && (
        <>
          <label>
            Book{' '}
            <select value={bookId} onChange={(e) => setBookId(e.target.value)}>
              {books.list.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.title}
                </option>
              ))}
            </select>
          </label>
          {bookId && <ReviewList bookId={bookId} userId={userId} />}
        </>
      )}
    </section>
  )
}
