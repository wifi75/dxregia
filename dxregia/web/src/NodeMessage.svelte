<script lang="ts">
  import { onMount, tick } from 'svelte'
  import Hint from './Hint.svelte'

  type Entry = { Minute: string; Hour: string; Day: string; Month: string; Weekday: string; Command: string }
  const languages = [
    { code: 'en', flag: 'gb', name: 'English · predefinito' }, { code: 'it', flag: 'it', name: 'Italiano' },
    { code: 'es', flag: 'es', name: 'Español' }, { code: 'de', flag: 'de', name: 'Deutsch' },
    { code: 'fr', flag: 'fr', name: 'Français' }, { code: 'pt', flag: 'pt', name: 'Português' },
  ]
  let language = $state('en')
  let banners = $state<Record<string, string>>({ en: '', it: '', es: '', de: '', fr: '', pt: '' })
  let loaded = $state(false)
  let maxBytes = $state(2000)
  let entries = $state<Entry[]>([])
  let extra = $state<string[]>([])
  let motdStatus = $state('')
  let cronStatus = $state('')
  let busy = $state(false)
  // La casella si adatta al testo: alta quanto serve (almeno 6 righe), senza barra di scorrimento.
  let box = $state<HTMLTextAreaElement>()
  function fit() {
    if (!box) return
    box.style.height = 'auto'
    box.style.height = Math.min(box.scrollHeight + 2, window.innerHeight * 0.8) + 'px'
  }
  $effect(() => { banners[language]; box; void tick().then(fit) })
  const blank = (): Entry => ({ Minute: '0', Hour: '3', Day: '*', Month: '*', Weekday: '*', Command: '' })
  const size = $derived(new TextEncoder().encode(banners[language]).length)
  const flag = $derived(languages.find(option => option.code === language)?.flag ?? 'gb')

  onMount(async () => {
    try {
      const [motd, cron] = await Promise.all([fetch('/api/motd'), fetch('/api/crontab')])
      if (motd.ok) {
        const data = await motd.json()
        banners = { ...banners, ...data.Translations, en: data.Text }
        maxBytes = data.MaxBytes
        loaded = true
      } else motdStatus = 'Impossibile leggere i banner Telnet.'
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
    if (!loaded || busy) return
    motdStatus = await save('/api/motd', { Text: banners[language], Language: language }, 'Testo della lingua selezionata salvato.')
  }
  const saveCron = async (event: SubmitEvent) => {
    event.preventDefault()
    cronStatus = await save('/api/crontab', { Entries: entries }, 'Pianificazione salvata: DXSpider la rilegge da solo.')
  }
</script>

<section class="glass-panel" aria-labelledby="motd-title">
  <h1 id="motd-title">Messaggio del giorno</h1>
  <form onsubmit={saveMotd}>
    <p>Il banner appare prima di <code>login:</code>. Il nodo usa il Paese dell’IP pubblico o la lingua dell’indicativo del nodo per gli accessi dalla rete locale. Se non trova una traduzione, mostra l’inglese.</p>
    <label>Lingua del banner
      <span class="language-picker">
        <img src={'/flags/' + flag + '.svg'} width="24" height="17" alt="" />
        <select bind:value={language} disabled={!loaded || busy} onchange={() => { motdStatus = '' }}>
          {#each languages as option (option.code)}<option value={option.code}>{option.name}</option>{/each}
        </select>
      </span>
    </label>
    <label>Testo<Hint text="Si mostra a chi si collega via Telnet, prima del prompt. Puoi andare a capo. Massimo 2000 byte: le lettere accentate ne occupano 2." />
      <textarea bind:this={box} bind:value={banners[language]} rows="6" wrap="off" disabled={!loaded || busy} class="telnet-banner"></textarea>
    </label>
    <small class:over={size > maxBytes}>{size} / {maxBytes} byte</small>
    <button disabled={!loaded || busy || size > maxBytes}>Salva messaggio</button>
  </form>
  {#if motdStatus}<p role="status">{motdStatus}</p>{/if}

  <h2>Pianificazione dei comandi<Hint text="Come il cron di Linux: ogni riga ha cinque campi di orario e un comando di DXSpider. Il nodo la rilegge da solo, senza riavvio." /></h2>
  <form onsubmit={saveCron}>
    <p>Comandi che il nodo esegue da solo agli orari indicati, nel formato cron: minuto, ora, giorno del mese, mese, giorno della settimana (0 = domenica). Usa <code>*</code> per «ogni».</p>
    {#each entries as entry, i}
      <div class="row">
        <input aria-label="Minuto" title="Minuto: 0–59, oppure * per ogni minuto, */5 per ogni 5" bind:value={entry.Minute} />
        <input aria-label="Ora" title="Ora: 0–23 (UTC), oppure *" bind:value={entry.Hour} />
        <input aria-label="Giorno del mese" title="Giorno del mese: 1–31, oppure *" bind:value={entry.Day} />
        <input aria-label="Mese" title="Mese: 1–12 oppure jan–dec, oppure *" bind:value={entry.Month} />
        <input aria-label="Giorno della settimana" title="Giorno della settimana: 0 (domenica)–6, oppure *" bind:value={entry.Weekday} />
        <input class="command" aria-label="Comando" title="Comando DXSpider senza virgolette, ad esempio export_users" placeholder="es. export_users" bind:value={entry.Command} />
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
  .language-picker { display: inline-flex; align-items: center; gap: .55rem; }
  .telnet-banner { font-family: ui-monospace, monospace; }
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.8rem 0 .8rem; font-size: 1.1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p { margin: 0; color: #465e70; }
  label { display: grid; gap: .35rem; width: min(100%, 560px); font-weight: 650; }
  textarea, input { width: 100%; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
  textarea { resize: vertical; overflow: auto; white-space: pre; font-family: ui-monospace, monospace; }
  small.over { color: #b3261e; }
  .row { display: grid; grid-template-columns: repeat(5, 4.2rem) minmax(10rem, 1fr) auto; gap: .4rem; width: 100%; }
  .row input { min-height: 40px; }
  .actions { display: flex; gap: .6rem; flex-wrap: wrap; }
  pre { margin: 0; padding: .6rem .8rem; overflow-x: auto; border-radius: 6px; background: #0001; }
  @media (max-width: 700px) { .row { grid-template-columns: repeat(5, 1fr); } .row .command { grid-column: 1 / -1; } }
</style>
