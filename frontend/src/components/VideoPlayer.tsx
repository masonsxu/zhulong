import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Play, Pause, Volume2, VolumeX, Maximize, Minimize, SkipBack, SkipForward, Settings } from 'lucide-react'

const VideoPlayer = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const videoRef = useRef<HTMLVideoElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [volume, setVolume] = useState(1)
  const [isMuted, setIsMuted] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // 模拟视频加载
    const timer = setTimeout(() => {
      setIsLoading(false)
    }, 2000)

    return () => clearTimeout(timer)
  }, [id])

  const formatTime = (time: number) => {
    const minutes = Math.floor(time / 60)
    const seconds = Math.floor(time % 60)
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  const togglePlay = () => {
    if (videoRef.current) {
      if (isPlaying) {
        videoRef.current.pause()
      } else {
        videoRef.current.play()
      }
      setIsPlaying(!isPlaying)
    }
  }

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newVolume = parseFloat(e.target.value)
    setVolume(newVolume)
    if (videoRef.current) {
      videoRef.current.volume = newVolume
    }
    setIsMuted(newVolume === 0)
  }

  const toggleMute = () => {
    if (videoRef.current) {
      videoRef.current.muted = !isMuted
      setIsMuted(!isMuted)
    }
  }

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen()
      setIsFullscreen(true)
    } else {
      document.exitFullscreen()
      setIsFullscreen(false)
    }
  }

  const handleTimeUpdate = () => {
    if (videoRef.current) {
      setCurrentTime(videoRef.current.currentTime)
    }
  }

  const handleLoadedMetadata = () => {
    if (videoRef.current) {
      setDuration(videoRef.current.duration)
    }
  }

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value)
    setCurrentTime(time)
    if (videoRef.current) {
      videoRef.current.currentTime = time
    }
  }

  const progress = duration > 0 ? (currentTime / duration) * 100 : 0

  return (
    <div className="min-h-screen py-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="max-w-6xl mx-auto px-4"
      >
        <div className="mb-6">
          <button
            onClick={() => navigate('/')}
            className="cyber-button mb-4"
          >
            ← 返回列表
          </button>
          <h1 className="text-3xl font-bold neon-text mb-2">视频播放器</h1>
          <p className="text-gray-400">ID: {id}</p>
        </div>

        <div className="cyber-card overflow-hidden">
          <div className="relative aspect-video bg-black rounded-lg overflow-hidden">
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <div className="w-16 h-16 border-4 border-cyber-blue border-t-transparent rounded-full animate-spin mb-4"></div>
                  <p className="neon-text">加载视频中...</p>
                </div>
              </div>
            ) : (
              <>
                <video
                  ref={videoRef}
                  className="w-full h-full object-cover"
                  onTimeUpdate={handleTimeUpdate}
                  onLoadedMetadata={handleLoadedMetadata}
                  onPlay={() => setIsPlaying(true)}
                  onPause={() => setIsPlaying(false)}
                  onEnded={() => setIsPlaying(false)}
                >
                  <source src="/sample-video.mp4" type="video/mp4" />
                  您的浏览器不支持视频播放
                </video>

                <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 hover:opacity-100 transition-opacity duration-300"></div>
              </>
            )}
          </div>

          <div className="p-6 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-xl font-semibold text-white mb-2">科幻电影预告片</h2>
                <div className="flex items-center space-x-4 text-sm text-gray-400">
                  <span>时长: {formatTime(duration)}</span>
                  <span>分辨率: 1920x1080</span>
                  <span>格式: MP4</span>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <span className="cyber-badge">HD</span>
                <span className="cyber-badge">科幻</span>
              </div>
            </div>

            <div className="space-y-3">
              <div className="cyber-progress">
                <div
                  className="cyber-progress-bar"
                  style={{ width: `${progress}%` }}
                ></div>
              </div>

              <div className="flex items-center justify-between text-sm text-gray-400">
                <span>{formatTime(currentTime)}</span>
                <span>{formatTime(duration)}</span>
              </div>

              <div className="flex items-center justify-center space-x-4">
                <button className="p-2 rounded-lg hover:bg-dark-card transition-colors">
                  <SkipBack className="w-5 h-5 text-gray-400" />
                </button>

                <button
                  onClick={togglePlay}
                  className="p-3 rounded-full bg-gradient-to-r from-cyber-blue to-cyber-purple hover:from-cyber-purple hover:to-cyber-pink transition-all duration-300"
                >
                  {isPlaying ? (
                    <Pause className="w-6 h-6 text-white" />
                  ) : (
                    <Play className="w-6 h-6 text-white" />
                  )}
                </button>

                <button className="p-2 rounded-lg hover:bg-dark-card transition-colors">
                  <SkipForward className="w-5 h-5 text-gray-400" />
                </button>

                <div className="flex items-center space-x-2">
                  <button onClick={toggleMute} className="p-2 rounded-lg hover:bg-dark-card transition-colors">
                    {isMuted ? (
                      <VolumeX className="w-5 h-5 text-gray-400" />
                    ) : (
                      <Volume2 className="w-5 h-5 text-gray-400" />
                    )}
                  </button>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={volume}
                    onChange={handleVolumeChange}
                    className="w-20 h-2 bg-dark-bg rounded-lg appearance-none cursor-pointer"
                  />
                </div>

                <button className="p-2 rounded-lg hover:bg-dark-card transition-colors">
                  <Settings className="w-5 h-5 text-gray-400" />
                </button>

                <button
                  onClick={toggleFullscreen}
                  className="p-2 rounded-lg hover:bg-dark-card transition-colors"
                >
                  {isFullscreen ? (
                    <Minimize className="w-5 h-5 text-gray-400" />
                  ) : (
                    <Maximize className="w-5 h-5 text-gray-400" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-6">
          <div className="cyber-card">
            <h3 className="text-lg font-semibold text-white mb-4">视频信息</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-400">文件大小:</span>
                <span className="text-white">512 MB</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">上传时间:</span>
                <span className="text-white">2024-01-15</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">播放次数:</span>
                <span className="text-white">1,234</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">比特率:</span>
                <span className="text-white">4000 kbps</span>
              </div>
            </div>
          </div>

          <div className="cyber-card">
            <h3 className="text-lg font-semibold text-white mb-4">系统状态</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-400">网络状态:</span>
                <span className="text-cyber-green">良好</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">缓冲状态:</span>
                <span className="text-cyber-blue">已缓存</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">播放质量:</span>
                <span className="text-cyber-purple">1080p</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">延迟:</span>
                <span className="text-cyber-yellow">12ms</span>
              </div>
            </div>
          </div>
        </div>
      </motion.div>
    </div>
  )
}

export default VideoPlayer