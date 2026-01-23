import { ElMessage } from 'element-plus'

export const TRACE_COPY_HOTKEY = 'Ctrl+Shift+C'

let lastTraceId = ''
let lastTraceTime = ''
let hotkeyInitialized = false

const formatTime = (value: Date) => value.toLocaleString('zh-CN')

const resolveHeaderTraceId = (headers: any): string => {
  if (!headers) return ''
  if (typeof headers.get === 'function') {
    const headerValue = headers.get('x-trace-id') || headers.get('x-request-id')
    if (headerValue) return String(headerValue)
  }
  const direct =
    headers['x-trace-id'] ||
    headers['X-Trace-Id'] ||
    headers['x-request-id'] ||
    headers['X-Request-Id']
  if (direct) return String(direct)
  return ''
}

export const resolveTraceIdFromError = (error: any): string => {
  const traceFromData = error?.response?.data?.trace_id
  if (traceFromData) return String(traceFromData)

  const traceFromResponseHeaders = resolveHeaderTraceId(error?.response?.headers)
  if (traceFromResponseHeaders) return traceFromResponseHeaders

  const traceFromConfigHeaders = resolveHeaderTraceId(error?.config?.headers)
  if (traceFromConfigHeaders) return traceFromConfigHeaders

  const traceFromConfig = error?.config?._traceId
  if (traceFromConfig) return String(traceFromConfig)

  return ''
}

export const recordTraceInfo = (traceId: string, at: Date) => {
  if (!traceId) return
  lastTraceId = traceId
  lastTraceTime = formatTime(at)
}

export const buildTraceMessage = (message: string, error?: any) => {
  const now = new Date()
  const traceId = resolveTraceIdFromError(error)
  const timeText = formatTime(now)
  if (traceId) {
    recordTraceInfo(traceId, now)
  }
  const traceText = traceId || 'unknown'
  return `${message} | TraceID: ${traceText} | 时间: ${timeText} | 快捷键: ${TRACE_COPY_HOTKEY}`
}

export const initTraceHotkey = () => {
  if (hotkeyInitialized || typeof window === 'undefined') return
  hotkeyInitialized = true
  window.addEventListener('keydown', handleTraceHotkey)
}

const handleTraceHotkey = (event: KeyboardEvent) => {
  const isCopyHotkey = (event.ctrlKey || event.metaKey) && event.shiftKey && event.code === 'KeyC'
  if (!isCopyHotkey) return
  if (!lastTraceId) {
    ElMessage.warning('暂无可复制的 TraceID')
    return
  }
  event.preventDefault()
  copyText(lastTraceId)
    .then(() => {
      const timeSuffix = lastTraceTime ? ` (${lastTraceTime})` : ''
      ElMessage.success(`TraceID 已复制${timeSuffix}`)
    })
    .catch(() => {
      ElMessage.error('复制 TraceID 失败')
    })
}

const copyText = async (value: string) => {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value)
    return
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  textarea.style.pointerEvents = 'none'
  document.body.appendChild(textarea)
  textarea.focus()
  textarea.select()
  const ok = document.execCommand('copy')
  document.body.removeChild(textarea)
  if (!ok) {
    throw new Error('copy failed')
  }
}
