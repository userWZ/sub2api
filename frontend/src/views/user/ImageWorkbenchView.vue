<template>
  <AppLayout>
    <div class="image-workbench">
      <section class="workbench-head">
        <div>
          <p class="eyebrow">{{ copy.eyebrow }}</p>
          <h1>{{ copy.title }}</h1>
          <p>{{ copy.description }}</p>
        </div>
        <div class="key-box">
          <span>{{ copy.defaultKey }}</span>
          <strong>{{ selectedKey ? maskKey(selectedKey.key) : copy.noKey }}</strong>
          <button v-if="selectedKey" class="btn btn-secondary btn-sm" type="button" @click="copyDefaultKey">
            {{ copy.copyKey }}
          </button>
          <button v-else class="btn btn-primary btn-sm" type="button" :disabled="keysLoading || creatingKey" @click="createDefaultKey">
            {{ creatingKey ? copy.creatingKey : copy.createKey }}
          </button>
        </div>
      </section>

      <div class="workbench-grid">
        <section class="control-panel">
          <label class="field">
            <span>{{ copy.prompt }}</span>
            <textarea v-model="form.prompt" :placeholder="copy.promptPlaceholder" rows="8" />
          </label>

          <div class="field-row">
            <label class="field">
              <span>{{ copy.model }}</span>
              <select v-model="form.model">
                <option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option>
              </select>
            </label>
            <label class="field">
              <span>{{ copy.count }}</span>
              <input v-model.number="form.n" type="number" min="1" max="4" step="1" />
            </label>
          </div>

          <div class="field-row">
            <label class="field">
              <span>{{ copy.size }}</span>
              <select v-model="form.size">
                <option value="1024x1024">1024 x 1024</option>
                <option value="1024x1536">1024 x 1536</option>
                <option value="1536x1024">1536 x 1024</option>
                <option value="auto">{{ copy.auto }}</option>
              </select>
            </label>
            <label class="field">
              <span>{{ copy.quality }}</span>
              <select v-model="form.quality">
                <option value="auto">{{ copy.auto }}</option>
                <option value="low">{{ copy.low }}</option>
                <option value="medium">{{ copy.medium }}</option>
                <option value="high">{{ copy.high }}</option>
              </select>
            </label>
          </div>

          <div class="field-row">
            <label class="field">
              <span>{{ copy.format }}</span>
              <select v-model="form.output_format">
                <option value="png">PNG</option>
                <option value="jpeg">JPEG</option>
                <option value="webp">WebP</option>
              </select>
            </label>
            <label class="field">
              <span>{{ copy.background }}</span>
              <select v-model="form.background">
                <option value="auto">{{ copy.auto }}</option>
                <option value="transparent">{{ copy.transparent }}</option>
                <option value="opaque">{{ copy.opaque }}</option>
              </select>
            </label>
          </div>

          <label class="field">
            <span>{{ copy.moderation }}</span>
            <select v-model="form.moderation">
              <option value="auto">{{ copy.auto }}</option>
              <option value="low">{{ copy.low }}</option>
            </select>
          </label>

          <div class="action-row">
            <button class="btn btn-primary" type="button" :disabled="!canGenerate" @click="runGeneration">
              {{ generating ? copy.generating : copy.generate }}
            </button>
            <button class="btn btn-secondary" type="button" :disabled="generating" @click="resetForm">
              {{ copy.reset }}
            </button>
          </div>

          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>
          <p class="hint">{{ copy.hint }}</p>
        </section>

        <section class="result-panel">
          <div class="result-head">
            <div>
              <p class="eyebrow">{{ copy.results }}</p>
              <h2>{{ resultTitle }}</h2>
            </div>
            <button class="btn btn-secondary btn-sm" type="button" :disabled="!images.length" @click="clearResults">
              {{ copy.clear }}
            </button>
          </div>

          <div v-if="generating" class="empty-state">
            <LoadingSpinner />
            <span>{{ copy.waiting }}</span>
          </div>
          <div v-else-if="!images.length" class="empty-state">
            <span>{{ copy.empty }}</span>
          </div>
          <div v-else class="image-grid">
            <article v-for="(image, index) in images" :key="`${image.src}-${index}`" class="image-card">
              <img :src="image.src" :alt="copy.imageAlt.replace('{index}', String(index + 1))" />
              <div class="image-actions">
                <a class="btn btn-secondary btn-sm" :href="image.src" :download="downloadName(index)">
                  {{ copy.download }}
                </a>
              </div>
              <p v-if="image.revised_prompt" class="revised-prompt">{{ image.revised_prompt }}</p>
            </article>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { keysAPI } from '@/api'
