# OceanWay 统一 API 对外接入文档

文档版本：`1.1`

更新日期：`2026-07-13`

适用对象：通过 API 接入 OceanWay 文本、推理、生图和视频生成能力的下游用户。

> 生产状态说明：文本、Responses、同步图片接口和异步视频任务接口已经在 `new.ocean-way.top` 提供服务。视频协议以《OceanWay 视频生成 API 对接文档》2.2 版为准，现有下游无需更换路径或调用流程。OceanWay 自有媒体域名仍属于统一网关建设项；在完成媒体归档前，图片或视频结果可能是第三方临时签名 URL。客户端必须始终把结果 URL 当作不透明临时地址，不得依赖其域名和路径结构。

## 1. 接入信息

API 地址：

```text
https://new.ocean-way.top/v1
```

鉴权方式：

```http
Authorization: Bearer sk-你的API密钥
```

除图片编辑等文件上传接口外，请求体和响应体均使用 JSON：

```http
Content-Type: application/json
```

请勿在浏览器前端、公开仓库、日志或错误截图中暴露 API 密钥。若密钥泄露，请立即在控制台删除并重新生成。

## 2. 接口概览

| 能力 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| 查询模型 | `GET` | `/v1/models` | 获取当前密钥可以调用的模型 |
| Chat Completions | `POST` | `/v1/chat/completions` | OpenAI Chat Completions 兼容接口 |
| Responses | `POST` | `/v1/responses` | 推理、Codex 和工具调用接口 |
| 图片生成 | `POST` | `/v1/images/generations` | 文生图、JSON 图生图，同步返回结果 |
| 图片编辑 | `POST` | `/v1/images/edits` | 图生图、图片编辑 |
| 创建视频任务 | `POST` | `/v1/videos` | 文生视频及受支持的视频生成任务 |
| 查询视频任务 | `GET` | `/v1/videos/{task_id}` | 查询任务状态和结果 |
| 下载视频 | `GET` | `/v1/videos/{task_id}/content` | 视频完成后下载文件，需要 Bearer 鉴权 |

不同密钥拥有的模型权限可能不同。模型是否可用，以 `GET /v1/models` 的实际返回和控制台配置为准。

图片和视频共享 Base URL、Bearer 鉴权、公开模型权限和媒体链接安全规则，但不共享任务生命周期：图片生成和编辑是同步 HTTP 请求；视频生成是异步任务，必须保存任务 ID 并轮询。

## 3. 快速开始

### 3.1 查询模型

```bash
curl https://new.ocean-way.top/v1/models \
  -H "Authorization: Bearer sk-你的API密钥"
```

响应示例：

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-5.4",
      "object": "model",
      "owned_by": "oceanway"
    },
    {
      "id": "gpt-image-2",
      "object": "model",
      "owned_by": "oceanway"
    }
  ]
}
```

不要根据模型名称推断上游供应商或实际渠道。OceanWay 可能在不改变公开模型名称的情况下进行渠道切换、故障转移或容量调整。

### 3.2 最小文本请求

```bash
curl https://new.ocean-way.top/v1/chat/completions \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "user", "content": "用一句话介绍 OceanWay"}
    ]
  }'
