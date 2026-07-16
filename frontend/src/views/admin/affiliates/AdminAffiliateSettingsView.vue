<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.settings.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.affiliates.settings.description') }}</p>
        </div>
        <button type="button" class="btn btn-primary inline-flex items-center gap-2" :disabled="loading || saving" @click="saveSettings">
          <Icon name="check" size="sm" />
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>

      <div v-if="loading" class="flex justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <template v-else>
        <div class="rounded-lg border border-sky-200 bg-sky-50 p-4 text-sm text-sky-800 dark:border-sky-900 dark:bg-sky-950/40 dark:text-sky-200">
          {{ t('admin.affiliates.settings.creditNotice') }}
        </div>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.settings.baseTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.affiliates.settings.baseHint') }}</p>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.affiliate_enabled" :label="t('admin.affiliates.settings.enabled')" :hint="t('admin.affiliates.settings.enabledHint')" />

            <div v-if="form.affiliate_enabled" class="grid gap-5 md:grid-cols-2">
              <NumberField v-model="form.rebate_rate" :label="t('admin.affiliates.settings.rebateRate')" suffix="%" :min="0" :max="100" :step="0.01" />
              <NumberField v-model="form.rebate_freeze_hours" :label="t('admin.affiliates.settings.freezeHours')" :min="0" :max="720" :step="1" />
              <NumberField v-model="form.rebate_duration_days" :label="t('admin.affiliates.settings.durationDays')" :min="0" :max="3650" :step="1" />
              <NumberField v-model="form.rebate_per_invitee_cap" :label="t('admin.affiliates.settings.perInviteeCap')" :min="0" :step="0.01" />
            </div>

            <div v-if="form.affiliate_enabled" class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
              <SettingToggle v-model="form.discount_enabled" :label="t('admin.affiliates.settings.discountEnabled')" :hint="t('admin.affiliates.settings.discountEnabledHint')" />
              <div v-if="form.discount_enabled" class="mt-5 grid gap-5 md:grid-cols-2">
                <NumberField v-model="form.discount_max_percent" :label="t('admin.affiliates.settings.discountMaxPercent')" suffix="%" :min="0" :max="100" :step="0.01" />
                <NumberField v-model="form.discount_min_pay_amount" :label="t('admin.affiliates.settings.discountMinPay')" :min="0" :step="0.01" />
              </div>
            </div>
          </div>
        </section>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.settings.usageRewardTitle') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.affiliates.settings.usageRewardHint') }}</p>
          </div>
          <div class="space-y-6 p-6">
            <SettingToggle v-model="form.usage_reward_enabled" :label="t('admin.affiliates.settings.usageRewardEnabled')" :hint="t('admin.affiliates.settings.usageRewardEnabledHint')" />
            <div v-if="form.usage_reward_enabled" class="grid gap-5 md:grid-cols-2">
              <NumberField v-model="form.usage_reward_amount" :label="t('admin.affiliates.settings.rewardAmount')" :hint="t('admin.affiliates.settings.rewardAmountHint')" :min="0.01" :step="0.01" />
            </div>
          </div>
        </section>

        <section class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.settings.riskTitle') }}</h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.affiliates.settings.riskHint') }}</p>
              </div>
              <router-link to="/admin/affiliates/usage-rewards" class="text-sm font-medium text-primary-600 hover:underline dark:text-primary-400">
                {{ t('admin.affiliates.settings.openReview') }} →
              </router-link>
            </div>
          </div>
          <div class="space-y-6 p-6">
            <div class="grid gap-5 md:grid-cols-3">
              <NumberField v-model="form.reward_inviter_daily_limit" :label="t('admin.affiliates.settings.inviterDailyLimit')" :hint="t('admin.affiliates.settings.zeroUnlimited')" :min="0" :step="1" />
              <NumberField v-model="form.reward_inviter_30d_limit" :label="t('admin.affiliates.settings.inviter30dLimit')" :hint="t('admin.affiliates.settings.zeroUnlimited')" :min="0" :step="1" />
              <NumberField v-model="form.reward_ip_daily_limit" :label="t('admin.affiliates.settings.ipDailyLimit')" :hint="t('admin.affiliates.settings.zeroUnlimited')" :min="0" :step="1" />
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <SettingToggle v-model="form.reward_review_same_ip" :label="t('admin.affiliates.settings.reviewSameIP')" :hint="t('admin.affiliates.settings.reviewSameIPHint')" boxed />
              <SettingToggle v-model="form.reward_review_missing_ip" :label="t('admin.affiliates.settings.reviewMissingIP')" :hint="t('admin.affiliates.settings.reviewMissingIPHint')" boxed />
            </div>
          </div>
        </section>

        <AffiliateCustomUsersCard />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import AffiliateCustomUsersCard from './AffiliateCustomUsersCard.vue'
