<script lang="ts">
  import { onMount } from 'svelte'
  import { tr } from './adminI18n'
  import { dxccFlags, dxccFlagName } from './dxccFlags'
  import { lang } from './i18n.svelte'

  import type { ContactRequest } from './ContactRequests.svelte'
  let { draft }: { draft?: ContactRequest } = $props()

  type Details = { Call: string; Name: string; Contact: string; Email: string; Host: string; Port?: number; Lang?: string; Person?: string }
  type Partner = Details & { DXCC?: number; RegisteredAt: string; FirstConnectedAt: string | null; FirstInboundAt?: string | null; Provisioning?: string; Connection?: string; MailStatus?: string; MailRetryAt?: string; IPMailStatus?: string; OutgoingReady?: boolean; CanResetCCInbound?: boolean; CanSetInboundIP?: boolean; InboundIP?: string; InboundTestUntil?: string; InboundAutoNext?: string; PublicMailAcceptedAt?: string; Reconnect?: string; LastConnectAttempt?: string; RetryEligibleAt?: string; CanResend?: boolean; Locked?: boolean; CanUnlock?: boolean; Direction?: string; Notes: { Text: string; At: string }[] | null }
  const empty = (): Details => ({ Call: '', Name: '', Contact: '', Email: '', Host: '', Port: undefined, Lang: '', Person: '' })
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })
  const regionNames = $derived(new Intl.DisplayNames([lang()], { type: 'region' }))
  let partners = $state<Partner[]>([])
  const isOnline = (partner: Partner) => !partner.Locked && partner.Connection === 'Attivo'
  const sortedPartners = $derived(partners.toSorted((a, b) =>
    Number(isOnline(b)) - Number(isOnline(a)) || a.Call.localeCompare(b.Call)))
  let onlyInbound = $state(false)
  let inboundPartners = $derived(sortedPartners.filter(partner => isOnline(partner) && partner.Direction === 'Ingresso'))
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
  let credentialRequest = 0
  let loadingCredential = $state(false)
  let remoteAuth = $state<'none' | 'password'>('none')
  let storedPassword = false
  let autoReconnect = $state(false)
  let emailStatus = $state('')
  let inboundIP = $state('')
  let inboundMode = $state<'password' | 'ip'>('password')
  let ipMailStatus = $state<{ text: string; error: boolean }>()
  let publicMail = $state<{ call: string; state: 'sending' | 'accepted' | 'failed'; text: string }>()
  let localNode = $state('')

  let checkingStatus = $state(false)
  let manualStatusCall = $state('')
  let statusError = $state('')

  onMount(() => {
    void fetch('/api/node-settings', { cache: 'no-store' }).then(async response => {
      if (response.status === 200) localNode = (await response.json()).node || ''
    }).catch(() => {})
    void refreshProvisioning().finally(() => { busy = false })
    const timer = setInterval(() => {
      if (!enabling && !busy) void refreshProvisioning()
    }, 5000)
    return () => clearInterval(timer)
  })

  function open(partner?: Partner) {
    peerPassword = ''
    outgoingPassword = ''
    remoteAuth = 'none'
    storedPassword = false
    autoReconnect = partner?.Reconnect === 'Abilitata'
    emailStatus = ''
    selected = partner
    inboundIP = partner?.InboundIP || partner?.Host || ''
    inboundMode = partner?.InboundIP ? 'ip' : 'password'
    ipMailStatus = undefined
    form = partner
      ? { Call: partner.Call, Name: partner.Name, Contact: partner.Contact, Email: partner.Email, Host: partner.Host, Port: partner.Port, Lang: partner.Lang ?? '', Person: partner.Person ?? '' }
      : empty()
    note = ''
    error = ''
    message = ''
    editing = true
    const request = ++credentialRequest
    loadingCredential = Boolean(partner)
    if (partner) {
      void fetch('/api/partners/' + encodeURIComponent(partner.Call) + '/credential', { cache: 'no-store' })
        .then(async response => {
          if (!response.ok || request !== credentialRequest) return
          const { password } = await response.json()
          if (request === credentialRequest && password) storedPassword = true
          if (request === credentialRequest && password && remoteAuth === 'none' && !outgoingPassword) {
            remoteAuth = 'password'
            outgoingPassword = password
          }
        }).catch(() => {}).finally(() => {
          if (request === credentialRequest) loadingCredential = false
        })
    }
  }

  $effect(() => {
    if (!draft) return
    const request = draft
    open()
    form = { Call: request.Call, Name: request.Name, Contact: request.Contact, Email: request.Email, Host: request.Host, Port: request.Port, Lang: request.Lang ?? '' }
    note = ['Richiesta ricevuta dal portale il ' + date.format(new Date(request.ReceivedAt)), request.Message].filter(Boolean).join('\n\n')
    message = 'Scheda precompilata. Controlla i dati prima di salvare.'
    requestAnimationFrame(() => document.getElementById('partner-draft')?.focus())
  })

  async function refreshProvisioning() {
    if (checkingStatus) return
    checkingStatus = true
    try {
      const response = await fetch('/api/partners', { cache: 'no-store', signal: AbortSignal.timeout(4000) })
      if (!response.ok) throw new Error()
      const rows: Record<string, Partner> = await response.json()
      partners = Object.values(rows)
      statusError = ''
      if (selected && rows[selected.Call]) {
        selected = rows[selected.Call]
        autoReconnect = selected.Reconnect === 'Abilitata'
        if (selected.Provisioning !== 'In attesa di Mojo' && selected.Provisioning !== 'Abilitato in Mojo') peerPassword = ''
      }
    } catch {
      statusError = 'Impossibile aggiornare lo stato dei nodi.'
    } finally {
      checkingStatus = false
    }
  }

  let passwordDialog: HTMLDialogElement
  let passwordTarget = $state('')
  let confirmation = $state<'password' | 'manual' | 'automatic'>('password')

  function openPartnerConfirmation(action: typeof confirmation) {
    if (!selected || enabling || passwordDialog.open) return
    confirmation = action
    passwordTarget = selected.Call
    passwordDialog.returnValue = ''
    passwordDialog.showModal()
  }

  function finishPasswordReset() {
    if (passwordDialog.returnValue === 'confirm' && selected?.Call === passwordTarget) {
      if (confirmation === 'password') void enable(false, true, false, true, Boolean(selected?.InboundIP))
      else void startInboundTest(confirmation === 'automatic')
    }
  }

  async function enable(outgoingOnly = false, mail = false, unlock = false, resetCCInbound = false, switchFromIP = false) {
    const call = selected?.Call
    if (!call || enabling) return
    if (unlock && !confirm(tr('Sbloccare {} in DXSpider e preparare il collegamento? Se il nodo ha già una password locale, potrà usarla per collegarsi.', call))) return
    enabling = true
    if (mail) emailStatus = 'In attesa dell’abilitazione'
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/enable', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ SendMail: !outgoingOnly && mail, OutgoingOnly: outgoingOnly, Unlock: unlock, ResetCCInbound: resetCCInbound, SwitchFromIP: switchFromIP }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        if (mail) emailStatus = error || "Richiesta di invio non riuscita"
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

  async function revealPassword() {
    const call = selected?.Call
    if (!call || enabling) return
    error = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/inbound-password', { cache: 'no-store' })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      peerPassword = (await response.json()).password
    } catch {
      error = 'Impossibile leggere la password.'
    }
  }

  async function copyPassword() {
    try {
      await navigator.clipboard.writeText(peerPassword)
      message = 'Password copiata.'
    } catch {
      message = 'Copia non riuscita: selezionala a mano.'
    }
  }

  async function removePartner() {
    const call = selected?.Call
    if (!call || enabling) return
    if (!confirm(tr('Eliminare il nodo {}?\n\nIl nodo viene scollegato, il suo accesso al tuo cluster viene rimosso (la sua password smette di funzionare) e la scheda viene cancellata. Non si può annullare.', call))) return
    enabling = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call), { method: 'DELETE' })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      partners = partners.filter(partner => partner.Call !== call)
      editing = false
      selected = undefined
      message = 'Nodo ' + call + ' eliminato.'
    } catch {
      error = 'Impossibile eliminare il nodo.'
    } finally {
      enabling = false
    }
  }

  async function resend() {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/resend', { method: 'POST' })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      message = 'Email accodata: il partner riceve di nuovo host, porta e password.'
      await refreshProvisioning()
    } catch {
      error = 'Impossibile reinviare le istruzioni.'
    } finally {
      enabling = false
    }
  }

  async function sendNodeDetails() {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = message = ''
    publicMail = { call, state: 'sending', text: 'Invio dei dati pubblici in corso…' }
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/send-node-details', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
      })
      publicMail = response.ok
        ? { call, state: 'accepted', text: 'Email con i dati pubblici accettata dal server di posta. La trovi nello Storico email partner.' }
        : { call, state: 'failed', text: (await response.text()).trim() || 'Il server di posta non ha confermato l’invio.' }
    } catch {
      publicMail = { call, state: 'failed', text: 'Invio non confermato. Controlla lo Storico email partner.' }
    } finally {
      enabling = false
    }
  }

  async function saveInboundIP(sendMail = false) {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = message = ''
    if (sendMail) ipMailStatus = { text: 'Invio delle istruzioni in corso…', error: false }
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/inbound-ip' + (sendMail ? '?send-mail=true' : ''), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(inboundIP),
      })
      if (selected?.Call !== call) return
      if (!response.ok) {
        const text = (await response.text()).trim() || 'Impossibile aggiornare l’accesso per IP.'
        if (sendMail) ipMailStatus = { text, error: true }
        else error = text
      } else if (sendMail) {
        ipMailStatus = { text: 'Istruzioni senza password accettate da SMTP. Vedi Storico email partner.', error: false }
      } else {
        message = 'Accesso senza password abilitato per l’indirizzo IP del partner.'
      }
      await refreshProvisioning()
    } catch {
      const text = 'Operazione non confermata. Controlla lo stato e lo Storico email partner.'
      if (sendMail && selected?.Call === call) ipMailStatus = { text, error: true }
      else error = text
    } finally {
      enabling = false
    }
  }

  async function startInboundTest(automatic: boolean, disable = false) {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    try {
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/inbound-test' + (automatic ? '?automatic=true' : ''), {
        method: disable ? 'DELETE' : 'POST', headers: { 'Content-Type': 'application/json' },
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      await refreshProvisioning()
    } catch {
      error = 'Impossibile avviare la prova in ingresso.'
    } finally {
      enabling = false
    }
  }

  function mailTone(status?: string, retryAt?: string) {
    if (retryAt || status?.startsWith('In attesa') || status?.includes('accodate')) return 'amber'
    if (status?.includes('non riuscit') || status?.includes('non inviate')) return 'red'
    if ((status?.startsWith('Istruzioni inviate') || status?.startsWith('Istruzioni senza password inviate')) || status?.startsWith('Conferma inviata') || status?.includes('avviso inviato')) return 'green'
    return 'neutral'
  }

  function inboundInfo(provisioning?: string) {
    if (provisioning === 'Abilitato in Mojo') return { tone: 'green', text: 'Abilitato: il partner può collegarsi al tuo nodo.' }
    if (provisioning === 'In attesa di Mojo') return { tone: 'amber', text: 'Operazione in corso: attendi qualche secondo.' }
    if (provisioning === 'Nodo bloccato in Mojo') return { tone: 'red', text: 'Nodo bloccato in DXSpider' }
    if (provisioning?.includes('verificare')) return { tone: 'red', text: 'Accesso in ingresso sospeso: verifica il tipo e le credenziali locali del nodo.' }
    return { tone: 'neutral', text: 'Non abilitato: il partner non può collegarsi al tuo nodo.' }
  }

  // Stato reale del collegamento verso di te: il canale di DXSpider è uno solo, quindi si distingue chi l'ha aperto.
  function incomingInfo(partner?: Partner) {
    if (!partner) return { tone: 'neutral', text: '' }
    if (partner.Connection === 'Attivo' && partner.Direction === 'Ingresso') return { tone: 'green', text: 'Attivo in ingresso · sessione avviata dal partner' }
    if (partner.Connection === 'Attivo' && partner.Direction === 'Uscita') return { tone: 'neutral', text: 'Sessione attuale avviata dal tuo nodo (in uscita).' }
    return { tone: 'neutral', text: 'Nessuna sessione in ingresso attiva.' }
  }

  function directionLabel(direction?: string) {
    if (direction === 'Uscita') return 'In uscita ↗ (avviato da te)'
    if (direction === 'Ingresso') return 'In ingresso ↙ (avviato dal partner)'
    return '—'
  }

  function connectionLabel(state?: string) {
    if (state === 'Stato Mojo non disponibile') return 'Stato DXSpider non disponibile'
    if (state === 'Account Mojo non abilitato') return 'Accesso del partner non abilitato'
    return state || 'Stato in attesa'
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
      await refreshProvisioning()
    } catch {
      error = 'Impossibile avviare il collegamento.'
    } finally {
      enabling = false
    }
  }

  async function updateReconnect(enabled: boolean) {
    const call = selected?.Call
    if (!call || enabling) return
    enabling = true
    error = ''
    try {
      if (enabled && remoteAuth === 'password' && !/^[A-Za-z0-9]{1,72}$/.test(outgoingPassword)) {
        throw new Error('Inserisci la password fornita dal nodo remoto.')
      }
      const response = await fetch('/api/partners/' + encodeURIComponent(call) + '/reconnect', {
        method: enabled ? 'PUT' : 'DELETE', headers: { 'Content-Type': 'application/json' },
        ...(enabled ? { body: JSON.stringify(remoteAuth === 'password' ? outgoingPassword : '') } : {}),
      })
      if (!response.ok) throw new Error((await response.text()).trim())
      await refreshProvisioning()
    } catch (cause) {
      if (selected?.Call === call) {
        autoReconnect = selected.Reconnect === 'Abilitata'
        error = cause instanceof Error && cause.message ? cause.message : 'Impossibile salvare la riconnessione automatica.'
      }
    } finally {
      enabling = false
    }
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (busy || enabling) return
    const connectNow = !selected && (event.submitter as HTMLButtonElement | null)?.value !== 'directory'
    const password = remoteAuth === 'password' ? outgoingPassword : ''
    if (remoteAuth === 'password' && (connectNow || password) && !/^[A-Za-z0-9]{1,72}$/.test(password)) {
      error = 'Inserisci la password fornita dal nodo remoto.'
      return
    }
    busy = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners' + (selected ? '/' + encodeURIComponent(selected.Call) : ''), {
        method: selected ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...form, Note: note, ConnectNow: connectNow,
          // Omesso = lascia invariata la password salvata; stringa vuota = rimuovila.
          ...(remoteAuth === 'password' ? (password ? { RemotePassword: password } : {}) : storedPassword ? { RemotePassword: '' } : {}),
        }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const saved: Partner = await response.json()
      partners = [...partners.filter(partner => partner.Call !== saved.Call), saved]
      open(saved)
      if (password) {
        remoteAuth = 'password'
        outgoingPassword = password
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
  {#if statusError}<p role="status">{statusError}</p>{/if}
  {#if partners.length}
    <div class="actions partner-filters" role="group" aria-label="Mostra nodi partner">
      <button type="button" class:secondary={onlyInbound} aria-pressed={!onlyInbound} onclick={() => onlyInbound = false}>Tutti i nodi</button>
      <button type="button" class:secondary={!onlyInbound} aria-pressed={onlyInbound} onclick={() => onlyInbound = true}>Nodi collegati in ingresso <span>({statusError ? '—' : inboundPartners.length})</span></button>
    </div>
    {#if onlyInbound && (statusError || !inboundPartners.length)}
      <p class="muted">{statusError ? 'Stato delle sessioni non disponibile.' : 'Nessun nodo collegato in ingresso.'}</p>
    {:else}
      <div class="table-wrap">
        <table>
          <thead><tr><th>Indicativo</th><th scope="col">Paese</th><th>Nome nodo</th><th>Referente</th><th scope="col">Stato</th><th scope="col" title="Chi ha avviato il collegamento">Direzione</th><th>Registrato il</th></tr></thead>
          <tbody>
            {#each (onlyInbound ? inboundPartners : sortedPartners) as partner (partner.Call)}
              {@const flag = dxccFlags[partner.DXCC ?? 0]}
              {@const country = flag ? dxccFlagName(flag, regionNames) : '—'}
              {@const state = statusError ? 'Stato non aggiornato' : partner.Locked ? 'Nodo bloccato in DXSpider' : partner.InboundTestUntil && partner.Connection !== 'Attivo' ? 'Prova in ingresso' : connectionLabel(partner.Connection)}
              <tr class:current={selected?.Call === partner.Call}>
                <td><button class="call" type="button" onclick={() => open(partner)} disabled={busy || enabling}>{partner.Call}</button></td>
                <td><span class="partner-country">
                  {#if flag}<img src={'/flags/' + flag + '.svg'} alt="" width="22" height="16" loading="lazy" />{/if}
                  <span>{country}</span>
                </span></td>
                <td>{partner.Name || '—'}</td>
                <td>{partner.Contact || '—'}</td>
                <td>
                  <span class="observed-state">
                    <span class="led" data-tone={statusError ? 'neutral' : partner.Locked ? 'red' : state === 'Prova in ingresso' ? 'amber' : connectionTone(partner.Connection)}
                      title={state} aria-hidden="true"></span>
                    {state}
                  </span>
                </td>
                <td>{directionLabel(partner.Direction)}</td>
                <td>{date.format(new Date(partner.RegisteredAt))}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {:else if !busy && !error && !statusError}
    <p>Nessun nodo partner registrato.</p>
  {/if}

  {#if editing}
    {@const inboundByIP = Boolean(selected?.CanSetInboundIP && selected?.InboundIP)}
    {@const inbound = inboundInfo(selected?.Locked ? 'Nodo bloccato in Mojo' : inboundByIP ? 'Abilitato in Mojo' : selected?.Provisioning)}
    {@const incoming = incomingInfo(selected)}
    {@const inboundTestWaiting = Boolean(selected?.InboundTestUntil && selected.Connection !== 'Attivo')}
    {@const lockedInMojo = Boolean(selected?.Locked)}
    {@const canEnableInbound = selected && !lockedInMojo && !['In attesa di Mojo', 'Abilitato in Mojo'].includes(selected.Provisioning || '') && !selected.Provisioning?.includes('verificare')}
    {@const outgoingReady = selected?.OutgoingReady ?? (selected?.Provisioning === 'Abilitato in Mojo' || selected?.Provisioning === 'Pronto in uscita')}
    <div class="detail">
      <header class="detail-head">
        <h3 id="partner-draft" tabindex="-1">{selected ? selected.Call : 'Nuovo nodo partner'}</h3>
        {#if selected}
          <p class="observed-state">
            <span class="led" data-tone={inboundTestWaiting ? 'amber' : connectionTone(selected.Connection)} aria-hidden="true"></span>
            {inboundTestWaiting ? 'Prova in ingresso' : connectionLabel(selected.Connection)}
            {#if selected.Direction}<span class="muted">· {directionLabel(selected.Direction)}</span>{/if}
            <span class="muted">· Prima connessione: {selected.FirstConnectedAt ? date.format(new Date(selected.FirstConnectedAt)) : 'non ancora rilevata'}</span>
          </p>
        {/if}
        <div class="head-actions">
          {#if selected}<button class="danger" type="button" disabled={enabling || busy} onclick={removePartner}>Elimina nodo</button>{/if}
          <button class="secondary" type="button" onclick={() => { editing = false; selected = undefined }}>Chiudi</button>
        </div>
      </header>

      {#if selected}
        <div class="cards">
          <article class="card">
            <h4><span class="step">1</span> Il tuo nodo si collega a {selected.Call}</h4>
            <p class="hint">Usa questa sezione per collegarti tu al nodo del partner (ad esempio per ricevere i suoi spot).</p>
            <p class="observed-state" role="status">
              <span class="led" data-tone={inboundTestWaiting ? 'amber' : connectionTone(selected.Connection)} aria-hidden="true"></span>
              {inboundTestWaiting ? 'Uscita sospesa durante la prova in ingresso' : connectionLabel(selected.Connection)}
            </p>
            {#if !inboundTestWaiting && selected.Connection === 'Collegamento in corso' && selected.Direction === 'Uscita'}
              <p class="hint">In attesa della risposta del nodo remoto. Se l’attesa continua, chiedi al referente di verificare che il tuo indicativo completo, SSID incluso, sia configurato come nodo partner e che la richiesta di password corrisponda.</p>
            {:else if !inboundTestWaiting && selected.Connection === 'Collegamento non confermato'}
              <p class="hint">La sessione tra i nodi non è stata confermata. Cerca l’indicativo nel Log DXSpider per vedere a quale fase si è fermato il tentativo.</p>
            {/if}
            {#if selected.Reconnect}
              <p class="muted">Riconnessione automatica: {selected.Reconnect}.
                {#if selected.LastConnectAttempt}Ultimo tentativo {new Date(selected.LastConnectAttempt).toLocaleString('it-IT', { timeZone: 'UTC' })} UTC.{/if}
                {#if selected.RetryEligibleAt}Prossimo tentativo dopo le {new Date(selected.RetryEligibleAt).toLocaleString('it-IT', { timeZone: 'UTC' })} UTC.{/if}
              </p>
            {/if}
            <form onsubmit={connect}>
              <div class="actions">
                {#if lockedInMojo}
                  {#if selected.CanUnlock}
                    <button class="unlock-node" type="button" disabled={enabling || busy} onclick={() => enable(true, false, true)}>Sblocca questo nodo</button>
                  {/if}
                  <span class="hint">DXSpider ha un blocco locale per questo indicativo. Può essere stato impostato dal sysop o applicato quando DXSpider ha scoperto il nodo in rete.</span>
                {:else if selected.Connection !== 'Attivo' && outgoingReady}
                  <button disabled={enabling || busy || inboundTestWaiting || selected.Connection === "In coda" || selected.Connection === "Collegamento in corso"}>Connetti ora</button>
                {:else if selected.Connection !== 'Attivo'}
                  <button type="button" disabled={enabling || busy || selected.Provisioning === 'In attesa di Mojo'}
                    onclick={() => enable(true)}>Prepara il collegamento</button>
                  <span class="hint">Prima va preparato l’account sul tuo nodo; poi compare «Connetti ora».</span>
                {/if}
              </div>
              <label>Il nodo remoto richiede una password?
                <select bind:value={remoteAuth} disabled={enabling || busy}>
                  <option value="none">No, senza password</option>
                  <option value="password">Sì, ha una password</option>
                </select>
              </label>
              {#if remoteAuth === 'password'}
                <label>Password che ti ha dato il nodo remoto
                  <input type="text" bind:value={outgoingPassword} pattern={'[A-Za-z0-9]{1,72}'} maxlength="72" autocomplete="off" spellcheck="false" required />
                </label>
              {/if}
              <label class="check"><input type="checkbox" checked={autoReconnect} onchange={event => {
                const input = event.currentTarget
                void updateReconnect(input.checked).finally(() => { input.checked = autoReconnect })
              }} disabled={enabling || busy || checkingStatus || loadingCredential} /> Riprova da solo se il collegamento cade</label>
            </form>
            <details class="script-preview">
              <summary>Mostra lo script di connessione</summary>
              <pre>{[
                'timeout 15',
                'connect telnet ' + selected.Host + ' ' + selected.Port,
                "'login' '" + (localNode || 'INDICATIVO NODO LOCALE') + "'",
                ...(remoteAuth === 'password' ? ["'password:' '••••••••'"] : []),
                'client ' + selected.Call + ' telnet',
              ].join('\n')}</pre>
            </details>
          </article>

          <article class="card">
            <h4><span class="step">2</span> {selected.Call} si collega al tuo nodo</h4>
            <p class="hint">Serve solo se il partner deve collegarsi lui al tuo server. Gli dai host, porta e una password dedicata.</p>
            <p class="observed-state" role="status">
              <span class="led" data-tone={inbound.tone} aria-hidden="true"></span>
              {inbound.text}
            </p>
            {#if incoming.text}
              <p class="observed-state" role="status">
                <span class="led" data-tone={incoming.tone} aria-hidden="true"></span>
                {incoming.text}
              </p>
              <p class="muted">Prima connessione in ingresso: {selected.FirstInboundAt ? date.format(new Date(selected.FirstInboundAt)) : 'non ancora rilevata'}</p>
            {/if}
            <label>Autenticazione in ingresso
              <select bind:value={inboundMode} disabled={enabling || busy}>
                <option value="password">Con password</option>
                <option value="ip">Senza password · IP o nome host autorizzato</option>
              </select>
            </label>
            {#if inboundMode === 'ip'}
              {#if !selected.CanSetInboundIP}
                <p class="hint">Prepara o sblocca l’account locale del partner per attivare l’accesso tramite IP.</p>
              {/if}
              <form onsubmit={event => {
                event.preventDefault()
                void saveInboundIP((event.submitter as HTMLButtonElement)?.value === 'send')
              }}>
                <label>IP pubblico o nome host del partner
                  <input bind:value={inboundIP} required maxlength="253" spellcheck="false" autocomplete="off" disabled={!selected.CanSetInboundIP || enabling} />
                </label>
                <p class="hint">Il campo riprende l’host salvato nella scheda del nodo. Verifica che punti all’IP pubblico da cui esce il partner. Gli indirizzi del nome host si aggiornano automaticamente.</p>
                <div class="actions">
                  {#if selected.Email}
                    <button type="submit" value="send" disabled={enabling || busy || !selected.CanSetInboundIP}>Salva e invia senza password</button>
                  {/if}
                  <button class="secondary" type="submit" value="save" disabled={enabling || busy || !selected.CanSetInboundIP}>Salva senza inviare</button>
                </div>
              </form>
              {#if ipMailStatus?.error || (ipMailStatus && enabling) || selected.IPMailStatus || ipMailStatus}
                <p class="mail-notice" data-tone={ipMailStatus?.error ? 'red' : enabling ? 'amber' : mailTone(selected.IPMailStatus ?? 'Istruzioni inviate')} role="status">
                  {ipMailStatus && (ipMailStatus.error || enabling) ? ipMailStatus.text : selected.IPMailStatus || ipMailStatus?.text}
                </p>
              {/if}
            {/if}
            {#if selected.InboundAutoNext}
              <p class="muted">Prova automatica attiva.
                {#if selected.Connection === 'Attivo' && selected.Direction === 'Ingresso'}
                  Il partner è collegato in ingresso.
                {:else}
                  Prossima finestra indicativa {date.format(new Date(selected.InboundAutoNext))}.
                {/if}
              </p>
            {/if}
            {#if selected.InboundTestUntil && !(selected.Connection === 'Attivo' && selected.Direction === 'Ingresso')}
              <div class="mail-notice" data-tone="amber" role="status">
                Prova in ingresso fino alle {date.format(new Date(selected.InboundTestUntil))}. Il partner può tentare il collegamento adesso.
              </div>
            {/if}
            {#if selected.PublicMailAcceptedAt}
              <div class="mail-notice" data-tone="green" role="status">
                <span class="led" data-tone="green" aria-hidden="true"></span>
                <span><strong>Email con dati pubblici</strong> · Accettata da SMTP il {date.format(new Date(selected.PublicMailAcceptedAt))}.</span>
              </div>
            {/if}
            {#if !inboundByIP && (selected.MailStatus || emailStatus)}
              {@const tone = mailTone(selected.MailStatus || emailStatus, selected.MailRetryAt)}
              <div class="mail-notice" data-tone={tone} role="status">
                <span class="led" data-tone={tone} aria-hidden="true"></span>
                <span><strong>Email con credenziali</strong> · {selected.MailStatus || emailStatus}
                  {#if selected.MailRetryAt} · Prossimo tentativo {date.format(new Date(selected.MailRetryAt))}{/if}
                </span>
              </div>
            {/if}
            {#if !inboundByIP && peerPassword && selected.Provisioning === 'Abilitato in Mojo'}
              <p>Password del partner:</p>
              <code>{peerPassword}</code>
              <div class="actions">
                <button class="secondary" type="button" onclick={copyPassword}>Copia</button>
                <button class="secondary" type="button" onclick={() => { peerPassword = '' }}>Nascondi</button>
              </div>
            {:else if !inboundByIP && selected.Provisioning === 'Abilitato in Mojo' && selected.CanResend}
              <div class="actions"><button class="secondary" type="button" onclick={revealPassword}>Mostra la password del partner</button></div>
            {/if}
            <div class="actions">
              {#if inboundMode === 'password' && (selected.CanResetCCInbound || inboundByIP) && selected.Email}
                <button type="button" disabled={enabling || busy} onclick={() => openPartnerConfirmation('password')}
                  title="Sostituisce la password locale del partner e la invia dopo l’abilitazione.">{inboundByIP ? 'Passa alla password e invia l’email' : 'Imposta nuova password e invia l’email'}</button>
              {:else if inboundMode === 'password' && selected.Provisioning !== 'Abilitato in Mojo' && selected.Email && !inboundByIP}
                <button type="button" disabled={!canEnableInbound || enabling || busy} onclick={() => enable(false, true)}>Abilita e invia l’email a {selected.Email}</button>
                {#if canEnableInbound}
                  <button class="secondary" type="button" disabled={enabling || busy} onclick={() => enable()}>Abilita senza email</button>
                {/if}
              {:else if inboundMode === 'password' && canEnableInbound && !inboundByIP}
                <button type="button" disabled={enabling || busy} onclick={() => enable()}>Abilita l’accesso</button>
                <span class="hint">Aggiungi l’email del referente in «Dati del nodo» per poter inviare le istruzioni.</span>
              {/if}
              {#if inboundMode === 'password' && selected.Provisioning === 'Abilitato in Mojo' && !inboundByIP}
                <button type="button" disabled={enabling || busy || !selected.CanResend} onclick={resend}>Reinvia le istruzioni via email</button>
                {#if !selected.CanResend}<span class="hint">Per reinviare servono l’email del referente e una password conservata da questa versione in poi.</span>{/if}
              {/if}
              {#if (selected.Provisioning === 'Abilitato in Mojo' || inboundByIP) && !selected.InboundTestUntil && !(selected.Connection === 'Attivo' && selected.Direction === 'Ingresso')}
                <button class="secondary" type="button" disabled={enabling || busy} onclick={() => openPartnerConfirmation('manual')}
                  title="Apre una finestra in ingresso di cinque minuti.">Prova ingresso ora</button>
              {/if}
              {#if selected.InboundAutoNext}
                <button class="secondary" type="button" disabled={enabling || busy} onclick={() => startInboundTest(false, true)}
                  title="Interrompe la programmazione delle prossime finestre.">Disattiva prova automatica</button>
              {:else if selected.Provisioning === 'Abilitato in Mojo' || inboundByIP}
                <button class="secondary" type="button" disabled={enabling || busy} onclick={() => openPartnerConfirmation('automatic')}
                  title="Programma finestre in ingresso anche a portale chiuso.">Attiva prova automatica</button>
              {/if}
              {#if selected.Email}
                <button class="secondary" type="button" disabled={enabling || busy} onclick={sendNodeDetails}
                  title="Invia indicativo, host e porta. L’accesso e le password restano invariati.">{publicMail?.call === selected.Call && publicMail.state === 'sending' ? 'Invio in corso…' : 'Invia dati pubblici senza password'}</button>
              {/if}
              <button class="secondary" type="button" disabled={enabling || busy || checkingStatus} onclick={async () => {
                manualStatusCall = selected?.Call ?? ''
                await refreshProvisioning()
              }}>Aggiorna stato</button>
              {#if publicMail?.call === selected.Call}
              <div class="mail-notice public-mail-result" data-tone={publicMail.state === 'accepted' ? 'green' : publicMail.state === 'failed' ? 'red' : 'amber'} role="status">
                <span class="led" data-tone={publicMail.state === 'accepted' ? 'green' : publicMail.state === 'failed' ? 'red' : 'amber'} aria-hidden="true"></span>
                <span>{publicMail.text}</span>
              </div>
            {/if}
            {#if manualStatusCall === selected.Call}
                <span class="hint" role="status">
                  {checkingStatus ? 'Controllo dello stato in corso.' : statusError || `Ultimo stato: ${inbound.text}`}
                </span>
              {/if}
            </div>
          </article>
        </div>
      {/if}

      <article class="card">
        <h4>{selected ? 'Dati del nodo' : 'Dati del nuovo nodo'}</h4>
        <form onsubmit={save}>
          <div class="fields">
            <label>Indicativo del nodo (SSID incluso)
              <input bind:value={form.Call} maxlength="12" placeholder="EA7JXH-6" aria-describedby="partner-call-hint" readonly={Boolean(selected)} required />
            </label>
            <label>Nome del nodo <input bind:value={form.Name} maxlength="120" /></label>
            <label>Nome del referente <input bind:value={form.Contact} maxlength="120" /></label>
            <label>Nome della persona <input bind:value={form.Person} maxlength="120" autocomplete="off" placeholder="Es. Anthony" /></label>
            <label>Email del referente <input type="email" bind:value={form.Email} maxlength="254" /></label>
            <label>Lingua delle email
              <select bind:value={form.Lang}>
                <option value="">Inglese (predefinita)</option>
                <option value="it">Italiano</option><option value="en">English</option><option value="es">Español</option>
                <option value="de">Deutsch</option><option value="fr">Français</option><option value="pt">Português</option>
              </select>
            </label>
            <label>Host del nodo <input bind:value={form.Host} maxlength="253" required /></label>
            <label>Porta <input type="number" bind:value={form.Port} min="1" max="65535" required /></label>
          </div>
          <p class="hint" id="partner-call-hint">Per il collegamento usa l’indicativo completo del nodo remoto, con il suffisso se presente. Il nome del nodo è un’etichetta facoltativa.</p>
          <p class="hint">«Nome della persona» serve solo per la nostra anagrafica: non entra nella configurazione del nodo.</p>
          {#if !selected}
            <label class="remote-auth">Il nodo remoto richiede una password?
              <select bind:value={remoteAuth}>
                <option value="none">No, senza password</option>
                <option value="password">Sì, ha una password</option>
              </select>
            </label>
            {#if remoteAuth === 'password'}
              <label class="remote-auth">Password del nodo remoto
                <input type="text" bind:value={outgoingPassword} maxlength="72" autocomplete="off" spellcheck="false" />
              </label>
            {/if}
          {/if}
          <label class="note">Nota <textarea bind:value={note} maxlength="4000" rows="3" placeholder="Es. Inviata email al referente"></textarea></label>
          <div class="actions">
            <button type="submit" value={selected ? 'directory' : 'connect'} disabled={busy || enabling}>
              {busy ? 'Salvataggio…' : selected ? 'Salva dati' : 'Salva e connetti'}
            </button>
            {#if !selected}
              <button class="secondary" type="submit" value="directory" disabled={busy || enabling}>Salva solo anagrafica</button>
            {/if}
          </div>
        </form>
      </article>

      {#if selected?.Notes?.length}
        <article class="card">
          <h4>Storico note</h4>
          <ul>
            {#each selected.Notes.toReversed() as entry}
              <li><time datetime={entry.At}>{date.format(new Date(entry.At))}</time><p>{entry.Text}</p></li>
            {/each}
          </ul>
        </article>
      {/if}
    </div>
  {/if}
</section>

<dialog class="glass-panel password-dialog" bind:this={passwordDialog} onclose={finishPasswordReset}
  aria-labelledby="partner-password-title" aria-describedby="partner-password-description">
  <h2 id="partner-password-title">{confirmation === 'password' ? 'Nuova password Telnet' : confirmation === 'automatic' ? 'Prova automatica in ingresso' : 'Prova in ingresso'}</h2>
  <p id="partner-password-description">{#if confirmation === 'password'}
      {tr('La password Telnet di {} verrà sostituita. Il partner riceverà quella nuova via email.', passwordTarget)}
      {#if selected?.InboundIP}L’accesso per IP verrà rimosso dopo la configurazione della password.{/if}
    {:else if confirmation === 'automatic'}
      Finché il partner è scollegato in ingresso, ogni 30 minuti DXRegia sospende l’uscita per 5 minuti. La modalità resta attiva anche a portale chiuso e dopo un riavvio.
    {:else}
      DXRegia sospende l’uscita per 5 minuti. Il partner può tentare il collegamento in quella finestra.
    {/if}</p>
  <form method="dialog" class="actions">
    <button class="secondary" value="cancel" autofocus>Annulla</button>
    <button value="confirm">{confirmation === 'password' ? 'Imposta e invia' : confirmation === 'automatic' ? 'Attiva' : 'Avvia prova'}</button>
  </form>
</dialog>

<style>
  .password-dialog { width: min(440px, calc(100% - 32px)); box-sizing: border-box; padding: 24px; color: var(--glass-text); }
  .password-dialog::backdrop { background: #0b2238a0; backdrop-filter: blur(3px); }
  .password-dialog .actions { justify-content: flex-end; }

  .mail-notice { display: flex; align-items: baseline; gap: .5rem; padding: .65rem .75rem; border: 1px solid #b4cbd8; border-radius: 6px; }
  .public-mail-result { flex-basis: 100%; }
  .mail-notice[data-tone='green'] { border-color: #7db59d; background: #e8f6ef; color: #145b3d; }
  .mail-notice[data-tone='red'] { border-color: #d8898b; background: #fff0f0; color: #8c2127; }
  .mail-notice[data-tone='amber'] { border-color: #d7ad67; background: #fff5de; color: #764800; }

  .unlock-node { background: #a91e24; border-color: #84191e; color: #fff; }
  section { margin-top: 20px; padding: 16px; }
  .toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
  h2 { margin: 0; font-size: 1.15rem; }
  h3 { margin: 0; font-size: 1.2rem; }
  h4 { display: flex; align-items: center; gap: 8px; margin: 0 0 6px; font-size: 1rem; }
  .partner-filters { margin-top: 12px; }
  .table-wrap { overflow: auto; margin-top: 12px; }
  table { width: 100%; text-align: left; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 5px 8px; border-bottom: 1px solid #d7dfe7; }
  th { background: #edf3f9; }
  .partner-country { display: inline-flex; align-items: center; gap: 8px; }
  .partner-country img { flex: none; object-fit: cover; border: 1px solid #17384c26; border-radius: 2px; }
  tr.current td { background: #e4eef7; }
  .detail { margin-top: 18px; display: grid; gap: 14px; }
  .detail-head { display: flex; align-items: center; flex-wrap: wrap; gap: 6px 14px; }
  .detail-head p { margin: 0; }
  .head-actions { display: flex; gap: 10px; margin-left: auto; }
  button.danger { --button-face: #fff1f1; --button-edge: #d99a9d; --button-ink: #9b2424; }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr)); gap: 14px; }
  .cards { align-items: start; }
  .cards .card { display: flex; flex-direction: column; }
  .cards .card > .actions { padding-top: 10px; }
  .card { padding: 14px 16px; border: 1px solid #cdd9e4; border-radius: 8px; background: #ffffffb8; }
  .card p { margin: 6px 0; }
  .card form { display: grid; gap: 10px; margin-top: 8px; }
  .step { display: inline-grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: #214e80; color: #fff; font-size: .85rem; }
  .hint, .muted { color: #485b6d; font-size: .9rem; font-weight: 450; }
  .actions { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
  .check { display: flex; align-items: center; gap: 8px; font-weight: 450; }
  .check input { min-height: auto; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); gap: 14px 16px; align-items: start; }
  .fields label { align-content: start; }
  .fields input, .fields select { width: 100%; }
  label { display: grid; min-width: 0; gap: 5px; font-weight: 600; }
  .note { margin: 12px 0; }
  .observed-state { display: flex; align-items: center; flex-wrap: wrap; gap: .55rem; }
  .led { width: 11px; height: 11px; flex: 0 0 11px; border-radius: 50%; background: #708391; box-shadow: 0 0 0 3px #70839123; }
  .led[data-tone=green] { background: #167b50; box-shadow: 0 0 0 3px #167b5029; }
  .led[data-tone=red] { background: #b32731; box-shadow: 0 0 0 3px #b3273129; }
  .led[data-tone=amber] { background: #ad6a10; box-shadow: 0 0 0 3px #ad6a1029; }
  .script-preview { margin-top: 10px; }
  .script-preview pre { padding: .65rem .9rem; background: #edf4f8; overflow: auto; }
  .remote-auth { margin-top: 12px; max-width: 300px; }
  input, textarea, select { min-width: 0; padding: 8px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-height: 44px; }

  .call { background: none; color: #214e80; border: 0; padding: 4px; min-height: 34px; text-decoration: underline; box-shadow: none; }
  :is(button,input,textarea):focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
  [role=alert] { color: #9b2424; }
  ul { padding: 0; list-style: none; margin: 0; }
  li { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  li p { white-space: pre-wrap; overflow-wrap: anywhere; margin: 3px 0; }
  code { overflow-wrap: anywhere; }
  time { color: #485b6d; font-size: .85rem; }
</style>
