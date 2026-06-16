<template>
  <AppLayout>
    <div class="image-shell">
      <aside class="image-sidebar">
        <div class="sidebar-actions">
          <button type="button" class="new-chat-button" @click="createDraftConversation">
            <span>+</span>
            {{ copy.newChat }}
          </button>
          <button type="button" class="icon-button" :disabled="!conversations.length" :title="copy.clearHistory" @click="clearHistory">
            <span aria-hidden="true">×</span>
          </button>
        </div>

        <div class="key-strip">
          <span>{{ copy.defaultKey }}</span>
          <strong>{{ selectedKey ? maskKey(selectedKey.key) : copy.noKey }}</strong>
          <button v-if="selectedKey" type="button" @click="copyDefaultKey">{{ copy.copyKey }}</button>
          <button v-else type="button" :disabled="keysLoading || creatingKey" @click="createDefaultKey">
            {{ creatingKey ? copy.creatingKey : copy.createKey }}
          </button>
        </div>

        <div class="history-list">
          <div v-if="historyLoading" class="history-empty">
            <LoadingSpinner />
            <span>{{ copy.loadingHistory }}</span>
          </div>
          <div v-else-if="!conversations.length" class="history-empty">{{ copy.noHistory }}</div>
          <button
            v-for="conversation in conversations"
            v-else
            :key="conversation.id"
            type="button"
            :class="['history-item', { active: conversation.id === activeConversationId }]"
            @click="selectConversation(conversation.id)"
          >
            <span class="history-title">{{ conversation.title }}</span>
            <span class="history-meta">{{ conversation.turns.length }} {{ copy.rounds }} · {{ formatTime(conversation.updatedAt) }}</span>
            <span v-if="conversationHasRunningTurn(conversation)" class="history-running">{{ copy.running }}</span>
          </button>
        </div>
      </aside>

      <section class="image-main">
        <div ref="resultsEl" class="result-stream">
          <div v-if="!activeConversation || activeConversation.turns.length === 0" class="welcome-state">
            <h1>{{ copy.heroTitle }}</h1>
            <p>{{ copy.heroDescription }}</p>
          </div>

          <div v-else class="turn-stack">
            <article v-for="(turn, turnIndex) in activeConversation.turns" :key="turn.id" class="turn">
              <div class="prompt-row">
                <div class="prompt-bubble">
                  <div class="turn-meta">
                    <span>{{ copy.round }} {{ turnIndex + 1 }}</span>
                    <span>{{ turn.model }}</span>
                    <span>{{ turn.size }}</span>
                    <span>{{ statusLabel(turn.status) }}</span>
                  </div>
                  <p>{{ turn.prompt }}</p>
                  <div class="turn-actions">
                    <button type="button" @click="reuseTurn(turn)">{{ copy.reuse }}</button>
                    <button type="button" @click="deleteTurn(activeConversation!.id, turn.id)">{{ copy.delete }}</button>
                  </div>
                </div>
              </div>

              <div class="result-row">
                <div class="result-bubble">
                  <div v-if="turn.status === 'generating'" class="loading-result">
                    <LoadingSpinner />
                    <span>{{ copy.waiting }}</span>
                  </div>
                  <div v-else-if="turn.status === 'error'" class="error-result">
                    <strong>{{ copy.failed }}</strong>
                    <p>{{ turn.error || copy.generateFailed }}</p>
                    <button type="button" @click="regenerateTurn(turn)">{{ copy.retry }}</button>
                  </div>
                  <div v-else class="generated-grid">
                    <figure v-for="(image, index) in turn.images" :key="`${turn.id}-${index}`" class="generated-card">
                      <button type="button" class="image-preview" @click="previewImage = image.src">
                        <img :src="image.src" :alt="copy.imageAlt.replace('{index}', String(index + 1))" loading="lazy" decoding="async" />
                      </button>
                      <figcaption>
                        <span>{{ copy.result }} {{ index + 1 }}</span>
                        <a :href="image.src" :download="downloadName(turn, index)">{{ copy.download }}</a>
                      </figcaption>
                      <p v-if="image.revised_prompt" class="revised-prompt">{{ image.revised_prompt }}</p>
                    </figure>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </div>

        <form class="composer-wrap" @submit.prevent="runGeneration">
          <div class="composer-card">
            <textarea
              v-model="form.prompt"
              :placeholder="copy.promptPlaceholder"
              rows="4"
              @keydown.enter.exact.prevent="runGeneration"
            />

            <div class="composer-toolbar">
              <div class="pill-group">
                <select v-model="form.model" :aria-label="copy.model">
                  <option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option>
                </select>
                <select v-model="form.size" :aria-label="copy.size">
                  <option value="1024x1024">1:1</option>
                  <option value="1024x1536">2:3</option>
                  <option value="1536x1024">3:2</option>
                  <option value="auto">{{ copy.auto }}</option>
                </select>
                <select v-model="form.quality" :aria-label="copy.quality">
                  <option value="auto">{{ copy.auto }}</option>
                  <option value="low">{{ copy.low }}</option>
                  <option value="medium">{{ copy.medium }}</option>
                  <option value="high">{{ copy.high }}</option>
                </select>
                <select v-model.number="form.n" :aria-label="copy.count">
                  <option v-for="count in countOptions" :key="count" :value="count">{{ count }} {{ copy.images }}</option>
                </select>
                <select v-model="form.output_format" :aria-label="copy.format">
                  <option value="png">PNG</option>
                  <option value="jpeg">JPEG</option>
                  <option value="webp">WebP</option>
                </select>
              </div>

              <button class="send-button" type="submit" :disabled="!canGenerate">
                <span v-if="generating" aria-hidden="true">…</span>
                <span v-else aria-hidden="true">↑</span>
                <span class="sr-only">{{ copy.generate }}</span>
              </button>
            </div>
          </div>
          <p v-if="errorMessage" class="composer-error">{{ errorMessage }}</p>
        </form>
      </section>
    </div>

    <button v-if="previewImage" type="button" class="lightbox" @click="previewImage = ''">
      <img :src="previewImage" :alt="copy.preview" />
    </button>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { keysAPI } from '@/api'
