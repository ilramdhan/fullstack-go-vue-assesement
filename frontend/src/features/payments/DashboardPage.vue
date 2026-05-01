<script setup lang="ts">
import AppLayout from '@/components/AppLayout.vue'
import SummaryWidget from './SummaryWidget.vue'
import PaymentsTable from './PaymentsTable.vue'
</script>

<template>
  <AppLayout>
    <div class="mb-6">
      <h2 class="text-2xl font-semibold tracking-tight">Payments</h2>
      <p class="text-sm text-muted-foreground">
        Monitor incoming payments. Operations can review processing transactions.
      </p>
    </div>

    <SummaryWidget class="mb-6" />
    <PaymentsTable />
  </AppLayout>
</template>
