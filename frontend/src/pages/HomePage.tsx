import { useState } from 'react'
import VideoList from '../components/VideoList'
import type { Video } from '../types'

export default function HomePage() {
  const [searchQuery, setSearchQuery] = useState('')

  const handleVideoSelect = (video: Video) => {
    // 可以在这里添加选择视频的逻辑，比如打开模态框或跳转到详情页
    console.log('Selected video:', video)
  }

  return (
    <div className="px-4 py-6">
      {/* 页面标题和搜索 */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">视频库</h1>
        <div className="max-w-md">
          <div className="relative">
            <input
              type="text"
              placeholder="搜索视频..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center">
              <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      {/* 视频列表 */}
      <VideoList 
        searchQuery={searchQuery}
        onVideoSelect={handleVideoSelect}
      />
    </div>
  )
}