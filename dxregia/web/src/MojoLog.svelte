<script lang="ts">
  import { onMount } from 'svelte'
  import { lang } from './i18n.svelte'

  type Event = { At: number; Kind: string; Message: string; Call?: string; Host?: string; Session?: string; Direction?: string; Country?: string }
  let events = $state<Event[]>([])
  let category = $state('')
  let busy = $state(false)
  let error = $state('')
  let query = $state('')
  let anonymous = $state(false)
  let visible = $derived(category ? events.filter(event => event.Kind === category) : events)
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'medium', timeZone: 'UTC' })
  const countryNames = $derived(new Intl.DisplayNames([lang()], { type: 'region' }))

  async function refresh() {
    if (busy) return
    busy = true
    error = ''
    try {
      const params = new URLSearchParams({ q: query, anonymous: String(anonymous) })
      const response = await fetch('/api/mojo-log?' + params, { cache: 'no-store', signal: AbortSignal.timeout(8000) })
      if (!response.ok) throw new Error()
      events = await response.json()
    } catch {
      error = 'Impossibile leggere il log DXSpider.'
    } finally {
      busy = false
    }
  }

  onMount(() => {
    void refresh()
    const timer = setInterval(refresh, 5000)
    return () => clearInterval(timer)
  })
</script>

<section class="glass-panel" aria-labelledby="mojo-log-heading">
  <div class="heading">
    <h1 id="mojo-log-heading">Log DXSpider</h1>
    <label>Categoria
      <select bind:value={category}><option value="">Tutte</option><option>Collegamenti</option><option>Avvio</option></select>
    </label>
    <button class="secondary" onclick={refresh} disabled={busy}>{busy ? 'Aggiornamento…' : 'Aggiorna'}</button>
  </div>
  <form class="heading" onsubmit={event => { event.preventDefault(); void refresh() }}>
    <label>Nominativo, IP o sessione <input type="search" bind:value={query} placeholder="DA0BCC-7" /></label>
    <button class="secondary" disabled={busy}>Cerca</button>
    <label><input type="checkbox" bind:checked={anonymous} onchange={() => queueMicrotask(refresh)} /> Mostra accessi in ingresso senza nominativo</label>
  </form>
  {#if error}
    <p role="alert">{error}</p>
  {:else if visible.length}
    <div class="table-wrap">
      <table>
        <thead><tr><th scope="col">Data e ora UTC</th><th scope="col">Categoria</th><th scope="col">Nominativo</th><th scope="col">IP o host</th><th scope="col">Direzione</th><th scope="col">Sessione</th><th scope="col">Evento</th></tr></thead>
        <tbody>{#each visible as event}<tr><td>{event.At ? date.format(new Date(event.At * 1000)) : '—'}</td><td>{event.Kind}</td><td><span class="callsign">{event.Call || '—'}{#if event.Call && event.Country && /^[A-Z]{2}$/.test(event.Country)}<img src={'/flags/' + event.Country.toLowerCase() + '.svg'} width="22" height="16" alt={countryNames.of(event.Country) || event.Country} title={countryNames.of(event.Country)} />{/if}</span></td><td>{event.Host || '—'}</td><td>{event.Direction || '—'}</td><td>{event.Session || '—'}</td><td>{event.Message}</td></tr>{/each}</tbody>
      </table>
    </div>
    {#if visible.some(event => event.Call && event.Country)}<p class="geo-credit"><a href="https://db-ip.com/" target="_blank" rel="noreferrer">IP Geolocation by DB-IP</a></p>{/if}
  {:else if !busy}
    <p>Nessun evento corrisponde ai filtri.</p>
  {/if}
</section>

<style>
  .heading { display: flex; align-items: center; flex-wrap: wrap; gap: .8rem; margin-bottom: 14px; }
  h1 { margin: 0 auto 0 0; font-size: 1.6rem; }
  label { display: flex; align-items: center; gap: .5rem; }
  select, input[type="search"] { min-height: 40px; border: 1px solid #a3bccc; border-radius: 7px; background: #fff; font: inherit; padding: .3rem .5rem; }
  .table-wrap { overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; }
  th, td { padding: .5rem .6rem; border-bottom: 1px solid #c5d7e2; }
  th { background: #edf6fa; }
  td:first-child { white-space: nowrap; }
  .callsign { display: inline-flex; align-items: center; gap: .45rem; white-space: nowrap; }
  .geo-credit { margin-bottom: 0; font-size: .82rem; }
</style>
