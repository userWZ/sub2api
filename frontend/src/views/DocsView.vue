<template>
  <div class="docs-shell min-h-screen text-slate-950">
    <header class="relative z-20 border-b border-slate-200/70 bg-white/75 backdrop-blur">
      <nav class="mx-auto flex max-w-7xl items-center justify-between px-5 py-4">
        <router-link :to="homePath" class="flex min-w-0 items-center gap-3">
          <span :class="['brand-mark', siteLogo ? 'brand-mark-image' : '']">
            <img
              v-if="siteLogo"
              :src="siteLogo"
              alt=""
              class="brand-logo-image"
            />
            <span v-else>{{ logoLetters }}</span>
          </span>
          <span class="min-w-0">
            <span class="block truncate text-base font-semibold leading-5 text-slate-700">
              {{ siteName }} Docs
            </span>
            <span class="block truncate text-sm leading-5 text-slate-500">
              {{ copy.headerSubtitle }}
            </span>
          </span>
        </router-link>

        <div class="flex items-center gap-3 sm:gap-5">
          <LocaleSwitcher />
          <router-link
            to="/agents"
            class="shrink-0 text-sm font-medium text-slate-500 transition hover:text-slate-900"
          >
            {{ copy.agentsHub }}
          </router-link>
          <router-link
            :to="homePath"
            class="hidden text-sm font-medium text-slate-500 transition hover:text-slate-900 sm:inline"
          >
            {{ copy.mainSite }}
          </router-link>
          <router-link to="/login" class="nav-button">
            {{ copy.signIn }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="relative z-10 mx-auto grid max-w-7xl gap-8 px-5 py-8 lg:grid-cols-[17rem_1fr]">
      <aside class="hidden lg:block">
        <nav class="sticky top-6 space-y-6 rounded-lg border border-slate-200/80 bg-white/72 p-4 shadow-sm">
          <div>
            <p class="px-2 text-xs font-bold uppercase tracking-[0.14em] text-slate-400">
              {{ copy.contents }}
            </p>
            <div class="mt-3 space-y-1">
              <a
                v-for="item in copy.nav"
                :key="item.href"
                :href="item.href"
                class="block rounded-md px-2 py-2 text-sm font-semibold text-slate-600 transition hover:bg-sky-50 hover:text-sky-800"
              >
                {{ item.label }}
              </a>
            </div>
          </div>

          <div class="rounded-lg bg-slate-950 p-4 text-white">
            <p class="text-xs font-semibold uppercase tracking-[0.14em] text-sky-200">
              Codex Base URL
            </p>
            <p class="mt-2 break-words font-mono text-sm text-slate-200">{{ apiHost }}</p>
          </div>
        </nav>
      </aside>

      <div class="min-w-0">
        <section id="overview" class="docs-hero">
          <div class="eyebrow-pill">
            <span class="h-2 w-2 rounded-full bg-sky-500"></span>
            {{ copy.eyebrow }}
          </div>
          <h1 class="mt-6 max-w-4xl text-4xl font-black leading-tight tracking-normal text-slate-950 sm:text-5xl">
            {{ copy.titleLead }}
            <span class="hero-highlight">{{ copy.titleHighlight }}</span>
          </h1>
          <p class="mt-6 max-w-3xl text-lg leading-8 text-slate-600">
            {{ copy.description }}
          </p>

          <div class="mt-8 flex flex-wrap gap-3">
            <a href="#quick-start" class="primary-action">{{ copy.quickStartAction }}</a>
            <a href="#clients" class="secondary-action">{{ copy.apiAction }}</a>
          </div>

          <dl class="mt-9 grid gap-3 md:grid-cols-3">
            <div
              v-for="fact in copy.facts"
              :key="fact.label"
              class="fact-panel"
            >
              <dt class="text-xs font-bold uppercase tracking-[0.12em] text-slate-400">
                {{ fact.label }}
              </dt>
              <dd class="mt-2 break-words text-sm font-semibold text-slate-800">
                {{ fact.value }}
              </dd>
            </div>
          </dl>
        </section>

        <section class="docs-section">
          <h2>{{ copy.needTitle }}</h2>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <a
              v-for="card in copy.needCards"
              :key="card.title"
              :href="card.href"
              class="guide-card"
            >
              <p class="text-sm font-bold text-slate-950">{{ card.title }}</p>
              <p class="mt-2 text-sm leading-6 text-slate-600">{{ card.description }}</p>
            </a>
          </div>
        </section>

        <section id="quick-start" class="docs-section">
          <h2>{{ copy.quickStartTitle }}</h2>
          <p class="section-lead">{{ copy.quickStartIntro }}</p>

          <div class="mt-6 grid gap-4 md:grid-cols-3">
            <article
              v-for="(step, index) in copy.steps"
              :key="step.title"
              class="step-panel"
            >
              <span class="step-number">{{ index + 1 }}</span>
              <h3>{{ step.title }}</h3>
              <p>{{ step.description }}</p>
            </article>
          </div>

          <div class="mt-8 space-y-5">
            <CodeSnippet
              v-for="block in quickStartBlocks"
              :key="block.key"
              :title="block.title"
              :code="block.code"
              :copied="copiedKey === block.key"
              :copy-label="copy.copy"
              :copied-label="copy.copied"
              @copy="copyCode(block.key, block.code)"
            />
          </div>
        </section>

        <section id="clients" class="docs-section">
          <h2>{{ copy.clientsTitle }}</h2>
          <p class="section-lead">{{ copy.clientsIntro }}</p>

          <div class="mt-6 space-y-5">
            <CodeSnippet
              v-for="block in clientBlocks"
              :key="block.key"
              :title="block.title"
              :description="block.description"
              :code="block.code"
              :copied="copiedKey === block.key"
              :copy-label="copy.copy"
              :copied-label="copy.copied"
              @copy="copyCode(block.key, block.code)"
            />
          </div>
        </section>

        <section id="api" class="docs-section">
          <h2>{{ copy.apiTitle }}</h2>
          <p class="section-lead">{{ copy.apiIntro }}</p>

          <div class="mt-6 grid gap-4 md:grid-cols-3">
            <article
              v-for="endpoint in copy.endpoints"
              :key="endpoint.path"
              class="endpoint-panel"
            >
              <p class="font-mono text-xs font-bold text-sky-700">{{ endpoint.method }}</p>
              <h3>{{ endpoint.path }}</h3>
              <p>{{ endpoint.description }}</p>
            </article>
          </div>

          <div class="mt-8 space-y-5">
            <CodeSnippet
              v-for="block in apiBlocks"
              :key="block.key"
              :title="block.title"
              :code="block.code"
              :copied="copiedKey === block.key"
              :copy-label="copy.copy"
              :copied-label="copy.copied"
              @copy="copyCode(block.key, block.code)"
            />
          </div>
        </section>

        <section id="errors" class="docs-section">
          <h2>{{ copy.errorsTitle }}</h2>
          <p class="section-lead">{{ copy.errorsIntro }}</p>
          <div class="mt-6 overflow-hidden rounded-lg border border-slate-200 bg-white/82">
            <div
              v-for="row in copy.errorRows"
              :key="row.code"
              class="grid gap-2 border-b border-slate-100 px-4 py-4 last:border-b-0 md:grid-cols-[7rem_1fr_1.4fr]"
            >
              <p class="font-mono text-sm font-bold text-slate-950">{{ row.code }}</p>
              <p class="text-sm font-semibold text-slate-700">{{ row.meaning }}</p>
              <p class="text-sm leading-6 text-slate-600">{{ row.fix }}</p>
            </div>
          </div>
        </section>

        <section id="faq" class="docs-section">
          <h2>{{ copy.faqTitle }}</h2>
          <div class="mt-6 space-y-3">
            <details
              v-for="item in copy.faq"
              :key="item.question"
              class="faq-item"
            >
              <summary>{{ item.question }}</summary>
              <p>{{ item.answer }}</p>
            </details>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { resolveHomePathForHost } from '@/utils/homeDomain'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import CodeSnippet from '@/components/common/CodeSnippet.vue'

type NavItem = { href: string; label: string }
type Fact = { label: string; value: string }
type GuideCard = { href: string; title: string; description: string }
type Step = { title: string; description: string }
type Endpoint = { method: string; path: string; description: string }
type ErrorRow = { code: string; meaning: string; fix: string }
type FaqItem = { question: string; answer: string }
type CodeBlock = { key: string; title: string; description?: string; code: string }

const appStore = useAppStore()
const { locale } = useI18n()
const copiedKey = ref('')

const isZh = computed(() => locale.value.toLowerCase().startsWith('zh'))
const homePath = computed(() => resolveHomePathForHost(appStore.cachedPublicSettings ?? window.__APP_CONFIG__))
const siteName = computed(() => {
  const configured = appStore.cachedPublicSettings?.site_name || appStore.siteName
  return configured?.trim() || 'OceanWay AI'
})
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const apiBaseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl || '')

