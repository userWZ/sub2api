<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-lg font-bold text-gray-900 dark:text-white">前台商品配置</h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">配置首页定价区和收银台商品卡片。顶部标题只用于 /home，站内收银台会使用自己的购买文案。</p>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-secondary" :disabled="loading" @click="loadConfig">刷新</button>
        <button class="btn btn-primary" :disabled="saving || loading" @click="saveConfig">{{ saving ? '保存中...' : '保存配置' }}</button>
      </div>
    </div>

    <div v-if="loading" class="card py-12 text-center text-sm text-gray-500 dark:text-gray-400">正在加载前台商品配置...</div>

    <template v-else>
      <div class="card space-y-4 p-5">
        <div class="grid gap-4 md:grid-cols-2">
          <LocalizedInput v-model="config.eyebrow" label="眉标" />
          <LocalizedInput v-model="config.title" label="标题" />
        </div>
        <LocalizedInput v-model="config.description" label="说明" multiline />
      </div>

      <section class="card space-y-4 p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 class="text-base font-bold text-gray-900 dark:text-white">商品类型</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">订阅和余额充值分开维护，保存后同时影响 /home 和用户端收银台。</p>
          </div>
          <div class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-1 dark:border-dark-600 dark:bg-dark-800">
            <button
              v-for="tab in pricingTabs"
              :key="tab.value"
              type="button"
              :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                activePricingTab === tab.value
                  ? 'bg-white text-primary-700 shadow-sm dark:bg-dark-700 dark:text-primary-300'
                  : 'text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white'
              ]"
              @click="activePricingTab = tab.value"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <div v-if="activePricingTab === 'subscription'" class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-base font-bold text-gray-900 dark:text-white">订阅商品</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">每张卡选择一个后台套餐。前台展示为订阅，价格使用套餐价格。</p>
            </div>
            <button class="btn btn-secondary" @click="addSubscriptionCard">添加订阅卡</button>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <LocalizedInput v-model="config.subscription_group.title" label="分组标题" />
            <LocalizedInput v-model="config.subscription_group.description" label="分组说明" />
          </div>

          <div class="space-y-3">
            <article v-for="(card, index) in config.subscription_cards" :key="card.id" class="rounded-xl border border-gray-200 p-4 dark:border-dark-600">
              <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="card.enabled" type="checkbox" class="rounded border-gray-300" />
                    启用
                  </label>
                  <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="card.highlight" type="checkbox" class="rounded border-gray-300" />
                    推荐高亮
                  </label>
                </div>
                <button class="btn btn-danger btn-sm" @click="removeSubscriptionCard(index)">删除</button>
              </div>

              <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_8rem]">
                <div>
                  <label class="input-label">选择套餐</label>
                  <select v-model.number="card.subscription_plan_id" class="input">
                    <option :value="0">请选择套餐</option>
                    <option v-for="plan in plans" :key="plan.id" :value="plan.id">
                      {{ plan.name }} - ¥{{ plan.price }}{{ plan.for_sale ? '' : '（已下架）' }}
                    </option>
                  </select>
                </div>
                <div>
                  <label class="input-label">排序</label>
                  <input v-model.number="card.sort_order" type="number" class="input" />
                </div>
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <LocalizedInput v-model="card.name" label="卡片名称" />
                <LocalizedInput v-model="card.badge" label="角标" optional />
                <LocalizedInput v-model="card.period" label="价格周期" optional />
                <LocalizedInput v-model="card.description" label="描述" multiline />
              </div>
              <MetricsEditor v-model="card.metrics" class="mt-4" />
            </article>
          </div>
        </div>

        <div v-else class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-base font-bold text-gray-900 dark:text-white">余额积分包</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">用户只能购买这里启用的积分包。支付金额是用户实付人民币，到账积分是支付成功后写入余额的积分。</p>
            </div>
            <button class="btn btn-secondary" @click="addCreditCard">添加积分包</button>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <LocalizedInput v-model="config.credit_group.title" label="分组标题" />
            <LocalizedInput v-model="config.credit_group.description" label="分组说明" />
          </div>

          <div class="space-y-3">
            <article v-for="(card, index) in config.credit_cards" :key="card.id" class="rounded-xl border border-gray-200 p-4 dark:border-dark-600">
              <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="card.enabled" type="checkbox" class="rounded border-gray-300" />
                    启用
                  </label>
                  <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="card.highlight" type="checkbox" class="rounded border-gray-300" />
                    推荐高亮
                  </label>
                </div>
                <button class="btn btn-danger btn-sm" @click="removeCreditCard(index)">删除</button>
              </div>

              <div class="grid gap-4 md:grid-cols-4">
                <div>
                  <label class="input-label">支付金额（元）</label>
                  <input v-model.number="card.recharge_amount" type="number" min="0.01" step="0.01" class="input" />
                </div>
                <div>
                  <label class="input-label">到账积分</label>
                  <input v-model.number="card.credited_amount" type="number" min="0.01" step="0.01" class="input" />
                </div>
                <div>
                  <label class="input-label">排序</label>
                  <input v-model.number="card.sort_order" type="number" class="input" />
                </div>
                <LocalizedInput v-model="card.badge" label="角标" optional />
              </div>

              <div class="mt-4 grid gap-4 md:grid-cols-2">
                <LocalizedInput v-model="card.name" label="卡片名称" />
                <LocalizedInput v-model="card.period" label="价格周期" optional />
                <LocalizedInput v-model="card.description" label="描述" multiline />
              </div>
              <MetricsEditor v-model="card.metrics" class="mt-4" />
            </article>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import type {
  HomePricingConfig,
  HomePricingLocalizedText,
  HomePricingMetricConfig,
  HomePricingSubscriptionCardConfig,
  HomePricingCreditCardConfig,
} from '@/types'
import type { SubscriptionPlan } from '@/types/payment'

