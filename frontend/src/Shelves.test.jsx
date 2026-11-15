import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Shelves } from './Shelves.jsx'

const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
const shelves = [
  { id: 10, name: 'Read' },
  { id: 11, name: 'Want to read' },
]
const book = (o = {}) => ({
  id: 1, title: 'Dune', author: 'Herbert', pages_read: 100, progress_percent: 25,
  page_count: { Int64: 400, Valid: true }, ...o,
})

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('Shelves', () => {
  it('shows loading while pending', () => {
    fetch.mockReturnValue(new Promise(() => {}))
    render(<Shelves userId="1" />)
    expect(screen.getByText('Loading shelves…')).toBeTruthy()
  })

  it('lists shelves', async () => {
    fetch.mockReturnValue(json(shelves))
    render(<Shelves userId="1" />)
    expect(await screen.findByText('Want to read')).toBeTruthy()
    expect(screen.getByText('Read')).toBeTruthy()
    expect(fetch.mock.calls[0][0]).toBe('/api/users/1/shelves')
    expect(fetch).toHaveBeenCalledTimes(1)
  })

  it('shows books with progress when a shelf is selected', async () => {
    fetch.mockReturnValueOnce(json(shelves)).mockReturnValueOnce(json([book()]))
    render(<Shelves userId="1" />)
    fireEvent.click(await screen.findByText('Read'))
    expect(await screen.findByText('Dune')).toBeTruthy()
    expect(fetch.mock.calls[1][0]).toBe('/api/shelves/10/books')
    const bar = document.querySelector('progress')
    expect(bar.value).toBe(25)
    expect(bar.max).toBe(100)
    expect(screen.getByText(/25%/).textContent).toContain('100 of 400 pages')
  })

  it('handles unknown page counts', async () => {
    fetch
      .mockReturnValueOnce(json(shelves))
      .mockReturnValueOnce(json([book({ page_count: { Int64: 0, Valid: false } })]))
    render(<Shelves userId="1" />)
    fireEvent.click(await screen.findByText('Read'))
    expect(await screen.findByText(/100 pages read/)).toBeTruthy()
  })

  it('shows an empty state for no shelves', async () => {
    fetch.mockReturnValue(json([]))
    render(<Shelves userId="1" />)
    expect(await screen.findByText('No shelves yet.')).toBeTruthy()
  })

  it('shows an empty state for a shelf with no books', async () => {
    fetch.mockReturnValueOnce(json(shelves)).mockReturnValueOnce(json([]))
    render(<Shelves userId="1" />)
    fireEvent.click(await screen.findByText('Read'))
    expect(await screen.findByText('This shelf has no books yet.')).toBeTruthy()
  })

  it('shows an alert when shelves fail to load', async () => {
    fetch.mockReturnValue(json({}, 500))
    render(<Shelves userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain(
      'Could not load shelves',
    )
  })

  it('shows an alert when books fail to load, keeping the shelf list', async () => {
    fetch.mockReturnValueOnce(json(shelves)).mockRejectedValueOnce(new Error('boom'))
    render(<Shelves userId="1" />)
    fireEvent.click(await screen.findByText('Read'))
    expect((await screen.findByRole('alert')).textContent).toContain('boom')
    expect(screen.getByText('Want to read')).toBeTruthy()
  })

  it('switches shelves', async () => {
    fetch
      .mockReturnValueOnce(json(shelves))
      .mockReturnValueOnce(json([book()]))
      .mockReturnValueOnce(new Promise(() => {}))
    render(<Shelves userId="1" />)
    fireEvent.click(await screen.findByText('Read'))
    await screen.findByText('Dune')
    fireEvent.click(screen.getByText('Want to read'))
    expect(screen.getByText('Loading books…')).toBeTruthy()
    expect(screen.queryByText('Dune')).toBeNull()
  })

  it('resets when the user changes', async () => {
    fetch.mockReturnValueOnce(json(shelves)).mockReturnValueOnce(new Promise(() => {}))
    const { rerender } = render(<Shelves userId="1" />)
    await screen.findByText('Read')
    rerender(<Shelves userId="2" />)
    expect(screen.getByText('Loading shelves…')).toBeTruthy()
    expect(fetch.mock.calls[1][0]).toBe('/api/users/2/shelves')
  })
})