```

## 4. 通用约定

### 4.1 模型选择

调用前建议先查询 `/v1/models`。当指定模型不可用时，接口通常返回 `404`、`403` 或 `503`，不会自动替换成语义完全不同的模型。

OceanWay 会在同一个公开模型对应的可用渠道之间自动调度。客户端不需要指定渠道、供应商或账号池。

OpenAI、Claude、Gemini、GLM 和 Grok 文本模型使用相同的 OceanWay 鉴权方式和 OpenAI 兼容请求格式，差异只体现在 `model` 字段及模型自身支持的参数。下游不应调用供应商私有域名，也不需要为每一家供应商分别维护 API Key。

供应商独有能力只有在 OceanWay 明确提供统一映射后才能使用。例如 Gemini 图片、Grok 特殊媒体接口或视频首尾帧功能，不能把供应商原始字段直接附加到通用请求中。

### 4.2 时间戳

响应中的 `created`、`created_at` 等字段为 Unix 时间戳，单位为秒。

### 4.3 请求追踪

发生错误时，请保留以下信息并提交给管理员：

- 请求时间和时区。
- 请求路径和模型名称。
- 响应中的 `request_id`、任务 `id` 或 `task_id`。
- HTTP 状态码和完整错误 JSON。

不要发送完整 API 密钥。只需提供密钥末尾 4 位用于定位。

### 4.4 超时建议

| 请求类型 | 建议客户端读取超时 |
| --- | ---: |
| 普通文本请求 | 120 秒 |
| 流式文本请求 | 600 秒 |
| 图片生成/编辑 | 1,800 秒 |
| 创建视频任务 | 60 秒 |
| 视频任务轮询 | 前台最多 30 分钟，之后转后台每 60–120 秒继续查询 |

图片和视频生成耗时受排队、分辨率、内容审核和模型负载影响。不要把同步生图接口设置成 30 秒或 60 秒超时。

### 4.5 重试原则

- 文本接口的 `429`、`502`、`503`、`504` 可以使用指数退避重试；视频创建接口只在明确返回 `429`、`502`、`503` 且尚未获得任务 ID 时有限重试，`504` 应先查询任务和消费记录。
- 建议退避间隔为 2、4、8、16 秒，并加入少量随机抖动。
- 文本请求最多自动重试 2 次。
- 图片和视频都属于高成本请求。同步图片请求超时且无法确认结果时，不要无条件重发，应先检查消费日志或联系技术支持；视频一旦获得任务 `id`，只能查询已有任务，不能重复创建。
- `400`、`401`、`403`、`404` 不应盲目重试，应先修正参数、密钥或权限。

## 5. Chat Completions

接口：

```http
POST /v1/chat/completions
```

### 5.1 非流式请求

```bash
curl https://new.ocean-way.top/v1/chat/completions \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "system", "content": "你是一个简洁的助手。"},
      {"role": "user", "content": "解释什么是向量数据库。"}
    ],
    "temperature": 0.7,
    "stream": false
  }'
```

响应示例：

```json
{
  "id": "chatcmpl_xxx",
  "object": "chat.completion",
  "created": 1783670400,
  "model": "gpt-5.4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "向量数据库是专门存储和检索向量表示的数据系统。"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 25,
    "total_tokens": 45
  }
}
```

### 5.2 流式请求

```bash
curl -N https://new.ocean-way.top/v1/chat/completions \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
    "messages": [
      {"role": "user", "content": "写一段简短的产品介绍。"}
    ],
    "stream": true
  }'
```

流式响应使用 Server-Sent Events。客户端应逐行读取 `data:`，收到以下内容后结束：

```text
data: [DONE]
```

不要对流式响应启用反向代理缓冲，否则会表现为长时间无输出后一次性返回。

### 5.3 常用参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 模型名称 |
| `messages` | array | 是 | 对话消息数组 |
| `stream` | boolean | 否 | 是否流式返回，默认 `false` |
| `temperature` | number | 否 | 采样温度，是否生效取决于模型 |
| `max_tokens` | integer | 否 | 最大输出 token 数 |
| `tools` | array | 否 | 函数工具定义，是否支持取决于模型 |
| `tool_choice` | string/object | 否 | 工具选择策略 |

不要同时向不支持某个参数的模型强制发送该参数。推理模型可能忽略或拒绝 `temperature` 等采样参数。

## 6. Responses API

接口：

```http
POST /v1/responses
```

Responses API 适合推理模型、Codex 客户端、工具调用和多轮响应。

### 6.1 基础请求

```bash
curl https://new.ocean-way.top/v1/responses \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
    "input": "分析使用异步任务处理视频生成的好处",
    "reasoning": {
      "effort": "medium"
    }
  }'
```

### 6.2 流式请求

```bash
curl -N https://new.ocean-way.top/v1/responses \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4",
    "input": "给出一个三步排查 API 超时的方法",
    "stream": true
  }'
