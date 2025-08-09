import { Link, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Play, Upload, Home, Cpu } from 'lucide-react'

const Header = () => {
  const location = useLocation()

  const navItems = [
    { path: '/', label: '首页', icon: Home },
    { path: '/upload', label: '上传', icon: Upload },
  ]

  return (
    <header className="bg-dark-bg/80 backdrop-blur-md border-b border-dark-border sticky top-0 z-50">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
            className="flex items-center space-x-3"
          >
            <div className="relative">
              <Cpu className="w-8 h-8 text-cyber-blue" />
              <div className="absolute -top-1 -right-1 w-3 h-3 bg-cyber-green rounded-full animate-pulse"></div>
            </div>
            <div>
              <h1 className="text-2xl font-bold neon-text">烛龙</h1>
              <p className="text-xs text-gray-400">科幻视频播放系统</p>
            </div>
          </motion.div>

          <nav className="flex items-center space-x-6">
            {navItems.map((item, index) => {
              const Icon = item.icon
              const isActive = location.pathname === item.path

              return (
                <motion.div
                  key={item.path}
                  initial={{ opacity: 0, y: -20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.5, delay: index * 0.1 }}
                >
                  <Link
                    to={item.path}
                    className={`flex items-center space-x-2 px-4 py-2 rounded-lg transition-all duration-300 ${
                      isActive
                        ? 'bg-gradient-to-r from-cyber-blue/20 to-cyber-purple/20 border border-cyber-blue/50'
                        : 'hover:bg-dark-card'
                    }`}
                  >
                    <Icon className={`w-5 h-5 ${isActive ? 'text-cyber-blue' : 'text-gray-400'}`} />
                    <span className={isActive ? 'text-cyber-blue font-semibold' : 'text-gray-400'}>
                      {item.label}
                    </span>
                  </Link>
                </motion.div>
              )
            })}
          </nav>

          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5 }}
            className="flex items-center space-x-4"
          >
            <div className="data-stream">
              <div className="flex items-center space-x-2">
                <div className="w-2 h-2 bg-cyber-green rounded-full animate-pulse"></div>
                <span>系统在线</span>
              </div>
            </div>
            <div className="text-xs text-gray-500">
              <div>v1.0.0</div>
            </div>
          </motion.div>
        </div>
      </div>
    </header>
  )
}

export default Header