import {
  generateImages,
  ImageWorkbenchTimeoutError,
  imageDataUrl,
  type ImageWorkbenchBackground,
  type ImageWorkbenchFormat,
  type ImageWorkbenchModel,
  type ImageWorkbenchModeration,
  type ImageWorkbenchQuality,
  type ImageWorkbenchSize,
} from '@/api/imageWorkbench'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey } from '@/types'

type TurnStatus = 'generating' | 'success' | 'error'

interface StoredWorkbenchImage {
  src: string
  revised_prompt?: string
}

interface WorkbenchTurn {
  id: string
  prompt: string
  model: ImageWorkbenchModel
  size: ImageWorkbenchSize
  quality: ImageWorkbenchQuality
  output_format: ImageWorkbenchFormat
  count: number
  status: TurnStatus
  images: StoredWorkbenchImage[]
  createdAt: string
  error?: string
}

interface WorkbenchConversation {
  id: string
  title: string
  createdAt: string
  updatedAt: string
  turns: WorkbenchTurn[]
}

const DB_NAME = 'oceanway-image-workbench'
const DB_VERSION = 1
const STORE_NAME = 'conversations'
const ACTIVE_ID_KEY = 'oceanway:image-workbench:active-id'
const MAX_HISTORY = 30

const { locale } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const modelOptions: ImageWorkbenchModel[] = ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1', 'gpt-image-1-mini']
const countOptions = [1, 2, 3, 4]
const keys = ref<ApiKey[]>([])
const keysLoading = ref(false)
const creatingKey = ref(false)
const historyLoading = ref(false)
const generating = ref(false)
const errorMessage = ref('')
const conversations = ref<WorkbenchConversation[]>([])
const activeConversationId = ref<string | null>(null)
const previewImage = ref('')
const resultsEl = ref<HTMLElement | null>(null)
let controller: AbortController | null = null

