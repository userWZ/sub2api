/**
 * Admin API Keys API endpoints
 * Handles API key management for administrators
 */

import { apiClient } from '../client'
import type { ApiKey } from '@/types'

export interface UpdateApiKeyGroupResult {
  api_key: ApiKey
  auto_granted_group_access: boolean
  granted_group_id?: number
  granted_group_name?: string
}

export interface UpdateApiKeyPolicyRequest {
  group_id?: number | null
  status?: 'active' | 'inactive'
  quota_disabled?: boolean
  reset_rate_limit_usage?: boolean
}

/**
 * Update an API key's group binding
 * @param id - API Key ID
 * @param groupId - Group ID (0 to unbind, positive to bind, null/undefined to skip)
 * @returns Updated API key with auto-grant info
 */
export async function updateApiKeyGroup(id: number, groupId: number | null): Promise<UpdateApiKeyGroupResult> {
  const { data } = await apiClient.put<UpdateApiKeyGroupResult>(`/admin/api-keys/${id}`, {
    group_id: groupId === null ? 0 : groupId
  })
  return data
}

export async function updateApiKeyPolicy(id: number, payload: UpdateApiKeyPolicyRequest): Promise<UpdateApiKeyGroupResult> {
  const body = { ...payload }
  if (Object.prototype.hasOwnProperty.call(body, 'group_id') && body.group_id === null) {
    body.group_id = 0
  }
  const { data } = await apiClient.put<UpdateApiKeyGroupResult>(`/admin/api-keys/${id}`, body)
  return data
}

export const apiKeysAPI = {
  updateApiKeyGroup,
  updateApiKeyPolicy
}

export default apiKeysAPI
