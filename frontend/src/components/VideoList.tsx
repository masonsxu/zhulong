import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Play, Clock, HardDrive, Calendar, Search } from 'lucide-react'
import { Link } from 'react-router-dom'

interface Video {
  id: string
  title: string
  duration: number
  size: number
  uploadDate: string
  thumbnail?: string
}

const VideoList = () => {
  const [videos, setVideos] = useState<Video[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')

  useEffect(() => {
    // 模拟API调用
    const fetchVideos = async () => {
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      const mockVideos: Video[] = [
        {
          id: '1',
          title: '科幻电影预告片',
          duration: 180,
          size: 51200000,
          uploadDate: '2024-01-15',
        },
        {
          id: '2',
          title: '技术演示视频',
          duration: 300,
          size: 102400000,
          uploadDate: '2024-01-14',
        },
        {
          id: '3',
          title: '游戏剪辑合集',
          duration: 600,
          size: 256000000,
          uploadDate: '2024-01-13',
        },
      ]
      
      setVideos(mockVideos)
      setLoading(false)
    }

    fetchVideos()
  }, [])

  const filteredVideos = videos.filter(video =>
    video.title.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const formatSize = (bytes: number) => {
    const mb = bytes / (1024 * 1024)
    return `${mb.toFixed(1)} MB`
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-cyber-blue border-t-transparent rounded-full animate-spin mb-4"></div>
          <p className="neon-text">加载视频数据...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="text-center"
      >
        <h2 className="text-4xl font-bold mb-4 neon-text">视频库</h2>
        <p className="text-gray-400 mb-8">探索科幻视频世界</p>
      </motion.div>

      <div className="flex items-center justify-between mb-8">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            placeholder="搜索视频..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="cyber-input pl-10"
          />
        </div>
        <div className="flex items-center space-x-4 text-sm text-gray-400">
          <span>共 {filteredVideos.length} 个视频</span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredVideos.map((video, index) => (
          <motion.div
            key={video.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: index * 0.1 }}
            className="cyber-card group cursor-pointer hover:transform hover:scale-105 transition-all duration-300"
          >
            <div className="aspect-video bg-gradient-to-br from-cyber-blue/20 to-cyber-purple/20 rounded-lg mb-4 flex items-center justify-center relative overflow-hidden">
              <div className="absolute inset-0 bg-black/50 group-hover:bg-black/30 transition-all duration-300"></div>
              <Play className="w-16 h-16 text-cyber-blue z-10 opacity-80 group-hover:opacity-100 transition-all duration-300" />
              <div className="absolute top-2 right-2 bg-black/70 px-2 py-1 rounded text-xs text-cyber-green">
                {formatDuration(video.duration)}
              </div>
            </div>
            
            <div className="space-y-3">
              <h3 className="text-lg font-semibold text-white group-hover:text-cyber-blue transition-colors duration-300">
                {video.title}
              </h3>
              
              <div className="flex items-center justify-between text-sm text-gray-400">
                <div className="flex items-center space-x-1">
                  <HardDrive className="w-4 h-4" />
                  <span>{formatSize(video.size)}</span>
                </div>
                <div className="flex items-center space-x-1">
                  <Calendar className="w-4 h-4" />
                  <span>{video.uploadDate}</span>
                </div>
              </div>
              
              <Link
                to={`/video/${video.id}`}
                className="cyber-button w-full text-center block"
              >
                播放视频
              </Link>
            </div>
          </motion.div>
        ))}
      </div>

      {filteredVideos.length === 0 && (
        <div className="text-center py-12">
          <div className="text-6xl mb-4">📺</div>
          <h3 className="text-xl font-semibold text-gray-400 mb-2">没有找到视频</h3>
          <p className="text-gray-500">尝试不同的搜索词或上传新视频</p>
        </div>
      )}
    </div>
  )
}

export default VideoList