const form = reactive<{
  prompt: string
  model: ImageWorkbenchModel
  size: ImageWorkbenchSize
  quality: ImageWorkbenchQuality
  output_format: ImageWorkbenchFormat
  background: ImageWorkbenchBackground
  moderation: ImageWorkbenchModeration
  n: number
}>({
  prompt: '',
  model: 'gpt-image-2',
  size: '1024x1024',
  quality: 'medium',
  output_format: 'png',
  background: 'auto',
  moderation: 'auto',
  n: 1,
})

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const copy = computed(() => isZh.value ? zhCopy : enCopy)
const selectedKey = computed(() => keys.value.find(key => key.status === 'active') || keys.value[0] || null)
const activeConversation = computed(() => conversations.value.find(item => item.id === activeConversationId.value) || null)
const canGenerate = computed(() => Boolean(form.prompt.trim() && !generating.value))

watch(activeConversationId, (id) => {
  if (id) localStorage.setItem(ACTIVE_ID_KEY, id)
})

async function loadKeys() {
  keysLoading.value = true
  try {
    const page = await keysAPI.list(1, 20, { sort_by: 'created_at', sort_order: 'desc' })
    keys.value = page.items || []
  } catch {
    appStore.showError(copy.value.loadKeysFailed)
  } finally {
    keysLoading.value = false
  }
}

async function createDefaultKey() {
  if (creatingKey.value) return
  creatingKey.value = true
  try {
    const created = await keysAPI.create(copy.value.defaultKeyName, null, undefined, undefined, undefined, undefined, undefined, undefined, false)
    keys.value = [created, ...keys.value]
    appStore.showSuccess(copy.value.keyCreated)
  } catch {
    appStore.showError(copy.value.keyCreateFailed)
  } finally {
    creatingKey.value = false
  }
}

async function copyDefaultKey() {
  if (!selectedKey.value) return
  await copyToClipboard(selectedKey.value.key, copy.value.copied)
}

async function loadHistory() {
  historyLoading.value = true
  try {
    const loaded = await listConversations()
    const { conversations: recovered, changed } = recoverInterruptedTurns(loaded)
    conversations.value = recovered
    if (changed) {
      await persistCurrentHistory()
    }
    const activeId = localStorage.getItem(ACTIVE_ID_KEY)
    activeConversationId.value = conversations.value.some(item => item.id === activeId)
      ? activeId
      : conversations.value[0]?.id ?? null
  } catch {
    appStore.showError(copy.value.loadHistoryFailed)
  } finally {
    historyLoading.value = false
  }
}

function recoverInterruptedTurns(items: WorkbenchConversation[]) {
  let changed = false
  const interruptedMessage = copy.value.generateInterrupted
  const conversations = items.map(conversation => {
    const turns = conversation.turns.map(turn => {
      if (turn.status !== 'generating') return turn
      changed = true
      return {
        ...turn,
        status: 'error' as TurnStatus,
        error: turn.error || interruptedMessage,
      }
    })
    return turns === conversation.turns ? conversation : { ...conversation, turns }
  })

  return { conversations, changed }
}

function createDraftConversation() {
  const now = new Date().toISOString()
  const conversation: WorkbenchConversation = {
    id: createId(),
    title: copy.value.untitled,
    createdAt: now,
    updatedAt: now,
    turns: [],
  }
  conversations.value = [conversation, ...conversations.value].slice(0, MAX_HISTORY)
  activeConversationId.value = conversation.id
  void saveConversation(conversation)
}

function selectConversation(id: string) {
  activeConversationId.value = id
  void nextTick(scrollToBottom)
}

async function clearHistory() {
  conversations.value = []
  activeConversationId.value = null
  localStorage.removeItem(ACTIVE_ID_KEY)
  await clearConversations()
}

