import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AvailableChannelsView from '../AvailableChannelsView.vue'

const getAvailable = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => (
        params?.count != null ? `${key}:${params.count}` : key
      ),
    }),
  }
})

vi.mock('@/api/channels', () => ({
  default: {
    getAvailable,
  },
}))

vi.mock('@/api/groups', () => ({
  default: {
    getUserGroupRates,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
  }),
}))

describe('AvailableChannelsView', () => {
  beforeEach(() => {
    getAvailable.mockReset()
    getUserGroupRates.mockReset()
    showError.mockReset()
    getUserGroupRates.mockResolvedValue({})
  })

  it('summarizes account-level service capabilities from available models', async () => {
    getAvailable.mockResolvedValue([
      {
        name: 'OceanWay OpenAI',
        description: 'primary',
        platforms: [
          {
            platform: 'openai',
            groups: [],
            supported_models: [
              { name: 'gpt-4o', platform: 'openai', pricing: null },
              {
                name: 'dall-e-3',
                platform: 'openai',
                pricing: {
                  billing_mode: 'image',
                  input_price: null,
                  output_price: null,
                  cache_write_price: null,
                  cache_read_price: null,
                  image_output_price: 0.02,
                  per_request_price: null,
                  intervals: [],
                },
              },
            ],
          },
        ],
      },
    ])

    const wrapper = mount(AvailableChannelsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
          AvailableChannelsTable: true,
          RouterLink: { template: '<a><slot /></a>' },
        },
      },
    })
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('availableChannels.serviceOverviewTitle')
    expect(text).toContain('availableChannels.services.text')
    expect(text).toContain('availableChannels.serviceModelCount:2')
    expect(text).toContain('availableChannels.services.image')
    expect(text).toContain('availableChannels.serviceImageCount:1')
    expect(text).toContain('availableChannels.services.openai')
    expect(text).toContain('availableChannels.serviceIncluded')
  })
})
