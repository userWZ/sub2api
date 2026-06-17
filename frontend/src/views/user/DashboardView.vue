<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else>
        <section class="ow-dashboard-hero">
          <div class="ow-dashboard-main">
            <p class="ow-eyebrow">{{ copy.workbench }}</p>
            <div class="flex flex-wrap items-center gap-3">
              <h1>{{ accountStatus.title }}</h1>
              <span :class="['ow-status-pill', accountStatus.tone]">{{ accountStatus.label }}</span>
            </div>
            <p class="ow-desc">{{ accountStatus.description }}</p>

            <div class="ow-quota-grid">
              <div class="ow-quota-card">
                <span>{{ copy.balanceQuota }}</span>
                <strong>{{ formatQuota(balanceQuota) }}</strong>
                <small>{{ copy.balanceQuotaHint }}</small>
              </div>
              <div class="ow-quota-card">
                <span>{{ copy.subscriptionQuota }}</span>
                <strong>{{ subscriptionQuotaText }}</strong>
                <small>{{ copy.subscriptionQuotaHint }}</small>
              </div>
              <div class="ow-quota-card">
                <span>{{ copy.activeSubscriptions }}</span>
                <strong>{{ activeSubscriptions.length }}</strong>
                <small>{{ copy.activeSubscriptionsHint }}</small>
              </div>
              <div class="ow-quota-card">
                <span>{{ copy.concurrency }}</span>
                <strong>{{ user?.concurrency ?? '-' }}</strong>
                <small>{{ copy.concurrencyHint }}</small>
              </div>
            </div>

            <div class="ow-action-row">
              <button class="btn btn-primary" type="button" :disabled="!defaultKey" @click="copyDefaultKey">
                {{ copy.copyDefaultKey }}
              </button>
              <button class="btn btn-secondary" type="button" @click="copyBaseUrl">
                {{ copy.copyBaseUrl }}
              </button>
              <RouterLink class="btn btn-secondary" to="/purchase">{{ copy.recharge }}</RouterLink>
            </div>
          </div>

          <aside class="ow-start-card">
            <div class="ow-start-head">
              <div>
                <p>{{ copy.quickStart }}</p>
                <h2>{{ defaultKey ? defaultKey.name : copy.noDefaultKey }}</h2>
              </div>
              <RouterLink to="/keys">{{ copy.manageKeys }}</RouterLink>
            </div>

            <div class="ow-start-field">
              <span>{{ copy.apiKey }}</span>
              <code>{{ defaultKey ? maskKey(defaultKey.key) : copy.notCreated }}</code>
              <button v-if="defaultKey" type="button" @click="copyDefaultKey">{{ copy.copy }}</button>
              <button v-else type="button" :disabled="creatingKey" @click="createDefaultKey">
                {{ creatingKey ? copy.creating : copy.createDefaultKey }}
              </button>
            </div>

            <div class="ow-start-field">
              <span>{{ copy.baseUrl }}</span>
              <code>{{ baseUrl }}</code>
              <button type="button" @click="copyBaseUrl">{{ copy.copy }}</button>
            </div>

            <div class="ow-route-note">
              <strong>{{ copy.autoRouteTitle }}</strong>
              <p>{{ routeDescription }}</p>
            </div>
          </aside>
        </section>

        <div v-if="stats" class="space-y-6">
          <UserDashboardStats :stats="stats" :balance="balanceQuota" :is-simple="authStore.isSimpleMode" :platform-quotas="platformQuotas" />
          <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
            <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { authAPI, keysAPI, usageAPI } from '@/api'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey, PublicSettings, UsageLog, TrendDataPoint, ModelStat, PlatformQuotaItem, UserSubscription } from '@/types'
import { getMyPlatformQuotas } from '@/api/user'

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()
const { copyToClipboard } = useClipboard()
const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const creatingKey = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const platformQuotas = ref<PlatformQuotaItem[] | null>(null)
const apiKeys = ref<ApiKey[]>([])
const publicSettings = ref<PublicSettings | null>(null)
const activeSubscriptions = ref<UserSubscription[]>([])

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const copy = computed(() => isZh.value ? zhCopy : enCopy)
const balanceQuota = computed(() => Number(user.value?.balance || 0))
const defaultKey = computed(() => apiKeys.value.find(key => key.status === 'active') || apiKeys.value[0] || null)
const baseUrl = computed(() => publicSettings.value?.api_base_url || window.location.origin)
const subscriptionQuota = computed(() => {
  let total = 0
  let hasUnlimited = false
  for (const sub of activeSubscriptions.value) {
    const remaining = getSubscriptionRemaining(sub)
    if (remaining === null) hasUnlimited = true
    else total += remaining
  }
  return { total, hasUnlimited }
})
const subscriptionQuotaText = computed(() => subscriptionQuota.value.hasUnlimited ? copy.value.unlimited : formatQuota(subscriptionQuota.value.total))

