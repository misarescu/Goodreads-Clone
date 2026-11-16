import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from './App.jsx'

const year = new Date().getFullYear()
const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
const months = (overrides = {}) =>
  Array.from({ length: 12 }, (_, i) => ({
    month: i + 1, books_finished: 0, pages_finished: 0, ...overrides[i + 1],
  }))

function route(monthly) {
  fetch.mockImplementation((url) => {
    if (url === '/api/users') return json([{ id: 1, username: 'alice' }])
    if (url === '/api/users/1/stats') {
      return json({ books_on_shelves: 1, pages_read: 2, books_finished: 3 })
    }
    if (url.startsWith('/api/users/1/stats/monthly')) return monthly(url)
    return new Promise(() => {})
  })
}

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('monthly reading stats', () => {
  it('shows a twelve-row table for the current year', async () => {
    route(() => json(months({ 2: { books_finished: 3, pages_finished: 150 } })))
    render(<App />)
    const panel = await screen.findByRole('region', { name: 'Monthly reading stats' })
    const feb = await within(panel).findByText('February')
    expect(fetch.mock.calls.map((c) => c[0])).toContain(
      `/api/users/1/stats/monthly?year=${year}`,
    )
    const rows = within(panel).getAllByRole('row')
    expect(rows).toHaveLength(13)
    expect(rows[0].textContent).toBe('MonthBooksPages')
    expect(rows[1].textContent).toBe('January00')
    expect(feb.closest('tr').textContent).toBe('February3150')
    expect(rows[12].textContent).toBe('December00')
  })

  it('reloads when the year changes', async () => {
    route((url) =>
      json(url.endsWith(`year=${year - 1}`)
        ? months({ 12: { books_finished: 1, pages_finished: 99 } })
        : months()))
    render(<App />)
    const select = await screen.findByLabelText('Year')
    expect(select.value).toBe(String(year))
    await screen.findByText('January')
    fireEvent.change(select, { target: { value: String(year - 1) } })
    await waitFor(() =>
      expect(screen.getByText('December').closest('tr').textContent).toBe('December199'))
    expect(fetch.mock.calls.map((c) => c[0])).toContain(
      `/api/users/1/stats/monthly?year=${year - 1}`,
    )
  })

  it('shows an error when the request fails', async () => {
    route(() => json({}, 500))
    render(<App />)
    expect(
      await screen.findByText('Could not load monthly stats: Request failed (500)'),
    ).toBeTruthy()
  })
})
