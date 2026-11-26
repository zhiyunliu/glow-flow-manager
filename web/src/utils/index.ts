const ID_ALPHABET = 'abcdefghijklmnopqrstuvwxyz0123456789'

/** 节点编号：n + 11 位随机小写字母数字，与后端生成规则一致 */
export function newNodeId(): string {
  const bytes = new Uint8Array(11)
  crypto.getRandomValues(bytes)
  let id = 'n'
  for (const b of bytes) id += ID_ALPHABET[b % ID_ALPHABET.length]
  return id
}

export function deepClone<T>(val: T): T {
  if (val === undefined || val === null) return val
  return JSON.parse(JSON.stringify(val))
}

export function prettyJson(val: any): string {
  if (val === undefined) return ''
  if (typeof val === 'string') {
    try {
      return JSON.stringify(JSON.parse(val), null, 2)
    } catch {
      return val
    }
  }
  return JSON.stringify(val, null, 2)
}

/** 解析 JSON 文本，返回 [值, 错误信息] */
export function parseJson(text: string): [any, string] {
  if (!text || !text.trim()) return [undefined, '']
  try {
    return [JSON.parse(text), '']
  } catch (err) {
    return [undefined, (err as Error).message]
  }
}

export function formatDuration(ms?: number): string {
  if (ms === undefined || ms === null) return '-'
  if (ms < 1000) return `${ms} ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)} s`
  const minutes = Math.floor(ms / 60000)
  const seconds = Math.round((ms % 60000) / 1000)
  return `${minutes} 分 ${seconds} 秒`
}

export function downloadJson(fileName: string, content: any) {
  const blob = new Blob([JSON.stringify(content, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  a.click()
  URL.revokeObjectURL(url)
}

export function readFileText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsText(file, 'utf-8')
  })
}

export async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    return false
  }
}

/** 时间范围转查询参数 */
export function rangeParams(range?: [string, string] | null) {
  if (!range || range.length !== 2) return {}
  return { start_time: range[0], end_time: range[1] }
}
