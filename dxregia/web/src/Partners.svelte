<script lang="ts">
  import { onMount } from 'svelte'

  import type { ContactRequest } from './ContactRequests.svelte'
  let { draft }: { draft?: ContactRequest } = $props()

  type Details = { Call: string; Name: string; Contact: string; Email: string; Host: string; Port?: number }
  type Partner = Details & { RegisteredAt: string; FirstConnectedAt: string | null; Provisioning?: string; Connection?: string; MailStatus?: string; Reconnect?: string; LastConnectAttempt?: string; RetryEligibleAt?: string; Notes: { Text: string; At: string }[] | null }
  const empty = (): Details => ({ Call: '', Name: '', Contact: '', Email: '', Host: '', Port: undefined })
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })
  let partners = $state<Partner[]>([])
  let selected = $state<Partner>()
  let form = $state<Details>(empty())
  let editing = $state(false)
  let note = $state('')
  let error = $state('')
  let message = $state('')
  let busy = $state(true)
  let peerPassword = $state('')
  let enabling = $state(false)
  let outgoingPassword = $state('')
  let remoteAuth = $state<'none' | 'password'>('none')
  let autoReconnect = $state(false)
  let sendInstructions = $state(false)
  let emailStatus = $state('')
  let localNode = $state('')

  onMount(async () => {
    try {
      const response = await fetch('/api/partners')
      if (!response.ok) throw new Error()
      partners = Object.values(await response.json() as Record<string, Partner>)
    } catch {
      error = 'Impossibile caricare i nodi partner.'
    } finally {
      busy = false
    }
  })

  onMount(() => {
    void fetch('/api/node-settings', { cache: 'no-store' }).then(async response => {
      if (response.status === 200) localNode = (await response.json()).node || ''
    }).catch(() => {})
    const timer = setInterval(() => {
      if (selected && !enabling && !busy) void refreshProvisioning()
    }, 5000)
    return () => clearInterval(timer)
  })

  function open(partner?: Partner) {
    peerPassword = ''
    outgoingPassword = ''
    remoteAuth = 'none'
    autoReconnect = partner?.Reconnect === 'Abilitata'
    sendInstructions = false
    emailStatus = ''
    selected = partner
    form = partner
      ? { Call: partner.Call, Name: partner.Name, Contact: partner.Contact, Email: partner.Email, Host: partner.Host, Port: partner.Port }
      : empty()
    note = ''
    error = ''
    message = ''
    editing = true
  }

  $effect(() => {
    if (!draft) return
    const request = draft
    open()
    form = { Call: request.Call, Name: request.Name, Contact: request.Contact, Email: request.Email, Host: request.Host, Port: request.Port }
    note = ['Richiesta ricevuta dal portale il ' + date.format(new Date(request.ReceivedAt)), request.Message].filter(Boolean).join('\n\n')
    message = 'Scheda precompilata. Controlla i dati prima di salvare.'
    requestAnimationFrame(() => document.getElementById('partner-draft')?.focus())
  })

  async function refreshProvisioning() {
    const call = selected?.Call
    if (!call) return
    try {
      const response = await fetch('/api/partners')
      if (!response.ok) throw new Error()
      const rows: Record<string, Partner> = await response.json()
      if (rows[call] && selected?.Call === call) {
        selected = rows[call]
        if (selected.Provisioning !== 'In attesa di Mojo' && selected.Provisioning !== 'Abilitato in Mojo') peerPassword = ''
      }
    } catch {
      error = 'Impossibile leggere lo stato da Mojo.'
    }
  }

  async function enable(outgoingOnly = false) {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/enable', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ SendMail: !outgoingOnly && sendInstructions, OutgoingOnly: outgoingOnly }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      if (selected?.Call !== call) return
      if (response.headers.get('content-type')?.includes('application/json')) {
        const result = await response.json()
        peerPassword = result.Password ?? ''
        emailStatus = result.MailStatus ?? ''
      }
      await refreshProvisioning()
    } catch {
      error = 'Impossibile inviare la richiesta di abilitazione.'
    } finally {
      enabling = false
    }
  }

  function connectionTone(state?: string) {
    if (state === 'Attivo') return 'green'
    if (state === 'In coda' || state === 'Collegamento in corso') return 'amber'
    if (!state || state === 'Stato Mojo non disponibile') return 'neutral'
    return 'red'
  }

  async function connect(event: SubmitEvent) {
    event.preventDefault()
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/connect' + (autoReconnect ? '?retry=true' : ''), {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(remoteAuth === 'password' ? outgoingPassword : ''),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      outgoingPassword = ''
      await refreshProvisioning()
    } catch {
      error = 'Impossibile avviare il collegamento.'
    } finally {
      enabling = false
    }
  }

  async function disableReconnect() {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/reconnect', {
        method: 'DELETE', headers: { 'Content-Type': 'application/json' },
      })
      if (!response.ok) throw new Error()
      await refreshProvisioning()
    } catch {
      error = 'Impossibile disattivare la riconnessione automatica.'
    } finally {
      enabling = false
    }
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (busy || enabling) return
    const prepareOnly = !selected && remoteAuth === 'password' && (event.submitter as HTMLButtonElement | null)?.value !== 'directory'
    const connectNow = !selected && !prepareOnly && (event.submitter as HTMLButtonElement | null)?.value !== 'directory'
    busy = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners' + (selected ? '/' + encodeURIComponent(selected.Call) : ''), {
        method: selected ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, Note: note, ConnectNow: connectNow }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const saved: Partner = await response.json()
      partners = [...partners.filter(partner => partner.Call !== saved.Call), saved]
      open(saved)
      if (prepareOnly) {
        remoteAuth = 'password'
        await enable(true)
        message = 'Nodo salvato. Quando Mojo conferma, inserisci la password fornita dal partner e premi Connetti nodo.'
        return
      }
      message = connectNow ? 'Nodo salvato. DXRegia prepara il collegamento e riproverà anche dopo il riavvio.' : 'Anagrafica salvata.'
    } catch {
      error = 'Impossibile salvare il nodo partner.'
    } finally {
      busy = false
    }
  }