const logoLetters = computed(() => {
  const uppercaseLetters = siteName.value.match(/[A-Z]/g)?.slice(0, 2).join('')
  return uppercaseLetters || siteName.value.slice(0, 2).toUpperCase()
})

const endpointBase = computed(() => {
  const configured = apiBaseUrl.value.trim().replace(/\/+$/, '')
  if (configured) return configured
  const origin = getCurrentOrigin()
  return isLocalOrigin(origin) ? 'https://ocean-way.top/v1' : `${origin}/v1`
})

const apiHost = computed(() => endpointBase.value.replace(/\/v1$/i, ''))
const registerUrl = computed(() => `${apiHost.value}/register`)
const dashboardUrl = computed(() => `${apiHost.value}/dashboard`)
const keysUrl = computed(() => `${apiHost.value}/keys`)
const subscriptionsUrl = computed(() => `${apiHost.value}/subscriptions`)
const redeemUrl = computed(() => `${apiHost.value}/redeem`)
const copy = computed(() => isZh.value ? zhCopy.value : enCopy.value)

const quickStartBlocks = computed<CodeBlock[]>(() => [
  {
    key: 'codex-entry',
    title: isZh.value ? '常用入口' : 'Useful links',
    code: `${siteName.value}: ${apiHost.value}
注册账号: ${registerUrl.value}
用户仪表盘: ${dashboardUrl.value}
API 密钥: ${keysUrl.value}
我的订阅: ${subscriptionsUrl.value}
兑换码: ${redeemUrl.value}

Windows Codex: https://apps.microsoft.com/detail/9plm9xgg6vks?hl=en-US&gl=US
Codex 官网: https://openai.com/zh-Hant/codex/
macOS Codex: https://persistent.oaistatic.com/codex-app-prod/Codex.dmg`
  },
  {
    key: 'default-key',
    title: isZh.value ? '默认 Key 获取路径' : 'Default key path',
    code: isZh.value
      ? `1. 登录 ${dashboardUrl.value}
2. 在仪表盘右侧找到 default-key
3. 点击复制 API Key
4. 点击复制 Base URL
5. 如需多个 Key，再到 ${keysUrl.value} 创建`
      : `1. Sign in at ${dashboardUrl.value}
2. Find default-key on the dashboard
3. Copy the API key
4. Copy the Base URL
5. Create extra keys at ${keysUrl.value} only when needed`
  },
  {
    key: 'quota-paths',
    title: isZh.value ? '额度、订阅和兑换码' : 'Quota, subscriptions, and redemption',
    code: isZh.value
      ? `注册后先使用默认赠送额度体验。
新用户可联系客服领取试用额度。
站内购买: ${subscriptionsUrl.value}
站外兑换码: ${redeemUrl.value}
购买或兑换订阅后，在“我的订阅”查看分组。
购买或兑换额度后，在左上角余额查看变化。`
      : `Try the default sign-up credit first.
Contact support for trial credit when available.
Buy in-site: ${subscriptionsUrl.value}
Redeem code: ${redeemUrl.value}
After subscribing, check My subscriptions.
After topping up credit, check the balance in the header.`
  }
])

