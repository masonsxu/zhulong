import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import HomePage from './pages/HomePage'
import VideoUploadPage from './pages/VideoUploadPage'
import VideoDetailPage from './pages/VideoDetailPage'

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/upload" element={<VideoUploadPage />} />
          <Route path="/video/:id" element={<VideoDetailPage />} />
        </Routes>
      </Layout>
    </Router>
  )
}

export default App