const accountStatus = computed(() => {
  if (user.value?.status && user.value.status !== 'active') {
    return { tone: 'danger', label: copy.value.unavailable, title: copy.value.accountDisabledTitle, description: copy.value.accountDisabledDescription }
  }
  if (!defaultKey.value) {
    return { tone: 'warning', label: copy.value.needAction, title: copy.value.needKeyTitle, description: copy.value.needKeyDescription }
  }
  if (balanceQuota.value <= 0 && activeSubscriptions.value.length === 0) {
    return { tone: 'warning', label: copy.value.needQuota, title: copy.value.needQuotaTitle, description: copy.value.needQuotaDescription }
  }
  return { tone: 'success', label: copy.value.available, title: copy.value.availableTitle, description: copy.value.availableDescription }
})

const routeDescription = computed(() => {
  if (activeSubscriptions.value.length) return copy.value.routeUseSubscription.replace('{count}', String(activeSubscriptions.value.length))
  if (balanceQuota.value > 0) return copy.value.routeUseBalance
  return copy.value.routeUnavailable
})

const loadStats = async () => {
  loading.value = true
  try {
    await authStore.refreshUser()
    const [dashboardStats, keyPage, settings, subscriptions] = await Promise.all([
      usageAPI.getDashboardStats(),
      keysAPI.list(1, 20, { sort_by: 'created_at', sort_order: 'desc' }),
      authAPI.getPublicSettings().catch(() => null),
      subscriptionStore.fetchActiveSubscriptions(true).catch(() => []),
    ])
    stats.value = dashboardStats
    apiKeys.value = keyPage.items || []
    publicSettings.value = settings
    activeSubscriptions.value = subscriptions
  } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false }
}
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const loadPlatformQuotas = async () => { try { const data = await getMyPlatformQuotas(); platformQuotas.value = data.platform_quotas ?? [] } catch (error) { console.warn('Failed to load platform quotas:', error); platformQuotas.value = [] } }
const refreshAll = () => { loadStats(); loadCharts(); loadRecent(); loadPlatformQuotas() }

const copyDefaultKey = async () => {
  if (!defaultKey.value) return
  await copyToClipboard(defaultKey.value.key, copy.value.copied)
}

const copyBaseUrl = async () => {
  await copyToClipboard(baseUrl.value, copy.value.copied)
}

const createDefaultKey = async () => {
  if (creatingKey.value) return
  creatingKey.value = true
  try {
    const created = await keysAPI.create(copy.value.defaultKeyName, null, undefined, undefined, undefined, undefined, undefined, undefined, false)
    apiKeys.value = [created, ...apiKeys.value]
    appStore.showSuccess(copy.value.defaultKeyCreated)
    await copyToClipboard(created.key, copy.value.copied)
  } catch (error) {
    appStore.showError(copy.value.defaultKeyFailed)
  } finally {
    creatingKey.value = false
  }
}

const getSubscriptionRemaining = (sub: UserSubscription): number | null => {
  const candidates: number[] = []
  if (sub.group?.daily_limit_usd && sub.group.daily_limit_usd > 0) candidates.push(Math.max(0, sub.group.daily_limit_usd - sub.daily_usage_usd))
  if (sub.group?.weekly_limit_usd && sub.group.weekly_limit_usd > 0) candidates.push(Math.max(0, sub.group.weekly_limit_usd - sub.weekly_usage_usd))
  if (sub.group?.monthly_limit_usd && sub.group.monthly_limit_usd > 0) candidates.push(Math.max(0, sub.group.monthly_limit_usd - sub.monthly_usage_usd))
  if (!candidates.length) return null
  return Math.min(...candidates)
}

const formatQuota = (value: number) => {
  const safeValue = Number.isFinite(value) ? value : 0
  return `$${safeValue.toLocaleString(isZh.value ? 'zh-CN' : 'en-US', { minimumFractionDigits: safeValue >= 100 ? 0 : 2, maximumFractionDigits: 2 })}`
}

const maskKey = (value: string) => {
  if (!value) return ''
  if (value.length <= 12) return value
  return `${value.slice(0, 8)}...${value.slice(-4)}`
}

onMounted(() => { refreshAll() })

