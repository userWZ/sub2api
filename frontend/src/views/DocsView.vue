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
            class="hidden shrink-0 text-sm font-medium text-slate-500 transition hover:text-slate-900 sm:inline"
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
            <a href="#step-6-config" class="secondary-action">{{ copy.apiAction }}</a>
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

        <section id="support" class="docs-section">
          <h2>{{ copy.needTitle }}</h2>
          <p class="section-lead">{{ copy.needIntro }}</p>
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
          <div class="tutorial-image-grid mt-6">
            <figure
              v-for="image in copy.supportImages"
              :key="image.src"
              :class="['tutorial-image-card', image.variant ? `tutorial-image-card-${image.variant}` : '']"
            >
              <img :src="image.src" :alt="image.alt" loading="lazy" />
              <figcaption>{{ image.caption }}</figcaption>
            </figure>
          </div>
        </section>

        <section id="quick-start" class="docs-section">
          <h2>{{ copy.quickStartTitle }}</h2>
          <p class="section-lead">{{ copy.quickStartIntro }}</p>

          <div class="mt-8 space-y-8">
            <article
              v-for="section in copy.tutorialSections"
              :id="section.id"
              :key="section.id"
              class="tutorial-block"
            >
              <div class="flex flex-wrap items-center gap-3">
                <span v-if="section.step" class="step-number">{{ section.step }}</span>
                <span class="tutorial-label">{{ section.label }}</span>
              </div>
              <h3>{{ section.title }}</h3>
              <div class="mt-4 space-y-3">
                <p
                  v-for="paragraph in section.paragraphs"
                  :key="paragraph"
                  class="tutorial-paragraph"
                >
                  <template
                    v-for="(part, partIndex) in linkifyText(paragraph)"
                    :key="`${paragraph}-${partIndex}`"
                  >
                    <a
                      v-if="part.href"
                      :href="part.href"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-doc-link"
                    >
                      {{ part.text }}
                    </a>
                    <template v-else>{{ part.text }}</template>
                  </template>
                </p>
              </div>
              <ul v-if="section.bullets?.length" class="tutorial-list">
                <li v-for="bullet in section.bullets" :key="bullet">
                  <template
                    v-for="(part, partIndex) in linkifyText(bullet)"
                    :key="`${bullet}-${partIndex}`"
                  >
                    <a
                      v-if="part.href"
                      :href="part.href"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-doc-link"
                    >
                      {{ part.text }}
                    </a>
                    <template v-else>{{ part.text }}</template>
                  </template>
                </li>
              </ul>
              <p v-if="section.callout" class="tutorial-callout">
                <template
                  v-for="(part, partIndex) in linkifyText(section.callout)"
                  :key="`callout-${section.id}-${partIndex}`"
                >
                  <a
                    v-if="part.href"
                    :href="part.href"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="inline-doc-link"
                  >
                    {{ part.text }}
                  </a>
                  <template v-else>{{ part.text }}</template>
                </template>
              </p>
              <div v-if="section.images?.length" class="tutorial-image-grid mt-5">
                <figure
                  v-for="image in section.images"
                  :key="image.src"
                  :class="['tutorial-image-card', image.variant ? `tutorial-image-card-${image.variant}` : '']"
                >
                  <img :src="image.src" :alt="image.alt" loading="lazy" />
                  <figcaption>{{ image.caption }}</figcaption>
                </figure>
              </div>
            </article>
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
type Endpoint = { method: string; path: string; description: string }
type ErrorRow = { code: string; meaning: string; fix: string }
type FaqItem = { question: string; answer: string }
type CodeBlock = { key: string; title: string; description?: string; code: string }
type DocImage = { src: string; alt: string; caption: string; variant?: 'wide' | 'compact' | 'qr' | 'pair' }
type LinkifiedPart = { text: string; href?: string }
type TutorialSection = {
  id: string
  step?: string
  label: string
  title: string
  paragraphs: string[]
  bullets?: string[]
  callout?: string
  images?: DocImage[]
}

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

