import { useState, useCallback } from 'react'
import { VideoService } from '../../services'
import type { VideoUploadForm, UploadStatus, VideoUploadProgress } from '../../types'

interface VideoUploadProps {
  onUploadSuccess?: (videoId: string) => void
  onUploadError?: (error: string) => void
}

export default function VideoUpload({ onUploadSuccess, onUploadError }: VideoUploadProps) {
  const [form, setForm] = useState<VideoUploadForm>({
    title: '',
    description: '',
    file: null,
  })
  const [uploadStatus, setUploadStatus] = useState<UploadStatus>('idle')
  const [uploadProgress, setUploadProgress] = useState<VideoUploadProgress>({
    loaded: 0,
    total: 0,
    percentage: 0,
  })
  const [dragActive, setDragActive] = useState(false)

  const handleInputChange = useCallback((field: keyof VideoUploadForm, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
  }, [])

  const handleFileSelect = useCallback((file: File) => {
    // 验证文件类型
    const supportedTypes = ['video/mp4', 'video/webm', 'video/avi', 'video/mov']
    if (!supportedTypes.includes(file.type)) {
      onUploadError?.('不支持的文件格式。支持的格式：MP4, WebM, AVI, MOV')
      return
    }

    // 验证文件大小 (2GB限制)
    const maxSize = 2 * 1024 * 1024 * 1024
    if (file.size > maxSize) {
      onUploadError?.('文件太大。最大支持2GB')
      return
    }

    setForm(prev => ({ ...prev, file }))
    
    // 如果标题为空，使用文件名作为默认标题
    if (!form.title) {
      const filename = file.name.replace(/\.[^/.]+$/, '')
      setForm(prev => ({ ...prev, title: filename }))
    }
  }, [form.title, onUploadError])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
    
    const files = Array.from(e.dataTransfer.files)
    if (files.length > 0) {
      handleFileSelect(files[0])
    }
  }, [handleFileSelect])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
  }, [])

  const handleFileInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      handleFileSelect(files[0])
    }
  }, [handleFileSelect])

  const handleUpload = useCallback(async () => {
    if (!form.file || !form.title.trim()) {
      onUploadError?.('请选择文件并填写标题')
      return
    }

    setUploadStatus('uploading')
    setUploadProgress({ loaded: 0, total: form.file.size, percentage: 0 })

    try {
      const response = await VideoService.uploadVideo(
        {
          title: form.title.trim(),
          description: form.description.trim(),
          file: form.file,
        },
        (progress) => {
          setUploadProgress(progress)
        }
      )

      setUploadStatus('success')
      onUploadSuccess?.(response.video_id)
      
      // 重置表单
      setForm({ title: '', description: '', file: null })
      setUploadProgress({ loaded: 0, total: 0, percentage: 0 })
    } catch (error) {
      setUploadStatus('error')
      onUploadError?.(error instanceof Error ? error.message : '上传失败')
    }
  }, [form, onUploadSuccess, onUploadError])

  const resetUpload = useCallback(() => {
    setForm({ title: '', description: '', file: null })
    setUploadStatus('idle')
    setUploadProgress({ loaded: 0, total: 0, percentage: 0 })
  }, [])

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  return (
    <div className="max-w-2xl mx-auto p-6 bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold text-gray-900 mb-6">上传视频</h2>
      
      {/* 文件拖拽区域 */}
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          dragActive
            ? 'border-blue-400 bg-blue-50'
            : form.file
            ? 'border-green-400 bg-green-50'
            : 'border-gray-300 hover:border-gray-400'
        }`}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
      >
        {form.file ? (
          <div className="space-y-2">
            <div className="text-lg font-medium text-green-600">✓ 文件已选择</div>
            <div className="text-gray-600">{form.file.name}</div>
            <div className="text-sm text-gray-500">{formatFileSize(form.file.size)}</div>
            <button
              type="button"
              onClick={resetUpload}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              重新选择
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="text-6xl text-gray-400">📁</div>
            <div>
              <p className="text-lg text-gray-600">拖拽视频文件到这里</p>
              <p className="text-sm text-gray-500">或者</p>
            </div>
            <label className="inline-block bg-blue-600 text-white px-6 py-2 rounded-md cursor-pointer hover:bg-blue-700">
              选择文件
              <input
                type="file"
                accept="video/*"
                onChange={handleFileInputChange}
                className="hidden"
              />
            </label>
            <p className="text-xs text-gray-500">
              支持格式：MP4, WebM, AVI, MOV (最大2GB)
            </p>
          </div>
        )}
      </div>

      {/* 表单输入 */}
      <div className="mt-6 space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            视频标题 *
          </label>
          <input
            type="text"
            value={form.title}
            onChange={(e) => handleInputChange('title', e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请输入视频标题"
            disabled={uploadStatus === 'uploading'}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            视频描述
          </label>
          <textarea
            value={form.description}
            onChange={(e) => handleInputChange('description', e.target.value)}
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            placeholder="请输入视频描述（可选）"
            disabled={uploadStatus === 'uploading'}
          />
        </div>
      </div>

      {/* 上传进度 */}
      {uploadStatus === 'uploading' && (
        <div className="mt-6">
          <div className="flex justify-between text-sm text-gray-600 mb-1">
            <span>上传进度</span>
            <span>{uploadProgress.percentage}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress.percentage}%` }}
            />
          </div>
          <div className="text-xs text-gray-500 mt-1">
            {formatFileSize(uploadProgress.loaded)} / {formatFileSize(uploadProgress.total)}
          </div>
        </div>
      )}

      {/* 上传按钮 */}
      <div className="mt-6">
        <button
          onClick={handleUpload}
          disabled={!form.file || !form.title.trim() || uploadStatus === 'uploading'}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
        >
          {uploadStatus === 'uploading' ? '上传中...' : '开始上传'}
        </button>
      </div>

      {/* 状态消息 */}
      {uploadStatus === 'success' && (
        <div className="mt-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
          ✓ 视频上传成功！
        </div>
      )}
    </div>
  )
}