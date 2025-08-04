// 视频相关类型定义

// 从后端API接收的原始视频对象
export interface BackendVideo {
  id: string;
  title: string;
  filename: string;
  content_type: string;
  size: number;
  duration: number;
  width: number;
  height: number;
  storage_path: string;
  thumbnail_path: string;
  uploaded_at: number;
  updated_at: number;
  description?: string;
}

// 从后端API接收的原始视频列表响应
export interface BackendVideoListResponse {
  base: {
    code: number;
    message: string;
  };
  videos: BackendVideo[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}


// 前端组件使用的经过转换的视频对象
export interface Video {
  id: string
  title: string
  description?: string
  filename: string
  file_size: number
  duration: number
  mime_type: string
  thumbnail_url?: string
  play_url?: string
  created_at: string
  updated_at: string
}

export interface VideoListRequest {
  page?: number
  limit?: number
  search?: string
  sort_by?: 'created_at' | 'title' | 'duration' | 'file_size'
  sort_order?: 'asc' | 'desc'
}

export interface VideoListResponse {
  videos: Video[]
  total: number
  page: number
  limit: number
  has_next: boolean
  has_prev: boolean
}

export interface VideoUploadRequest {
  title: string
  description?: string
  file: File
}

export interface VideoUploadResponse {
  base: {
    code: number
    message: string
  }
  video_id: string
  upload_url?: string
}

export interface VideoUploadProgress {
  loaded: number
  total: number
  percentage: number
}

// API响应类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

export interface ApiError {
  code: number
  message: string
  details?: string
}

// 上传状态类型
export type UploadStatus = 'idle' | 'uploading' | 'success' | 'error'

// 播放器状态类型
export interface PlayerState {
  isPlaying: boolean
  currentTime: number
  duration: number
  volume: number
  isMuted: boolean
  isFullscreen: boolean
  isLoading: boolean
}

// 分页类型
export interface Pagination {
  page: number
  limit: number
  total: number
  has_next: boolean
  has_prev: boolean
}

// 表单类型
export interface VideoUploadForm {
  title: string
  description: string
  file: File | null
}

export interface VideoEditForm {
  title: string
  description: string
}

// 搜索过滤类型
export interface VideoFilter {
  search: string
  sortBy: string
  sortOrder: 'asc' | 'desc'
  dateRange?: {
    start: string
    end: string
  }
}