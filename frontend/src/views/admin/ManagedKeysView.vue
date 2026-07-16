<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="text-sm text-gray-500 dark:text-dark-400">
              共 {{ pagination.total }} 个托管用户
            </div>
          </div>
          <div class="flex flex-wrap items-center justify-end gap-2">
            <button
              type="button"
              class="btn btn-secondary px-2 md:px-3"
              :disabled="loading"
              title="刷新"
              @click="loadManagedKeys"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button type="button" class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-2" />
              新建托管用户
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="managedKeys" :loading="loading">
          <template #cell-customer="{ row: item }">
            <div class="min-w-52">
              <div class="font-medium text-gray-900 dark:text-white">
                {{ displayCustomerName(item) }}
              </div>
              <div class="mt-1 max-w-72 truncate text-xs text-gray-500 dark:text-dark-400">
                {{ managedContact(item) || item.user.email }}
              </div>
              <div class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                账户总额度 {{ formatPoints(item.user.balance) }}
              </div>
            </div>
          </template>

          <template #cell-api_key="{ row: item }">
            <div v-if="item.api_key" class="flex min-w-56 items-center gap-2">
              <code class="rounded-md bg-gray-100 px-2 py-1 font-mono text-xs text-gray-800 dark:bg-dark-700 dark:text-dark-100">
                {{ maskApiKey(item.api_key.key) }}
              </code>
              <button
                type="button"
                class="btn btn-ghost btn-sm px-2"
                title="复制 API Key"
                @click="copyText(item.api_key.key, `key-${item.api_key.id}`)"
              >
                <Icon :name="copiedField === `key-${item.api_key.id}` ? 'check' : 'copy'" size="sm" />
              </button>
            </div>
            <span v-else class="text-sm text-gray-400">未创建</span>
          </template>

          <template #cell-group="{ row: item }">
            <GroupBadge
              v-if="groupFor(item)"
              :name="groupFor(item)!.name"
              :platform="groupFor(item)!.platform"
              :subscription-type="groupFor(item)!.subscription_type"
              :rate-multiplier="groupFor(item)!.rate_multiplier"
            />
            <span v-else class="text-sm text-gray-400">未绑定</span>
          </template>

          <template #cell-concurrency="{ row: item }">
            <span class="font-medium text-gray-900 dark:text-white">{{ item.user.concurrency }}</span>
            <span v-if="item.user.rpm_limit" class="ml-2 text-xs text-gray-500 dark:text-dark-400">
              {{ item.user.rpm_limit }} RPM
            </span>
          </template>

          <template #cell-quota="{ row: item }">
            <div v-if="item.api_key" class="space-y-1 text-sm">
              <div>
                <span class="text-gray-900 dark:text-white">{{ formatPoints(item.api_key.quota_used) }}</span>
                <span class="text-gray-500 dark:text-dark-400">
                  / {{ item.api_key.quota > 0 ? formatPoints(item.api_key.quota) : '不限' }}
                </span>
              </div>
              <span :class="['badge text-xs', item.api_key.quota_disabled ? 'badge-gray' : 'badge-warning']">
                {{ item.api_key.quota_disabled ? 'Key 额度关闭' : 'Key 额度启用' }}
              </span>
            </div>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-window_usage="{ row: item }">
            <div v-if="item.api_key" class="min-w-44 space-y-1 text-xs text-gray-600 dark:text-dark-300">
              <div>{{ usageLine(item.api_key, '5h') }}</div>
              <div>{{ usageLine(item.api_key, '1d') }}</div>
              <div>{{ usageLine(item.api_key, '7d') }}</div>
            </div>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-expires_at="{ row: item }">
            <span class="text-sm text-gray-600 dark:text-dark-300">
              {{ item.api_key?.expires_at ? formatDateTime(item.api_key.expires_at) : '长期' }}
            </span>
          </template>

          <template #cell-last_used_at="{ row: item }">
            <span class="text-sm text-gray-600 dark:text-dark-300">
              {{ item.api_key?.last_used_at ? formatDateTime(item.api_key.last_used_at) : '从未' }}
            </span>
          </template>

          <template #cell-status="{ row: item }">
            <span :class="statusClass(item.api_key?.status)">
              {{ statusLabel(item.api_key?.status) }}
            </span>
          </template>

          <template #cell-actions="{ row: item }">
            <div v-if="item.api_key" class="flex items-center gap-1">
              <button
                type="button"
                class="btn btn-ghost btn-sm px-2"
                title="查看交付信息"
                @click="showDeliveryForExisting(item)"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-sm px-2"
                title="编辑托管 Key"
                @click="openEditDialog(item)"
              >
                <Icon name="edit" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              title="还没有托管用户"
              description="暂无记录"
              action-text="新建托管用户"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="loadManagedKeys"
          @update:page-size="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showCreateDialog"
      title="新建托管用户"
      width="wide"
      @close="closeCreateDialog"
    >
      <form class="space-y-5" @submit.prevent="submitCreate">
        <section class="space-y-4">
          <div>
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">客户信息</h4>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">用于内部识别和交付，不要求客户登录系统。</p>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">客户名称</span>
              <input v-model.trim="form.customer_name" class="input" required placeholder="例如 OceanWay 客户 A" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">联系方式</span>
              <input v-model.trim="form.contact" class="input" placeholder="微信、邮箱或备注名" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">Key 名称</span>
              <input v-model.trim="form.key_name" class="input" placeholder="留空则自动生成" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">优质额度分组</span>
              <Select v-model="form.group_id" :options="groupOptions" searchable />
            </label>
          </div>
        </section>

        <section class="space-y-4 border-t border-gray-100 pt-5 dark:border-dark-700">
          <div>
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">额度与有效期</h4>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">托管用户默认走优质分组，Key 到期后自动不可用。</p>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">并发数</span>
              <input v-model.number="form.concurrency" type="number" min="1" class="input" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">账户余额</span>
              <input v-model.number="form.balance" type="number" min="0" step="0.01" class="input" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">Key 额度</span>
              <input v-model.number="form.quota" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">每分钟限制</span>
              <input v-model.number="form.rpm_limit" type="number" min="0" class="input" placeholder="0 = 不限" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">有效天数</span>
              <input v-model.number="form.expires_in_days" type="number" min="1" class="input" required placeholder="默认 30 天" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">自定义 Key</span>
              <input v-model.trim="form.custom_key" class="input font-mono" placeholder="留空则自动生成" />
            </label>
          </div>
        </section>

        <section class="space-y-4 border-t border-gray-100 pt-5 dark:border-dark-700">
          <div>
            <h4 class="text-sm font-medium text-gray-900 dark:text-white">访问限制</h4>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">留空表示不限制。</p>
          </div>
          <div class="grid gap-4 md:grid-cols-3">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">5 小时限额</span>
              <input v-model.number="form.rate_limit_5h" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">每日限额</span>
              <input v-model.number="form.rate_limit_1d" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">7 日限额</span>
              <input v-model.number="form.rate_limit_7d" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
            </label>
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">IP 白名单</span>
              <textarea v-model="form.ip_whitelist" rows="3" class="input resize-none" placeholder="每行一个 IP 或 CIDR"></textarea>
            </label>
            <label class="block">
              <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">IP 黑名单</span>
              <textarea v-model="form.ip_blacklist" rows="3" class="input resize-none" placeholder="每行一个 IP 或 CIDR"></textarea>
            </label>
          </div>

          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">内部备注</span>
            <textarea v-model="form.notes" rows="3" class="input resize-none" placeholder="仅管理员可见"></textarea>
          </label>
        </section>

        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-300">
          {{ formError }}
        </div>
      </form>

      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="submitting" @click="closeCreateDialog">
          取消
        </button>
        <button type="button" class="btn btn-primary" :disabled="submitting || !form.group_id" @click="submitCreate">
          <Icon v-if="submitting" name="refresh" size="md" class="mr-2 animate-spin" />
          {{ submitting ? '创建中...' : '创建并生成 Key' }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showEditDialog"
      title="编辑托管 Key"
      width="normal"
      @close="closeEditDialog"
    >
      <form class="space-y-4" @submit.prevent="submitEdit">
        <section class="grid gap-4 md:grid-cols-2">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">Key 名称</span>
            <input v-model.trim="editForm.name" class="input" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">账户总额度</span>
            <input v-model.number="editForm.balance" type="number" min="0.01" step="0.01" class="input" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">优质额度分组</span>
            <Select v-model="editForm.group_id" :options="groupOptions" searchable />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">状态</span>
            <Select v-model="editForm.status" :options="statusOptions" />
          </label>
        </section>

        <section class="grid gap-4 md:grid-cols-2">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">Key 总额度</span>
            <input v-model.number="editForm.quota" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">有效期</span>
            <input v-model="editForm.expires_at" type="datetime-local" class="input" />
          </label>
        </section>

        <label class="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
          <span class="text-sm font-medium text-gray-700 dark:text-dark-200">启用 Key 自身额度</span>
          <input v-model="editForm.quota_enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
        </label>
        <label class="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
          <span class="text-sm font-medium text-gray-700 dark:text-dark-200">保存时重置 Key 已用额度</span>
          <input v-model="editForm.reset_quota" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
        </label>

        <section class="grid gap-4 md:grid-cols-3">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">5 小时限额</span>
            <input v-model.number="editForm.rate_limit_5h" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">每日限额</span>
            <input v-model.number="editForm.rate_limit_1d" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">7 日限额</span>
            <input v-model.number="editForm.rate_limit_7d" type="number" min="0" step="0.01" class="input" placeholder="0 = 不限" />
          </label>
        </section>

        <section class="grid gap-4 md:grid-cols-2">
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">IP 白名单</span>
            <textarea v-model="editForm.ip_whitelist" rows="3" class="input resize-none" placeholder="每行一个 IP 或 CIDR"></textarea>
          </label>
          <label class="block">
            <span class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">IP 黑名单</span>
            <textarea v-model="editForm.ip_blacklist" rows="3" class="input resize-none" placeholder="每行一个 IP 或 CIDR"></textarea>
          </label>
        </section>
        <div v-if="editError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-300">
          {{ editError }}
        </div>
      </form>

      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="submittingEdit" @click="closeEditDialog">
          取消
        </button>
        <button type="button" class="btn btn-primary" :disabled="submittingEdit || !editForm.group_id" @click="submitEdit">
          <Icon v-if="submittingEdit" name="refresh" size="md" class="mr-2 animate-spin" />
          {{ submittingEdit ? '保存中...' : '保存修改' }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="deliveryDialogVisible"
      title="交付信息"
      width="wide"
      @close="closeDeliveryDialog"
    >
      <div v-if="deliveryData" class="space-y-4">
        <div class="grid gap-3 md:grid-cols-2">
          <DeliveryField
            label="API Key"
            :value="deliveryData.delivery.api_key"
            field-id="delivery-key"
            :copied-field="copiedField"
            @copy="copyText"
          />
          <DeliveryField
            label="Authorization"
            :value="deliveryData.delivery.authorization_header"
            field-id="delivery-auth"
            :copied-field="copiedField"
            @copy="copyText"
          />
          <DeliveryField
            label="OpenAI / Claude Base URL"
            :value="deliveryData.delivery.openai_base_url"
            field-id="delivery-v1"
            :copied-field="copiedField"
            @copy="copyText"
          />
          <DeliveryField
            label="Gemini Base URL"
            :value="deliveryData.delivery.gemini_base_url"
            field-id="delivery-gemini"
            :copied-field="copiedField"
            @copy="copyText"
          />
        </div>
        <div class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-900/40 dark:text-dark-200">
          客户：{{ deliveryData.user.username || deliveryData.user.email }}，
          并发：{{ deliveryData.user.concurrency }}，
          Key 额度：{{ deliveryData.api_key.quota > 0 ? formatPoints(deliveryData.api_key.quota) : '不限' }}。
        </div>
      </div>

      <template #footer>
        <button type="button" class="btn btn-primary" @click="closeDeliveryDialog">
          完成
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { adminAPI } from '@/api/admin'
import type { CreateManagedKeyRequest, ManagedKey, ManagedKeyResponse, UpdateApiKeyPolicyRequest } from '@/api/admin/apiKeys'
import type { AdminGroup, ApiKey } from '@/types'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime, formatPoints } from '@/utils/format'
import { maskApiKey } from '@/utils/maskApiKey'

const appStore = useAppStore()
const managedPremiumGroupName = 'codex-managed-premium'

const managedKeys = ref<ManagedKey[]>([])
const groups = ref<AdminGroup[]>([])
const loading = ref(false)
const submitting = ref(false)
const submittingEdit = ref(false)
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const deliveryDialogVisible = ref(false)
const deliveryData = ref<ManagedKeyResponse | null>(null)
const formError = ref('')
const editError = ref('')
const copiedField = ref('')
const editingItem = ref<ManagedKey | null>(null)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

function createEmptyForm() {
  return {
    customer_name: '',
    contact: '',
    key_name: '',
    group_id: null as number | null,
    balance: 1000,
    concurrency: 1,
    rpm_limit: 0,
    quota: 0,
    expires_in_days: 30 as number | null,
    custom_key: '',
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    ip_whitelist: '',
    ip_blacklist: '',
    notes: ''
  }
}

const form = reactive(createEmptyForm())

function createEmptyEditForm() {
  return {
    name: '',
    balance: 1000,
    group_id: null as number | null,
    status: 'active' as 'active' | 'inactive',
    quota: 0,
    expires_at: '',
    quota_enabled: true,
    reset_quota: false,
    rate_limit_5h: 0,
    rate_limit_1d: 0,
    rate_limit_7d: 0,
    ip_whitelist: '',
    ip_blacklist: ''
  }
}

const editForm = reactive(createEmptyEditForm())

const managedKeyGroups = computed(() =>
  groups.value.filter((group) => group.status === 'active' && group.subscription_type !== 'subscription')
)

const defaultManagedGroup = computed(() => {
  return managedKeyGroups.value.find((group) => group.name === managedPremiumGroupName)
    || managedKeyGroups.value.find((group) => group.is_exclusive && group.allow_image_generation)
    || managedKeyGroups.value[0]
    || null
})

const groupOptions = computed(() =>
  managedKeyGroups.value.map((group) => ({
    value: group.id,
    label: `${group.name} · ${group.platform}${group.name === managedPremiumGroupName ? ' · 默认优质' : ''}`
  }))
)

const statusOptions = [
  { value: 'active', label: '启用' },
  { value: 'inactive', label: '停用' }
]

const columns: Column[] = [
  { key: 'customer', label: '客户', class: 'min-w-64' },
  { key: 'api_key', label: 'API Key', class: 'min-w-64' },
  { key: 'group', label: '分组', class: 'min-w-48' },
  { key: 'concurrency', label: '并发', class: 'min-w-24' },
  { key: 'quota', label: '额度', class: 'min-w-40' },
  { key: 'window_usage', label: '窗口用量', class: 'min-w-48' },
  { key: 'expires_at', label: '到期', class: 'min-w-40' },
  { key: 'last_used_at', label: '最后使用', class: 'min-w-40' },
  { key: 'status', label: '状态', class: 'min-w-28' },
  { key: 'actions', label: '操作', class: 'min-w-28' }
]

const groupById = computed(() => {
  const map = new Map<number, AdminGroup>()
  for (const group of groups.value) {
    map.set(group.id, group)
  }
  return map
})

async function loadGroups() {
  groups.value = await adminAPI.groups.getAll()
  ensureManagedGroupSelected()
}

async function loadManagedKeys() {
  loading.value = true
  try {
    const result = await adminAPI.apiKeys.listManagedKeys(pagination.page, pagination.page_size)
    managedKeys.value = result.items
    pagination.total = result.total
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '加载托管用户失败'))
  } finally {
    loading.value = false
  }
}

