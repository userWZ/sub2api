# OceanWay 统一网关内部规范

文档版本：`1.0`

更新日期：`2026-07-10`

本文档供 OceanWay 内部实现和运维使用，不应发送给下游用户。

## 1. 目标

OceanWay 对外提供稳定的模型名、请求格式、任务状态和错误结构，内部允许接入 OpenAI、Claude、Gemini、GLM、Grok、图片和视频供应商，并按能力、健康状态、成本和优先级自动切换渠道。

对外接口不应暴露：

- 上游供应商名称和域名。
- new-api 渠道 ID、账号 ID、渠道分组名。
- 上游任务 ID。
- 上游原始错误响应。
- 上游媒体存储地址和签名参数。
- 上游 API Key 或其可识别片段。

### 1.1 当前生产状态

截至 `2026-07-10`，`new-api-oceanway` 已经承载用户、Key、额度、渠道和计费，并且线上存在 `/v1/models`、`/v1/responses`、`/v1/images/generations`、`/v1/images/edits`、`POST /v1/videos` 和 `GET /v1/videos/{task_id}` 的实际调用。

当前尚未完成的统一网关能力：

- 视频任务仍可能向下游暴露上游任务 ID。
- 图片和视频结果仍可能返回上游临时签名 URL。
- 渠道能力主要依赖现有渠道配置和分组，尚未形成独立的实测能力注册表。
- 高成本媒体请求的统一幂等、归档和计费状态机尚未完整落地。
- 跨不同供应商私有视频格式的转换仍依赖单独适配。
- 渠道 88 已公开多参考图、参考视频、参考音频、命名参考和首尾帧字段，但尚未对 `new-api-oceanway -> 渠道 88` 的每一种组合完成端到端验收。

因此本文后续内容是目标实现规范，不代表所有能力已经上线。

## 2. 系统边界

```text
下游用户
  -> new.ocean-way.top
  -> OceanWay Gateway
       -> 鉴权、额度、计费：new-api-oceanway
       -> 能力路由、格式转换、任务映射：OceanWay Gateway
       -> 文本/图片/视频上游渠道
       -> 媒体归档：OceanWay Object Storage
  -> media.ocean-way.top
```

职责划分：

| 组件 | 职责 |
| --- | --- |
| `new-api-oceanway` | 用户、API Key、分组权限、额度、模型价格、消费日志 |
| `OceanWay Gateway` | 统一协议、能力判断、渠道调度、格式转换、任务状态机、错误清洗 |
| PostgreSQL | OceanWay 任务、渠道能力、幂等记录、媒体映射 |
| Redis | 限流、短期任务状态、渠道熔断、分布式锁 |
| 对象存储 | 图片和视频成品文件 |
| `media.ocean-way.top` | OceanWay 自有媒体访问域名 |

原有 `sub2api` 继续负责现有零售用户的订阅、积分和账号池逻辑，不把它的业务分组直接映射为统一网关的渠道能力分组。

## 3. 统一模型与能力

内部不要只根据模型名称或供应商套餐名判断能力。每个渠道账号必须维护经过实际测试的能力标签：

```json
{
  "channel_id": 72,
  "capabilities": [
    "chat",
    "responses",
    "image_generation_endpoint"
  ],
  "verified_at": "2026-07-10T00:00:00Z",
  "health": "healthy"
}
```

建议能力标签：

| 标签 | 含义 |
| --- | --- |
| `chat` | 支持 `/v1/chat/completions` |
| `responses` | 支持 `/v1/responses` |
| `function_tools` | 支持普通函数工具 |
| `image_generation_tool` | 支持 Responses 内置 `image_generation` |
| `image_generation_endpoint` | 支持 `/v1/images/generations` |
| `image_edit` | 支持 `/v1/images/edits` |
| `video_text` | 支持文生视频 |
| `video_image` | 支持图生视频 |
| `video_multi_image` | 支持多参考图 |
| `video_named_reference` | 支持命名角色参考图 |
| `video_reference_video` | 支持参考视频 |
| `video_reference_audio` | 支持参考音频 |
| `video_native_audio` | 支持生成原生音频开关 |
| `video_negative_prompt` | 支持独立负面提示词字段 |
| `video_first_last_frame` | 支持首尾帧视频 |
| `video_multimodal` | 支持多模态参考输入 |

`pro-kedaya` 等账号只有在真实请求验证成功后才能标记 `image_generation_tool`。套餐名称、Pro/Plus 标签或模型列表中出现图片模型，都不能代替能力验证。

截至 `2026-07-10`，渠道 88 当前公开的 Seedance 2.0 `api_doc.params` 包含参考图、参考视频、参考音频、命名参考、首尾帧和部分模型的 `audio`，但没有列出 `negative_prompt`。在实际探测成功前，不得给该渠道标记 `video_negative_prompt`。

