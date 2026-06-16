import { apiClient } from './client'

export type ImageWorkbenchModel = 'gpt-image-2' | 'gpt-image-1.5' | 'gpt-image-1' | 'gpt-image-1-mini'
export type ImageWorkbenchSize = '1024x1024' | '1024x1536' | '1536x1024' | 'auto'
export type ImageWorkbenchQuality = 'auto' | 'low' | 'medium' | 'high'
export type ImageWorkbenchFormat = 'png' | 'jpeg' | 'webp'
export type ImageWorkbenchBackground = 'auto' | 'transparent' | 'opaque'
export type ImageWorkbenchModeration = 'auto' | 'low'

export interface GenerateImagesRequest {
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

const IMAGE_WORKBENCH_TIMEOUT_MS = 180_000

export class ImageWorkbenchTimeoutError extends Error {
  constructor() {
    super('Image generation timed out')
    this.name = 'ImageWorkbenchTimeoutError'
  }
}

function createTimedSignal(signal?: AbortSignal) {
  const controller = new AbortController()
  let timedOut = false

  const timeoutId = window.setTimeout(() => {
    timedOut = true
    controller.abort()
  }, IMAGE_WORKBENCH_TIMEOUT_MS)

  const abortFromParent = () => controller.abort()
  if (signal?.aborted) {
    controller.abort()
  } else {
    signal?.addEventListener('abort', abortFromParent, { once: true })
  }

  return {
    signal: controller.signal,
    timedOut: () => timedOut,
    cleanup: () => {
      window.clearTimeout(timeoutId)
      signal?.removeEventListener('abort', abortFromParent)
    },
  }
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

  const timedSignal = createTimedSignal(signal)

  let response: { data: GenerateImagesResponse }
  try {
    response = await apiClient.post<GenerateImagesResponse>('/user/image-workbench/images/generations', payload, {
      signal: timedSignal.signal,
      timeout: IMAGE_WORKBENCH_TIMEOUT_MS + 5_000,
    })
  } catch (error) {
    if (timedSignal.timedOut()) {
      throw new ImageWorkbenchTimeoutError()
    }
    throw error
  } finally {
    timedSignal.cleanup()
  }

  const result = response.data
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
