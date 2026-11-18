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
  let autoReconnect = $state(false)
  let sendInstructions = $state(false)
  let emailStatus = $state('')

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
    const timer = setInterval(() => {
      if (selected && !enabling && !busy) void refreshProvisioning()
    }, 5000)
    return () => clearInterval(timer)
  })

  function open(partner?: Partner) {
    peerPassword = ''
    outgoingPassword = ''
    autoReconnect = false
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
      if (rows[call] && selected?.Call === call) selected = rows[call]
    } catch {
      error = 'Impossibile leggere lo stato da Mojo.'
    }
  }

  async function enable() {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/enable', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ SendMail: sendInstructions }),
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

  async function connect(event: SubmitEvent) {
    event.preventDefault()
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/connect' + (autoReconnect ? '?retry=true' : ''), {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(outgoingPassword),
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
    busy = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners' + (selected ? '/' + encodeURIComponent(selected.Call) : ''), {
        method: selected ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, Note: note }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const saved: Partner = await response.json()
      partners = [...partners.filter(partner => partner.Call !== saved.Call), saved]
      open(saved)
      message = 'Anagrafica salvata.'
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
      <p>Accesso in ingresso: {selected.Provisioning || 'Non abilitato dal pannello'}</p>
      {#if !selected.Provisioning || !['In attesa di Mojo', 'Abilitato in Mojo'].includes(selected.Provisioning)}
        {#if selected.Email}<label class="mail-opt"><input type="checkbox" bind:checked={sendInstructions} /> Invia le istruzioni e l’esito via email a {selected.Email}</label>{/if}
        <button type="button" disabled={enabling || busy} onclick={enable}>Abilita accesso in ingresso</button>
      {/if}
      <button type="button" disabled={enabling} onclick={refreshProvisioning}>Aggiorna stato Mojo</button>
      {#if selected.MailStatus || emailStatus}<p>Email: {selected.MailStatus || emailStatus}</p>{/if}
      {#if peerPassword}
        <p>Conserva la password del partner. Servirà per configurare il collegamento dal suo nodo.</p>
        <code>{peerPassword}</code>
      {/if}
    {/if}
    {#if selected?.Provisioning === 'Abilitato in Mojo'}
      <p>Collegamento: {selected.Connection || 'Aggiorna lo stato per verificare'}</p>
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
      <form onsubmit={connect}>
        <label>Password impostata dal nodo remoto, se richiesta
          <input type="password" bind:value={outgoingPassword} pattern="[A-Za-z0-9]{0,72}" maxlength="72" autocomplete="off" />
        </label>
        <label class="mail-opt"><input type="checkbox" bind:checked={autoReconnect} /> Riprova automaticamente dopo una caduta del collegamento</label>
        <button disabled={enabling || busy}>Connetti nodo</button>
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
      <label class="note">Aggiungi una nota <textarea bind:value={note} maxlength="4000" rows="3" placeholder="Es. Inviata email al referente"></textarea></label>
      <button disabled={busy}>{busy ? 'Salvataggio…' : 'Salva anagrafica'}</button>
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
  input, textarea { min-width: 0; padding: 8px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-height: 44px; }

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
