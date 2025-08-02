import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  
  // 路径别名配置
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@services': path.resolve(__dirname, './src/services'),
      '@types': path.resolve(__dirname, './src/types'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
    }
  },

  // 开发服务器配置
  server: {
    port: parseInt(process.env.VITE_DEV_SERVER_PORT || '5173'),
    host: process.env.VITE_DEV_SERVER_HOST || 'localhost',
    open: true,
    cors: true,
    // 代理配置，用于开发环境API请求
    proxy: {
      '/api': {
        target: 'http://localhost:8888',
        changeOrigin: true,
        secure: false,
      }
    }
  },

  // 构建配置
  build: {
    outDir: 'dist',
    sourcemap: true,
    // 分包策略
    rollupOptions: {
      output: {
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
        manualChunks: {
          // 将React相关库打包到单独的chunk
          react: ['react', 'react-dom'],
          // 将路由相关库打包到单独的chunk
          router: ['react-router-dom'],
          // 将HTTP客户端打包到单独的chunk
          http: ['axios'],
        }
      }
    },
    // 压缩配置
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
  },

  // 预览服务器配置
  preview: {
    port: 4173,
    host: 'localhost',
    open: true
  },

  // 环境变量配置
  envPrefix: 'VITE_',
})