```

客户端应按照 Responses API 的事件类型处理流式响应，不要假定所有事件都是纯文本增量。

### 6.3 Codex 生图工具

只有模型和账号明确支持内置生图工具时，才可以发送：

```json
{
  "model": "支持生图工具的模型",
  "input": "生成一张白色小狗在草地上奔跑的图片",
  "tools": [
    {"type": "image_generation"}
  ]
}
```

规则如下：

- 请求带有 `image_generation` 工具时，OceanWay 只会选择经过验证、支持该工具的能力池。
- 请求没有携带该工具时，OceanWay 不会自行注入生图工具。
- 仅仅在提示词中出现“图片”“生图”等文字，不等于客户端已经启用了工具。
- 文本账号和生图账号可能属于不同能力池，延迟和计费也可能不同。
- 若返回“当前线程未暴露内置出图工具”，通常说明所选模型、客户端或账号权限不支持该工具，应改用 `/v1/images/generations` 或联系管理员确认权限。

## 7. 图片生成

### 7.1 文生图

接口：

```http
POST /v1/images/generations
```

请求示例：

```bash
curl https://new.ocean-way.top/v1/images/generations \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "一只白色小狗坐在窗边，柔和自然光，写实摄影风格",
    "size": "1024x1024",
    "n": 1
  }'
```

响应可能返回临时 URL。以下为 OceanWay 自有媒体域名上线后的目标格式：

```json
{
  "created": 1783670400,
  "data": [
    {
      "url": "https://media.ocean-way.top/images/img_xxx.png"
    }
  ]
}
```

也可能根据模型和请求参数返回 Base64：

```json
{
  "created": 1783670400,
  "data": [
    {
      "b64_json": "iVBORw0KGgoAAA..."
    }
  ]
}
```

客户端应兼容 `data[].url` 和 `data[].b64_json` 两种结果。

常用参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 图片模型，例如 `gpt-image-2` |
| `prompt` | string | 是 | 图片描述 |
| `n` | integer | 否 | 生成数量，默认 `1` |
| `size` | string | 否 | 尺寸，支持值取决于模型 |
| `quality` | string | 否 | 质量档位，支持值取决于模型 |
| `aspect_ratio` | string | 否 | 画幅比例，例如 `1:1`、`16:9` |
| `output_resolution` | string | 否 | 输出档位，例如 `1K`、`2K`、`4K`，仅部分模型支持 |
| `image_size` | string | 否 | `output_resolution` 的兼容别名，不要传互相矛盾的值 |
| `image` | string | 否 | JSON 图生图的单张参考图 URL 或 Data URI |
| `images` | array | 否 | JSON 图生图的多张参考图 |
| `reference_images` | array | 否 | JSON 图生图的参考图兼容字段 |
| `stream` | boolean | 否 | 图片接口通常使用 `false`，是否支持取决于模型 |
| `response_format` | string | 否 | `url` 或 `b64_json`，是否生效取决于模型 |

图片接口同步返回 `data` 数组，不接受 `async` 参数，也不提供图片任务轮询路径。客户端读取超时应覆盖完整生成时间。

### 7.2 JSON 图生图

单张参考图：

```bash
curl https://new.ocean-way.top/v1/images/generations \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "支持图生图的模型",
    "prompt": "保留人物主体，把背景改成傍晚海边",
    "aspect_ratio": "16:9",
    "image": "https://cdn.example.com/reference.png",
    "stream": false
  }'
```

`image`、`images` 和 `reference_images` 可以使用 HTTPS URL 或图片 Data URI。具体参考图数量和格式限制以模型能力为准。

### 7.3 multipart 图片编辑

接口：

```http
POST /v1/images/edits
```

该接口使用 `multipart/form-data`：

```bash
curl https://new.ocean-way.top/v1/images/edits \
  -H "Authorization: Bearer sk-你的API密钥" \
  -F "model=gpt-image-2" \
  -F "prompt=把背景改成傍晚海边，保留主体" \
  -F "image=@./input.png"
```

图片文件要求、最大尺寸和支持格式取决于具体模型。建议使用 PNG、JPEG 或 WebP，并将单张输入图片控制在 20 MB 以内。

多参考图需要重复使用 `image` 字段，不要写成 `image[]`：

```bash
curl https://new.ocean-way.top/v1/images/edits \
  -H "Authorization: Bearer sk-你的API密钥" \
  -F "model=支持多图编辑的模型" \
  -F "prompt=结合两张参考图生成新场景" \
  -F "image=@./reference-1.png" \
  -F "image=@./reference-2.png"
