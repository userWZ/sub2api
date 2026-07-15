/**
 * Admin Affiliate API endpoints
 * Manage per-user affiliate (邀请返利) configurations:
 * exclusive invite codes (overrides aff_code) and exclusive rebate rates.
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface AffiliateAdminEntry {
  user_id: number
  email: string
  username: string
  aff_code: string
  aff_code_custom: boolean
  aff_rebate_rate_percent?: number | null
  aff_count: number
}

export interface ListAffiliateUsersParams {
  page?: number
  page_size?: number
  search?: string
}

export interface ListAffiliateRecordsParams {
  page?: number
  page_size?: number
  search?: string
  start_at?: string
  end_at?: string
  action?: AffiliateWalletAction | ''
  sort_by?: string
  sort_order?: 'asc' | 'desc'
  timezone?: string
}

export type AffiliateWalletAction = 'discount' | 'withdraw'

export interface AffiliateInviteRecord {
  inviter_id: number
  inviter_email: string
  inviter_username: string
  invitee_id: number
  invitee_email: string
  invitee_username: string
  aff_code: string
  total_rebate: number
  created_at: string
}

export interface AffiliateRebateRecord {
  order_id: number
  out_trade_no: string
  inviter_id: number
  inviter_email: string
  inviter_username: string
  invitee_id: number
  invitee_email: string
  invitee_username: string
  order_amount: number
  pay_amount: number
  rebate_amount: number
  payment_type: string
  order_status: string
  created_at: string
}

export interface AffiliateTransferRecord {
  ledger_id: number
  user_id: number
  user_email: string
  username: string
  action: AffiliateWalletAction | string
  amount: number
  source_order_id?: number | null
  out_trade_no?: string | null
  operator_user_id?: number | null
  operator_email?: string | null
  remark?: string | null
  external_ref?: string | null
  available_quota_after?: number | null
  frozen_quota_after?: number | null
  history_quota_after?: number | null
  snapshot_available: boolean
  created_at: string
}

export interface AffiliateUserOverview {
  user_id: number
  email: string
  username: string
  aff_code: string
  rebate_rate_percent: number
  invited_count: number
  rebated_invitee_count: number
  available_quota: number
  history_quota: number
}

export interface AffiliateWithdrawRequest {
  amount: number
  remark?: string
  external_ref?: string
}

export interface AffiliateWithdrawResult {
  user_id: number
  amount: number
  available_quota_after: number
  frozen_quota_after: number
  history_quota_after: number
}

export interface AffiliateSettings {
  affiliate_enabled: boolean
  rebate_rate: number
  rebate_freeze_hours: number
  rebate_duration_days: number
  rebate_per_invitee_cap: number
  discount_enabled: boolean
  discount_max_percent: number
  discount_min_pay_amount: number
  usage_reward_enabled: boolean
  usage_reward_amount: number
  reward_inviter_daily_limit: number
  reward_inviter_30d_limit: number
  reward_ip_daily_limit: number
  reward_review_same_ip: boolean
  reward_review_missing_ip: boolean
}

export type AffiliateUsageRewardStatus = 'pending' | 'granted' | 'rejected'

export interface AffiliateUsageRewardRecord {
  id: number
  inviter_id: number
  inviter_email: string
  inviter_username: string
  invitee_id: number
  invitee_email: string
  invitee_username: string
  amount: number
  status: AffiliateUsageRewardStatus
  risk_reason: string
  trigger_request_id: string
  trigger_ip?: string | null
  trigger_actual_cost: number
  reviewer_user_id?: number | null
  reviewer_email?: string | null
  review_remark: string
  reviewed_at?: string | null
  created_at: string
}

export interface ListAffiliateUsageRewardsParams {
  page?: number
  page_size?: number
  search?: string
  status?: AffiliateUsageRewardStatus | ''
  start_at?: string
  end_at?: string
  timezone?: string
}

export interface UpdateAffiliateUserRequest {
  aff_code?: string
  aff_rebate_rate_percent?: number | null
  /** Set true to explicitly clear the per-user rate (sets it to NULL). */
  clear_rebate_rate?: boolean
}

export interface BatchSetRateRequest {
  user_ids: number[]
  aff_rebate_rate_percent?: number | null
  /** Set true to clear rates instead of setting. */
  clear?: boolean
}

export interface SimpleUser {
  id: number
  email: string
  username: string
}

