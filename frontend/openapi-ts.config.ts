import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: '../openapi.yaml',
  output: {
    path: 'src/api/generated',
    format: 'prettier',
  },
  plugins: ['@hey-api/client-axios', '@hey-api/typescript', '@hey-api/sdk'],
})
