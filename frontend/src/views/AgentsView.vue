<template>
  <div class="min-h-screen bg-slate-50 text-slate-950">
    <header class="border-b border-slate-200 bg-white/85 backdrop-blur">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-5 py-4">
        <router-link :to="homePath" class="flex items-center gap-3 font-semibold">
          <img :src="siteLogo || '/logo.png'" alt="" class="h-9 w-9 rounded-lg object-contain" />
          <span>{{ siteName }} Agents</span>
        </router-link>
        <div class="flex items-center gap-4 text-sm font-medium text-slate-600">
          <LocaleSwitcher />
          <router-link to="/docs" class="hover:text-slate-950">Docs</router-link>
          <router-link :to="homePath" class="hover:text-slate-950">Home</router-link>
          <router-link to="/login" class="rounded-full bg-slate-950 px-4 py-2 text-white hover:bg-slate-800">
            Sign in
          </router-link>
        </div>
      </nav>
    </header>

    <main class="mx-auto max-w-6xl px-5 py-10">
      <section class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_20rem] lg:items-end">
        <div>
          <p class="text-sm font-bold uppercase tracking-[0.16em] text-sky-700">Agents Hub</p>
          <h1 class="mt-4 max-w-4xl text-4xl font-black tracking-normal sm:text-6xl">
            Task-ready prompts for professional work.
          </h1>
          <p class="mt-5 max-w-3xl text-lg leading-8 text-slate-600">
            Start with curated agent templates for research, writing, legal review, operations, and software tasks.
          </p>
        </div>
        <div class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <p class="text-xs font-bold uppercase tracking-[0.14em] text-slate-400">Available templates</p>
          <p class="mt-2 text-4xl font-black">{{ agents.length }}</p>
          <p class="mt-2 text-sm leading-6 text-slate-600">Copy a template into Codex, Claude Code, or another agent-capable client.</p>
        </div>
      </section>

      <section class="mt-8 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <input
            v-model="query"
            type="search"
            placeholder="Search agents"
            class="w-full rounded-md border border-slate-200 px-3 py-2 text-sm outline-none focus:border-sky-400 md:max-w-sm"
          />
          <div class="flex flex-wrap gap-2">
            <button
              v-for="category in categories"
              :key="category"
              type="button"
              class="rounded-full border px-3 py-1.5 text-xs font-bold transition"
              :class="activeCategory === category ? 'border-slate-950 bg-slate-950 text-white' : 'border-slate-200 text-slate-600 hover:border-slate-400'"
              @click="activeCategory = category"
            >
              {{ category }}
            </button>
          </div>
        </div>
      </section>

      <section class="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <article v-for="agent in filteredAgents" :key="agent.name" class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-lg font-black">{{ agent.name }}</p>
              <p class="mt-1 text-xs font-bold uppercase tracking-[0.12em] text-sky-700">{{ agent.category }}</p>
            </div>
            <span class="text-2xl">{{ agent.mark }}</span>
          </div>
          <p class="mt-4 text-sm leading-6 text-slate-600">{{ agent.description }}</p>
          <pre class="mt-4 max-h-36 overflow-auto rounded-md bg-slate-950 p-3 text-xs leading-5 text-slate-100">{{ agent.prompt }}</pre>
          <button type="button" class="mt-4 rounded-full bg-sky-700 px-4 py-2 text-sm font-bold text-white hover:bg-sky-800" @click="copyAgent(agent)">
            {{ copiedAgent === agent.name ? 'Copied' : 'Copy template' }}
          </button>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { resolveHomePathForHost } from '@/utils/homeDomain'

interface AgentTemplate {
  name: string
  category: string
  mark: string
  description: string
  prompt: string
}

const appStore = useAppStore()
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'OceanWay AI')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const homePath = computed(() => resolveHomePathForHost(appStore.cachedPublicSettings ?? window.__APP_CONFIG__))
const query = ref('')
const activeCategory = ref('All')
const copiedAgent = ref('')

const agents: AgentTemplate[] = [
  {
    name: 'Research Synthesizer',
    category: 'Research',
    mark: 'R',
    description: 'Turn papers, notes, and source material into a grounded brief with assumptions and open questions.',
    prompt: 'You are a research synthesizer. Identify the user question, summarize only supported facts, separate inference from evidence, and end with open questions.',
  },
  {
    name: 'Patent Drafting Partner',
    category: 'Legal',
    mark: 'P',
    description: 'Structure technical invention material into claims, embodiments, and support checks.',
    prompt: 'You are a patent drafting partner. Extract technical features, map them to claim elements, flag missing support, and draft in precise legal-technical language.',
  },
  {
    name: 'Clinical Writing Assistant',
    category: 'Writing',
    mark: 'C',
    description: 'Prepare patient-facing or professional medical writing with careful uncertainty language.',
    prompt: 'You are a clinical writing assistant. Use plain language, avoid diagnosis beyond supplied facts, and clearly mark when clinician review is required.',
  },
  {
    name: 'Code Review Lead',
    category: 'Engineering',
    mark: 'E',
    description: 'Review code changes for defects, regressions, missing tests, and maintainability risks.',
    prompt: 'You are a code review lead. Prioritize bugs and behavioral regressions first. Cite file and line references. Keep summaries secondary.',
  },
  {
    name: 'Operations Runbook Writer',
    category: 'Operations',
    mark: 'O',
    description: 'Convert deployment context into a practical runbook with rollback and verification steps.',
    prompt: 'You are an operations runbook writer. Produce prerequisites, deploy steps, verification, rollback, monitoring, and owner handoff.',
  },
  {
    name: 'Contract Risk Reader',
    category: 'Legal',
    mark: 'L',
    description: 'Review contract text for commercial risk, ambiguous duties, renewal traps, and missing protections.',
    prompt: 'You are a contract risk reader. Extract obligations, deadlines, payment terms, termination rules, liability exposure, and negotiation points.',
  },
]

const categories = computed(() => ['All', ...Array.from(new Set(agents.map((agent) => agent.category)))])

const filteredAgents = computed(() => {
  const text = query.value.trim().toLowerCase()
  return agents.filter((agent) => {
    const categoryMatch = activeCategory.value === 'All' || agent.category === activeCategory.value
    const textMatch = !text || `${agent.name} ${agent.category} ${agent.description}`.toLowerCase().includes(text)
    return categoryMatch && textMatch
  })
})

async function copyAgent(agent: AgentTemplate) {
  await navigator.clipboard.writeText(agent.prompt)
  copiedAgent.value = agent.name
  window.setTimeout(() => {
    if (copiedAgent.value === agent.name) copiedAgent.value = ''
  }, 1200)
}
</script>