function handlePageSizeChange() {
  pagination.page = 1
  loadManagedKeys()
}

function openCreateDialog() {
  resetForm()
  ensureManagedGroupSelected()
  formError.value = ''
  showCreateDialog.value = true
}

function closeCreateDialog() {
  if (!submitting.value) {
    showCreateDialog.value = false
  }
}

function openEditDialog(item: ManagedKey) {
  if (!item.api_key) return
  editingItem.value = item
  Object.assign(editForm, createEmptyEditForm(), {
    name: item.api_key.name || '',
    balance: Math.max(0, Number(item.user.balance || 0)),
    group_id: item.api_key.group_id ?? defaultManagedGroup.value?.id ?? null,
    status: item.api_key.status === 'active' ? 'active' : 'inactive',
    quota: Math.max(0, Number(item.api_key.quota || 0)),
    expires_at: formatDateTimeLocalValue(item.api_key.expires_at),
    quota_enabled: !item.api_key.quota_disabled,
    reset_quota: false,
    rate_limit_5h: Math.max(0, Number(item.api_key.rate_limit_5h || 0)),
    rate_limit_1d: Math.max(0, Number(item.api_key.rate_limit_1d || 0)),
    rate_limit_7d: Math.max(0, Number(item.api_key.rate_limit_7d || 0)),
    ip_whitelist: (item.api_key.ip_whitelist || []).join('\n'),
    ip_blacklist: (item.api_key.ip_blacklist || []).join('\n')
  })
  editError.value = ''
  showEditDialog.value = true
}