const imageBase = '/docs/oceanway-codex'
const urlPattern = /https?:\/\/[^\s，。；、)）]+/g
const docImages = {
  homepage: `${imageBase}/page-01-img-01-X16.png`,
  groupQr: `${imageBase}/page-02-img-01-X43.jpg`,
  supportQr: `${imageBase}/page-03-img-02-X46.png`,
  register: `${imageBase}/page-04-img-01-X66.png`,
  billing: `${imageBase}/page-05-img-01-X117.png`,
  redeem: `${imageBase}/page-05-img-02-X127.png`,
  redeemBalance: `${imageBase}/page-06-img-01-X131.png`,
  redeemSubscription: `${imageBase}/page-06-img-02-X132.png`,
  subscriptions: `${imageBase}/page-06-img-03-X135.png`,
  dashboardKey: `${imageBase}/page-07-img-01-X148.png`,
  apiKeys: `${imageBase}/page-07-img-02-X151.png`,
  windowsStart: `${imageBase}/page-08-img-01-X157.jpg`,
  windowsStore: `${imageBase}/page-08-img-02-X158.jpg`,
  codexWebsite: `${imageBase}/page-08-img-03-X160.png`,
  configTool: `${imageBase}/page-10-img-01-X205.png`
}

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
    { href: '#support', label: '官网与支持' },
    { href: '#quick-start', label: '图文步骤' },
    { href: '#api', label: 'API 补充' },
    { href: '#errors', label: '错误排查' },
    { href: '#faq', label: 'Q&A' }
  ] satisfies NavItem[],
  needTitle: '官网、讨论组和客服入口',
  needIntro: '使用 API 有任何问题，可以加入讨论组，或者扫码添加微信咨询开通与使用方式。',
  needCards: [
    { href: apiHost.value, title: '官网首页', description: `访问 ${apiHost.value} 进入 OceanWay AI。` },
    { href: '#step-1-register', title: '注册账号', description: `访问 ${registerUrl.value}，用邮箱注册并登录控制台。` },
    { href: '#step-2-quota', title: '额度获取使用', description: '注册赠送额度可先体验，也可以站内充值、购买订阅或兑换兑换码。' },
    { href: '#step-6-config', title: '一键配置', description: '输入 API Key，点击一键配置，再重新打开 Codex 即可使用。' }
  ] satisfies GuideCard[],
  supportImages: [
    { src: docImages.homepage, alt: 'OceanWay AI 官网首页截图', caption: '官网首页', variant: 'wide' },
    { src: docImages.groupQr, alt: 'OceanWay AI 讨论 2 群二维码', caption: '讨论组二维码', variant: 'qr' },
    { src: docImages.supportQr, alt: '微信扫码咨询开通与使用方式二维码', caption: '微信扫码咨询开通与使用方式', variant: 'compact' }
  ] satisfies DocImage[],
  quickStartTitle: '按以下步骤获取 API 密钥',
  quickStartIntro: '以下结构按原始教程重写：先注册账号和获取额度，再获取 API Key，随后安装 Codex、运行一键配置并验证可用。',
  tutorialSections: [
    {
      id: 'step-1-register',
      step: '1',
      label: '注册账号',
      title: '访问官网注册账号',
      paragraphs: [`访问 ${registerUrl.value}，使用邮箱注册 OceanWay AI 账号。注册完成后登录控制台。`],
      images: [
        { src: docImages.register, alt: 'OceanWay AI 注册页面截图', caption: '邮箱注册页面', variant: 'wide' }
      ]
    },
    {
      id: 'step-2-quota',
      step: '2',
      label: '额度获取使用',
      title: '先领取或购买可用额度',
      paragraphs: [
        '注册就可以获取 1 刀额度，建议先体验再充值。',
        '新用户注册后，可以联系微信客服领取 10 刀试用额度。',
        '系统内包括订阅和额度充值两种额度获取方式，可以通过站内充值，也可站外购买兑换码。'
      ],
      bullets: [
        `站内充值/订阅：访问 ${subscriptionsUrl.value}，切换额度套餐和订阅套餐。`,
        `站外充值：主要通过购买兑换码，再到 ${redeemUrl.value} 输入兑换码。`,
        '兑换成功后可看到兑换结果；订阅在我的订阅中查看，额度在左上角余额查看。'
      ],
      images: [
        { src: docImages.billing, alt: '充值订阅页面切换额度套餐和订阅套餐截图', caption: '站内充值/订阅入口', variant: 'wide' },
        { src: docImages.redeem, alt: '兑换码页面输入兑换码截图', caption: '兑换码使用方式', variant: 'wide' },
        { src: docImages.redeemBalance, alt: '兑换余额成功页面截图', caption: '兑换余额成功', variant: 'compact' },
        { src: docImages.redeemSubscription, alt: '兑换订阅成功页面截图', caption: '兑换订阅成功', variant: 'compact' },
        { src: docImages.subscriptions, alt: '我的订阅页面截图', caption: '我的订阅和余额变化', variant: 'wide' }
      ]
    },
    {
      id: 'step-3-key',
      step: '3',
      label: '获取 API KEY',
      title: '优先使用仪表盘默认 Key',
      paragraphs: [
        '有别于登录使用 Codex 的方式，OceanWay AI 平台提供官方满血 GPT 最新模型 API 接入方式配置使用，因此需要在平台上获取个人 API Key。',
        `用户注册登录到系统后，可以在仪表盘 ${dashboardUrl.value} 看到自己的默认 key，直接复制默认 key 和 Base URL 即可开始使用。`,
        `如果需要试用多个 key，可以在 API 密钥页 ${keysUrl.value} 额外创建新的 key，点击复制图标即可复制 API 密钥。`
      ],
      images: [
        { src: docImages.dashboardKey, alt: '仪表盘默认 Key 和 Base URL 复制按钮截图', caption: '仪表盘默认 Key 和 Base URL', variant: 'wide' },
        { src: docImages.apiKeys, alt: 'API 密钥页面复制多个 Key 截图', caption: 'API 密钥页额外创建和复制 Key', variant: 'wide' }
      ]
    },
    {
      id: 'install-codex',
      label: '下载安装 Codex',
      title: '先安装官方 Codex 客户端',
      paragraphs: ['完成账号、额度和 API Key 准备后，再下载安装 Codex。Windows 和 Mac 的安装方式不同，任选对应系统步骤即可。'],
      images: [
        { src: docImages.codexWebsite, alt: 'Codex 官网下载页面截图', caption: 'Codex 官网下载入口', variant: 'wide' }
      ]
    },
    {
      id: 'step-4-windows',
      step: '4',
      label: 'Windows 安装',
      title: '通过 Microsoft Store 或官网安装',
      paragraphs: [
        '搜索 Microsoft Store 并打开，随后搜索 Codex 安装即可。',
        '网页访问也可使用 Microsoft Store 链接，或者直接访问 Codex 官网，方式二选一。'
      ],
      bullets: [
        'Microsoft Store: https://apps.microsoft.com/detail/9plm9xgg6vks?hl=en-US&gl=US',
        'Codex 官网: https://openai.com/zh-Hant/codex/'
      ],
      callout: '安装时如果有报错，请携带错误代码和截图到群里，我们帮助解决。',
      images: [
        { src: docImages.windowsStart, alt: 'Windows 搜索 Microsoft Store 截图', caption: '搜索并打开 Microsoft Store', variant: 'pair' },
        { src: docImages.windowsStore, alt: 'Microsoft Store 搜索 Codex 截图', caption: '在 Microsoft Store 搜索 Codex', variant: 'pair' }
      ]
    },
    {
      id: 'step-5-mac',
      step: '5',
      label: 'Mac 安装',
      title: '下载官方 macOS 安装包',
      paragraphs: [
        'Mac M 系列芯片下载链接：https://persistent.oaistatic.com/codex-app-prod/Codex.dmg',
        'Intel 芯片 Mac 下载链接：https://persistent.oaistatic.com/codex-app-prod/Codex-latest-x64.dmg',
        '下载后双击安装，按系统提示拖入 Applications。'
      ],
      callout: '注意：下载完后，要完全关闭 Codex。左面右下角图标右键，exit 退出，再使用一键配置软件进行环境配置。'
    },
    {
      id: 'step-6-config',
      step: '6',
      label: '一键配置软件',
      title: '输入 API Key 后点击一键配置',
      paragraphs: [
        'OceanWay 团队开发的一键配置环境软件，免去繁琐的环境配置工作。',
        `打开软件，输入 ${apiHost.value} 激活的 API Key，Base URL 填 ${apiHost.value}，点击一键配置，然后重新打开 Codex，用 sk 登录即可使用。`
      ],
      bullets: [
        'Windows 版本配置软件：http://download.czroad.xyz/codex-config-Windows.zip',
        'Mac Intel 版本配置软件：http://download.czroad.xyz/codex-config-macOS-intel.zip',
        'Mac M 系列版本配置软件：http://download.czroad.xyz/codex-config-macOS.zip'
      ],
      callout: 'Mac 版本如无法打开，可在终端运行：xattr -dr com.apple.quarantine ~/Downloads/codex-config.app',
      images: [
        { src: docImages.configTool, alt: 'OceanWay AI Codex 一键配置软件截图', caption: '填写 API Key 和 Base URL 后点击一键配置', variant: 'wide' }
      ]
    },
    {
      id: 'step-7-ready',
      step: '7',
      label: '开启使用',
      title: '重新打开 Codex 并验证',
      paragraphs: [
        '配置前别打开 Codex，如果打开了记得彻底退出。Windows 需要把任务栏里的也退出。',
        '重新打开后，无需任何登录环节，无需任何代理服务。发送 hi 能正常收到回复，即配置完毕。'
      ]
    }
  ] satisfies TutorialSection[],
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
    { code: 'Codex 无法回复', meaning: '配置未生效或 Key 不可用', fix: '彻底退出 Codex 后重新运行一键配置，再重新打开。Windows 也要退出任务栏中的 Codex；仍失败时重新复制 default-key。' },
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
    { question: '配置前为什么要完全退出 Codex？', answer: 'Codex 可能已经读取了旧配置。完全退出后再配置、再重新打开，可以确保新配置和 Key 被加载。' },
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
    { href: '#support', label: 'Site and support' },
    { href: '#quick-start', label: 'Visual guide' },
    { href: '#api', label: 'API supplement' },
    { href: '#errors', label: 'Troubleshooting' },
    { href: '#faq', label: 'Q&A' }
  ] satisfies NavItem[],
  needTitle: 'Site, group, and support',
  needIntro: 'If you have any issue with API usage, join the discussion group or scan the WeChat support QR code.',
  needCards: [
    { href: apiHost.value, title: 'Official site', description: `Open ${apiHost.value} to enter OceanWay AI.` },
    { href: '#step-1-register', title: 'Register', description: `Create an account at ${registerUrl.value} and sign in.` },
    { href: '#step-2-quota', title: 'Get credit', description: 'Use the sign-up credit first, then top up, subscribe, or redeem a code.' },
    { href: '#step-6-config', title: 'One-click config', description: 'Paste your API key, click one-click config, then reopen Codex.' }
  ] satisfies GuideCard[],
  supportImages: [
    { src: docImages.homepage, alt: 'OceanWay AI official homepage screenshot', caption: 'Official homepage', variant: 'wide' },
    { src: docImages.groupQr, alt: 'OceanWay AI discussion group QR code', caption: 'Discussion group QR code', variant: 'qr' },
    { src: docImages.supportQr, alt: 'WeChat support QR code', caption: 'WeChat support QR code', variant: 'compact' }
  ] satisfies DocImage[],
  quickStartTitle: 'Get an API key and connect Codex',
  quickStartIntro: 'This follows the original guide order: register, get credit, copy the API key, install Codex, run one-click config, and verify.',
  tutorialSections: [
    {
      id: 'step-1-register',
      step: '1',
      label: 'Register',
      title: 'Create an OceanWay AI account',
      paragraphs: [`Visit ${registerUrl.value}, register with email, then sign in to the console.`],
      images: [
        { src: docImages.register, alt: 'OceanWay AI registration page screenshot', caption: 'Email registration page', variant: 'wide' }
      ]
    },
    {
      id: 'step-2-quota',
      step: '2',
      label: 'Get credit',
      title: 'Prepare usable credit',
      paragraphs: [
        'New accounts receive sign-up credit. Try it before topping up.',
        'New users can contact WeChat support for trial credit when available.',
        'You can buy balance or subscriptions in-site, or redeem an external redemption code.'
      ],
      bullets: [
        `In-site top-up/subscription: open ${subscriptionsUrl.value} and switch between balance plans and subscription plans.`,
        `External recharge: buy a redemption code and enter it at ${redeemUrl.value}.`,
        'After redemption, subscriptions appear under My subscriptions and balance changes appear in the header.'
      ],
      images: [
        { src: docImages.billing, alt: 'Top-up and subscription page screenshot', caption: 'In-site top-up and subscription', variant: 'wide' },
        { src: docImages.redeem, alt: 'Redemption code page screenshot', caption: 'Redeem a code', variant: 'wide' },
        { src: docImages.redeemBalance, alt: 'Balance redemption success screenshot', caption: 'Balance redemption success', variant: 'compact' },
        { src: docImages.redeemSubscription, alt: 'Subscription redemption success screenshot', caption: 'Subscription redemption success', variant: 'compact' },
        { src: docImages.subscriptions, alt: 'My subscriptions page screenshot', caption: 'My subscriptions and balance', variant: 'wide' }
      ]
    },
    {
      id: 'step-3-key',
      step: '3',
      label: 'Get API key',
      title: 'Use the dashboard default key first',
      paragraphs: [
        'OceanWay AI connects Codex through an API key instead of direct account login.',
        `After signing in, open ${dashboardUrl.value} and copy the default-key and Base URL. This is enough for normal use.`,
        `If you need multiple keys, create and copy extra keys at ${keysUrl.value}.`
      ],
      images: [
        { src: docImages.dashboardKey, alt: 'Dashboard default key and Base URL screenshot', caption: 'Dashboard default key and Base URL', variant: 'wide' },
        { src: docImages.apiKeys, alt: 'API keys page screenshot', caption: 'Create and copy extra API keys', variant: 'wide' }
      ]
    },
    {
      id: 'install-codex',
      label: 'Install Codex',
      title: 'Install the official Codex client',
      paragraphs: ['After account, credit, and API key are ready, install Codex for your operating system.'],
      images: [
        { src: docImages.codexWebsite, alt: 'Codex official download page screenshot', caption: 'Codex official download page', variant: 'wide' }
      ]
    },
    {
      id: 'step-4-windows',
      step: '4',
      label: 'Windows install',
      title: 'Install from Microsoft Store or the Codex website',
      paragraphs: [
        'Open Microsoft Store, search for Codex, then install it.',
        'You can also use the Microsoft Store web link or the Codex official site.'
      ],
      bullets: [
        'Microsoft Store: https://apps.microsoft.com/detail/9plm9xgg6vks?hl=en-US&gl=US',
        'Codex website: https://openai.com/zh-Hant/codex/'
      ],
      callout: 'If installation fails, send the error code and screenshot to the group for help.',
      images: [
        { src: docImages.windowsStart, alt: 'Windows Microsoft Store search screenshot', caption: 'Open Microsoft Store', variant: 'pair' },
        { src: docImages.windowsStore, alt: 'Microsoft Store Codex search screenshot', caption: 'Search Codex in Microsoft Store', variant: 'pair' }
      ]
    },
    {
      id: 'step-5-mac',
      step: '5',
      label: 'Mac install',
      title: 'Download the official macOS installer',
      paragraphs: [
        'Mac Apple Silicon download: https://persistent.oaistatic.com/codex-app-prod/Codex.dmg',
        'Intel Mac download: https://persistent.oaistatic.com/codex-app-prod/Codex-latest-x64.dmg',
        'Open the dmg after download and drag Codex into Applications.'
      ],
      callout: 'Important: fully quit Codex before using the one-click config tool.'
    },
    {
      id: 'step-6-config',
      step: '6',
      label: 'One-click config',
      title: 'Paste the API key and configure',
      paragraphs: [
        'The OceanWay one-click config tool writes the connection settings for you.',
        `Paste the API key activated on ${apiHost.value}, set Base URL to ${apiHost.value}, click one-click config, then reopen Codex.`
      ],
      bullets: [
        'Windows package: http://download.czroad.xyz/codex-config-Windows.zip',
        'Mac Intel package: http://download.czroad.xyz/codex-config-macOS-intel.zip',
        'Mac Apple Silicon package: http://download.czroad.xyz/codex-config-macOS.zip'
      ],
      callout: 'If macOS blocks the app, run: xattr -dr com.apple.quarantine ~/Downloads/codex-config.app',
      images: [
        { src: docImages.configTool, alt: 'OceanWay AI Codex config tool screenshot', caption: 'Paste API key and Base URL, then click one-click config', variant: 'wide' }
      ]
    },
    {
      id: 'step-7-ready',
      step: '7',
      label: 'Start using',
      title: 'Reopen Codex and verify',
      paragraphs: [
        'Do not keep Codex open before configuring. On Windows, quit the taskbar instance too.',
        'After reopening, no proxy is needed. Send hi; receiving a normal reply means setup is complete.'
      ]
    }
  ] satisfies TutorialSection[],
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
    { code: 'Codex does not reply', meaning: 'Config not loaded or key unavailable', fix: 'Quit Codex completely, run the config tool again, then reopen Codex. If it still fails, copy the default-key again.' },
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
    { question: 'Why must I quit Codex before configuring?', answer: 'Codex may have already loaded the old config. Reopening it after configuration ensures the new config and key are loaded.' },
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

