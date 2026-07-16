<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('admin.renewal.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.renewal.description') }}</p>
        </div>
        <button type="button" class="btn btn-primary inline-flex items-center gap-2" :disabled="loading || saving || !formValid" @click="saveSettings">
          <Icon name="check" size="sm" />
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>

      <div v-if="loading" class="flex justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <template v-else>
        <div class="rounded-lg border border-sky-200 bg-sky-50 p-4 text-sm leading-6 text-sky-800 dark:border-sky-900 dark:bg-sky-950/40 dark:text-sky-200">
          {{ t('admin.renewal.orderNotice') }}
        </div>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.renewal.baseTitle') }}</h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.renewal.baseHint') }}</p>
              </div>
              <span class="inline-flex w-fit items-center rounded-full px-3 py-1 text-xs font-medium" :class="form.offer_enabled ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'">
                {{ form.offer_enabled ? t('admin.renewal.running') : t('admin.renewal.stopped') }}
              </span>
            </div>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.offer_enabled" :label="t('admin.renewal.offerEnabled')" :hint="t('admin.renewal.offerEnabledHint')" />

            <div class="grid gap-5 md:grid-cols-2">
              <NumberField v-model="form.before_expiry_days" :label="t('admin.renewal.beforeDays')" :hint="t('admin.renewal.beforeDaysHint')" :min="0" :max="365" :suffix="t('admin.renewal.daysUnit')" />
              <NumberField v-model="form.after_expiry_days" :label="t('admin.renewal.afterDays')" :hint="t('admin.renewal.afterDaysHint')" :min="0" :max="365" :suffix="t('admin.renewal.daysUnit')" />
            </div>

            <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-800/60 dark:text-gray-300">
              {{ eligibilityPreview }}
            </div>
          </div>
        </section>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.renewal.discountTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.renewal.discountHint') }}</p>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.discount_enabled" :label="t('admin.renewal.discountEnabled')" :hint="t('admin.renewal.discountEnabledHint')" boxed />
            <div v-if="form.discount_enabled" class="grid gap-5 md:grid-cols-3">
              <NumberField v-model="form.discount_percent" :label="t('admin.renewal.discountPercent')" :min="0" :max="99.99" :step="0.01" suffix="%" />
              <NumberField v-model="form.discount_min_order_amount" :label="t('admin.renewal.discountMinOrder')" :hint="t('admin.renewal.zeroNoLimit')" :min="0" :step="0.01" />
              <NumberField v-model="form.discount_max_amount" :label="t('admin.renewal.discountMaxAmount')" :hint="t('admin.renewal.zeroNoLimit')" :min="0" :step="0.01" />
            </div>
            <div v-if="form.discount_enabled" class="rounded-lg border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-800 dark:border-emerald-900 dark:bg-emerald-950/30 dark:text-emerald-200">
              {{ discountPreview }}
            </div>
          </div>
        </section>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.renewal.rolloverTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.renewal.rolloverHint') }}</p>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.rollover_enabled" :label="t('admin.renewal.rolloverEnabled')" :hint="t('admin.renewal.rolloverEnabledHint')" boxed />
            <div v-if="form.rollover_enabled" class="grid gap-5 md:grid-cols-3">
              <NumberField v-model="form.rollover_percent" :label="t('admin.renewal.rolloverPercent')" :min="0" :max="100" :step="0.01" suffix="%" />
              <NumberField v-model="form.rollover_min_unused_amount" :label="t('admin.renewal.rolloverMinUnused')" :hint="t('admin.renewal.zeroNoLimit')" :min="0" :step="0.01" />
              <NumberField v-model="form.rollover_max_amount" :label="t('admin.renewal.rolloverMaxAmount')" :hint="t('admin.renewal.zeroNoLimit')" :min="0" :step="0.01" />
            </div>
            <div v-if="form.rollover_enabled" class="rounded-lg border border-violet-200 bg-violet-50 p-4 text-sm text-violet-800 dark:border-violet-900 dark:bg-violet-950/30 dark:text-violet-200">
              {{ rolloverPreview }}
            </div>
          </div>
        </section>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.renewal.emailTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.renewal.emailHint') }}</p>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.email_enabled" :label="t('admin.renewal.emailEnabled')" :hint="t('admin.renewal.emailEnabledHint')" boxed />

            <div v-if="form.email_enabled" class="space-y-3">
              <div>
                <div class="input-label">{{ t('admin.renewal.reminderDays') }}</div>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.renewal.reminderDaysHint') }}</p>
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="day in reminderPresets"
                  :key="day"
                  type="button"
                  class="rounded-md border px-3 py-2 text-sm font-medium transition-colors"
                  :class="hasReminderDay(day) ? 'border-primary-500 bg-primary-50 text-primary-700 dark:bg-primary-950/30 dark:text-primary-300' : 'border-gray-200 bg-white text-gray-600 hover:border-primary-300 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300'"
                  :disabled="day > form.before_expiry_days"
                  :aria-pressed="hasReminderDay(day)"
                  @click="toggleReminderDay(day)"
                >
                  {{ day === 0 ? t('admin.renewal.expiryDay') : t('admin.renewal.daysBefore', { day }) }}
                </button>
              </div>
              <div class="flex max-w-sm gap-2">
                <input v-model.number="customReminderDay" type="number" class="input" min="0" :max="form.before_expiry_days" :placeholder="t('admin.renewal.customReminderPlaceholder')" @keyup.enter="addCustomReminderDay" />
                <button type="button" class="btn btn-secondary whitespace-nowrap" @click="addCustomReminderDay">{{ t('common.add') }}</button>
              </div>
              <p v-if="form.email_reminder_days.length === 0" class="text-sm text-red-600 dark:text-red-400">{{ t('admin.renewal.reminderRequired') }}</p>
            </div>
          </div>
        </section>

        <EmailTemplateEditor v-if="form.email_enabled" event="subscription.renewal_offer" lock-event />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import EmailTemplateEditor from '@/views/admin/settings/EmailTemplateEditor.vue'
