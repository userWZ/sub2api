export type ImageWorkbenchModel = 'gpt-image-2' | 'gpt-image-1.5' | 'gpt-image-1' | 'gpt-image-1-mini'
export type ImageWorkbenchSize = '1024x1024' | '1024x1536' | '1536x1024' | 'auto'
export type ImageWorkbenchQuality = 'auto' | 'low' | 'medium' | 'high'
export type ImageWorkbenchFormat = 'png' | 'jpeg' | 'webp'
export type ImageWorkbenchBackground = 'auto' | 'transparent' | 'opaque'
export type ImageWorkbenchModeration = 'auto' | 'low'

export interface GenerateImagesRequest {
  apiKey: string
  model: ImageWorkbenchModel
  prompt: string
  size: ImageWorkbenchSize
  quality: ImageWorkbenchQuality
  output_format: ImageWorkbenchFormat
  background: ImageWorkbenchBackground
  moderation: ImageWorkbenchModeration
  n: number
}

export interface GeneratedImage {
  url?: string
  b64_json?: string
  revised_prompt?: string
}

export interface GenerateImagesResponse {
  created?: number
  data: GeneratedImage[]
  usage?: unknown
}

interface OpenAIErrorBody {
  error?: {
    message?: string
    type?: string
    code?: string
  }
  message?: string
}

export async function generateImages(req: GenerateImagesRequest, signal?: AbortSignal): Promise<GenerateImagesResponse> {
  const payload: Record<string, unknown> = {
    model: req.model,
    prompt: req.prompt,
    n: req.n,
  }

  if (req.size !== 'auto') payload.size = req.size
  if (req.quality !== 'auto') payload.quality = req.quality
  if (req.output_format) payload.output_format = req.output_format
  if (req.background !== 'auto') payload.background = req.background
  if (req.moderation !== 'auto') payload.moderation = req.moderation

  const response = await fetch('/v1/images/generations', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${req.apiKey}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
    signal,
  })

  const contentType = response.headers.get('content-type') || ''
  const body = contentType.includes('application/json')
    ? await response.json().catch(() => ({}))
    : await response.text().catch(() => '')

  if (!response.ok) {
    const errorBody = body as OpenAIErrorBody
    const message =
      errorBody?.error?.message ||
      errorBody?.message ||
      (typeof body === 'string' && body.trim()) ||
      `Image generation failed with HTTP ${response.status}`
    throw new Error(message)
  }

  const result = body as GenerateImagesResponse
  return {
    ...result,
    data: Array.isArray(result.data) ? result.data : [],
  }
}

export function imageDataUrl(image: GeneratedImage, fallbackFormat: ImageWorkbenchFormat = 'png'): string {
  if (image.url) return image.url
  if (!image.b64_json) return ''
  return `data:image/${fallbackFormat};base64,${image.b64_json}`
}
