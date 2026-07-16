# OceanWay 视频生成 API 对接文档

文档版本：`2.2`

更新日期：`2026-07-11`

适用对象：通过 OceanWay API 接入文生视频、图生视频、多模态参考和首尾帧视频生成能力的开发者。

> 能力说明：不同视频模型支持的素材数量和参数不完全相同。请仅使用本文列出的字段，并以 `/v1/models` 返回结果和账号权限为准。未列出的字段不保证生效。

## 1. 基础信息

API Base URL：

```text
https://new.ocean-way.top/v1
```

鉴权方式：

```http
Authorization: Bearer sk-你的API密钥
```

JSON 请求：

```http
Content-Type: application/json
```

单图文件上传使用 `multipart/form-data`，不需要手动设置 `Content-Type`，由 HTTP 客户端自动生成 boundary。

## 2. 接口概览

| 能力 | 方法 | 路径 |
| --- | --- | --- |
| 创建视频任务 | `POST` | `/v1/videos` |
| 查询任务状态 | `GET` | `/v1/videos/{id}` |
| 下载视频 | `GET` | `/v1/videos/{id}/content` |

视频生成采用异步任务模式：

```text
提交任务 -> 保存创建响应的 id -> 轮询状态 -> completed 后读取或下载视频
```

不要保持一个 HTTP 请求等待视频生成完成，也不要在未确认任务状态前重复创建同一个高成本任务。

## 3. 查询可用模型

```bash
curl https://new.ocean-way.top/v1/models \
  -H "Authorization: Bearer sk-你的API密钥"
```

模型是否对某个 API Key 可见，以该接口实际返回和账号权限为准。

建议优先使用带固定分辨率后缀的模型：

| 档位 | 480p | 720p |
| --- | --- | --- |
| 标准版 | `seedance-2.0-480p` | `seedance-2.0-720p` |
| Mini | `seedance-2.0-mini-480p` | `seedance-2.0-mini-720p` |
| Fast | `seedance-2.0-fast-480p` | `seedance-2.0-fast-720p` |

模型价格和计费单位以账号控制台显示为准。

## 4. 创建视频任务

接口：

```http
POST /v1/videos
```

最小文生视频请求：

```bash
curl https://new.ocean-way.top/v1/videos \
  -H "Authorization: Bearer sk-你的API密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "seedance-2.0-fast-480p",
    "prompt": "雨夜霓虹街道，镜头缓慢推进，电影感光影，无人物，无文字",
    "aspect_ratio": "16:9",
    "duration": 5
  }'
```

### 4.1 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 公开视频模型名 |
| `prompt` | string | 是 | 视频描述，最长 5,000 字符 |
| `duration` | integer | 是 | 视频时长，Seedance 2.0 通常支持 4–15 秒任意整数 |
| `aspect_ratio` | string | 否 | 默认 `16:9` |
| `resolution` | string | 否 | `480p` 或 `720p`；固定分辨率模型通常无需传 |
| `audio` | boolean | 否 | 原生音频开关，仅部分模型支持；是否包含音轨以最终视频文件为准 |
| `image_url` | string | 否 | 主参考图 URL 或图片 Data URI |
| `reference_image_urls` | array | 否 | 额外参考图 URL 数组，也可使用带名称对象 |
| `reference_images` | array | 否 | 命名参考图对象数组 |
| `reference_image_names` | array | 否 | 与 `reference_image_urls` 同序的名称数组 |
| `reference_videos` | array | 否 | 参考视频 HTTPS URL 数组 |
| `reference_audios` | array | 否 | 参考音频 HTTPS URL 数组 |
| `first_image_url` | string | 否 | 首帧，必须与 `last_image_url` 成对 |
| `last_image_url` | string | 否 | 尾帧，必须与 `first_image_url` 成对 |
| `image` | file | 否 | multipart 单张主参考图 |

### 4.2 画幅比例

Seedance 2.0 常用画幅：

```text
16:9
9:16
1:1
21:9
3:4
4:3
```

不要把 `16:9-720p` 等 UI 标签直接传给 `aspect_ratio`。清晰度由模型名或 `resolution` 单独控制。

### 4.3 生成模式

服务端根据素材字段自动判断模式，不需要传 `mode`：

