import { useEffect, useState } from 'react'
import { fetchJSON } from './api.js'

function plural(n) {
  return `${n} ${n === 1 ? 'book' : 'books'}`
}

function schedule(aheadBy) {
  if (aheadBy > 0) return `${plural(aheadBy)} ahead of schedule`
  if (aheadBy < 0) return `${plural(-aheadBy)} behind schedule`
  return 'On schedule'
}

function GoalForm({ userId, year, onSaved }) {
  const [goal, setGoal] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  function submit(e) {
    e.preventDefault()
    setSaving(true)
    setError('')
    fetchJSON(`/api/users/${userId}/challenges/${year}`, undefined, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ goal: Number(goal) }),
    })
      .then(onSaved)
      .catch((err) => {
        setError(err.message)
        setSaving(false)
      })
  }

  return (
    <form onSubmit={submit}>
      <label>
        Books to read in {year}{' '}
        <input
          type="number"
          min="1"
          max="1000"
          required
          value={goal}
          onChange={(e) => setGoal(e.target.value)}
        />
      </label>{' '}
      <button type="submit" disabled={saving}>
        {saving ? 'Saving…' : 'Set goal'}
      </button>
      {error && <p role="alert">Could not save reading challenge: {error}</p>}
    </form>
  )
}

export function ReadingChallenge({ userId }) {
  const year = new Date().getFullYear()
  const [state, setState] = useState({ userId, status: 'loading' })
  const [version, setVersion] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    fetchJSON(`/api/users/${userId}/challenges/${year}`, controller.signal)
      .then((challenge) => setState({ userId, status: 'ok', challenge }))
      .catch((err) => {
        if (err.name === 'AbortError') return
        if (err.status === 404) setState({ userId, status: 'none' })
        else setState({ userId, status: 'error', message: err.message })
      })
    return () => controller.abort()
  }, [userId, year, version])

  // Stale state from a previous user counts as loading.
  if (state.userId !== userId || state.status === 'loading') {
    return <p>Loading challenge…</p>
  }
  if (state.status === 'error') {
    return <p role="alert">Could not load reading challenge: {state.message}</p>
  }
  if (state.status === 'none') {
    return (
      <section className="reading-challenge">
        <h2>{year} reading challenge</h2>
        <GoalForm
          userId={userId}
          year={year}
          onSaved={() => {
            setState({ userId, status: 'loading' })
            setVersion((v) => v + 1)
          }}
        />
      </section>
    )
  }
  const { goal, finished, ahead_by: aheadBy } = state.challenge
  return (
    <section className="reading-challenge">
      <h2>{year} reading challenge</h2>
      <progress value={Math.min(finished, goal)} max={goal} />
      <p>
        {finished} of {goal} books
      </p>
      <p>{schedule(aheadBy)}</p>
      <p>
        {plural(finished)} finished in {year}
      </p>
    </section>
  )
}
