import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { BookSearch, DEBOUNCE_MS } from './BookSearch.jsx'

const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
const book = (id, o = {}) => ({
  id, title: `Book ${id}`, author: `Author ${id}`, language: { String: 'en', Valid: true }, ...o,
})

beforeEach(() => {
  vi.useFakeTimers()
  vi.stubGlobal('fetch', vi.fn())
})
afterEach(() => {
  cleanup()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

async function type(value) {
  fireEvent.change(screen.getByLabelText('Search books'), { target: { value } })
  await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MS))
}

describe('BookSearch', () => {
  it('shows the filter hint and does not fetch when empty', async () => {
    render(<BookSearch />)
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MS * 2))
    expect(screen.getAllByText(/author:/)[0]).toBeTruthy()
    expect(screen.getAllByText(/lang:/)[0]).toBeTruthy()
    expect(fetch).not.toHaveBeenCalled()
  })

  it('debounces typing into a single request and shows results with total', async () => {
    fetch.mockReturnValue(json({ books: [book(1), book(2, { language: null })], total: 2 }))
    render(<BookSearch />)
    fireEvent.change(screen.getByLabelText('Search books'), { target: { value: 'du' } })
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MS - 1))
    fireEvent.change(screen.getByLabelText('Search books'), { target: { value: 'dune' } })
    await act(() => vi.advanceTimersByTimeAsync(DEBOUNCE_MS - 1))
    expect(fetch).not.toHaveBeenCalled()
    await act(() => vi.advanceTimersByTimeAsync(1))
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(fetch.mock.calls[0][0]).toBe('/api/books/search?q=dune&limit=10&offset=0')
    expect(screen.getByText('2 results')).toBeTruthy()
    const items = screen.getAllByRole('listitem')
    expect(items[0].textContent).toBe('Book 1 by Author 1 (en)')
    expect(items[1].textContent).toBe('Book 2 by Author 2')
  })

  it('pages with previous and next', async () => {
    fetch.mockImplementation((url) => {
      const offset = new URL(url, 'http://x').searchParams.get('offset')
      return json({ books: [book(Number(offset) + 1)], total: 15 })
    })
    render(<BookSearch />)
    await type('a')
    expect(screen.getByRole('listitem').textContent).toBe('Book 1 by Author 1 (en)')
    expect(screen.getByText('Previous').disabled).toBe(true)
    fireEvent.click(screen.getByText('Next'))
    await act(() => vi.advanceTimersByTimeAsync(0))
    expect(fetch.mock.calls[1][0]).toBe('/api/books/search?q=a&limit=10&offset=10')
    expect(screen.getByRole('listitem').textContent).toBe('Book 11 by Author 11 (en)')
    expect(screen.getByText('Next').disabled).toBe(true)
    fireEvent.click(screen.getByText('Previous'))
    await act(() => vi.advanceTimersByTimeAsync(0))
    expect(fetch.mock.calls[2][0]).toBe('/api/books/search?q=a&limit=10&offset=0')
    expect(screen.getByRole('listitem').textContent).toBe('Book 1 by Author 1 (en)')
  })

  it('shows a message when nothing matches', async () => {
    fetch.mockReturnValue(json({ books: [], total: 0 }))
    render(<BookSearch />)
    await type('zzz')
    expect(screen.getByText('No books found.')).toBeTruthy()
    expect(screen.queryByText('Next')).toBeNull()
  })

  it('shows an alert on API error', async () => {
    fetch.mockReturnValue(json({ error: 'boom' }, 500))
    render(<BookSearch />)
    await type('dune')
    expect(screen.getByRole('alert').textContent).toContain('Request failed (500)')
  })

  it('shows covers, a placeholder without one, and a placeholder after load failure', async () => {
    fetch.mockReturnValue(json({
      books: [
        book(1, { cover_url: 'http://x/1.jpg' }),
        book(2, { cover_url: null }),
        book(3, { cover_url: 'http://x/3.jpg' }),
      ],
      total: 3,
    }))
    render(<BookSearch />)
    await type('a')
    const img = screen.getByAltText('Cover of Book 1')
    expect(img.getAttribute('src')).toBe('http://x/1.jpg')
    expect(img.getAttribute('loading')).toBe('lazy')
    expect(img.getAttribute('width')).toBe('48')
    expect(img.getAttribute('height')).toBe('72')
    expect(img.style.objectFit).toBe('cover')
    expect(screen.getAllByTestId('cover-placeholder')).toHaveLength(1)
    fireEvent.error(screen.getByAltText('Cover of Book 3'))
    expect(screen.queryByAltText('Cover of Book 3')).toBeNull()
    expect(screen.getAllByTestId('cover-placeholder')).toHaveLength(2)
    expect(screen.getByAltText('Cover of Book 1')).toBeTruthy()
  })
})

describe('BookSearch tags', () => {
  const tagged = (id, tags) => book(id, { tags })
  const searchWith = async (books) => {
    fetch.mockReturnValue(json({ books, total: books.length }))
    render(<BookSearch />)
    await type('x')
  }

  it('shows tag chips on results and none for an untagged book', async () => {
    await searchWith([tagged(1, ['scifi', 'classic']), tagged(2, []), book(3)])
    const items = screen.getAllByRole('listitem')
    expect(items[0].querySelectorAll('button.tag-chip')).toHaveLength(2)
    expect(items[1].querySelectorAll('button')).toHaveLength(0)
    expect(items[2].querySelectorAll('button')).toHaveLength(0)
  })

  it('lists books with a tag, loads more, and goes back', async () => {
    await searchWith([tagged(1, ['scifi'])])
    const page = (from, n) => Array.from({ length: n }, (_, i) => tagged(from + i, ['scifi']))
    fetch.mockImplementation((url) => {
      const offset = Number(new URL(url, 'http://x').searchParams.get('offset'))
      return json(offset === 0 ? page(1, 20) : page(21, 3))
    })
    fireEvent.click(screen.getByRole('button', { name: 'scifi' }))
    await act(() => vi.advanceTimersByTimeAsync(0))
    expect(fetch.mock.calls.at(-1)[0]).toBe('/api/books?tag=scifi&limit=20&offset=0')
    expect(screen.getByRole('heading', { name: /scifi/ })).toBeTruthy()
    expect(screen.getAllByRole('listitem')).toHaveLength(20)
    fireEvent.click(screen.getByText('Load more'))
    await act(() => vi.advanceTimersByTimeAsync(0))
    expect(fetch.mock.calls.at(-1)[0]).toBe('/api/books?tag=scifi&limit=20&offset=20')
    expect(screen.getAllByRole('listitem')).toHaveLength(23)
    expect(screen.queryByText('Load more')).toBeNull()
    fireEvent.click(screen.getByText('Back to search results'))
    expect(screen.getByText('1 result')).toBeTruthy()
    expect(screen.getByLabelText('Search books').value).toBe('x')
  })

  it('shows loading, then an error when the tag request fails', async () => {
    await searchWith([tagged(1, ['scifi'])])
    fetch.mockReturnValue(json({}, 500))
    fireEvent.click(screen.getByRole('button', { name: 'scifi' }))
    expect(screen.getByText('Loading…')).toBeTruthy()
    await act(() => vi.advanceTimersByTimeAsync(0))
    expect(screen.getByRole('alert').textContent).toContain('Request failed (500)')
  })
})