```

### 7.4 图片链接处理

- 将结果 URL 视为不透明字符串，不要解析或依赖其域名和路径结构。
- 图片链接可能带有有效期，业务需要长期保存时请及时下载并存入自己的对象存储。
- 不要在日志中完整记录带签名参数的 URL。
- URL 下载失败时，请提供生成请求的 `request_id`，不要只提供已经过期的链接。

## 8. 视频生成

视频采用异步任务模式：先创建任务，再轮询任务状态。不要保持一个 HTTP 请求等待视频生成完成。

### 8.1 支持模型

当前建议使用以下公开模型名：

| 档位 | 480p | 720p |
| --- | --- | --- |
| 标准版 | `seedance-2.0-480p` | `seedance-2.0-720p` |
| Mini | `seedance-2.0-mini-480p` | `seedance-2.0-mini-720p` |
| Fast | `seedance-2.0-fast-480p` | `seedance-2.0-fast-720p` |

最终模型和价格以控制台对当前密钥显示的配置为准。模型可能因容量、维护或权限暂时不可用。

### 8.2 创建任务

接口：

```http
POST /v1/videos
```

请求示例：

```bash
curl https://new.ocean-way.top/v1/videos \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedance-2.0-fast-480p",
    "prompt": "雨夜霓虹街道，镜头缓慢向前推进，电影感光影，无人物，无文字",
    "aspect_ratio": "16:9",
    "duration": 5
  }'
