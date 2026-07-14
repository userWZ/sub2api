<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <section class="purchase-hero">
        <div>
          <p class="purchase-eyebrow">{{ copy.eyebrow }}</p>
          <h1>{{ copy.title }}</h1>
          <p>{{ copy.description }}</p>
        </div>

        <div class="account-panel">
          <span>{{ copy.account }}</span>
          <strong>{{ user?.username || user?.email || '-' }}</strong>
          <small>{{ copy.balance }}: {{ userBalance }}</small>
          <router-link class="orders-link" to="/orders">
            {{ copy.ordersLink }}
          </router-link>
        </div>
      </section>

      <section class="notice-panel">
        <div>
          <h2>{{ copy.cashierTitle }}</h2>
          <p>{{ copy.cashierDescription }}</p>
        </div>
        <div class="method-strip">
          <span
            v-for="method in availablePaymentMethods"
            :key="method"
            class="method-pill"
          >
            {{ paymentMethodLabel(method) }}
          </span>
          <span v-if="!availablePaymentMethods.length" class="method-pill method-pill-muted">
            {{ copy.noPaymentMethod }}
          </span>
        </div>
      </section>

      <section v-if="selectedProduct" ref="confirmPanelRef" class="confirm-panel" aria-live="polite">
        <div>
          <p class="purchase-eyebrow">{{ copy.confirmEyebrow }}</p>
          <h2>{{ selectedProduct.name }}</h2>
          <p>{{ selectedProduct.description }}</p>
        </div>

        <div class="confirm-grid">
          <div class="confirm-detail">
            <span>{{ copy.productAmount }}</span>
            <strong>{{ selectedOriginalAmountText }}</strong>
          </div>
          <div class="confirm-detail">
            <span>{{ copy.arrival }}</span>
            <strong>{{ selectedProduct.arrivalText }}</strong>
          </div>
        </div>

        <div class="checkout-summary">
          <div class="checkout-summary-row">
            <span>{{ copy.orderAmount }}</span>
            <strong>{{ selectedOriginalAmountText }}</strong>
          </div>

          <div v-if="affiliateDiscountVisible" class="affiliate-discount-panel">
            <label class="affiliate-discount-toggle">
              <input
                v-model="useAffiliateDiscount"
                type="checkbox"
                :disabled="!affiliateDiscountSelectable"
              />
              <span class="affiliate-discount-copy">
                <strong>{{ copy.useAffiliateDiscount }}</strong>
                <small>{{ copy.availableDiscount }}: {{ formatCnyPrice(affiliateAvailableQuota) }}</small>
              </span>
            </label>
            <div class="affiliate-discount-summary">
              <span>{{ selectedAffiliateDiscountStatus }}</span>
              <strong>{{ selectedAffiliateDiscountText }}</strong>
            </div>
          </div>

          <div v-if="selectedFeeAmount > 0" class="checkout-summary-row">
            <span>{{ copy.fee }}</span>
            <strong>{{ formatCnyPrice(selectedFeeAmount) }}</strong>
          </div>

          <div class="checkout-summary-row checkout-summary-total">
            <span>{{ copy.payAmount }}</span>
            <strong>{{ selectedPayableText }}</strong>
          </div>
        </div>

        <div class="method-picker">
          <button
            v-for="method in availablePaymentMethods"
            :key="method"
            type="button"
            class="method-button"
            :class="{ 'method-button-active': selectedPaymentMethod === method }"
            :disabled="!methodIsUsableForProduct(method, selectedProduct)"
            @click="selectedPaymentMethod = method"
          >
            {{ paymentMethodLabel(method) }}
          </button>
        </div>

        <p v-if="selectedMethodLimitMessage" class="method-warning">
          {{ selectedMethodLimitMessage }}
        </p>

        <button
          type="button"
          class="submit-button"
          data-testid="selected-product-submit"
          :disabled="submitting || !canSubmitSelectedProduct"
          @click="submitSelectedOrder"
        >
          {{ submitting ? copy.creatingOrder : copy.confirmPay }}
        </button>
      </section>

      <section class="pricing-section">
        <div class="section-heading">
          <p>{{ pricingHeader.eyebrow }}</p>
          <h2>{{ pricingHeader.title }}</h2>
          <span>{{ pricingHeader.description }}</span>
        </div>

        <div v-if="loadingCheckout" class="loading-panel">
          {{ copy.loading }}
        </div>

        <template v-else>
          <div v-if="availablePricingTabs.length > 1" class="pricing-tabs" role="tablist">
            <button
              v-for="tab in availablePricingTabs"
              :key="tab.kind"
              type="button"
              class="pricing-tab"
              :class="{ 'pricing-tab-active': currentProductKind === tab.kind }"
              role="tab"
              :aria-selected="currentProductKind === tab.kind"
              @click="activeProductKind = tab.kind"
            >
              {{ tab.title }}
            </button>
          </div>

          <div
            v-for="group in visiblePricingGroups"
            :key="group.kind"
            class="pricing-group"
          >
            <div class="pricing-group-heading">
              <h3>{{ group.title }}</h3>
              <p>{{ group.description }}</p>
            </div>

            <div class="pricing-grid">
              <article
                v-for="plan in group.plans"
                :key="plan.id"
                class="pricing-card"
                :class="{ 'pricing-card-highlight': plan.highlight }"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="text-base font-black text-slate-950 dark:text-white">{{ plan.name }}</p>
                    <p class="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">{{ plan.description }}</p>
                  </div>
                  <span v-if="plan.badge" class="plan-badge">{{ plan.badge }}</span>
                </div>

                <div class="mt-5">
                  <span class="text-3xl font-black text-slate-950 dark:text-white">{{ plan.priceText }}</span>
                  <span v-if="plan.originalPriceText" class="original-price">{{ plan.originalPriceText }}</span>
                  <span class="ml-1 text-sm font-semibold text-slate-500 dark:text-slate-400">{{ plan.period }}</span>
                </div>

                <dl class="plan-metrics">
                  <div
                    v-for="item in plan.metrics"
                    :key="item.label"
                    class="plan-metric-row"
                  >
                    <dt>{{ item.label }}</dt>
                    <dd>{{ item.value }}</dd>
                  </div>
                </dl>

                <button
                  type="button"
                  class="buy-link"
                  :disabled="plan.disabled || !availablePaymentMethods.length"
                  @click="selectProduct(plan)"
                >
                  {{ plan.kind === 'balance' ? copy.rechargeAction : copy.buyAction }}
                </button>
              </article>
            </div>
          </div>

          <div v-if="!renderedPricingGroups.length" class="pricing-group">
            <div class="pricing-group-heading">
              <h3>{{ copy.serviceGroupTitle }}</h3>
              <p>{{ copy.serviceGroupDescription }}</p>
            </div>
            <div class="empty-service-panel">
              {{ copy.noServicePlans }}
            </div>
          </div>
        </template>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { paymentAPI } from '@/api/payment'
