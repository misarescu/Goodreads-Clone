import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { BookReviews } from './BookReviews.jsx'

const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
const books = [{ id: 7, title: 'Dune' }, { id: 8, title: 'Emma' }]
const stats = (o = {}) => ({ id: 7, average_rating: 4.5, review_count: 2, ...o })
const review = (id, o = {}) => ({
  id, book_id: 7, user_id: id, username: `user${id}`, rating: 4, body: `body ${id}`, ...o,
})

// Routes fetch calls by "METHOD url" prefix; first match wins.
function route(handlers) {
  fetch.mockImplementation((url, opts = {}) => {
    const key = `${opts.method || 'GET'} ${url}`
    for (const [pattern, res] of handlers) {
      if (key.startsWith(pattern)) return typeof res === 'function' ? res() : res
    }
    return json({}, 404)
  })
}

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('BookReviews', () => {
  it('shows loading while pending', () => {
    fetch.mockReturnValue(new Promise(() => {}))
    render(<BookReviews userId="1" />)
    expect(screen.getByText('Loading books…')).toBeTruthy()
  })

  it('lists stats and reviews in API order', async () => {
    route([
      ['GET /api/books?', json(books)],
      ['GET /api/books/7/reviews', json([review(2), review(1)])],
      ['GET /api/books/7', json(stats())],
    ])
    render(<BookReviews userId="9" />)
    expect(await screen.findByText('body 2')).toBeTruthy()
    expect(screen.getByText(/Average rating: 4.5/).textContent).toContain('2 reviews')
    const items = screen.getAllByRole('listitem').map((li) => li.textContent)
    expect(items[0]).toContain('user2')
    expect(items[1]).toContain('user1')
    expect(screen.queryByText('Load more')).toBeNull()
  })

  it('shows the empty state', async () => {
    route([
      ['GET /api/books?', json(books)],
      ['GET /api/books/7/reviews', json([])],
      ['GET /api/books/7', json(stats({ average_rating: null, review_count: 0 }))],
    ])
    render(<BookReviews userId="1" />)
    expect(await screen.findByText('No reviews yet.')).toBeTruthy()
    expect(screen.getByText(/No ratings yet/).textContent).toContain('0 reviews')
  })

  it('loads more reviews', async () => {
    const page1 = Array.from({ length: 20 }, (_, i) => review(i + 100))
    route([
      ['GET /api/books?', json(books)],
      ['GET /api/books/7/reviews?limit=20&offset=0', json(page1)],
      ['GET /api/books/7/reviews?limit=20&offset=20', json([review(1)])],
      ['GET /api/books/7', json(stats({ review_count: 21 }))],
    ])
    render(<BookReviews userId="9" />)
    fireEvent.click(await screen.findByText('Load more'))
    expect(await screen.findByText('body 1')).toBeTruthy()
    expect(screen.getAllByRole('listitem')).toHaveLength(21)
    expect(screen.queryByText('Load more')).toBeNull()
  })

  it('submits a review and reloads', async () => {
    let saved = false
    route([
      ['GET /api/books?', json(books)],
      ['POST /api/books/7/reviews', () => { saved = true; return json(review(9), 201) }],
      ['GET /api/books/7/reviews', () => json(saved ? [review(9, { body: 'great' })] : [])],
      ['GET /api/books/7', () => json(stats({ review_count: saved ? 1 : 0 }))],
    ])
    render(<BookReviews userId="9" />)
    await screen.findByText('No reviews yet.')
    fireEvent.change(screen.getByLabelText('Rating'), { target: { value: '3' } })
    fireEvent.change(screen.getByLabelText('Review'), { target: { value: 'great' } })
    fireEvent.click(screen.getByText('Save review'))
    expect(await screen.findByText('great')).toBeTruthy()
    const post = fetch.mock.calls.find(([, o]) => o?.method === 'POST')
    expect(post[0]).toBe('/api/books/7/reviews')
    expect(JSON.parse(post[1].body)).toEqual({ user_id: 9, rating: 3, body: 'great' })
  })

  it('deletes the review', async () => {
    let deleted = false
    route([
      ['GET /api/books?', json(books)],
      ['DELETE /api/books/7/reviews/9', () => { deleted = true; return json({}) }],
      ['GET /api/books/7/reviews', () => json(deleted ? [] : [review(9)])],
      ['GET /api/books/7', json(stats())],
    ])
    render(<BookReviews userId="9" />)
    expect((await screen.findByLabelText('Review')).value).toBe('body 9')
    fireEvent.click(screen.getByText('Delete my review'))
    expect(await screen.findByText('No reviews yet.')).toBeTruthy()
    expect(screen.getByLabelText('Review').value).toBe('')
  })

  it('shows an alert when saving fails', async () => {
    route([
      ['GET /api/books?', json(books)],
      ['POST /api/books/7/reviews', json({}, 400)],
      ['GET /api/books/7/reviews', json([])],
      ['GET /api/books/7', json(stats())],
    ])
    render(<BookReviews userId="9" />)
    await screen.findByText('No reviews yet.')
    fireEvent.change(screen.getByLabelText('Review'), { target: { value: 'x' } })
    fireEvent.click(screen.getByText('Save review'))
    expect((await screen.findByRole('alert')).textContent).toContain('400')
  })

  it('shows an alert when loading reviews fails', async () => {
    route([
      ['GET /api/books?', json(books)],
      ['GET /api/books/7/reviews', json({}, 500)],
      ['GET /api/books/7', json(stats())],
    ])
    render(<BookReviews userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain('Could not load reviews')
  })

  it('shows an alert when loading books fails', async () => {
    fetch.mockRejectedValue(new Error('boom'))
    render(<BookReviews userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain('boom')
  })
})