const clientBlocks = computed<CodeBlock[]>(() => [
  {
    key: 'one-click-config',
    title: isZh.value ? '一键配置软件' : 'One-click Codex config tool',
    description: isZh.value
      ? '配置前先完全退出 Codex。Windows 需要把任务栏里的 Codex 也退出。'
      : 'Fully quit Codex before configuring. On Windows, also quit it from the taskbar.',
    code: isZh.value
      ? `1. 从讨论组或客服处获取对应系统的一键配置软件
2. 解压并打开 codex-config
3. API Key 填入仪表盘复制的 default-key
4. Base URL 填入 ${apiHost.value}
5. 点击测试连接
6. 点击一键配置
7. 重新打开 Codex，看到 ${siteName.value} 标记后发送 hi 验证`
      : `1. Get the matching config tool from support
2. Unzip and open codex-config
3. Paste the default-key from the dashboard
4. Set Base URL to ${apiHost.value}
5. Test the connection
6. Click one-click config
7. Reopen Codex and send hi after the ${siteName.value} badge appears`
  },
  {
    key: 'mac-quarantine',
    title: isZh.value ? 'macOS 配置器无法打开时' : 'macOS quarantine command',
    description: isZh.value
      ? '如果 macOS 阻止打开配置器，在终端运行后再打开。'
      : 'Run this in Terminal if macOS blocks the config app.',
    code: `xattr -dr com.apple.quarantine ~/Downloads/codex-config.app`
  },
  {
    key: 'codex-manual',
    title: isZh.value ? '手动 Codex 配置备用' : 'Manual Codex config fallback',
    description: isZh.value
      ? '一键配置不可用时，可以按这个 Provider 结构手动配置。'
      : 'Use this provider structure if the config tool is unavailable.',
    code: `disable_response_storage = true
model = "your-model"
model_provider = "${siteName.value}"
model_reasoning_effort = "high"

[model_providers."${siteName.value}"]
name = "${siteName.value}"
base_url = "${endpointBase.value}"
requires_openai_auth = true
wire_api = "responses"`
  },
  {
    key: 'auth-json',
    title: 'auth.json',
    description: isZh.value
      ? 'API Key 使用默认 Key 或你在 API 密钥页额外创建的 Key。'
      : 'Use your default key or an extra key created on the API keys page.',
    code: `{
  "OPENAI_API_KEY": "sk-your-token"
}`
  }
])

