/// <reference types="vite/client" />
interface ImportMetaEnv {
  readonly VITE_MODE: string
  readonly VITE_APP_API_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
