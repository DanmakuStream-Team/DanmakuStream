/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_DEV_GATEWAY_TARGET?: string
  readonly VITE_DEV_GATEWAY_WS_TARGET?: string
  readonly VITE_RECOMMEND_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
