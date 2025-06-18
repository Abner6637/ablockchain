import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,          // 让服务器监听所有地址（0.0.0.0），可远程访问
    port: 5173,          // 你想改成的端口，比如 5173
    strictPort: true     // 如果端口被占用则报错，不自动切换
  }
})
