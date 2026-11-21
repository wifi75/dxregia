<script lang="ts">
  import { onMount } from 'svelte'

  type Info = {
    Installed: boolean; State: 'idle' | 'queued' | 'running' | 'success' | 'failed'
    Current: string; Available: string; CurrentTag?: string; AvailableTag?: string; UpToDate: boolean
    CheckedAt: number; CheckError?: string; Log?: string[]
  }
  let info = $state<Info>()
  let offline = $state(false)
  let loadError = $state('')
  let checking = $state(false)
  let starting = $state(false)
  let message = $state('')
  let dialog: HTMLDialogElement
  let timer: ReturnType<typeof setTimeout> | undefined
  let seenBusy = false

  const busy = $derived(info?.State === 'running' || info?.State === 'queued')
  const checked = $derived(info?.CheckedAt
    ? new Intl.DateTimeFormat(undefined, { dateStyle: 'short', timeStyle: 'medium' }).format(info.CheckedAt * 1000)
    : '')

  async function load(check = false) {
    try {
      const response = await fetch('/api/update' + (check ? '?check=1' : ''), { cache: 'no-store' })
      if (!response.ok) throw new Error(response.status === 401 ? 'Accesso sysop richiesto.' : 'Impossibile leggere lo stato.')
      info = await response.json()
      offline = false
      loadError = ''
      if (busy) seenBusy = true
    } catch (error) {
      // Durante il riavvio del servizio il pannello non risponde per qualche secondo: è normale.
      if (busy || seenBusy) offline = true
      else loadError = error instanceof Error ? error.message : 'Impossibile leggere lo stato.'
    }
  }

  function schedule() {
    clearTimeout(timer)
    timer = setTimeout(async () => {
      await load()
      schedule()
    }, busy || offline ? 2000 : 15000)
  }

  async function check() {
    checking = true
    message = ''
    await load(true)
    checking = false
  }

  async function start(event: SubmitEvent) {
    event.preventDefault()
    if (starting) return
    starting = true
    message = ''
    try {
      const response = await fetch('/api/update/start', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: '{}',
      })
      if (!response.ok) {
        message = (await response.text()).trim() || 'Impossibile avviare l’aggiornamento.'
        return
      }
      dialog.close()
      seenBusy = true
      await load()
      schedule()
    } catch {
      message = 'Connessione interrotta. Riprova.'
    } finally {
      starting = false
    }
  }

  function ask() {
    message = ''
    dialog.showModal()
  }

  onMount(() => {
    void load(true).then(schedule)
    return () => clearTimeout(timer)
  })
</script>

<section class="glass-panel" aria-labelledby="update-heading">
  <h2 id="update-heading">Aggiornamenti di DXRegia</h2>
  {#if loadError}
    <p role="alert">{loadError}</p>
  {:else if !info}
    <p role="status">Controllo in corso…</p>
  {:else}
    <dl class="versions">
      <div><dt>Installata</dt><dd>{#if info.CurrentTag}{info.CurrentTag} <small>({info.Current})</small>{:else}{info.Current || 'sconosciuta'}{/if}</dd></div>
      <div><dt>Disponibile</dt><dd>{#if info.AvailableTag}{info.AvailableTag} <small>({info.Available})</small>{:else}{info.Available || '—'}{/if}</dd></div>
    </dl>
    {#if info.CheckError}
      <p role="alert">{info.CheckError}</p>
    {:else if info.State === 'idle' || info.State === 'success' || info.State === 'failed'}
      <p class="status" role="status">
        {#if info.UpToDate}DXRegia è aggiornato.{:else if info.Available}Aggiornamento disponibile.{/if}
        {#if checked}<small>Ultimo controllo: {checked}</small>{/if}
      </p>
    {/if}

    {#if !info.Installed}
      <p class="hint">L’aggiornamento dal pannello si attiva con un ultimo aggiornamento fatto da terminale: collegati in SSH ed esegui <code>dxregia-update</code>. Dopo, basterà questo pulsante.</p>
    {/if}

    <div class="actions">
      <button type="button" class="secondary" onclick={check} disabled={checking || busy}>{checking ? 'Controllo…' : 'Controlla ora'}</button>
      <button type="button" onclick={ask} disabled={!info.Installed || busy || info.UpToDate || !info.Available}>
        {busy ? 'Aggiornamento in corso' : 'Aggiorna ora'}
      </button>
    </div>

    {#if busy || offline}
      <p class="status" role="status">
        {#if offline}Il servizio si sta riavviando. La pagina si ricollega da sola.{:else if info.State === 'queued'}Aggiornamento in partenza…{:else}Aggiornamento in corso. Non chiudere il server.{/if}
      </p>
    {:else if info.State === 'success' && info.Log?.length}
      <p class="status ok" role="status">Aggiornamento completato.</p>
    {:else if info.State === 'failed'}
      <p role="alert">L’aggiornamento non è riuscito. DXRegia è tornato alla versione precedente: leggi il registro qui sotto.</p>
    {/if}

    {#if info.Log?.length}
      <h3>Registro dell’ultimo aggiornamento</h3>
      <pre class="log" aria-label="Registro dell’aggiornamento">{info.Log.join('\n')}</pre>
    {/if}
  {/if}
</section>

<dialog class="glass-panel update-dialog" bind:this={dialog} aria-labelledby="update-title" aria-describedby="update-text">
  <h2 id="update-title">Aggiorna DXRegia</h2>
  <p id="update-text">Durante l’aggiornamento Telnet e i collegamenti con i nodi partner si interrompono per circa un minuto. Se qualcosa va storto, DXRegia torna da solo alla versione precedente.</p>
  <form onsubmit={start}>
    {#if message}<p role="alert">{message}</p>{/if}
    <div class="actions">
      <button type="button" class="secondary" onclick={() => dialog.close()}>Annulla</button>
      <button type="submit" disabled={starting} autofocus>{starting ? 'Avvio…' : 'Aggiorna ora'}</button>
    </div>
  </form>
</dialog>

<style>
  .versions { display: flex; flex-wrap: wrap; gap: 8px 32px; margin: 10px 0; }
  .versions div { display: grid; gap: 2px; }
  dt { font-size: .8rem; color: #3f5c71; }
  dd { margin: 0; font-weight: 650; font-variant-numeric: tabular-nums; }
  small { font-weight: 400; color: #506678; }
  .status { margin: 8px 0; }
  .status.ok { color: #12563e; font-weight: 650; }
  .hint { margin: 8px 0; }
  .actions { display: flex; flex-wrap: wrap; gap: 10px; margin: 12px 0; }
  h3 { margin: 14px 0 6px; font-size: 1rem; }
  .log { max-height: 320px; overflow: auto; margin: 0; padding: 10px 12px; border: 1px solid #cdd9e4; border-radius: 6px; background: #0f2233; color: #dce9f3; font-size: .82rem; white-space: pre-wrap; }
  .update-dialog { width: min(440px, calc(100% - 32px)); box-sizing: border-box; padding: 24px; color: var(--glass-text); }
  .update-dialog::backdrop { background: #0b2238a0; backdrop-filter: blur(3px); }
  .update-dialog .actions { justify-content: flex-end; }
</style>