## 4. 路由决策

路由顺序：

1. 校验用户、Key、余额、分组和公开模型权限。
2. 解析请求路径、模型和工具，计算必须具备的能力标签。
3. 排除禁用、熔断、并发已满和能力未验证的渠道。
4. 按公开模型映射、优先级、健康分、成本和近期延迟排序。
5. 选择渠道并记录内部路由决策。
6. 在允许自动重试的错误类型下切换到下一个渠道。
7. 对下游只返回统一模型名和 OceanWay request_id。

Responses 生图工具判断：

```text
request.tools 包含 type=image_generation
  -> required capability = image_generation_tool
  -> 只进入已验证生图工具池

request.tools 不包含 type=image_generation
  -> 不注入工具
  -> 按普通 responses/text 池调度
```

图片接口判断：

```text
POST /v1/images/generations
  -> required capability = image_generation_endpoint

POST /v1/images/edits
  -> required capability = image_edit
```

视频接口必须根据每一个输入字段计算能力要求，不能把所有 Seedance 渠道视为同一能力：

```text
无素材字段
  -> video_text

image_url 或 multipart image
  -> video_image

reference_image_urls 数量 > 1
  -> video_image + video_multi_image

reference_images 包含 name
  -> video_image + video_multi_image + video_named_reference

reference_videos 非空
  -> video_multimodal + video_reference_video

reference_audios 非空
  -> video_multimodal + video_reference_audio

audio 已显式指定
  -> video_native_audio

negative_prompt 非空
  -> video_negative_prompt

first_image_url 或 last_image_url
  -> video_first_last_frame
```

任何已传入但目标渠道不支持的参数都必须返回 `unsupported_parameter`，不得静默删除后继续计费。

### 4.1 视频统一请求对象

入口层接受对外字段后，应先转换为内部标准对象，再由渠道适配器生成上游请求：

```json
{
  "model": "seedance-2.0-720p",
  "prompt": "...",
  "negative_prompt": null,
  "duration_seconds": 10,
  "aspect_ratio": "16:9",
  "resolution": "720p",
  "generate_audio": null,
  "references": {
    "images": [
      {"url": "https://...", "name": "志强"}
    ],
    "videos": [
      {"url": "https://..."}
    ],
    "audios": [
      {"url": "https://..."}
    ],
    "first_frame": null,
    "last_frame": null
  }
}
```

适配器负责把内部字段转换为渠道 88 的 `image_url`、`reference_image_urls`、`reference_images`、`reference_videos`、`reference_audios`、`first_image_url` 和 `last_image_url`，或其他供应商对应字段。

### 4.2 视频参数校验

统一网关必须在选渠道和扣费前完成校验：

- `first_image_url` 与 `last_image_url` 必须成对出现。
- 首尾帧模式与普通参考图、参考视频和参考音频互斥。
- 参考视频和参考音频必须同时存在至少一张主参考图。
- 素材数量、文件大小、时长、分辨率和格式必须按候选渠道能力取交集。
- `reference_images[].name` 必须唯一，且不能包含控制字符。
- 提示词中的素材引用不能超过实际素材数量。
- `negative_prompt` 只能发给带 `video_negative_prompt` 标签的渠道。
- 不支持的字段返回 `400 unsupported_parameter`，不能忽略。

## 5. 渠道选择评分

可以使用以下基础评分：

```text
score =
  priority_weight
  + health_weight
  + success_rate_weight
  - latency_penalty
  - recent_error_penalty
  - concurrency_penalty
  - cost_penalty
```

硬条件必须先过滤，评分不能让不具备所需能力的低价渠道获得请求。

推荐的渠道状态：

```text
healthy
degraded
cooldown
disabled
```

熔断建议：

- 连续 3 次确定性渠道错误：进入 `cooldown` 60 秒。
- 5 分钟成功率低于 70%：进入 `degraded`，降低权重。
- 鉴权失败、余额不足、账号封禁：立即 `disabled`，等待人工处理。
- 内容审核拒绝属于请求级错误，不计入渠道故障率。

## 6. 自动重试和故障转移

可以切换渠道重试：

- 上游连接失败。
- DNS、TLS、连接重置。
- 上游 `429`、`502`、`503`、`504`。
- 返回空响应或无法解析的响应，且尚未确认任务已创建。

不应自动切换重试：

- 用户参数错误。
- 内容审核拒绝。
- 用户无模型权限。
- 已经拿到上游视频任务 ID，或同步图片请求已经完成上游派发。
- 无法判断上游是否已经成功创建高成本任务。

