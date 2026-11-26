import { optionApi } from '@/api'
import type { OptionItem } from '@/types'

/** 动态下拉选项缓存有效期：同一数据源在有效期内只请求一次 */
const CACHE_TTL = 60_000

const cache = new Map<string, { at: number; promise: Promise<OptionItem[]> }>()

/** 获取动态下拉数据源的选项，force 为 true 时忽略缓存 */
export function loadOptions(source: string, force = false): Promise<OptionItem[]> {
  const hit = cache.get(source)
  if (!force && hit && Date.now() - hit.at < CACHE_TTL) return hit.promise
  const entry = { at: Date.now(), promise: optionApi.list(source).then((list) => list ?? []) }
  cache.set(source, entry)
  // 失败的请求不缓存，下次打开下拉时重试
  entry.promise.catch(() => {
    if (cache.get(source) === entry) cache.delete(source)
  })
  return entry.promise
}
