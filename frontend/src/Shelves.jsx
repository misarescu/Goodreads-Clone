import { useEffect, useState } from 'react'
import {
  createShelf,
  deleteShelf,
  fetchShelfBooks,
  fetchUserShelves,
  renameShelf,
  updateShelfBookProgress,
} from './api.js'

function BookProgress({ shelfId, book, onSaved }) {
  const hasPages = Boolean(book.page_count?.Valid) && book.page_count.Int64 > 0
  const [value, setValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)

  const save = async (progress) => {
    setSaving(true)
    setError(null)
    try {
      const updated = await updateShelfBookProgress(shelfId, book.book_id, progress)
      onSaved(updated)
      setValue('')
    } catch (err) {
      setError(err.apiMessage ?? 'Could not save progress. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  const field = hasPages ? 'pages_read' : 'progress_percent'
  const label = hasPages ? 'Pages read' : 'Percent read'
  const submit = (e) => {
    e.preventDefault()
    if (value.trim() === '') return
    save({ [field]: Number(value) })
  }
  // A book with a page count takes pages, so finishing means reading them all.
  const finish = () =>
    save(hasPages ? { pages_read: book.page_count.Int64 } : { progress_percent: 100 })

  return (
    <form onSubmit={submit} noValidate>
      <label>
        {label}{' '}
        <input
          type="number"
          min="0"
          max={hasPages ? book.page_count.Int64 : 100}
          value={value}
          disabled={saving}
          onChange={(e) => setValue(e.target.value)}
        />
      </label>{' '}
      <button type="submit" disabled={saving}>Save progress</button>{' '}
      <button type="button" disabled={saving} onClick={finish}>Mark as finished</button>
      {error && <p role="alert">{error}</p>}
    </form>
  )
}

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
          <BookProgress
            shelfId={shelfId}
            book={b}
            onSaved={(updated) =>
              setState((cur) =>
                cur.shelfId === shelfId && cur.status === 'ok'
                  ? {
                      ...cur,
                      books: cur.books.map((x) =>
                        x.id === b.id
                          ? {
                              ...x,
                              pages_read: updated.pages_read,
                              progress_percent: updated.progress_percent,
                              finished_at: updated.finished_at,
                            }
                          : x,
                      ),
                    }
                  : cur,
              )
            }
          />
        </li>
      ))}
    </ul>
  )
}

function CreateShelfForm({ userId, onCreated }) {
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)

  const submit = async (e) => {
    e.preventDefault()
    const trimmed = name.trim()
    if (trimmed === '') return
    setBusy(true)
    setError(null)
    try {
      onCreated(await createShelf(userId, trimmed))
      setName('')
    } catch (err) {
      setError(err.apiMessage ?? 'Could not create the shelf. Please try again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <form onSubmit={submit} noValidate>
      <label>
        New shelf{' '}
        <input value={name} disabled={busy} onChange={(e) => setName(e.target.value)} />
      </label>{' '}
      <button type="submit" disabled={busy}>Create shelf</button>
      {error && <p role="alert">{error}</p>}
    </form>
  )
}

function ShelfActions({ shelf, onRenamed, onDeleted }) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(shelf.name)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState(null)

  const run = async (fn, fallback) => {
    setBusy(true)
    setError(null)
    try {
      await fn()
    } catch (err) {
      setError(err.apiMessage ?? fallback)
    } finally {
      setBusy(false)
    }
  }

  const save = (e) => {
    e.preventDefault()
    const trimmed = name.trim()
    if (trimmed === '') return
    run(async () => {
      onRenamed(await renameShelf(shelf.id, trimmed))
      setEditing(false)
    }, 'Could not rename the shelf. Please try again.')
  }

  const remove = () => {
    if (!window.confirm(`Delete the shelf "${shelf.name}"? Its books will leave this shelf.`)) {
      return
    }
    run(async () => {
      await deleteShelf(shelf.id)
      onDeleted(shelf.id)
    }, 'Could not delete the shelf. Please try again.')
  }

  return (
    <>
      {editing ? (
        <form onSubmit={save} noValidate style={{ display: 'inline' }}>
          <input
            aria-label={`New name for ${shelf.name}`}
            value={name}
            disabled={busy}
            onChange={(e) => setName(e.target.value)}
          />{' '}
          <button type="submit" disabled={busy}>Save name</button>{' '}
          <button
            type="button"
            disabled={busy}
            onClick={() => {
              setEditing(false)
              setName(shelf.name)
              setError(null)
            }}
          >
            Cancel
          </button>
        </form>
      ) : (
        <>
          <button
            type="button"
            disabled={busy}
            onClick={() => {
              setName(shelf.name)
              setEditing(true)
            }}
          >
            Rename
          </button>{' '}
          <button type="button" disabled={busy} onClick={remove}>Delete</button>
        </>
      )}
      {error && <p role="alert">{error}</p>}
    </>
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
      <CreateShelfForm
        userId={userId}
        onCreated={(shelf) =>
          setState((cur) =>
            cur.userId === userId && cur.status === 'ok'
              ? { ...cur, shelves: [...cur.shelves, shelf] }
              : cur,
          )
        }
      />
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
                </button>{' '}
                {!s.is_default && (
                  <ShelfActions
                    shelf={s}
                    onRenamed={(updated) =>
                      setState((cur) =>
                        cur.status === 'ok'
                          ? {
                              ...cur,
                              shelves: cur.shelves.map((x) =>
                                x.id === s.id ? { ...x, ...updated } : x,
                              ),
                            }
                          : cur,
                      )
                    }
                    onDeleted={(id) =>
                      setState((cur) =>
                        cur.status === 'ok'
                          ? { ...cur, shelves: cur.shelves.filter((x) => x.id !== id) }
                          : cur,
                      )
                    }
                  />
                )}
              </li>
            ))}
          </ul>
          {selected && <ShelfBooks shelfId={selected.id} />}
        </>
      )}
    </section>
  )
}
