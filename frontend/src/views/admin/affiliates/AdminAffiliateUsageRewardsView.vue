<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.rewards.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.affiliates.rewards.description') }}</p>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <div class="relative w-full md:w-80">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input v-model="filters.search" class="input pl-10" :placeholder="t('admin.affiliates.rewards.searchPlaceholder')" @input="scheduleLoad" />
        </div>
        <select v-model="filters.status" class="input w-full sm:w-40" @change="reloadFirstPage">
          <option value="">{{ t('admin.affiliates.rewards.allStatuses') }}</option>
          <option value="pending">{{ t('admin.affiliates.rewards.status.pending') }}</option>
          <option value="granted">{{ t('admin.affiliates.rewards.status.granted') }}</option>
          <option value="rejected">{{ t('admin.affiliates.rewards.status.rejected') }}</option>
        </select>
        <input v-model="filters.start_at" type="date" class="input w-full sm:w-44" @change="reloadFirstPage" />
        <input v-model="filters.end_at" type="date" class="input w-full sm:w-44" @change="reloadFirstPage" />
        <button type="button" class="btn btn-secondary px-3" :disabled="loading" @click="loadRewards">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th v-for="column in columns" :key="column" class="whitespace-nowrap px-5 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t(`admin.affiliates.rewards.columns.${column}`) }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
              <tr v-if="loading">
                <td :colspan="columns.length" class="px-5 py-12 text-center text-sm text-gray-500">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="items.length === 0">
                <td :colspan="columns.length" class="px-5 py-12 text-center text-sm text-gray-500">{{ t('admin.affiliates.rewards.empty') }}</td>
              </tr>
              <tr v-for="item in items" v-else :key="item.id" class="hover:bg-gray-50/70 dark:hover:bg-dark-800/60">
                <td class="whitespace-nowrap px-5 py-4">
                  <UserIdentity :id="item.inviter_id" :email="item.inviter_email" :username="item.inviter_username" />
                </td>
                <td class="whitespace-nowrap px-5 py-4">
                  <UserIdentity :id="item.invitee_id" :email="item.invitee_email" :username="item.invitee_username" />
                </td>
                <td class="whitespace-nowrap px-5 py-4 text-sm font-semibold text-gray-900 dark:text-white">{{ formatNumber(item.amount) }}</td>
                <td class="whitespace-nowrap px-5 py-4">
                  <span class="inline-flex rounded-full px-2.5 py-1 text-xs font-medium" :class="statusClass(item.status)">
                    {{ t(`admin.affiliates.rewards.status.${item.status}`) }}
                  </span>
                </td>
                <td class="whitespace-nowrap px-5 py-4 text-sm text-gray-600 dark:text-gray-300">
                  {{ t(`admin.affiliates.rewards.reasons.${item.risk_reason}`, item.risk_reason || '-') }}
                </td>
                <td class="whitespace-nowrap px-5 py-4 text-sm text-gray-600 dark:text-gray-300">{{ item.trigger_ip || '-' }}</td>
                <td class="whitespace-nowrap px-5 py-4 text-sm text-gray-600 dark:text-gray-300">{{ formatDateTime(item.created_at) }}</td>
                <td class="whitespace-nowrap px-5 py-4 text-sm">
                  <button v-if="item.status === 'pending'" type="button" class="font-medium text-primary-600 hover:underline dark:text-primary-400" @click="openReview(item)">
                    {{ t('admin.affiliates.rewards.review') }}
                  </button>
                  <span v-else class="text-gray-400">{{ item.reviewer_email || '-' }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        v-if="total > 0"
        :page="page"
        :page-size="pageSize"
        :total="total"
        @update:page="changePage"
        @update:pageSize="changePageSize"
      />
    </div>

    <BaseDialog :show="reviewOpen" :title="t('admin.affiliates.rewards.reviewTitle')" width="narrow" @close="closeReview">
      <div v-if="reviewTarget" class="space-y-4">
        <div class="rounded-lg bg-gray-50 p-4 text-sm dark:bg-dark-800">
          <p class="font-medium text-gray-900 dark:text-white">{{ reviewTarget.inviter_email }} ← {{ reviewTarget.invitee_email }}</p>
          <p class="mt-1 text-gray-500 dark:text-gray-400">
            {{ t('admin.affiliates.rewards.reviewSummary', { amount: formatNumber(reviewTarget.amount), reason: t(`admin.affiliates.rewards.reasons.${reviewTarget.risk_reason}`, reviewTarget.risk_reason) }) }}
          </p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.affiliates.rewards.reviewRemark') }}</label>
          <textarea v-model="reviewRemark" class="input" rows="3" maxlength="500" :placeholder="t('admin.affiliates.rewards.reviewRemarkPlaceholder')"></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-danger" :disabled="reviewing" @click="submitReview(false)">{{ t('admin.affiliates.rewards.reject') }}</button>
          <button type="button" class="btn btn-primary" :disabled="reviewing" @click="submitReview(true)">{{ t('admin.affiliates.rewards.approve') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { affiliatesAPI, type AffiliateUsageRewardRecord, type AffiliateUsageRewardStatus } from '@/api/admin/affiliates'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime as formatDisplayDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const columns = ['inviter', 'invitee', 'amount', 'status', 'riskReason', 'triggerIP', 'createdAt', 'actions'] as const
const items = ref<AffiliateUsageRewardRecord[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const filters = reactive<{ search: string; status: AffiliateUsageRewardStatus | ''; start_at: string; end_at: string }>({
  search: '',
  status: '',
  start_at: '',
  end_at: '',
})
let searchTimer: number | null = null

const reviewOpen = ref(false)
const reviewTarget = ref<AffiliateUsageRewardRecord | null>(null)
const reviewRemark = ref('')
const reviewing = ref(false)

const UserIdentity = defineComponent({
  props: { id: Number, email: String, username: String },
  setup(props) {
    return () => h('div', { class: 'min-w-[180px]' }, [
      h('div', { class: 'text-sm font-medium text-gray-900 dark:text-white' }, props.email || '-'),
      h('div', { class: 'mt-0.5 text-xs text-gray-500 dark:text-gray-400' }, `#${props.id} · ${props.username || '-'}`),
    ])
  },
})

function statusClass(status: AffiliateUsageRewardStatus) {
  if (status === 'granted') return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  if (status === 'rejected') return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
}

function formatNumber(value: number) {
  return Number(value || 0).toFixed(2).replace(/\.00$/, '')
}

function formatDateTime(value: string) {
  return formatDisplayDateTime(value)
}

async function loadRewards() {
  loading.value = true
  try {
    const response = await affiliatesAPI.listUsageRewards({
      page: page.value,
      page_size: pageSize.value,
      search: filters.search,
      status: filters.status,
      start_at: filters.start_at,
      end_at: filters.end_at,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    })
    items.value = response.items || []
    total.value = response.total || 0
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function scheduleLoad() {
  if (searchTimer != null) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    page.value = 1
    void loadRewards()
  }, 300)
}

function reloadFirstPage() {
  page.value = 1
  void loadRewards()
}

function changePage(value: number) {
  page.value = value
  void loadRewards()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadRewards()
}

function openReview(item: AffiliateUsageRewardRecord) {
  reviewTarget.value = item
  reviewRemark.value = ''
  reviewOpen.value = true
}

function closeReview() {
  if (reviewing.value) return
  reviewOpen.value = false
  reviewTarget.value = null
}

async function submitReview(approve: boolean) {
  if (!reviewTarget.value) return
  reviewing.value = true
  try {
    await affiliatesAPI.reviewUsageReward(reviewTarget.value.id, { approve, remark: reviewRemark.value })
    appStore.showSuccess(approve ? t('admin.affiliates.rewards.approved') : t('admin.affiliates.rewards.rejected'))
    reviewOpen.value = false
    reviewTarget.value = null
    await loadRewards()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    reviewing.value = false
  }
}

onMounted(loadRewards)
onBeforeUnmount(() => {
  if (searchTimer != null) window.clearTimeout(searchTimer)
})
</script>
