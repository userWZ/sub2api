import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import AdminRenewalSettingsView from '../AdminRenewalSettingsView.vue'
import type { RenewalSettings } from '@/api/admin/payment'

const { getRenewalSettings, updateRenewalSettings, showError, showSuccess } = vi.hoisted(() => ({
  getRenewalSettings: vi.fn(),
  updateRenewalSettings: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: { getRenewalSettings, updateRenewalSettings },
  default: { getRenewalSettings, updateRenewalSettings },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('@/utils/apiError', () => ({
  extractI18nErrorMessage: (_error: unknown, _t: unknown, _prefix: string, fallback: string) => fallback,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => key.replace(/\{(\w+)\}/g, (_, token) => String(params?.[token] ?? `{${token}}`)),
    }),
  }
})

const settings = (): RenewalSettings => ({
  offer_enabled: false,
  before_expiry_days: 14,
  after_expiry_days: 7,
  discount_enabled: true,
  discount_percent: 10,
  discount_min_order_amount: 0,
  discount_max_amount: 20,
  rollover_enabled: true,
  rollover_percent: 20,
  rollover_min_unused_amount: 1,
  rollover_max_amount: 30,
  email_enabled: true,
  email_reminder_days: [14, 7, 1],
})

function mountView() {
  return mount(AdminRenewalSettingsView, {
    global: {
      stubs: {
        AppLayout: defineComponent({ setup: (_, { slots }) => () => h('main', slots.default?.()) }),
        Icon: true,
        EmailTemplateEditor: defineComponent({ props: ['event', 'lockEvent'], template: '<div data-test="email-editor" />' }),
      },
    },
  })
}

describe('AdminRenewalSettingsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getRenewalSettings.mockResolvedValue({ data: settings() })
    updateRenewalSettings.mockImplementation(async (payload: RenewalSettings) => ({ data: payload }))
  })

  it('loads independent benefit switches and the locked renewal email editor', async () => {
    const wrapper = mountView()
    await flushPromises()

    const switches = wrapper.findAll('[role="switch"]')
    expect(switches).toHaveLength(4)
    expect(switches.map(item => item.attributes('aria-checked'))).toEqual(['false', 'true', 'true', 'true'])
    expect(wrapper.find('[data-test="email-editor"]').exists()).toBe(true)
  })

  it('persists the complete dedicated configuration', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('button.btn-primary').trigger('click')
    await flushPromises()

    expect(updateRenewalSettings).toHaveBeenCalledWith(expect.objectContaining({
      offer_enabled: false,
      before_expiry_days: 14,
      after_expiry_days: 7,
      discount_enabled: true,
      rollover_enabled: true,
      email_enabled: true,
      email_reminder_days: [14, 7, 1],
    }))
    expect(showSuccess).toHaveBeenCalledWith('admin.renewal.saved')
  })
})
