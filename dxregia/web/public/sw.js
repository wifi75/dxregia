// Service worker di DXRegia: rende il portale installabile come app e lo apre anche senza rete.
// Le pagine si chiedono sempre prima alla rete; solo gli elementi grafici (assets, loghi, bandiere) vengono tenuti in cache.
// Le API e il pannello /admin non si memorizzano mai.
const CACHE = 'dxregia-v1'
const STATIC = ['/assets/', '/brand/', '/flags/']

self.addEventListener('install', () => self.skipWaiting())

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys()
      .then(keys => Promise.all(keys.filter(key => key !== CACHE).map(key => caches.delete(key))))
      .then(() => self.clients.claim()),
  )
})

self.addEventListener('fetch', event => {
  const request = event.request
  if (request.method !== 'GET') return
  const url = new URL(request.url)
  if (url.origin !== self.location.origin || url.pathname.startsWith('/api/') || url.pathname.startsWith('/admin')) return

  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then(response => {
          if (response.ok && (url.pathname === '/spots' || url.pathname === '/')) {
            const copy = response.clone()
            caches.open(CACHE).then(cache => cache.put('/spots', copy))
          }
          return response
        })
        .catch(() => caches.match('/spots').then(cached => cached || Response.error())),
    )
    return
  }

  if (STATIC.some(prefix => url.pathname.startsWith(prefix))) {
    event.respondWith(
      caches.match(request).then(cached => cached || fetch(request).then(response => {
        if (response.ok) {
          const copy = response.clone()
          caches.open(CACHE).then(cache => cache.put(request, copy))
        }
        return response
      })),
    )
  }
})

self.addEventListener('notificationclick', event => {
  if (event.notification.tag !== 'dxregia-radio-alerts') return
  event.notification.close()
  const target = new URL('/spots?view=alerts', self.location.origin).href
  event.waitUntil(self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then(async windows => {
    const existing = windows.find(client => client.url === target)
    if (existing) return existing.focus()
    return self.clients.openWindow(target)
  }))
})
