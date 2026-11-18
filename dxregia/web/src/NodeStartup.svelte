<script lang="ts">
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'

  let commands = $state<string[]>([])
  let extra = $state<string[]>([])
  let status = $state('')
  let busy = $state(false)

  onMount(async () => {
    try {
      const response = await fetch('/api/startup')
      if (response.ok) { const data = await response.json(); commands = data.Commands; extra = data.Extra }
    } catch { status = 'Impostazioni non disponibili.' }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    status = ''
    try {
      const response = await fetch('/api/startup', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Commands: commands.filter(command => command.trim()) }),
      })
      status = response.ok ? 'Script salvato. Si esegue al prossimo avvio del nodo.' : (await response.text()).trim()
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }
</script>

<section class="glass-panel" aria-labelledby="startup-title">
  <h1 id="startup-title">Script di avvio<Hint text="Comandi che DXSpider esegue una volta sola, ogni volta che il nodo parte. Servono per impostare valori che non cambiano, ad esempio il numero massimo di spot in memoria." /></h1>
  <form onsubmit={save}>
    <p>Un comando per riga. Sono ammessi <code>set/…</code>, <code>unset/…</code>, <code>load/…</code>, <code>connect NOMINATIVO</code> e <code>set/var $Pacchetto::nome = valore</code> con un numero o un testo semplice.</p>
    {#each commands as _, i}
      <div class="row">
        <input aria-label="Comando" title="Esempio: set/var $Spot::maxspots = 100" bind:value={commands[i]} />
        <button type="button" class="ghost" aria-label="Elimina la riga" onclick={() => commands.splice(i, 1)}>✕</button>
      </div>
    {:else}
      <p>Nessun comando di avvio impostato dal pannello.</p>
    {/each}
    <div class="actions">
      <button type="button" class="ghost" onclick={() => commands.push('')}>Aggiungi comando</button>
      <button disabled={busy}>Salva script</button>
    </div>
  </form>
  {#if extra.length}
    <p class="note">Righe già presenti che il pannello conserva senza modificarle<Hint text="Sono commenti o comandi in una forma che il pannello non gestisce. Restano nel file così come sono." /></p>
    <pre>{extra.join('\n')}</pre>
  {/if}
  {#if status}<p class="note" role="status">{status}</p>{/if}
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p, .note { margin: 0 0 .5rem; color: #465e70; }
  .row { display: flex; gap: .4rem; width: min(100%, 640px); }
  input { flex: 1; min-height: 40px; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; font-family: ui-monospace, monospace; }
  .actions { display: flex; gap: .6rem; flex-wrap: wrap; }
  pre { margin: 0 0 1rem; padding: .6rem .8rem; overflow-x: auto; border-radius: 6px; background: #0001; }
</style>
