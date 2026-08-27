import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    tailwindcss(),
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: false,
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: false,
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_DEV_GATEWAY_TARGET ?? 'http://localhost:8888',
        changeOrigin: true,
      },
      '/media': {
        target: process.env.VITE_DEV_GATEWAY_TARGET ?? 'http://localhost:8888',
        changeOrigin: true,
      },
      '/ws': {
        target: process.env.VITE_DEV_GATEWAY_WS_TARGET ?? 'ws://localhost:8888',
        ws: true,
      },
    },
  },
})