const props = defineProps<{
  plans: SubscriptionPlan[]
}>()

const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)
const config = ref<HomePricingConfig>(createDefaultConfig())
const plans = computed(() => props.plans || [])
const activePricingTab = ref<'subscription' | 'credit'>('subscription')
const pricingTabs = [
  { value: 'subscription' as const, label: '订阅' },
  { value: 'credit' as const, label: '余额充值' },
]

function localized(zh = '', en = ''): HomePricingLocalizedText {
  return { zh, en }
}

function createDefaultConfig(): HomePricingConfig {
  return {
    eyebrow: localized('定价', 'Pricing'),
    title: localized('选择适合你的使用方式。', 'Choose the usage option that fits you.'),
    description: localized('可以购买订阅获得周期服务额度，也可以购买余额积分包作为补充；实际消耗会随模型、输入输出长度和工具行为变化。', 'Buy a subscription for recurring service credit, or add balance credit packages as backup. Actual usage varies by model, input/output length, and tool behavior.'),
    subscription_group: {
      title: localized('订阅', 'Subscription'),
      description: localized('适合每天持续使用 AI 工具的用户。', 'For users who run AI tools continuously.'),
    },
    credit_group: {
      title: localized('余额积分包', 'Credit Packages'),
      description: localized('适合灵活补充，额度进入余额后按实际使用消耗。', 'Flexible top-ups for account balance, consumed by actual usage.'),
    },
    subscription_cards: [],
    credit_cards: [
      createCreditCard(20, 80, 10),
      createCreditCard(40, 180, 20, true, localized('常用', 'Common')),
      createCreditCard(200, 1000, 30),
    ],
  }
}

function createMetric(labelZh: string, labelEn: string, valueZh: string, valueEn: string): HomePricingMetricConfig {
  return { label: localized(labelZh, labelEn), value: localized(valueZh, valueEn) }
}

function createSubscriptionCard(plan?: SubscriptionPlan): HomePricingSubscriptionCardConfig {
  const id = `subscription-${Date.now()}-${Math.random().toString(16).slice(2)}`
  return {
    id,
    enabled: true,
    sort_order: (config.value.subscription_cards.length + 1) * 10,
    subscription_plan_id: plan?.id || 0,
    name: localized(plan?.name || '', plan?.name || ''),
    description: localized(plan?.description || '', plan?.description || ''),
    badge: localized('', ''),
    period: localized('/ 月', '/ month'),
    highlight: false,
    metrics: [
      createMetric('每日额度', 'Daily credit', plan?.daily_limit_usd != null ? `约 ${plan.daily_limit_usd} 积分` : '约 20 积分', plan?.daily_limit_usd != null ? `about ${plan.daily_limit_usd} credits` : 'about 20 credits'),
      createMetric('token 估算', 'Token estimate', '按实际模型消耗', 'varies by model'),
    ],
  }
}