| 模式 | 最少必传 | 触发字段 | 规则 |
| --- | --- | --- | --- |
| 文生视频 | `prompt` | 不传素材字段 | 只使用文本描述 |
| 单图生视频 | `prompt` + 1 张图 | `image_url` 或 multipart `image` | 使用主参考图 |
| 多参考图 | `prompt` + 多张图 | `reference_image_urls` 或 `reference_images` | 可在提示词中引用图片序号或角色名 |
| 全能参考 | `prompt` + 至少 1 张主图 | 图像 + `reference_videos`/`reference_audios` | 视频和音频参考必须搭配主图 |
| 首尾帧 | `prompt` + 首帧 + 尾帧 | `first_image_url` + `last_image_url` | 客户端不得同时传其他参考素材 |

首尾帧与其他参考素材互斥。客户端必须在提交前完成校验，不得组合使用。

## 5. 模型素材限制

当前 Seedance 模型存在两类参数档案，必须按照实际模型名使用：

| 模型类型 | 参考图 | 参考视频 | 参考音频 | `audio` 开关 |
| --- | ---: | ---: | ---: | --- |
| `seedance-2.0`、`seedance-2.0-fast`、`seedance-2.0-mini` | 合计最多 4 张 | 最多 3 条，单条 4–15 秒，总时长 ≤15 秒 | 最多 1 条，≤15 秒 | 支持 |
| 带 `-480p`、`-720p` 等固定分辨率后缀 | 合计最多 9 张 | 最多 3 条，单条 2–15 秒，总时长 ≤15 秒 | 最多 3 条，须搭配主图 | 以模型能力为准 |

### 5.1 参考图要求

- 格式：JPEG、PNG、WebP。
- 单张不超过 30 MB。
- 每边至少 300 px。
- 宽高比建议保持在 0.4–2.5。
- 长边不超过 4,000 px。
- 支持 HTTPS URL 或 `data:image/...;base64,...`。
- multipart 仅建议上传一张主参考图。
- 图片应具有清晰主体和正常动态范围。纯色、无明显主体或低质量图片可能导致任务失败。

### 5.2 参考视频要求

- 使用 HTTPS 公网直链。
- 格式：MP4 或 MOV。
- 帧率：24–60 fps。
- 单条不超过 50 MB。
- 多条参考视频总时长不超过 15 秒。
- 带参考视频时必须同时提供至少一张主参考图。
- URL 必须支持无 Cookie、无登录态的服务端直接 `GET`，并正确返回视频 `Content-Type`。带防盗链、地区限制或复杂跳转的链接可能无法读取。

### 5.3 参考音频要求

- 使用 HTTPS 公网直链。
- 常见格式：MP3、WAV、M4A。
- 单条建议不超过 15 MB。
- 带参考音频时必须同时提供至少一张主参考图。
- URL 必须支持无 Cookie、无登录态的服务端直接下载。

### 5.4 素材引用

序号从 1 开始：

```text
@image1 ... @image9
@video1 ... @video3
@audio1 ... @audio3
```

使用命名参考图时，可以在提示词中直接引用角色名：

```text
@志强 与 @清雅 在医院走廊相遇
```

## 6. 调用示例

### 6.1 文生视频

```json
{
  "model": "seedance-2.0-fast-480p",
  "prompt": "雨夜霓虹街道，镜头缓慢推进，电影感光影",
  "aspect_ratio": "16:9",
  "duration": 8
}
```

### 6.2 单图生视频：HTTPS URL

```json
{
  "model": "seedance-2.0-fast-480p",
  "prompt": "保持人物一致，人物缓慢向前走动",
  "aspect_ratio": "16:9",
  "duration": 5,
  "image_url": "https://cdn.example.com/person.jpg"
}
```

### 6.3 单图生视频：Base64

```json
{
  "model": "seedance-2.0-fast-480p",
  "prompt": "让画面自然动起来",
  "duration": 5,
  "image_url": "data:image/png;base64,iVBORw0KGgo..."
}
```

Base64 会显著增大请求体。较大的图片建议使用 HTTPS URL 或 multipart 上传。

### 6.4 单图生视频：multipart 上传

```bash
curl https://new.ocean-way.top/v1/videos \
  -H "Authorization: Bearer sk-你的API密钥" \
  -F "model=seedance-2.0-fast-480p" \
  -F "prompt=保持主体一致，让画面自然动起来" \
  -F "duration=5" \
  -F "aspect_ratio=16:9" \
  -F "image=@./photo.jpg"
```