文本非流式请求可以最多尝试 2 个渠道。流式请求一旦向用户发送首个有效事件，不允许切换渠道重新生成。

## 7. 请求和任务标识

每个入口请求生成 OceanWay request_id：

```text
req_ow_<ulid>
```

当前异步视频任务生成 OceanWay task_id：

```text
video_ow_<ulid>
```

同步图片请求只生成 `request_id`，不生成公开任务 ID。未来可选图片异步模式可以使用 `image_ow_<ulid>`，但在通用持久化 Worker、R2 归档、幂等和失败结算全部上线前，不得注册图片任务查询路由，也不得写入外部文档。

内部映射表示例：

```sql
oceanway_task_id
request_id
user_id
token_id
public_model
internal_channel_id
upstream_task_id
upstream_status
public_status
upstream_result_url_encrypted
oceanway_media_id
billing_status
created_at
updated_at
completed_at
```

下游查询 `/v1/videos/{task_id}` 时只能使用 OceanWay task_id。网关在内部转换为上游任务 ID。

## 8. 视频状态机

统一状态：

```text
queued -> in_progress -> archiving -> completed
                      -> failed
                      -> expired
```

`archiving` 是内部状态。对外可以继续显示 `in_progress`，直到媒体文件已成功复制到 OceanWay 对象存储。

只有满足以下条件才能标记 `completed`：

1. 上游任务成功。
2. 已取得有效媒体 URL。
3. 已完成文件下载和校验。
4. 已上传对象存储。
5. 已生成 OceanWay 媒体 URL。
6. 已完成成功计费或计费记录已进入可靠队列。

### 8.1 可选图片异步模式上线门槛

图片保持同步默认。未来只有显式传入 `async: true` 的请求可以进入可选异步模式，并且必须同时满足：

1. 图片任务和原始请求已持久化，进程重启后可以重新领取。
2. API Key 不以明文写入任务；Worker 使用内部用户、Token 和计费上下文执行。
3. 创建请求具备幂等键，派发结果不确定时不会自动重复生成。
4. 预扣、成功结算、失败退款和重复回调均具备数据库幂等保护。
5. 所有结果已转存 R2，并以与同步图片相同的 `data` 数组返回。
6. 已实现按用户鉴权的任务查询，并阻止跨用户读取。

建议的未来兼容路径为 `POST /v1/images/generations` 携带 `async: true`，并通过 `GET /v1/images/generations/{task_id}` 查询。该路径当前不得对外启用。

## 9. 媒体归档与链接改写

### 9.1 输入参考素材

统一网关要先标准化输入素材，再交给渠道适配器：

- 图片允许 HTTPS URL、图片 Data URI 和 multipart 文件。
- 参考视频和参考音频对外只接受 HTTPS URL；需要支持本地上传时，先上传至 OceanWay 临时对象存储，再转换为短期签名 HTTPS URL。
- Data URI 和 multipart 图片应先落到临时对象存储，避免把超大 Base64 重复写入队列、数据库和日志。
- 下载用户提供的 URL 前必须阻止私网、环回、链路本地、云元数据和保留地址，DNS 解析和每次重定向后都要重新校验，防止 SSRF。
- 校验 Content-Type、文件头、大小、图片尺寸、音视频时长、分辨率和帧率后再创建上游任务。
- 输入临时文件使用与任务绑定的不可预测对象 Key，并设置 24 小时以内生命周期。
- 向上游发送的素材 URL 使用短期签名，不暴露用户原始 URL 中的查询参数和凭证。

### 9.2 输出媒体归档

推荐流程：

```text
上游返回视频 URL
  -> 校验域名是否在渠道允许列表
  -> 服务端下载，禁止任意 URL 访问
  -> 校验 Content-Type、Content-Length 和文件头
  -> 计算 SHA-256
  -> 上传对象存储
  -> 保存内部映射
  -> 返回 media.ocean-way.top 地址
```

禁止直接用 `302` 把用户重定向到上游 URL，因为最终访问仍会暴露上游域名。

短期过渡可以由 `/v1/videos/{task_id}/content` 在服务端流式代理文件，但该方案会消耗 `la-vps2` 带宽，只适合作为临时兜底。

正式方案使用 Cloudflare R2、腾讯云 COS、阿里云 OSS 或兼容 S3 的对象存储，并绑定：

```text
https://media.ocean-way.top
```

对象 Key 不应包含上游渠道或任务标识：

```text
videos/2026/07/<oceanway_media_id>.mp4
images/2026/07/<oceanway_media_id>.png
```

建议安全限制：

