// 通用类型定义

export interface BaseEntity {
  id: string
  created_at: string
  updated_at: string
}

export interface SelectOption {
  value: string
  label: string
  disabled?: boolean
}

export interface TableColumn<T = any> {
  key: keyof T
  title: string
  width?: string
  align?: 'left' | 'center' | 'right'
  sortable?: boolean
  render?: (value: any, record: T) => React.ReactNode
}

export interface ModalProps {
  open: boolean
  title: string
  onCancel: () => void
  onOk?: () => void
  confirmLoading?: boolean
  width?: number | string
  footer?: React.ReactNode
}

export interface ToastMessage {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

// 环境配置类型
export interface AppConfig {
  API_BASE_URL: string
  API_VERSION: string
  MAX_FILE_SIZE: number
  SUPPORTED_VIDEO_FORMATS: string[]
  CHUNK_SIZE: number
}

// 错误类型
export interface ErrorBoundaryState {
  hasError: boolean
  error?: Error
  errorInfo?: React.ErrorInfo
}

// 路由类型
export interface RouteConfig {
  path: string
  component: React.ComponentType
  title: string
  exact?: boolean
  requireAuth?: boolean
}

// HTTP客户端类型
export interface HttpClientConfig {
  baseURL: string
  timeout: number
  headers?: Record<string, string>
}

export interface RequestConfig {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  data?: any
  params?: Record<string, any>
  headers?: Record<string, string>
  onUploadProgress?: (progressEvent: any) => void
}

// 本地存储类型
export type StorageKey = 'theme' | 'language' | 'user_preferences' | 'recent_videos'

export interface UserPreferences {
  theme: 'light' | 'dark' | 'auto'
  language: 'zh-CN' | 'en-US'
  video_quality: 'auto' | '720p' | '1080p'
  auto_play: boolean
  volume: number
}