</script>

<section class="glass-panel" aria-labelledby="partners-title">
  <div class="toolbar">
    <h2 id="partners-title">Nodi partner</h2>
    <button type="button" onclick={() => open()} disabled={busy || enabling}>Aggiungi nodo</button>
  </div>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if message}<p role="status">{message}</p>{/if}
  {#if partners.length}
    <div class="table-wrap">
      <table>
        <thead><tr><th>Indicativo</th><th>Nome nodo</th><th>Referente</th><th>Registrato il</th></tr></thead>
        <tbody>
          {#each partners.toSorted((a, b) => a.Call.localeCompare(b.Call)) as partner (partner.Call)}
            <tr>
              <td><button class="call" type="button" onclick={() => open(partner)} disabled={busy || enabling}>{partner.Call}</button></td>
              <td>{partner.Name || '—'}</td>
              <td>{partner.Contact || '—'}</td>
              <td>{date.format(new Date(partner.RegisteredAt))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if !busy && !error}
    <p>Nessun nodo partner registrato.</p>
  {/if}

  {#if editing}
    <h3 id="partner-draft" tabindex="-1">{selected ? selected.Call : 'Nuovo nodo partner'}</h3>
    {#if selected}
      <p>Prima connessione: {selected.FirstConnectedAt ? date.format(new Date(selected.FirstConnectedAt)) : 'Da rilevare'}</p>
      <p class="observed-state">
        <span class="led" data-tone={selected.Provisioning === 'Abilitato in Mojo' ? 'green' : selected.Provisioning === 'In attesa di Mojo' ? 'amber' : selected.Provisioning?.includes('verificare') ? 'red' : 'neutral'}
          title={selected.Provisioning || 'Accesso in ingresso non abilitato'} aria-hidden="true"></span>
        Accesso in ingresso: {selected.Provisioning === 'Pronto in uscita' ? 'Non abilitato' : selected.Provisioning || 'Non abilitato dal pannello'}
      </p>
      {#if !selected.Provisioning || !['In attesa di Mojo', 'Abilitato in Mojo'].includes(selected.Provisioning)}
        {#if selected.Email}<label class="mail-opt"><input type="checkbox" bind:checked={sendInstructions} /> Invia le istruzioni e l’esito via email a {selected.Email}</label>{/if}
        <button type="button" disabled={enabling || busy} onclick={() => enable()}>Abilita accesso in ingresso</button>
      {/if}
      <button type="button" disabled={enabling} onclick={refreshProvisioning}>Aggiorna stato Mojo</button>
      {#if selected.MailStatus || emailStatus}<p>Email: {selected.MailStatus || emailStatus}</p>{/if}
      {#if peerPassword && selected.Provisioning === 'Abilitato in Mojo'}
        <p>Conserva la password del partner. Servirà per configurare il collegamento dal suo nodo.</p>
        <code>{peerPassword}</code>
      {/if}
    {/if}
    {#if selected}
      {@const outgoingReady = selected.Provisioning === 'Abilitato in Mojo' || selected.Provisioning === 'Pronto in uscita'}
      <h3>Collegamento in uscita</h3>
      <p class="observed-state" role="status">
        <span class="led" data-tone={connectionTone(selected.Connection)}
          title={selected.Connection || 'Mojo non ha ancora riportato lo stato'} aria-hidden="true"></span>
        Collegamento: {selected.Connection || 'Stato in attesa'}
      </p>
      {#if selected.Reconnect}
        <p>Riconnessione automatica: {selected.Reconnect}</p>
        {#if selected.LastConnectAttempt}
          <p>Ultimo tentativo {new Date(selected.LastConnectAttempt).toLocaleString('it-IT', { timeZone: 'UTC' })} UTC</p>
        {/if}
        {#if selected.RetryEligibleAt}
          <p>Nuovo tentativo consentito dal {new Date(selected.RetryEligibleAt).toLocaleString('it-IT', { timeZone: 'UTC' })} UTC</p>
        {/if}
        <button type="button" disabled={enabling} onclick={disableReconnect}>Disattiva riconnessione</button>
      {/if}
      <details class="script-preview">
        <summary>Anteprima dello script Mojo</summary>
        <pre>{[
          'timeout 15',
          'connect telnet ' + selected.Host + ' ' + selected.Port,
          "'login' '" + (localNode || 'INDICATIVO NODO LOCALE') + "'",
          ...(remoteAuth === 'password' ? ["'password:' '••••••••'"] : []),
          'client ' + selected.Call + ' telnet',
        ].join('\n')}</pre>
      </details>
      <form onsubmit={connect}>
        <label>Autenticazione del nodo remoto
          <select bind:value={remoteAuth} disabled={enabling || busy}>
            <option value="none">Senza password</option>
            <option value="password">Con password</option>
          </select>
        </label>
        {#if remoteAuth === 'password'}
          <label>Password fornita dal nodo remoto
            <input type="password" bind:value={outgoingPassword} pattern={'[A-Za-z0-9]{1,72}'} maxlength="72" autocomplete="off" required />
          </label>
        {/if}
        <label class="mail-opt"><input type="checkbox" bind:checked={autoReconnect} /> Riprova automaticamente dopo una caduta del collegamento</label>
        {#if outgoingReady}
          <button disabled={enabling || busy}>Connetti nodo</button>
        {:else}
          <button type="button" disabled={enabling || busy || selected.Provisioning === 'In attesa di Mojo'}
            onclick={() => enable(true)}>Prepara collegamento in uscita</button>
        {/if}
      </form>
    {/if}
    <form onsubmit={save}>
      <div class="fields">
        <label>Indicativo <input bind:value={form.Call} maxlength="12" readonly={Boolean(selected)} required /></label>
        <label>Nome del nodo <input bind:value={form.Name} maxlength="120" /></label>
        <label>Nome del referente <input bind:value={form.Contact} maxlength="120" /></label>
        <label>Email <input type="email" bind:value={form.Email} maxlength="254" /></label>
        <label>Host <input bind:value={form.Host} maxlength="253" required /></label>
        <label>Porta remota <input type="number" bind:value={form.Port} min="1" max="65535" required /></label>
      </div>
      {#if !selected}
        <label class="remote-auth">Autenticazione del nodo remoto
          <select bind:value={remoteAuth}>
            <option value="none">Senza password</option>
            <option value="password">Con password</option>
          </select>
        </label>
      {/if}
      <label class="note">Aggiungi una nota <textarea bind:value={note} maxlength="4000" rows="3" placeholder="Es. Inviata email al referente"></textarea></label>
      <button type="submit" value={selected ? 'directory' : 'connect'} disabled={busy || enabling}>
        {busy ? 'Salvataggio…' : selected ? 'Salva anagrafica' : remoteAuth === 'password' ? 'Salva e prepara' : 'Salva e connetti'}
      </button>
      {#if !selected}<button class="secondary" type="submit" value="directory" disabled={busy || enabling}>Salva solo anagrafica</button>{/if}
    </form>
    {#if selected?.Notes?.length}
      <h3>Storico note</h3>
      <ul>
        {#each selected.Notes.toReversed() as entry}
          <li><time datetime={entry.At}>{date.format(new Date(entry.At))}</time><p>{entry.Text}</p></li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>

<style>
  section { margin-top: 20px; padding: 16px; }
  .toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
  h2 { margin: 0; font-size: 1.15rem; }
  h3 { margin: 18px 0 6px; font-size: 1rem; }
  .table-wrap { overflow: auto; margin-top: 12px; }
  table { width: 100%; text-align: left; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 5px 8px; border-bottom: 1px solid #d7dfe7; }
  th { background: #edf3f9; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); gap: 12px 16px; }
  label { display: grid; min-width: 0; gap: 5px; font-weight: 600; }
  .note { margin: 12px 0; }
  .observed-state { display: flex; align-items: center; gap: .55rem; }
  .led { width: 11px; height: 11px; flex: 0 0 11px; border-radius: 50%; background: #708391; box-shadow: 0 0 0 3px #70839123; }
  .led[data-tone=green] { background: #167b50; box-shadow: 0 0 0 3px #167b5029; }
  .led[data-tone=red] { background: #b32731; box-shadow: 0 0 0 3px #b3273129; }
  .led[data-tone=amber] { background: #ad6a10; box-shadow: 0 0 0 3px #ad6a1029; }
  .script-preview pre { padding: .65rem .9rem; background: #edf4f8; overflow: auto; }
  .remote-auth { margin-top: 12px; max-width: 300px; }
  input, textarea, select { min-width: 0; padding: 8px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-height: 44px; }

  .call { background: none; color: #214e80; border: 0; padding: 4px; min-height: 34px; text-decoration: underline; box-shadow: none; }
  :is(button,input,textarea):focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
  [role=alert] { color: #9b2424; }
  ul { padding: 0; list-style: none; }
  li { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  li p { white-space: pre-wrap; overflow-wrap: anywhere; margin: 3px 0; }
  .mail-opt { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
  .mail-opt input { min-height: auto; }
  code { overflow-wrap: anywhere; }
  time { color: #485b6d; font-size: .85rem; }
</style>
