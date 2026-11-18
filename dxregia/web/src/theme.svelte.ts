// Tema chiaro o scuro: la scelta resta nel browser; senza scelta si segue quello del sistema.
const key = 'dxregia-theme'

function initial(): 'light' | 'dark' {
  try {
    const saved = localStorage.getItem(key)
    if (saved === 'light' || saved === 'dark') return saved
  } catch { /* si usa il tema del sistema */ }
  return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

let theme = $state<'light' | 'dark'>(initial())
document.documentElement.dataset.theme = theme

export const currentTheme = () => theme

export function toggleTheme() {
  theme = theme === 'dark' ? 'light' : 'dark'
  document.documentElement.dataset.theme = theme
  try { localStorage.setItem(key, theme) } catch { /* solo una comodità */ }
}