const apiBlocks = computed<CodeBlock[]>(() => [
  {
    key: 'models',
    title: 'GET /v1/models',
    code: `curl ${endpointBase.value}/models \\
  -H "Authorization: Bearer sk-your-token"`
  },
  {
    key: 'chat-completions',
    title: 'POST /v1/chat/completions',
    code: `curl ${endpointBase.value}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-token" \\
  -d '{
    "model": "your-model",
    "messages": [
      {"role": "user", "content": "Hello from ${siteName.value}"}
    ]
  }'`
  },
  {
    key: 'responses',
    title: 'POST /v1/responses',
    code: `curl ${endpointBase.value}/responses \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-token" \\
  -d '{
    "model": "your-model",
    "input": "Say hi from ${siteName.value}",
    "reasoning": { "effort": "medium" }
  }'`
  }
])

const zhCopy = computed(() => ({
  headerSubtitle: 'Codex 接入教程',
  mainSite: '主站',
  agentsHub: 'Agents Hub',
  signIn: '登录',
  contents: '目录',
  eyebrow: `${siteName.value} Codex`,
  titleLead: 'OceanWay AI',
  titleHighlight: 'Codex 接入教程',
  description:
    `全程无需代理。按本文步骤注册账号、获取额度、复制默认 Key，然后用一键配置软件把 Codex 接入 ${siteName.value}。使用中遇到问题，可加入讨论组或联系微信客服。`,
  quickStartAction: '按步骤接入',
  apiAction: '查看配置',
  facts: [
    { label: '官网首页', value: apiHost.value },
    { label: 'Codex Base URL', value: apiHost.value },
    { label: '推荐 Key', value: '仪表盘 default-key' }
  ] satisfies Fact[],
  nav: [
    { href: '#overview', label: '首页' },
    { href: '#quick-start', label: '接入步骤' },
    { href: '#clients', label: '一键配置' },
    { href: '#api', label: 'API 补充' },
    { href: '#errors', label: '错误排查' },
    { href: '#faq', label: 'Q&A' }
  ] satisfies NavItem[],
  needTitle: '先确认这四件事',
  needCards: [
    { href: '#quick-start', title: '注册账号', description: `访问 ${registerUrl.value}，用邮箱注册并登录控制台。` },
    { href: '#quick-start', title: '获取额度', description: '新账号先使用赠送额度；也可以站内充值、购买订阅或兑换兑换码。' },
    { href: '#quick-start', title: '复制默认 Key', description: '仪表盘中的 default-key 可以直接使用，通常无需先创建新 Key。' },
    { href: '#clients', title: '配置 Codex', description: '完全退出 Codex 后，用一键配置软件写入 Key 和 Base URL，再重新打开验证。' }
  ] satisfies GuideCard[],
  quickStartTitle: '按以下步骤获取 API 密钥并接入 Codex',
  quickStartIntro: '推荐顺序是先注册和确认额度，再复制默认 Key，最后安装 Codex 并用配置器接入。',
  steps: [
    { title: '注册账号', description: `访问 ${registerUrl.value}，使用邮箱注册并登录。` },
    { title: '获取额度', description: '注册赠送额度可先体验；新用户可联系客服领取试用额度。' },
    { title: '充值或兑换', description: '在充值/订阅页购买余额或订阅，也可以在兑换页输入兑换码。' },
    { title: '复制默认 Key', description: '登录后在仪表盘复制 default-key 和 Base URL；多 Key 场景再去 API 密钥页创建。' },
    { title: '安装 Codex', description: 'Windows 可通过 Microsoft Store 或官网安装；Mac 使用官方 dmg。' },
    { title: '一键配置', description: `配置器中 API Key 填 default-key，Base URL 填 ${apiHost.value}。` },
    { title: '重新打开验证', description: `打开 Codex 后看到 ${siteName.value} 标记，发送 hi 有回复即配置完成。` }
  ] satisfies Step[],
  clientsTitle: '一键配置 Codex',
  clientsIntro:
    '一键配置软件由 OceanWay 团队提供，适合不想手动编辑配置文件的用户。配置前请先完全退出 Codex。',
  apiTitle: 'API 调用补充',
  apiIntro: '普通 Codex 用户优先使用一键配置。需要脚本或 SDK 调用时，OpenAI 兼容接口使用带 /v1 的 Base URL。',
  endpoints: [
    { method: 'GET', path: '/v1/models', description: '查看当前令牌可见模型。' },
    { method: 'POST', path: '/v1/chat/completions', description: '兼容传统 OpenAI 对话接口。' },
    { method: 'POST', path: '/v1/responses', description: '适合新版客户端和推理任务。' }
  ] satisfies Endpoint[],
  errorsTitle: '错误排查',
  errorsIntro: '如果安装或配置时报错，优先截图并记录错误代码。API 调用失败时，再按状态码排查。',
  errorRows: [
    { code: 'Codex 无 OceanWay 标记', meaning: '配置未生效', fix: '彻底退出 Codex 后重新运行一键配置，再重新打开。Windows 也要退出任务栏中的 Codex。' },
    { code: '401', meaning: '认证失败', fix: '确认使用的是仪表盘 default-key 或 API 密钥页创建的有效 Key。' },
    { code: '403', meaning: '权限不足', fix: '检查账户是否有有效订阅、余额或对应模型权限。' },
    { code: '404', meaning: '路径不存在', fix: 'Codex 配置器填裸域；脚本调用 OpenAI 兼容 API 时填带 /v1 的地址。' },
    { code: '429', meaning: '频率或额度限制', fix: '降低并发，或检查 Key、账户余额、订阅额度和平台限额。' },
    { code: '503', meaning: '服务暂不可用', fix: '稍后重试；如果持续出现，带截图和请求时间联系讨论组或客服。' }
  ] satisfies ErrorRow[],
  faqTitle: 'Q&A',
  faq: [
    { question: '我需要自己创建 API Key 吗？', answer: '通常不需要。注册登录后，仪表盘里的 default-key 可以直接作为 Codex Key 使用。只有需要多个 Key 时，再到 API 密钥页创建。' },
    { question: 'Codex 配置器里的 Base URL 应该填什么？', answer: `按教程填裸域 ${apiHost.value}。如果你写脚本直接调 OpenAI 兼容 API，则使用 ${endpointBase.value}。` },
    { question: '配置前为什么要完全退出 Codex？', answer: 'Codex 可能已经读取了旧配置。完全退出后再配置、再重新打开，可以确保新 Provider 和 Key 被加载。' },
    { question: '没有额度可以测试吗？', answer: '新注册账号可先使用默认赠送额度；如需更多试用额度，可以联系微信客服或讨论组。' }
  ] satisfies FaqItem[],
  copy: '复制',
  copied: '已复制'
}))

