export type Spot = {
  Frequency: number; Call: string; At: number; Comment: string; Spotter: string; DXCC: number
  Cursor?: string; Entity?: string; Continent?: string; SpotterContinent?: string; CQ?: number; ITU?: number
  Grids?: string[]; Special?: boolean
}

export const bands = [
  ['160 m', 1800, 2000], ['80 m', 3500, 4000], ['60 m', 5250, 5450],
  ['40 m', 7000, 7300], ['30 m', 10100, 10150], ['20 m', 14000, 14350],
  ['17 m', 18068, 18168], ['15 m', 21000, 21450], ['12 m', 24890, 24990],
  ['10 m', 28000, 29700], ['6 m', 50000, 54000], ['VHF+', 70000, Infinity],
] as const
export const bandFor = (frequency: number) => bands.find(([, low, high]) => frequency >= low && frequency <= high)?.[0] ?? ''
export const spotKey = (spot: Spot) => [spot.At, spot.Call, spot.Frequency, spot.Spotter, spot.Comment].join('|')

type SearchField = 'dx' | 'de' | 'comment'
type SearchTerm = { word: string; exclude: boolean; field?: SearchField }

function searchGroups(query: string, defaultField?: SearchField) {
  const groups: SearchTerm[][] = [[]]
  for (const match of query.matchAll(/(-)?(?:(dx|de|comment|commento):)?(?:"([^"]+)"|([^"\s]+))/gi)) {
    const [, exclude, field, phrase, token] = match
    if (!exclude && !field && token?.toUpperCase() === 'OR' && groups.at(-1)!.length) {
      groups.push([])
    } else {
      const normalized = field?.toLowerCase().replace('commento', 'comment') as SearchField | undefined
      groups.at(-1)!.push({ word: phrase ?? token, exclude: Boolean(exclude), field: normalized ?? defaultField })
    }
  }
  return groups.filter(group => group.length)
}

export function spotMatcher(query: string) {
  const groups = searchGroups(query)
  return (spot: Spot) => {
    if (!groups.length) return true
    const values = { dx: spot.Call.toLowerCase(), de: spot.Spotter.toLowerCase(), comment: spot.Comment.toLowerCase() }
    const all = Object.values(values)
    return groups.some(group => group.every(term => {
      const field = term.field ? [values[term.field]] : all
      return field.some(value => value.includes(term.word.toLowerCase())) !== term.exclude
    }))
  }
}

export function migrateSpotSearch(query: string, comment: string) {
  if (!comment.trim()) return query
  const plain = query.trim().replaceAll('"', '')
  return searchGroups(comment, 'comment').map(group => [
    ...(plain ? ['"' + plain + '"'] : []),
    ...group.map(term => (term.exclude ? '-' : '') + term.field + ':"' + term.word + '"'),
  ].join(' ')).join(' OR ')
}

export function corroboratedSpots(rows: Spot[]) {
  const confirmed = new Set<string>()
  const byCall = new Map<string, Spot[]>()
  for (const spot of [...rows].sort((a, b) => a.At - b.At)) {
    const key = spot.Call.toUpperCase()
    const reporters = byCall.get(key) ?? []
    const nearby = reporters.filter(previous => spot.At - previous.At <= 180)
    for (const previous of nearby) {
      if (spot.Spotter && previous.Spotter && spot.Spotter.toUpperCase() !== previous.Spotter.toUpperCase() &&
          Math.abs(spot.Frequency - previous.Frequency) <= 1) {
        confirmed.add(spotKey(spot))
        confirmed.add(spotKey(previous))
      }
    }
    nearby.push(spot)
    byCall.set(key, nearby)
  }
  return confirmed
}

export function exportSpotCSV(rows: Spot[]) {
  const cell = (value: string | number) => {
    const text = String(value)
    const safe = typeof value === 'string' && /^[\s\u0000-\u001f]*[=+\-@]/.test(text) ? "'" + text : text
    return '"' + safe.replaceAll('"', '""') + '"'
  }
  const csv = ['UTC,DX,kHz,Spotter,DXCC,Comment',
    ...rows.map(row => [new Date(row.At * 1000).toISOString(), row.Call, row.Frequency, row.Spotter, row.Entity ?? row.DXCC, row.Comment]
      .map(cell).join(','))].join('\r\n')
  const url = URL.createObjectURL(new Blob(['\uFEFF', csv], { type: 'text/csv;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = 'dxregia-spots.csv'
  link.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

const modeFor = (comment: string) => comment.match(/\b(FT8|FT4|SSB|CW|RTTY|PSK31|USB|LSB)\b/i)?.[1].toUpperCase() ?? ''
// Se il commento non indica il modo, lo si deduce dalla frequenza (piano di banda IARU regione 1).
const ft8 = [1840, 3573, 5357, 7074, 10136, 14074, 18100, 21074, 24915, 28074, 50313]
const ft4 = [3575.5, 7047.5, 10140, 14080, 18104, 21140, 24919, 28180, 50318]
const segments: [number, number, string][] = [
  [1800, 1838, 'CW'], [1838, 1843, 'DIGI'], [1843, 2000, 'SSB'],
  [3500, 3570, 'CW'], [3570, 3600, 'DIGI'], [3600, 3800, 'SSB'],
  [5351, 5367, 'DIGI'],
  [7000, 7040, 'CW'], [7040, 7060, 'DIGI'], [7060, 7200, 'SSB'],
  [10100, 10130, 'CW'], [10130, 10150, 'DIGI'],
  [14000, 14070, 'CW'], [14070, 14099, 'DIGI'], [14112, 14350, 'SSB'],
  [18068, 18095, 'CW'], [18095, 18109, 'DIGI'], [18111, 18168, 'SSB'],
  [21000, 21070, 'CW'], [21070, 21150, 'DIGI'], [21151, 21450, 'SSB'],
  [24890, 24915, 'CW'], [24915, 24931, 'DIGI'], [24931, 24990, 'SSB'],
  [28000, 28070, 'CW'], [28070, 28190, 'DIGI'], [28225, 29700, 'SSB'],
  [50000, 50100, 'CW'], [50100, 50300, 'SSB'], [50300, 50500, 'DIGI'],
]
function inferMode(khz: number) {
  if (ft8.some(value => Math.abs(khz - value) <= 1.2)) return 'FT8'
  if (ft4.some(value => Math.abs(khz - value) <= 1.2)) return 'FT4'
  return segments.find(([low, high]) => khz >= low && khz < high)?.[2] ?? ''
}
export const modeInfo = (spot: Spot) => {
  const declared = modeFor(spot.Comment)
  return declared ? { mode: declared, inferred: false } : { mode: inferMode(spot.Frequency), inferred: true }
}
export const groupOf = (value: string) => ['FT8', 'FT4', 'RTTY', 'PSK31'].includes(value) ? 'DIGI' : ['SSB', 'USB', 'LSB'].includes(value) ? 'SSB' : value

export const alertStorageKey = 'dxregia-radio-alerts-v1'

export function localDay(date: Date) {
  return [date.getFullYear(), date.getMonth() + 1, date.getDate()].map((part, index) =>
    String(part).padStart(index ? 2 : 4, '0')).join('-')
}