async function runGeneration() {
  if (!form.prompt.trim() || generating.value) return

  const prompt = form.prompt.trim()
  const hadDefaultKey = Boolean(selectedKey.value)
  const now = new Date().toISOString()
  let conversation = activeConversation.value
  if (!conversation) {
    conversation = {
      id: createId(),
      title: buildConversationTitle(prompt),
      createdAt: now,
      updatedAt: now,
      turns: [],
    }
    conversations.value = [conversation, ...conversations.value]
    activeConversationId.value = conversation.id
  }

  const turn: WorkbenchTurn = {
    id: createId(),
    prompt,
    model: form.model,
    size: form.size,
    quality: form.quality,
    output_format: form.output_format,
    count: normalizeCount(form.n),
    status: 'generating',
    images: [],
    createdAt: now,
  }

  conversation.title = conversation.turns.length ? conversation.title : buildConversationTitle(prompt)
  conversation.updatedAt = now
  conversation.turns = [...conversation.turns, turn]
  touchConversation(conversation)
  errorMessage.value = ''
  generating.value = true
  form.prompt = ''
  await nextTick(scrollToBottom)
  void persistCurrentHistory().catch(() => appStore.showError(copy.value.saveHistoryFailed))

  controller?.abort()
  controller = new AbortController()

  try {
    const result = await generateImages({
      model: turn.model,
      prompt: turn.prompt,
      size: turn.size,
      quality: turn.quality,
      output_format: turn.output_format,
      background: form.background,
      moderation: form.moderation,
      n: turn.count,
    }, controller.signal)

    turn.images = result.data
      .map(image => ({
        src: imageDataUrl(image, turn.output_format),
        revised_prompt: image.revised_prompt,
      }))
      .filter(image => Boolean(image.src))

    if (!turn.images.length) {
      turn.status = 'error'
      turn.error = copy.value.noImageReturned
      errorMessage.value = turn.error
    } else {
      turn.status = 'success'
    }
  } catch (error: any) {
    if (error?.name === 'AbortError') return
    const message = error instanceof ImageWorkbenchTimeoutError
      ? copy.value.generateTimeout
      : error?.message || copy.value.generateFailed
    turn.status = 'error'
    turn.error = message
    errorMessage.value = message
    appStore.showError(message)
  } finally {
    generating.value = false
    if (!hadDefaultKey) {
      void loadKeys()
    }
    conversation.updatedAt = new Date().toISOString()
    touchConversation(conversation)
    void persistCurrentHistory().catch(() => appStore.showError(copy.value.saveHistoryFailed))
    await nextTick(scrollToBottom)
  }
}

function reuseTurn(turn: WorkbenchTurn) {
  form.prompt = turn.prompt
  form.model = turn.model
  form.size = turn.size
  form.quality = turn.quality
  form.output_format = turn.output_format
  form.n = turn.count
}

function regenerateTurn(turn: WorkbenchTurn) {
  reuseTurn(turn)
  void runGeneration()
}

async function deleteTurn(conversationId: string, turnId: string) {
  const conversation = conversations.value.find(item => item.id === conversationId)
  if (!conversation) return
  conversation.turns = conversation.turns.filter(turn => turn.id !== turnId)
  conversation.updatedAt = new Date().toISOString()
  touchConversation(conversation)
  await persistCurrentHistory()
}

function touchConversation(conversation: WorkbenchConversation) {
  conversations.value = [
    conversation,
    ...conversations.value.filter(item => item.id !== conversation.id),
  ].slice(0, MAX_HISTORY)
}

async function persistCurrentHistory() {
  for (const conversation of conversations.value) {
    await saveConversation(conversation)
  }
}

function conversationHasRunningTurn(conversation: WorkbenchConversation) {
  return conversation.turns.some(turn => turn.status === 'generating')
}

function statusLabel(status: TurnStatus) {
  if (status === 'generating') return copy.value.generating
  if (status === 'error') return copy.value.failed
  return copy.value.completed
}

function normalizeCount(value: number): number {
  if (!Number.isFinite(value)) return 1
  return Math.max(1, Math.min(4, Math.floor(value)))
}

function buildConversationTitle(prompt: string): string {
  const trimmed = prompt.trim()
  if (!trimmed) return copy.value.untitled
  return trimmed.length <= 18 ? trimmed : `${trimmed.slice(0, 18)}...`
}