import userAPI from '@/api/user'
import {
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  writePaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import type {
  CheckoutInfoResponse,
  CreateOrderRequest,
  CreateOrderResult,
  MethodLimit,
  OrderType,
  SubscriptionPlan,
} from '@/types/payment'
import type { HomePricingLocalizedText, UserAffiliateDetail } from '@/types'
import { isMobileDevice } from '@/utils/device'
import { extractApiErrorCode, extractI18nErrorMessage } from '@/utils/apiError'
import { formatPoints } from '@/utils/format'
import {
  buildPaymentErrorToastMessage,
  describePaymentScenarioError,
} from './paymentUx'
import {
  formatPaymentAmount,
  normalizePaymentCurrency,
} from '@/components/payment/currency'
import {
  hasWechatResumeQuery,
  parseWechatResumeRoute,
  stripWechatResumeQuery,
  type ParsedWechatResumeRoute,
} from './paymentWechatResume'

type VisibleLocalPaymentMethod = 'alipay' | 'wxpay'

type TextPair = {
  label: string
  value: string
}

type CashierProduct = {
  id: string
  kind: OrderType
  name: string
  description: string
  priceText: string
  originalPriceText?: string
  period: string
  badge?: string
  highlight?: boolean
  metrics: TextPair[]
  amount: number
  arrivalText: string
  planId?: number
  disabled?: boolean
}

type PriceGroup = {
  kind: 'balance' | 'service'
  title: string
  description: string
  plans: CashierProduct[]
}

type ProductTabKind = PriceGroup['kind']

type WechatBridgeResponse = {
  err_msg?: string
}

type WechatBridge = {
  invoke: (
    action: string,
    payload: Record<string, unknown>,
    callback: (response: WechatBridgeResponse) => void,
  ) => void
}

const i18n = useI18n()
const t = i18n.t
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const paymentStore = usePaymentStore()

const checkoutInfo = ref<CheckoutInfoResponse | null>(null)
const affiliateDetail = ref<UserAffiliateDetail | null>(null)
const loadingCheckout = ref(false)
const submitting = ref(false)
const selectedProduct = ref<CashierProduct | null>(null)
const selectedPaymentMethod = ref<VisibleLocalPaymentMethod | ''>('')
const activeProductKind = ref<ProductTabKind>('balance')
const confirmPanelRef = ref<HTMLElement | null>(null)
const useAffiliateDiscount = ref(true)

const user = computed(() => authStore.user)
const localeValue = computed(() => {
  const raw = (i18n as { locale?: { value?: string } }).locale?.value
  return typeof raw === 'string' ? raw : 'zh'
})
const isZh = computed(() => localeValue.value.toLowerCase().startsWith('zh'))
const copy = computed(() => isZh.value ? zhCopy : enCopy)
const homePricingConfig = computed(() => appStore.cachedPublicSettings?.home_pricing_config || null)

const userBalance = computed(() => {
  const balance = Number(user.value?.balance || 0)
  return formatCreditAmount(balance)
})

const affiliateDiscountSettings = computed(() => checkoutInfo.value?.affiliate_discount)

const affiliateAvailableQuota = computed(() => {
  const quota = Number(affiliateDetail.value?.aff_quota || 0)
  return Number.isFinite(quota) && quota > 0 ? quota : 0
})

const affiliateDiscountVisible = computed(() => {
  return affiliateDiscountSettings.value?.enabled === true && affiliateAvailableQuota.value > 0
})

const selectedAffiliateDiscount = computed(() => {
  if (!useAffiliateDiscount.value) return 0
  return selectedAffiliateDiscountPreview.value
})

const selectedAffiliateDiscountPreview = computed(() => {
  if (!selectedProduct.value || !affiliateDiscountVisible.value) return 0
  return estimateAffiliateDiscount(paymentBaseAmountForProduct(selectedProduct.value))
})

const affiliateDiscountSelectable = computed(() => {
  return affiliateDiscountVisible.value && selectedAffiliateDiscountPreview.value > 0
})

const selectedOriginalAmountText = computed(() => {
  if (!selectedProduct.value) return formatCnyPrice(0)
  return formatCnyPrice(paymentBaseAmountForProduct(selectedProduct.value))
})

const selectedDiscountedBaseAmount = computed(() => {
  if (!selectedProduct.value) return 0
  return roundPaymentAmount(Math.max(0, paymentBaseAmountForProduct(selectedProduct.value) - selectedAffiliateDiscount.value))
})

const selectedFeeAmount = computed(() => {
  const rate = Number(checkoutInfo.value?.recharge_fee_rate || 0)
  if (!Number.isFinite(rate) || rate <= 0) return 0
  return ceilPaymentAmount(selectedDiscountedBaseAmount.value * rate / 100)
})

const selectedPayableAmount = computed(() => {
  return roundPaymentAmount(selectedDiscountedBaseAmount.value + selectedFeeAmount.value)
})

const selectedPayableText = computed(() => formatCnyPrice(selectedPayableAmount.value))

const selectedAffiliateDiscountText = computed(() => {
  return selectedAffiliateDiscount.value > 0 ? `-${formatCnyPrice(selectedAffiliateDiscount.value)}` : formatCnyPrice(0)
})

const selectedAffiliateDiscountStatus = computed(() => {
  if (!affiliateDiscountSelectable.value) return copy.value.discountUnavailable
  if (!useAffiliateDiscount.value) return copy.value.discountNotApplied
  return copy.value.discountApplied
})

const pricingHeader = computed(() => {
  return {
    eyebrow: copy.value.pricingEyebrow,
    title: copy.value.pricingTitle,
    description: copy.value.pricingDescription,
  }
})

const availablePaymentMethods = computed<VisibleLocalPaymentMethod[]>(() => {
  const methods = getVisibleMethods(checkoutInfo.value?.methods || {})
  return (['alipay', 'wxpay'] as VisibleLocalPaymentMethod[])
    .filter(method => methods[method]?.available)
})

const visibleMethodLimits = computed<Record<string, MethodLimit>>(() => {
  return getVisibleMethods(checkoutInfo.value?.methods || {})
})

const selectedPaymentCurrency = computed(() => {
  return normalizePaymentCurrency(visibleMethodLimits.value[selectedPaymentMethod.value]?.currency)
})

const renderedPricingGroups = computed<PriceGroup[]>(() => {
  const groups: PriceGroup[] = []
  const balancePlans = buildBalanceProducts()
  const servicePlans = buildServiceProducts()

  if (balancePlans.length) {
    groups.push({
      kind: 'balance',
      title: pickHomePricingText(homePricingConfig.value?.credit_group?.title) || copy.value.balanceGroupTitle,
      description: pickHomePricingText(homePricingConfig.value?.credit_group?.description) || copy.value.balanceGroupDescription,
      plans: balancePlans,
    })
  }
  if (servicePlans.length) {
    groups.push({
      kind: 'service',
      title: pickHomePricingText(homePricingConfig.value?.subscription_group?.title) || copy.value.serviceGroupTitle,
      description: pickHomePricingText(homePricingConfig.value?.subscription_group?.description) || copy.value.serviceGroupDescription,
      plans: servicePlans,
    })
  }

  return groups
})

const availablePricingTabs = computed(() => renderedPricingGroups.value.map(group => ({
  kind: group.kind,
  title: group.title,
})))

const visiblePricingGroups = computed<PriceGroup[]>(() => {
  const groups = renderedPricingGroups.value
  if (!groups.length) return []
  const active = groups.find(group => group.kind === activeProductKind.value)
  return active ? [active] : [groups[0]]
})

const currentProductKind = computed<ProductTabKind>(() => {
  return visiblePricingGroups.value[0]?.kind || activeProductKind.value
})

const canSubmitSelectedProduct = computed(() => {
  if (!selectedProduct.value || !selectedPaymentMethod.value) return false
  return methodIsUsableForProduct(selectedPaymentMethod.value, selectedProduct.value)
})

const selectedMethodLimitMessage = computed(() => {
  if (!selectedProduct.value || !selectedPaymentMethod.value) return ''
  const limit = visibleMethodLimits.value[selectedPaymentMethod.value]
  if (!limit) return copy.value.noPaymentMethod
  if (!limit.available) return copy.value.methodUnavailable
  const amount = selectedPayableAmount.value
  if (limit.single_min > 0 && amount < limit.single_min) {
    return copy.value.amountTooLow.replace('{amount}', formatCnyPrice(limit.single_min))
  }
  if (limit.single_max > 0 && amount > limit.single_max) {
    return copy.value.amountTooHigh.replace('{amount}', formatCnyPrice(limit.single_max))
  }
  return ''
})

watch(availablePaymentMethods, (methods) => {
  if (!methods.length) {
    selectedPaymentMethod.value = ''
    return
  }
  if (!selectedPaymentMethod.value || !methods.includes(selectedPaymentMethod.value)) {
    selectedPaymentMethod.value = methods[0]
  }
}, { immediate: true })

onMounted(async () => {
  await Promise.resolve(appStore.fetchPublicSettings?.(true)).catch(() => {})
  Promise.resolve(authStore.refreshUser()).catch(() => {})
  await loadCheckoutInfo()
  await loadAffiliateDetail()
  applyPurchaseQueryDefaults()
  await resumeWechatPaymentIfNeeded()
})

async function loadCheckoutInfo() {
  loadingCheckout.value = true
  try {
    const response = await paymentAPI.getCheckoutInfo()
    checkoutInfo.value = {
      ...response.data,
      plans: normalizePlans(response.data.plans || []),
    }
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    loadingCheckout.value = false
  }
}

async function loadAffiliateDetail() {
  if (checkoutInfo.value?.affiliate_discount?.enabled !== true) return
  try {
    affiliateDetail.value = await userAPI.getAffiliateDetail()
  } catch {
    affiliateDetail.value = null
  }
}

function normalizePlans(plans: SubscriptionPlan[]): SubscriptionPlan[] {
  return plans
    .map(plan => ({
      ...plan,
      features: Array.isArray(plan.features) ? plan.features : [],
    }))
    .filter(plan => plan.for_sale !== false)
    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
}

function buildBalanceProducts(): CashierProduct[] {
  if (checkoutInfo.value?.balance_disabled) return []
  const cards = homePricingConfig.value?.credit_cards || []
  const configured = cards
    .filter(card => card.enabled)
    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
    .map(card => {
      const amount = Number(card.price || card.recharge_amount || 0)
      const credited = getCreditedBalanceAmount(card.credited_amount, amount)
      return {
        id: `balance-${card.id}`,
        kind: 'balance' as const,
        name: normalizeCreditText(pickHomePricingText(card.name)),
        description: normalizeCreditText(pickHomePricingText(card.description)),
        priceText: formatCnyPrice(amount),
        originalPriceText: '',
        period: pickHomePricingText(card.period),
        badge: pickHomePricingOptionalText(card.badge),
        highlight: card.highlight,
        metrics: buildCreditProductMetrics(card.metrics, credited),
        amount,
        arrivalText: formatCreditAmount(credited),
        disabled: amount <= 0,
      }
    })
    .filter(plan => plan.amount > 0)

  return configured
}

function buildServiceProducts(): CashierProduct[] {
  const plans = checkoutInfo.value?.plans || []
  const planById = new Map(plans.map(plan => [plan.id, plan]))
  const cards = homePricingConfig.value?.subscription_cards || []

  const configured = cards
    .filter(card => card.enabled && card.for_sale !== false)
    .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
    .map((card): CashierProduct | null => {
      const plan = planById.get(card.subscription_plan_id)
      if (!plan) return null
      const amount = Number(plan.price || card.price || 0)
      const originalAmount = Number(card.original_price || plan.original_price || 0)
      const paymentAmount = subscriptionPaymentBaseAmount(amount)
      const originalPaymentAmount = originalAmount > amount
        ? subscriptionPaymentBaseAmount(originalAmount)
        : 0
      return {
        id: `service-${card.id}`,
        kind: 'subscription' as const,
        name: pickHomePricingText(card.name) || plan.name,
        description: pickHomePricingText(card.description) || plan.description,
        priceText: formatCnyPrice(paymentAmount),
        originalPriceText: originalPaymentAmount > paymentAmount ? formatCnyPrice(originalPaymentAmount) : '',
        period: pickHomePricingText(card.period) || formatPlanPeriod(plan),
        badge: pickHomePricingOptionalText(card.badge),
        highlight: card.highlight,
        metrics: normalizeMetrics(card.metrics, buildPlanMetrics(plan)),
        amount,
        arrivalText: buildPlanArrivalText(plan),
        planId: plan.id,
        disabled: amount <= 0,
      }
    })
    .filter((plan): plan is CashierProduct => Boolean(plan))

  return configured
}

function normalizeMetrics(metrics: { label: HomePricingLocalizedText; value: HomePricingLocalizedText }[] | undefined, fallback: TextPair[]) {
  const rows = (metrics || [])
    .map(metric => ({
      label: pickHomePricingText(metric.label),
      value: pickHomePricingText(metric.value),
    }))
    .filter(row => row.label && row.value)
  return rows.length ? rows : fallback
}

function normalizeCreditMetrics(metrics: { label: HomePricingLocalizedText; value: HomePricingLocalizedText }[] | undefined, fallback: TextPair[]) {
  return normalizeMetrics(metrics, fallback).map(row => ({
    label: row.label,
    value: normalizeCreditText(row.value),
  }))
}

function buildCreditProductMetrics(metrics: { label: HomePricingLocalizedText; value: HomePricingLocalizedText }[] | undefined, credited: number) {
  const systemMetrics = [
    { label: copy.value.creditMetric, value: formatCreditAmount(credited) },
    { label: copy.value.typeMetric, value: copy.value.balanceCredit },
  ]
  const customMetrics = normalizeCreditMetrics(metrics, [])
    .filter(row => !isCreditSystemMetric(row.label))
  return [...systemMetrics, ...customMetrics]
}

function isCreditSystemMetric(label: string) {
  const normalized = label.trim().toLowerCase()
  return [
    copy.value.creditMetric,
    copy.value.typeMetric,
    '额度',
    '到账积分',
    '类型',
    'credit',
    'credits',
    'credits received',
    'type',
  ].some(item => normalized === item.trim().toLowerCase())
}

function getCreditedBalanceAmount(configuredCredit: number | undefined, paymentAmount: number) {
  const configured = Number(configuredCredit || 0)
  if (Number.isFinite(configured) && configured > 0) return configured
  return calculateBalanceCreditAmount(paymentAmount)
}

function calculateBalanceCreditAmount(paymentAmount: number) {
  const multiplier = Number(checkoutInfo.value?.balance_recharge_multiplier || 1)
  const safeMultiplier = Number.isFinite(multiplier) && multiplier > 0 ? multiplier : 1
  return Math.round(paymentAmount * safeMultiplier * 100) / 100
}

function currencyFractionDigits(currency = selectedPaymentCurrency.value) {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: normalizePaymentCurrency(currency),
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function roundPaymentAmount(value: number, currency = selectedPaymentCurrency.value) {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.round(value * factor) / factor
}

function ceilPaymentAmount(value: number, currency = selectedPaymentCurrency.value) {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.ceil(value * factor) / factor
}

function subscriptionPaymentBaseAmount(amount: number, currency = selectedPaymentCurrency.value) {
  const normalizedCurrency = normalizePaymentCurrency(currency)
  const safeAmount = Number.isFinite(amount) && amount > 0 ? amount : 0
  return roundPaymentAmount(safeAmount, normalizedCurrency)
}

function paymentBaseAmountForProduct(product: CashierProduct) {
  if (product.kind === 'subscription') return subscriptionPaymentBaseAmount(product.amount)
  return roundPaymentAmount(product.amount)
}

function estimateAffiliateDiscount(amount: number) {
  const settings = affiliateDiscountSettings.value
  if (!settings?.enabled || affiliateAvailableQuota.value <= 0 || amount <= 0) return 0
  const maxPercent = Number.isFinite(settings.max_percent) ? Math.max(0, Math.min(100, settings.max_percent)) : 50
  const minPayAmount = Number.isFinite(settings.min_pay_amount) ? Math.max(0, settings.min_pay_amount) : 1
  const maxByPercent = amount * maxPercent / 100
  const maxByMinPay = Math.max(0, amount - minPayAmount)
  return roundMoney(Math.min(affiliateAvailableQuota.value, maxByPercent, maxByMinPay))
}

function effectivePaymentAmountForProduct(product: CashierProduct) {
  if (selectedProduct.value?.id !== product.id) return paymentBaseAmountForProduct(product)
  return selectedPayableAmount.value
}

function buildPlanMetrics(plan: SubscriptionPlan): TextPair[] {
  const metrics: TextPair[] = []
  if (plan.daily_limit_usd) {
    metrics.push({ label: copy.value.dailyCredit, value: formatCreditAmount(plan.daily_limit_usd) })
  }
  metrics.push({ label: copy.value.validity, value: formatPlanValidity(plan) })
  if (plan.group_name) {
    metrics.push({ label: copy.value.serviceScope, value: plan.group_name })
  }
  return metrics
}

function buildPlanArrivalText(plan: SubscriptionPlan): string {
  const daily = plan.daily_limit_usd ? `${formatCreditAmount(plan.daily_limit_usd)} / ${copy.value.day}` : ''
  const validity = formatPlanValidity(plan)
  return daily ? `${daily}, ${validity}` : validity
}

function formatPlanPeriod(plan: SubscriptionPlan): string {
  const validity = formatPlanValidity(plan)
  return validity ? `/ ${validity}` : ''
}

function formatPlanValidity(plan: SubscriptionPlan): string {
  if (!plan.validity_days) return ''
  if (isZh.value) return `${plan.validity_days} 天`
  return `${plan.validity_days} days`
}

function selectProduct(product: CashierProduct, options: { scroll?: boolean } = {}) {
  selectedProduct.value = product
  activeProductKind.value = product.kind === 'subscription' ? 'service' : 'balance'
  if (!selectedPaymentMethod.value && availablePaymentMethods.value.length) {
    selectedPaymentMethod.value = availablePaymentMethods.value[0]
  }
  if (options.scroll !== false) {
    nextTick(() => {
      const scrollIntoView = confirmPanelRef.value?.scrollIntoView
      if (typeof scrollIntoView === 'function') {
        scrollIntoView.call(confirmPanelRef.value, { behavior: 'smooth', block: 'start' })
      }
    })
  }
}

function applyPurchaseQueryDefaults() {
  const amount = Number(route.query.amount || 0)
  if (Number.isFinite(amount) && amount > 0 && !checkoutInfo.value?.balance_disabled) {
    const product = renderedPricingGroups.value
      .find(group => group.kind === 'balance')
      ?.plans.find(item => Math.abs(item.amount - amount) < 0.001)
    if (product) selectProduct(product, { scroll: false })
  }

  const planId = Number(route.query.plan_id || 0)
  if (Number.isFinite(planId) && planId > 0) {
    const plan = renderedPricingGroups.value
      .flatMap(group => group.plans)
      .find(item => item.planId === planId)
    if (plan) selectProduct(plan, { scroll: false })
  }
}

async function submitSelectedOrder() {
  if (!selectedProduct.value || !selectedPaymentMethod.value || !canSubmitSelectedProduct.value) return
  await createAndLaunchOrder({
    amount: selectedProduct.value.amount,
    orderType: selectedProduct.value.kind,
    paymentType: selectedPaymentMethod.value,
    planId: selectedProduct.value.planId,
    useAffiliateDiscount: affiliateDiscountSelectable.value && useAffiliateDiscount.value,
  })
}

async function createAndLaunchOrder(input: {
  amount: number
  orderType: OrderType
  paymentType: string
  planId?: number
  openid?: string
  wechatResumeToken?: string
  useAffiliateDiscount?: boolean
}) {
  const mobile = isMobileDevice()
  const inWechat = isWechatBrowser()
  const basePayload = buildCreateOrderPayload({
    amount: input.amount,
    paymentType: input.paymentType,
    orderType: input.orderType,
    planId: input.planId,
    origin: window.location.origin,
    isMobile: mobile,
    isWechatBrowser: inWechat,
    useAffiliateDiscount: input.useAffiliateDiscount,
  })
  const payload: CreateOrderRequest = {
    ...basePayload,
    ...(input.openid ? { openid: input.openid } : {}),
    ...(input.wechatResumeToken ? { wechat_resume_token: input.wechatResumeToken } : {}),
    use_affiliate_discount: input.useAffiliateDiscount !== false,
  }

  submitting.value = true
  try {
    const result = await paymentStore.createOrder(payload)
    await handleOrderResult(result, {
      paymentType: input.paymentType,
      orderType: input.orderType,
      isMobile: mobile,
      isWechatBrowser: inWechat,
      planId: input.planId,
      useAffiliateDiscount: input.useAffiliateDiscount,
    })
  } catch (error) {
    const recovered = await retryWithQrFallbackIfNeeded(error, payload, {
      paymentType: input.paymentType,
      orderType: input.orderType,
      planId: input.planId,
      useAffiliateDiscount: input.useAffiliateDiscount,
    })
    if (!recovered) {
      showPaymentError(error, input.paymentType, mobile, inWechat)
      clearPaymentRecoverySnapshot(window.localStorage)
    }
  } finally {
    submitting.value = false
  }
}

async function retryWithQrFallbackIfNeeded(
  error: unknown,
  payload: CreateOrderRequest,
  context: { paymentType: string; orderType: OrderType; planId?: number; useAffiliateDiscount?: boolean },
): Promise<boolean> {
  if (normalizeVisibleMethod(payload.payment_type) !== 'wxpay') return false
  if (extractApiErrorCode(error) !== 'WECHAT_H5_NOT_AUTHORIZED') return false

  appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
  const retryPayload: CreateOrderRequest = {
    ...payload,
    is_mobile: false,
    payment_source: 'hosted_redirect',
  }
  const result = await paymentStore.createOrder(retryPayload)
  await handleOrderResult(result, {
    paymentType: context.paymentType,
    orderType: context.orderType,
    isMobile: false,
    isWechatBrowser: false,
    planId: context.planId,
    useAffiliateDiscount: context.useAffiliateDiscount,
  })
  return true
}

async function handleOrderResult(
  result: CreateOrderResult,
  context: {
    paymentType: string
    orderType: OrderType
    isMobile: boolean
    isWechatBrowser: boolean
    planId?: number
    useAffiliateDiscount?: boolean
  },
) {
  const visibleMethod = normalizeVisibleMethod(result.payment_type || context.paymentType) || context.paymentType
  const decision = decidePaymentLaunch(result, {
    visibleMethod,
    orderType: context.orderType,
    isMobile: context.isMobile,
    isWechatBrowser: context.isWechatBrowser,
  })

  if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
    writePaymentRecoverySnapshot(window.localStorage, decision.recovery)
    window.location.href = buildWechatAuthorizeUrl(decision.oauth.authorize_url, {
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      useAffiliateDiscount: context.useAffiliateDiscount,
    })
    return
  }

  if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
    writePaymentRecoverySnapshot(window.localStorage, decision.recovery)
    await launchWechatJsapiPayment(decision.jsapi as Record<string, unknown>, decision.recovery)
    return
  }

  if (decision.kind === 'qr_waiting' || decision.kind === 'redirect_waiting') {
    writePaymentRecoverySnapshot(window.localStorage, decision.recovery)
    if (decision.kind === 'redirect_waiting' && context.isMobile && decision.paymentState.payUrl) {
      window.location.href = decision.paymentState.payUrl
      return
    }
    await router.push({
      path: '/payment/qrcode',
      query: buildQrRouteQuery(decision.paymentState),
    })
    return
  }

  clearPaymentRecoverySnapshot(window.localStorage)
  appStore.showError(t('payment.errors.PAYMENT_GATEWAY_ERROR'))
}

