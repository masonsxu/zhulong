import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { motion } from 'framer-motion'
import Header from './components/Header'
import VideoList from './components/VideoList'
import VideoPlayer from './components/VideoPlayer'
import VideoUpload from './components/VideoUpload'
import CyberBackground from './components/CyberBackground'

function App() {
  return (
    <Router>
      <div className="min-h-screen relative">
        <CyberBackground />
        
        <div className="relative z-10">
          <Header />
          
          <main className="container mx-auto px-4 py-8">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <Routes>
                <Route path="/" element={<VideoList />} />
                <Route path="/video/:id" element={<VideoPlayer />} />
                <Route path="/upload" element={<VideoUpload />} />
              </Routes>
            </motion.div>
          </main>
        </div>
      </div>
    </Router>
  )
}

export default App