function closeEditDialog() {
  if (!submittingEdit.value) {
    showEditDialog.value = false
    editingItem.value = null
  }
}

function closeDeliveryDialog() {
  deliveryDialogVisible.value = false
  deliveryData.value = null
}

function resetForm() {
  Object.assign(form, createEmptyForm())
}

function ensureManagedGroupSelected() {
  if (form.group_id && managedKeyGroups.value.some(group => group.id === form.group_id)) return
  form.group_id = defaultManagedGroup.value?.id || null
}

function splitLines(value: string): string[] {
  return value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
}

function optionalPositiveInt(value: number | null): number | null {
  const normalized = Number(value)
  return Number.isFinite(normalized) && normalized > 0 ? Math.floor(normalized) : null
}

function numericValue(value: number, fallback = 0): number {
  const normalized = Number(value)
  return Number.isFinite(normalized) ? normalized : fallback
}

function formatDateTimeLocalValue(value: string | null | undefined): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (part: number) => String(part).padStart(2, '0')
  const year = date.getFullYear()
  const month = pad(date.getMonth() + 1)
  const day = pad(date.getDate())
  const hours = pad(date.getHours())
  const minutes = pad(date.getMinutes())
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

function dateTimeLocalToISOString(value: string): string {
  return new Date(value).toISOString()
}