async function launchWechatJsapiPayment(payload: Record<string, unknown>, recovery: PaymentRecoverySnapshot) {
  const bridge = await waitForWechatBridge()
  if (!bridge) {
    showPaymentError(new Error('WECHAT_JSAPI_UNAVAILABLE'), 'wxpay', true, true)
    clearPaymentRecoverySnapshot(window.localStorage)
    return
  }

  await new Promise<void>((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (response) => {
      const message = response.err_msg || ''
      if (message.includes(':ok')) {
        clearPaymentRecoverySnapshot(window.localStorage)
        router.push({
          path: '/payment/result',
          query: {
            order_id: String(recovery.orderId),
            out_trade_no: recovery.outTradeNo,
            resume_token: recovery.resumeToken,
          },
        })
      } else if (message.includes(':cancel')) {
        clearPaymentRecoverySnapshot(window.localStorage)
        appStore.showInfo(t('payment.qr.cancelled'))
      } else {
        clearPaymentRecoverySnapshot(window.localStorage)
        showPaymentError(new Error(message || 'WECHAT_JSAPI_FAILED'), 'wxpay', true, true)
      }
      resolve()
    })
  })
}

async function waitForWechatBridge(timeoutMs = 3000): Promise<WechatBridge | null> {
  const existing = (window as Window & { WeixinJSBridge?: WechatBridge }).WeixinJSBridge
  if (existing) return existing

  return new Promise(resolve => {
    let resolved = false
    const done = (bridge: WechatBridge | null) => {
      if (resolved) return
      resolved = true
      document.removeEventListener('WeixinJSBridgeReady', onReady)
      resolve(bridge)
    }
    const onReady = () => {
      done((window as Window & { WeixinJSBridge?: WechatBridge }).WeixinJSBridge || null)
    }
    document.addEventListener('WeixinJSBridgeReady', onReady)
    window.setTimeout(() => done(null), timeoutMs)
  })
}