export async function listUsers(
  params: ListAffiliateUsersParams = {},
): Promise<PaginatedResponse<AffiliateAdminEntry>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateAdminEntry>>(
    '/admin/affiliates/users',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
      },
    },
  )
  return data
}

export async function lookupUsers(q: string): Promise<SimpleUser[]> {
  const { data } = await apiClient.get<SimpleUser[]>(
    '/admin/affiliates/users/lookup',
    { params: { q } },
  )
  return data
}

export async function updateUserSettings(
  userId: number,
  payload: UpdateAffiliateUserRequest,
): Promise<{ user_id: number }> {
  const { data } = await apiClient.put<{ user_id: number }>(
    `/admin/affiliates/users/${userId}`,
    payload,
  )
  return data
}

export async function clearUserSettings(
  userId: number,
): Promise<{ user_id: number }> {
  const { data } = await apiClient.delete<{ user_id: number }>(
    `/admin/affiliates/users/${userId}`,
  )
  return data
}

export async function batchSetRate(
  payload: BatchSetRateRequest,
): Promise<{ affected: number }> {
  const { data } = await apiClient.post<{ affected: number }>(
    '/admin/affiliates/users/batch-rate',
    payload,
  )
  return data
}

function recordParams(params: ListAffiliateRecordsParams = {}) {
  return {
    page: params.page ?? 1,
    page_size: params.page_size ?? 20,
    search: params.search ?? '',
    start_at: params.start_at || undefined,
    end_at: params.end_at || undefined,
    action: params.action || undefined,
    sort_by: params.sort_by || undefined,
    sort_order: params.sort_order || undefined,
    timezone: params.timezone || undefined,
  }
}

export async function listInviteRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateInviteRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateInviteRecord>>(
    '/admin/affiliates/invites',
    { params: recordParams(params) },
  )
  return data
}

export async function listRebateRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateRebateRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateRebateRecord>>(
    '/admin/affiliates/rebates',
    { params: recordParams(params) },
  )
  return data
}

export async function listTransferRecords(
  params: ListAffiliateRecordsParams = {},
): Promise<PaginatedResponse<AffiliateTransferRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateTransferRecord>>(
    '/admin/affiliates/transfers',
    { params: recordParams(params) },
  )
  return data
}

export async function getUserOverview(
  userId: number,
): Promise<AffiliateUserOverview> {
  const { data } = await apiClient.get<AffiliateUserOverview>(
    `/admin/affiliates/users/${userId}/overview`,
  )
  return data
}

export async function withdrawAffiliateQuota(
  userId: number,
  payload: AffiliateWithdrawRequest,
): Promise<AffiliateWithdrawResult> {
  const { data } = await apiClient.post<AffiliateWithdrawResult>(
    `/admin/affiliates/users/${userId}/withdraw`,
    payload,
  )
  return data
}

export async function getSettings(): Promise<AffiliateSettings> {
  const { data } = await apiClient.get<AffiliateSettings>('/admin/affiliates/settings')
  return data
}

export async function updateSettings(payload: AffiliateSettings): Promise<AffiliateSettings> {
  const { data } = await apiClient.put<AffiliateSettings>('/admin/affiliates/settings', payload)
  return data
}

export async function listUsageRewards(
  params: ListAffiliateUsageRewardsParams = {},
): Promise<PaginatedResponse<AffiliateUsageRewardRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<AffiliateUsageRewardRecord>>(
    '/admin/affiliates/usage-rewards',
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
        search: params.search ?? '',
        status: params.status || undefined,
        start_at: params.start_at || undefined,
        end_at: params.end_at || undefined,
        timezone: params.timezone || undefined,
      },
    },
  )
  return data
}

export async function reviewUsageReward(
  rewardId: number,
  payload: { approve: boolean; remark?: string },
): Promise<{ reward_id: number; inviter_id: number; status: AffiliateUsageRewardStatus; amount: number }> {
  const { data } = await apiClient.post<{
    reward_id: number
    inviter_id: number
    status: AffiliateUsageRewardStatus
    amount: number
  }>(`/admin/affiliates/usage-rewards/${rewardId}/review`, payload)
  return data
}

export const affiliatesAPI = {
  listUsers,
  lookupUsers,
  updateUserSettings,
  clearUserSettings,
  batchSetRate,
  listInviteRecords,
  listRebateRecords,
  listTransferRecords,
  getUserOverview,
  withdrawAffiliateQuota,
  getSettings,
  updateSettings,
  listUsageRewards,
  reviewUsageReward,
}

export default affiliatesAPI