const zhCopy = {
  workbench: '用户工作台',
  available: '可用',
  unavailable: '不可用',
  needAction: '需要处理',
  needQuota: '需要额度',
  availableTitle: '当前账户可以发起 API 调用',
  availableDescription: '默认 Key 会优先使用有效订阅，订阅不可用或额度用完后自动按余额兜底。',
  needKeyTitle: '还没有可用 API Key',
  needKeyDescription: '创建默认 Key 后即可复制到客户端开始调用。',
  needQuotaTitle: '当前没有可用额度',
  needQuotaDescription: '请充值额度或购买订阅后再开始调用。',
  accountDisabledTitle: '账户已被禁用',
  accountDisabledDescription: '账户状态会阻止 API 调用，请联系管理员处理。',
  balanceQuota: '余额额度',
  balanceQuotaHint: '账户余额兜底额度',
  subscriptionQuota: '订阅额度',
  subscriptionQuotaHint: '按当前周期剩余额度估算',
  activeSubscriptions: '有效订阅',
  activeSubscriptionsHint: '后台已分配且未过期',
  concurrency: '并发额度',
  concurrencyHint: '账户并发上限',
  unlimited: '不限量',
  copyDefaultKey: '复制默认 Key',
  copyBaseUrl: '复制 Base URL',
  recharge: '充值/购买订阅',
  quickStart: '开始使用',
  noDefaultKey: '默认 Key 未创建',
  manageKeys: '管理密钥',
  apiKey: 'API Key',
  baseUrl: 'Base URL',
  notCreated: '未创建',
  copy: '复制',
  creating: '创建中...',
  createDefaultKey: '创建默认 Key',
  autoRouteTitle: '额度使用方式',
  routeUseSubscription: '当前会优先使用 {count} 个有效订阅；订阅不可用或额度用完后再按余额兜底。',
  routeUseBalance: '当前没有可用订阅，会使用余额额度按量计费。',
  routeUnavailable: '当前没有可用额度，创建调用前需要先充值或购买订阅。',
  defaultKeyName: '默认 API Key',
  defaultKeyCreated: '默认 API Key 已创建并复制',
  defaultKeyFailed: '创建默认 API Key 失败',
  copied: '已复制',
}

const enCopy = {
  workbench: 'User Workbench',
  available: 'Available',
  unavailable: 'Unavailable',
  needAction: 'Action Needed',
  needQuota: 'Quota Needed',
  availableTitle: 'Your account can make API calls',
  availableDescription: 'The default key prefers active subscriptions, then automatically falls back to balance credit.',
  needKeyTitle: 'No API key is ready',
  needKeyDescription: 'Create a default key and copy it into your client to start.',
  needQuotaTitle: 'No usable credit',
  needQuotaDescription: 'Top up balance credit or buy a subscription before calling the API.',
  accountDisabledTitle: 'Account disabled',
  accountDisabledDescription: 'Account status blocks API calls. Contact an administrator.',
  balanceQuota: 'Balance Credit',
  balanceQuotaHint: 'Fallback account balance',
  subscriptionQuota: 'Subscription Credit',
  subscriptionQuotaHint: 'Estimated by current period remaining',
  activeSubscriptions: 'Active Subscriptions',
  activeSubscriptionsHint: 'Assigned and not expired',
  concurrency: 'Concurrency',
  concurrencyHint: 'Account concurrency limit',
  unlimited: 'Unlimited',
  copyDefaultKey: 'Copy Default Key',
  copyBaseUrl: 'Copy Base URL',
  recharge: 'Recharge / Subscription',
  quickStart: 'Start',
  noDefaultKey: 'Default key not created',
  manageKeys: 'Manage Keys',
  apiKey: 'API Key',
  baseUrl: 'Base URL',
  notCreated: 'Not created',
  copy: 'Copy',
  creating: 'Creating...',
  createDefaultKey: 'Create Default Key',
  autoRouteTitle: 'Credit Source',
  routeUseSubscription: 'The default key will prefer {count} active subscription source(s), then fall back to balance when subscription credit is unavailable or exhausted.',
  routeUseBalance: 'No active subscription is available, so the default key will use balance credit with usage-based billing.',
  routeUnavailable: 'No usable credit is available. Recharge or buy a subscription first.',
  defaultKeyName: 'Default API Key',
  defaultKeyCreated: 'Default API key created and copied',
  defaultKeyFailed: 'Failed to create default API key',
  copied: 'Copied',
}
</script>

<style scoped>
.ow-dashboard-hero {
  display: grid;
  gap: 1rem;
  border: 1px solid rgba(203, 213, 225, 0.8);
  border-radius: 0.5rem;
  background: linear-gradient(135deg, rgba(232, 246, 255, 0.92), rgba(255, 255, 255, 0.96));
  padding: 1.25rem;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.06);
}

@media (min-width: 1024px) {
  .ow-dashboard-hero {
    grid-template-columns: minmax(0, 1.4fr) minmax(22rem, 0.8fr);
  }
}