async function resumeWechatPaymentIfNeeded() {
  if (!hasWechatResumeQuery(route.query)) return
  clearPaymentRecoverySnapshot(window.localStorage)
  const fallbackAmount = renderedPricingGroups.value
    .find(group => group.kind === 'balance')
    ?.plans[0]?.amount || 0
  const parsed = parseWechatResumeRoute(route.query, checkoutInfo.value?.plans || [], fallbackAmount)
  await router.replace({ path: '/purchase', query: stripWechatResumeQuery(route.query) })
  if (!parsed) return
  await resumeWechatPayment(parsed)
}

async function resumeWechatPayment(parsed: ParsedWechatResumeRoute) {
  const plan = parsed.planId
    ? checkoutInfo.value?.plans.find(item => item.id === parsed.planId)
    : undefined
  const amount = parsed.orderAmount > 0
    ? parsed.orderAmount
    : parsed.orderType === 'subscription'
      ? Number(plan?.price || 0)
      : (renderedPricingGroups.value.find(group => group.kind === 'balance')?.plans[0]?.amount || 0)
  await createAndLaunchOrder({
    amount,
    orderType: parsed.orderType,
    paymentType: parsed.paymentType,
    planId: parsed.planId,
    openid: parsed.openid,
    wechatResumeToken: parsed.wechatResumeToken,
    useAffiliateDiscount: parsed.useAffiliateDiscount,
  })
}

