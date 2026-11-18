// Suono per gli spot nuovi, condiviso tra l'intestazione (interruttore) e la pagina degli spot.
const key = 'dxregia-sound'

function preference(): boolean {
  try { return localStorage.getItem(key) !== 'off' } catch { return true }
}

let on = $state(preference())
let audio: AudioContext | undefined

export const soundOn = () => on

// I browser permettono l'audio solo dopo un gesto dell'utente: il primo tocco o tasto lo sblocca.
export function unlockAudio() {
  try {
    audio ??= new AudioContext()
    if (audio.state === 'suspended') void audio.resume()
  } catch { /* senza audio il resto funziona uguale */ }
}

// Un breve «bip» morbido: due note ravvicinate, volume basso.
export function chirp(at = 0) {
  if (!audio || audio.state !== 'running') return
  const start = audio.currentTime + at
  const gain = audio.createGain()
  gain.connect(audio.destination)
  gain.gain.setValueAtTime(0.0001, start)
  gain.gain.exponentialRampToValueAtTime(0.09, start + 0.015)
  gain.gain.exponentialRampToValueAtTime(0.0001, start + 0.2)
  const tone = audio.createOscillator()
  tone.type = 'sine'
  tone.frequency.setValueAtTime(880, start)
  tone.frequency.exponentialRampToValueAtTime(1320, start + 0.09)
  tone.connect(gain)
  tone.start(start)
  tone.stop(start + 0.22)
}

export function toggleSound() {
  on = !on
  try { localStorage.setItem(key, on ? 'on' : 'off') } catch { /* solo una comodità */ }
  if (on) { unlockAudio(); chirp() }
}
