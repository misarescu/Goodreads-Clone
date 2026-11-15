import { useEffect, useState } from 'react'
import { fetchShelfBooks, fetchUserShelves } from './api.js'

function ShelfBooks({ shelfId }) {
  const [state, setState] = useState({ shelfId, status: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    fetchShelfBooks(shelfId, controller.signal)
      .then((books) => setState({ shelfId, status: 'ok', books: books ?? [] }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState({ shelfId, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [shelfId])

  // Stale state from a previous shelf counts as loading.
  if (state.shelfId !== shelfId || state.status === 'loading') {
    return <p>Loading books…</p>
  }
  if (state.status === 'error') {
    return <p role="alert">Could not load shelf books: {state.message}</p>
  }
  if (state.books.length === 0) return <p>This shelf has no books yet.</p>
  return (
    <ul>
      {state.books.map((b) => (
        <li key={b.id}>
          <strong>{b.title}</strong> by {b.author}
          <div>
            <progress value={b.progress_percent} max="100" /> {b.progress_percent}%
            {b.page_count?.Valid
              ? ` (${b.pages_read} of ${b.page_count.Int64} pages)`
              : ` (${b.pages_read} pages read)`}
          </div>
        </li>
      ))}
    </ul>
  )
}

export function Shelves({ userId }) {
  const [state, setState] = useState({ userId, status: 'loading' })
  const [selection, setSelection] = useState({ userId, shelfId: null })

  useEffect(() => {
    const controller = new AbortController()
    fetchUserShelves(userId, controller.signal)
      .then((shelves) => setState({ userId, status: 'ok', shelves: shelves ?? [] }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState({ userId, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [userId])

  // Stale state from a previous user counts as loading.
  if (state.userId !== userId || state.status === 'loading') {
    return <p>Loading shelves…</p>
  }
  if (state.status === 'error') {
    return <p role="alert">Could not load shelves: {state.message}</p>
  }
  const selectedId = selection.userId === userId ? selection.shelfId : null
  const selected = state.shelves.find((s) => s.id === selectedId)
  return (
    <section className="shelves">
      <h2>Shelves</h2>
      {state.shelves.length === 0 ? (
        <p>No shelves yet.</p>
      ) : (
        <>
          <ul>
            {state.shelves.map((s) => (
              <li key={s.id}>
                <button
                  type="button"
                  aria-pressed={s.id === selectedId}
                  onClick={() => setSelection({ userId, shelfId: s.id })}
                >
                  {s.name}
                </button>
              </li>
            ))}
          </ul>
          {selected && <ShelfBooks shelfId={selected.id} />}
        </>
      )}
    </section>
  )
}