function buildQrRouteQuery(state: PaymentRecoverySnapshot) {
  const query: Record<string, string> = {
    order_id: String(state.orderId),
    payment_type: state.paymentType,
  }
  if (state.qrCode) query.qr = state.qrCode
  if (state.payUrl) query.pay_url = state.payUrl
  if (state.expiresAt) query.expires_at = state.expiresAt
  return query
}

function buildWechatAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; useAffiliateDiscount?: boolean },
): string {
  const url = new URL(authorizeUrl, window.location.origin)
  const redirect = new URLSearchParams()
  redirect.set('from', 'wechat')
  redirect.set('payment_type', normalizeVisibleMethod(context.paymentType) || context.paymentType)
  redirect.set('order_type', context.orderType)
  if (context.planId) redirect.set('plan_id', String(context.planId))
  if (typeof context.useAffiliateDiscount === 'boolean') {
    redirect.set('use_affiliate_discount', String(context.useAffiliateDiscount))
  }
  url.searchParams.set('redirect', `/purchase?${redirect.toString()}`)
  return url.toString()
}

function showPaymentError(error: unknown, paymentType: string, mobile: boolean, inWechat: boolean) {
  const scenario = describePaymentScenarioError(error, {
    paymentMethod: paymentType,
    isMobile: mobile,
    isWechatBrowser: inWechat,
  })
  if (scenario) {
    appStore.showError(buildPaymentErrorToastMessage(t(scenario.messageKey), scenario.hintKey ? t(scenario.hintKey) : undefined))
    return
  }
  appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
}

