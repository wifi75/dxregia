<script lang="ts">
  import { onMount } from 'svelte'

  type Entry = { Minute: string; Hour: string; Day: string; Month: string; Weekday: string; Command: string }
  let text = $state('')
  let maxBytes = $state(2000)
  let entries = $state<Entry[]>([])
  let extra = $state<string[]>([])
  let motdStatus = $state('')
  let cronStatus = $state('')
  let busy = $state(false)
  const blank = (): Entry => ({ Minute: '0', Hour: '3', Day: '*', Month: '*', Weekday: '*', Command: '' })
  const size = $derived(new TextEncoder().encode(text).length)

  onMount(async () => {
    try {
      const [motd, cron] = await Promise.all([fetch('/api/motd'), fetch('/api/crontab')])
      if (motd.ok) { const data = await motd.json(); text = data.Text; maxBytes = data.MaxBytes }
      if (cron.ok) { const data = await cron.json(); entries = data.Entries; extra = data.Extra }
    } catch {
      motdStatus = cronStatus = 'Impostazioni non disponibili.'
    }
  })

  async function save(url: string, body: unknown, done: string): Promise<string> {
    busy = true
    try {
      const response = await fetch(url, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
      return response.ok ? done : (await response.text()).trim()
    } catch {
      return 'Connessione interrotta.'
    } finally {
      busy = false
    }
  }
  const saveMotd = async (event: SubmitEvent) => {
    event.preventDefault()
    motdStatus = await save('/api/motd', { Text: text }, text.trim() ? 'Messaggio salvato: lo vedranno i prossimi collegamenti.' : 'Messaggio rimosso.')
  }
  const saveCron = async (event: SubmitEvent) => {
    event.preventDefault()
    cronStatus = await save('/api/crontab', { Entries: entries }, 'Pianificazione salvata: DXSpider la rilegge da solo.')
  }
</script>

<section class="glass-panel" aria-labelledby="motd-title">
  <h1 id="motd-title">Messaggio del giorno</h1>
  <form onsubmit={saveMotd}>
    <p>Appare agli utenti subito dopo il collegamento Telnet. Lascialo vuoto per non mostrare nulla.</p>
    <label>Testo
      <textarea bind:value={text} rows="6" disabled={busy}></textarea>
    </label>
    <small class:over={size > maxBytes}>{size} / {maxBytes} byte</small>
    <button disabled={busy || size > maxBytes}>Salva messaggio</button>
  </form>
  {#if motdStatus}<p role="status">{motdStatus}</p>{/if}

  <h2>Pianificazione dei comandi</h2>
  <form onsubmit={saveCron}>
    <p>Comandi che il nodo esegue da solo agli orari indicati, nel formato cron: minuto, ora, giorno del mese, mese, giorno della settimana (0 = domenica). Usa <code>*</code> per «ogni».</p>
    {#each entries as entry, i}
      <div class="row">
        <input aria-label="Minuto" bind:value={entry.Minute} />
        <input aria-label="Ora" bind:value={entry.Hour} />
        <input aria-label="Giorno del mese" bind:value={entry.Day} />
        <input aria-label="Mese" bind:value={entry.Month} />
        <input aria-label="Giorno della settimana" bind:value={entry.Weekday} />
        <input class="command" aria-label="Comando" placeholder="es. export_users" bind:value={entry.Command} />
        <button type="button" class="ghost" onclick={() => entries.splice(i, 1)} aria-label="Elimina la riga">✕</button>
      </div>
    {:else}
      <p>Nessun comando pianificato dal pannello.</p>
    {/each}
    <div class="actions">
      <button type="button" class="ghost" onclick={() => entries.push(blank())}>Aggiungi comando</button>
      <button disabled={busy}>Salva pianificazione</button>
    </div>
  </form>
  {#if extra.length}
    <p>Righe già presenti che il pannello conserva senza modificarle:</p>
    <pre>{extra.join('\n')}</pre>
  {/if}
  {#if cronStatus}<p role="status">{cronStatus}</p>{/if}
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.8rem 0 .8rem; font-size: 1.1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p { margin: 0; color: #465e70; }
  label { display: grid; gap: .35rem; width: min(100%, 560px); font-weight: 650; }
  textarea, input { width: 100%; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
  textarea { resize: vertical; font-family: ui-monospace, monospace; }
  small.over { color: #b3261e; }
  .row { display: grid; grid-template-columns: repeat(5, 4.2rem) minmax(10rem, 1fr) auto; gap: .4rem; width: 100%; }
  .row input { min-height: 40px; }
  .actions { display: flex; gap: .6rem; flex-wrap: wrap; }
  pre { margin: 0; padding: .6rem .8rem; overflow-x: auto; border-radius: 6px; background: #0001; }
  @media (max-width: 700px) { .row { grid-template-columns: repeat(5, 1fr); } .row .command { grid-column: 1 / -1; } }
</style>