function formatTime(value: string): string {
  const date = new Date(value)
  if (!Number.isFinite(date.getTime())) return ''
  return new Intl.DateTimeFormat(isZh.value ? 'zh-CN' : 'en-US', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function maskKey(value: string): string {
  if (!value) return ''
  if (value.length <= 14) return value
  return `${value.slice(0, 8)}...${value.slice(-6)}`
}

function downloadName(turn: WorkbenchTurn, index: number): string {
  return `oceanway-image-${turn.id}-${index + 1}.${turn.output_format}`
}

function createId(): string {
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function scrollToBottom() {
  const el = resultsEl.value
  if (!el) return
  el.scrollTop = el.scrollHeight
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION)
    req.onupgradeneeded = () => {
      const db = req.result
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME, { keyPath: 'id' })
      }
    }
    req.onsuccess = () => resolve(req.result)
    req.onerror = () => reject(req.error)
  })
}

async function listConversations(): Promise<WorkbenchConversation[]> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly')
    const req = tx.objectStore(STORE_NAME).getAll()
    req.onsuccess = () => {
      resolve((req.result as WorkbenchConversation[]).sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)).slice(0, MAX_HISTORY))
    }
    req.onerror = () => reject(req.error)
  })
}

async function saveConversation(conversation: WorkbenchConversation): Promise<void> {
  const db = await openDB()
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite')
    tx.objectStore(STORE_NAME).put(serializeConversation(conversation))
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

function serializeConversation(conversation: WorkbenchConversation): WorkbenchConversation {
  return {
    id: conversation.id,
    title: conversation.title,
    createdAt: conversation.createdAt,
    updatedAt: conversation.updatedAt,
    turns: conversation.turns.map(turn => ({
      id: turn.id,
      prompt: turn.prompt,
      model: turn.model,
      size: turn.size,
      quality: turn.quality,
      output_format: turn.output_format,
      count: turn.count,
      status: turn.status,
      images: turn.images.map(image => ({
        src: image.src,
        revised_prompt: image.revised_prompt,
      })),
      createdAt: turn.createdAt,
      error: turn.error,
    })),
  }
}

async function clearConversations(): Promise<void> {
  const db = await openDB()
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite')
    tx.objectStore(STORE_NAME).clear()
    tx.oncomplete = () => resolve()
    tx.onerror = () => reject(tx.error)
  })
}

onMounted(() => {
  void loadKeys()
  void loadHistory()
})

onBeforeUnmount(() => {
  controller?.abort()
})

const zhCopy = {
  newChat: '新建对话',
  clearHistory: '清空历史',
  defaultKey: '默认生图 Key',
  noKey: '未创建',
  copyKey: '复制',
  createKey: '创建',
  creatingKey: '创建中...',
  loadingHistory: '正在读取会话记录',
  noHistory: '还没有图片记录，输入提示词后会在这里显示。',
  rounds: '轮',
  running: '处理中',
  heroTitle: 'Turn ideas into images',
  heroDescription: '在同一窗口里保留本地历史，用默认 Key 直接走 OceanWay 生图调度。',
  round: '第',
  reuse: '复用配置',
  delete: '删除',
  waiting: '正在等待上游图片返回...',
  failed: '失败',
  retry: '重试',
  result: '结果',
  download: '下载',
  promptPlaceholder: '输入你想要生成的画面，按 Enter 发送，Shift + Enter 换行',
  model: '模型',
  size: '尺寸',
  quality: '质量',
  count: '数量',
  format: '格式',
  auto: '自动',
  low: '低',
  medium: '中',
  high: '高',
  images: '张',
  generate: '生成图片',
  generating: '生成中',
  completed: '已完成',
  imageAlt: '生成图片 {index}',
  preview: '图片预览',
  copied: '已复制',
  untitled: '新的生图对话',
  defaultKeyName: '默认 API Key',
  keyCreated: '默认 API Key 已创建',
  keyCreateFailed: '创建默认 API Key 失败',
  loadKeysFailed: '加载 API Key 失败',
  loadHistoryFailed: '加载本地生图历史失败',
  saveHistoryFailed: '保存本地生图历史失败',
  noImageReturned: '请求成功，但没有返回图片。',
  generateFailed: '图片生成失败',
  generateTimeout: '图片生成超时。请稍后重试，或联系管理员检查工作台账号池的上游状态。',
  generateInterrupted: '上次生成已中断，请重新发起生成。',
}