function methodIsUsableForProduct(method: VisibleLocalPaymentMethod, product: CashierProduct) {
  return methodIsUsableForAmount(method, effectivePaymentAmountForProduct(product))
}

function methodIsUsableForAmount(method: VisibleLocalPaymentMethod, amount: number) {
  const limit = visibleMethodLimits.value[method]
  if (!limit?.available) return false
  const globalMin = Number(checkoutInfo.value?.global_min || 0)
  const globalMax = Number(checkoutInfo.value?.global_max || 0)
  if (globalMin > 0 && amount < globalMin) return false
  if (globalMax > 0 && amount > globalMax) return false
  if (limit.single_min > 0 && amount < limit.single_min) return false
  if (limit.single_max > 0 && amount > limit.single_max) return false
  return true
}

function paymentMethodLabel(method: VisibleLocalPaymentMethod) {
  return method === 'alipay' ? copy.value.alipay : copy.value.wxpay
}

function pickHomePricingText(value?: HomePricingLocalizedText | null) {
  if (!value) return ''
  const preferred = isZh.value ? value.zh : value.en
  const fallback = isZh.value ? value.en : value.zh
  return preferred?.trim() || fallback?.trim() || ''
}

function pickHomePricingOptionalText(value?: HomePricingLocalizedText | null) {
  const text = pickHomePricingText(value)
  return text || undefined
}

function formatCnyPrice(value: number) {
  return formatPaymentAmount(roundPaymentAmount(Number.isFinite(value) && value > 0 ? value : 0), selectedPaymentCurrency.value, localeValue.value)
}

function roundMoney(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.round(value * 100) / 100
}

function formatCreditAmount(value: number) {
  return formatPoints(Number.isFinite(value) && value > 0 ? value : 0)
}

function normalizeCreditText(text: string) {
  if (!text) return text
  return text.replace(/\$([\d,]+(?:\.\d+)?)/g, (_, raw: string) => {
    const amount = Number(raw.replace(/,/g, ''))
    return formatCreditAmount(amount)
  })
}

function isWechatBrowser() {
  if (typeof navigator === 'undefined') return false
  return /micromessenger/i.test(navigator.userAgent || '')
}

const zhCopy = {
  eyebrow: '站内收银台',
  title: '充值额度，支付后自动到账。',
  description: '选择充值额度或订阅后，直接使用支付宝或微信支付。订单创建、扫码、到账状态都在站内完成。',
  account: '当前账号',
  balance: '当前余额',
  ordersLink: '查看历史订单',
  cashierTitle: '支付宝 / 微信支付',
  cashierDescription: '请选择可用的支付方式完成站内充值或订阅购买。',
  noPaymentMethod: '支付宝/微信暂不可用',
  loading: '正在读取收银台配置...',
  pricingEyebrow: '充值',
  pricingTitle: '选择要到账的额度类型',
  pricingDescription: '充值额度直接进入余额；订阅按套餐规则生效，支付和到账都在当前页面完成。',
  balanceGroupTitle: '充值额度',
  balanceGroupDescription: '适合灵活补充，额度进入余额后按实际使用消耗。',
  serviceGroupTitle: '订阅',
  serviceGroupDescription: '适合持续使用，按服务套餐提供周期内额度。',
  noServicePlans: '暂无可购买的订阅套餐，请管理员在后台配置并上架套餐。',
  rechargeAction: '立即充值',
  buyAction: '立即购买',
  confirmEyebrow: '确认支付',
  productAmount: '商品金额',
  orderAmount: '订单金额',
  payAmount: '实付金额',
  arrival: '到账内容',
  fee: '手续费',
  useAffiliateDiscount: '使用返利折扣余额',
  availableDiscount: '可用返利',
  discountApplied: '返利抵扣',
  discountNotApplied: '未使用返利抵扣',
  discountUnavailable: '本单暂无可用抵扣',
  confirmPay: '确认并支付',
  creatingOrder: '正在创建订单...',
  methodUnavailable: '该支付方式暂不可用',
  amountTooLow: '该支付方式最低金额为 {amount}',
  amountTooHigh: '该支付方式最高金额为 {amount}',
  alipay: '支付宝',
  wxpay: '微信支付',
  creditMetric: '额度',
  typeMetric: '类型',
  balanceCredit: '余额额度',
  dailyCredit: '每日额度',
  validity: '有效期',
  serviceScope: '服务范围',
  serviceProductDescription: '订阅套餐',
  day: '天',
}

const enCopy = {
  eyebrow: 'Internal Checkout',
  title: 'Top up credits and get them applied automatically.',
  description: 'Choose balance credit or subscription, then pay with Alipay or WeChat Pay. Orders, QR codes, and payment status stay inside the app.',
  account: 'Current account',
  balance: 'Current balance',
  ordersLink: 'Order history',
  cashierTitle: 'Alipay / WeChat Pay',
  cashierDescription: 'Choose an available payment method to complete your top-up or subscription purchase.',
  noPaymentMethod: 'Alipay/WeChat unavailable',
  loading: 'Loading checkout configuration...',
  pricingEyebrow: 'Checkout',
  pricingTitle: 'Choose the credit type to receive',
  pricingDescription: 'Balance credit goes directly to your account. Subscriptions follow package rules, with payment and status handled here.',
  balanceGroupTitle: 'Balance Credit',
  balanceGroupDescription: 'Flexible top-ups for account balance, consumed by actual usage.',
  serviceGroupTitle: 'Subscription',
  serviceGroupDescription: 'For steady usage, with package rules applying during the validity period.',
  noServicePlans: 'No subscription packages are currently for sale. Configure and publish packages in admin.',
  rechargeAction: 'Top Up Now',
  buyAction: 'Buy Now',
  confirmEyebrow: 'Confirm Payment',
  productAmount: 'Item Amount',
  orderAmount: 'Order Amount',
  payAmount: 'Amount Due',
  arrival: 'Credit Received',
  fee: 'Fee',
  useAffiliateDiscount: 'Use referral discount credit',
  availableDiscount: 'Available rebate',
  discountApplied: 'Referral discount',
  discountNotApplied: 'Referral discount not used',
  discountUnavailable: 'No discount available for this order',
  confirmPay: 'Confirm and Pay',
  creatingOrder: 'Creating order...',
  methodUnavailable: 'This payment method is unavailable.',
  amountTooLow: 'Minimum amount for this method is {amount}.',
  amountTooHigh: 'Maximum amount for this method is {amount}.',
  alipay: 'Alipay',
  wxpay: 'WeChat Pay',
  creditMetric: 'Credit',
  typeMetric: 'Type',
  balanceCredit: 'Balance credit',
  dailyCredit: 'Daily credit',
  validity: 'Validity',
  serviceScope: 'Service scope',
  serviceProductDescription: 'Subscription package',
  day: 'day',
}

