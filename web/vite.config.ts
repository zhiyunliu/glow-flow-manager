import { writeFileSync } from 'node:fs'
import { resolve } from 'node:path'
import vue from '@vitejs/plugin-vue'
import { defineConfig, type Plugin } from 'vite'

// 构建后保留 dist/readme.md：Go 通过 embed 嵌入 dist 目录，未构建前端时目录也必须存在
function keepDistReadme(): Plugin {
  return {
    name: 'keep-dist-readme',
    apply: 'build',
    closeBundle() {
      writeFileSync(resolve(import.meta.dirname, 'dist/readme.md'), '前端构建产物目录，由 web 下 npm run build 生成。\n')
    },
  }
}

// 开发时代理到本地后端，并模拟网关注入的操作人请求头
const backend = process.env.GFM_BACKEND || 'http://127.0.0.1:8190'

export default defineConfig(({ command }) => ({
  // 生产构建使用相对路径，页面挂载在 /api/manager-glow/web/ 或网关前缀下均可访问
  base: command === 'build' ? './' : '/',
  plugins: [vue(), keepDistReadme()],
  resolve: {
    alias: { '@': resolve(import.meta.dirname, 'src') },
  },
  server: {
    port: 5173,
    proxy: {
      '/api/manager-glow': {
        target: backend,
        changeOrigin: true,
        headers: { 'X-Auth-User-Id': 'dev' },
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
  },
}))