import { generateImages, imageDataUrl, type ImageWorkbenchModel, type ImageWorkbenchSize, type ImageWorkbenchQuality, type ImageWorkbenchFormat, type ImageWorkbenchBackground, type ImageWorkbenchModeration } from '@/api/imageWorkbench'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey } from '@/types'

const { locale } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const modelOptions: ImageWorkbenchModel[] = ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1', 'gpt-image-1-mini']
const keys = ref<ApiKey[]>([])
const keysLoading = ref(false)
const creatingKey = ref(false)
const generating = ref(false)
const errorMessage = ref('')
const images = ref<Array<{ src: string; revised_prompt?: string }>>([])
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
const canGenerate = computed(() => Boolean(selectedKey.value && form.prompt.trim() && !generating.value))
const resultTitle = computed(() => images.value.length ? copy.value.resultCount.replace('{count}', String(images.value.length)) : copy.value.resultPending)

async function loadKeys() {
  keysLoading.value = true
  try {
    const page = await keysAPI.list(1, 20, { sort_by: 'created_at', sort_order: 'desc' })
    keys.value = page.items || []
  } catch (error) {
    appStore.showError(copy.value.loadKeysFailed)
  } finally {
    keysLoading.value = false
  }
}

async function createDefaultKey() {
  if (creatingKey.value) return
  creatingKey.value = true
  try {
    const created = await keysAPI.create(copy.value.defaultKeyName, null, undefined, undefined, undefined, undefined, undefined, undefined, true)
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

async function runGeneration() {
  if (!selectedKey.value || generating.value) return
  errorMessage.value = ''
  generating.value = true
  controller?.abort()
  controller = new AbortController()

  try {
    const result = await generateImages({
      apiKey: selectedKey.value.key,
      model: form.model,
      prompt: form.prompt.trim(),
      size: form.size,
      quality: form.quality,
      output_format: form.output_format,
      background: form.background,
      moderation: form.moderation,
      n: normalizeCount(form.n),
    }, controller.signal)
    images.value = result.data
      .map(image => ({
        src: imageDataUrl(image, form.output_format),
        revised_prompt: image.revised_prompt,
      }))
      .filter(image => Boolean(image.src))
    if (!images.value.length) {
      errorMessage.value = copy.value.noImageReturned
    }
  } catch (error: any) {
    if (error?.name === 'AbortError') return
    errorMessage.value = error?.message || copy.value.generateFailed
    appStore.showError(errorMessage.value)
  } finally {
    generating.value = false
  }
}

function resetForm() {
  form.prompt = ''
  form.model = 'gpt-image-2'
  form.size = '1024x1024'
  form.quality = 'medium'
  form.output_format = 'png'
  form.background = 'auto'
  form.moderation = 'auto'
  form.n = 1
  errorMessage.value = ''
}

function clearResults() {
  images.value = []
}

function normalizeCount(value: number): number {
  if (!Number.isFinite(value)) return 1
  return Math.max(1, Math.min(4, Math.floor(value)))
}

function maskKey(value: string): string {
  if (!value) return ''
  if (value.length <= 14) return value
  return `${value.slice(0, 8)}...${value.slice(-6)}`
}

function downloadName(index: number): string {
  return `oceanway-image-${Date.now()}-${index + 1}.${form.output_format}`
}

onMounted(loadKeys)

onBeforeUnmount(() => {
  controller?.abort()
})

const zhCopy = {
  eyebrow: 'Image Workbench',
  title: '生图工作台',
  description: '使用当前账户默认 Key 直接调用图片生成接口，走平台现有分组权限、账号调度和额度计费。',
  defaultKey: '默认生图 Key',
  noKey: '未创建',
  copyKey: '复制 Key',
  createKey: '创建默认 Key',
  creatingKey: '创建中...',
  prompt: '提示词',
  promptPlaceholder: '描述你想生成的画面、风格、主体、构图、背景和细节。',
  model: '模型',
  count: '数量',
  size: '尺寸',
  quality: '质量',
  format: '格式',
  background: '背景',
  moderation: '审核强度',
  auto: '自动',
  low: '低',
  medium: '中',
  high: '高',
  transparent: '透明',
  opaque: '不透明',
  generate: '生成图片',
  generating: '生成中...',
  reset: '重置',
  hint: '建议优先使用 1 张图测试提示词，再提高质量或增加数量。',
  results: 'Results',
  resultPending: '等待生成',
  resultCount: '已生成 {count} 张图片',
  clear: '清空',
  waiting: '正在等待上游图片返回...',
  empty: '生成结果会显示在这里。',
  imageAlt: '生成图片 {index}',
  download: '下载',
  copied: '已复制',
  defaultKeyName: '默认 API Key',
  keyCreated: '默认 API Key 已创建',
  keyCreateFailed: '创建默认 API Key 失败',
  loadKeysFailed: '加载 API Key 失败',
  noImageReturned: '请求成功，但没有返回图片。',
  generateFailed: '图片生成失败',
}

const enCopy = {
  eyebrow: 'Image Workbench',
  title: 'Image Workbench',
  description: 'Generate images with your default key while keeping the existing group permissions, scheduler, and billing flow.',
  defaultKey: 'Default Image Key',
  noKey: 'Not created',
  copyKey: 'Copy Key',
  createKey: 'Create Default Key',
  creatingKey: 'Creating...',
  prompt: 'Prompt',
  promptPlaceholder: 'Describe the image, style, subject, composition, background, and details.',
  model: 'Model',
  count: 'Count',
  size: 'Size',
  quality: 'Quality',
  format: 'Format',
  background: 'Background',
  moderation: 'Moderation',
  auto: 'Auto',
  low: 'Low',
  medium: 'Medium',
  high: 'High',
  transparent: 'Transparent',
  opaque: 'Opaque',
  generate: 'Generate',
  generating: 'Generating...',
  reset: 'Reset',
  hint: 'Start with one image to tune the prompt, then increase quality or count.',
  results: 'Results',
  resultPending: 'Waiting',
  resultCount: '{count} image(s) generated',
  clear: 'Clear',
  waiting: 'Waiting for the upstream image response...',
  empty: 'Generated images will appear here.',
  imageAlt: 'Generated image {index}',
  download: 'Download',
  copied: 'Copied',
  defaultKeyName: 'Default API Key',
  keyCreated: 'Default API key created',
  keyCreateFailed: 'Failed to create default API key',
  loadKeysFailed: 'Failed to load API keys',
  noImageReturned: 'The request succeeded, but no image was returned.',
  generateFailed: 'Image generation failed',
}
</script>

<style scoped>
.image-workbench {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.workbench-head,
.control-panel,
.result-panel {
  border: 1px solid rgba(203, 213, 225, 0.85);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 16px 36px rgba(15, 23, 42, 0.06);
}

.dark .workbench-head,
.dark .control-panel,
.dark .result-panel {
  border-color: rgba(51, 65, 85, 0.9);
  background: rgba(15, 23, 42, 0.96);
}

.workbench-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem;
}

.workbench-head h1,
.result-head h2 {
  color: #0f172a;
  font-weight: 900;
}

.dark .workbench-head h1,
.dark .result-head h2 {
  color: #f8fafc;
}

.workbench-head h1 {
  font-size: clamp(1.8rem, 3vw, 2.5rem);
  line-height: 1.05;
}

.workbench-head p:not(.eyebrow) {
  margin-top: 0.5rem;
  max-width: 48rem;
  color: #64748b;
}

.dark .workbench-head p:not(.eyebrow) {
  color: #94a3b8;
}

.eyebrow {
  color: #0074d9;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0;
  text-transform: uppercase;
}

.key-box {
  display: grid;
  min-width: min(100%, 17rem);
  gap: 0.45rem;
  border-radius: 0.5rem;
  background: #f8fafc;
  padding: 0.85rem;
}

.dark .key-box {
  background: #111827;
}

.key-box span {
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 800;
}

.key-box strong {
  color: #0f172a;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.9rem;
}

.dark .key-box strong {
  color: #e2e8f0;
}

.workbench-grid {
  display: grid;
  gap: 1rem;
}

@media (min-width: 1024px) {
  .workbench-grid {
    grid-template-columns: minmax(20rem, 0.72fr) minmax(0, 1.28fr);
  }
}

.control-panel,
.result-panel {
  padding: 1rem;
}

.field,
.field-row {
  display: grid;
  gap: 0.45rem;
}

.field-row {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 0.85rem;
}

.field + .field,
.field-row + .field,
.field + .action-row {
  margin-top: 0.85rem;
}

.field span {
  color: #334155;
  font-size: 0.78rem;
  font-weight: 850;
}

.dark .field span {
  color: #cbd5e1;
}

.field textarea,
.field select,
.field input {
  width: 100%;
  border: 1px solid #cbd5e1;
  border-radius: 0.45rem;
  background: #ffffff;
  color: #0f172a;
  font-size: 0.92rem;
  outline: none;
  padding: 0.65rem 0.75rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.field textarea {
  min-height: 12rem;
  resize: vertical;
}

.dark .field textarea,
.dark .field select,
.dark .field input {
  border-color: #334155;
  background: #0f172a;
  color: #f8fafc;
}

.field textarea:focus,
.field select:focus,
.field input:focus {
  border-color: #0ea5e9;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.16);
}

.action-row,
.result-head,
.image-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.action-row {
  margin-top: 1rem;
  justify-content: flex-start;
}

.hint,
.error-message {
  margin-top: 0.85rem;
  font-size: 0.85rem;
}

.hint {
  color: #64748b;
}

.error-message {
  color: #dc2626;
  font-weight: 800;
}

.dark .hint {
  color: #94a3b8;
}

.empty-state {
  display: grid;
  min-height: 28rem;
  place-items: center;
  align-content: center;
  gap: 0.75rem;
  color: #64748b;
  text-align: center;
}

.dark .empty-state {
  color: #94a3b8;
}

.image-grid {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
}

@media (min-width: 768px) {
  .image-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.image-card {
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  background: #f8fafc;
}

.dark .image-card {
  border-color: #334155;
  background: #111827;
}

.image-card img {
  display: block;
  width: 100%;
  aspect-ratio: 1 / 1;
  background: #e2e8f0;
  object-fit: contain;
}

.image-actions {
  justify-content: flex-end;
  padding: 0.7rem;
}

.revised-prompt {
  border-top: 1px solid #e2e8f0;
  color: #475569;
  font-size: 0.78rem;
  line-height: 1.45;
  padding: 0.7rem;
}

.dark .revised-prompt {
  border-top-color: #334155;
  color: #cbd5e1;
}

@media (max-width: 767px) {
  .workbench-head,
  .field-row {
    grid-template-columns: 1fr;
  }

  .workbench-head {
    display: grid;
  }

  .action-row {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
