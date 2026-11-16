import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
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

  describe('progress form', () => {
    const open = async (b) => {
      fetch.mockReturnValueOnce(json(shelves)).mockReturnValueOnce(json([b]))
      render(<Shelves userId="1" />)
      fireEvent.click(await screen.findByText('Read'))
      await screen.findByText('Dune')
    }
    const noPages = { Int64: 0, Valid: false }
    const put = (o) => ({ id: 1, shelf_id: 10, book_id: 7, pages_read: 0, progress_percent: 0, ...o })
    const lastCall = () => fetch.mock.calls[2]

    it('saves pages read and shows the response', async () => {
      await open(book({ book_id: 7 }))
      fetch.mockReturnValueOnce(json(put({ pages_read: 200, progress_percent: 50 })))
      fireEvent.change(screen.getByLabelText('Pages read'), { target: { value: '200' } })
      fireEvent.click(screen.getByText('Save progress'))
      expect(await screen.findByText(/50%/)).toBeTruthy()
      expect(screen.getByText(/50%/).textContent).toContain('200 of 400 pages')
      expect(lastCall()[0]).toBe('/api/shelves/10/books/7/progress')
      expect(lastCall()[1].method).toBe('PUT')
      expect(JSON.parse(lastCall()[1].body)).toEqual({ pages_read: 200 })
      expect(fetch).toHaveBeenCalledTimes(3)
    })

    it('saves a percentage for a book without a page count', async () => {
      await open(book({ book_id: 7, page_count: noPages }))
      fetch.mockReturnValueOnce(json(put({ pages_read: 100, progress_percent: 60 })))
      fireEvent.change(screen.getByLabelText('Percent read'), { target: { value: '60' } })
      fireEvent.click(screen.getByText('Save progress'))
      expect(await screen.findByText(/60%/)).toBeTruthy()
      expect(JSON.parse(lastCall()[1].body)).toEqual({ progress_percent: 60 })
    })

    it('marks a book without a page count as finished', async () => {
      await open(book({ book_id: 7, page_count: noPages }))
      fetch.mockReturnValueOnce(json(put({ progress_percent: 100 })))
      fireEvent.click(screen.getByText('Mark as finished'))
      expect(await screen.findByText(/100%/)).toBeTruthy()
      expect(JSON.parse(lastCall()[1].body)).toEqual({ progress_percent: 100 })
    })

    it('marks a book with a page count as finished by reading all pages', async () => {
      await open(book({ book_id: 7 }))
      fetch.mockReturnValueOnce(json(put({ pages_read: 400, progress_percent: 100 })))
      fireEvent.click(screen.getByText('Mark as finished'))
      expect(await screen.findByText(/400 of 400 pages/)).toBeTruthy()
      expect(JSON.parse(lastCall()[1].body)).toEqual({ pages_read: 400 })
    })

    it('disables the form while saving', async () => {
      await open(book({ book_id: 7 }))
      let resolve
      fetch.mockReturnValueOnce(new Promise((r) => (resolve = r)))
      fireEvent.change(screen.getByLabelText('Pages read'), { target: { value: '5' } })
      fireEvent.click(screen.getByText('Save progress'))
      await waitFor(() => expect(screen.getByText('Save progress').disabled).toBe(true))
      expect(screen.getByLabelText('Pages read').disabled).toBe(true)
      expect(screen.getByText('Mark as finished').disabled).toBe(true)
      resolve(await json(put({ pages_read: 5, progress_percent: 1 })))
      await waitFor(() => expect(screen.getByText('Save progress').disabled).toBe(false))
    })

    it('shows the API message on a 400 and stays usable', async () => {
      await open(book({ book_id: 7 }))
      fetch.mockReturnValueOnce(json({ Error: 'pages_read must be between 0 and 400' }, 400))
      fireEvent.change(screen.getByLabelText('Pages read'), { target: { value: '999' } })
      fireEvent.click(screen.getByText('Save progress'))
      expect((await screen.findByRole('alert')).textContent).toBe('pages_read must be between 0 and 400')
      expect(screen.getByText(/25%/)).toBeTruthy()
      expect(screen.getByText('Save progress').disabled).toBe(false)
      expect(screen.getByLabelText('Pages read').disabled).toBe(false)
    })

    it('shows a generic message for other errors and stays usable', async () => {
      await open(book({ book_id: 7 }))
      fetch.mockReturnValueOnce(json({ Error: 'database is on fire' }, 500))
      fireEvent.click(screen.getByText('Mark as finished'))
      const alert = await screen.findByRole('alert')
      expect(alert.textContent).toBe('Could not save progress. Please try again.')
      expect(screen.getByText('Mark as finished').disabled).toBe(false)
    })
  })
})
