import { describe, expect, it } from 'vitest'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { renderWithStack } from '@/test/utils'
import PaymentsTable from './PaymentsTable.vue'
import { useAuthStore } from '@/stores/auth'
import '@/api/http'

function loginAs(role: 'cs' | 'operation') {
  useAuthStore().setSession({
    token: 't',
    email: role === 'cs' ? 'cs@test.com' : 'operation@test.com',
    role,
  })
}

async function setup(role: 'cs' | 'operation') {
  const result = renderWithStack(PaymentsTable)
  loginAs(role)
  await waitFor(() => expect(screen.getByText('Tokopedia')).toBeInTheDocument())
  return result
}

describe('PaymentsTable', () => {
  it('renders rows for cs without an Actions column', async () => {
    await setup('cs')
    expect(screen.queryByRole('columnheader', { name: /actions/i })).toBeNull()
    expect(screen.getByText('Shopee')).toBeInTheDocument()
    expect(screen.getByText('Lazada')).toBeInTheDocument()
  })

  it('shows Actions with review buttons for operation', async () => {
    await setup('operation')
    expect(screen.getByRole('columnheader', { name: /actions/i })).toBeInTheDocument()

    const processingRow = screen.getByText('Shopee').closest('tr')!
    expect(within(processingRow).getByRole('button', { name: /approve/i })).toBeInTheDocument()
    expect(within(processingRow).getByRole('button', { name: /reject/i })).toBeInTheDocument()

    const completedRow = screen.getByText('Tokopedia').closest('tr')!
    expect(within(completedRow).queryByRole('button', { name: /approve/i })).toBeNull()
  })

  it('filters by status', async () => {
    await setup('cs')

    await fireEvent.click(screen.getByRole('button', { name: /^processing$/i }))

    await waitFor(() => {
      expect(screen.queryByText('Tokopedia')).toBeNull()
      expect(screen.getByText('Shopee')).toBeInTheDocument()
    })
  })

  it('approve flow flips the row to completed', async () => {
    await setup('operation')

    const row = screen.getByText('Shopee').closest('tr')!
    await fireEvent.click(within(row).getByRole('button', { name: /approve/i }))

    const dialog = await screen.findByRole('dialog')
    await fireEvent.click(within(dialog).getByRole('button', { name: /^approve$/i }))

    await waitFor(() => {
      const updated = screen.getByText('Shopee').closest('tr')!
      expect(within(updated).queryByRole('button', { name: /approve/i })).toBeNull()
    })
  })
})