```

#### 8.2.1 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 视频模型名称，清晰度包含在模型名中 |
| `prompt` | string | 是 | 视频描述，最长 5,000 字符；多素材时可用 `@image1`、`@video1`、`@audio1` 或命名引用 |
| `duration` | integer | 是 | 时长，Seedance 2.0 通常支持 4 至 15 秒任意整数 |
| `aspect_ratio` | string | 否 | 默认 `16:9`；支持 `16:9`、`9:16`、`1:1`、`21:9`、`3:4`、`4:3` |
| `resolution` | string | 否 | `480p` 或 `720p`；固定分辨率模型通常无需再传 |
| `audio` | boolean | 否 | 是否生成原生音频；仅部分非固定分辨率模型支持，默认值由模型决定 |
| `image_url` | string | 否 | 主参考图，支持 HTTPS URL 或 `data:image/...;base64,...` |
| `reference_image_urls` | array | 否 | 额外参考图数组；元素可以是 URL 字符串或带名称的对象 |
| `reference_images` | array | 否 | 命名参考图，推荐格式为 `[{"url":"...","name":"角色名"}]` |
| `reference_image_names` | array | 否 | 与 `reference_image_urls` 同序的名称数组 |
| `reference_videos` | array | 否 | 参考视频 HTTPS URL 数组 |
| `reference_audios` | array | 否 | 参考音频 HTTPS URL 数组 |
| `first_image_url` | string | 否 | 首帧，必须与 `last_image_url` 成对出现 |
| `last_image_url` | string | 否 | 尾帧，必须与 `first_image_url` 成对出现 |
| `image` | file | 否 | `multipart/form-data` 单图上传字段；多参考图请使用 JSON 数组 |

当前视频接口不对外支持 `negative_prompt`。需要表达排除项时，请将“不要出现字幕、水印、Logo、闪烁、变形”等要求直接写入 `prompt`。

#### 8.2.2 生成模式

服务端根据素材字段自动判断生成模式，不需要额外传 `mode`：

| 模式 | 最少必传 | 触发字段 | 主要规则 |
| --- | --- | --- | --- |
| 文生视频 | `prompt` | 不传任何素材字段 | 仅根据文本生成 |
| 单图生视频 | `prompt` + 1 张图 | 仅 `image_url`，或 multipart `image` | 主体或场景以参考图为基础 |
| 多模态/全能参考 | `prompt` + 至少 1 张主图 | `reference_image_urls`、`reference_images`、`reference_videos`、`reference_audios` | 视频和音频参考必须搭配主参考图 |
| 首尾帧 | `prompt` + 首帧 + 尾帧 | `first_image_url` + `last_image_url` | 两个字段必须成对，与多模态参考互斥 |

多模态提示词中的素材序号从 1 开始：

```text
@image1 ... @image9
@video1 ... @video3
@audio1 ... @audio3
```

使用 `reference_images` 的 `name` 后，也可以直接在提示词中使用命名引用，例如 `@志强`、`@清雅`。

#### 8.2.3 Seedance 2.0 素材限制

当前已接入的 Seedance 模型存在两类参数档案，必须按实际公开模型名判断，不能混用上限：

| 模型类型 | 参考图 | 参考视频 | 参考音频 | 原生音频开关 |
| --- | ---: | ---: | ---: | --- |
| `seedance-2.0`、`seedance-2.0-fast`、`seedance-2.0-mini` | 合计最多 4 张 | 最多 3 条，单条 4–15 秒，总时长不超过 15 秒 | 最多 1 条，不超过 15 秒 | 支持 `audio` |
| 带 `-480p`、`-720p` 等固定分辨率后缀的模型 | 合计最多 9 张 | 最多 3 条，单条 2–15 秒，总时长不超过 15 秒 | 最多 3 条，必须搭配主图 | 以该模型文档为准 |

固定分辨率模型的素材要求：

- 参考图：JPEG、PNG 或 WebP；单张不超过 30 MB；每边至少 300 px；宽高比 0.4–2.5；长边不超过 4,000 px。
- 参考视频：MP4 或 MOV；24–60 fps；单条不超过 50 MB；多条总时长不超过 15 秒。
- 参考音频：MP3、WAV、M4A 等常见格式；单条不超过 15 MB；必须同时提供至少一张主参考图。
- `image_url`、首帧和尾帧支持 HTTPS URL 或图片 Data URI。
- `reference_videos` 和 `reference_audios` 必须使用可由服务端下载的 HTTPS 公网 URL。
- 首尾帧模式不得再传 `image_url`、`reference_image_urls`、`reference_images`、`reference_videos` 或 `reference_audios`。

#### 8.2.4 单图生视频示例

```json
{
  "model": "seedance-2.0-fast-480p",
  "prompt": "保持人物一致，人物缓慢向前走动",
  "duration": 5,
  "aspect_ratio": "16:9",
  "image_url": "https://cdn.example.com/person.jpg"
}
```

参考图也可以使用 Data URI：

```json
{
  "model": "seedance-2.0-fast-480p",
  "prompt": "让画面自然动起来",
  "duration": 5,
  "image_url": "data:image/png;base64,iVBORw0KGgo..."
}
```

#### 8.2.5 多参考图与角色绑定示例

序号引用：

```json
{
  "model": "seedance-2.0-mini-720p",
  "prompt": "@image1 的人物在 @image2 的场景中行走",
  "duration": 8,
  "aspect_ratio": "16:9",
  "image_url": "https://cdn.example.com/person.jpg",
  "reference_image_urls": [
    "https://cdn.example.com/scene.jpg"
  ]
}
```

命名角色引用：

```json
{
  "model": "seedance-2.0-720p",
  "prompt": "@志强 与 @清雅 在医院走廊相遇",
  "duration": 10,
  "aspect_ratio": "9:16",
  "reference_images": [
    {
      "name": "志强",
      "url": "https://cdn.example.com/zhiqiang.jpg"
    },
    {
      "name": "清雅",
      "url": "https://cdn.example.com/qingya.jpg"
    }
  ]
}
```

#### 8.2.6 图、视频和音频全能参考示例

```json
{
  "model": "seedance-2.0-720p",
  "prompt": "以 @image1 的人物、@video1 的动作和运镜，配合 @audio1 的节奏生成广告片",
  "duration": 10,
  "aspect_ratio": "16:9",
  "image_url": "https://cdn.example.com/main-character.jpg",
  "reference_image_urls": [
    "https://cdn.example.com/product.jpg"
  ],
  "reference_videos": [
    "https://cdn.example.com/camera-motion.mp4"
  ],
  "reference_audios": [
    "https://cdn.example.com/music.mp3"
  ]
}
```

`reference_videos` 或 `reference_audios` 不能单独使用，必须至少同时提供 `image_url`、`reference_image_urls` 或 `reference_images` 中的一张主图。

#### 8.2.7 首尾帧示例

```json
{
  "model": "seedance-2.0-fast-720p",
  "prompt": "从清晨城市平滑过渡到夜晚霓虹，保持镜头位置一致",
  "duration": 5,
  "first_image_url": "https://cdn.example.com/start.jpg",
  "last_image_url": "https://cdn.example.com/end.jpg"
}
```

以下请求属于参数冲突，网关应返回 `400`，不能自行删除其中某些素材：

```json
{
  "first_image_url": "https://cdn.example.com/start.jpg",
  "last_image_url": "https://cdn.example.com/end.jpg",
  "reference_videos": ["https://cdn.example.com/ref.mp4"]
}
```

#### 8.2.8 原生音频

支持原生音频开关的模型可以使用：

```json
{
  "model": "seedance-2.0-fast",
  "prompt": "雨夜街道，汽车驶过积水，保留环境声和轮胎水声",
  "duration": 8,
  "resolution": "720p",
  "aspect_ratio": "16:9",
  "audio": true
}
```

#### 8.2.9 multipart 单图上传

```bash
curl https://new.ocean-way.top/v1/videos \
  -H "Authorization: Bearer sk-你的API密钥" \
  -F "model=seedance-2.0-fast-480p" \
  -F "prompt=保持主体一致，让画面自然动起来" \
  -F "duration=5" \
  -F "aspect_ratio=16:9" \
  -F "image=@./photo.jpg"
