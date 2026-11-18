import { locales, languages, type Key } from './translations'

export { languages }
const storageKey = 'dxregia-lang'

function initial(): string {
  try {
    const saved = localStorage.getItem(storageKey)
    if (saved && saved in locales) return saved
  } catch { /* archivio non disponibile: si usa la lingua del browser */ }
  for (const candidate of navigator.languages ?? [navigator.language]) {
    const code = candidate.toLowerCase().split('-')[0]
    if (code in locales) return code
  }
  return 'en' // ogni altra lingua (ad esempio il cinese) ripiega sull'inglese
}

let current = $state(initial())
if (typeof document !== 'undefined') document.documentElement.lang = current

export function lang() { return current }

export function setLang(code: string) {
  if (!(code in locales)) return
  current = code
  document.documentElement.lang = code
  try { localStorage.setItem(storageKey, code) } catch { /* solo una comodità */ }
}

export function t(key: Key, values?: Record<string, string | number>): string {
  let text = locales[current]?.[key] ?? locales.en[key]
  for (const [name, value] of Object.entries(values ?? {})) text = text.replaceAll('{' + name + '}', String(value))
  return text
}

export const locale = () => ({ it: 'it-IT', en: 'en-GB', es: 'es-ES', de: 'de-DE', fr: 'fr-FR', pt: 'pt-PT' })[current] ?? 'en-GB'
