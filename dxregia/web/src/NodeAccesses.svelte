<script lang="ts">
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'
  import { lang } from './i18n.svelte'

  type Session = { Call: string; Kind: 'user' | 'node'; Host: string; Start: number; End?: number; Seconds?: number; Open?: boolean }
  type Caller = { Call: string; Kind: 'user' | 'node'; Count: number; Last: number; Host: string }
  let days = $state(7)
  let kind = $state<'user' | 'node' | 'all'>('user')
  let sessions = $state<Session[]>([])
  let callers = $state<Caller[]>([])
  let loaded = $state(false)
  let status = $state('')

  const date = $derived(new Intl.DateTimeFormat(lang() === 'it' ? 'it-IT' : lang(), { dateStyle: 'short', timeStyle: 'medium', timeZone: 'UTC' }))
  const shown = $derived(sessions.filter(item => kind === 'all' || item.Kind === kind))
  const people = $derived(callers.filter(item => kind === 'all' || item.Kind === kind))
  const stamp = (at: number) => date.format(new Date(at * 1000))
  function length(seconds?: number) {
    if (seconds === undefined) return '—'
    if (seconds < 60) return seconds + ' s'
    if (seconds < 3600) return Math.floor(seconds / 60) + ' min ' + (seconds % 60) + ' s'
    return Math.floor(seconds / 3600) + ' h ' + Math.floor((seconds % 3600) / 60) + ' min'
  }

  async function load() {
    status = ''
    try {
      const response = await fetch('/api/connections?days=' + days, { cache: 'no-store' })
      if (!response.ok) throw new Error()
      const data = await response.json()
      sessions = data.Sessions
      callers = data.Callers
    } catch {
      status = 'Impossibile leggere lo storico degli accessi.'
    } finally {
      loaded = true
    }
  }
  onMount(load)
</script>

<section class="glass-panel" aria-labelledby="access-title">
  <h1 id="access-title">Accessi al nodo<Hint text="Chi si è collegato al nodo via Telnet o dal web, con il nominativo, l'indirizzo di provenienza e la durata. I dati vengono dal registro di DXSpider." /></h1>
  <div class="bar">
    <label>Periodo
      <select bind:value={days} onchange={load}>
        <option value={1}>24 ore</option>
        <option value={7}>7 giorni</option>
        <option value={30}>30 giorni</option>
      </select>
    </label>
    <label>Mostra<Hint text="Utenti: radioamatori che si collegano al cluster. Nodi: altri cluster collegati al tuo come partner." />
      <select bind:value={kind}>
        <option value="user">Radioamatori</option>
        <option value="node">Nodi partner</option>
        <option value="all">Tutti</option>
      </select>
    </label>
    <button type="button" class="secondary" onclick={load}>Aggiorna</button>
  </div>

  {#if status}<p class="note" role="alert">{status}</p>{/if}
  {#if !loaded}
    <p class="note">Caricamento…</p>
  {:else if !shown.length}
    <p class="note">Nessun accesso nel periodo.</p>
  {:else}
    <h2>Chi si è collegato<Hint text="Un riepilogo per nominativo: quante volte si è collegato nel periodo e l'ultimo accesso." /></h2>
    <div class="table-wrap">
      <table>
        <thead><tr><th>Nominativo</th><th>Tipo</th><th>N. accessi</th><th>Ultimo accesso (UTC)</th><th>Da</th></tr></thead>
        <tbody>
          {#each people as caller (caller.Kind + caller.Call)}
            <tr><td><strong>{caller.Call}</strong></td><td>{caller.Kind === 'node' ? 'Nodo' : 'Radioamatore'}</td><td>{caller.Count}</td><td>{stamp(caller.Last)}</td><td>{caller.Host}</td></tr>
          {/each}
        </tbody>
      </table>
    </div>

    <h2>Elenco dei collegamenti</h2>
    <div class="table-wrap">
      <table>
        <thead><tr><th>Inizio (UTC)</th><th>Nominativo</th><th>Tipo</th><th>Da</th><th>Durata</th></tr></thead>
        <tbody>
          {#each shown as session (session.Kind + session.Call + session.Start)}
            <tr>
              <td>{stamp(session.Start)}</td>
              <td><strong>{session.Call}</strong></td>
              <td>{session.Kind === 'node' ? 'Nodo' : 'Radioamatore'}</td>
              <td>{session.Host}</td>
              <td>{#if session.Open}<span class="live">Collegato ora</span>{:else if session.End}{length(session.Seconds)}{:else}—{/if}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.6rem 0 .6rem; font-size: 1.1rem; }
  .bar { display: flex; flex-wrap: wrap; gap: .8rem 1.2rem; align-items: end; }
  label { display: grid; gap: .3rem; font-weight: 650; }
  select { min-height: 40px; border: 1px solid #a4bbc9; border-radius: 6px; padding: .4rem .6rem; font: inherit; }
  .note { margin: .8rem 0 0; color: #465e70; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 7px 12px; text-align: left; border-bottom: 1px solid #c4d6e3; }
  th { font-size: .8rem; text-transform: uppercase; letter-spacing: .05em; color: #1d5d8c; }
  tbody tr:nth-child(even) { background: #0000000a; }
  .live { color: #1a7a4a; font-weight: 700; }
</style>