```

multipart 方式仅建议用于单张主图。多图、参考视频、参考音频和命名角色应使用 JSON 请求及 HTTPS 素材 URL。

#### 8.2.10 创建任务响应

创建成功示例：

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "",
  "model": "",
  "status": "queued",
  "progress": 0,
  "created_at": 1783670400
}
```

`object`、`model`、`seconds` 和 `size` 等字段可能为空或不存在，客户端不能依赖这些可选字段。客户端必须优先保存创建响应的 `id`；只有旧版响应缺少 `id` 时才兼容读取 `task_id`。后续查询、下载、账单核对和故障定位始终使用保存的任务 ID。

### 8.3 查询任务

接口：

```http
GET /v1/videos/{task_id}
```

请求示例：

```bash
curl https://new.ocean-way.top/v1/videos/task_xxx \
  -H "Authorization: Bearer sk-你的API密钥"
```

任务状态：

| 状态 | 含义 | 客户端行为 |
| --- | --- | --- |
| `queued` | 已提交，等待处理 | 继续轮询 |
| `in_progress` | 正在生成 | 继续轮询 |
| `completed` | 已完成 | 读取结果链接 |
| `failed` | 生成失败 | 停止轮询并记录错误 |

生成中响应：

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "video",
  "model": "seedance-2.0-fast-480p",
  "status": "in_progress",
  "progress": 45,
  "created_at": 1783670400
}
```

渠道兼容响应可能直接返回 `video_url`：

```json
{
  "id": "task_xxx",
  "status": "completed",
  "progress": 100,
  "video_url": "https://temporary.example.com/output.mp4"
}
```

完成媒体归档后的 OceanWay 目标格式：

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "video",
  "model": "seedance-2.0-fast-480p",
  "status": "completed",
  "progress": 100,
  "metadata": {
    "url": "https://media.ocean-way.top/videos/video_xxx.mp4"
  }
}
```

为兼容不同版本，建议同时读取以下位置：

```javascript
const videoUrl =
  response.video_url ??
  response.metadata?.url ??
  response.data?.[0]?.url ??
  response.output?.url;
```

### 8.4 轮询策略

建议在创建成功 5 秒后开始查询：

- 前 2 分钟：每 10 秒查询一次。
- 2 至 10 分钟：每 15 秒查询一次。
- 10 至 30 分钟：每 30 秒查询一次。
- 超过 30 分钟：转入后台观察，每 60 至 120 秒查询一次；不要认定失败，也不要重复创建同一任务。