import { affiliatesAPI, type AffiliateSettings } from '@/api/admin/affiliates'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)

const defaults: AffiliateSettings = {
  affiliate_enabled: false,
  rebate_rate: 20,
  rebate_freeze_hours: 0,
  rebate_duration_days: 0,
  rebate_per_invitee_cap: 0,
  discount_enabled: true,
  discount_max_percent: 50,
  discount_min_pay_amount: 1,
  usage_reward_enabled: false,
  usage_reward_amount: 20,
  reward_inviter_daily_limit: 5,
  reward_inviter_30d_limit: 30,
  reward_ip_daily_limit: 2,
  reward_review_same_ip: true,
  reward_review_missing_ip: true,
}
const form = reactive<AffiliateSettings>({ ...defaults })

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
      class: [
        'flex items-center justify-between gap-4',
        props.boxed ? 'rounded-lg border border-gray-200 p-4 dark:border-dark-700' : '',
      ],
    }, [
      h('div', { class: 'min-w-0' }, [
        h('div', { class: 'text-sm font-medium text-gray-800 dark:text-gray-200' }, props.label),
        props.hint ? h('p', { class: 'mt-1 text-xs text-gray-500 dark:text-gray-400' }, props.hint) : null,
      ]),
      h(Toggle, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value: boolean) => emit('update:modelValue', value),
      }),
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
          type: 'number',
          class: ['input', props.suffix ? 'pr-9' : ''],
          value: localValue.value,
          min: props.min,
          max: props.max,
          step: props.step,
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

async function loadSettings() {
  loading.value = true
  try {
    Object.assign(form, await affiliatesAPI.getSettings())
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    const payload: AffiliateSettings = {
      ...form,
      rebate_rate: Math.min(100, Math.max(0, Number(form.rebate_rate) || 0)),
      rebate_freeze_hours: Math.min(720, Math.max(0, Math.floor(Number(form.rebate_freeze_hours) || 0))),
      rebate_duration_days: Math.min(3650, Math.max(0, Math.floor(Number(form.rebate_duration_days) || 0))),
      rebate_per_invitee_cap: Math.max(0, Number(form.rebate_per_invitee_cap) || 0),
      discount_max_percent: Math.min(100, Math.max(0, Number(form.discount_max_percent) || 0)),
      discount_min_pay_amount: Math.max(0, Number(form.discount_min_pay_amount) || 0),
      usage_reward_amount: Math.max(0.01, Number(form.usage_reward_amount) || 20),
      reward_inviter_daily_limit: Math.max(0, Math.floor(Number(form.reward_inviter_daily_limit) || 0)),
      reward_inviter_30d_limit: Math.max(0, Math.floor(Number(form.reward_inviter_30d_limit) || 0)),
      reward_ip_daily_limit: Math.max(0, Math.floor(Number(form.reward_ip_daily_limit) || 0)),
    }
    Object.assign(form, await affiliatesAPI.updateSettings(payload))
    appStore.showSuccess(t('admin.affiliates.settings.saved'))
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>