import { adminPaymentAPI, type RenewalSettings } from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const customReminderDay = ref<number | null>(null)
const reminderPresets = [14, 7, 3, 1, 0]

const defaults: RenewalSettings = {
  offer_enabled: false,
  before_expiry_days: 14,
  after_expiry_days: 14,
  discount_enabled: true,
  discount_percent: 10,
  discount_min_order_amount: 0,
  discount_max_amount: 0,
  rollover_enabled: true,
  rollover_percent: 20,
  rollover_min_unused_amount: 0,
  rollover_max_amount: 0,
  email_enabled: true,
  email_reminder_days: [14],
}

const form = reactive<RenewalSettings>({ ...defaults, email_reminder_days: [...defaults.email_reminder_days] })

const SettingToggle = defineComponent({
  props: {
    modelValue: { type: Boolean, required: true },
    label: { type: String, required: true },
    hint: { type: String, default: '' },
    boxed: { type: Boolean, default: false },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('div', {
      class: ['flex items-center justify-between gap-4', props.boxed ? 'rounded-lg border border-gray-200 p-4 dark:border-dark-700' : ''],
    }, [
      h('div', { class: 'min-w-0' }, [
        h('div', { class: 'text-sm font-medium text-gray-800 dark:text-gray-200' }, props.label),
        props.hint ? h('p', { class: 'mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400' }, props.hint) : null,
      ]),
      h(Toggle, { modelValue: props.modelValue, ariaLabel: props.label, 'onUpdate:modelValue': (value: boolean) => emit('update:modelValue', value) }),
    ])
  },
})

const NumberField = defineComponent({
  props: {
    modelValue: { type: Number, required: true },
    label: { type: String, required: true },
    hint: { type: String, default: '' },
    suffix: { type: String, default: '' },
    min: { type: Number, default: undefined },
    max: { type: Number, default: undefined },
    step: { type: Number, default: 1 },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const localValue = ref(props.modelValue)
    watch(() => props.modelValue, value => { localValue.value = value })
    return () => h('label', { class: 'block' }, [
      h('span', { class: 'input-label' }, props.label),
      h('div', { class: 'relative' }, [
        h('input', {
          type: 'number', class: ['input', props.suffix ? 'pr-12' : ''], value: localValue.value,
          min: props.min, max: props.max, step: props.step,
          onInput: (event: Event) => {
            const value = Number((event.target as HTMLInputElement).value)
            localValue.value = value
            emit('update:modelValue', value)
          },
        }),
        props.suffix ? h('span', { class: 'pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-gray-400' }, props.suffix) : null,
      ]),
      props.hint ? h('p', { class: 'mt-1 text-xs text-gray-500 dark:text-gray-400' }, props.hint) : null,
    ])
  },
})