</script>

<style scoped>
.purchase-hero {
  display: grid;
  gap: 1rem;
  align-items: stretch;
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.75rem;
  background: linear-gradient(135deg, rgba(232, 246, 255, 0.92), rgba(255, 255, 255, 0.95));
  padding: 1.25rem;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.06);
}

@media (min-width: 768px) {
  .purchase-hero {
    grid-template-columns: minmax(0, 1fr) 18rem;
  }
}

.purchase-eyebrow,
.section-heading p {
  color: #006fd6;
  font-size: 0.78rem;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.purchase-hero h1 {
  margin-top: 0.5rem;
  color: #0f172a;
  font-size: clamp(1.75rem, 4vw, 3rem);
  font-weight: 950;
  line-height: 1.08;
}

.purchase-hero p:not(.purchase-eyebrow) {
  margin-top: 0.85rem;
  max-width: 46rem;
  color: #475569;
  font-size: 1rem;
  line-height: 1.8;
}

.account-panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  justify-content: center;
  border: 1px solid rgba(148, 163, 184, 0.28);
  border-radius: 0.65rem;
  background: rgba(255, 255, 255, 0.74);
  padding: 1rem;
}

.account-panel span,
.account-panel small {
  color: #64748b;
  font-size: 0.8rem;
  font-weight: 800;
}

.account-panel strong {
  margin-top: 0.4rem;
  overflow-wrap: anywhere;
  color: #0f172a;
  font-size: 1.05rem;
}

.account-panel small {
  margin-top: 0.45rem;
}

.orders-link {
  display: inline-flex;
  width: fit-content;
  margin-top: 0.85rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 111, 214, 0.2);
  border-radius: 0.5rem;
  background: rgba(232, 243, 255, 0.8);
  color: #0057b8;
  padding: 0.45rem 0.7rem;
  font-size: 0.82rem;
  font-weight: 900;
}

.notice-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border: 1px solid rgba(0, 111, 214, 0.18);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.9);
  padding: 1rem;
}

@media (min-width: 768px) {
  .notice-panel {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}

.notice-panel h2,
.confirm-panel h2 {
  color: #0f172a;
  font-size: 1rem;
  font-weight: 900;
}

.notice-panel p,
.confirm-panel p {
  margin-top: 0.35rem;
  color: #64748b;
  font-size: 0.9rem;
  line-height: 1.7;
}

.method-strip,
.method-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.method-pill,
.method-button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(0, 111, 214, 0.18);
  border-radius: 999px;
  background: #f8fafc;
  color: #0f172a;
  padding: 0.4rem 0.8rem;
  font-size: 0.85rem;
  font-weight: 900;
}

.method-pill-muted {
  border-color: rgba(148, 163, 184, 0.35);
  color: #64748b;
}

.method-button {
  border-radius: 0.55rem;
  cursor: pointer;
}

.method-button-active {
  border-color: #006fd6;
  background: #e8f3ff;
  color: #0057b8;
}

.method-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.pricing-section {
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.78);
  padding: 1.25rem;
}

.section-heading h2 {
  margin-top: 0.35rem;
  color: #0f172a;
  font-size: clamp(1.35rem, 3vw, 2rem);
  font-weight: 950;
  line-height: 1.16;
}

.section-heading span {
  display: block;
  margin-top: 0.55rem;
  color: #64748b;
  font-size: 0.95rem;
  line-height: 1.7;
}

.loading-panel {
  margin-top: 1.25rem;
  border: 1px dashed rgba(148, 163, 184, 0.55);
  border-radius: 0.65rem;
  color: #64748b;
  padding: 1rem;
  text-align: center;
}

.empty-service-panel {
  margin-top: 0.9rem;
  border: 1px dashed rgba(148, 163, 184, 0.55);
  border-radius: 0.65rem;
  color: #64748b;
  padding: 1rem;
}

.pricing-tabs {
  display: inline-flex;
  gap: 0.35rem;
  margin-top: 1.15rem;
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.65rem;
  background: rgba(248, 250, 252, 0.88);
  padding: 0.3rem;
}

.pricing-tab {
  min-height: 2.35rem;
  border-radius: 0.45rem;
  color: #64748b;
  padding: 0.45rem 0.85rem;
  font-size: 0.86rem;
  font-weight: 900;
  transition:
    background-color 0.16s ease,
    color 0.16s ease;
}

.pricing-tab-active {
  background: #001040;
  color: #ffffff;
}

.pricing-group {
  margin-top: 1.5rem;
}

.pricing-group-heading h3 {
  color: #0f172a;
  font-size: 1.05rem;
  font-weight: 950;
}

.pricing-group-heading p {
  margin-top: 0.3rem;
  color: #64748b;
  font-size: 0.9rem;
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 16rem), 1fr));
  gap: 1rem;
  margin-top: 0.9rem;
}

.pricing-card {
  display: flex;
  min-height: 22rem;
  flex-direction: column;
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.65rem;
  background: white;
  padding: 1rem;
}

.pricing-card-highlight {
  border-color: rgba(0, 111, 214, 0.48);
  box-shadow: 0 18px 36px rgba(0, 111, 214, 0.1);
}

.plan-badge {
  flex: 0 0 auto;
  border-radius: 999px;
  background: #e8f3ff;
  color: #0057b8;
  padding: 0.3rem 0.55rem;
  font-size: 0.75rem;
  font-weight: 900;
}

.original-price {
  margin-left: 0.5rem;
  color: #94a3b8;
  font-size: 1rem;
  font-weight: 800;
  text-decoration: line-through;
}

.plan-metrics {
  margin: 1rem 0;
  display: grid;
  gap: 0.55rem;
}

.plan-metric-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(226, 232, 240, 0.85);
  padding-bottom: 0.55rem;
}

.plan-metric-row dt {
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 800;
}

.plan-metric-row dd {
  overflow-wrap: anywhere;
  text-align: right;
  color: #0f172a;
  font-size: 0.86rem;
  font-weight: 900;
}

.buy-link,
.submit-button {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: #001040;
  color: white;
  font-size: 0.9rem;
  font-weight: 900;
  transition:
    background-color 0.16s ease,
    transform 0.16s ease;
}

.buy-link {
  width: 100%;
  margin-top: auto;
  padding: 0.65rem 0.9rem;
}

.buy-link:hover,
.submit-button:hover {
  background: #002080;
  transform: translateY(-1px);
}

.buy-link:disabled,
.submit-button:disabled {
  cursor: not-allowed;
  opacity: 0.48;
  transform: none;
}

.confirm-panel {
  scroll-margin-top: 5.5rem;
  border: 1px solid rgba(0, 111, 214, 0.24);
  border-radius: 0.75rem;
  background: #ffffff;
  padding: 1.25rem;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.08);
}

