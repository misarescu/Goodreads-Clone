export async function fetchJSON(url, signal, options = {}) {
  const res = await fetch(url, { ...options, signal })
  if (!res.ok) {
    const err = new Error(`Request failed (${res.status})`)
    err.status = res.status
    throw err
  }
  if (res.status === 204) return null
  return res.json()
}

export const fetchUserShelves = (userId, signal) =>
  fetchJSON(`/api/users/${userId}/shelves`, signal)

export const fetchShelfBooks = (shelfId, signal) =>
  fetchJSON(`/api/shelves/${shelfId}/books`, signal)

// Sets a book's reading progress. Send pages_read for a book with a page count,
// progress_percent for one without. Failures carry the API's message (the
// server's {"Error": "..."} body) in err.apiMessage for 400 responses.
export async function updateShelfBookProgress(shelfId, bookId, progress, signal) {
  const res = await fetch(`/api/shelves/${shelfId}/books/${bookId}/progress`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(progress),
    signal,
  })
  if (!res.ok) {
    const err = new Error(`Request failed (${res.status})`)
    err.status = res.status
    if (res.status === 400) {
      const body = await res.json().catch(() => null)
      if (typeof body?.Error === 'string') err.apiMessage = body.Error
    }
    throw err
  }
  return res.json()
}

// Sends a JSON shelf mutation. Failures carry the API's message (the server's
// {"Error": "..."} body) in err.apiMessage for 400 and 409 responses.
async function mutateShelf(url, method, body) {
  const res = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  if (!res.ok) {
    const err = new Error(`Request failed (${res.status})`)
    err.status = res.status
    if (res.status === 400 || res.status === 409) {
      const parsed = await res.json().catch(() => null)
      if (typeof parsed?.Error === 'string') err.apiMessage = parsed.Error
    }
    throw err
  }
  return res.status === 204 ? null : res.json()
}

export const createShelf = (userId, name) =>
  mutateShelf('/api/shelves', 'POST', { user_id: Number(userId), name })

export const renameShelf = (shelfId, name) =>
  mutateShelf(`/api/shelves/${shelfId}`, 'PUT', { name })

export const deleteShelf = (shelfId) => mutateShelf(`/api/shelves/${shelfId}`, 'DELETE')

// Creates a book. Failures carry the API's message (the server's {"Error": "..."}
// body) in err.apiMessage for 400 responses.
export async function createBook(book, signal) {
  const res = await fetch('/api/books', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(book),
    signal,
  })
  if (!res.ok) {
    const err = new Error(`Request failed (${res.status})`)
    err.status = res.status
    if (res.status === 400) {
      const body = await res.json().catch(() => null)
      if (typeof body?.Error === 'string') err.apiMessage = body.Error
    }
    throw err
  }
  return res.json()
}
