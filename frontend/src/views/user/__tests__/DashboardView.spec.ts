import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DashboardView from '../DashboardView.vue'

const refreshUser = vi.hoisted(() => vi.fn())
const getDashboardStats = vi.hoisted(() => vi.fn())
const getDashboardTrend = vi.hoisted(() => vi.fn())
const getDashboardModels = vi.hoisted(() => vi.fn())
const getByDateRange = vi.hoisted(() => vi.fn())
const listKeys = vi.hoisted(() => vi.fn())
const createKey = vi.hoisted(() => vi.fn())
const getPublicSettings = vi.hoisted(() => vi.fn())
const getMyPlatformQuotas = vi.hoisted(() => vi.fn())
const fetchActiveSubscriptions = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

const authState = vi.hoisted(() => ({
  user: {
    id: 1,
    username: 'demo',
    email: 'demo@example.com',
    role: 'user',
    balance: 12.5,
    concurrency: 3,
    status: 'active',
  },
  isSimpleMode: false,
  refreshUser,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      locale: { value: 'zh-CN' },
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api', () => ({
  authAPI: {
    getPublicSettings,
  },
  keysAPI: {
    list: listKeys,
    create: createKey,
  },
  usageAPI: {
    getDashboardStats,
    getDashboardTrend,
    getDashboardModels,
    getByDateRange,
  },
}))

vi.mock('@/api/user', () => ({
  getMyPlatformQuotas,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authState,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showError,
  }),
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    fetchActiveSubscriptions,
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard,
  }),
}))

const statsFixture = {
  total_api_keys: 1,
  active_api_keys: 1,
  today_requests: 2,
  total_requests: 10,
  today_cost: 0.1,
  today_actual_cost: 0.08,
  total_cost: 1,
  total_actual_cost: 0.8,
  today_tokens: 100,
  today_input_tokens: 60,
  today_output_tokens: 40,
  total_tokens: 1000,
  total_input_tokens: 600,
  total_output_tokens: 400,
  rpm: 1,
  tpm: 100,
  average_duration_ms: 300,
  by_platform: [],
}

function activeKeyFixture() {
  return {
    id: 10,
    user_id: 1,
    key: 'sk-oceanway-default-1234567890',
    name: 'Default Key',
    group_id: null,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    quota_disabled: false,
    last_used_at: null,
    quota: 0,
    quota_used: 0,
    expires_at: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    usage_5h: 0,
    usage_1d: 0,
    usage_7d: 0,
    window_5h_start: null,
    window_1d_start: null,
    window_7d_start: null,
    reset_5h_at: null,
    reset_1d_at: null,
    reset_7d_at: null,
  }
}

function mountDashboard() {
  return mount(DashboardView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        LoadingSpinner: { template: '<span />' },
        UserDashboardStats: { template: '<div data-test="stats" />' },
        UserDashboardCharts: { template: '<div data-test="charts" />' },
        UserDashboardRecentUsage: { template: '<div data-test="recent" />' },
        UserDashboardQuickActions: { template: '<div data-test="quick" />' },
        RouterLink: { props: ['to'], template: '<a><slot /></a>' },
      },
    },
  })
}

describe('DashboardView', () => {
  beforeEach(() => {
    refreshUser.mockReset().mockResolvedValue(undefined)
    getDashboardStats.mockReset().mockResolvedValue(statsFixture)
    getDashboardTrend.mockReset().mockResolvedValue({ trend: [] })
    getDashboardModels.mockReset().mockResolvedValue({ models: [] })
    getByDateRange.mockReset().mockResolvedValue({ items: [] })
    listKeys.mockReset().mockResolvedValue({ items: [activeKeyFixture()] })
    createKey.mockReset().mockResolvedValue(activeKeyFixture())
    getPublicSettings.mockReset().mockResolvedValue({ api_base_url: 'https://api.oceanway.example/v1' })
    getMyPlatformQuotas.mockReset().mockResolvedValue({ platform_quotas: [] })
    fetchActiveSubscriptions.mockReset().mockResolvedValue([
      {
        id: 1,
        user_id: 1,
        group_id: 2,
        status: 'active',
        starts_at: '2026-01-01T00:00:00Z',
        daily_usage_usd: 1,
        weekly_usage_usd: 2,
        monthly_usage_usd: 3,
        daily_window_start: null,
        weekly_window_start: null,
        monthly_window_start: null,
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z',
        expires_at: null,
        group: {
          daily_limit_usd: 5,
          weekly_limit_usd: 20,
          monthly_limit_usd: 50,
        },
      },
    ])
    copyToClipboard.mockReset().mockResolvedValue(true)
    showSuccess.mockReset()
    showError.mockReset()
  })

  it('shows balance, subscription quota, default key, and subscription-first routing', async () => {
    const wrapper = mountDashboard()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('当前账户可以发起 API 调用')
    expect(text).toContain('$12.50')
    expect(text).toContain('$4.00')
    expect(text).toContain('Default Key')
    expect(text).toContain('sk-ocean...7890')
    expect(text).toContain('https://api.oceanway.example/v1')
    expect(text).toContain('优先使用 1 个有效订阅')
    expect(wrapper.find('[data-test="charts"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="recent"]').exists()).toBe(true)
  })

  it('creates an unbound quota-enabled default key when no key exists', async () => {
    listKeys.mockResolvedValue({ items: [] })
    const wrapper = mountDashboard()
    await flushPromises()

    await wrapper.findAll('button').find(button => button.text() === '创建默认 Key')?.trigger('click')
    await flushPromises()

    expect(createKey).toHaveBeenCalledWith('默认 API Key', null, undefined, undefined, undefined, undefined, undefined, undefined, false)
    expect(showSuccess).toHaveBeenCalledWith('默认 API Key 已创建并复制')
    expect(copyToClipboard).toHaveBeenCalledWith('sk-oceanway-default-1234567890', '已复制')
  })
})
