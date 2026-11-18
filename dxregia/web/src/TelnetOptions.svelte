<script lang="ts">
  import { onMount } from 'svelte'

  const toggles = [
    ['PasswordRequired', 'Richiedi una password a tutti i collegamenti Telnet'],
    ['SlashCallsigns', 'Ammetti nominativi con barra, ad esempio IU3WJQ/P'],
    ['ReplaceSession', 'Consenti al nuovo collegamento di sostituire la sessione precedente'],
  ] as const
  const limits = [
    ['ConcurrentUsers', 'Limite connessioni in ingresso per utente', 1000],
    ['ConcurrentNodes', 'Limite connessioni in ingresso per nodo', 1000],
    ['MultipleSessions', 'Sessioni aggiuntive per nominativo', 15],
    ['ReconnectSeconds', 'Soglia per riconoscere connessioni ravvicinate (secondi)', 86400],
    ['MaxSSID', 'SSID massimo da cercare per sessioni aggiuntive', 15],
  ] as const
  type Options = Record<typeof toggles[number][0], boolean> &
    Record<typeof limits[number][0], number> & { Registration: number }
  type Snapshot = { Options: Options; Applied: boolean }

  let options = $state<Options>()
  let busy = $state(false)
  let status = $state('')
  let error = $state(false)

  async function readOptions(): Promise<Snapshot> {
    const response = await fetch('/api/telnet-options', { cache: 'no-store' })
    if (!response.ok) throw new Error((await response.text()).trim())
    return await response.json() as Snapshot
  }

  onMount(async () => {
    try {
      const result = await readOptions()
      options = result.Options
      if (!result.Applied) status = 'Modifiche salvate, in attesa di Mojo.'
    } catch (failure) {
      error = true
      status = failure instanceof Error ? failure.message : 'Impostazioni non disponibili.'
    }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (!options || busy) return
    if (options.PasswordRequired &&
      !confirm('I collegamenti Telnet senza una password impostata verranno rifiutati. Vuoi continuare?')) return
    status = ''
    error = false
    busy = true
    try {
      const response = await fetch('/api/telnet-options', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(options),
      })
      if (!response.ok) throw new Error((await response.text()).trim())
      let result = await response.json() as Snapshot
      for (let i = 0; !result.Applied && i < 8; i++) {
        status = 'Salvato. Attendo conferma da Mojo…'
        await new Promise(resolve => setTimeout(resolve, 1000))
        result = await readOptions()
      }
      options = result.Options
      status = result.Applied ? 'Impostazioni applicate da Mojo.' : 'Salvato. Mojo non ha ancora confermato le modifiche.'
    } catch (failure) {
      error = true
      status = failure instanceof Error ? failure.message : 'Impossibile salvare.'
    } finally {
      busy = false
    }
  }
</script>

<section aria-labelledby="telnet-options-title">
  <h1 id="telnet-options-title">Opzioni Telnet</h1>
  {#if options}
    <form onsubmit={save}>
      <fieldset disabled={busy}>
        <legend>Accesso</legend>
        {#each toggles.slice(0, 2) as [key, label]}
          <label class="check"><input type="checkbox" bind:checked={options[key]} />{label}</label>
        {/each}
        <label>Registrazione utenti
          <select bind:value={options.Registration}>
            <option value={0}>Registrazione non richiesta</option>
            <option value={1}>Registrazione richiesta</option>
            <option value={2}>Modalità de-registrazione Mojo</option>
          </select>
        </label>
      </fieldset>
      <fieldset disabled={busy}>
        <legend>Sessioni e riconnessioni</legend>
        <label class="check"><input type="checkbox" bind:checked={options.ReplaceSession} />{toggles[2][1]}</label>
        <div class="limits">
          {#each limits as [key, label, max]}
            <label>{label}<input type="number" bind:value={options[key]} min="0" {max} step="1" required /></label>
          {/each}
        </div>
        <p>Imposta 0 nei limiti di connessione per disattivare il relativo limite. Le modifiche si applicano ai collegamenti successivi.</p>
      </fieldset>
      <button disabled={busy}>{busy ? 'Attendo Mojo…' : 'Salva opzioni'}</button>
    </form>
  {/if}
  {#if status}<p role={error ? 'alert' : 'status'}>{status}</p>{/if}
</section>

<style>
  section { padding: clamp(14px, 2vw, 24px); border: 1px solid #b7c9d5; border-radius: 8px; background: #fff; }
  h1 { margin: 0 0 1.2rem; }
  form, fieldset { display: grid; gap: .75rem; }
  form > button { justify-self: start; }
  fieldset { min-width: 0; margin: 0; border: 1px solid #c7d6e0; padding: 15px; border-radius: 7px; }
  legend { padding: 0 .4rem; font-weight: 700; }
  label { display: grid; gap: .3rem; }
  .check { display: flex; align-items: center; gap: .55rem; }
  .limits { display: grid; gap: .8rem; grid-template-columns: repeat(auto-fit, minmax(min(100%, 205px), 1fr)); }
  :is(select, input[type=number]) { width: min(100%, 22rem); min-height: 40px; font: inherit; padding: .4rem; border: 1px solid #a4bbc9; border-radius: 6px; }
  fieldset p { margin: 0; color: #435e72; }
  [role=alert] { color: #902a21; }
</style>
