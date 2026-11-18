import './glass.css'
import './controls.css'
import './theme.css'
import './theme.svelte'
import { mount } from 'svelte'

// Una risposta 401 dalle API del pannello significa sessione scaduta: il pannello torna all'accesso.
const nativeFetch = window.fetch.bind(window)
window.fetch = async (...args: Parameters<typeof fetch>) => {
  const response = await nativeFetch(...args)
  if (response.status === 401) {
    const input = args[0]
    const url = typeof input === 'string' ? input : input instanceof URL ? input.pathname : input.url
    if (new URL(url, location.href).pathname.startsWith('/api/') && !/\/api\/auth(\/|$)/.test(url)) {
      window.dispatchEvent(new Event('dxregia-unauthorized'))
    }
  }
  return response
}
import App from './App.svelte'

mount(App, { target: document.getElementById('app')! })
