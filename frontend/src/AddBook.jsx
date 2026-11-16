import { useState } from 'react'
import { createBook } from './api.js'

const EMPTY = {
  title: '', author: '', pageCount: '', isbn: '', releaseDate: '', language: '', coverUrl: '',
}

// Builds the request body from the filled-in fields only.
function toRequest(v) {
  const body = { title: v.title.trim(), author: v.author.trim() }
  if (v.pageCount.trim() !== '') body.page_count = Number(v.pageCount)
  if (v.isbn.trim() !== '') body.isbn = v.isbn.trim()
  // The API expects RFC3339; the date input gives YYYY-MM-DD.
  if (v.releaseDate !== '') body.release_date = `${v.releaseDate}T00:00:00Z`
  if (v.language.trim() !== '') body.language = v.language.trim()
  if (v.coverUrl.trim() !== '') body.cover_url = v.coverUrl.trim()
  return body
}

export function AddBook() {
  const [values, setValues] = useState(EMPTY)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState(null)
  const [added, setAdded] = useState(null)

  const set = (field) => (e) => setValues((v) => ({ ...v, [field]: e.target.value }))

  async function onSubmit(e) {
    e.preventDefault()
    setAdded(null)
    if (values.title.trim() === '' || values.author.trim() === '') {
      setError('Title and author are required.')
      return
    }
    if (values.pageCount.trim() !== '' && !/^\d+$/.test(values.pageCount.trim())) {
      setError('Page count must be a whole number.')
      return
    }
    setError(null)
    setSaving(true)
    try {
      const book = await createBook(toRequest(values))
      setAdded(book?.title ?? values.title.trim())
      setValues(EMPTY)
    } catch (err) {
      setError(err.apiMessage ?? 'Could not add the book. Please try again.')
    } finally {
      setSaving(false)
    }
  }

  return (
    <section>
      <h2>Add a book</h2>
      {error && <p role="alert">{error}</p>}
      {added && <p role="status">Added {added}</p>}
      <form onSubmit={onSubmit} noValidate>
        <fieldset disabled={saving}>
          <label>Title <input value={values.title} onChange={set('title')} required /></label>{' '}
          <label>Author <input value={values.author} onChange={set('author')} required /></label>{' '}
          <label>Page count <input type="number" min="1" value={values.pageCount} onChange={set('pageCount')} /></label>{' '}
          <label>ISBN <input value={values.isbn} onChange={set('isbn')} /></label>{' '}
          <label>Release date <input type="date" value={values.releaseDate} onChange={set('releaseDate')} /></label>{' '}
          <label>Language <input value={values.language} onChange={set('language')} /></label>{' '}
          <label>Cover URL <input value={values.coverUrl} onChange={set('coverUrl')} /></label>{' '}
          <button type="submit">{saving ? 'Saving…' : 'Add book'}</button>
        </fieldset>
      </form>
    </section>
  )
}
