import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { VideoService } from '../../services'
import type { Video, VideoListResponse, VideoFilter, Pagination } from '../../types'

interface VideoListProps {
  searchQuery?: string
  onVideoSelect?: (video: Video) => void
  onVideoDeleted?: (videoId: string) => void
}

export default function VideoList({ searchQuery, onVideoSelect, onVideoDeleted }: VideoListProps) {
  const [videos, setVideos] = useState<Video[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [deletingVideoId, setDeletingVideoId] = useState<string | null>(null)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [pagination, setPagination] = useState<Pagination>({
    page: 1,
    limit: 12,
    total: 0,
    has_next: false,
    has_prev: false,
  })
  const [filter, setFilter] = useState<VideoFilter>({
    search: searchQuery || '',
    sortBy: 'created_at',
    sortOrder: 'desc',
  })

  const fetchVideos = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      const response: VideoListResponse = await VideoService.getVideoList({
        page: pagination.page,
        limit: pagination.limit,
        search: filter.search,
        sort_by: filter.sortBy as any,
        sort_order: filter.sortOrder,
      })

      setVideos(response.videos)
      setPagination({
        page: response.page,
        limit: response.limit,
        total: response.total,
        has_next: response.has_next,
        has_prev: response.has_prev,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取视频列表失败')
    } finally {
      setLoading(false)
    }
  }, [pagination.page, pagination.limit, filter])

  useEffect(() => {
    fetchVideos()
  }, [fetchVideos])

  useEffect(() => {
    if (searchQuery !== undefined) {
      setFilter(prev => ({ ...prev, search: searchQuery }))
      setPagination(prev => ({ ...prev, page: 1 }))
    }
  }, [searchQuery])

  const handleSortChange = useCallback((sortBy: string) => {
    setFilter(prev => ({
      ...prev,
      sortBy,
      sortOrder: prev.sortBy === sortBy && prev.sortOrder === 'desc' ? 'asc' : 'desc',
    }))
    setPagination(prev => ({ ...prev, page: 1 }))
  }, [])

  const handlePageChange = useCallback((page: number) => {
    setPagination(prev => ({ ...prev, page }))
  }, [])

  const handleDeleteVideo = useCallback(async (videoId: string, videoTitle: string) => {
    if (!window.confirm(`确定要删除视频"${videoTitle}"吗？此操作不可撤销。`)) {
      return
    }

    setDeletingVideoId(videoId)
    
    try {
      await VideoService.deleteVideo(videoId)
      
      // 从列表中移除删除的视频
      setVideos(prev => prev.filter(video => video.id !== videoId))
      
      // 更新分页信息
      setPagination(prev => ({ ...prev, total: prev.total - 1 }))
      
      // 通知父组件
      onVideoDeleted?.(videoId)
      
      // 如果当前页没有视频了且不是第一页，则跳转到上一页
      if (videos.length === 1 && pagination.page > 1) {
        setPagination(prev => ({ ...prev, page: prev.page - 1 }))
      }
      
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除视频失败')
    } finally {
      setDeletingVideoId(null)
    }
  }, [videos.length, pagination.page, onVideoDeleted])

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const remainingSeconds = seconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`
    }
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 mb-4">❌ {error}</div>
        <button
          onClick={fetchVideos}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
        >
          重试
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* 工具栏 */}
      <div className="flex justify-between items-center">
        <div className="text-sm text-gray-600">
          共 {pagination.total} 个视频
        </div>
        <div className="flex items-center space-x-4">
          {/* 视图切换 */}
          <div className="flex border border-gray-300 rounded">
            <button
              onClick={() => setViewMode('grid')}
              className={`px-3 py-1 text-sm ${
                viewMode === 'grid' 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-white text-gray-600 hover:bg-gray-50'
              }`}
            >
              网格
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`px-3 py-1 text-sm ${
                viewMode === 'list' 
                  ? 'bg-blue-600 text-white' 
                  : 'bg-white text-gray-600 hover:bg-gray-50'
              }`}
            >
              列表
            </button>
          </div>
          
          {/* 排序选择 */}
          <select
            value={filter.sortBy}
            onChange={(e) => handleSortChange(e.target.value)}
            className="px-3 py-1 border border-gray-300 rounded text-sm"
          >
            <option value="created_at">按创建时间</option>
            <option value="title">按标题</option>
            <option value="duration">按时长</option>
            <option value="file_size">按文件大小</option>
          </select>
          <button
            onClick={() => setFilter(prev => ({ 
              ...prev, 
              sortOrder: prev.sortOrder === 'desc' ? 'asc' : 'desc' 
            }))}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            {filter.sortOrder === 'desc' ? '↓' : '↑'}
          </button>
        </div>
      </div>

      {/* 视频显示区域 */}
      {videos.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-500 mb-4">📺 暂无视频</div>
          <Link
            to="/upload"
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
          >
            上传第一个视频
          </Link>
        </div>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {videos.map((video) => (
            <div
              key={video.id}
              className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => onVideoSelect?.(video)}
            >
              {/* 网格视图的卡片内容 */}
              <div className="aspect-video bg-gray-200 relative">
                {video.thumbnail_url ? (
                  <img
                    src={video.thumbnail_url}
                    alt={video.title}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-4xl text-gray-400">
                    🎬
                  </div>
                )}
                <div className="absolute bottom-2 right-2 bg-black bg-opacity-70 text-white text-xs px-2 py-1 rounded">
                  {formatDuration(video.duration)}
                </div>
              </div>

              <div className="p-4">
                <h3 className="font-medium text-gray-900 mb-2 line-clamp-2">
                  {video.title}
                </h3>
                {video.description && (
                  <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                    {video.description}
                  </p>
                )}
                <div className="text-xs text-gray-500 space-y-1">
                  <div>{formatFileSize(video.file_size)}</div>
                  <div>{formatDate(video.created_at)}</div>
                </div>

                <div className="mt-3 flex space-x-2">
                  <Link
                    to={`/video/${video.id}`}
                    className="flex-1 bg-blue-600 text-white text-center py-1 px-2 rounded text-sm hover:bg-blue-700"
                    onClick={(e) => e.stopPropagation()}
                  >
                    播放
                  </Link>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDeleteVideo(video.id, video.title)
                    }}
                    disabled={deletingVideoId === video.id}
                    className="px-2 py-1 text-red-600 hover:bg-red-50 rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {deletingVideoId === video.id ? (
                      <div className="flex items-center">
                        <div className="animate-spin rounded-full h-3 w-3 border-b border-red-600 mr-1"></div>
                        删除中
                      </div>
                    ) : (
                      '删除'
                    )}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          {videos.map((video) => (
            <div
              key={video.id}
              className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
              onClick={() => onVideoSelect?.(video)}
            >
              {/* 列表视图的行内容 */}
              <div className="flex">
                <div className="w-48 h-28 bg-gray-200 relative flex-shrink-0">
                  {video.thumbnail_url ? (
                    <img
                      src={video.thumbnail_url}
                      alt={video.title}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-2xl text-gray-400">
                      🎬
                    </div>
                  )}
                  <div className="absolute bottom-1 right-1 bg-black bg-opacity-70 text-white text-xs px-1 py-0.5 rounded">
                    {formatDuration(video.duration)}
                  </div>
                </div>
                
                <div className="flex-1 p-4">
                  <div className="flex justify-between items-start">
                    <div className="flex-1 mr-4">
                      <h3 className="font-medium text-gray-900 mb-2">{video.title}</h3>
                      {video.description && (
                        <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                          {video.description}
                        </p>
                      )}
                      <div className="text-xs text-gray-500 space-x-4">
                        <span>{formatFileSize(video.file_size)}</span>
                        <span>{formatDate(video.created_at)}</span>
                      </div>
                    </div>
                    
                    <div className="flex space-x-2 flex-shrink-0">
                      <Link
                        to={`/video/${video.id}`}
                        className="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700"
                        onClick={(e) => e.stopPropagation()}
                      >
                        播放
                      </Link>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteVideo(video.id, video.title)
                        }}
                        disabled={deletingVideoId === video.id}
                        className="px-3 py-1 text-red-600 hover:bg-red-50 rounded text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {deletingVideoId === video.id ? (
                          <div className="flex items-center">
                            <div className="animate-spin rounded-full h-3 w-3 border-b border-red-600 mr-1"></div>
                            删除中
                          </div>
                        ) : (
                          '删除'
                        )}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 分页 */}
      {pagination.total > pagination.limit && (
        <div className="flex justify-center items-center space-x-2">
          <button
            onClick={() => handlePageChange(pagination.page - 1)}
            disabled={!pagination.has_prev}
            className="px-3 py-1 border border-gray-300 rounded disabled:bg-gray-100 disabled:text-gray-400"
          >
            上一页
          </button>
          
          <span className="text-sm text-gray-600">
            第 {pagination.page} 页，共 {Math.ceil(pagination.total / pagination.limit)} 页
          </span>

          <button
            onClick={() => handlePageChange(pagination.page + 1)}
            disabled={!pagination.has_next}
            className="px-3 py-1 border border-gray-300 rounded disabled:bg-gray-100 disabled:text-gray-400"
          >
            下一页
          </button>
        </div>
      )}
    </div>
  )
}