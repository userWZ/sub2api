<template>
  <div class="min-h-screen bg-slate-50 text-slate-950">
    <header class="border-b border-slate-200 bg-white/85 backdrop-blur">
      <nav class="mx-auto flex max-w-6xl items-center justify-between px-5 py-4">
        <router-link :to="homePath" class="flex items-center gap-3 font-semibold">
          <img :src="siteLogo || '/logo.png'" alt="" class="h-9 w-9 rounded-lg object-contain" />
          <span>{{ siteName }} Docs</span>
        </router-link>
        <div class="flex items-center gap-4 text-sm font-medium text-slate-600">
          <LocaleSwitcher />
          <router-link to="/agents" class="hover:text-slate-950">Agents</router-link>
          <router-link :to="homePath" class="hover:text-slate-950">Home</router-link>
          <router-link to="/login" class="rounded-full bg-slate-950 px-4 py-2 text-white hover:bg-slate-800">
            Sign in
          </router-link>
        </div>
      </nav>
    </header>

    <main class="mx-auto grid max-w-6xl gap-8 px-5 py-10 lg:grid-cols-[16rem_1fr]">
      <aside class="hidden lg:block">
        <nav class="sticky top-6 rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
          <p class="text-xs font-bold uppercase tracking-[0.14em] text-slate-400">Contents</p>
          <a v-for="item in navItems" :key="item.href" :href="item.href" class="mt-3 block rounded-md px-2 py-2 text-sm font-semibold text-slate-600 hover:bg-sky-50 hover:text-sky-800">
            {{ item.label }}
          </a>
        </nav>
      </aside>

      <div class="min-w-0 space-y-8">
        <section id="overview" class="rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
          <p class="text-sm font-bold uppercase tracking-[0.16em] text-sky-700">OceanWay API</p>
          <h1 class="mt-4 text-4xl font-black tracking-normal sm:text-5xl">OpenAI-compatible access for your workspace.</h1>
          <p class="mt-5 max-w-3xl text-lg leading-8 text-slate-600">
            Use one API key with OpenAI-compatible clients. OceanWay routes requests through your active subscriptions first, then account balance.
          </p>
          <div class="mt-6 rounded-lg bg-slate-950 p-4 font-mono text-sm text-slate-100">
            {{ endpointBase }}/chat/completions
          </div>
        </section>

        <section id="quick-start" class="docs-section">
          <h2>Quick Start</h2>
          <div class="mt-5 grid gap-4 md:grid-cols-3">
            <article v-for="step in steps" :key="step.title" class="rounded-lg border border-slate-200 bg-white p-5">
              <p class="text-sm font-black text-slate-950">{{ step.title }}</p>
              <p class="mt-2 text-sm leading-6 text-slate-600">{{ step.description }}</p>
            </article>
          </div>
          <CodeBlock class="mt-6" title="cURL" :code="curlExample" />
        </section>

        <section id="clients" class="docs-section">
          <h2>Client Configuration</h2>
          <p class="section-lead">Set the base URL and API key in any OpenAI-compatible app.</p>
          <CodeBlock title="Environment" :code="envExample" />
        </section>

        <section id="endpoints" class="docs-section">
          <h2>Common Endpoints</h2>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <article v-for="endpoint in endpoints" :key="endpoint.path" class="rounded-lg border border-slate-200 bg-white p-5">
              <p class="font-mono text-xs font-bold text-sky-700">{{ endpoint.method }}</p>
              <h3 class="mt-2 font-black">{{ endpoint.path }}</h3>
              <p class="mt-2 text-sm leading-6 text-slate-600">{{ endpoint.description }}</p>
            </article>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { resolveHomePathForHost } from '@/utils/homeDomain'

const appStore = useAppStore()
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'OceanWay AI')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const endpointBase = computed(() => (appStore.cachedPublicSettings?.api_base_url || window.location.origin).replace(/\/$/, ''))
const homePath = computed(() => resolveHomePathForHost(appStore.cachedPublicSettings ?? window.__APP_CONFIG__))

const navItems = [
  { href: '#overview', label: 'Overview' },
  { href: '#quick-start', label: 'Quick start' },
  { href: '#clients', label: 'Clients' },
  { href: '#endpoints', label: 'Endpoints' },
]

const steps = [
  { title: 'Create or use your default key', description: 'New accounts receive a default key. You can create more keys from the console.' },
  { title: 'Configure Base URL', description: 'Use the OceanWay endpoint as your OpenAI-compatible base URL.' },
  { title: 'Send requests', description: 'Requests consume active subscriptions first and fall back to balance when needed.' },
]

const endpoints = [
  { method: 'POST', path: '/v1/chat/completions', description: 'OpenAI-compatible chat completions.' },
  { method: 'POST', path: '/v1/responses', description: 'Responses API, including upstream image-generation support.' },
  { method: 'POST', path: '/v1/images/generations', description: 'OpenAI image endpoint routed through group-level image capability.' },
  { method: 'GET', path: '/v1/usage', description: 'Read usage and remaining entitlement for the API key.' },
]

const curlExample = computed(() => `curl ${endpointBase.value}/v1/chat/completions \\
  -H "Authorization: Bearer sk-..." \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-5","messages":[{"role":"user","content":"Hello"}]}'`)

const envExample = computed(() => `OPENAI_BASE_URL=${endpointBase.value}/v1
OPENAI_API_KEY=sk-...`)

const CodeBlock = defineComponent({
  props: {
    title: { type: String, required: true },
    code: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'rounded-lg border border-slate-800 bg-slate-950 p-5 text-slate-100' }, [
      h('p', { class: 'mb-3 text-xs font-bold uppercase tracking-[0.14em] text-sky-300' }, props.title),
      h('pre', { class: 'overflow-x-auto whitespace-pre-wrap font-mono text-sm leading-7' }, props.code),
    ])
  },
})
</script>

<style scoped>
.docs-section {
  @apply rounded-lg border border-slate-200 bg-white p-8 shadow-sm;
}

.docs-section h2 {
  @apply text-2xl font-black tracking-normal text-slate-950;
}

.section-lead {
  @apply mt-3 max-w-3xl text-sm leading-6 text-slate-600;
}
</style>
