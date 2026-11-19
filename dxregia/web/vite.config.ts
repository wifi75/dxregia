import { execFileSync } from 'node:child_process'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'

let revision = ''
try {
  revision = execFileSync('git', ['rev-parse', '--short=12', 'HEAD'], {
    cwd: new URL('.', import.meta.url),
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'ignore'],
  }).trim()
} catch {
  // Source archives and container builds can lack Git metadata.
}

export default defineConfig({
  plugins: [svelte()],
  define: { __DXREGIA_REVISION__: JSON.stringify(revision) },
})