async function submitCreate() {
  formError.value = ''
  if (!form.customer_name.trim()) {
    formError.value = '请填写客户名称'
    return
  }
  if (!form.group_id) {
    formError.value = `请先创建并选择标准优质分组 ${managedPremiumGroupName}`
    return
  }

  submitting.value = true
  try {
    const payload: CreateManagedKeyRequest = {
      customer_name: form.customer_name.trim(),
      contact: form.contact.trim(),
      key_name: form.key_name.trim(),
      group_id: form.group_id,
      balance: numericValue(form.balance, 1000),
      concurrency: Math.max(1, Math.floor(numericValue(form.concurrency, 1))),
      rpm_limit: Math.max(0, Math.floor(numericValue(form.rpm_limit, 0))),
      quota: Math.max(0, numericValue(form.quota, 0)),
      expires_in_days: optionalPositiveInt(form.expires_in_days),
      custom_key: form.custom_key.trim() || null,
      rate_limit_5h: Math.max(0, numericValue(form.rate_limit_5h, 0)),
      rate_limit_1d: Math.max(0, numericValue(form.rate_limit_1d, 0)),
      rate_limit_7d: Math.max(0, numericValue(form.rate_limit_7d, 0)),
      ip_whitelist: splitLines(form.ip_whitelist),
      ip_blacklist: splitLines(form.ip_blacklist),
      notes: form.notes.trim()
    }

    const result = await adminAPI.apiKeys.createManagedKey(payload)
    deliveryData.value = result
    deliveryDialogVisible.value = true
    showCreateDialog.value = false
    appStore.showSuccess('托管用户已创建')
    await loadManagedKeys()
  } catch (err) {
    formError.value = extractApiErrorMessage(err, '创建托管用户失败')
  } finally {
    submitting.value = false
  }
}

