<script lang="ts">
  import { tr } from './adminI18n'
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'

  type Filter = { Sort: string; Name: string; Text: string; Size: number }
  let fallback = $state(30)
  let rows = $state<{ pc: string; hops: number }[]>([])
  let custom = $state(false)
  let filters = $state<Filter[]>([])
  let status = $state('')
  let filterStatus = $state('')
  let busy = $state(false)
  const kinds: Record<string, string> = { ann: 'Annunci', route: 'Instradamento', spots: 'Spot', wcy: 'WCY', wwv: 'WWV' }

  async function loadFilters() {
    const response = await fetch('/api/filters')
    if (response.ok) filters = await response.json()
  }
  onMount(async () => {
    try {
      const response = await fetch('/api/hops')
      if (response.ok) {
        const data = await response.json()
        fallback = data.Default; custom = data.Custom
        rows = Object.entries<number>(data.Table).map(([pc, hops]) => ({ pc, hops })).sort((a, b) => Number(a.pc) - Number(b.pc))
      }
      await loadFilters()
    } catch { status = 'Impostazioni non disponibili.' }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    status = ''
    const Table: Record<string, number> = {}
    for (const row of rows) if (row.pc.trim()) Table[row.pc.trim()] = Number(row.hops)
    try {
      const response = await fetch('/api/hops', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ Default: Number(fallback), Table }),
      })
      status = response.ok ? 'Hop salvati. Si applicano entro pochi secondi, senza riavvio.' : (await response.text()).trim()
      if (response.ok) custom = true
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }
  async function remove(filter: Filter) {
    if (!confirm(tr('Eliminare il filtro {}?', filter.Name))) return
    const response = await fetch('/api/filters/' + filter.Sort + '/' + filter.Name, { method: 'DELETE' })
    filterStatus = response.ok ? 'Filtro eliminato.' : (await response.text()).trim()
    if (response.ok) await loadFilters()
  }
</script>

<section class="glass-panel" aria-labelledby="hops-title">
  <h1 id="hops-title">Hop<Hint text="L'hop count limita quanti nodi può attraversare un'informazione prima di essere scartata. Valori bassi evitano che spot e annunci girino all'infinito; valori alti li fanno arrivare più lontano." /></h1>
  <form onsubmit={save}>
    <label>Hop predefiniti<Hint text="Valore usato per ogni tipo di messaggio che non ha una riga propria qui sotto. Da 1 a 99; DXSpider usa 30." />
      <input type="number" min="1" max="99" bind:value={fallback} disabled={busy} />
    </label>
    <p>Eccezioni per tipo di messaggio<Hint text="Ogni tipo di pacchetto del protocollo si chiama «PC» seguito da un numero (ad esempio PC16 e PC17 per gli utenti collegati, PC19 e PC21 per i nodi). Qui puoi dargli un hop count diverso dal predefinito." /></p>
    {#each rows as row, i}
      <div class="row">
        <label class="small">PC<input aria-label="Numero PC" title="Numero del pacchetto, da 1 a 999" bind:value={row.pc} disabled={busy} /></label>
        <label class="small">Hop<input aria-label="Hop" type="number" min="1" max="99" bind:value={row.hops} disabled={busy} /></label>
        <button type="button" class="ghost" aria-label="Elimina la riga" onclick={() => rows.splice(i, 1)}>✕</button>
      </div>
    {/each}
    <div class="actions">
      <button type="button" class="ghost" onclick={() => rows.push({ pc: '', hops: 15 })}>Aggiungi eccezione</button>
      <button disabled={busy}>Salva hop</button>
    </div>
    {#if !custom}<small>Valori originali di DXSpider: nessuna modifica salvata finora.</small>{/if}
  </form>
  {#if status}<p class="note" role="status">{status}</p>{/if}

  <h2>Filtri<Hint text="Un filtro decide quali spot, annunci o bollettini passano verso un nodo o un utente. Si creano con i comandi accept e reject della Console Telnet; qui si controllano e si eliminano." /></h2>
  {#if filters.length}
    <ul class="filters">
      {#each filters as filter}
        <li>
          <details>
            <summary><strong>{kinds[filter.Sort] ?? filter.Sort}</strong> · {filter.Name.replace(/\.pl$/, '')} <small>{filter.Size} byte</small></summary>
            <pre>{filter.Text}</pre>
          </details>
          <button type="button" class="ghost" onclick={() => remove(filter)}>Elimina</button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="note">Nessun filtro attivo.</p>
  {/if}
  {#if filterStatus}<p class="note" role="status">{filterStatus}</p>{/if}
  <p class="note">Le liste «nodx», «noann» e «nowwv» del vecchio DXSpider non sono più lette da questo motore: per escludere un tipo di informazione usa i filtri.</p>
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.8rem 0 .8rem; font-size: 1.1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p, .note { margin: 0; color: #465e70; }
  .note { margin-top: .6rem; }
  label { display: grid; gap: .35rem; font-weight: 650; width: min(100%, 220px); }
  .row { display: flex; gap: .6rem; align-items: end; }
  .small { width: 7rem; }
  input { min-height: 40px; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; width: 100%; }
  .actions { display: flex; gap: .6rem; flex-wrap: wrap; }
  .filters { list-style: none; margin: 0 0 .8rem; padding: 0; display: grid; gap: .5rem; }
  .filters li { display: flex; gap: .6rem; align-items: start; justify-content: space-between; }
  details { flex: 1; min-width: 0; }
  summary { cursor: pointer; }
  pre { margin: .4rem 0 0; padding: .6rem .8rem; overflow-x: auto; border-radius: 6px; background: #0001; max-height: 14rem; }
  small { color: #465e70; }
</style>