const enCopy = {
  newChat: 'New chat',
  clearHistory: 'Clear history',
  defaultKey: 'Default image key',
  noKey: 'Not created',
  copyKey: 'Copy',
  createKey: 'Create',
  creatingKey: 'Creating...',
  loadingHistory: 'Loading conversations',
  noHistory: 'No image history yet. Prompts will appear here.',
  rounds: 'turns',
  running: 'Running',
  heroTitle: 'Turn ideas into images',
  heroDescription: 'Keep local image history in one window and route generation through OceanWay with your default key.',
  round: 'Turn',
  reuse: 'Reuse config',
  delete: 'Delete',
  waiting: 'Waiting for the upstream image response...',
  failed: 'Failed',
  retry: 'Retry',
  result: 'Result',
  download: 'Download',
  promptPlaceholder: 'Describe the image. Enter sends, Shift + Enter adds a line.',
  model: 'Model',
  size: 'Size',
  quality: 'Quality',
  count: 'Count',
  format: 'Format',
  auto: 'Auto',
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  images: 'image(s)',
  generate: 'Generate image',
  generating: 'Generating',
  completed: 'Completed',
  imageAlt: 'Generated image {index}',
  preview: 'Image preview',
  copied: 'Copied',
  untitled: 'New image chat',
  defaultKeyName: 'Default API Key',
  keyCreated: 'Default API key created',
  keyCreateFailed: 'Failed to create default API key',
  loadKeysFailed: 'Failed to load API keys',
  loadHistoryFailed: 'Failed to load local image history',
  saveHistoryFailed: 'Failed to save local image history',
  noImageReturned: 'The request succeeded, but no image was returned.',
  generateFailed: 'Image generation failed',
  generateTimeout: 'Image generation timed out. Please try again later or ask an admin to check the workbench account pool.',
  generateInterrupted: 'The previous generation was interrupted. Please start it again.',
}
</script>

<style scoped>
.image-shell {
  display: grid;
  grid-template-columns: minmax(14rem, 18rem) minmax(0, 1fr);
  height: calc(100vh - 6.5rem);
  min-height: 42rem;
  overflow: hidden;
  border: 1px solid #e7e5e4;
  border-radius: 0.5rem;
  background: #fafaf9;
}

.dark .image-shell {
  border-color: #292524;
  background: #0c0a09;
}

.image-sidebar {
  display: flex;
  min-height: 0;
  flex-direction: column;
  gap: 0.75rem;
  border-right: 1px solid #e7e5e4;
  background: #f5f5f4;
  padding: 0.75rem;
}

.dark .image-sidebar {
  border-right-color: #292524;
  background: #1c1917;
}

.sidebar-actions {
  display: flex;
  gap: 0.5rem;
}

.new-chat-button,
.icon-button,
.send-button,
.key-strip button,
.turn-actions button,
.error-result button {
  border: 0;
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease, opacity 0.15s ease;
}

.new-chat-button {
  display: inline-flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 0.5rem;
  background: #1c1917;
  color: #ffffff;
  font-size: 0.9rem;
  font-weight: 800;
  min-height: 2.5rem;
}

.new-chat-button:hover {
  background: #44403c;
}

.icon-button {
  width: 2.5rem;
  border: 1px solid #e7e5e4;
  border-radius: 0.5rem;
  background: #ffffff;
  color: #78716c;
  font-size: 1.35rem;
}

.icon-button:disabled,
.send-button:disabled,
.key-strip button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.key-strip {
  display: grid;
  gap: 0.35rem;
  border: 1px solid #e7e5e4;
  border-radius: 0.5rem;
  background: #ffffffcc;
  padding: 0.7rem;
}

.dark .key-strip,
.dark .icon-button {
  border-color: #44403c;
  background: #0c0a09;
}

