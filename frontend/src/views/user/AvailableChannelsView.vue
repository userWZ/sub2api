<template>
  <AppLayout>
    <section class="mb-6 rounded-xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800">
      <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <p class="text-xs font-bold uppercase tracking-wide text-primary-600 dark:text-primary-400">
            {{ t('availableChannels.serviceOverviewEyebrow') }}
          </p>
          <h1 class="mt-1 text-2xl font-black text-gray-900 dark:text-white">
            {{ t('availableChannels.serviceOverviewTitle') }}
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-500 dark:text-gray-400">
            {{ t('availableChannels.serviceOverviewDescription') }}
          </p>
        </div>
        <RouterLink to="/monitor" class="btn btn-secondary">
          {{ t('availableChannels.viewServiceStatus') }}
        </RouterLink>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-5">
        <div
          v-for="service in serviceCards"
          :key="service.key"
          class="rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900/40"
        >
          <div class="flex items-center justify-between gap-3">
            <span class="font-semibold text-gray-900 dark:text-white">{{ service.label }}</span>
            <span
              :class="[
                'rounded-full px-2 py-0.5 text-xs font-bold',
                service.available
                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                  : 'bg-gray-200 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
              ]"
            >
              {{ service.available ? t('availableChannels.serviceAvailable') : t('availableChannels.serviceUnavailable') }}
            </span>
          </div>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
            {{ service.detail }}
          </p>
        </div>
      </div>
    </section>

    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-80">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('availableChannels.searchPlaceholder')"
                class="input pl-10"
              />
            </div>
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadChannels"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh', 'Refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <AvailableChannelsTable
          :columns="columnLabels"
          :rows="filteredChannels"
          :loading="loading"
          :user-group-rates="userGroupRates"
          pricing-key-prefix="availableChannels.pricing"
          :no-pricing-label="t('availableChannels.noPricing')"
          :no-models-label="t('availableChannels.noModels')"
          :empty-label="t('availableChannels.empty')"
        />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AvailableChannelsTable from '@/components/channels/AvailableChannelsTable.vue'
import userChannelsAPI, { type UserAvailableChannel } from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const channels = ref<UserAvailableChannel[]>([])
const userGroupRates = ref<Record<number, number>>({})
const loading = ref(false)
const searchQuery = ref('')

const columnLabels = computed(() => ({
  name: t('availableChannels.columns.name'),
  description: t('availableChannels.columns.description'),
  platform: t('availableChannels.columns.platform'),
  groups: t('availableChannels.columns.groups'),
  supportedModels: t('availableChannels.columns.supportedModels'),
}))

const serviceCards = computed(() => {
  const platforms = new Set<string>()
  const models = new Set<string>()
  let imageModels = 0
  channels.value.forEach(channel => {
    channel.platforms.forEach(section => {
      platforms.add(section.platform.toLowerCase())
      section.supported_models.forEach(model => {
        const name = model.name.toLowerCase()
        models.add(name)
        if (
          name.includes('image') ||
          name.includes('dall') ||
          model.pricing?.image_output_price != null
        ) {
          imageModels += 1
        }
      })
    })
  })

  const hasText = models.size > 0
  const hasCode = [...models].some(name => name.includes('claude') || name.includes('gpt') || name.includes('gemini') || name.includes('code'))
  const hasOpenAI = platforms.has('openai')
  const hasClaude = platforms.has('anthropic') || platforms.has('claude')
  return [
    {
      key: 'text',
      label: t('availableChannels.services.text'),
      available: hasText,
      detail: hasText
        ? t('availableChannels.serviceModelCount', { count: models.size })
        : t('availableChannels.serviceNoModel'),
    },
    {
      key: 'code',
      label: t('availableChannels.services.code'),
      available: hasCode,
      detail: hasCode ? t('availableChannels.serviceIncluded') : t('availableChannels.serviceNoModel'),
    },
    {
      key: 'image',
      label: t('availableChannels.services.image'),
      available: imageModels > 0,
      detail: imageModels > 0
        ? t('availableChannels.serviceImageCount', { count: imageModels })
        : t('availableChannels.serviceNoModel'),
    },
    {
      key: 'openai',
      label: t('availableChannels.services.openai'),
      available: hasOpenAI,
      detail: hasOpenAI ? t('availableChannels.serviceIncluded') : t('availableChannels.serviceNoModel'),
    },
    {
      key: 'claude',
      label: t('availableChannels.services.claude'),
      available: hasClaude,
      detail: hasClaude ? t('availableChannels.serviceIncluded') : t('availableChannels.serviceNoModel'),
    },
  ]
})

/**
 * 搜索过滤：
 * - 命中渠道名/描述 → 整个渠道（所有 platforms）都保留
 * - 否则按 platform/group/model 维度在 sections 里过滤，保留有匹配的 section
 * - 所有 sections 都不匹配时，渠道本身被过滤掉
 */
const filteredChannels = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return channels.value
  return channels.value
    .map((ch) => {
      const nameHit = ch.name.toLowerCase().includes(q)
      const descHit = (ch.description || '').toLowerCase().includes(q)
      if (nameHit || descHit) return ch
      const matchingSections = ch.platforms.filter(
        (p) =>
          p.platform.toLowerCase().includes(q) ||
          p.groups.some((g) => g.name.toLowerCase().includes(q)) ||
          p.supported_models.some((m) => m.name.toLowerCase().includes(q)),
      )
      if (matchingSections.length === 0) return null
      return { ...ch, platforms: matchingSections }
    })
    .filter((ch): ch is UserAvailableChannel => ch !== null)
})

async function loadChannels() {
  loading.value = true
  try {
    // 渠道列表和用户专属倍率并发拉取。专属倍率失败不阻塞渠道展示——
    // 失败时只是无法渲染专属倍率角标，降级为仅显示默认倍率。
    const [list, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch((err: unknown) => {
        console.error('Failed to load user group rates:', err)
        return {} as Record<number, number>
      }),
    ])
    channels.value = list
    userGroupRates.value = rates
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    loading.value = false
  }
}

onMounted(loadChannels)
</script>
