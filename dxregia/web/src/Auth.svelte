<script lang="ts">
  import { onMount } from 'svelte'
  import type { Snippet } from 'svelte'
  import { toDataURL } from 'qrcode'
  import WebSettings from './WebSettings.svelte'
  import TelnetOptions from './TelnetOptions.svelte'
  import MojoLog from './MojoLog.svelte'
  import TelnetConsole from './TelnetConsole.svelte'
  import PasswordSettings from './PasswordSettings.svelte'
  import Setup from './Setup.svelte'
  import Partners from './Partners.svelte'
  import BackupVerify from './BackupVerify.svelte'
  import MailSettings from './MailSettings.svelte'
  import Logo from './Logo.svelte'
  import Analytics from './Analytics.svelte'
  import ContactRequests from './ContactRequests.svelte'
  import type { ContactRequest } from './ContactRequests.svelte'
  let draft = $state<ContactRequest>()

  type Session = { signedIn: boolean; mustChange: boolean; mustVerify?: boolean; enrollmentKey?: string; recoveryCodes?: string[] }
  let session = $state<Session>()
  let { onSaved, nodeStatus }: { onSaved: () => void; nodeStatus: Snippet } = $props()
  const groups = [
    { label: 'Nodo', items: [['stato', 'Stato e sessioni'], ['configurazione', 'Configurazione'], ['accesso-web', 'Accesso web'], ['console', 'Console Telnet'], ['opzioni-telnet', 'Opzioni Telnet'], ['log-dxspider', 'Log DXSpider']] },
    { label: 'Collegamenti', items: [['partner', 'Nodi partner'], ['richieste', 'Richieste ricevute']] },
    { label: 'Gestione', items: [['analytics', 'Analytics'], ['posta', 'Server di posta'], ['backup', 'Backup'], ['password', 'Password']] },
  ] as const
  type Page = typeof groups[number]['items'][number][0]
  let page = $state<Page>('stato')

  const groupColor: Record<string, string> = { Nodo: '#4a9be0', Collegamenti: '#2fc58a', Gestione: '#f0a23a' }
  const readJson = (key: string): Record<string, boolean> => {
    try { return JSON.parse(localStorage.getItem(key) ?? '{}') } catch { return {} }
  }
  let collapsed = $state<Record<string, boolean>>(readJson('dxregia-menu'))
  const isOpen = (group: typeof groups[number]) => !collapsed[group.label] || group.items.some(([id]) => id === page)
  function toggleGroup(label: string) {
    collapsed[label] = !collapsed[label]
    try { localStorage.setItem('dxregia-menu', JSON.stringify(collapsed)) } catch { /* solo una comodità */ }
  }

  // Segnali del menu: partner collegati, richieste non lette, stato del nodo.
  let partnersUp = $state(0)
  let partnersTotal = $state(0)
  let requestsTotal = $state(0)
  let seen = $state(Number((() => { try { return localStorage.getItem('dxregia-requests-seen') } catch { return '0' } })() ?? 0) || 0)
  let nodeCall = $state('')
  let listenerOk = $state<boolean>()
  let unread = $derived(Math.max(0, requestsTotal - seen))
  let inPanel = $derived(Boolean(session?.signedIn && !session.mustChange && !session.mustVerify))

  // Nella scheda di accesso il logo grande è già al centro: il logo piccolo nell'intestazione si nasconde.
  $effect(() => {
    document.documentElement.classList.toggle('auth-card-view', Boolean(session) && !inPanel)
    return () => document.documentElement.classList.remove('auth-card-view')
  })

  $effect(() => {
    if (page === 'richieste' && seen !== requestsTotal) {
      seen = requestsTotal
      try { localStorage.setItem('dxregia-requests-seen', String(seen)) } catch { /* solo una comodità */ }
    }
  })

  async function loadSide() {
    try {
      const [partners, requests, node, listener] = await Promise.all([
        fetch('/api/partners', { cache: 'no-store' }), fetch('/api/contact-requests', { cache: 'no-store' }),
        fetch('/api/node-settings', { cache: 'no-store' }), fetch('/api/listener', { cache: 'no-store' }),
      ])
      if (partners.ok) {
        const rows: { Connection?: string }[] = Object.values(await partners.json())
        partnersTotal = rows.length
        partnersUp = rows.filter(row => row.Connection === 'Attivo').length
      }
      if (requests.ok) requestsTotal = ((await requests.json()) ?? []).length
      if (node.ok) nodeCall = (await node.json()).node || ''
      if (listener.ok) listenerOk = (await listener.json()).state === 'ready'
    } catch { /* il menu resta com'è */ }
  }

  $effect(() => {
    if (!inPanel) return
    void loadSide()
    const timer = setInterval(loadSide, 10000)
    return () => clearInterval(timer)
  })

  function selectPage(event: MouseEvent, next: Page) {
    if (event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return
    event.preventDefault()
    if (location.hash !== '#' + next) location.hash = next
    page = next
  }

  function readPage() {
    const target = location.hash.slice(1)
    for (const group of groups) {
      for (const [id] of group.items) {
        if (id === target) {
          page = id
          return
        }
      }
    }
    page = 'stato'
  }
  let password = $state('')
  let user = $state('')
  let error = $state('')
  let busy = $state(false)
  let recovery = $state(false)
  let codes = $state<string[]>([])
  let qr = $state('')
  let challenge = $derived(session?.mustVerify && !session.mustChange)

  $effect(() => {
    const secret = session?.enrollmentKey
    qr = ''
    if (!secret) return
    let current = true
    const uri = 'otpauth://totp/DXRegia:sysop?secret=' + encodeURIComponent(secret) + '&issuer=DXRegia'
    void toDataURL(uri, { width: 180, margin: 3 }).then(value => {
      if (current) qr = value
    }).catch(() => {})
    return () => { current = false }
  })

  function toggleRecovery() {
    recovery = !recovery
    password = ''
    error = ''
  }

  onMount(async () => {
    readPage()
    try {
      const response = await fetch('/api/auth')
      if (response.ok && response.headers.get('content-type')?.includes('application/json')) {
        session = await response.json()
      }
    } catch {
      // The listener diagnostic is also available without setup mode.
    }
  })

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (!session || busy) return
    busy = true
    error = ''
    try {
      const endpoint = session.mustChange ? 'password' : session.mustVerify ? (recovery ? 'recovery' : 'totp') : 'login'
      const response = await fetch('/api/auth/' + endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(endpoint === 'login' ? { User: user, Password: password } : password),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const next: Session = await response.json()
      codes = next.recoveryCodes ?? []
      session = next
      recovery = false
      password = ''
    } catch {
      error = 'Il servizio di accesso non risponde.'
    } finally {
      busy = false
    }
  }

  async function logout() {
    error = ''
    try {
      const response = await fetch('/api/auth/logout', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
      })
      if (!response.ok) throw new Error('Logout failed')
      session = await response.json()
    } catch {
      error = 'Impossibile terminare la sessione.'
    }
  }