const enCopy = computed(() => ({
  headerSubtitle: 'Codex setup guide',
  mainSite: 'Main site',
  agentsHub: 'Agents Hub',
  signIn: 'Sign in',
  contents: 'Contents',
  eyebrow: `${siteName.value} Codex`,
  titleLead: 'OceanWay AI',
  titleHighlight: 'Codex setup guide',
  description:
    `No proxy is required. Register, confirm your credit, copy the default key, then use the one-click config tool to connect Codex to ${siteName.value}.`,
  quickStartAction: 'Setup steps',
  apiAction: 'View config',
  facts: [
    { label: 'Site', value: apiHost.value },
    { label: 'Codex Base URL', value: apiHost.value },
    { label: 'Recommended key', value: 'Dashboard default-key' }
  ] satisfies Fact[],
  nav: [
    { href: '#overview', label: 'Home' },
    { href: '#quick-start', label: 'Setup steps' },
    { href: '#clients', label: 'One-click config' },
    { href: '#api', label: 'API supplement' },
    { href: '#errors', label: 'Troubleshooting' },
    { href: '#faq', label: 'Q&A' }
  ] satisfies NavItem[],
  needTitle: 'Check these first',
  needCards: [
    { href: '#quick-start', title: 'Register', description: `Create an account at ${registerUrl.value} and sign in.` },
    { href: '#quick-start', title: 'Get credit', description: 'Use the sign-up credit first, then top up, subscribe, or redeem a code.' },
    { href: '#quick-start', title: 'Copy default key', description: 'The dashboard default-key is ready to use. Extra keys are optional.' },
    { href: '#clients', title: 'Configure Codex', description: 'Quit Codex, write the key and Base URL with the config tool, then reopen and test.' }
  ] satisfies GuideCard[],
  quickStartTitle: 'Get an API key and connect Codex',
  quickStartIntro: 'Register and confirm credit first, copy the default key, install Codex, then configure it.',
  steps: [
    { title: 'Register', description: `Visit ${registerUrl.value} and register with email.` },
    { title: 'Get credit', description: 'Try the sign-up credit first, or contact support for trial credit.' },
    { title: 'Top up or redeem', description: 'Buy credit or a subscription in-site, or redeem an external code.' },
    { title: 'Copy default key', description: 'Copy default-key and Base URL from the dashboard. Create extra keys only when needed.' },
    { title: 'Install Codex', description: 'Install from Microsoft Store, the OpenAI Codex page, or the macOS dmg.' },
    { title: 'One-click config', description: `Paste the default key and set Base URL to ${apiHost.value}.` },
    { title: 'Reopen and test', description: `Reopen Codex, confirm the ${siteName.value} badge, and send hi.` }
  ] satisfies Step[],
  clientsTitle: 'One-click Codex config',
  clientsIntro:
    'The one-click config tool is the recommended path for regular Codex users. Fully quit Codex before configuring.',
  apiTitle: 'API supplement',
  apiIntro: 'Regular Codex users should use the config tool. Scripts and SDKs can use the OpenAI-compatible /v1 Base URL.',
  endpoints: [
    { method: 'GET', path: '/v1/models', description: 'List models visible to the current token.' },
    { method: 'POST', path: '/v1/chat/completions', description: 'Traditional OpenAI-compatible chat API.' },
    { method: 'POST', path: '/v1/responses', description: 'Useful for newer clients and reasoning tasks.' }
  ] satisfies Endpoint[],
  errorsTitle: 'Troubleshooting',
  errorsIntro: 'For install or config errors, keep the screenshot and error code. For API errors, check the status code.',
  errorRows: [
    { code: 'No OceanWay badge', meaning: 'Config not loaded', fix: 'Quit Codex completely, run the config tool again, then reopen Codex.' },
    { code: '401', meaning: 'Auth failed', fix: 'Use the dashboard default-key or a valid key from the API keys page.' },
    { code: '403', meaning: 'No permission', fix: 'Check your subscription, balance, or model permission.' },
    { code: '404', meaning: 'Wrong path', fix: 'Use the bare domain in the Codex config tool; use /v1 for direct OpenAI-compatible API calls.' },
    { code: '429', meaning: 'Rate or quota limited', fix: 'Lower concurrency or check key, account balance, subscription quota, and platform limits.' },
    { code: '503', meaning: 'Temporarily unavailable', fix: 'Retry later. If it continues, contact support with the screenshot and request time.' }
  ] satisfies ErrorRow[],
  faqTitle: 'Q&A',
  faq: [
    { question: 'Do I need to create an API key manually?', answer: 'Usually no. The dashboard default-key is ready for Codex. Create extra keys only if you need multiple keys.' },
    { question: 'Which Base URL should I use?', answer: `Use ${apiHost.value} in the Codex config tool. Use ${endpointBase.value} for direct OpenAI-compatible API calls.` },
    { question: 'Why must I quit Codex before configuring?', answer: 'Codex may have already loaded the old provider. Reopening it after configuration ensures the new provider and key are loaded.' },
    { question: 'Can I test without topping up?', answer: 'Try the sign-up credit first. Contact support for trial credit when available.' }
  ] satisfies FaqItem[],
  copy: 'Copy',
  copied: 'Copied'
}))