- 下载来源采用渠道级域名允许列表，防止 SSRF。
- 限制重定向次数，重定向后的域名也必须校验。
- 限制最大图片和视频尺寸。
- 校验 MIME 类型与文件头，不信任 URL 后缀。
- 上游原始 URL 加密存储，日志中只记录哈希或域名分类。
- 对外链接使用短期签名 URL 或带授权的下载接口。
- 对象存储设置 7 至 30 天生命周期，永久保存作为单独收费能力。

HLS 结果需要下载 `.m3u8` 和全部分片，改写播放列表后再发布。未完成 HLS 归档前，不要把上游播放列表直接返回给下游。

## 10. 错误清洗

内部保留完整上游错误，外部只返回统一错误：

```json
{
  "error": {
    "code": "model_not_available",
    "message": "当前模型暂时不可用，请稍后重试",
    "type": "oceanway_api_error",
    "request_id": "req_ow_xxx"
  }
}
```

错误清洗必须移除：

- 上游域名和 URL。
- 渠道 ID 和渠道名称。
- 上游分组名。
- API Key、Cookie、Authorization 和签名参数。
- 上游账号邮箱、手机号或用户 ID。
- 数据库和容器内部地址。

内部日志通过 request_id 关联完整错误、渠道和账号。

## 11. 分组和权限

不要把 `sub2api` 的订阅定价分组原样复制为能力分组。

建议将权限拆成能力：

```text
text
responses
pro-text
pro-image-tool
image-generation
image-edit
video-fast
video-mini
video-standard
claude
gemini
glm
grok
```

用户套餐决定可使用的能力集合；倍率和具体售价由模型价格或套餐规则决定。渠道分组只决定哪些内部渠道可以承接某种能力。

## 12. 计费状态

推荐计费状态：

```text
not_started
reserved
charged
refunded
failed
```

文本请求按实际 usage 结算。图片和视频建议先预留额度，成功后确认扣费，最终失败则释放或退款。

视频计费必须绑定 OceanWay task_id，避免轮询、回调或重复状态更新造成重复扣费。数据库中应有唯一约束：

```text
unique(user_id, oceanway_task_id, billing_event_type)
```

## 13. 可观测性

每次请求至少记录：

```text
request_id
oceanway_task_id
user_id
token_id
public_model
required_capabilities
selected_channel_id
attempt_count
status_code
first_token_latency_ms
total_latency_ms
upstream_error_class
billing_status
```

管理端应能够按 request_id 或 task_id 查看完整调用链，但下游只能看到清洗后的公开信息。

建议监控：

- 各公开模型 5 分钟成功率。
- 各渠道首字延迟和总延迟。
- 图片和视频 P50、P95、P99 生成时间。
- 进行中超过 10、20、30 分钟的任务数。
- 媒体归档失败率和对象存储上传耗时。
- 上游 `401`、`403`、`429`、`5xx` 比例。
- 重复扣费和退款队列积压。

## 14. 部署顺序

1. 先实现 OceanWay request_id、统一错误和字段级渠道能力标签。
2. 实现视频统一请求对象、参考素材校验和各供应商适配器。
3. 修正 Responses `image_generation` 的能力路由，移除未经验证的生图账号。
4. 实现视频 task_id 映射，停止向下游暴露上游任务 ID。
5. 接入 R2 和 `media.ocean-way.top`，完成输入临时素材及输出成片归档。
6. 增加通用媒体 Worker、渠道健康检查、熔断和自动故障转移。
7. 实现高成本任务幂等和计费状态机。
8. 在上述能力验收后开放可选图片异步模式，图片同步仍为默认。
9. 最后扩展 HLS 归档、任务回调和永久媒体保存能力。

## 15. 上线验收

- 相同公开模型可以在两个内部渠道间切换，外部响应格式不变。
- 外部响应和错误中不存在渠道名、渠道 ID、上游域名和上游任务 ID。
- 带 `image_generation` 工具的请求不会进入不支持该工具的账号池。
- 不带 `image_generation` 工具的请求不会被强制注入生图能力。
- 文生、单图、多图、命名参考、视频参考、音频参考和首尾帧均有独立的路由验收用例。
- 首尾帧与多模态素材同时传入时，在扣费前返回 `parameter_conflict`。
- 带 `negative_prompt` 的请求只会进入已验证支持该字段的渠道；没有可用渠道时明确报错。
- 参考素材超出目标渠道数量、时长、大小或分辨率限制时，在创建上游任务前拒绝。
- 视频完成前不会返回上游媒体链接。
- 视频完成后只返回 `media.ocean-way.top` 或 OceanWay 授权下载地址。
- 视频最终失败不扣成功费用，重复回调不会重复扣费。
- 渠道故障时可以自动熔断，并通过 request_id 定位全部尝试记录。
- `la-vps2` 不承担长期大文件下载带宽，对象存储/CDN 正常命中。
