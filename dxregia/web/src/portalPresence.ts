import { onMount } from 'svelte'

export function trackPortalPresence() {
  onMount(() => {
    if (!globalThis.crypto?.getRandomValues) return
    const id = Array.from(crypto.getRandomValues(new Uint8Array(16)), byte =>
      byte.toString(16).padStart(2, '0')).join('')
    function pulse() {
      if (document.visibilityState !== 'visible') return
      void fetch('/api/visits/heartbeat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(id),
      }).catch(() => {})
    }
    pulse()
    const timer = setInterval(pulse, 30000)
    document.addEventListener('visibilitychange', pulse)
    return () => {
      clearInterval(timer)
      document.removeEventListener('visibilitychange', pulse)
    }
  })
}
