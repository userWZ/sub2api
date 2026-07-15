import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, put, post } = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
  post: vi.fn(),
}))

vi.mock('../client', () => ({
  apiClient: { get, put, post },
}))

import { getSettings, listUsageRewards, reviewUsageReward, updateSettings, type AffiliateSettings } from '@/api/admin/affiliates'

const settings: AffiliateSettings = {
  affiliate_enabled: true,
  rebate_rate: 20,
  rebate_freeze_hours: 0,
  rebate_duration_days: 0,
  rebate_per_invitee_cap: 0,
  discount_enabled: true,
  discount_max_percent: 50,
  discount_min_pay_amount: 1,
  usage_reward_enabled: true,
  usage_reward_amount: 20,
  reward_inviter_daily_limit: 5,
  reward_inviter_30d_limit: 30,
  reward_ip_daily_limit: 2,
  reward_review_same_ip: true,
  reward_review_missing_ip: true,
}

describe('admin affiliate usage reward api', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads and updates dedicated affiliate settings', async () => {
    get.mockResolvedValueOnce({ data: settings })
    put.mockResolvedValueOnce({ data: settings })

    await expect(getSettings()).resolves.toEqual(settings)
    await expect(updateSettings(settings)).resolves.toEqual(settings)
    expect(get).toHaveBeenCalledWith('/admin/affiliates/settings')
    expect(put).toHaveBeenCalledWith('/admin/affiliates/settings', settings)
  })

  it('lists and reviews first-usage rewards', async () => {
    get.mockResolvedValueOnce({ data: { items: [], total: 0, page: 1, page_size: 20 } })
    post.mockResolvedValueOnce({ data: { reward_id: 7, inviter_id: 9, status: 'granted', amount: 20 } })

    await listUsageRewards({ status: 'pending', page: 2, page_size: 50 })
    await reviewUsageReward(7, { approve: true, remark: 'verified' })

    expect(get).toHaveBeenCalledWith('/admin/affiliates/usage-rewards', expect.objectContaining({
      params: expect.objectContaining({ page: 2, page_size: 50, status: 'pending' }),
    }))
    expect(post).toHaveBeenCalledWith('/admin/affiliates/usage-rewards/7/review', { approve: true, remark: 'verified' })
  })
})