.key-strip span,
.history-meta {
  color: #78716c;
  font-size: 0.76rem;
}

.key-strip strong {
  color: #1c1917;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.84rem;
}

.dark .key-strip strong {
  color: #f5f5f4;
}

.key-strip button {
  justify-self: start;
  border-radius: 999px;
  background: #e7e5e4;
  color: #292524;
  font-size: 0.76rem;
  font-weight: 800;
  padding: 0.25rem 0.65rem;
}

.history-list {
  min-height: 0;
  flex: 1;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.history-empty {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #78716c;
  font-size: 0.86rem;
  line-height: 1.55;
  padding: 0.75rem 0.25rem;
}

.history-item {
  position: relative;
  display: grid;
  width: 100%;
  gap: 0.2rem;
  border: 0;
  border-left: 2px solid transparent;
  background: transparent;
  color: #57534e;
  cursor: pointer;
  margin-bottom: 0.25rem;
  padding: 0.65rem 0.65rem 0.65rem 0.75rem;
  text-align: left;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.history-item:hover,
.history-item.active {
  border-left-color: #1c1917;
  background: rgba(28, 25, 23, 0.055);
  color: #1c1917;
}

.dark .history-item {
  color: #d6d3d1;
}

.dark .history-item:hover,
.dark .history-item.active {
  border-left-color: #f5f5f4;
  background: rgba(245, 245, 244, 0.08);
  color: #ffffff;
}

.history-title {
  overflow: hidden;
  font-size: 0.9rem;
  font-weight: 850;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-running {
  justify-self: start;
  border-radius: 999px;
  background: #dbeafe;
  color: #2563eb;
  font-size: 0.68rem;
  font-weight: 850;
  padding: 0.15rem 0.5rem;
}

.image-main {
  display: grid;
  grid-template-rows: minmax(0, 1fr) auto;
  min-width: 0;
  min-height: 0;
  background: #fffaf5;
}

.dark .image-main {
  background: #0c0a09;
}

.result-stream {
  min-height: 0;
  overflow-y: auto;
  padding: 1.5rem 1rem 1rem;
}

.welcome-state {
  display: grid;
  min-height: 100%;
  place-content: center;
  text-align: center;
}

.welcome-state h1 {
  color: #1c1917;
  font-family: "Palatino Linotype", "Book Antiqua", "Times New Roman", serif;
  font-size: clamp(2.1rem, 6vw, 4.3rem);
  font-weight: 700;
  line-height: 1.04;
}

.welcome-state p {
  color: #78716c;
  font-family: "Palatino Linotype", "Book Antiqua", "Times New Roman", serif;
  font-size: 1rem;
  font-style: italic;
  margin: 0.85rem auto 0;
  max-width: 36rem;
}

.dark .welcome-state h1 {
  color: #fafaf9;
}

.turn-stack {
  display: flex;
  flex-direction: column;
  gap: 1.65rem;
  margin: 0 auto;
  max-width: 62rem;
}

.turn {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.prompt-row {
  display: flex;
  justify-content: flex-end;
}

.prompt-bubble {
  max-width: min(84%, 46rem);
  color: #1c1917;
  padding: 0.25rem 0.15rem;
  text-align: right;
}

.turn-meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.4rem;
  color: #a8a29e;
  font-size: 0.72rem;
  margin-bottom: 0.35rem;
}

.prompt-bubble p {
  font-size: 0.96rem;
  line-height: 1.75;
  white-space: pre-wrap;
}

.turn-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.35rem;
  margin-top: 0.45rem;
}

.turn-actions button {
  border-radius: 999px;
  background: #f5f5f4;
  color: #78716c;
  font-size: 0.72rem;
  font-weight: 800;
  padding: 0.25rem 0.6rem;
}

.turn-actions button:hover {
  background: #e7e5e4;
  color: #1c1917;
}

.result-row {
  display: flex;
  justify-content: flex-start;
}

.result-bubble {
  width: 100%;
  max-width: 58rem;
}

.loading-result,
.error-result {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  border-radius: 999px;
  background: #ffffff;
  color: #78716c;
  font-size: 0.9rem;
  padding: 0.65rem 0.9rem;
}

.error-result {
  align-items: flex-start;
  border-radius: 0.5rem;
  border: 1px solid #fecdd3;
  background: #fff1f2;
  color: #9f1239;
  flex-direction: column;
}

.error-result button {
  border-radius: 999px;
  background: #ffffff;
  color: #9f1239;
  font-weight: 850;
  padding: 0.35rem 0.75rem;
}

.generated-grid {
  column-count: 2;
  column-gap: 1rem;
}

.generated-card {
  break-inside: avoid;
  margin: 0 0 1rem;
}

.image-preview {
  display: block;
  width: 100%;
  overflow: hidden;
  border: 0;
  border-radius: 0.75rem;
  background: #e7e5e4;
  cursor: zoom-in;
  padding: 0;
}

.image-preview img {
  display: block;
  width: 100%;
  object-fit: contain;
}

.generated-card figcaption {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  color: #78716c;
  font-size: 0.78rem;
  padding: 0.45rem 0.2rem;
}

.generated-card a {
  color: #1c1917;
  font-weight: 850;
}

.revised-prompt {
  color: #78716c;
  font-size: 0.76rem;
  line-height: 1.55;
  padding: 0 0.2rem 0.5rem;
}

.composer-wrap {
  padding: 0.75rem 1rem 1rem;
}

.composer-card {
  overflow: hidden;
  border: 1px solid #e7e5e4;
  border-radius: 1.5rem;
  background: #ffffff;
  box-shadow: 0 16px 48px rgba(28, 25, 23, 0.08);
  margin: 0 auto;
  max-width: 62rem;
}

.dark .composer-card,
.dark .loading-result {
  border-color: #292524;
  background: #1c1917;
}

.composer-card textarea {
  display: block;
  width: 100%;
  min-height: 5.5rem;
  resize: none;
  border: 0;
  background: transparent;
  color: #1c1917;
  font-size: 0.96rem;
  line-height: 1.6;
  outline: none;
  padding: 1rem 1rem 0.65rem;
}

.dark .composer-card textarea {
  color: #fafaf9;
}

.composer-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0 0.75rem 0.75rem;
}

.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 0.45rem;
  min-width: 0;
}