</script>

<svelte:window onhashchange={readPage} />

{#if session}
  {#if codes.length}
    <section class="auth-card glass-panel glass-enter">
      <h2>Salva i codici di recupero</h2>
      <p>Conserva questi codici in un posto sicuro. Un codice e la password permettono di configurare una nuova app di autenticazione.</p>
      <ul class="recovery-codes">
        {#each codes as code}<li><code>{code}</code></li>{/each}
      </ul>
      <button type="button" onclick={() => { codes = []; if (session) session.recoveryCodes = undefined }}>Ho salvato i codici</button>
    </section>
  {:else if session.signedIn && !session.mustChange && !session.mustVerify}
    <div class="workspace">
      <aside class="sidebar">
        <p class="sidebar-title">Amministrazione</p>
        <nav aria-label="Amministrazione DXRegia">
          {#each groups as group (group.label)}
            <div class="nav-group" style:--c={groupColor[group.label]}>
              <button class="group-head" type="button" aria-expanded={isOpen(group)} onclick={() => toggleGroup(group.label)}>
                <span class="chev" class:open={isOpen(group)} aria-hidden="true"></span>
                <span class="dot" aria-hidden="true"></span>{group.label}
              </button>
              {#if isOpen(group)}
                <ul>
                  {#each group.items as [id, label]}
                    <li><a href={'#' + id} aria-current={page === id ? 'page' : undefined}
                      onclick={(event) => selectPage(event, id)}>
                      {label}
                      {#if id === 'partner' && partnersTotal}
                        <span class="led" data-tone={partnersUp === partnersTotal ? 'green' : partnersUp ? 'amber' : 'red'}
                          title={partnersUp + ' su ' + partnersTotal + ' collegati'}></span>
                      {:else if id === 'richieste' && unread}
                        <span class="count" title="Richieste non lette">{unread}</span>
                      {/if}
                    </a></li>
                  {/each}
                </ul>
              {/if}
            </div>
          {/each}
        </nav>
        <div class="side-foot">
          <p class="node"><span class="led" data-tone={listenerOk === undefined ? 'neutral' : listenerOk ? 'green' : 'red'}></span>{nodeCall || 'Nodo'} · {listenerOk === undefined ? '…' : listenerOk ? 'online' : 'offline'}</p>
          <button class="secondary" type="button" onclick={logout}>Esci</button>
        </div>
        {#if error}<p role="alert">{error}</p>{/if}
      </aside>
      {#key page}
      <div class="workspace-content glass-enter">
        {#if page === 'stato'}
          {@render nodeStatus()}
        {:else if page === 'configurazione'}
          <Setup {onSaved} />
        {:else if page === 'accesso-web'}
          <WebSettings />
        {:else if page === 'console'}
          <TelnetConsole />
        {:else if page === 'opzioni-telnet'}
          <TelnetOptions />
        {:else if page === 'log-dxspider'}
          <MojoLog />
        {:else if page === 'partner'}
          <Partners {draft} />
        {:else if page === 'richieste'}
          <ContactRequests onDraft={request => {
            draft = { ...request }
            location.hash = 'partner'
            page = 'partner'
          }} />
        {:else if page === 'analytics'}
          <Analytics />
        {:else if page === 'posta'}
          <MailSettings />
        {:else if page === 'backup'}
          <BackupVerify />
        {:else if page === 'password'}
          <PasswordSettings />
        {/if}
      </div>
      {/key}
    </div>
  {:else}
    <section class="auth-card glass-panel glass-enter">
      <Logo />
      {#if session.mustChange || session.mustVerify}
        <h2>{session.mustChange ? 'Cambia la password temporanea' : 'Verifica in due passaggi'}</h2>
      {/if}
      {#if session.enrollmentKey}
        <p>Aggiungi DXRegia alla tua app di autenticazione.</p>
        {#if qr}<img src={qr} alt="Codice QR per configurare DXRegia nell’app di autenticazione" width="180" height="180" />{/if}
        <details><summary>Inserisci la chiave a mano</summary><p><code>{session.enrollmentKey}</code></p></details>
      {/if}
      <form onsubmit={submit}>
        {#if !session.mustChange && !challenge}
          <label>
            Utente
            <input type="text" required bind:value={user} autocomplete="username" autocapitalize="off" spellcheck="false" maxlength="64" />
          </label>
        {/if}
        <label>
          {session.mustChange ? 'Nuova password' : challenge ? (recovery ? 'Codice di recupero' : 'Codice a sei cifre') : 'Password'}
          <input type={challenge && !recovery ? 'text' : 'password'} required bind:value={password}
            inputmode={challenge && !recovery ? 'numeric' : undefined}
            pattern={challenge ? (recovery ? '[0-9a-f]{48}' : '[0-9]{6}') : undefined}
            autocomplete={session.mustChange ? 'new-password' : session.mustVerify ? 'one-time-code' : 'current-password'}
            minlength={session.mustChange ? 12 : undefined} />
        </label>
        {#if session.mustChange}<p>Usa almeno 12 caratteri. Dopo il cambio imposterai il secondo fattore.</p>{/if}
        {#if error}<p role="alert">{error}</p>{/if}
        <button disabled={busy}>{session.mustChange ? 'Salva password' : session.mustVerify ? 'Verifica codice' : 'Accedi'}</button>
      </form>
      {#if challenge && !session.enrollmentKey}
        <button type="button" class="secondary" onclick={toggleRecovery} disabled={busy}>
          {recovery ? 'Usa l’app autenticatore' : 'Usa un codice di recupero'}
        </button>
      {/if}
    </section>
  {/if}
{/if}

<style>
  .auth-card { max-width: 440px; margin: 40px auto; padding: clamp(22px, 5vw, 36px); text-align: center; }
  .auth-card h2 { margin: 6px 0 22px; font-size: 1.25rem; text-align: center; }
  .auth-card form { display: grid; gap: 18px; text-align: left; max-width: 340px; margin: 0 auto; }
  .auth-card label { display: grid; gap: 8px; font-weight: 650; }
  .auth-card input { width: 100%; min-height: 46px; padding: 10px 12px; font: inherit; }
  .auth-card form button { justify-self: center; width: auto; min-width: 120px; min-height: 40px; padding: 0 28px; margin-top: 4px; }
  .auth-card > button { margin-top: 14px; }
  .auth-card img { display: block; margin: 10px auto; }
  .auth-card details, .auth-card > p { text-align: left; }
  .workspace { display: grid; grid-template-columns: clamp(190px, 19vw, 264px) minmax(0, 1fr); align-items: start; gap: clamp(14px, 2vw, 28px); }
  .workspace-content { min-width: 0; }
  .sidebar { position: sticky; top: 18px; display: flex; flex-direction: column; padding: 16px 12px; min-height: calc(100vh - 112px); border-radius: 14px; background: #0f2438; color: #c9d9e8; box-shadow: 0 8px 24px #0f243833; }
  .sidebar-title { margin: 0 0 14px; padding-inline: .7rem; font-size: 1.02rem; color: #ffffff; font-weight: 750; letter-spacing: .02em; }
  .sidebar nav { flex: 1; }
  .sidebar ul { margin: 0; padding: 0; list-style: none; }
  .nav-group + .nav-group { margin-top: 10px; }
  .group-head { display: flex; align-items: center; gap: 8px; width: 100%; min-height: 34px; padding: 4px 8px; border: 0; border-radius: 6px; background: none; box-shadow: none; color: var(--c); font: inherit; font-size: .78rem; font-weight: 700; letter-spacing: .07em; text-transform: uppercase; cursor: pointer; text-align: left; }
  .group-head:hover { background: #ffffff12; }
  .chev { width: 7px; height: 7px; border-right: 2px solid currentColor; border-bottom: 2px solid currentColor; transform: rotate(-45deg); transition: transform .15s; }
  .chev.open { transform: rotate(45deg); }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--c); }
  .nav-group a { display: flex; align-items: center; gap: 8px; min-height: 38px; margin: 1px 0 1px 14px; padding: .4rem .65rem; text-decoration: none; color: #c9d9e8; border-inline-start: 2px solid #ffffff1f; border-radius: 0 7px 7px 0; }
  .nav-group a:hover { background: #ffffff12; color: #ffffff; }
  .nav-group a[aria-current=page] { border-inline-start-color: var(--c); background: color-mix(in srgb, var(--c) 24%, transparent); color: #ffffff; font-weight: 700; }
  .led { display: inline-block; width: 9px; height: 9px; margin-left: auto; border-radius: 50%; background: #708391; flex: 0 0 9px; }
  .led[data-tone=green] { background: #4be0a0; box-shadow: 0 0 8px #4be0a099; }
  .led[data-tone=amber] { background: #f0a23a; box-shadow: 0 0 8px #f0a23a99; }
  .led[data-tone=red] { background: #f0646d; box-shadow: 0 0 8px #f0646d99; }
  .count { margin-left: auto; min-width: 20px; padding: 0 6px; border-radius: 10px; background: #f0a23a; color: #412402; font-size: .78rem; font-weight: 700; text-align: center; }
  .side-foot { display: grid; gap: 10px; padding-top: 12px; border-top: 1px solid #ffffff1f; }
  .node { display: flex; align-items: center; gap: 8px; margin: 0; padding-inline: .5rem; font-size: .88rem; color: #9fc2e0; }
  .node .led { margin-left: 0; }
  .sidebar .secondary { background: #ffffff14; color: #ffffff; border: 1px solid #ffffff2e; box-shadow: none; }
  .sidebar .secondary:hover { background: #ffffff24; }
  .sidebar [role=alert] { color: #ffb4b8; }
  @media (max-width: 650px) {
    .workspace { grid-template-columns: minmax(0, 1fr); }
    .sidebar { position: static; min-height: 0; }
  }
  h2 { margin: 0 0 12px; font-size: 1.15rem; }
  label { display: grid; max-width: 330px; gap: 5px; font-weight: 600; }
  input { border: 1px solid #aab8c5; border-radius: 4px; min-height: 44px; padding: 8px; font: inherit; }
  button { margin-top: 14px; }
  :is(button, input):focus-visible { outline: 2px solid #214e80; outline-offset: 3px; }
  .recovery-codes { padding-left: 20px; }
  code { overflow-wrap: anywhere; }
  summary { cursor: pointer; color: #214e80; }
  [role=alert] { color: #9b2424; }
</style>
