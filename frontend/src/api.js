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