async function copyCode(key: string, code: string) {
  try {
    await navigator.clipboard.writeText(code)
    copiedKey.value = key
    window.setTimeout(() => {
      if (copiedKey.value === key) copiedKey.value = ''
    }, 1600)
  } catch {
    copiedKey.value = ''
  }
}

function getCurrentOrigin() {
  if (typeof window === 'undefined') return ''
  return window.location.origin.replace(/\/+$/, '')
}

function isLocalOrigin(origin: string) {
  return /^https?:\/\/(localhost|127\.0\.0\.1|\[::1\])(?::\d+)?$/i.test(origin)
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.docs-shell {
  position: relative;
  background:
    linear-gradient(180deg, rgba(242, 248, 255, 0.96) 0%, rgba(250, 252, 255, 0.94) 38%, rgba(232, 246, 255, 0.98) 100%);
}

.docs-shell::before {
  position: fixed;
  inset: 0;
  pointer-events: none;
  content: '';
  background:
    linear-gradient(110deg, rgba(255, 255, 255, 0.86) 0%, rgba(255, 255, 255, 0) 44%),
    linear-gradient(180deg, rgba(196, 222, 250, 0) 55%, rgba(196, 222, 250, 0.38) 100%);
}

.brand-mark {
  display: inline-flex;
  width: 2.375rem;
  height: 2.375rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: linear-gradient(135deg, #002080 0%, #001040 78%);
  color: rgba(255, 255, 255, 0.78);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0;
  box-shadow: 0 10px 26px rgba(15, 23, 42, 0.18);
}

.brand-mark-image {
  overflow: hidden;
  border-radius: 0.65rem;
  background: #001040;
  box-shadow: none;
}

.brand-logo-image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.nav-button,
.primary-action,
.secondary-action {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 700;
  transition:
    border-color 0.16s ease,
    background-color 0.16s ease,
    color 0.16s ease,
    transform 0.16s ease;
}

.nav-button,
.secondary-action {
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: rgba(255, 255, 255, 0.82);
  padding: 0.5rem 1rem;
  color: #111827;
}

.primary-action {
  background: #001040;
  padding: 0.72rem 1.05rem;
  color: white;
}

.primary-action:hover,
.secondary-action:hover,
.nav-button:hover {
  transform: translateY(-1px);
}

.docs-hero,
.docs-section {
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.72);
  box-shadow: 0 18px 38px rgba(15, 23, 42, 0.05);
}

.docs-hero {
  padding: 2rem;
}

.docs-section {
  margin-top: 1rem;
  padding: 1.5rem;
}

.docs-section h2 {
  color: #0f172a;
  font-size: 1.5rem;
  font-weight: 850;
  letter-spacing: 0;
}

.section-lead {
  margin-top: 0.75rem;
  max-width: 46rem;
  color: #475569;
  font-size: 0.95rem;
  line-height: 1.75;
}

.eyebrow-pill {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.6rem;
  border: 1px solid rgba(0, 160, 255, 0.24);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.7);
  padding: 0.45rem 0.85rem;
  color: #005db8;
  font-size: 0.875rem;
  font-weight: 800;
}

