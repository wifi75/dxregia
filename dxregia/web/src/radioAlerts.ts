import { alertStorageKey, bandFor, groupOf, localDay, modeInfo, spotKey, type Spot } from './spotTools'

export type AlertRule = {
  id: string; label: string; calls: string; band: string; mode: string; frequency: string
  portable: boolean; newDXCC: boolean; expires: string; enabled: boolean; known: number[]
}
export type RadioAlert = {
  id: string; key: string; day: string; call: string; frequency: number; mode: string
  entity: string; rules: string[]; first: number; last: number; count: number; read: boolean
}
export type AlertBook = {
  rules: AlertRule[]; alerts: RadioAlert[]; recent: string[]; notifications: boolean
  quietStart: string; quietEnd: string; lastNotification?: number; days: Record<string, { groups: number; spots: number }>
}
export const emptyAlertBook = (): AlertBook => ({
  rules: [], alerts: [], recent: [], notifications: false, quietStart: '', quietEnd: '', days: {},
})
export const emptyRule = (): AlertRule => ({
  id: '', label: '', calls: '', band: '', mode: '', frequency: '',
  portable: false, newDXCC: false, expires: '', enabled: true, known: [],
})
const maxAlerts = 200
const groupMillis = 15 * 60 * 1000

export function inQuietHours(date: Date, start: string, end: string) {
  if (!start || !end || start === end) return false
  const current = [date.getHours(), date.getMinutes()].map(part => String(part).padStart(2, '0')).join(':')
  return start < end ? current >= start && current < end : current >= start || current < end
}

export function parseWatchCalls(value: string) {
  return [...new Set(value.toUpperCase().split(/[\s,;]+/).filter(Boolean))]
}

export function validRule(rule: AlertRule) {
  const frequency = Number(rule.frequency.replace(',', '.'))
  const calls = parseWatchCalls(rule.calls)
  return rule.label.trim().length > 0 && rule.label.length <= 60 &&
    calls.length <= 20 && calls.every(call => /^[A-Z0-9][A-Z0-9/-]{1,23}$/.test(call) && /\d/.test(call)) &&
    (!rule.frequency || Number.isFinite(frequency) && frequency > 0 && frequency <= 100_000_000) &&
    Boolean(calls.length || rule.band || rule.mode || rule.frequency || rule.portable || rule.newDXCC)
}

export function alertRuleMatches(rule: AlertRule, spot: Spot, mode: string, day: string) {
  if (!rule.enabled || (rule.expires && rule.expires < day)) return false
  if (rule.band && rule.band !== bandFor(spot.Frequency)) return false
  if (rule.mode && rule.mode !== mode && rule.mode !== groupOf(mode)) return false
  if (rule.frequency && Math.round(Number(rule.frequency) * 1000) !== Math.round(spot.Frequency * 1000)) return false
  const call = spot.Call.toUpperCase()
  const portable = /\/(?:P|M|MM|AM)(?:\/(?:QRP|[0-9]))?$/.test(call)
  if (rule.portable && !portable) return false
  const calls = parseWatchCalls(rule.calls)
  if (calls.length && !calls.some(watch => watch.includes('/') ? call === watch : call.split('/').includes(watch))) return false
  return !rule.newDXCC || spot.DXCC > 0 && spot.DXCC !== 666 && !rule.known.includes(spot.DXCC)
}

export function readAlertBook(): AlertBook {
  try {
    const json = localStorage.getItem(alertStorageKey)
    if (json) {
      const data = JSON.parse(json) as AlertBook
      if (Array.isArray(data.rules) && Array.isArray(data.alerts) && Array.isArray(data.recent) && data.days) return data
    }
  } catch { /* Browser storage can be unavailable. */ }
  return emptyAlertBook()
}

export function collectAlerts(book: AlertBook, spots: Spot[], now: Date) {
  const day = localDay(now)
  const ids = new Set(book.recent)
  const fresh = []
  const at = now.getTime()
  let updated = false

  for (const spot of spots) {
    const id = spotKey(spot)
    if (ids.has(id)) continue
    ids.add(id)
    const mode = modeInfo(spot).mode
    const rules = book.rules.filter(rule => alertRuleMatches(rule, spot, mode, day))
    if (!rules.length) continue
    for (const rule of rules) if (rule.newDXCC) rule.known.push(spot.DXCC)
    const labels = [...new Set(rules.map(rule => rule.label))]
    const key = [spot.Call.toUpperCase(), bandFor(spot.Frequency), mode].join('|')
    const match = book.alerts.find(alert => alert.key === key && alert.day === day && at - alert.first < groupMillis)
    const stats = book.days[day] ??= { groups: 0, spots: 0 }
    stats.spots++
    updated = true
    if (match) {
      match.count++
      match.last = at
      match.frequency = spot.Frequency
      match.rules = [...new Set([...match.rules, ...labels])]
    } else {
      const alert: RadioAlert = {
        id, key, day, call: spot.Call, frequency: spot.Frequency, mode, entity: spot.Entity ?? '',
        rules: labels, first: at, last: at, count: 1, read: false,
      }
      stats.groups++
      book.alerts.unshift(alert)
      fresh.push(alert)
    }
  }
  if (updated) {
    book.recent = [...ids].slice(-500)
    book.alerts = book.alerts.sort((a, b) => b.last - a.last).slice(0, maxAlerts)
    const days = Object.keys(book.days).sort().slice(-7)
    book.days = Object.fromEntries(days.map(date => [date, book.days[date]]))
  }
  return { fresh, updated }
}
