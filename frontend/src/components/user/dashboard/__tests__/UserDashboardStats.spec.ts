import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserDashboardStats from '../UserDashboardStats.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const stats = {
  total_api_keys: 1,
  active_api_keys: 1,
  total_requests: 10,
  total_input_tokens: 600,
  total_output_tokens: 400,
  total_cache_creation_tokens: 10,
  total_cache_read_tokens: 300,
  total_tokens: 1310,
  total_cost: 1,
  total_actual_cost: 0.8,
  today_requests: 2,
  today_input_tokens: 60,
  today_output_tokens: 40,
  today_cache_creation_tokens: 5,
  today_cache_read_tokens: 200,
  today_tokens: 305,
  today_cost: 0.1,
  today_actual_cost: 0.08,
  average_duration_ms: 300,
  rpm: 1,
  tpm: 305,
  by_platform: [],
}

describe('UserDashboardStats', () => {
  it('shows cache tokens in today and total token breakdowns', () => {
    const wrapper = mount(UserDashboardStats, {
      props: {
        stats,
        balance: 12.5,
        isSimple: false,
        platformQuotas: [],
      },
      global: {
        stubs: {
          Icon: { template: '<span />' },
        },
      },
    })

    const text = wrapper.text()

    expect(text).toContain('dashboard.todayTokens')
    expect(text).toContain('305')
    expect(text).toContain('dashboard.cache: 200')
    expect(text).toContain('dashboard.cacheCreate: 5')
    expect(text).toContain('dashboard.totalTokens')
    expect(text).toContain('1.3K')
    expect(text).toContain('dashboard.cache: 300')
    expect(text).toContain('dashboard.cacheCreate: 10')
  })
})