轮询接口本身返回 `200` 不代表视频已经成功。必须检查响应 JSON 中的 `status`。

### 8.5 下载视频和链接保护

视频完成后建议优先使用 OceanWay 下载接口，并继续携带创建任务时使用的 Bearer Key：

```bash
curl -L https://new.ocean-way.top/v1/videos/task_xxx/content \
  -H "Authorization: Bearer sk-你的API密钥" \
  -o output.mp4
```

- 视频 URL 是结果文件地址，不代表模型或渠道身份。
- `/content` 中必须使用创建响应阶段保存的任务 ID。
- 客户端必须把 URL 当作不透明字符串，不得依赖域名、文件名或查询参数。
- 现阶段部分任务可能返回临时签名链接，请在任务完成后及时下载。
- OceanWay 统一媒体域名完成部署后，结果将逐步统一为 `https://media.ocean-way.top/...`，客户端无需修改解析逻辑。
- 业务需要永久保存时，请把视频复制到自己的对象存储。
- 不要公开传播带签名参数的原始链接。

### 8.6 失败与计费

任务的最终 `status` 为 `failed` 时不应按成功任务计费。若出现失败任务扣费、重复扣费或长时间卡在生成中，请提交 `task_id` 申请核查。

内容审核拒绝、参数错误和账号无权限属于明确失败，不应通过反复重试绕过。

## 9. SDK 示例

### 9.1 Python

安装：

```bash
pip install openai
```

Chat Completions：

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OCEANWAY_API_KEY"],
    base_url="https://new.ocean-way.top/v1",
)

response = client.chat.completions.create(
    model="gpt-5.4",
    messages=[{"role": "user", "content": "你好"}],
)

print(response.choices[0].message.content)
```

Responses：

```python
response = client.responses.create(
    model="gpt-5.4",
    input="用三点解释异步任务",
)

print(response.output_text)
```

视频任务使用普通 HTTP 请求：

```python
import os
import time
import requests

base_url = "https://new.ocean-way.top/v1"
headers = {
    "Authorization": f"Bearer {os.environ['OCEANWAY_API_KEY']}",
    "Content-Type": "application/json",
}

created = requests.post(
    f"{base_url}/videos",
    headers=headers,
    json={
        "model": "seedance-2.0-fast-480p",
        "prompt": "雨夜霓虹街道，镜头缓慢推进，无人物，无文字",
        "aspect_ratio": "16:9",
        "duration": 5,
    },
    timeout=60,
)
created.raise_for_status()
task = created.json()
task_id = task.get("task_id") or task["id"]

for _ in range(120):
    result = requests.get(
        f"{base_url}/videos/{task_id}",
        headers=headers,
        timeout=30,
    )
    result.raise_for_status()
    body = result.json()

    if body["status"] == "completed":
        print(body.get("metadata", {}).get("url"))
        break
    if body["status"] == "failed":
        raise RuntimeError(body.get("error", "video generation failed"))

    time.sleep(15)
else:
    raise TimeoutError(f"video task still running: {task_id}")
```

### 9.2 Node.js

安装：

```bash
npm install openai
```

示例：

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.OCEANWAY_API_KEY,
  baseURL: "https://new.ocean-way.top/v1",
});

const response = await client.chat.completions.create({
  model: "gpt-5.4",
  messages: [{ role: "user", content: "你好" }],
});

console.log(response.choices[0].message.content);
```

## 10. 错误处理

推荐兼容以下两种错误格式。

标准格式：

```json
{
  "error": {
    "code": "model_not_available",
    "message": "当前模型暂时不可用，请稍后重试",
    "type": "oceanway_api_error",
    "request_id": "req_xxx"
  }
}
```

兼容格式：

```json
{
  "code": "invalid_request",
  "message": "prompt is required",
  "data": null
}
```

常见状态码：