async function submitEdit() {
  editError.value = ''
  const item = editingItem.value
  if (!item?.api_key) return
  if (!editForm.group_id) {
    editError.value = `请先选择优质额度分组 ${managedPremiumGroupName}`
    return
  }
  const nextBalance = Math.max(0, numericValue(editForm.balance, item.user.balance))
  if (nextBalance <= 0) {
    editError.value = '账户总额度必须大于 0'
    return
  }

  submittingEdit.value = true
  try {
    if (Math.abs(nextBalance - Number(item.user.balance || 0)) > 0.000001) {
      item.user = await adminAPI.users.updateBalance(item.user.id, nextBalance, 'set', '托管用户账户总额度调整')
    }
    let expiresAt: string | null = ''
    if (editForm.expires_at.trim()) {
      expiresAt = dateTimeLocalToISOString(editForm.expires_at)
    }
    const payload: UpdateApiKeyPolicyRequest = {
      group_id: editForm.group_id,
      name: editForm.name.trim(),
      status: editForm.status,
      quota_disabled: !editForm.quota_enabled,
      quota: Math.max(0, numericValue(editForm.quota, 0)),
      expires_at: expiresAt,
      reset_quota: editForm.reset_quota,
      rate_limit_5h: Math.max(0, numericValue(editForm.rate_limit_5h, 0)),
      rate_limit_1d: Math.max(0, numericValue(editForm.rate_limit_1d, 0)),
      rate_limit_7d: Math.max(0, numericValue(editForm.rate_limit_7d, 0)),
      ip_whitelist: splitLines(editForm.ip_whitelist),
      ip_blacklist: splitLines(editForm.ip_blacklist)
    }
    const result = await adminAPI.apiKeys.updateApiKeyPolicy(item.api_key.id, payload)
    item.api_key = result.api_key
    showEditDialog.value = false
    editingItem.value = null
    appStore.showSuccess('托管 Key 已更新')
    await loadManagedKeys()
  } catch (err) {
    editError.value = extractApiErrorMessage(err, '更新托管 Key 失败')
  } finally {
    submittingEdit.value = false
  }
}

