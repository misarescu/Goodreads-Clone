import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { cleanup, render, screen } from '@testing-library/react'
import { Recommendations } from './Recommendations.jsx'

const json = (body, status = 200) =>
  Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })

beforeEach(() => vi.stubGlobal('fetch', vi.fn()))
afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
})

describe('Recommendations', () => {
  it('shows loading while pending', () => {
    fetch.mockReturnValue(new Promise(() => {}))
    render(<Recommendations userId="1" />)
    expect(screen.getByText('Loading recommendations…')).toBeTruthy()
  })

  it('lists books with the reason', async () => {
    fetch.mockReturnValue(
      json([
        { id: 1, title: 'Dune', author: 'Herbert', score: 2, because: ['Emma', 'Ulysses'] },
        { id: 2, title: 'Solo', author: 'Nobody', score: 1, because: [] },
      ]),
    )
    render(<Recommendations userId="1" />)
    expect(await screen.findByText('Dune')).toBeTruthy()
    expect(fetch.mock.calls[0][0]).toBe('/api/users/1/recommendations')
    expect(screen.getByText('You might like')).toBeTruthy()
    expect(screen.getByText(/Herbert/)).toBeTruthy()
    expect(screen.getByText(/Emma, Ulysses/)).toBeTruthy()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
    expect(screen.getAllByText(/Because/)).toHaveLength(1)
  })

  it('shows the empty state', async () => {
    fetch.mockReturnValue(json([]))
    render(<Recommendations userId="1" />)
    expect(await screen.findByText(/No recommendations yet/)).toBeTruthy()
    expect(screen.queryByRole('list')).toBeNull()
  })

  it('shows an alert on API errors', async () => {
    fetch.mockReturnValue(json({}, 500))
    render(<Recommendations userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain(
      'Could not load recommendations',
    )
  })

  it('shows an alert on network errors', async () => {
    fetch.mockRejectedValue(new Error('boom'))
    render(<Recommendations userId="1" />)
    expect((await screen.findByRole('alert')).textContent).toContain('boom')
  })

  it('resets when the user changes', async () => {
    fetch
      .mockReturnValueOnce(json([{ id: 1, title: 'Dune', author: 'H', score: 1, because: [] }]))
      .mockReturnValueOnce(new Promise(() => {}))
    const { rerender } = render(<Recommendations userId="1" />)
    await screen.findByText('Dune')
    rerender(<Recommendations userId="2" />)
    expect(screen.getByText('Loading recommendations…')).toBeTruthy()
    expect(screen.queryByText('Dune')).toBeNull()
    expect(fetch.mock.calls[1][0]).toBe('/api/users/2/recommendations')
  })
})
