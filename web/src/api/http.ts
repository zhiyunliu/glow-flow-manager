import axios, { type AxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'

/**
 * 接口根路径：页面挂载在 {APIRoot}/web/ 下，去掉 /web/ 之后的部分即为接口根路径，
 * 因此经过网关前缀(如 /msweb/api/manager-glow/web/)访问时也能正确调用接口。
 */
export const API_BASE = (() => {
  const path = window.location.pathname
  const idx = path.indexOf('/web/')
  if (idx >= 0) return path.substring(0, idx)
  if (path.endsWith('/web')) return path.substring(0, path.length - 4)
  return '/api/manager-glow'
})()

/** 接口错误：HTTP 状态码 + 统一响应中的 code/sub_code/message/data */
export class ApiError extends Error {
  status: number
  code: number
  subCode: string
  data: any

  constructor(status: number, body: any, fallback: string) {
    const message = (body && typeof body === 'object' && body.message) || fallback || '请求失败'
    super(message)
    this.status = status
    this.code = body?.code ?? status
    this.subCode = body?.sub_code ?? ''
    this.data = body?.data
  }
}

const http = axios.create({ baseURL: API_BASE, timeout: 150000 })

http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) return body.data
      return Promise.reject(new ApiError(res.status, body, ''))
    }
    return body
  },
  (err) => {
    const status = err.response?.status ?? 0
    const fallback = status === 0 ? '无法连接服务' : err.message
    return Promise.reject(new ApiError(status, err.response?.data, fallback))
  },
)

export interface RequestOptions {
  /** 为 true 时不自动提示错误，由调用方处理 */
  silent?: boolean
}

async function request<T>(config: AxiosRequestConfig, opts?: RequestOptions): Promise<T> {
  try {
    return (await http.request(config)) as T
  } catch (err) {
    if (!opts?.silent) {
      ElMessage.error({ message: (err as Error).message, grouping: true })
    }
    throw err
  }
}

export function get<T>(url: string, params?: Record<string, any>, opts?: RequestOptions): Promise<T> {
  return request<T>({ method: 'GET', url, params: cleanParams(params) }, opts)
}

export function post<T>(url: string, data?: any, opts?: RequestOptions): Promise<T> {
  return request<T>({ method: 'POST', url, data: data ?? {} }, opts)
}

/** 去掉空查询参数 */
function cleanParams(params?: Record<string, any>) {
  if (!params) return params
  const result: Record<string, any> = {}
  for (const [key, val] of Object.entries(params)) {
    if (val === undefined || val === null || val === '') continue
    result[key] = val
  }
  return result
}
