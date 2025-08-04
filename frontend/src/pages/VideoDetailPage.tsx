import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import VideoPlayer from '../components/VideoPlayer'
import { VideoService } from '../services'
import type { Video } from '../types'

export default function VideoDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [video, setVideo] = useState<Video | null>(null)
  const [playUrl, setPlayUrl] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchVideoData = async () => {
      if (!id || id === 'undefined') {
        setError('视频ID无效')
        setLoading(false)
        return
      }

      try {
        setLoading(true)
        
        const videoData = await VideoService.getVideoById(id)
        setVideo(videoData)

        if (videoData) {
          const playUrlData = await VideoService.getVideoPlayUrl(id)
          setPlayUrl(playUrlData.play_url)
        }

      } catch (err) {
        setError(err instanceof Error ? err.message : '获取视频信息失败')
      } finally {
        setLoading(false)
      }
    }

    fetchVideoData()
  }, [id])

  const handleVideoError = (errorMessage: string) => {
    setError(errorMessage)
  }

  const handleVideoEnded = () => {
    // 视频播放结束的处理逻辑
    console.log('Video ended')
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const remainingSeconds = seconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`
    }
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (error || !video) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 mb-4">❌ {error || '视频不存在'}</div>
        <div className="space-x-4">
          <button
            onClick={() => window.location.reload()}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
          >
            重试
          </button>
          <Link
            to="/"
            className="bg-gray-600 text-white px-4 py-2 rounded-md hover:bg-gray-700"
          >
            返回首页
          </Link>
        </div>
      </div>
    )
  }

  // 等待播放URL加载完成
  if (!playUrl) {
    return (
      <div className="flex justify-center items-center min-h-96">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <div className="ml-3 text-gray-600">正在获取播放地址...</div>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-6">
      {/* 导航 */}
      <div className="mb-6">
        <nav className="flex items-center space-x-2 text-sm text-gray-600">
          <Link to="/" className="hover:text-blue-600">首页</Link>
          <span>/</span>
          <span className="text-gray-900">{video.title}</span>
        </nav>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* 视频播放器 */}
        <div className="lg:col-span-2">
          <VideoPlayer
            video={{...video, play_url: playUrl}}
            autoPlay={false}
            onError={handleVideoError}
            onEnded={handleVideoEnded}
          />
        </div>

        {/* 视频信息 */}
        <div className="space-y-6">
          {/* 基本信息 */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h1 className="text-2xl font-bold text-gray-900 mb-4">{video.title}</h1>
            
            {video.description && (
              <div className="mb-4">
                <h3 className="text-sm font-medium text-gray-700 mb-2">描述</h3>
                <p className="text-gray-600 text-sm leading-relaxed">{video.description}</p>
              </div>
            )}

            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">文件大小</span>
                <span className="font-medium">{formatFileSize(video.file_size)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">视频时长</span>
                <span className="font-medium">{formatDuration(video.duration)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">文件格式</span>
                <span className="font-medium">{video.mime_type}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">上传时间</span>
                <span className="font-medium">{formatDate(video.created_at)}</span>
              </div>
              {video.updated_at !== video.created_at && (
                <div className="flex justify-between">
                  <span className="text-gray-600">更新时间</span>
                  <span className="font-medium">{formatDate(video.updated_at)}</span>
                </div>
              )}
            </div>
          </div>

          {/* 操作按钮 */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900 mb-4">操作</h3>
            <div className="space-y-3">
              <button
                onClick={() => {
                  // TODO: 实现编辑功能
                  console.log('Edit video:', video.id)
                }}
                className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
              >
                编辑信息
              </button>
              <button
                onClick={() => {
                  // 使用获取到的预签名URL下载视频
                  window.open(playUrl, '_blank')
                }}
                className="w-full bg-green-600 text-white py-2 px-4 rounded-md hover:bg-green-700 transition-colors"
              >
                下载视频
              </button>
              <button
                onClick={() => {
                  // TODO: 实现删除功能
                  if (window.confirm('确定要删除这个视频吗？此操作不可撤销。')) {
                    console.log('Delete video:', video.id)
                  }
                }}
                className="w-full bg-red-600 text-white py-2 px-4 rounded-md hover:bg-red-700 transition-colors"
              >
                删除视频
              </button>
            </div>
          </div>

          {/* 技术信息 */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900 mb-4">技术信息</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">视频ID</span>
                <span className="font-mono text-xs">{video.id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">文件名</span>
                <span className="font-mono text-xs break-all">{video.filename}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}