function linkifyText(text: string): LinkifiedPart[] {
  const parts: LinkifiedPart[] = []
  let lastIndex = 0

  for (const match of text.matchAll(urlPattern)) {
    const url = match[0]
    const index = match.index ?? 0
    if (index > lastIndex) {
      parts.push({ text: text.slice(lastIndex, index) })
    }
    parts.push({ text: url, href: url })
    lastIndex = index + url.length
  }

  if (lastIndex < text.length) {
    parts.push({ text: text.slice(lastIndex) })
  }

  return parts.length > 0 ? parts : [{ text }]
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
  overflow-x: hidden;
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
  white-space: nowrap;
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

.tutorial-block {
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.78);
  padding: 1.15rem;
}

.tutorial-label {
  color: #005db8;
  font-size: 0.8rem;
  font-weight: 850;
}

.tutorial-block h3 {
  margin-top: 0.9rem;
  color: #0f172a;
  font-size: 1.2rem;
  font-weight: 850;
  letter-spacing: 0;
}

.tutorial-paragraph {
  color: #475569;
  font-size: 0.92rem;
  line-height: 1.75;
}

.tutorial-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.55rem;
  padding-left: 1.25rem;
  color: #475569;
  font-size: 0.9rem;
  line-height: 1.7;
  list-style: disc;
}

