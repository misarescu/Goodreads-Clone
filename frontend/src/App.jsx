import { useEffect, useState } from 'react'
import './App.css'
import { fetchJSON } from './api.js'
import { ReadingChallenge } from './ReadingChallenge.jsx'
import { BookSearch } from './BookSearch.jsx'


function UserStats({ userId }) {
  const [state, setState] = useState({ userId, status: 'loading' })

  useEffect(() => {
    const controller = new AbortController()
    fetchJSON(`/api/users/${userId}/stats`, controller.signal)
      .then((stats) => setState({ userId, status: 'ok', stats }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState({ userId, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [userId])

  // Stale state from a previous user counts as loading.
  if (state.userId !== userId || state.status === 'loading') {
    return <p>Loading stats…</p>
  }
  if (state.status === 'error') {
    return <p role="alert">Could not load reading stats: {state.message}</p>
  }
  const { stats } = state
  return (
    <dl className="reading-stats">
      <dt>Books on shelves</dt>
      <dd>{stats.books_on_shelves}</dd>
      <dt>Pages read</dt>
      <dd>{stats.pages_read}</dd>
      <dt>Books finished</dt>
      <dd>{stats.books_finished}</dd>
    </dl>
  )
}

function App() {
  const [users, setUsers] = useState({ status: 'loading', list: [] })
  const [selected, setSelected] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    fetchJSON('/api/users', controller.signal)
      .then((list) => {
        setUsers({ status: 'ok', list })
        if (list.length > 0) setSelected(String(list[0].id))
      })
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setUsers({ status: 'error', list: [], message: err.message })
        }
      })
    return () => controller.abort()
  }, [])

  return (
    <>
      <BookSearch />
      <h1>Reading stats</h1>
      {users.status === 'loading' && <p>Loading users…</p>}
      {users.status === 'error' && (
        <p role="alert">Could not load users: {users.message}</p>
      )}
      {users.status === 'ok' && users.list.length === 0 && <p>No users yet.</p>}
      {users.list.length > 0 && (
        <>
          <label>
            User{' '}
            <select value={selected} onChange={(e) => setSelected(e.target.value)}>
              {users.list.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.username}
                </option>
              ))}
            </select>
          </label>
          {selected && <UserStats userId={selected} />}
          {selected && <ReadingChallenge userId={selected} />}
        </>
      )}
    </>
  )
}

export default App