.confirm-grid {
  display: grid;
  gap: 0.75rem;
  margin: 1rem 0;
}

@media (min-width: 768px) {
  .confirm-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.confirm-detail {
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 0.6rem;
  padding: 0.85rem;
}

.confirm-detail span {
  color: #64748b;
  font-size: 0.8rem;
  font-weight: 800;
}

.confirm-detail strong {
  display: block;
  margin-top: 0.35rem;
  overflow-wrap: anywhere;
  color: #0f172a;
  font-size: 1rem;
  font-weight: 950;
}

.checkout-summary {
  display: grid;
  gap: 0.65rem;
  margin: -0.15rem 0 1rem;
  border: 1px solid rgba(226, 232, 240, 0.95);
  border-radius: 0.65rem;
  background: #f8fafc;
  padding: 0.85rem;
}

.checkout-summary-row {
  display: flex;
  min-height: 1.75rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  color: #475569;
  font-size: 0.86rem;
  font-weight: 800;
}

.checkout-summary-row strong {
  color: #0f172a;
  font-size: 0.95rem;
  font-weight: 950;
  text-align: right;
}

.checkout-summary-total {
  border-top: 1px solid rgba(203, 213, 225, 0.85);
  padding-top: 0.65rem;
  color: #0f172a;
}

.checkout-summary-total strong {
  color: #006fd6;
  font-size: 1.2rem;
}

.affiliate-discount-panel {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  border: 1px solid rgba(16, 185, 129, 0.28);
  border-radius: 0.6rem;
  background: rgba(236, 253, 245, 0.72);
  padding: 0.75rem;
}

.affiliate-discount-toggle {
  display: flex;
  align-items: center;
  gap: 0.55rem;
  color: #065f46;
  font-size: 0.9rem;
  font-weight: 850;
}

.affiliate-discount-toggle input {
  height: 1rem;
  width: 1rem;
  accent-color: #059669;
}

.affiliate-discount-toggle input:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.affiliate-discount-copy {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 0.12rem;
}

.affiliate-discount-copy strong {
  overflow-wrap: anywhere;
}

.affiliate-discount-copy small {
  color: #047857;
  font-size: 0.76rem;
  font-weight: 750;
}

.affiliate-discount-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  color: #047857;
  font-size: 0.8rem;
  font-weight: 750;
}

.affiliate-discount-summary strong {
  color: #047857;
  font-size: 0.95rem;
  font-weight: 950;
}

.method-warning {
  color: #b45309;
}

.submit-button {
  margin-top: 1rem;
  width: 100%;
  padding: 0.75rem 1rem;
}

:global(.dark .purchase-hero) {
  border-color: rgba(51, 65, 85, 0.8);
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.92));
  box-shadow: 0 18px 42px rgba(2, 6, 23, 0.32);
}

:global(.dark .purchase-eyebrow),
:global(.dark .section-heading p) {
  color: #38bdf8;
}

:global(.dark .purchase-hero h1),
:global(.dark .account-panel strong),
:global(.dark .notice-panel h2),
:global(.dark .confirm-panel h2),
:global(.dark .section-heading h2),
:global(.dark .pricing-group-heading h3),
:global(.dark .plan-metric-row dd),
:global(.dark .confirm-detail strong) {
  color: #f8fafc;
}

:global(.dark .purchase-hero p:not(.purchase-eyebrow)),
:global(.dark .notice-panel p),
:global(.dark .confirm-panel p),
:global(.dark .section-heading span),
:global(.dark .pricing-group-heading p),
:global(.dark .account-panel span),
:global(.dark .account-panel small),
:global(.dark .plan-metric-row dt),
:global(.dark .confirm-detail span),
:global(.dark .loading-panel) {
  color: #cbd5e1;
}

:global(.dark .account-panel),
:global(.dark .notice-panel),
:global(.dark .pricing-section),
:global(.dark .pricing-card),
:global(.dark .confirm-panel),
:global(.dark .confirm-detail),
:global(.dark .checkout-summary),
:global(.dark .method-pill),
:global(.dark .method-button) {
  border-color: rgba(71, 85, 105, 0.75);
  background: rgba(15, 23, 42, 0.86);
  color: #e2e8f0;
}

:global(.dark .orders-link) {
  border-color: rgba(56, 189, 248, 0.42);
  background: rgba(14, 116, 144, 0.24);
  color: #bae6fd;
}

:global(.dark .pricing-card-highlight),
:global(.dark .confirm-panel) {
  border-color: rgba(56, 189, 248, 0.46);
  box-shadow: 0 18px 42px rgba(2, 132, 199, 0.12);
}

:global(.dark .method-button-active),
:global(.dark .method-button-active:disabled) {
  border-color: rgba(56, 189, 248, 0.76);
  background: rgba(14, 116, 144, 0.28);
  color: #e0f2fe;
}

:global(.dark .affiliate-discount-panel) {
  border-color: rgba(16, 185, 129, 0.38);
  background: rgba(6, 78, 59, 0.24);
}

:global(.dark .affiliate-discount-toggle),
:global(.dark .affiliate-discount-summary),
:global(.dark .affiliate-discount-copy small),
:global(.dark .affiliate-discount-summary strong) {
  color: #a7f3d0;
}

:global(.dark .checkout-summary-row),
:global(.dark .checkout-summary-total) {
  color: #cbd5e1;
}

:global(.dark .checkout-summary-row strong) {
  color: #f8fafc;
}

:global(.dark .checkout-summary-total) {
  border-top-color: rgba(71, 85, 105, 0.85);
}

:global(.dark .checkout-summary-total strong) {
  color: #7dd3fc;
}

:global(.dark .pricing-tabs) {
  border-color: rgba(71, 85, 105, 0.75);
  background: rgba(2, 6, 23, 0.42);
}

:global(.dark .pricing-tab) {
  color: #cbd5e1;
}

:global(.dark .pricing-tab-active) {
  background: #0ea5e9;
  color: #082f49;
}

:global(.dark .method-pill-muted) {
  color: #94a3b8;
}

:global(.dark .plan-badge) {
  background: rgba(14, 116, 144, 0.28);
  color: #bae6fd;
}

:global(.dark .original-price) {
  color: #64748b;
}

:global(.dark .plan-metric-row) {
  border-bottom-color: rgba(51, 65, 85, 0.92);
}

:global(.dark .loading-panel) {
  border-color: rgba(71, 85, 105, 0.9);
  background: rgba(15, 23, 42, 0.5);
}

:global(.dark .empty-service-panel) {
  border-color: rgba(71, 85, 105, 0.9);
  background: rgba(15, 23, 42, 0.5);
  color: #cbd5e1;
}

:global(.dark .method-warning) {
  color: #fbbf24;
}

:global(.dark .buy-link),
:global(.dark .submit-button) {
  background: #0284c7;
  color: #f8fafc;
}

:global(.dark .buy-link:hover),
:global(.dark .submit-button:hover) {
  background: #0369a1;
}
</style>
