import axios from 'axios'
import type { AxiosInstance, AxiosResponse, AxiosError } from 'axios'
import type { ApiResponse, ApiError } from '../types'

// 创建axios实例
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8888/api/v1',
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // 请求拦截器
  client.interceptors.request.use(
    (config) => {
      // 在这里可以添加认证token等
      const token = localStorage.getItem('auth_token')
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      
      console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`)
      return config
    },
    (error) => {
      console.error('Request interceptor error:', error)
      return Promise.reject(error)
    }
  )

  // 响应拦截器
  client.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      console.log(`API Response: ${response.status} ${response.config.url}`)
      return response
    },
    (error: AxiosError<ApiError>) => {
      console.error('API Error:', error.response?.data || error.message)
      
      // 统一错误处理
      if (error.response?.status === 401) {
        // 未授权，清除token并重定向到登录页
        localStorage.removeItem('auth_token')
        window.location.href = '/login'
      } else if (error.response?.status === 403) {
        // 无权限
        console.error('Access forbidden')
      } else if (error.response?.status && error.response.status >= 500) {
        // 服务器错误
        console.error('Server error')
      }

      return Promise.reject(error)
    }
  )

  return client
}

// 导出API客户端实例
export const apiClient = createApiClient()

// 通用API调用封装
export const apiRequest = async <T = any>(
  url: string,
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' = 'GET',
  data?: any,
  config?: any
): Promise<T> => {
  try {
    const response = await apiClient.request<ApiResponse<T>>({
      url,
      method,
      data,
      ...config,
    })

    if (response.data.base.code === 0) {
      return response.data as T
    } else {
      throw new Error(response.data.base.message || 'API request failed')
    }
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const apiError = error.response?.data as ApiError
      throw new Error(apiError?.message || error.message)
    }
    throw error
  }
}

// 文件上传专用函数
export const uploadFile = async (
  url: string,
  file: File,
  onProgress?: (progress: number) => void
): Promise<any> => {
  const formData = new FormData()
  formData.append('file', file)

  try {
    const response = await apiClient.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total && onProgress) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      },
    })

    return response.data
  } catch (error) {
    console.error('File upload error:', error)
    throw error
  }
}

// 错误处理工具函数
export const handleApiError = (error: unknown): string => {
  if (axios.isAxiosError(error)) {
    const apiError = error.response?.data as ApiError
    return apiError?.message || error.message || '网络请求失败'
  }
  if (error instanceof Error) {
    return error.message
  }
  return '未知错误'
}