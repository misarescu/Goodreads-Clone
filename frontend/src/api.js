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
