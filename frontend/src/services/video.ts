import { apiRequest, apiClient } from './api'
import axios from 'axios'
import type {
  Video,
  VideoListRequest,
  VideoListResponse,
  VideoUploadRequest,
  VideoUploadResponse,
  VideoUploadProgress,
  BackendVideo,
  BackendVideoListResponse,
} from '../types'

// 将后端视频数据转换为前端格式
function transformVideoData(backendVideo: BackendVideo): Video {
  const minioBaseUrl = import.meta.env.VITE_MINIO_BASE_URL || 'http://localhost:9000';
  const bucketName = import.meta.env.VITE_MINIO_BUCKET_NAME || 'zhulong-videos';

  return {
    id: backendVideo.id,
    title: backendVideo.title,
    description: backendVideo.description,
    filename: backendVideo.filename,
    file_size: backendVideo.size,
    duration: backendVideo.duration,
    mime_type: backendVideo.content_type,
    thumbnail_url: backendVideo.thumbnail_path
      ? `${minioBaseUrl}/${bucketName}/${backendVideo.thumbnail_path}`
      : 'https://via.placeholder.com/1280x720.png?text=No+Thumbnail',
    created_at: new Date(backendVideo.uploaded_at).toISOString(),
    updated_at: new Date(backendVideo.updated_at).toISOString(),
  };
}

// 视频API服务类
export class VideoService {
  private static readonly BASE_PATH = '/videos'

  /**
   * 获取视频列表
   * @param params 查询参数
   * @returns 视频列表响应
   */
  static async getVideoList(params: VideoListRequest = {}): Promise<VideoListResponse> {
    const queryParams = new URLSearchParams()
    
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', params.limit.toString())
    if (params.search) queryParams.append('search', params.search)
    if (params.sort_by) queryParams.append('sort_by', params.sort_by)
    if (params.sort_order) queryParams.append('sort_order', params.sort_order)

    const queryString = queryParams.toString()
    const url = queryString ? `${this.BASE_PATH}?${queryString}` : this.BASE_PATH

    const backendResponse = await apiRequest<BackendVideoListResponse>(url, 'GET')

    if (backendResponse.base.code !== 0) {
      throw new Error(backendResponse.base.message || 'Failed to fetch video list');
    }

    return {
      videos: backendResponse.videos.map(transformVideoData),
      total: backendResponse.total,
      page: backendResponse.page,
      limit: backendResponse.page_size,
      has_next: backendResponse.page < backendResponse.total_pages,
      has_prev: backendResponse.page > 1,
    };
  }

  /**
   * 获取单个视频详情
   * @param id 视频ID
   * @returns 视频信息
   */
  static async getVideoById(id: string): Promise<Video> {
    return apiRequest<Video>(`${this.BASE_PATH}/${id}`, 'GET')
  }

  /**
   * 获取视频播放URL
   * @param id 视频ID
   * @returns 播放URL信息
   */
  static async getVideoPlayUrl(id: string): Promise<{ play_url: string }> {
    return apiRequest<{ play_url: string }>(`${this.BASE_PATH}/${id}/play`, 'GET')
  }

  /**
   * 上传视频文件
   * @param request 上传请求参数
   * @param onProgress 上传进度回调
   * @returns 上传响应
   */
  static async uploadVideo(
    request: VideoUploadRequest,
    onProgress?: (progress: VideoUploadProgress) => void
  ): Promise<VideoUploadResponse> {
    // 创建FormData对象
    const formData = new FormData()
    formData.append('file', request.file)
    formData.append('title', request.title)
    formData.append('description', request.description || '')

    try {
      // 使用FormData直接上传到后端API
      const response = await apiClient.post<VideoUploadResponse>(
        this.BASE_PATH,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
          onUploadProgress: (progressEvent) => {
            if (progressEvent.total && onProgress) {
              const loaded = progressEvent.loaded
              const total = progressEvent.total
              const percentage = Math.round((loaded * 100) / total)
              
              onProgress({
                loaded,
                total,
                percentage,
              })
            }
          },
        }
      )

      // 检查响应格式
      if (response.data.base?.code === 0) {
        return response.data
      } else {
        throw new Error(response.data.base?.message || '上传失败')
      }
    } catch (error) {
      console.error('Upload error:', error)
      if (axios.isAxiosError(error)) {
        const apiError = error.response?.data
        throw new Error(apiError?.base?.message || apiError?.message || '上传失败')
      }
      throw error
    }
  }

  /**
   * 删除视频
   * @param id 视频ID
   * @returns 删除结果
   */
  static async deleteVideo(id: string): Promise<{ message: string }> {
    return apiRequest<{ message: string }>(`${this.BASE_PATH}/${id}`, 'DELETE')
  }

  /**
   * 更新视频信息
   * @param id 视频ID
   * @param data 更新数据
   * @returns 更新后的视频信息
   */
  static async updateVideo(
    id: string,
    data: { title?: string; description?: string }
  ): Promise<Video> {
    return apiRequest<Video>(`${this.BASE_PATH}/${id}`, 'PUT', data)
  }

  /**
   * 搜索视频
   * @param query 搜索关键词
   * @param options 搜索选项
   * @returns 搜索结果
   */
  static async searchVideos(
    query: string,
    options: {
      page?: number
      limit?: number
      sort_by?: 'created_at' | 'title' | 'duration' | 'file_size'
      sort_order?: 'asc' | 'desc'
    } = {}
  ): Promise<VideoListResponse> {
    return this.getVideoList({
      search: query,
      ...options,
    })
  }

  /**
   * 获取视频统计信息
   * @returns 统计信息
   */
  static async getVideoStats(): Promise<{
    total_videos: number
    total_size: number
    total_duration: number
    recent_uploads: number
  }> {
    return apiRequest<{
      total_videos: number
      total_size: number
      total_duration: number
      recent_uploads: number
    }>(`${this.BASE_PATH}/stats`, 'GET')
  }
}