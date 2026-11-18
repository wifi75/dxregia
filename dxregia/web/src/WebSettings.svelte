<script lang="ts">
  import { onMount } from 'svelte'

  type Settings = { port: number; editable: boolean }
  let current = $state<Settings>()
  let requested = $state<number>()
  let busy = $state(false)
  let message = $state('')
  let destination = $state('')

  onMount(async () => {
    try {
      const response = await fetch('/api/web-settings', { cache: 'no-store' })
      if (!response.ok) throw new Error()
      const settings: Settings = await response.json()
      current = settings
      requested = settings.port
    } catch {
      message = 'Impostazioni web non disponibili.'
    }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (!current || !requested || busy) return
    busy = true
    message = ''
    try {
      const response = await fetch('/api/web-settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requested),
      })
      if (!response.ok) {
        message = (await response.text()).trim()
        return
      }
      current = await response.json() as Settings
      const next = new URL(location.href)
      next.port = String(current.port)
      destination = next.href
      message = 'Porta salvata. Apri il pannello al nuovo indirizzo.'
    } catch {
      message = 'Connessione interrotta. Prova ad aprire il pannello sulla porta scelta.'
    } finally {
      busy = false
    }
  }
</script>

<section aria-labelledby="web-settings-heading">
  <h1 id="web-settings-heading">Accesso web</h1>
  {#if current}
    <p>Porta HTTP in uso <strong>{current.port}</strong></p>
    {#if current.editable}
      <form onsubmit={save}>
        <label for="web-port">Nuova porta HTTP</label>
        <div class="actions">
          <input id="web-port" type="number" bind:value={requested} min="1024" max="65535" required />
          <button disabled={busy || requested === current.port}>{busy ? 'Salvataggio…' : 'Applica porta'}</button>
        </div>
      </form>
    {:else}
      <p>La porta pubblica è gestita dall’installazione.</p>
    {/if}
  {/if}
  {#if message}<p role="status">{message}</p>{/if}
  {#if destination}<a href={destination}>Apri il pannello sulla porta {current?.port}</a>{/if}
</section>

<style>
  section { padding: 20px; border: 1px solid #cbd9e4; border-radius: 8px; background: #fff; }
  h1 { margin: 0; font-size: 1.6rem; }
  label { display: block; font-weight: 650; margin-bottom: .5rem; }
  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: .75rem; }
  input { width: 9rem; min-height: 44px; padding: .4rem .65rem; border: 1px solid #a7bdcd; border-radius: 6px; font: inherit; }
</style>