const formValid = computed(() => {
  if (form.before_expiry_days < 0 || form.before_expiry_days > 365 || form.after_expiry_days < 0 || form.after_expiry_days > 365) return false
  if (form.discount_percent < 0 || form.discount_percent >= 100 || form.rollover_percent < 0 || form.rollover_percent > 100) return false
  if ([form.discount_min_order_amount, form.discount_max_amount, form.rollover_min_unused_amount, form.rollover_max_amount].some(value => value < 0)) return false
  if (form.email_enabled && form.email_reminder_days.length === 0) return false
  return form.email_reminder_days.every(day => Number.isInteger(day) && day >= 0 && day <= form.before_expiry_days)
})

const eligibilityPreview = computed(() => t('admin.renewal.eligibilityPreview', {
  before: form.before_expiry_days,
  after: form.after_expiry_days,
}))

const discountPreview = computed(() => {
  const sample = 100
  const rawDiscount = sample * Math.max(0, form.discount_percent) / 100
  const discount = form.discount_max_amount > 0 ? Math.min(rawDiscount, form.discount_max_amount) : rawDiscount
  return t('admin.renewal.discountPreview', { amount: sample.toFixed(2), discount: discount.toFixed(2), payable: (sample - discount).toFixed(2) })
})

const rolloverPreview = computed(() => {
  const sample = 100
  const rawRollover = sample * Math.max(0, form.rollover_percent) / 100
  const rollover = form.rollover_max_amount > 0 ? Math.min(rawRollover, form.rollover_max_amount) : rawRollover
  return t('admin.renewal.rolloverPreview', { unused: sample.toFixed(2), rollover: rollover.toFixed(2) })
})

function normalizeReminderDays(days: number[]) {
  return [...new Set(days.map(Number).filter(day => Number.isInteger(day) && day >= 0 && day <= 365))].sort((a, b) => b - a)
}

function hasReminderDay(day: number) {
  return form.email_reminder_days.includes(day)
}

function toggleReminderDay(day: number) {
  if (day > form.before_expiry_days) return
  form.email_reminder_days = hasReminderDay(day)
    ? form.email_reminder_days.filter(item => item !== day)
    : normalizeReminderDays([...form.email_reminder_days, day])
}

function addCustomReminderDay() {
  const day = Number(customReminderDay.value)
  if (!Number.isInteger(day) || day < 0 || day > form.before_expiry_days) {
    appStore.showError(t('admin.renewal.invalidReminderDay', { max: form.before_expiry_days }))
    return
  }
  form.email_reminder_days = normalizeReminderDays([...form.email_reminder_days, day])
  customReminderDay.value = null
}

async function loadSettings() {
  loading.value = true
  try {
    const response = await adminPaymentAPI.getRenewalSettings()
    Object.assign(form, response.data, { email_reminder_days: normalizeReminderDays(response.data.email_reminder_days || []) })
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.renewal.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  if (!formValid.value) {
    appStore.showError(t('admin.renewal.invalidForm'))
    return
  }
  saving.value = true
  try {
    const payload: RenewalSettings = {
      ...form,
      before_expiry_days: Math.floor(Number(form.before_expiry_days)),
      after_expiry_days: Math.floor(Number(form.after_expiry_days)),
      discount_percent: Number(form.discount_percent),
      discount_min_order_amount: Number(form.discount_min_order_amount),
      discount_max_amount: Number(form.discount_max_amount),
      rollover_percent: Number(form.rollover_percent),
      rollover_min_unused_amount: Number(form.rollover_min_unused_amount),
      rollover_max_amount: Number(form.rollover_max_amount),
      email_reminder_days: normalizeReminderDays(form.email_reminder_days),
    }
    const response = await adminPaymentAPI.updateRenewalSettings(payload)
    Object.assign(form, response.data, { email_reminder_days: normalizeReminderDays(response.data.email_reminder_days || []) })
    appStore.showSuccess(t('admin.renewal.saved'))
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.renewal.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}

watch(() => form.before_expiry_days, maxDay => {
  form.email_reminder_days = form.email_reminder_days.filter(day => day <= maxDay)
})

onMounted(loadSettings)
</script>
