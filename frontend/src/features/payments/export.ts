import { utils, writeFile } from 'xlsx'
import type { Payment } from '@/api/generated'

export function exportPaymentsToExcel(rows: Payment[], filenameHint = 'payments') {
  const data = rows.map((p) => ({
    'Payment ID': p.id,
    Merchant: p.merchant,
    'Created At': p.created_at,
    'Amount (minor)': p.amount,
    'Amount (IDR)': p.amount / 100,
    Currency: p.currency,
    Status: p.status,
    'Reviewed By': p.reviewed_by ?? '',
    'Reviewed At': p.reviewed_at ?? '',
  }))

  const ws = utils.json_to_sheet(data)
  ws['!cols'] = [
    { wch: 38 },
    { wch: 16 },
    { wch: 22 },
    { wch: 14 },
    { wch: 14 },
    { wch: 8 },
    { wch: 12 },
    { wch: 22 },
    { wch: 22 },
  ]

  const wb = utils.book_new()
  utils.book_append_sheet(wb, ws, 'Payments')

  const stamp = new Date().toISOString().slice(0, 10)
  writeFile(wb, `${filenameHint}-${stamp}.xlsx`)
}
