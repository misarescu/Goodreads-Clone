import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { AddBook } from './AddBook.jsx'

const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

const fill = (label, value) =>
  fireEvent.change(screen.getByLabelText(label), { target: { value } })
const submit = () => fireEvent.click(screen.getByRole('button', { name: 'Add book' }))

describe('AddBook', () => {
  it('sends only filled fields, disables while saving, then confirms and clears', async () => {
    let resolve
    fetch.mockReturnValue(new Promise((r) => { resolve = r }))
    render(<AddBook />)
    fill('Title', 'Dune')
    fill('Author', 'Frank Herbert')
    fill('Page count', '412')
    fill('Release date', '1965-08-01')
    submit()
    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, opts] = fetch.mock.calls[0]
    expect(url).toBe('/api/books')
    expect(opts.method).toBe('POST')
    expect(JSON.parse(opts.body)).toEqual({
      title: 'Dune', author: 'Frank Herbert', page_count: 412, release_date: '1965-08-01T00:00:00Z',
    })
    await waitFor(() => expect(screen.getByLabelText('Title').closest('fieldset').disabled).toBe(true))
    resolve(await json({ id: 1, title: 'Dune' }, 201))
    expect(await screen.findByText('Added Dune')).toBeTruthy()
    expect(screen.getByLabelText('Title').value).toBe('')
    expect(screen.getByLabelText('Page count').value).toBe('')
    expect(screen.getByLabelText('Title').closest('fieldset').disabled).toBe(false)
  })

  it('checks required fields before sending', () => {
    render(<AddBook />)
    fill('Title', 'Dune')
    submit()
    expect(screen.getByRole('alert').textContent).toBe('Title and author are required.')
    expect(fetch).not.toHaveBeenCalled()
  })

  it('shows the API 400 message and keeps the values', async () => {
    fetch.mockReturnValue(json({ Error: 'cover_url must be an absolute http or https URL' }, 400))
    render(<AddBook />)
    fill('Title', 'Dune')
    fill('Author', 'Frank Herbert')
    fill('Cover URL', 'nope')
    submit()
    expect((await screen.findByRole('alert')).textContent).toMatch(/cover_url must be/)
    expect(screen.getByLabelText('Cover URL').value).toBe('nope')
    expect(screen.getByLabelText('Title').value).toBe('Dune')
    expect(screen.queryByText(/^Added/)).toBeNull()
  })

  it('shows a generic message on a network error and keeps the values', async () => {
    fetch.mockRejectedValue(new TypeError('Failed to fetch'))
    render(<AddBook />)
    fill('Title', 'Dune')
    fill('Author', 'Frank Herbert')
    submit()
    expect((await screen.findByRole('alert')).textContent).toBe(
      'Could not add the book. Please try again.',
    )
    expect(screen.getByLabelText('Title').value).toBe('Dune')
  })
})
