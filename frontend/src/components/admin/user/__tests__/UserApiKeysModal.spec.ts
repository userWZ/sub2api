import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import UserApiKeysModal from '../UserApiKeysModal.vue'

const getUserApiKeys = vi.hoisted(() => vi.fn())
const getAllGroups = vi.hoisted(() => vi.fn())
const updateApiKeyGroup = vi.hoisted(() => vi.fn())
const updateApiKeyPolicy = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => (
        params?.group ? `${key}:${params.group}` : key
      ),
    }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess,
    showError,
  }),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserApiKeys,
    },
    groups: {
      getAll: getAllGroups,
    },
    apiKeys: {
      updateApiKeyGroup,
      updateApiKeyPolicy,
    },
  },
}))

function keyFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 10,
    user_id: 1,
    key: 'sk-admin-user-key-1234567890',
    name: 'Default Key',
    group_id: null,
    status: 'active',
    ip_whitelist: [],
    ip_blacklist: [],
    quota_disabled: true,
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
    ...overrides,
  }
}

function userFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    username: 'demo',
    email: 'demo@example.com',
    role: 'user',
    balance: 0,
    concurrency: 1,
    status: 'active',
    allowed_groups: null,
    balance_notify_enabled: false,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

function mountModal(user = userFixture()) {
  return mount(UserApiKeysModal, {
    props: {
      show: false,
      user,
    },
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /></div>' },
        GroupBadge: { template: '<span />' },
        GroupOptionItem: { template: '<span />' },
        Teleport: true,
      },
    },
  })
}

describe('UserApiKeysModal', () => {
  beforeEach(() => {
    getUserApiKeys.mockReset().mockResolvedValue({ items: [keyFixture()] })
    getAllGroups.mockReset().mockResolvedValue([])
    updateApiKeyGroup.mockReset()
    updateApiKeyPolicy.mockReset().mockResolvedValue({ api_key: keyFixture({ status: 'inactive' }) })
    showSuccess.mockReset()
    showError.mockReset()
  })

  it('updates key status and does not render key-level image controls', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('admin.users.keyQuotaDisabled')
    expect(wrapper.text()).not.toContain('可生图')
    expect(wrapper.text()).not.toContain('allow_image_generation')

    await wrapper.findAll('button').find(button => button.text() === 'admin.users.disableKey')?.trigger('click')
    await flushPromises()

    expect(updateApiKeyPolicy).toHaveBeenCalledWith(10, { status: 'inactive' })
    expect(showSuccess).toHaveBeenCalledWith('admin.users.keyPolicyUpdated')
  })

  it('hides quota toggle for managed users', async () => {
    const wrapper = mountModal(userFixture({ customer_type: 'managed' }))
    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).not.toContain('admin.users.keyQuotaDisabled')
    expect(wrapper.text()).not.toContain('admin.users.disableKeyQuota')
  })
})
