import { describe, expect, it } from 'vitest'
import { formatCurrency, formatDate, shortId } from './format'

describe('formatCurrency', () => {
  it('formats IDR minor units to major rupiah', () => {
    expect(formatCurrency(1_500_000_00)).toMatch(/1\.500\.000/)
  })

  it('handles zero', () => {
    expect(formatCurrency(0)).toMatch(/0/)
  })

  it('uses currency code for non-IDR', () => {
    expect(formatCurrency(100_00, 'USD')).toContain('USD')
  })
})

describe('formatDate', () => {
  it('returns em-dash for empty', () => {
    expect(formatDate(null)).toBe('—')
    expect(formatDate('')).toBe('—')
    expect(formatDate('not-a-date')).toBe('—')
  })

  it('formats ISO date', () => {
    const out = formatDate('2026-04-15T10:00:00Z')
    expect(out).toMatch(/2026/)
    expect(out).toMatch(/Apr/)
  })
})

describe('shortId', () => {
  it('truncates long ids with ellipsis', () => {
    expect(shortId('1234567890abcdef')).toBe('12345678…')
  })
  it('returns short ids unchanged', () => {
    expect(shortId('abc')).toBe('abc')
  })
})