多参考图、参考视频、参考音频和命名角色请使用 JSON 请求。

### 6.5 多参考图

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

### 6.6 命名角色参考图

```json
{
  "model": "seedance-2.0-720p",
  "prompt": "@志强 与 @清雅 在医院走廊相遇，镜头缓慢向前推进",
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

`name` 应保持唯一，且提示词中的角色名必须与请求中的名称完全一致。

### 6.7 图、视频和音频全能参考

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

`reference_videos` 或 `reference_audios` 不能单独使用，必须至少提供一张主参考图。

### 6.8 首尾帧视频

```json
{
  "model": "seedance-2.0-fast-720p",
  "prompt": "从清晨城市平滑过渡到夜晚霓虹，保持镜头位置一致",
  "duration": 5,
  "first_image_url": "https://cdn.example.com/start.jpg",
  "last_image_url": "https://cdn.example.com/end.jpg"
}
```

首尾帧必须成对出现，客户端不能同时传入以下字段：

```text
image_url
reference_image_urls
reference_images
reference_videos
reference_audios
```

首尾帧模式的输出画幅可能优先参考输入图片。需要固定画幅时，应先把首尾帧裁剪为相同宽高比。

### 6.9 原生音频

部分模型支持使用 `audio` 请求生成原生音频：

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

`audio` 控制是否请求生成原生音频；`reference_audios` 用于提供参考音频，两者含义不同。最终是否包含音轨以返回的视频文件为准。

### 6.10 暂不支持的字段

当前不对外支持 `negative_prompt`。需要表达排除项时，请把“不要出现字幕、水印、Logo、闪烁、变形”等要求直接写入 `prompt`。

## 7. 创建任务响应

创建响应示例：

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

`object`、`model`、`seconds` 和 `size` 等字段可能为空或不存在，客户端不能依赖这些可选字段。

客户端必须保存创建响应的 `id`，后续查询和下载始终使用该值。

兼容读取方式：

```javascript
const taskId = response.id ?? response.task_id;
```

## 8. 查询任务

接口：

```http
GET /v1/videos/{id}
```

请求示例：

```bash
curl https://new.ocean-way.top/v1/videos/task_xxx \
  -H "Authorization: Bearer sk-你的API密钥"
```

任务状态：

| 状态 | 含义 | 客户端行为 |
| --- | --- | --- |
| `queued` | 等待处理 | 继续轮询 |
| `in_progress` | 正在生成 | 继续轮询 |
| `completed` | 已完成 | 读取视频链接 |
| `failed` | 生成失败 | 停止轮询并记录错误 |

`progress` 仅表示参考进度，不代表线性剩余时间，也不能作为自动重试依据。

生成中：

```json
{
  "id": "task_xxx",
  "status": "in_progress",
  "progress": 50,
  "created_at": 1783670400
}
```

完成响应示例：

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "video",
  "model": "seedance-2.0-mini-480p",
  "status": "completed",
  "progress": 100,
  "metadata": {
    "url": "https://temporary.example.com/output.mp4"
  }
}
```

继续查询和调用 `/content` 时，始终使用创建阶段保存的 `id`。

只有在 `status === "completed"` 后才读取 URL，并校验它确实是 HTTP(S) 地址：

```javascript
let videoUrl = null;

if (response.status === "completed") {
  const candidate =
    response.video_url ??
    response.metadata?.url ??
    response.data?.[0]?.url ??
    response.output?.url;

  if (typeof candidate === "string" && /^https?:\/\//.test(candidate)) {
    videoUrl = candidate;
  }
}
```

## 9. 轮询策略

建议在任务创建成功 5 秒后开始查询：

- 前 2 分钟：每 10 秒查询一次。
- 2–10 分钟：每 15 秒查询一次。
- 10–30 分钟：每 30 秒查询一次。
- 超过 30 分钟：将任务转入后台观察，保留创建响应的 `id`，改为每 60–120 秒查询；不要认定失败，也不要重复创建。

轮询接口返回 HTTP `200` 不代表任务已经成功，必须检查 JSON 中的 `status`。

