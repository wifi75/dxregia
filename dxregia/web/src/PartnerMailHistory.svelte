<script lang="ts">
  import { onMount } from 'svelte'
  import { lang } from './i18n.svelte'

  type Message = { ID: string; At: string; Call: string; Recipient: string; Kind: string; Subject: string; Text: string; Status: string }

  const kinds: Record<string, string> = { 'node-details': 'Dati pubblici del nodo', 'ip-instructions': 'Istruzioni senza password', instructions: 'Istruzioni', confirmed: 'Conferma collegamento', pending: 'Avviso di mancato collegamento' }
  const statuses: Record<string, string> = { accepted: 'Accettata da SMTP', failed: 'Invio non riuscito', unknown: 'Esito non registrato' }
  const date = $derived(new Intl.DateTimeFormat(lang(), { dateStyle: 'short', timeStyle: 'medium', timeZone: 'UTC' }))

  let items = $state<Message[]>([])
  let latest = $state<{ Call: string; Recipient: string; Status: string }[]>([])
  let selected = $state<Message>()
  let next = $state('')
  let query = $state('')
  let activeQuery = $state('')
  let busy = $state(false)
  let error = $state('')

  async function load(more = false) {
    if (busy) return
    busy = true
    error = ''
    const search = more ? activeQuery : query.trim()
    const params = new URLSearchParams({ q: search })
    if (more) params.set('before', next)
    try {
      const response = await fetch('/api/partners/mail-history?' + params, { cache: 'no-store', signal: AbortSignal.timeout(8000) })
      if (!response.ok) throw new Error()
      const data: { Items: Message[]; Next: string; Latest?: typeof latest } = await response.json()
      items = more ? [...items, ...data.Items] : data.Items
      next = data.Next
      activeQuery = search
      if (!more) selected = items.find(item => item.ID === selected?.ID)
      if (!more) latest = data.Latest ?? []
    } catch {
      error = 'Impossibile leggere lo storico email.'
    } finally {
      busy = false
    }
  }

  onMount(() => { void load() })
</script>

<section class="glass-panel" aria-labelledby="partner-mail-heading">
  <h1 id="partner-mail-heading">Storico email partner</h1>
  <p>Apri un messaggio per leggerne il contenuto. Le password dei collegamenti sono oscurate.</p>
  <form class="actions" onsubmit={event => { event.preventDefault(); void load() }}>
    <label>Nominativo, destinatario o oggetto <input type="search" bind:value={query} /></label>
    <button disabled={busy}>Cerca</button>
    <button class="secondary" type="button" disabled={busy} onclick={() => load()}>Aggiorna</button>
    {#if query}<button class="secondary" type="button" disabled={busy} onclick={() => { query = ''; void load() }}>Mostra tutte</button>{/if}
  </form>

  {#if error}<p role="alert">{error}</p>{/if}
  {#if items.length}
    <div class="table-wrap" aria-busy={busy}>
      <table>
        <thead><tr><th scope="col">Data e ora UTC</th><th scope="col">Nodo</th><th scope="col">Destinatario</th><th scope="col">Oggetto</th><th scope="col">Esito</th></tr></thead>
        <tbody>
          {#each items as item (item.ID)}
            <tr class:current={selected?.ID === item.ID}>
              <td>{date.format(new Date(item.At))}</td>
              <td>{item.Call}</td>
              <td>{item.Recipient}</td>
              <td><button class="subject" type="button" aria-expanded={selected?.ID === item.ID} onclick={() => { selected = selected?.ID === item.ID ? undefined : item }}>{item.Subject}</button></td>
              <td><span class="delivery" data-status={item.Status}>{statuses[item.Status] ?? statuses.unknown}</span></td>
            </tr>
            {#if selected?.ID === item.ID}
              <tr class="preview"><td colspan="5"><strong>{kinds[item.Kind]}</strong><pre>{item.Text}</pre></td></tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
    {#if next}<button class="secondary more" type="button" disabled={busy} onclick={() => load(true)}>Carica email precedenti</button>{/if}
  {:else if busy}
    <p role="status">Caricamento…</p>
  {:else if !error}
    <p role="status">{latest.length ? 'Per questi esiti non è disponibile una copia del messaggio.' : activeQuery ? 'Nessuna email corrisponde alla ricerca.' : 'Nessuna email archiviata.'}</p>
  {/if}
  {#if latest.length}
    <h2>Ultimo esito per nodo</h2>
    <div class="table-wrap"><table>
      <thead><tr><th scope="col">Nodo</th><th scope="col">Destinatario</th><th scope="col">Ultimo esito registrato</th></tr></thead>
      <tbody>{#each latest as row}<tr><td>{row.Call}</td><td>{row.Recipient}</td><td>{row.Status}</td></tr>{/each}</tbody>
    </table></div>
  {/if}
</section>

<style>
  h1 { margin: 0 0 .5rem; }
  .actions { display: flex; flex-wrap: wrap; align-items: end; gap: .65rem; margin: 1rem 0; }
  label { display: grid; gap: .35rem; }
  input { min-height: 42px; min-width: 15rem; padding: .45rem .65rem; border: 1px solid #a3bccc; border-radius: 7px; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; }
  th, td { padding: .55rem .7rem; border-bottom: 1px solid #c5d7e2; }
  th { background: #edf6fa; }
  td:first-child { white-space: nowrap; }
  .current, .preview { background: #eaf4fa; }
  .subject { min-height: auto; padding: .1rem 0; border: 0; background: transparent; color: #174e79; text-decoration: underline; font: inherit; text-align: left; box-shadow: none; }
  .delivery { white-space: nowrap; }
  .delivery[data-status='accepted'] { color: #186642; }
  .delivery[data-status='failed'] { color: #a21e26; }
  .preview pre { font: inherit; white-space: pre-wrap; overflow-wrap: anywhere; }
  .more { margin-top: .8rem; }
</style>
