export type SetupFields = {
  node: string
  sysop: string
  locator: string
  port?: number
  name: string
  qth: string
  email: string
  advertisedHost: string
  latitude?: number
  longitude?: number
}

const aliases = {
  node: ['node', 'mycall'],
  sysop: ['sysop', 'myalias'],
  locator: ['locator', 'mylocator'],
  port: ['port', 'telnetPort'],
  name: ['name', 'myname'],
  qth: ['qth', 'myqth'],
  email: ['email', 'myemail'],
  advertisedHost: ['advertisedHost', 'publicHost'],
  latitude: ['latitude', 'mylatitude'],
  longitude: ['longitude', 'mylongitude'],
} as const

export function readSetupImport(text: string): Partial<SetupFields> {
  const source = text.replace(/^\uFEFF/, '').trim()
  let data: Record<string, unknown> = {}
  if (source.startsWith('{') || source.startsWith('[')) {
    const parsed: unknown = JSON.parse(source)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error('Il JSON deve contenere le impostazioni del nodo.')
    }
    data = parsed as Record<string, unknown>
  } else {
    for (const match of source.matchAll(/^\s*(?:our\s+)?\$([a-zA-Z]\w*)\s*=\s*(?:"([^"$@\\\r\n]*)"|'((?:\\['\\]|[^'\\\r\n])*)'|([+-]?\d+(?:\.\d+)?))\s*;\s*(?:#[^\n]*)?$/gm)) {
      const [, name, doubleQuoted, singleQuoted, number] = match
      data[name] = number !== undefined ? Number(number) : doubleQuoted ?? singleQuoted?.replace(/\\(['\\])/g, '$1')
    }
    // Listeners.pm can define multiple listeners. Only a single literal port is unambiguous.
    const active = source.split('\n').filter(line => !line.trimStart().startsWith('#')).join('\n')
    if (/^\s*(?:our\s+)?@listen\s*=/m.test(active)) {
      const listeners = [...active.matchAll(/\[\s*["'][^"'\r\n]+["']\s*,\s*(\d+)\s*\]/g)]
      if (listeners.length === 1) data.port = Number(listeners[0][1])
    }
  }

  const fields: Record<string, string | number> = {}
  for (const [field, keys] of Object.entries(aliases)) {
    const found = keys.find(key => Object.hasOwn(data, key))
    if (!found) continue
    const value = data[found]
    if (field === 'port' || field === 'latitude' || field === 'longitude') {
      if (typeof value === 'number' && Number.isFinite(value)) fields[field] = value
    } else if (typeof value === 'string') {
      fields[field] = value
    }
  }
  if (!Object.keys(fields).length) {
    throw new Error('Nessuna impostazione riconosciuta. Usa DXVars.pm, Listeners.pm con un solo listener, o un JSON di configurazione del nodo.')
  }
  return fields
}
