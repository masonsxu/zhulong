import { useState, useRef, useEffect, useCallback } from 'react'
import type { Video, PlayerState } from '../../types'

interface VideoPlayerProps {
  video: Video
  autoPlay?: boolean
  onTimeUpdate?: (currentTime: number) => void
  onEnded?: () => void
  onError?: (error: string) => void
}

export default function VideoPlayer({ 
  video, 
  autoPlay = false,
  onTimeUpdate,
  onEnded,
  onError 
}: VideoPlayerProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [playerState, setPlayerState] = useState<PlayerState>({
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
    isMuted: false,
    isFullscreen: false,
    isLoading: true,
  })
  const [showControls, setShowControls] = useState(true)
  const [controlsTimer, setControlsTimer] = useState<NodeJS.Timeout | null>(null)
  const [playbackRate, setPlaybackRate] = useState(1)
  const [showSpeedMenu, setShowSpeedMenu] = useState(false)

  // 播放/暂停
  const togglePlay = useCallback(async () => {
    if (!videoRef.current) return

    try {
      if (playerState.isPlaying) {
        await videoRef.current.pause()
      } else {
        await videoRef.current.play()
      }
    } catch (error) {
      onError?.('播放失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }, [playerState.isPlaying, onError])

  // 设置音量
  const setVolume = useCallback((volume: number) => {
    if (!videoRef.current) return
    
    videoRef.current.volume = volume
    setPlayerState(prev => ({ ...prev, volume, isMuted: volume === 0 }))
  }, [])

  // 静音/取消静音
  const toggleMute = useCallback(() => {
    if (!videoRef.current) return

    const newMuted = !playerState.isMuted
    videoRef.current.muted = newMuted
    setPlayerState(prev => ({ ...prev, isMuted: newMuted }))
  }, [playerState.isMuted])

  // 跳转到指定时间
  const seekTo = useCallback((time: number) => {
    if (!videoRef.current) return
    
    videoRef.current.currentTime = time
    setPlayerState(prev => ({ ...prev, currentTime: time }))
  }, [])

  // 设置播放速度
  const setPlaybackSpeed = useCallback((rate: number) => {
    if (!videoRef.current) return
    
    videoRef.current.playbackRate = rate
    setPlaybackRate(rate)
    setShowSpeedMenu(false)
  }, [])

  // 全屏/退出全屏
  const toggleFullscreen = useCallback(async () => {
    if (!videoRef.current) return

    try {
      if (!document.fullscreenElement) {
        await videoRef.current.requestFullscreen()
        setPlayerState(prev => ({ ...prev, isFullscreen: true }))
      } else {
        await document.exitFullscreen()
        setPlayerState(prev => ({ ...prev, isFullscreen: false }))
      }
    } catch (error) {
      onError?.('全屏操作失败: ' + (error instanceof Error ? error.message : '未知错误'))
    }
  }, [onError])

  // 显示控制条
  const showControlsTemporarily = useCallback(() => {
    setShowControls(true)
    
    if (controlsTimer) {
      clearTimeout(controlsTimer)
    }
    
    const timer = setTimeout(() => {
      if (playerState.isPlaying) {
        setShowControls(false)
      }
    }, 3000)
    
    setControlsTimer(timer)
  }, [controlsTimer, playerState.isPlaying])

  // 键盘控制
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!videoRef.current) return

    switch (e.code) {
      case 'Space':
        e.preventDefault()
        togglePlay()
        break
      case 'ArrowLeft':
        e.preventDefault()
        seekTo(Math.max(0, playerState.currentTime - 10))
        break
      case 'ArrowRight':
        e.preventDefault()
        seekTo(Math.min(playerState.duration, playerState.currentTime + 10))
        break
      case 'ArrowUp':
        e.preventDefault()
        setVolume(Math.min(1, playerState.volume + 0.1))
        break
      case 'ArrowDown':
        e.preventDefault()
        setVolume(Math.max(0, playerState.volume - 0.1))
        break
      case 'KeyM':
        e.preventDefault()
        toggleMute()
        break
      case 'KeyF':
        e.preventDefault()
        toggleFullscreen()
        break
    }
  }, [togglePlay, seekTo, setVolume, toggleMute, toggleFullscreen, playerState])

  // 格式化时间
  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const remainingSeconds = Math.floor(seconds % 60)

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`
    }
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
  }

  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    const handleLoadedData = () => {
      setPlayerState(prev => ({ 
        ...prev, 
        duration: video.duration,
        isLoading: false 
      }))
    }

    const handleTimeUpdate = () => {
      const currentTime = video.currentTime
      setPlayerState(prev => ({ ...prev, currentTime }))
      onTimeUpdate?.(currentTime)
    }

    const handlePlay = () => {
      setPlayerState(prev => ({ ...prev, isPlaying: true }))
    }

    const handlePause = () => {
      setPlayerState(prev => ({ ...prev, isPlaying: false }))
    }

    const handleEnded = () => {
      setPlayerState(prev => ({ ...prev, isPlaying: false }))
      onEnded?.()
    }

    const handleError = () => {
      setPlayerState(prev => ({ ...prev, isLoading: false }))
      onError?.('视频加载失败')
    }

    const handleVolumeChange = () => {
      setPlayerState(prev => ({ 
        ...prev, 
        volume: video.volume,
        isMuted: video.muted 
      }))
    }

    video.addEventListener('loadeddata', handleLoadedData)
    video.addEventListener('timeupdate', handleTimeUpdate)
    video.addEventListener('play', handlePlay)
    video.addEventListener('pause', handlePause)
    video.addEventListener('ended', handleEnded)
    video.addEventListener('error', handleError)
    video.addEventListener('volumechange', handleVolumeChange)

    return () => {
      video.removeEventListener('loadeddata', handleLoadedData)
      video.removeEventListener('timeupdate', handleTimeUpdate)
      video.removeEventListener('play', handlePlay)
      video.removeEventListener('pause', handlePause)
      video.removeEventListener('ended', handleEnded)
      video.removeEventListener('error', handleError)
      video.removeEventListener('volumechange', handleVolumeChange)
    }
  }, [onTimeUpdate, onEnded, onError])

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  useEffect(() => {
    return () => {
      if (controlsTimer) {
        clearTimeout(controlsTimer)
      }
    }
  }, [controlsTimer])

  return (
    <div 
      className="relative bg-black rounded-lg overflow-hidden group"
      onMouseMove={showControlsTemporarily}
      onMouseLeave={() => playerState.isPlaying && setShowControls(false)}
    >
      {/* 视频元素 */}
      <video
        ref={videoRef}
        src={video.play_url}
        className="w-full h-auto"
        autoPlay={autoPlay}
        onClick={togglePlay}
        poster={video.thumbnail_url}
      />

      {/* 加载指示器 */}
      {playerState.isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-50">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white"></div>
        </div>
      )}

      {/* 播放按钮覆盖层 */}
      {!playerState.isPlaying && !playerState.isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-30">
          <button
            onClick={togglePlay}
            className="bg-white bg-opacity-90 rounded-full p-4 hover:bg-opacity-100 transition-all"
          >
            <svg className="w-8 h-8 text-black" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      )}

      {/* 控制条 */}
      <div 
        className={`absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black to-transparent p-4 transition-opacity duration-300 ${
          showControls ? 'opacity-100' : 'opacity-0'
        }`}
      >
        {/* 进度条 */}
        <div className="mb-4">
          <div className="relative">
            <input
              type="range"
              min={0}
              max={playerState.duration || 0}
              value={playerState.currentTime}
              onChange={(e) => seekTo(Number(e.target.value))}
              className="w-full h-1 bg-gray-600 rounded-lg appearance-none cursor-pointer slider"
              style={{
                background: `linear-gradient(to right, #3b82f6 0%, #3b82f6 ${(playerState.currentTime / (playerState.duration || 1)) * 100}%, #4b5563 ${(playerState.currentTime / (playerState.duration || 1)) * 100}%, #4b5563 100%)`
              }}
            />
          </div>
        </div>

        {/* 控制按钮 */}
        <div className="flex items-center justify-between text-white">
          <div className="flex items-center space-x-4">
            {/* 播放/暂停 */}
            <button onClick={togglePlay} className="hover:text-gray-300">
              {playerState.isPlaying ? (
                <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              ) : (
                <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clipRule="evenodd" />
                </svg>
              )}
            </button>

            {/* 音量控制 */}
            <div className="flex items-center space-x-2">
              <button onClick={toggleMute} className="hover:text-gray-300">
                {playerState.isMuted || playerState.volume === 0 ? (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M9.383 3.076A1 1 0 0110 4v12a1 1 0 01-1.707.707L4.586 13H2a1 1 0 01-1-1V8a1 1 0 011-1h2.586l3.707-3.707a1 1 0 011.09-.217zM12.293 7.293a1 1 0 011.414 0L15 8.586l1.293-1.293a1 1 0 111.414 1.414L16.414 10l1.293 1.293a1 1 0 01-1.414 1.414L15 11.414l-1.293 1.293a1 1 0 01-1.414-1.414L13.586 10l-1.293-1.293a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M9.383 3.076A1 1 0 0110 4v12a1 1 0 01-1.707.707L4.586 13H2a1 1 0 01-1-1V8a1 1 0 011-1h2.586l3.707-3.707a1 1 0 011.09-.217zM12.146 5.146a.5.5 0 01.708 0c.647.647 1.146 1.428 1.146 2.354s-.5 1.707-1.146 2.354a.5.5 0 01-.708-.708c.647-.647.854-1.026.854-1.646s-.207-.999-.854-1.646a.5.5 0 010-.708z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
              <input
                type="range"
                min={0}
                max={1}
                step={0.1}
                value={playerState.volume}
                onChange={(e) => setVolume(Number(e.target.value))}
                className="w-20 h-1 bg-gray-600 rounded-lg appearance-none cursor-pointer"
              />
            </div>

            {/* 时间显示 */}
            <div className="text-sm">
              {formatTime(playerState.currentTime)} / {formatTime(playerState.duration)}
            </div>
          </div>

          <div className="flex items-center space-x-4">
            {/* 播放速度控制 */}
            <div className="relative">
              <button
                onClick={() => setShowSpeedMenu(!showSpeedMenu)}
                className="hover:text-gray-300 text-sm"
              >
                {playbackRate}x
              </button>
              {showSpeedMenu && (
                <div className="absolute bottom-8 left-0 bg-black bg-opacity-90 rounded-md p-2 min-w-16">
                  {[0.5, 0.75, 1, 1.25, 1.5, 2].map(rate => (
                    <button
                      key={rate}
                      onClick={() => setPlaybackSpeed(rate)}
                      className={`block w-full text-left px-2 py-1 text-sm hover:bg-gray-700 rounded ${
                        playbackRate === rate ? 'text-blue-400' : ''
                      }`}
                    >
                      {rate}x
                    </button>
                  ))}
                </div>
              )}
            </div>
            
            {/* 全屏按钮 */}
            <button onClick={toggleFullscreen} className="hover:text-gray-300">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M3 4a1 1 0 011-1h4a1 1 0 010 2H6.414l2.293 2.293a1 1 0 11-1.414 1.414L5 6.414V8a1 1 0 01-2 0V4zm9 1a1 1 0 010-2h4a1 1 0 011 1v4a1 1 0 01-2 0V6.414l-2.293 2.293a1 1 0 11-1.414-1.414L13.586 5H12zm-9 7a1 1 0 012 0v1.586l2.293-2.293a1 1 0 111.414 1.414L6.414 15H8a1 1 0 010 2H4a1 1 0 01-1-1v-4zm13-1a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 010-2h1.586l-2.293-2.293a1 1 0 111.414-1.414L15 13.586V12a1 1 0 011-1z" clipRule="evenodd" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* 键盘快捷键提示 */}
      <div className="absolute top-4 right-4 text-white text-xs bg-black bg-opacity-50 p-2 rounded opacity-0 group-hover:opacity-100 transition-opacity">
        <div>空格: 播放/暂停</div>
        <div>← →: 后退/前进10秒</div>
        <div>↑ ↓: 音量+/-</div>
        <div>M: 静音</div>
        <div>F: 全屏</div>
      </div>
    </div>
  )
}