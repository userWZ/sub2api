import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import enAdminOverview from '@/i18n/locales/en/admin/overview'
import enCommon from '@/i18n/locales/en/common'
import zhAdminOverview from '@/i18n/locales/zh/admin/overview'
import zhCommon from '@/i18n/locales/zh/common'
import AdminAffiliateRecordsTable from '../AdminAffiliateRecordsTable.vue'

const { listTransferRecords, showError, showSuccess } = vi.hoisted(() => ({
  listTransferRecords: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/affiliates', () => {
  const affiliatesAPI = {
    listInviteRecords: vi.fn(),
    listRebateRecords: vi.fn(),
    listTransferRecords,
    getUserOverview: vi.fn(),
    withdrawAffiliateQuota: vi.fn(),
  }
  return { affiliatesAPI, default: affiliatesAPI }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

function getPath(source: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((current, part) => {
    if (!current || typeof current !== 'object') return undefined
    return (current as Record<string, unknown>)[part]
  }, source)
}

function zhTranslate(key: string, fallbackOrParams?: unknown): string {
  const path = key.startsWith('admin.') ? key.slice('admin.'.length) : key
  const source = key.startsWith('admin.') ? zhAdminOverview : zhCommon
  const translated = getPath(source as Record<string, unknown>, path)
  if (typeof translated === 'string') {
    if (fallbackOrParams && typeof fallbackOrParams === 'object') {
      return Object.entries(fallbackOrParams as Record<string, unknown>).reduce(
        (result, [name, value]) => result.replace(`{${name}}`, String(value)),
        translated,
      )
    }
    return translated
  }
  return typeof fallbackOrParams === 'string' ? fallbackOrParams : key
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: zhTranslate }),
  }
})

const affiliateKeys = [
  'affiliates.records.allActions',
  'affiliates.records.action',
  'affiliates.records.actionDiscount',
  'affiliates.records.actionWithdraw',
  'affiliates.records.walletAmount',
  'affiliates.records.sourceOrder',
  'affiliates.records.availableQuotaAfter',
  'affiliates.records.frozenQuotaAfter',
  'affiliates.records.historyQuotaAfter',
  'affiliates.records.operator',
  'affiliates.records.externalRef',
  'affiliates.records.remark',
  'affiliates.records.transferredAt',
  'affiliates.withdraw.openButton',
  'affiliates.withdraw.title',
  'affiliates.withdraw.amount',
  'affiliates.withdraw.all',
  'affiliates.withdraw.externalRef',
  'affiliates.withdraw.externalRefPlaceholder',
  'affiliates.withdraw.remark',
  'affiliates.withdraw.remarkPlaceholder',
  'affiliates.withdraw.afterQuota',
  'affiliates.withdraw.confirm',
  'affiliates.withdraw.invalidAmount',
  'affiliates.withdraw.success',
]

describe('admin affiliate wallet records translations', () => {
  it.each([
    ['zh', zhAdminOverview, zhCommon],
    ['en', enAdminOverview, enCommon],
  ])('%s locale contains every wallet record and withdrawal key', (_locale, admin, common) => {
    for (const key of affiliateKeys) {
      expect(getPath(admin as Record<string, unknown>, key), key).toBeTypeOf('string')
    }
    expect(getPath(common as Record<string, unknown>, 'nav.affiliateTransferRecords')).toBeTypeOf('string')
  })

  it('does not describe wallet activity as a transfer into account balance', () => {
    expect(JSON.stringify(zhAdminOverview.affiliates)).not.toContain('转入账户余额')
    expect(JSON.stringify(enAdminOverview.affiliates)).not.toContain('into account balance')
  })
})

describe('AdminAffiliateRecordsTable', () => {
  beforeEach(() => {
    localStorage.clear()
    listTransferRecords.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    listTransferRecords.mockResolvedValue({
      items: [
        {
          ledger_id: 1,
          user_id: 3,
          user_email: 'user@example.com',
          username: 'user',
          action: 'discount',
          amount: 50,
          source_order_id: 101,
          out_trade_no: 'order-101',
          available_quota_after: 150,
          frozen_quota_after: 0,
          history_quota_after: 200,
          snapshot_available: true,
          created_at: '2026-07-14T00:00:00Z',
        },
        {
          ledger_id: 2,
          user_id: 3,
          user_email: 'user@example.com',
          username: 'user',
          action: 'withdraw',
          amount: 100,
          available_quota_after: 50,
          frozen_quota_after: 0,
          history_quota_after: 200,
          snapshot_available: true,
          created_at: '2026-07-14T01:00:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
      pages: 1,
    })
  })

  it('renders localized wallet headers and action labels without leaking i18n keys', async () => {
    const wrapper = mount(AdminAffiliateRecordsTable, {
      props: { type: 'transfers' },
      global: {
        stubs: {
          AppLayout: { template: '<main><slot /></main>' },
          TablePageLayout: {
            template: '<section><slot name="filters" /><slot name="table" /><slot name="pagination" /></section>',
          },
          DataTable: {
            props: ['columns', 'data'],
            template: `<div>
              <div class="headers"><span v-for="column in columns" :key="column.key">{{ column.label }}</span></div>
              <div v-for="row in data" :key="row.ledger_id"><slot name="cell-action" :row="row" /></div>
            </div>`,
          },
          Pagination: true,
          BaseDialog: true,
          Select: true,
          Icon: true,
          OrderStatusBadge: true,
        },
      },
    })

    await flushPromises()

    const text = wrapper.text()
    for (const label of [
      '变动类型',
      '变动金额',
      '关联订单',
      '操作后可用返利',
      '操作后冻结返利',
      '操作后累计返利',
      '操作管理员',
      '外部凭证',
      '备注',
      '操作时间',
      '订单抵扣',
      '管理员提现',
    ]) {
      expect(text).toContain(label)
    }
    expect(text).not.toContain('admin.affiliates')
    expect(text).not.toContain('ADMIN.AFFILIATES')
    expect(showError).not.toHaveBeenCalled()

    wrapper.unmount()
  })
})
