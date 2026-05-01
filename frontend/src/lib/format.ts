const idrFormatter = new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
})

const numberFormatter = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 })

const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

export function formatCurrency(amountMinor: number, currency = 'IDR'): string {
  const major = amountMinor / 100
  if (currency === 'IDR') return idrFormatter.format(major)
  return `${numberFormatter.format(major)} ${currency}`
}

export function formatDate(iso: string | Date | null | undefined): string {
  if (!iso) return '—'
  const d = typeof iso === 'string' ? new Date(iso) : iso
  if (Number.isNaN(d.getTime())) return '—'
  return dateFormatter.format(d)
}

export function shortId(id: string, length = 8): string {
  return id.length > length ? id.slice(0, length) + '…' : id
}