| HTTP 状态码 | 含义 | 建议处理 |
| ---: | --- | --- |
| `400` | 参数错误或格式不支持 | 检查请求体和 Content-Type |
| `401` | 密钥缺失、错误或失效 | 检查 Authorization 请求头 |
| `403` | 用户分组或模型权限不足 | 联系管理员开通对应能力 |
| `404` | 模型、接口或任务不存在 | 检查模型名、路径和 task_id |
| `409` | 任务状态冲突 | 查询现有任务，不要重复创建 |
| `429` | 额度、并发或频率受限 | 降低并发并指数退避 |
| `500` | OceanWay 内部错误 | 保留 request_id 后联系管理员 |
| `502` | 上游响应异常 | 稍后重试 |
| `503` | 当前没有可用能力池 | 稍后重试或切换公开模型 |
| `504` | 上游处理超时 | 查询任务状态，避免直接重复扣费 |

常见错误码：

| 错误码 | 说明 |
| --- | --- |
| `invalid_request` | 请求参数错误 |
| `invalid_api_key` | API 密钥无效 |
| `insufficient_quota` | 余额或额度不足 |
| `model_not_found` | 模型不存在或当前密钥不可见 |
| `model_not_available` | 模型存在但暂时没有可用渠道 |
| `permission_denied` | 用户分组无权使用该能力 |
| `rate_limit_exceeded` | 请求频率或并发超限 |
| `unsupported_parameter` | 当前模型不支持传入的参数 |
| `parameter_conflict` | 素材字段互斥，例如首尾帧与多模态同时使用 |
| `reference_media_invalid` | 参考素材格式、大小、时长或分辨率不合规 |
| `content_policy_violation` | 内容审核未通过 |
| `PROMPT_BLOCKED` | 提示词被内容策略拒绝 |
| `GENERATION_FAILED` | 媒体生成失败或被内容策略拦截 |
| `TIMEOUT` | 生成任务超时 |
| `NO_ACCOUNT` | 当前能力池繁忙或没有可用账号 |
| `upstream_timeout` | 模型服务处理超时 |
| `task_failed` | 异步任务最终失败 |

下游程序应根据 HTTP 状态码和错误码处理，不要依赖错误消息的中文文本做逻辑判断。

## 11. 计费与额度

- 文本模型通常按输入、缓存输入和输出 token 计费。
- 图片模型通常按模型、尺寸、质量和生成数量计费。
- 视频模型通常按模型档位、清晰度和任务次数计费。
- 控制台展示的余额和调用日志为账单核对依据。
- 生成接口返回 HTTP `200` 不一定代表最终成功；视频必须以最终任务状态为准。
- 因网络中断而无法确认结果时，请先查询日志或任务状态，再决定是否重试。

## 12. 安全与合规

- 禁止生成违法、色情、极端暴力、政治敏感、侵犯隐私或明显侵权的内容。
- 不要上传无权处理的个人照片、商业机密或受保护数据。
- API 密钥应仅保存在服务端环境变量或密钥管理系统中。
- 建议为不同业务创建不同密钥，并分别设置额度和模型权限。
- 发现异常调用时应立即停用密钥并联系管理员。

## 13. 上线检查清单

- 已使用 `/v1/models` 验证目标模型对当前密钥可见。
- 已设置正确的 Base URL，且没有重复拼接 `/v1/v1`。
- 已将 API 密钥放在服务端环境变量中。
- 已分别设置文本、图片和视频请求超时。
- 已支持解析标准和兼容错误格式。
- 已区分文本重试和视频创建重试；视频 `504` 不会被直接重复创建。
- 已避免对同步图片请求和视频任务进行无条件自动重试。
- 已优先保存视频创建响应的 `id`，并正确轮询最终状态。
- 已分别测试文生视频、单图生视频、多参考图、视频/音频参考和首尾帧。
- 已验证首尾帧与多模态同时出现时会返回参数冲突错误。
- 视频请求没有发送当前协议不支持的 `negative_prompt` 字段。
- 已在提交前校验参考素材的数量、格式、大小、时长和分辨率。
- 已把媒体 URL 当作不透明临时地址处理。
- 已记录请求时间、模型名和 request_id，但未记录完整密钥。

## 14. 技术支持

提交问题时，请提供：

```text
请求时间：
接口路径：
模型名称：
HTTP 状态码：
request_id/task_id：
密钥末尾 4 位：
错误响应 JSON：
```

请勿提供完整 API 密钥、Authorization 请求头或带有长期有效签名的媒体链接。