function displayCustomerName(item: ManagedKey): string {
  return item.user.username || managedNoteValue(item, 'customer') || item.user.email
}

function managedContact(item: ManagedKey): string {
  return managedNoteValue(item, 'contact')
}

function managedNoteValue(item: ManagedKey, key: string): string {
  const prefix = `${key}:`
  const line = (item.user.notes || '')
    .split('\n')
    .find((part) => part.toLowerCase().startsWith(prefix))
  return line ? line.slice(prefix.length).trim() : ''
}

function groupFor(item: ManagedKey): AdminGroup | null {
  const apiKey = item.api_key
  if (!apiKey) return null
  if (apiKey.group) return apiKey.group as AdminGroup
  return apiKey.group_id ? groupById.value.get(apiKey.group_id) ?? null : null
}

function statusLabel(status?: string): string {
  const labels: Record<string, string> = {
    active: '活跃',
    inactive: '停用',
    quota_exhausted: '额度用尽',
    expired: '已过期'
  }
  return status ? labels[status] ?? status : '未知'
}

function statusClass(status?: string): string {
  const base = 'inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium'
  if (status === 'active') {
    return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300`
  }
  if (status === 'quota_exhausted' || status === 'expired') {
    return `${base} bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300`
  }
  return `${base} bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300`
}

function usageLine(apiKey: ApiKey | null, window: '5h' | '1d' | '7d'): string {
  if (!apiKey) return '-'
  const fields = {
    '5h': { label: '5h', used: apiKey.usage_5h, limit: apiKey.rate_limit_5h },
    '1d': { label: '日', used: apiKey.usage_1d, limit: apiKey.rate_limit_1d },
    '7d': { label: '7日', used: apiKey.usage_7d, limit: apiKey.rate_limit_7d }
  }[window]
  const limit = fields.limit > 0 ? formatPoints(fields.limit) : '不限'
  return `${fields.label}: ${formatPoints(fields.used || 0)} / ${limit}`
}

async function showDeliveryForExisting(item: ManagedKey) {
  if (!item.api_key) return
  try {
    deliveryData.value = await adminAPI.apiKeys.getManagedKeyDelivery(item.user.id)
    deliveryDialogVisible.value = true
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, '加载交付信息失败'))
  }
}

async function copyText(value: string, fieldId: string) {
  await navigator.clipboard.writeText(value)
  copiedField.value = fieldId
  window.setTimeout(() => {
    if (copiedField.value === fieldId) {
      copiedField.value = ''
    }
  }, 1400)
}

const DeliveryField = defineComponent({
  name: 'DeliveryField',
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    fieldId: { type: String, required: true },
    copiedField: { type: String, required: true }
  },
  emits: ['copy'],
  setup(props, { emit }) {
    return () =>
      h('div', { class: 'rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-800' }, [
        h('div', { class: 'mb-2 text-xs font-medium uppercase text-gray-500 dark:text-dark-400' }, props.label),
        h('div', { class: 'flex items-center gap-2' }, [
          h('code', { class: 'min-w-0 flex-1 truncate rounded-md bg-gray-100 px-2 py-1.5 font-mono text-xs text-gray-800 dark:bg-dark-700 dark:text-dark-100' }, props.value),
          h(
            'button',
            {
              type: 'button',
              class: 'btn btn-ghost btn-sm px-2',
              title: '复制',
              onClick: () => emit('copy', props.value, props.fieldId)
            },
            [h(Icon, { name: props.copiedField === props.fieldId ? 'check' : 'copy', size: 'sm' })]
          )
        ])
      ])
  }
})

onMounted(async () => {
  await Promise.all([loadGroups(), loadManagedKeys()])
})

watch(defaultManagedGroup, () => ensureManagedGroupSelected())
</script>