.ow-eyebrow {
  color: #006fd6;
  font-size: 0.78rem;
  font-weight: 900;
  text-transform: uppercase;
}

.ow-dashboard-main h1 {
  color: #0f172a;
  font-size: clamp(1.8rem, 4vw, 3rem);
  font-weight: 950;
  line-height: 1.08;
}

.ow-desc {
  margin-top: 0.75rem;
  max-width: 48rem;
  color: #475569;
  line-height: 1.75;
}

.ow-status-pill {
  border-radius: 999px;
  padding: 0.35rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 900;
}

.ow-status-pill.success { background: #dcfce7; color: #166534; }
.ow-status-pill.warning { background: #fef3c7; color: #92400e; }
.ow-status-pill.danger { background: #fee2e2; color: #991b1b; }

.ow-quota-grid {
  display: grid;
  gap: 0.75rem;
  margin-top: 1rem;
}

@media (min-width: 768px) {
  .ow-quota-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.ow-quota-card,
.ow-start-card {
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.82);
  padding: 0.9rem;
}

.ow-quota-card span,
.ow-start-field span {
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 850;
}

.ow-quota-card strong {
  display: block;
  margin-top: 0.35rem;
  color: #0f172a;
  font-size: 1.2rem;
  font-weight: 950;
}

.ow-quota-card small {
  display: block;
  margin-top: 0.25rem;
  color: #64748b;
  font-size: 0.78rem;
}

.ow-action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  margin-top: 1rem;
}

.ow-start-card {
  background: rgba(255, 255, 255, 0.9);
}

.ow-start-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.ow-start-head p {
  color: #006fd6;
  font-size: 0.78rem;
  font-weight: 900;
}

.ow-start-head h2 {
  margin-top: 0.25rem;
  color: #0f172a;
  font-size: 1rem;
  font-weight: 950;
}

.ow-start-head a {
  color: #006fd6;
  font-size: 0.82rem;
  font-weight: 900;
}

.ow-start-field {
  margin-top: 0.85rem;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.45rem;
}

.ow-start-field span {
  grid-column: 1 / -1;
}

.ow-start-field code {
  min-width: 0;
  overflow-wrap: anywhere;
  border-radius: 0.5rem;
  background: #f8fafc;
  color: #0f172a;
  padding: 0.55rem;
  font-size: 0.8rem;
}

.ow-start-field button {
  border-radius: 0.5rem;
  background: #001040;
  color: white;
  padding: 0.45rem 0.7rem;
  font-size: 0.8rem;
  font-weight: 900;
}

.ow-start-field button:disabled {
  opacity: 0.5;
}

.ow-route-note {
  margin-top: 0.85rem;
  border-radius: 0.5rem;
  background: #eef7ff;
  padding: 0.75rem;
}

.ow-route-note strong {
  color: #0f172a;
  font-size: 0.86rem;
}

.ow-route-note p {
  margin-top: 0.25rem;
  color: #475569;
  font-size: 0.84rem;
  line-height: 1.65;
}

:global(.dark .ow-dashboard-hero) {
  border-color: rgba(51, 65, 85, 0.8);
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.92));
  box-shadow: 0 18px 42px rgba(2, 6, 23, 0.32);
}

:global(.dark .ow-eyebrow),
:global(.dark .ow-start-head p),
:global(.dark .ow-start-head a) {
  color: #38bdf8;
}

:global(.dark .ow-dashboard-main h1),
:global(.dark .ow-quota-card strong),
:global(.dark .ow-start-head h2),
:global(.dark .ow-route-note strong),
:global(.dark .ow-start-field code) {
  color: #f8fafc;
}

:global(.dark .ow-desc),
:global(.dark .ow-route-note p),
:global(.dark .ow-quota-card span),
:global(.dark .ow-start-field span),
:global(.dark .ow-quota-card small) {
  color: #cbd5e1;
}

:global(.dark .ow-quota-card),
:global(.dark .ow-start-card),
:global(.dark .ow-start-field code) {
  border-color: rgba(71, 85, 105, 0.75);
  background: rgba(15, 23, 42, 0.82);
}

:global(.dark .ow-status-pill.success) {
  background: rgba(34, 197, 94, 0.16);
  color: #86efac;
}

:global(.dark .ow-status-pill.warning) {
  background: rgba(245, 158, 11, 0.18);
  color: #fcd34d;
}

:global(.dark .ow-status-pill.danger) {
  background: rgba(239, 68, 68, 0.18);
  color: #fca5a5;
}

:global(.dark .ow-start-field button) {
  background: #0284c7;
  color: #f8fafc;
}

:global(.dark .ow-route-note) {
  background: rgba(14, 116, 144, 0.16);
}
</style>
