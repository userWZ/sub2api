<template>
  <section class="card">
    <div class="flex flex-col gap-3 border-b border-gray-100 px-6 py-4 sm:flex-row sm:items-center sm:justify-between dark:border-dark-700">
      <div>
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('admin.settings.features.affiliate.customUsers.title') }}
        </h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.settings.features.affiliate.customUsers.description') }}
        </p>
      </div>
      <button type="button" class="btn btn-primary btn-sm whitespace-nowrap" @click="openEditor(null)">
        + {{ t('admin.settings.features.affiliate.customUsers.addButton') }}
      </button>
    </div>

    <div class="p-6">
      <div class="mb-4 flex flex-col gap-2 sm:flex-row">
        <input
          v-model="state.search"
          type="search"
          class="input flex-1"
          :placeholder="t('admin.settings.features.affiliate.customUsers.searchPlaceholder')"
          @input="queueListSearch"
        />
        <button
          v-if="state.selected.length > 0"
          type="button"
          class="btn btn-secondary btn-sm whitespace-nowrap"
          @click="openBatchEditor"
        >
          {{ t('admin.settings.features.affiliate.customUsers.batchButton', { count: state.selected.length }) }}
        </button>
      </div>

      <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
        <table class="min-w-[760px] divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-800">
            <tr>
              <th class="w-10 px-3 py-2 text-left">
                <input
                  type="checkbox"
                  :checked="allVisibleSelected"
                  :aria-label="t('common.selectAll')"
                  @change="toggleAll"
                />
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.email') }}</th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.username') }}</th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.code') }}</th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.rate') }}</th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.col.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
            <tr v-if="state.loading">
              <td colspan="6" class="px-3 py-8 text-center text-sm text-gray-500">{{ t('common.loading') }}</td>
            </tr>
            <tr v-else-if="state.entries.length === 0">
              <td colspan="6" class="px-3 py-8 text-center text-sm text-gray-500">
                {{ t('admin.settings.features.affiliate.customUsers.empty') }}
              </td>
            </tr>
            <tr v-for="entry in state.entries" v-else :key="entry.user_id">
              <td class="px-3 py-2">
                <input
                  type="checkbox"
                  :checked="state.selected.includes(entry.user_id)"
                  :aria-label="entry.email"
                  @change="toggleOne(entry.user_id)"
                />
              </td>
              <td class="px-3 py-2 text-sm text-gray-900 dark:text-white">{{ entry.email }}</td>
              <td class="px-3 py-2 text-sm text-gray-600 dark:text-gray-300">{{ entry.username }}</td>
              <td class="px-3 py-2 text-sm font-mono text-gray-700 dark:text-gray-200">
                {{ entry.aff_code }}
                <span
                  v-if="entry.aff_code_custom"
                  class="ml-1 inline-block rounded bg-primary-100 px-1.5 py-0.5 text-[10px] font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                >
                  {{ t('admin.settings.features.affiliate.customUsers.customBadge') }}
                </span>
              </td>
              <td class="px-3 py-2 text-sm">
                <span v-if="entry.aff_rebate_rate_percent != null">{{ entry.aff_rebate_rate_percent }}%</span>
                <span v-else class="text-gray-400">{{ t('admin.settings.features.affiliate.customUsers.useGlobal') }}</span>
              </td>
              <td class="px-3 py-2 text-sm">
                <div class="flex items-center gap-3">
                  <button type="button" class="text-primary-600 hover:underline dark:text-primary-400" @click="openEditor(entry)">
                    {{ t('common.edit') }}
                  </button>
                  <button type="button" class="text-red-600 hover:underline dark:text-red-400" @click="askReset(entry)">
                    {{ t('common.delete') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="state.total > state.pageSize" class="mt-4 flex flex-col gap-3 text-sm sm:flex-row sm:items-center sm:justify-between">
        <span class="text-gray-500">{{ t('admin.settings.features.affiliate.customUsers.totalLabel', { total: state.total }) }}</span>
        <div class="flex items-center gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="state.page <= 1" @click="changePage(state.page - 1)">
            {{ t('pagination.previous') }}
          </button>
          <span class="text-gray-500">{{ state.page }} / {{ pageCount }}</span>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="state.page >= pageCount" @click="changePage(state.page + 1)">
            {{ t('pagination.next') }}
          </button>
        </div>
      </div>
    </div>
  </section>

  <div v-if="editor.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="closeEditor">
    <div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-dark-900">
      <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
        {{ editor.mode === 'add' ? t('admin.settings.features.affiliate.modal.addTitle') : t('admin.settings.features.affiliate.modal.editTitle') }}
      </h3>
      <div class="space-y-4">
        <div v-if="editor.mode === 'add'">
          <label class="input-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
          <div v-if="editor.selectedUser" class="flex items-center justify-between rounded-md border border-primary-200 bg-primary-50 px-3 py-2 dark:border-primary-700/50 dark:bg-primary-900/20">
            <div class="min-w-0 text-sm">
              <span class="font-medium text-gray-900 dark:text-white">{{ editor.selectedUser.email }}</span>
              <span class="ml-1 text-xs text-gray-500">({{ editor.selectedUser.username }})</span>
            </div>
            <button type="button" class="ml-2 text-xl leading-none text-gray-400 hover:text-red-600" :title="t('admin.settings.features.affiliate.modal.changeUser')" @click="clearSelectedUser">×</button>
          </div>
          <template v-else>
            <input
              v-model="editor.userQuery"
              type="search"
              class="input"
              :placeholder="t('admin.settings.features.affiliate.modal.userPlaceholder')"
              @input="queueUserSearch"
            />
            <div v-if="editor.userResults.length" class="mt-1 max-h-44 overflow-y-auto rounded border border-gray-200 dark:border-dark-700">
              <button
                v-for="user in editor.userResults"
                :key="user.id"
                type="button"
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-800"
                @click="selectUser(user)"
              >
                {{ user.email }} <span class="text-xs text-gray-500">({{ user.username }})</span>
              </button>
            </div>
          </template>
        </div>
        <div v-else>
          <label class="input-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
          <input type="text" class="input" :value="editor.editingEntry?.email || ''" disabled />
        </div>

        <label class="block">
          <span class="input-label">{{ t('admin.settings.features.affiliate.modal.codeLabel') }}</span>
          <input v-model="editor.code" type="text" class="input font-mono" :placeholder="t('admin.settings.features.affiliate.modal.codePlaceholder')" maxlength="32" />
          <span class="mt-1 block text-xs text-gray-400">{{ t('admin.settings.features.affiliate.modal.codeHint') }}</span>
        </label>

        <label class="block">
          <span class="input-label">{{ t('admin.settings.features.affiliate.modal.rateLabel') }}</span>
          <div class="relative">
            <input v-model="editor.rate" type="number" min="0" max="100" step="0.01" class="input pr-8" :placeholder="t('admin.settings.features.affiliate.modal.ratePlaceholder')" />
            <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">%</span>
          </div>
          <span class="mt-1 block text-xs text-gray-400">{{ t('admin.settings.features.affiliate.modal.rateHint') }}</span>
        </label>
      </div>

      <div class="mt-6 flex items-center justify-between gap-3">
        <p v-if="!canSubmit" class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.features.affiliate.modal.errorEmpty') }}</p>
        <span v-else></span>
        <div class="flex gap-2">
          <button type="button" class="btn btn-secondary" @click="closeEditor">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="editor.saving || !canSubmit" @click="submitEditor">
            {{ editor.saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>

  <div v-if="batch.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" @click.self="batch.open = false">
    <div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-dark-900">
      <h3 class="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.features.affiliate.batchModal.title', { count: state.selected.length }) }}
      </h3>
      <p class="mb-4 text-sm text-gray-500">{{ t('admin.settings.features.affiliate.batchModal.hint') }}</p>
      <div class="relative">
        <input v-model="batch.rate" type="number" min="0" max="100" step="0.01" class="input pr-8" :placeholder="t('admin.settings.features.affiliate.batchModal.placeholder')" />
        <span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400">%</span>
      </div>
      <p class="mt-2 text-xs text-gray-400">{{ t('admin.settings.features.affiliate.batchModal.clearHint') }}</p>
      <div class="mt-6 flex justify-end gap-2">
        <button type="button" class="btn btn-secondary" @click="batch.open = false">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="batch.saving" @click="submitBatch">
          {{ batch.saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </div>
  </div>

  <ConfirmDialog
    :show="confirmation.show"
    :title="confirmation.title"
    :message="confirmation.message"
    :confirm-text="confirmation.confirmText"
    danger
    @confirm="confirmReset"
    @cancel="cancelReset"
  />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import {
  affiliatesAPI,
  type AffiliateAdminEntry,
  type SimpleUser,
} from '@/api/admin/affiliates'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const state = reactive({
  loading: false,
  entries: [] as AffiliateAdminEntry[],
  total: 0,
  page: 1,
  pageSize: 20,
  search: '',
  selected: [] as number[],
})

const editor = reactive({
  open: false,
  mode: 'add' as 'add' | 'edit',
  saving: false,
  userQuery: '',
  userResults: [] as SimpleUser[],
  selectedUser: null as SimpleUser | null,
  editingEntry: null as AffiliateAdminEntry | null,
  code: '',
  rate: '' as string | number,
})

const batch = reactive({ open: false, saving: false, rate: '' as string | number })
const confirmation = reactive({
  show: false,
  title: '',
  message: '',
  confirmText: '',
  userId: 0,
})

let listSearchTimer: number | undefined
let userSearchTimer: number | undefined

const pageCount = computed(() => Math.max(1, Math.ceil(state.total / state.pageSize)))
const allVisibleSelected = computed(() => state.entries.length > 0 && state.entries.every(entry => state.selected.includes(entry.user_id)))
const canSubmit = computed(() => {
  if (editor.mode === 'add' ? !editor.selectedUser : !editor.editingEntry) return false
  if (editor.code.trim() || String(editor.rate ?? '').trim()) return true
  return editor.mode === 'edit' && editor.editingEntry?.aff_rebate_rate_percent != null
})

function showError(error: unknown) {
  appStore.showError(extractApiErrorMessage(error, t('common.error')))
}

function parseRate(raw: unknown): number | null | undefined {
  const value = String(raw ?? '').trim()
  if (!value) return null
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0 || parsed > 100) {
    appStore.showError(t('admin.settings.features.affiliate.modal.errorBadRate'))
    return undefined
  }
  return parsed
}

async function loadUsers() {
  state.loading = true
  try {
    const result = await affiliatesAPI.listUsers({ page: state.page, page_size: state.pageSize, search: state.search })
    state.entries = result.items ?? []
    state.total = result.total ?? 0
    const visible = new Set(state.entries.map(entry => entry.user_id))
    state.selected = state.selected.filter(id => visible.has(id))
  } catch (error) {
    showError(error)
  } finally {
    state.loading = false
  }
}

function queueListSearch() {
  window.clearTimeout(listSearchTimer)
  listSearchTimer = window.setTimeout(() => {
    state.page = 1
    void loadUsers()
  }, 300)
}

function changePage(page: number) {
  if (page < 1 || page > pageCount.value) return
  state.page = page
  void loadUsers()
}

function toggleAll(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  state.selected = checked ? state.entries.map(entry => entry.user_id) : []
}

function toggleOne(userId: number) {
  const index = state.selected.indexOf(userId)
  if (index >= 0) state.selected.splice(index, 1)
  else state.selected.push(userId)
}

function openEditor(entry: AffiliateAdminEntry | null) {
  editor.open = true
  editor.mode = entry ? 'edit' : 'add'
  editor.userQuery = ''
  editor.userResults = []
  editor.selectedUser = null
  editor.editingEntry = entry
  editor.code = entry?.aff_code_custom ? entry.aff_code : ''
  editor.rate = entry?.aff_rebate_rate_percent != null ? entry.aff_rebate_rate_percent : ''
}

function closeEditor() {
  editor.open = false
  window.clearTimeout(userSearchTimer)
}

function queueUserSearch() {
  const query = editor.userQuery.trim()
  window.clearTimeout(userSearchTimer)
  if (!query) {
    editor.userResults = []
    return
  }
  userSearchTimer = window.setTimeout(async () => {
    try {
      editor.userResults = await affiliatesAPI.lookupUsers(query)
    } catch (error) {
      showError(error)
    }
  }, 300)
}

function selectUser(user: SimpleUser) {
  editor.selectedUser = user
  editor.userQuery = ''
  editor.userResults = []
}

function clearSelectedUser() {
  editor.selectedUser = null
}

async function submitEditor() {
  if (!canSubmit.value) return
  const userId = editor.mode === 'add' ? editor.selectedUser!.id : editor.editingEntry!.user_id
  const payload: Parameters<typeof affiliatesAPI.updateUserSettings>[1] = {}
  const code = editor.code.trim()
  if (code) payload.aff_code = code.toUpperCase()
  const rate = parseRate(editor.rate)
  if (rate === undefined) return
  if (rate === null) {
    if (editor.mode === 'edit' && editor.editingEntry?.aff_rebate_rate_percent != null) payload.clear_rebate_rate = true
  } else {
    payload.aff_rebate_rate_percent = rate
  }

  editor.saving = true
  try {
    await affiliatesAPI.updateUserSettings(userId, payload)
    appStore.showSuccess(t('common.saved'))
    closeEditor()
    state.page = 1
    await loadUsers()
  } catch (error) {
    showError(error)
  } finally {
    editor.saving = false
  }
}

function askReset(entry: AffiliateAdminEntry) {
  confirmation.show = true
  confirmation.title = t('admin.settings.features.affiliate.customUsers.resetTitle')
  confirmation.message = t('admin.settings.features.affiliate.customUsers.resetMessage', { email: entry.email || `#${entry.user_id}` })
  confirmation.confirmText = t('common.delete')
  confirmation.userId = entry.user_id
}

function cancelReset() {
  confirmation.show = false
  confirmation.userId = 0
}

async function confirmReset() {
  const userId = confirmation.userId
  cancelReset()
  if (!userId) return
  try {
    await affiliatesAPI.clearUserSettings(userId)
    appStore.showSuccess(t('common.saved'))
    await loadUsers()
  } catch (error) {
    showError(error)
  }
}

function openBatchEditor() {
  if (!state.selected.length) return
  batch.open = true
  batch.rate = ''
}

async function submitBatch() {
  const rate = parseRate(batch.rate)
  if (rate === undefined) return
  const payload: Parameters<typeof affiliatesAPI.batchSetRate>[0] = rate === null
    ? { user_ids: [...state.selected], clear: true }
    : { user_ids: [...state.selected], aff_rebate_rate_percent: rate }
  batch.saving = true
  try {
    await affiliatesAPI.batchSetRate(payload)
    appStore.showSuccess(t('common.saved'))
    batch.open = false
    state.selected = []
    await loadUsers()
  } catch (error) {
    showError(error)
  } finally {
    batch.saving = false
  }
}

onMounted(loadUsers)
onBeforeUnmount(() => {
  window.clearTimeout(listSearchTimer)
  window.clearTimeout(userSearchTimer)
})
</script>