.hero-highlight {
  color: #006fd6;
}

.fact-panel,
.guide-card,
.step-panel,
.endpoint-panel,
.faq-item {
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.78);
  padding: 1rem;
}

.guide-card {
  transition:
    border-color 0.16s ease,
    background-color 0.16s ease,
    transform 0.16s ease;
}

.guide-card:hover {
  border-color: rgba(0, 160, 255, 0.42);
  background: rgba(232, 246, 255, 0.76);
  transform: translateY(-1px);
}

.step-number {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.5rem;
  background: #001040;
  color: white;
  font-size: 0.875rem;
  font-weight: 800;
}

.step-panel h3,
.endpoint-panel h3 {
  margin-top: 0.9rem;
  color: #0f172a;
  font-size: 1rem;
  font-weight: 800;
}

.step-panel p,
.endpoint-panel p {
  margin-top: 0.5rem;
  color: #475569;
  font-size: 0.875rem;
  line-height: 1.65;
}

:deep(.code-panel) {
  overflow: hidden;
  border: 1px solid rgba(203, 213, 225, 0.8);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.9);
}

:deep(.code-panel-header) {
  display: flex;
  min-height: 3.2rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.8rem 1rem;
}

:deep(.code-panel-title) {
  color: #0f172a;
  font-size: 0.9rem;
  font-weight: 800;
}

:deep(.code-panel-desc) {
  margin-top: 0.2rem;
  color: #64748b;
  font-size: 0.8rem;
  line-height: 1.5;
}

:deep(.copy-button) {
  flex: 0 0 auto;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 0.5rem;
  background: white;
  padding: 0.42rem 0.7rem;
  color: #334155;
  font-size: 0.75rem;
  font-weight: 800;
}

:deep(.code-pre) {
  overflow-x: auto;
  background: #001040;
  padding: 1.1rem;
  color: #d1d5db;
  font-family:
    ui-monospace,
    SFMono-Regular,
    Menlo,
    Monaco,
    Consolas,
    'Liberation Mono',
    'Courier New',
    monospace;
  font-size: 0.8125rem;
  line-height: 1.75;
}

.faq-item summary {
  cursor: pointer;
  color: #0f172a;
  font-size: 0.95rem;
  font-weight: 800;
}

.faq-item p {
  margin-top: 0.75rem;
  color: #475569;
  font-size: 0.9rem;
  line-height: 1.7;
}

@media (max-width: 640px) {
  .docs-hero,
  .docs-section {
    padding: 1.1rem;
  }
}
</style>
