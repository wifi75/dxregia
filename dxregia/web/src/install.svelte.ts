// Installazione come app su smartphone. Chrome e i browser Android offrono un evento che si può
// mostrare con un pulsante; Safari su iPhone non ha l'evento, quindi si spiega il passaggio manuale.
type InstallPrompt = Event & { prompt: () => Promise<void>; userChoice: Promise<{ outcome: string }> }

let promptEvent = $state<InstallPrompt | undefined>()
let installed = $state(false)

const standalone = () =>
  typeof window !== 'undefined' &&
  (window.matchMedia('(display-mode: standalone)').matches || (navigator as unknown as { standalone?: boolean }).standalone === true)

if (typeof window !== 'undefined') {
  installed = standalone()
  window.addEventListener('beforeinstallprompt', event => {
    event.preventDefault()
    promptEvent = event as InstallPrompt
  })
  window.addEventListener('appinstalled', () => {
    installed = true
    promptEvent = undefined
  })
}

export const isIos = () =>
  typeof navigator !== 'undefined' && /iphone|ipad|ipod/i.test(navigator.userAgent) && !/crios|fxios/i.test(navigator.userAgent)

// Il pulsante serve se l'app non è già installata e il browser sa installarla o è Safari su iPhone.
export const canInstall = () => !installed && (promptEvent !== undefined || isIos())
export const needsManualInstall = () => promptEvent === undefined && isIos()

export async function installApp() {
  if (!promptEvent) return
  const event = promptEvent
  promptEvent = undefined
  await event.prompt()
  try { await event.userChoice } catch { /* scelta annullata */ }
}
