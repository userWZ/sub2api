<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="dollar" size="sm" class="text-primary-500" />
              {{ t('affiliate.stats.rebateRate') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-primary-600 dark:text-primary-400">
              {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.stats.rebateRateHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCount(detail.aff_count) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(detail.aff_quota) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCurrency(detail.aff_history_quota) }}
            </p>
            <p v-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.description') }}</p>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ t('affiliate.tips.title') }}</p>
            <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
              <li>1. {{ t('affiliate.tips.line1') }}</li>
              <li>2. {{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
            </ul>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.records.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.records.description') }}</p>
            </div>
            <Select
              v-model="recordFilters.action"
              :options="recordActionOptions"
              class="w-full sm:w-44"
              @change="reloadRecordsFromFirstPage"
            />
          </div>

          <div v-if="recordsLoading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          </div>
          <div v-else-if="records.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.records.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[760px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.type') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.records.columns.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.sourceOrder') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.records.columns.availableAfter') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.status') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.records.columns.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="record in records"
                  :key="record.ledger_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3">
                    <span :class="recordBadgeClass(record.action)">
                      {{ recordActionLabel(record.action) }}
                    </span>
                  </td>
                  <td class="px-3 py-3 text-right font-semibold" :class="recordAmountClass(record.action)">
                    {{ formatRecordAmount(record) }}
                  </td>
                  <td class="px-3 py-3">
                    <div v-if="record.source_order_id" class="space-y-0.5">
                      <div class="font-mono text-sm text-gray-900 dark:text-white">#{{ record.source_order_id }}</div>
                      <div class="max-w-52 truncate text-xs text-gray-500 dark:text-dark-400">{{ record.out_trade_no || '-' }}</div>
                    </div>
                    <span v-else class="text-gray-400 dark:text-dark-500">-</span>
                  </td>
                  <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-300">{{ formatOptionalCurrency(record.available_quota_after) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">
                    <span class="max-w-56 truncate" :title="record.remark || recordStatus(record)">
                      {{ record.remark || recordStatus(record) }}
                    </span>
                  </td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(record.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <Pagination
            v-if="recordPagination.total > recordPagination.page_size"
            class="-mx-6 -mb-6 mt-4"
            :page="recordPagination.page"
            :total="recordPagination.total"
            :page-size="recordPagination.page_size"
            :show-jump="false"
            @update:page="handleRecordsPageChange"
            @update:pageSize="handleRecordsPageSizeChange"
          />
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 font-medium text-right">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail, UserAffiliateLedgerAction, UserAffiliateLedgerRecord } from '@/types'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const detail = ref<UserAffiliateDetail | null>(null)
const recordsLoading = ref(false)
const records = ref<UserAffiliateLedgerRecord[]>([])
const recordFilters = reactive<{ action: UserAffiliateLedgerAction | '' }>({ action: '' })
const recordPagination = reactive({ page: 1, page_size: 10, total: 0 })

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

const recordActionOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('affiliate.records.allTypes') },
  { value: 'accrue', label: t('affiliate.records.actions.accrue') },
  { value: 'discount', label: t('affiliate.records.actions.discount') },
  { value: 'withdraw', label: t('affiliate.records.actions.withdraw') },
])

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function loadAffiliateRecords(): Promise<void> {
  recordsLoading.value = true
  try {
    const res = await userAPI.getAffiliateRecords({
      page: recordPagination.page,
      page_size: recordPagination.page_size,
      action: recordFilters.action,
    })
    records.value = res.items || []
    recordPagination.total = res.total || 0
    recordPagination.page = res.page || recordPagination.page
    recordPagination.page_size = res.page_size || recordPagination.page_size
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.records.loadFailed')))
  } finally {
    recordsLoading.value = false
  }
}

function reloadRecordsFromFirstPage(): void {
  recordPagination.page = 1
  void loadAffiliateRecords()
}

function handleRecordsPageChange(page: number): void {
  recordPagination.page = page
  void loadAffiliateRecords()
}

function handleRecordsPageSizeChange(size: number): void {
  recordPagination.page_size = size
  recordPagination.page = 1
  void loadAffiliateRecords()
}

function isPositiveRecord(action: string | null | undefined): boolean {
  return action === 'accrue'
}

function recordActionLabel(action: string | null | undefined): string {
  if (action === 'accrue') return t('affiliate.records.actions.accrue')
  if (action === 'discount') return t('affiliate.records.actions.discount')
  if (action === 'withdraw') return t('affiliate.records.actions.withdraw')
  return action || '-'
}

function recordBadgeClass(action: string | null | undefined): string {
  const base = 'inline-flex rounded px-2 py-0.5 text-xs font-medium'
  if (isPositiveRecord(action)) {
    return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300`
  }
  return `${base} bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300`
}

function recordAmountClass(action: string | null | undefined): string {
  return isPositiveRecord(action) ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'
}

function formatRecordAmount(record: UserAffiliateLedgerRecord): string {
  const amount = Math.abs(Number(record.amount || 0))
  const sign = isPositiveRecord(record.action) ? '+' : '-'
  return `${sign}${formatCurrency(amount)}`
}

function formatOptionalCurrency(value: number | null | undefined): string {
  return value === null || value === undefined ? '-' : formatCurrency(value)
}

function recordStatus(record: UserAffiliateLedgerRecord): string {
  if (record.action === 'accrue' && record.frozen_until) {
    const frozenUntil = new Date(record.frozen_until)
    if (!Number.isNaN(frozenUntil.getTime()) && frozenUntil.getTime() > Date.now()) {
      return t('affiliate.records.status.frozenUntil', { time: formatDateTime(record.frozen_until) })
    }
  }
  if (record.action === 'accrue') return t('affiliate.records.status.posted')
  if (record.action === 'withdraw') return t('affiliate.records.status.withdrawn')
  if (record.action === 'discount') return t('affiliate.records.status.deducted')
  return '-'
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

onMounted(() => {
  void loadAffiliateDetail()
  void loadAffiliateRecords()
})
</script>
