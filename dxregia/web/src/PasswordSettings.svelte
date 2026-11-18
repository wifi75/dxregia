<script lang="ts">
  let kind = $state<'web' | 'telnet'>('telnet')
  let password = $state('')
  let confirmation = $state('')
  let code = $state('')
  let busy = $state(false)
  let status = $state('')

  async function change(event: SubmitEvent) {
    event.preventDefault()
    if (busy) return
    if (password !== confirmation) {
      status = 'Le due password non coincidono.'
      return
    }
    busy = true
    status = ''
    try {
      const response = await fetch('/api/' + kind + '-password', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password, code }),
      })
      if (!response.ok) {
        status = (await response.text()).trim()
        return
      }
      password = confirmation = code = ''
      if (kind === 'web') {
        location.reload()
        return
      }
      status = 'In attesa di conferma da DXSpider…'
      for (let i = 0; i < 8; i++) {
        await new Promise(resolve => setTimeout(resolve, 1000))
        const check = await fetch('/api/telnet-password', { cache: 'no-store' })
        if (!check.ok) break
        const result = (await check.json()).status
        if (result === 'applied' || result === 'failed') {
          status = result === 'applied' ? 'Password Telnet impostata.' : 'DXSpider non ha applicato la password.'
          return
        }
      }
      status = 'Il nodo non ha ancora confermato il cambio.'
    } catch {
      status = 'Connessione interrotta durante il cambio password.'
    } finally {
      busy = false
    }
  }
</script>

<section aria-labelledby="password-title">
  <h1 id="password-title">Password sysop</h1>
  <form onsubmit={change}>
    <label>Accesso da aggiornare
      <select bind:value={kind} disabled={busy} onchange={() => { password = confirmation = code = ''; status = '' }}>
        <option value="telnet">Telnet DXSpider</option>
        <option value="web">Pannello web</option>
      </select>
    </label>
    <p>{kind === 'telnet' ? 'Usa questa password con il nominativo sysop su Telnet. Almeno 12 caratteri, senza spazi.' : 'Dopo il cambio dovrai accedere di nuovo con la nuova password.'}</p>
    <label>Nuova password <input type="password" bind:value={password} minlength="12" maxlength="72" autocomplete="new-password" required disabled={busy} /></label>
    <label>Conferma password <input type="password" bind:value={confirmation} autocomplete="new-password" required disabled={busy} /></label>
    <label>Nuovo codice 2FA <input bind:value={code} inputmode="numeric" autocomplete="one-time-code" pattern="[0-9]{6}" required disabled={busy} /></label>
    <button disabled={busy}>{busy ? 'Salvataggio…' : 'Cambia password'}</button>
  </form>
  {#if status}<p role="status">{status}</p>{/if}
</section>

<style>
  section { padding: 22px; border: 1px solid #b7c9d5; border-radius: 9px; background: #fff; }
  h1 { margin: 0 0 1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p { margin: 0; color: #465e70; }
  label { display: grid; gap: .35rem; width: min(100%, 360px); font-weight: 650; }
  :is(input, select) { width: 100%; min-height: 42px; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; }
</style>
