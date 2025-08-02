import { apiRequest, uploadFile } from './api'
import type {
  Video,
  VideoListRequest,
  VideoListResponse,
  VideoUploadRequest,
  VideoUploadResponse,
  VideoUploadProgress,
} from '../types'

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

    return apiRequest<VideoListResponse>(url, 'GET')
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
    // 首先创建视频记录并获取上传URL
    const uploadInfo = await apiRequest<VideoUploadResponse>(
      this.BASE_PATH,
      'POST',
      {
        title: request.title,
        description: request.description,
        filename: request.file.name,
        file_size: request.file.size,
        mime_type: request.file.type,
      }
    )

    // 使用预签名URL上传文件
    if (uploadInfo.upload_url) {
      await uploadFile(
        uploadInfo.upload_url,
        request.file,
        (progress) => {
          if (onProgress) {
            onProgress({
              loaded: (request.file.size * progress) / 100,
              total: request.file.size,
              percentage: progress,
            })
          }
        }
      )
    }

    return uploadInfo
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