复杂任务可能需要较长处理时间。超过前台等待时间后，应保留任务 ID 并继续低频轮询，不要自动重复创建。

## 10. 下载视频

建议优先使用 OceanWay 下载接口：

```http
GET /v1/videos/{id}/content
```

该接口返回视频文件内容：

```bash
curl -L https://new.ocean-way.top/v1/videos/task_xxx/content \
  -H "Authorization: Bearer sk-你的API密钥" \
  -o output.mp4
```

媒体处理规则：

- `/content` 中的任务 ID 必须使用创建响应的 `id`。
- 直接结果 URL 可能是临时存储地址；不需要直接链接时不要向最终用户返回它。
- 必须直接使用结果 URL 时，将其视为不透明字符串，不要依赖域名、文件名或查询参数。
- 现阶段部分结果可能是带有效期的临时签名 URL，请在任务完成后及时下载。
- 业务需要永久保存时，应复制到自己的对象存储。
- 不要在日志或公开页面中完整记录带签名参数的视频 URL。
- 大规模或长期交付建议将视频保存到自己的对象存储或 CDN。

## 11. 错误处理

任务失败示例：

```json
{
  "id": "task_xxx",
  "status": "failed",
  "error": {
    "code": "",
    "message": "生成失败，请更换参考素材或调整提示词后重试"
  },
  "metadata": {
    "url": "生成失败，请更换参考素材或调整提示词后重试"
  }
}
```

失败响应中的 `metadata.url` 可能错误地包含错误文本，因此必须先判断 `status`，不能直接读取该字段作为链接。

同步参数错误目前存在两种格式。缺少提示词通常是普通 JSON：

```json
{
  "code": "invalid_request",
  "message": "prompt is required",
  "data": null
}
```

部分参数错误会返回 `fail_to_fetch_task`，且详细原因可能包含在 `message` 中。客户端应展示原始消息并记录 request ID，不要依赖消息文本编写业务逻辑。

常见 HTTP 状态码：

| 状态码 | 含义 | 建议处理 |
| ---: | --- | --- |
| `400` | 参数错误、素材不合规或服务拒绝 | 修正请求，不要盲目重试 |
| `401` | API Key 无效 | 检查鉴权请求头 |
| `403` | 模型或用户分组无权限 | 联系 OceanWay 技术支持开通权限 |
| `404` | 模型或任务不存在 | 检查模型名和创建响应的 id |
| `429` | 并发、频率或额度受限 | 降低并发并退避重试 |
| `500` | 服务错误 | 保留 request_id 和创建响应的 id |
| `502` | 模型服务响应异常 | 稍后重试 |
| `503` | 当前没有可用能力池 | 稍后重试 |
| `504` | 处理超时 | 先查询任务，不要立即重复创建 |

常见错误码：

| 错误码 | 说明 |
| --- | --- |
| `invalid_request` | 请求参数错误 |
| `fail_to_fetch_task` | 任务创建失败，详细原因可能包含在 message 中 |
| `PROMPT_BLOCKED` | 提示词被内容策略拒绝 |
| `GENERATION_FAILED` | 生成失败或内容策略拦截 |
| `TIMEOUT` | 生成任务超时 |
| `NO_ACCOUNT` | 当前服务繁忙或没有可用账号 |
| `model_not_available` | 当前没有可用模型能力池 |
| `insufficient_quota` | 余额或额度不足 |

客户端应根据状态码和错误码处理，不要依赖错误消息的中文或英文文本。

只发送本文列出的字段，并在客户端执行素材数量、格式和互斥规则校验。

## 12. 超时与重试

建议配置：

```text
创建任务连接超时：10 秒
创建任务读取超时：60 秒
单次查询读取超时：30 秒
前台任务等待时间：最长 30 分钟
后台任务观察时间：建议至少 60 分钟
```

重试规则：

- 创建任务明确返回 `429`、`502`、`503` 时，可以使用 2、4、8、16 秒指数退避。
- 最多自动重试 2 次。
- 已获得创建响应的 `id` 后不要重新创建，只轮询已有任务。
- 创建请求超时且无法确认是否已创建成功时，不要立刻重发，以免重复扣费。
- `400`、`401`、`403`、`404` 不应自动重试。
- 流程中断时保留创建响应的 `id`，便于恢复查询。

## 13. Python 完整示例