function createCreditCard(amount: number, creditedAmount: number, sortOrder: number, highlight = false, badge = localized('', '')): HomePricingCreditCardConfig {
  return {
    id: `credit-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    enabled: true,
    sort_order: sortOrder,
    recharge_amount: amount,
    credited_amount: creditedAmount,
    name: localized(`${creditedAmount} 积分`, `${creditedAmount} Credits`),
    description: localized(`${amount} 元购买 ${creditedAmount} 积分。`, `Pay ¥${amount} for ${creditedAmount} credits.`),
    badge,
    period: localized('', ''),
    highlight,
    metrics: [
      createMetric('到账积分', 'Credits received', `${creditedAmount} 积分`, `${creditedAmount} credits`),
      createMetric('类型', 'Type', '余额额度', 'balance credit'),
    ],
  }
}

async function loadConfig() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getHomePricingConfig()
    config.value = normalizeConfig(res.data)
  } catch (error: unknown) {
    appStore.showError(error instanceof Error ? error.message : '加载前台商品配置失败')
  } finally {
    loading.value = false
  }
}

function normalizeConfig(input: HomePricingConfig | null | undefined): HomePricingConfig {
  const fallback = createDefaultConfig()
  if (!input) return fallback
  return {
    ...fallback,
    ...input,
    eyebrow: input.eyebrow || fallback.eyebrow,
    title: input.title || fallback.title,
    description: input.description || fallback.description,
    subscription_group: input.subscription_group || fallback.subscription_group,
    credit_group: input.credit_group || fallback.credit_group,
    subscription_cards: input.subscription_cards || [],
    credit_cards: (input.credit_cards || fallback.credit_cards).map((card, index) => normalizeCreditCard(card, index)),
  }
}

function normalizeCreditCard(card: HomePricingCreditCardConfig, index: number): HomePricingCreditCardConfig {
  const amount = Number(card.recharge_amount || card.price || 0)
  const creditedAmount = Number(card.credited_amount || amount)
  return {
    ...card,
    sort_order: card.sort_order || (index + 1) * 10,
    recharge_amount: amount,
    credited_amount: creditedAmount,
    name: required(card.name) ? card.name : localized(`${creditedAmount} 积分`, `${creditedAmount} Credits`),
    description: required(card.description) ? card.description : localized(`${amount} 元购买 ${creditedAmount} 积分。`, `Pay ¥${amount} for ${creditedAmount} credits.`),
    period: card.period || localized('', ''),
    badge: card.badge || localized('', ''),
    highlight: Boolean(card.highlight),
    metrics: card.metrics?.length ? card.metrics : [
      createMetric('到账积分', 'Credits received', `${creditedAmount} 积分`, `${creditedAmount} credits`),
      createMetric('类型', 'Type', '余额额度', 'balance credit'),
    ],
  }
}

function validateConfig() {
  for (const card of config.value.subscription_cards) {
    if (!card.subscription_plan_id) return '订阅卡必须选择后台套餐'
    if (!required(card.name) || !required(card.description)) return '订阅卡名称和描述必须填写中英文'
  }
  for (const card of config.value.credit_cards) {
    if (!card.recharge_amount || card.recharge_amount <= 0) return '余额积分包支付金额必须大于 0'
    if (!card.credited_amount || card.credited_amount <= 0) return '余额积分包到账积分必须大于 0'
    if (!required(card.name) || !required(card.description)) return '余额积分包名称和描述必须填写中英文'
  }
  return ''
}

function required(value: HomePricingLocalizedText) {
  return !!value?.zh?.trim() && !!value?.en?.trim()
}

async function saveConfig() {
  const validationError = validateConfig()
  if (validationError) {
    appStore.showError(validationError)
    return
  }
  saving.value = true
  try {
    const payload = {
      ...config.value,
      credit_cards: config.value.credit_cards.map((card, index) => normalizeCreditCard(card, index)),
    }
    const res = await adminPaymentAPI.updateHomePricingConfig(payload)
    config.value = normalizeConfig(res.data)
    await appStore.fetchPublicSettings(true)
    appStore.showSuccess('前台商品配置已保存')
  } catch (error: unknown) {
    appStore.showError(error instanceof Error ? error.message : '保存前台商品配置失败')
  } finally {
    saving.value = false
  }
}

function addSubscriptionCard() {
  config.value.subscription_cards.push(createSubscriptionCard(plans.value[0]))
}

function removeSubscriptionCard(index: number) {
  config.value.subscription_cards.splice(index, 1)
}

function addCreditCard() {
  config.value.credit_cards.push(createCreditCard(20, 80, (config.value.credit_cards.length + 1) * 10))
}

function removeCreditCard(index: number) {
  config.value.credit_cards.splice(index, 1)
}

const LocalizedInput = defineComponent({
  props: {
    modelValue: { type: Object as () => HomePricingLocalizedText, required: true },
    label: { type: String, required: true },
    multiline: { type: Boolean, default: false },
    optional: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
  setup(componentProps, { emit }) {
    const update = (key: 'zh' | 'en', value: string) => {
      emit('update:modelValue', { ...componentProps.modelValue, [key]: value })
    }
    return () => h('div', [
      h('label', { class: 'input-label' }, `${componentProps.label}${componentProps.optional ? '' : ' *'}`),
      h('div', { class: 'grid gap-2 sm:grid-cols-2' }, ['zh', 'en'].map((key) => h(componentProps.multiline ? 'textarea' : 'input', {
        class: 'input',
        rows: componentProps.multiline ? 2 : undefined,
        value: componentProps.modelValue[key as 'zh' | 'en'] || '',
        placeholder: key === 'zh' ? '中文' : 'English',
        onInput: (event: Event) => update(key as 'zh' | 'en', (event.target as HTMLInputElement).value),
      }))),
    ])
  },
})

const MetricsEditor = defineComponent({
  props: {
    modelValue: { type: Array as () => HomePricingMetricConfig[], required: true },
  },
  emits: ['update:modelValue'],
  setup(componentProps, { emit, attrs }) {
    const updateMetric = (index: number, key: 'label' | 'value', lang: 'zh' | 'en', value: string) => {
      const next = componentProps.modelValue.map(metric => ({
        label: { ...metric.label },
        value: { ...metric.value },
      }))
      next[index][key][lang] = value
      emit('update:modelValue', next)
    }
    const addMetric = () => {
      emit('update:modelValue', [...componentProps.modelValue, createMetric('', '', '', '')])
    }
    const removeMetric = (index: number) => {
      emit('update:modelValue', componentProps.modelValue.filter((_, i) => i !== index))
    }
    return () => h('div', { class: ['space-y-2', attrs.class] }, [
      h('div', { class: 'flex items-center justify-between gap-2' }, [
        h('p', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, '指标文案'),
        h('button', { class: 'btn btn-secondary btn-sm', type: 'button', onClick: addMetric }, '添加指标'),
      ]),
      ...componentProps.modelValue.map((metric, index) => h('div', { class: 'grid gap-2 rounded-lg bg-gray-50 p-3 dark:bg-dark-800 md:grid-cols-[1fr_1fr_auto]' }, [
        h('div', { class: 'grid gap-2 sm:grid-cols-2' }, [
          h('input', { class: 'input', value: metric.label.zh, placeholder: '标签中文', onInput: (event: Event) => updateMetric(index, 'label', 'zh', (event.target as HTMLInputElement).value) }),
          h('input', { class: 'input', value: metric.label.en, placeholder: 'Label EN', onInput: (event: Event) => updateMetric(index, 'label', 'en', (event.target as HTMLInputElement).value) }),
        ]),
        h('div', { class: 'grid gap-2 sm:grid-cols-2' }, [
          h('input', { class: 'input', value: metric.value.zh, placeholder: '值中文', onInput: (event: Event) => updateMetric(index, 'value', 'zh', (event.target as HTMLInputElement).value) }),
          h('input', { class: 'input', value: metric.value.en, placeholder: 'Value EN', onInput: (event: Event) => updateMetric(index, 'value', 'en', (event.target as HTMLInputElement).value) }),
        ]),
        h('button', { class: 'btn btn-danger btn-sm', type: 'button', onClick: () => removeMetric(index) }, '删除'),
      ])),
    ])
  },
})

onMounted(loadConfig)
</script>