.tutorial-list li {
  overflow-wrap: anywhere;
}

.inline-doc-link {
  color: #0066cc;
  font-weight: 800;
  overflow-wrap: anywhere;
  text-decoration: underline;
  text-decoration-thickness: 0.08em;
  text-underline-offset: 0.16em;
}

.inline-doc-link:hover {
  color: #004c99;
}

.tutorial-callout {
  margin-top: 1rem;
  border: 1px solid rgba(0, 160, 255, 0.22);
  border-radius: 0.5rem;
  background: rgba(232, 246, 255, 0.74);
  padding: 0.85rem 1rem;
  color: #0f3f6c;
  font-size: 0.9rem;
  font-weight: 700;
  line-height: 1.7;
}

.tutorial-image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(18rem, 100%), 1fr));
  align-items: start;
  gap: 1rem;
}

.tutorial-image-card {
  overflow: hidden;
  border: 1px solid rgba(203, 213, 225, 0.72);
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.9);
}

.tutorial-image-card-wide {
  grid-column: 1 / -1;
}

.tutorial-image-card-compact {
  justify-self: center;
  width: min(100%, 34rem);
}

.tutorial-image-card-qr {
  justify-self: center;
  width: min(100%, 22rem);
}

.tutorial-image-card-pair {
  align-self: stretch;
}

.tutorial-image-card img {
  display: block;
  width: 100%;
  max-height: 38rem;
  object-fit: contain;
  background: #f8fafc;
}

.tutorial-image-card-wide img {
  max-height: 44rem;
}

.tutorial-image-card-pair img {
  height: 18rem;
  max-height: 18rem;
}

.tutorial-image-card-qr img {
  max-height: 34rem;
}

@media (max-width: 639px) {
  .tutorial-image-card-pair img {
    height: auto;
    max-height: 28rem;
  }
}

.tutorial-image-card figcaption {
  border-top: 1px solid rgba(226, 232, 240, 0.88);
  padding: 0.75rem 0.9rem;
  color: #334155;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.45;
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
  max-width: 100%;
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
