<template>
  <AppLayout>
    <div class="space-y-4">
      <section class="card space-y-4 p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">订阅套餐管理</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              这里才是真正的套餐配置：价格、有效期、绑定订阅分组和是否上架都在这里修改。下面的前台商品配置只决定用户端展示哪几个套餐。
            </p>
          </div>
          <button class="btn btn-primary" @click="openCreatePlan">新增套餐</button>
        </div>

        <div v-if="plansLoading" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">正在加载订阅套餐...</div>
        <div v-else-if="!plans.length" class="rounded-lg border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
          当前还没有订阅套餐。请先新增套餐，再到前台商品配置里选择它。
        </div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-600">
            <thead>
              <tr class="text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                <th class="px-3 py-2">套餐</th>
                <th class="px-3 py-2">价格</th>
                <th class="px-3 py-2">有效期</th>
                <th class="px-3 py-2">订阅分组</th>
                <th class="px-3 py-2">额度</th>
                <th class="px-3 py-2">状态</th>
                <th class="px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="plan in plans" :key="plan.id" class="text-gray-700 dark:text-gray-300">
                <td class="px-3 py-3">
                  <div class="font-medium text-gray-900 dark:text-white">{{ plan.name }}</div>
                  <div class="mt-0.5 max-w-md truncate text-xs text-gray-500 dark:text-gray-400">{{ plan.description || '-' }}</div>
                </td>
                <td class="px-3 py-3">
                  <div class="font-semibold text-gray-900 dark:text-white">¥{{ formatAmount(plan.price) }}</div>
                  <div v-if="plan.original_price" class="text-xs text-gray-400 line-through">¥{{ formatAmount(plan.original_price) }}</div>
                </td>
                <td class="px-3 py-3">{{ plan.validity_days }} {{ unitLabel(plan.validity_unit) }}</td>
                <td class="px-3 py-3">{{ plan.group_name || groupName(plan.group_id) || `#${plan.group_id}` }}</td>
                <td class="px-3 py-3 text-xs leading-5">
                  <div>日：{{ formatLimit(plan.daily_limit_usd) }}</div>
                  <div>周：{{ formatLimit(plan.weekly_limit_usd) }}</div>
                  <div>月：{{ formatLimit(plan.monthly_limit_usd) }}</div>
                </td>
                <td class="px-3 py-3">
                  <span :class="['badge', plan.for_sale ? 'badge-success' : 'badge-gray']">
                    {{ plan.for_sale ? '已上架' : '已下架' }}
                  </span>
                </td>
                <td class="px-3 py-3">
                  <div class="flex justify-end gap-2">
                    <button class="btn btn-secondary btn-sm" @click="openEditPlan(plan)">编辑</button>
                    <button class="btn btn-danger btn-sm" @click="deletePlan(plan)">删除</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
      <HomePricingConfigPanel :plans="plans" />
    </div>

    <PlanEditDialog
      :show="showPlanDialog"
      :plan="editingPlan"
      :groups="groups"
      @close="showPlanDialog = false"
      @saved="handlePlanSaved"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import HomePricingConfigPanel from './HomePricingConfigPanel.vue'
import PlanEditDialog from './PlanEditDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const plans = ref<SubscriptionPlan[]>([])
const groups = ref<AdminGroup[]>([])
const plansLoading = ref(false)
const showPlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    plansLoading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function openCreatePlan() {
  editingPlan.value = null
  showPlanDialog.value = true
}

function openEditPlan(plan: SubscriptionPlan) {
  editingPlan.value = plan
  showPlanDialog.value = true
}

async function deletePlan(plan: SubscriptionPlan) {
  if (!window.confirm(`确定删除套餐「${plan.name}」吗？`)) return
  try {
    await adminPaymentAPI.deletePlan(plan.id)
    appStore.showSuccess('套餐已删除')
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

async function handlePlanSaved() {
  await loadPlans()
}

function formatAmount(value?: number) {
  return Number(value || 0).toFixed(2)
}

function formatLimit(value?: number | null) {
  return value == null ? '不限' : `${Number(value).toFixed(2)} 积分`
}

function groupName(groupID: number) {
  return groups.value.find(group => group.id === groupID)?.name || ''
}

function unitLabel(unit?: string) {
  if (unit === 'weeks') return '周'
  if (unit === 'months') return '月'
  return '天'
}

onMounted(() => {
  loadPlans()
  loadGroups()
})
</script>