```python
import os
import time
import requests

BASE_URL = "https://new.ocean-way.top/v1"
API_KEY = os.environ["OCEANWAY_API_KEY"]

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json",
}

created = requests.post(
    f"{BASE_URL}/videos",
    headers=headers,
    json={
        "model": "seedance-2.0-fast-480p",
        "prompt": "雨夜霓虹街道，镜头缓慢推进，无人物，无文字",
        "aspect_ratio": "16:9",
        "duration": 5,
    },
    timeout=(10, 60),
)
created.raise_for_status()

task = created.json()
task_id = task.get("id") or task["task_id"]

for _ in range(240):
    response = requests.get(
        f"{BASE_URL}/videos/{task_id}",
        headers={"Authorization": f"Bearer {API_KEY}"},
        timeout=(10, 30),
    )
    response.raise_for_status()
    result = response.json()
    status = result.get("status")

    if status == "completed":
        direct_url = (
            result.get("video_url")
            or result.get("metadata", {}).get("url")
            or (result.get("data") or [{}])[0].get("url")
        )
        if not isinstance(direct_url, str) or not direct_url.startswith(("http://", "https://")):
            raise RuntimeError("completed task did not return a valid video URL")

        # Prefer this OceanWay URL when delivering the file to your application.
        print(f"{BASE_URL}/videos/{task_id}/content")
        break

    if status == "failed":
        error = result.get("error") or "video generation failed"
        raise RuntimeError(str(error))

    time.sleep(15)
else:
    print(f"task is still running; keep it for background polling: {task_id}")
```

## 14. 计费说明

- 视频费用由模型、清晰度和生成时长共同决定，具体价格以账号控制台为准。
- 创建任务返回 HTTP `200` 只表示任务已受理，最终结果以任务状态为准。
- 任务创建后可能暂时冻结或预扣额度，最终账单可能存在短暂更新延迟。
- 网络中断时应先查询任务状态和消费记录，再决定是否重新创建。
- 发现重复扣费、账单异常或任务长时间没有终态时，请提交创建响应的 `id` 核查。

## 15. 安全与合规

- API Key 只能保存在服务端环境变量或密钥管理系统中。
- 禁止在浏览器前端、公开仓库、截图或日志中暴露完整 Key。
- 不要上传无权处理的个人影像、商业机密或受保护素材。
- 避免真人正脸、版权 IP、色情、暴力、政治敏感和其他违规内容。
- 参考素材 URL 不应包含长期凭证；建议使用短期签名 HTTPS URL。
- 任务报错时只提供 Key 末尾 4 位，不要发送完整 Key。

## 16. 上线检查清单

- 已通过 `/v1/models` 确认模型对当前 API Key 可见。
- Base URL 为 `https://new.ocean-way.top/v1`，没有重复拼接 `/v1/v1`。
- 已在控制台确认模型价格，并核算目标时长的预期费用。
- 优先使用带 `-480p` 或 `-720p` 后缀的模型。
- 已确认业务所需的生成模式和素材类型。
- 使用首尾帧时，客户端已保证两个字段成对出现，且没有同时传其他参考素材。
- 已按照具体模型校验参考素材数量、格式、大小、时长、分辨率和帧率。
- 参考视频和音频 URL 支持无需 Cookie 的公网直接下载。
- 没有传入当前暂不支持的 `negative_prompt`。
- 已保存创建响应的 `id`，后续查询没有使用其他兼容字段覆盖该值。
- 已正确处理 `queued`、`in_progress`、`completed` 和 `failed`。
- 已先判断 `status`，再读取并校验媒体 URL。
- 已兼容 `video_url`、`metadata.url` 和 `data[0].url`。
- 已设置 30 分钟前台等待和至少 60 分钟后台观察策略。
- 没有对高成本任务进行无条件自动重试。
- 对外交付优先使用 `/v1/videos/{id}/content`，避免依赖临时结果地址。

## 17. 技术支持信息

提交问题时请提供：

```text
请求时间和时区：
模型名称：
生成模式：文生/单图/多参考/首尾帧
HTTP 状态码：
创建响应 id/request_id：
API Key 末尾 4 位：
错误响应 JSON：
```

请勿提供完整 API Key、Authorization 请求头或包含长期有效凭证的素材 URL。
