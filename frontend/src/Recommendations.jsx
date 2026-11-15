import { useEffect, useState } from 'react'
import { fetchJSON } from './api.js'

export function Recommendations({ userId }) {
  const [state, setState] = useState({ userId, status: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    fetchJSON(`/api/users/${userId}/recommendations`, controller.signal)
      .then((books) => setState({ userId, status: 'ok', books }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState({ userId, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [userId])

  // Stale state from a previous user counts as loading.
  if (state.userId !== userId || state.status === 'loading') {
    return <p>Loading recommendations…</p>
  }
  if (state.status === 'error') {
    return <p role="alert">Could not load recommendations: {state.message}</p>
  }
  return (
    <section className="recommendations">
      <h2>You might like</h2>
      {state.books.length === 0 ? (
        <p>No recommendations yet. Finish some books to get suggestions.</p>
      ) : (
        <ul>
          {state.books.map((b) => (
            <li key={b.id}>
              <strong>{b.title}</strong> by {b.author}
              {b.because?.length > 0 && (
                <p>Because readers who finished {b.because.join(', ')} also finished this</p>
              )}
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