.pill-group select {
  height: 2.15rem;
  max-width: 10.5rem;
  border: 1px solid #e7e5e4;
  border-radius: 999px;
  background: #ffffff;
  color: #57534e;
  font-size: 0.78rem;
  font-weight: 750;
  outline: none;
  padding: 0 0.7rem;
}

.dark .pill-group select {
  border-color: #44403c;
  background: #0c0a09;
  color: #e7e5e4;
}

.send-button {
  display: inline-grid;
  flex: 0 0 auto;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border-radius: 999px;
  background: #1c1917;
  color: #ffffff;
  font-size: 1.2rem;
  font-weight: 900;
}

.composer-error {
  color: #dc2626;
  font-size: 0.85rem;
  font-weight: 800;
  margin: 0.6rem auto 0;
  max-width: 62rem;
}

.lightbox {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: grid;
  place-items: center;
  border: 0;
  background: rgba(12, 10, 9, 0.82);
  cursor: zoom-out;
  padding: 2rem;
}

.lightbox img {
  max-width: min(100%, 86rem);
  max-height: 100%;
  border-radius: 0.5rem;
  object-fit: contain;
}

@media (max-width: 1023px) {
  .image-shell {
    grid-template-columns: 1fr;
    height: auto;
    min-height: 0;
  }

  .image-sidebar {
    max-height: 16rem;
    border-right: 0;
    border-bottom: 1px solid #e7e5e4;
  }

  .image-main {
    min-height: 38rem;
  }
}

@media (max-width: 767px) {
  .result-stream {
    padding: 1rem 0.75rem;
  }

  .prompt-bubble {
    max-width: 94%;
  }

  .generated-grid {
    column-count: 1;
  }

  .composer-wrap {
    padding: 0.75rem;
  }
}
</style>
