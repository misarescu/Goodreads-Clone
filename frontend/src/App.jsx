import { useEffect, useState } from 'react'
import './App.css'
import { fetchJSON } from './api.js'
import { ReadingChallenge } from './ReadingChallenge.jsx'
import { BookReviews } from './BookReviews.jsx'
import { BookSearch } from './BookSearch.jsx'
import { AddBook } from './AddBook.jsx'
import { Recommendations } from './Recommendations.jsx'
import { Shelves } from './Shelves.jsx'


const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

// Years offered by the selector: next year back to nine years ago.
function yearOptions() {
  const next = new Date().getFullYear() + 1
  return Array.from({ length: 11 }, (_, i) => next - i)
}

function MonthlyStats({ userId }) {
  const [year, setYear] = useState(new Date().getFullYear())
  const [state, setState] = useState({ key: '', status: 'loading' })
  const key = `${userId}:${year}`

  useEffect(() => {
    const controller = new AbortController()
    fetchJSON(`/api/users/${userId}/stats/monthly?year=${year}`, controller.signal)
      .then((months) => setState({ key, status: 'ok', months }))
      .catch((err) => {
        if (err.name !== 'AbortError') {
          setState({ key, status: 'error', message: err.message })
        }
      })
    return () => controller.abort()
  }, [userId, year, key])

  return (
    <section aria-label="Monthly reading stats">
      <label>
        Year{' '}
        <select value={year} onChange={(e) => setYear(Number(e.target.value))}>
          {yearOptions().map((y) => (
            <option key={y} value={y}>
              {y}
            </option>
          ))}
        </select>
      </label>
      {state.key !== key || state.status === 'loading' ? (
        <p>Loading monthly stats…</p>
      ) : state.status === 'error' ? (
        <p role="alert">Could not load monthly stats: {state.message}</p>
      ) : (
        <table className="monthly-stats">
          <thead>
            <tr>
              <th>Month</th>
              <th>Books</th>
              <th>Pages</th>
            </tr>
          </thead>
          <tbody>
            {state.months.map((m) => (
              <tr key={m.month}>
                <td>{MONTHS[m.month - 1]}</td>
                <td>{m.books_finished}</td>
                <td>{m.pages_finished}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  )
}

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
    <>
      <dl className="reading-stats">
        <dt>Books on shelves</dt>
        <dd>{stats.books_on_shelves}</dd>
        <dt>Pages read</dt>
        <dd>{stats.pages_read}</dd>
        <dt>Books finished</dt>
        <dd>{stats.books_finished}</dd>
      </dl>
      <MonthlyStats userId={userId} />
    </>
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
      <AddBook />
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
          {selected && <Shelves userId={selected} />}
          {selected && <BookReviews userId={selected} />}
          {selected && <Recommendations userId={selected} />}
        </>
      )}
    </>
  )
}

export default App
