import { useState, useRef } from 'react'
import { motion } from 'framer-motion'
import { Upload, FileVideo, X, CheckCircle, AlertCircle, Loader2 } from 'lucide-react'
import { Link } from 'react-router-dom'

interface VideoFile {
  file: File
  id: string
  title: string
  size: number
  duration?: number
  progress: number
  status: 'pending' | 'uploading' | 'completed' | 'error'
  error?: string
}

const VideoUpload = () => {
  const [videos, setVideos] = useState<VideoFile[]>([])
  const [dragActive, setDragActive] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFiles(e.dataTransfer.files)
    }
  }

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFiles(e.target.files)
    }
  }

  const handleFiles = (files: FileList) => {
    const newVideos: VideoFile[] = []
    
    Array.from(files).forEach(file => {
      if (file.type.startsWith('video/')) {
        newVideos.push({
          file,
          id: Math.random().toString(36).substr(2, 9),
          title: file.name.replace(/\.[^/.]+$/, ''),
          size: file.size,
          progress: 0,
          status: 'pending'
        })
      }
    })

    setVideos(prev => [...prev, ...newVideos])
  }

  const removeVideo = (id: string) => {
    setVideos(prev => prev.filter(video => video.id !== id))
  }

  const formatSize = (bytes: number) => {
    const mb = bytes / (1024 * 1024)
    return `${mb.toFixed(1)} MB`
  }

  const uploadVideo = async (video: VideoFile) => {
    setVideos(prev => prev.map(v => 
      v.id === video.id ? { ...v, status: 'uploading' } : v
    ))

    // 模拟上传过程
    for (let progress = 0; progress <= 100; progress += 10) {
      await new Promise(resolve => setTimeout(resolve, 200))
      setVideos(prev => prev.map(v => 
        v.id === video.id ? { ...v, progress } : v
      ))
    }

    // 模拟上传完成
    setVideos(prev => prev.map(v => 
      v.id === video.id ? { ...v, status: 'completed' } : v
    ))
  }

  const getStatusIcon = (status: VideoFile['status']) => {
    switch (status) {
      case 'uploading':
        return <Loader2 className="w-4 h-4 animate-spin text-cyber-blue" />
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-cyber-green" />
      case 'error':
        return <AlertCircle className="w-4 h-4 text-red-500" />
      default:
        return <FileVideo className="w-4 h-4 text-gray-400" />
    }
  }

  const getStatusText = (status: VideoFile['status']) => {
    switch (status) {
      case 'pending':
        return '等待上传'
      case 'uploading':
        return '上传中'
      case 'completed':
        return '上传完成'
      case 'error':
        return '上传失败'
      default:
        return '未知状态'
    }
  }

  const totalSize = videos.reduce((sum, video) => sum + video.size, 0)
  const uploadingCount = videos.filter(v => v.status === 'uploading').length
  const completedCount = videos.filter(v => v.status === 'completed').length

  return (
    <div className="min-h-screen py-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="max-w-4xl mx-auto px-4"
      >
        <div className="mb-6">
          <Link
            to="/"
            className="cyber-button mb-4 inline-block"
          >
            ← 返回首页
          </Link>
          <h1 className="text-3xl font-bold neon-text mb-2">视频上传</h1>
          <p className="text-gray-400">支持多种视频格式，快速上传到烛龙系统</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <div className="cyber-card">
              <div
                className={`border-2 border-dashed rounded-lg p-8 text-center transition-all duration-300 ${
                  dragActive
                    ? 'border-cyber-blue bg-cyber-blue/10'
                    : 'border-dark-border hover:border-cyber-blue/50'
                }`}
                onDragEnter={handleDrag}
                onDragLeave={handleDrag}
                onDragOver={handleDrag}
                onDrop={handleDrop}
              >
                <Upload className="w-16 h-16 text-cyber-blue mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-white mb-2">拖拽视频文件到这里</h3>
                <p className="text-gray-400 mb-4">或者点击选择文件</p>
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="cyber-button"
                >
                  选择文件
                </button>
                <input
                  ref={fileInputRef}
                  type="file"
                  multiple
                  accept="video/*"
                  onChange={handleFileInput}
                  className="hidden"
                />
                <div className="mt-4 text-xs text-gray-500">
                  支持格式: MP4, WebM, AVI, MOV (最大2GB)
                </div>
              </div>
            </div>

            {videos.length > 0 && (
              <div className="cyber-card">
                <h3 className="text-lg font-semibold text-white mb-4">上传队列</h3>
                <div className="space-y-4">
                  {videos.map((video) => (
                    <div key={video.id} className="border border-dark-border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center space-x-3">
                          {getStatusIcon(video.status)}
                          <div>
                            <h4 className="text-white font-medium">{video.title}</h4>
                            <p className="text-sm text-gray-400">
                              {formatSize(video.size)} • {getStatusText(video.status)}
                            </p>
                          </div>
                        </div>
                        <button
                          onClick={() => removeVideo(video.id)}
                          className="p-1 rounded hover:bg-dark-card transition-colors"
                        >
                          <X className="w-4 h-4 text-gray-400" />
                        </button>
                      </div>

                      {video.status === 'uploading' && (
                        <div className="space-y-2">
                          <div className="cyber-progress">
                            <div
                              className="cyber-progress-bar"
                              style={{ width: `${video.progress}%` }}
                            ></div>
                          </div>
                          <div className="text-xs text-gray-400 text-right">
                            {video.progress}% 完成
                          </div>
                        </div>
                      )}

                      {video.status === 'pending' && (
                        <button
                          onClick={() => uploadVideo(video)}
                          className="cyber-button w-full"
                        >
                          开始上传
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="space-y-6">
            <div className="cyber-card">
              <h3 className="text-lg font-semibold text-white mb-4">上传统计</h3>
              <div className="space-y-4">
                <div className="flex justify-between">
                  <span className="text-gray-400">队列中:</span>
                  <span className="text-white">{videos.length}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">上传中:</span>
                  <span className="text-cyber-blue">{uploadingCount}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">已完成:</span>
                  <span className="text-cyber-green">{completedCount}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">总大小:</span>
                  <span className="text-white">{formatSize(totalSize)}</span>
                </div>
              </div>
            </div>

            <div className="cyber-card">
              <h3 className="text-lg font-semibold text-white mb-4">系统状态</h3>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">存储空间:</span>
                  <div className="flex items-center space-x-2">
                    <div className="w-2 h-2 bg-cyber-green rounded-full animate-pulse"></div>
                    <span className="text-cyber-green">充足</span>
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">网络带宽:</span>
                  <span className="text-cyber-blue">100 Mbps</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">服务器状态:</span>
                  <span className="text-cyber-green">在线</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">最大文件大小:</span>
                  <span className="text-white">2 GB</span>
                </div>
              </div>
            </div>

            <div className="cyber-card">
              <h3 className="text-lg font-semibold text-white mb-4">支持格式</h3>
              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">MP4:</span>
                  <span className="text-cyber-green">✓ 推荐</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">WebM:</span>
                  <span className="text-cyber-green">✓ 支持</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">AVI:</span>
                  <span className="text-cyber-green">✓ 支持</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-400">MOV:</span>
                  <span className="text-cyber-green">✓ 支持</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </motion.div>
    </div>
  )
}

export default VideoUpload