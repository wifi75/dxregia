<script lang="ts">
  let file = $state<File>()
  let busy = $state(false)
  let message = $state('')

  async function verify(event: SubmitEvent) {
    event.preventDefault()
    if (!file || busy) return
    busy = true
    message = ''
    try {
      const response = await fetch('/api/backups/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/zip' },
        body: file,
      })
      message = response.ok ? 'Archivio verificato.' : await response.text()
      if (response.status === 404) message = 'Verifica non disponibile su questa installazione.'
    } catch {
      message = 'Connessione interrotta durante la verifica.'
    } finally {
      busy = false
    }
  }
</script>

<section class="glass-panel" aria-labelledby="backup-heading">
  <h2 id="backup-heading">Verifica un backup</h2>
  <form onsubmit={verify}>
    <label>
      Archivio DXRegia (.zip)
      <input type="file" accept=".zip,application/zip" disabled={busy}
        onchange={(event) => { file = event.currentTarget.files?.[0]; message = '' }} />
    </label>
    <button type="submit" disabled={!file || busy}>{busy ? 'Verifica in corso' : 'Verifica archivio'}</button>
  </form>
  <p role="status">{message}</p>
</section>
