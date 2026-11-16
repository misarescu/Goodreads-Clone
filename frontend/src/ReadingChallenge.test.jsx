import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { ReadingChallenge } from './ReadingChallenge.jsx'

const year = new Date().getFullYear()
const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
const challenge = (o = {}) => ({
  year, goal: 24, finished: 3, expected_by_now: 3, ahead_by: 0, books: [], ...o,
})

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('ReadingChallenge', () => {
  it('shows loading while pending', () => {
    fetch.mockReturnValue(new Promise(() => {}))
    render(<ReadingChallenge userId="1" />)
    expect(screen.getByText('Loading challenge…')).toBeTruthy()
  })

  it('shows the card', async () => {
    fetch.mockReturnValue(json(challenge()))
    render(<ReadingChallenge userId="1" />)
    expect(await screen.findByText(`${year} reading challenge`)).toBeTruthy()
    expect(fetch.mock.calls[0][0]).toBe(`/api/users/1/challenges/${year}`)
    const bar = document.querySelector('progress')
    expect(bar.max).toBe(24)
    expect(bar.value).toBe(3)
    expect(screen.getByText('3 of 24 books')).toBeTruthy()
    expect(screen.getByText(`3 books finished in ${year}`)).toBeTruthy()
    expect(screen.getByText('On schedule')).toBeTruthy()
  })

  it('lists the books finished this year', async () => {
    const books = [
      { id: 2, title: 'Dune', author: 'Frank Herbert', finished_at: `${year}-05-01T10:00:00Z` },
      { id: 1, title: 'Emma', author: 'Jane Austen', finished_at: `${year}-02-01T10:00:00Z` },
    ]
    fetch.mockReturnValue(json(challenge({ finished: 2, books })))
    render(<ReadingChallenge userId="1" />)
    expect(await screen.findByText('Finished this year')).toBeTruthy()
    const items = screen.getAllByRole('listitem')
    expect(items).toHaveLength(2)
    expect(items[0].textContent).toContain('Dune by Frank Herbert')
    expect(items[1].textContent).toContain('Emma by Jane Austen')
    expect(screen.queryByText('No books finished yet this year.')).toBeNull()
  })

  it('shows an empty state when no books are finished', async () => {
    fetch.mockReturnValue(json(challenge({ finished: 0, books: [] })))
    render(<ReadingChallenge userId="1" />)
    expect(await screen.findByText('Finished this year')).toBeTruthy()
    expect(screen.getByText('No books finished yet this year.')).toBeTruthy()
    expect(screen.queryAllByRole('listitem')).toHaveLength(0)
  })

  it.each([
    [3, '3 books ahead of schedule'],
    [1, '1 book ahead of schedule'],
    [-2, '2 books behind schedule'],
    [-1, '1 book behind schedule'],
    [0, 'On schedule'],
  ])('schedule line for ahead_by %i', async (ahead, text) => {
    fetch.mockReturnValue(json(challenge({ ahead_by: ahead })))
    render(<ReadingChallenge userId="1" />)
    expect(await screen.findByText(text)).toBeTruthy()
  })

  it('caps progress at goal', async () => {
    fetch.mockReturnValue(json(challenge({ finished: 30 })))
    render(<ReadingChallenge userId="1" />)
    await screen.findByText('30 of 24 books')
    expect(document.querySelector('progress').value).toBe(24)
  })

  it('shows a form on 404 and saves then reloads', async () => {
    fetch
      .mockReturnValueOnce(json({}, 404))
      .mockReturnValueOnce(json({ year, goal: 12 }))
      .mockReturnValueOnce(json(challenge({ goal: 12, finished: 0 })))
    render(<ReadingChallenge userId="1" />)
    const input = await screen.findByLabelText(/Books to read/)
    expect(input.min).toBe('1')
    expect(input.max).toBe('1000')
    expect(document.querySelector('progress')).toBeNull()
    fireEvent.change(input, { target: { value: '12' } })
    fireEvent.click(screen.getByRole('button'))
    expect(await screen.findByText('0 of 12 books')).toBeTruthy()
    const [url, opts] = fetch.mock.calls[1]
    expect(url).toBe(`/api/users/1/challenges/${year}`)
    expect(opts.method).toBe('PUT')
    expect(JSON.parse(opts.body)).toEqual({ goal: 12 })
    expect(fetch).toHaveBeenCalledTimes(3)
  })

  it('keeps the form usable when saving fails', async () => {
    fetch.mockReturnValueOnce(json({}, 404)).mockReturnValueOnce(json({}, 400))
    render(<ReadingChallenge userId="1" />)
    fireEvent.change(await screen.findByLabelText(/Books to read/), {
      target: { value: '5' },
    })
    fireEvent.click(screen.getByRole('button'))
    expect((await screen.findByRole('alert')).textContent).toContain('400')
    await waitFor(() => expect(screen.getByRole('button').disabled).toBe(false))
    expect(screen.getByLabelText(/Books to read/)).toBeTruthy()
  })

  it('shows an alert and no form on non-404 errors', async () => {
    fetch.mockReturnValue(json({}, 500))
    render(<ReadingChallenge userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain(
      'Could not load reading challenge',
    )
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('shows an alert on network errors', async () => {
    fetch.mockRejectedValue(new Error('boom'))
    render(<ReadingChallenge userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain('boom')
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('resets when the user changes', async () => {
    fetch
      .mockReturnValueOnce(json(challenge({ finished: 3 })))
      .mockReturnValueOnce(new Promise(() => {}))
    const { rerender } = render(<ReadingChallenge userId="1" />)
    await screen.findByText('3 of 24 books')
    rerender(<ReadingChallenge userId="2" />)
    expect(screen.getByText('Loading challenge…')).toBeTruthy()
    expect(screen.queryByText('3 of 24 books')).toBeNull()
    expect(fetch.mock.calls[1][0]).toBe(`/api/users/2/challenges/${year}`)
  })
})
