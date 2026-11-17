<script lang="ts">
  import { onMount } from 'svelte'

  let available = $state(false)
  let saved = $state(false)
  let pending = $state(false)
  let error = $state('')
  let node = $state('')
  let sysop = $state('')
  let locator = $state('')
  let port = $state<number | undefined>(7300)
  let code = $state('')

  onMount(async () => {
    try {
      const response = await fetch('/api/setup', { cache: 'no-store' })
      if (response.ok && response.headers.get('content-type')?.includes('application/json')) {
        available = (await response.json()).available === true
      }
    } catch {
      // The diagnostic page can run without the setup service.
    }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (pending) return
    pending = true
    error = ''
    try {
      const response = await fetch('/api/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + code.trim() },
        body: JSON.stringify({ node, sysop, locator, port }),
      })
      if (!response.ok) {
        error = (await response.text()).trim() || 'Salvataggio non riuscito.'
        return
      }
      code = ''
      saved = true
      available = false
    } catch {
      error = 'Il pannello non riesce a raggiungere il servizio di configurazione.'
    } finally {
      pending = false
    }
  }
</script>

{#if available || saved}
  <section aria-labelledby="setup-title">
    <h2 id="setup-title">Prima configurazione</h2>
    {#if saved}
      <p role="status">Configurazione Mojo salvata. L'avvio automatico del nodo è ancora in sviluppo.</p>
    {:else}
      <form onsubmit={save}>
        <div class="fields">
          <label>Nominativo del nodo <input bind:value={node} maxlength="12" required placeholder="IU3ABC-2" /></label>
          <label>Nominativo del sysop <input bind:value={sysop} maxlength="12" required placeholder="IU3ABC" /></label>
          <label>Locator <input bind:value={locator} maxlength="6" required placeholder="JN55AA" /></label>
          <label>Porta Telnet <input type="number" bind:value={port} min="1" max="65535" required /></label>
          <label class="code">Codice temporaneo <input type="password" bind:value={code} autocomplete="off" maxlength="48" required /></label>
        </div>
        {#if error}<p class="error" role="alert">{error}</p>{/if}
        <button disabled={pending}>{pending ? 'Salvataggio…' : 'Salva configurazione'}</button>
      </form>
    {/if}
  </section>
{/if}

<style>
  section { margin-top: 20px; padding: 16px; background: #fff; border: 1px solid #d7dfe7; }
  h2 { margin: 0 0 12px; font-size: 1.15rem; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); gap: 12px 16px; }
  label { display: grid; gap: 5px; font-weight: 600; }
  input { color: #182636; border: 1px solid #aab8c5; border-radius: 4px; min-height: 44px; padding: 8px; font: inherit; }
  button { margin-top: 14px; min-height: 44px; padding: 7px 13px; border: 1px solid #214e80; border-radius: 4px; background: #214e80; color: #fff; font: inherit; font-weight: 650; cursor: pointer; }
  :is(button, input):focus-visible { outline: 2px solid #214e80; outline-offset: 3px; }
  button:disabled { opacity: .7; cursor: wait; }
  .error { color: #9b2424; }
</style>
