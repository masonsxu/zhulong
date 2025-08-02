import { useNavigate } from 'react-router-dom'
import VideoUpload from '../components/VideoUpload'

export default function VideoUploadPage() {
  const navigate = useNavigate()

  const handleUploadSuccess = (videoId: string) => {
    // 上传成功后跳转到视频详情页
    navigate(`/video/${videoId}`)
  }

  const handleUploadError = (error: string) => {
    // 可以在这里显示错误提示
    console.error('Upload error:', error)
    // TODO: 实现toast提示
  }

  return (
    <div className="px-4 py-6">
      <VideoUpload 
        onUploadSuccess={handleUploadSuccess}
        onUploadError={handleUploadError}
      />
    </div>
  )
}