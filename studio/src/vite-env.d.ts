/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Base URL of the coursesmith-tutor service. Defaults to localhost:8765. */
  readonly VITE_TUTOR_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
