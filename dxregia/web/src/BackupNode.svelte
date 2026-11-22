<script lang="ts">
  import { onMount } from 'svelte'

  type File = { Name: string; Size: number; At: number }
  type Schedule = { Enabled: boolean; Hour: number; Keep: number }
  let backups = $state<File[]>([])
  let schedule = $state<Schedule>({ Enabled: false, Hour: 3, Keep: 7 })
  let running = $state(false)
  let lastError = $state('')
  let loaded = $state(false)
  let busy = $state(false)
  let message = $state('')
  let removing = $state<File>()
  let dialog: HTMLDialogElement
  let timer: ReturnType<typeof setTimeout> | undefined

  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'medium', timeZone: 'UTC' })
  const size = (bytes: number) => bytes >= 1048576 ? (bytes / 1048576).toFixed(1) + ' MiB' : Math.max(1, Math.round(bytes / 1024)) + ' KiB'

  async function load() {
    try {
      const response = await fetch('/api/backups', { cache: 'no-store' })
      if (!response.ok) throw new Error()
      const data = await response.json()
      backups = data.Backups ?? []
      running = data.Running
      lastError = data.LastError ?? ''
      if (!loaded) schedule = data.Schedule
      loaded = true
    } catch {
      message = 'Impossibile leggere i backup.'
    }
  }

  function poll() {
    clearTimeout(timer)
    timer = setTimeout(async () => { await load(); poll() }, running ? 2000 : 20000)
  }

  async function send(path: string, method: string, body?: unknown) {
    return fetch(path, { method, headers: { 'Content-Type': 'application/json' }, body: body === undefined ? undefined : JSON.stringify(body) })
  }

  async function createNow() {
    busy = true
    message = ''
    try {
      const response = await send('/api/backups/run', 'POST', {})
      if (!response.ok) message = (await response.text()).trim() || 'Impossibile avviare il backup.'
      else { running = true; poll() }
    } catch { message = 'Connessione interrotta.' } finally { busy = false }
  }

  async function saveSchedule(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    message = ''
    try {
      const response = await send('/api/backups/schedule', 'PUT', { Enabled: schedule.Enabled, Hour: Number(schedule.Hour), Keep: Number(schedule.Keep) })
      message = response.ok ? 'Pianificazione salvata.' : (await response.text()).trim()
    } catch { message = 'Connessione interrotta.' } finally { busy = false }
  }

  async function remove() {
    if (!removing) return
    const name = removing.Name
    dialog.close()
    try {
      const response = await send('/api/backups/files/' + encodeURIComponent(name), 'DELETE')
      message = response.ok ? 'Backup eliminato.' : 'Impossibile eliminare il backup.'
      await load()
    } catch { message = 'Connessione interrotta.' }
  }

  onMount(() => {
    void load().then(poll)
    return () => clearTimeout(timer)
  })
</script>

<section class="glass-panel" aria-labelledby="node-backup-heading">
  <h2 id="node-backup-heading">Backup del nodo</h2>
  <p class="hint">La copia si fa con il nodo acceso, senza interrompere Telnet né i collegamenti. Contiene credenziali e segreti del nodo: scaricala e conservala in un posto sicuro. Per riprendere un nodo da un backup serve il ripristino da terminale.</p>
  <div class="actions">
    <button type="button" onclick={createNow} disabled={busy || running}>{running ? 'Backup in corso…' : 'Crea backup ora'}</button>
  </div>
  {#if lastError}<p role="alert">{lastError}</p>{/if}

  {#if loaded && backups.length}
    <div class="table-wrap">
      <table>
        <thead><tr><th scope="col">Data e ora UTC</th><th scope="col">Dimensione</th><th scope="col">Azioni</th></tr></thead>
        <tbody>
          {#each backups as file (file.Name)}
            <tr>
              <td>{file.At ? date.format(new Date(file.At * 1000)) : file.Name}</td>
              <td>{size(file.Size)}</td>
              <td class="row-actions">
                <a href={'/api/backups/files/' + encodeURIComponent(file.Name)} download={file.Name}>Scarica</a>
                <button type="button" class="secondary" onclick={() => { removing = file; dialog.showModal() }}>Elimina</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if loaded}
    <p>Non ci sono ancora backup.</p>
  {/if}

  <h3>Backup automatico</h3>
  <form onsubmit={saveSchedule}>
    <label class="choice"><input type="checkbox" bind:checked={schedule.Enabled} disabled={busy} /> Crea un backup ogni giorno</label>
    <label class="field">Ora (UTC) <input type="number" min="0" max="23" step="1" required bind:value={schedule.Hour} disabled={busy} /></label>
    <label class="field">Copie da conservare <input type="number" min="1" max="30" step="1" required bind:value={schedule.Keep} disabled={busy} /></label>
    <button disabled={busy}>Salva pianificazione</button>
  </form>
  {#if message}<p class="note" role="status">{message}</p>{/if}
</section>

<dialog class="glass-panel ask-dialog" bind:this={dialog} aria-labelledby="backup-remove-title">
  <h2 id="backup-remove-title">Elimina backup</h2>
  <p>Il backup selezionato verrà eliminato e non si potrà recuperare.</p>
  {#if removing}<p><strong>{date.format(new Date(removing.At * 1000))} UTC</strong></p>{/if}
  <div class="actions end">
    <button type="button" class="secondary" onclick={() => dialog.close()}>Annulla</button>
    <button type="button" onclick={remove}>Elimina</button>
  </div>
</dialog>

<style>
  h3 { margin: 18px 0 8px; font-size: 1rem; }
  .hint, .note { color: #465e70; }
  .actions { display: flex; flex-wrap: wrap; gap: 10px; margin: 12px 0; }
  .actions.end { justify-content: flex-end; }
  .table-wrap { overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; }
  th, td { padding: .5rem .6rem; border-bottom: 1px solid #c5d7e2; }
  th { background: #edf6fa; }
  td:first-child { white-space: nowrap; }
  .row-actions { display: flex; align-items: center; gap: 12px; }
  form { display: grid; gap: .7rem; justify-items: start; }
  .choice { display: flex; align-items: center; gap: .5rem; }
  .field { display: flex; align-items: center; gap: .5rem; }
  .field input { width: 5rem; min-height: 40px; border: 1px solid #a3bccc; border-radius: 7px; background: #fff; font: inherit; padding: .3rem .5rem; }
  .ask-dialog { width: min(440px, calc(100% - 32px)); box-sizing: border-box; padding: 24px; color: var(--glass-text); }
  .ask-dialog::backdrop { background: #0b2238a0; backdrop-filter: blur(3px); }
</style>
