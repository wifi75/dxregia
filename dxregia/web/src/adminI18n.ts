// Traduzione del pannello sysop. I testi del pannello sono scritti in italiano; qui si sostituiscono
// con la lingua scelta cercandoli nel dizionario di adminTranslations.ts (testo italiano → traduzione).
// Funziona sui nodi di testo e su title, placeholder, aria-label e alt, anche quando Svelte li aggiorna.
import { adminLanguages, adminRows } from './adminTranslations'

type Code = typeof adminLanguages[number]

const normalize = (text: string) => text.replace(/\s+/g, ' ').trim()
const escape = (text: string) => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const byKey = new Map<string, string[]>()
const patterns: { re: RegExp; row: string[] }[] = []
for (const row of adminRows) {
  const key = normalize(row[0])
  byKey.set(key, row)
  if (key.includes('{}')) patterns.push({ re: new RegExp('^' + key.split('{}').map(escape).join('(.+?)') + '$'), row })
}

let current: Code = 'it'
const index = (code: Code) => Math.max(0, adminLanguages.indexOf(code))

function fill(template: string, values: string[]) {
  let next = 0
  return template.replace(/\{\}/g, () => values[next++] ?? '')
}

export function translate(text: string, code: Code = current): string {
  if (code === 'it') return text
  const key = normalize(text)
  if (!key) return text
  let row = byKey.get(key)
  let groups: string[] = []
  if (!row) {
    for (const pattern of patterns) {
      const match = pattern.re.exec(key)
      if (match) { row = pattern.row; groups = match.slice(1); break }
    }
  }
  if (!row) return text
  const lead = /^\s*/.exec(text)![0]
  const trail = /\s*$/.exec(text)![0]
  return lead + fill(row[index(code)] || row[0], groups.map(group => translate(group, code))) + trail
}

// Per i testi costruiti dal codice (conferme): tr('Eliminare il filtro {}?', nome).
export function tr(template: string, ...values: (string | number)[]): string {
  const row = byKey.get(normalize(template))
  const text = row && current !== 'it' ? row[index(current)] || template : template
  return fill(text, values.map(String))
}

type Seen = { orig: string; shown: string }
const textSeen = new WeakMap<Node, Seen>()
const attrSeen = new WeakMap<Element, Record<string, Seen>>()
const attributes = ['title', 'placeholder', 'aria-label', 'alt']
const skipSelector = 'pre, code, textarea, script, style, [data-no-translate]'

function applyText(node: Text) {
  const parent = node.parentElement
  if (!parent || parent.closest(skipSelector)) return
  const seen = textSeen.get(node)
  const orig = seen && node.data === seen.shown ? seen.orig : node.data
  const shown = translate(orig)
  textSeen.set(node, { orig, shown })
  if (node.data !== shown) node.data = shown
}

function applyAttribute(element: Element, name: string) {
  const value = element.getAttribute(name)
  if (value === null || element.closest(skipSelector)) return
  const record = attrSeen.get(element) ?? {}
  const seen = record[name]
  const orig = seen && value === seen.shown ? seen.orig : value
  const shown = translate(orig)
  record[name] = { orig, shown }
  attrSeen.set(element, record)
  if (value !== shown) element.setAttribute(name, shown)
}

function applyTree(root: Node) {
  if (root.nodeType === Node.TEXT_NODE) { applyText(root as Text); return }
  if (root.nodeType !== Node.ELEMENT_NODE) return
  const element = root as Element
  for (const name of attributes) applyAttribute(element, name)
  const walker = document.createTreeWalker(element, NodeFilter.SHOW_TEXT)
  for (let node = walker.nextNode(); node; node = walker.nextNode()) applyText(node as Text)
  element.querySelectorAll('[title], [placeholder], [aria-label], [alt]').forEach(child => {
    for (const name of attributes) applyAttribute(child, name)
  })
}

export function startAdminTranslation(root: HTMLElement) {
  const observer = new MutationObserver(records => {
    for (const record of records) {
      if (record.type === 'characterData') applyText(record.target as Text)
      else if (record.type === 'attributes') applyAttribute(record.target as Element, record.attributeName!)
      else record.addedNodes.forEach(applyTree)
    }
  })
  observer.observe(root, { subtree: true, childList: true, characterData: true, attributes: true, attributeFilter: attributes })
  applyTree(root)
  return {
    setLanguage(code: string) {
      current = (adminLanguages as readonly string[]).includes(code) ? code as Code : 'en'
      applyTree(root)
    },
    stop() { observer.disconnect() },
  }
}
