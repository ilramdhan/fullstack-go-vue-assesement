import { describe, expect, it } from 'vitest'
import { screen, waitFor } from '@testing-library/vue'
import { renderWithStack } from '@/test/utils'
import SummaryWidget from './SummaryWidget.vue'
import '@/api/http'

describe('SummaryWidget', () => {
  it('renders aggregated counts from the API', async () => {
    renderWithStack(SummaryWidget)

    await waitFor(() => {
      expect(screen.getByText('Total payments')).toBeInTheDocument()
    })
    await waitFor(() => {
      expect(screen.getByText('4')).toBeInTheDocument()
      expect(screen.getAllByText(/^[12]$/).length).toBeGreaterThan(0)
    })
  })
})
