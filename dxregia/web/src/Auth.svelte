<script lang="ts">
  import { onMount } from 'svelte'
  import { version as dxregiaVersion } from '../package.json'
  import { t } from './i18n.svelte'
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
  import Updates from './Updates.svelte'
  import EmailSettings from './EmailSettings.svelte'
  import PartnerMailHistory from './PartnerMailHistory.svelte'
  import NodeMessage from './NodeMessage.svelte'
  import NodeStartup from './NodeStartup.svelte'
  import NodeAccesses from './NodeAccesses.svelte'
  import NodeRegistrations from './NodeRegistrations.svelte'
  import NodeLists from './NodeLists.svelte'
  import NodeHops from './NodeHops.svelte'
  import Logo from './Logo.svelte'
  import Analytics from './Analytics.svelte'
  import ContactRequests from './ContactRequests.svelte'
  import type { ContactRequest } from './ContactRequests.svelte'
  let draft = $state<ContactRequest>()

  type Session = { signedIn: boolean; mustChange: boolean; mustVerify?: boolean; enrollmentKey?: string; recoveryCodes?: string[] }
  let session = $state<Session>()
  let { onSaved, nodeStatus }: { onSaved: () => void; nodeStatus: Snippet } = $props()
  const groups = [
    { label: 'Nodo', items: [['stato', 'Stato e sessioni'], ['accessi', 'Accessi al nodo'], ['configurazione', 'Configurazione'], ['accesso-web', 'Accesso web'], ['console', 'Console Telnet'], ['opzioni-telnet', 'Opzioni Telnet'], ['messaggio', 'Messaggio e pianificazione'], ['avvio', 'Script di avvio'], ['liste', 'Liste di blocco'], ['hop', 'Hop e filtri']] },
    { label: 'Diagnostica', items: [['log-dxspider', 'Log DXSpider']] },
    { label: 'Collegamenti', items: [['partner', 'Nodi partner'], ['rbn', 'Reverse Beacon Network'], ['email-partner', 'Storico email partner'], ['richieste', 'Richieste ricevute'], ['utenti', 'Utenti registrati']] },
    { label: 'Gestione', items: [['analytics', 'Analytics'], ['posta', 'Server di posta'], ['backup', 'Backup'], ['aggiornamenti', 'Aggiornamenti'], ['password', 'Password']] },
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
  let revealPassword = $state(false)
  let capsLock = $state(false)
  declare const __DXREGIA_REVISION__: string
  let engineVersion = $state('')
  let user = $state('')
  let error = $state('')
  let busy = $state(false)
  let recovery = $state(false)
  let codes = $state<string[]>([])
  let qr = $state('')
  let challenge = $derived(session?.mustVerify && !session.mustChange)

  function updateCapsLock(event: KeyboardEvent) {
    capsLock = !challenge && event.getModifierState('CapsLock')
  }

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
    void fetch('/api/engine').then(async response => {
      if (!response.ok) return
      const info = await response.json()
      if (info.Version) engineVersion = info.Version + (info.Build ? ` (build ${info.Build})` : '')
    }).catch(() => {})
    try {
      const response = await fetch('/api/auth')
      if (response.ok && response.headers.get('content-type')?.includes('application/json')) {
        session = await response.json()
      }
    } catch {
      // The listener diagnostic is also available without setup mode.
    }
  })

  // Sessione scaduta: si torna alla pagina di accesso, restando sulla stessa voce del menu.
  function expire() {
    if (!session?.signedIn || session.mustChange || session.mustVerify) return
    session = { signedIn: false, mustChange: false }
    password = ''
    codes = []
    error = 'Sessione scaduta. Accedi di nuovo.'
  }
  async function checkSession() {
    if (!session?.signedIn || session.mustChange || session.mustVerify) return
    try {
      const response = await fetch('/api/auth', { cache: 'no-store' })
      if (!response.ok || !response.headers.get('content-type')?.includes('application/json')) return
      const next: Session = await response.json()
      if (!next.signedIn) expire()
    } catch {
      // Server non raggiungibile: la sessione non è detto sia scaduta.
    }
  }
  onMount(() => {
    const timer = setInterval(checkSession, 30000)
    const onVisible = () => { if (document.visibilityState === 'visible') checkSession() }
    window.addEventListener('focus', checkSession)
    document.addEventListener('visibilitychange', onVisible)
    window.addEventListener('dxregia-unauthorized', expire)
    return () => {
      clearInterval(timer)
      window.removeEventListener('focus', checkSession)
      document.removeEventListener('visibilitychange', onVisible)
      window.removeEventListener('dxregia-unauthorized', expire)
    }
  })

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (!session || busy) return
    busy = true
    error = ''
    revealPassword = false
    capsLock = false
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
        <div class="sidebar-heading">
          <p class="sidebar-title">Amministrazione</p>
          <p class="sidebar-version">DXRegia v{dxregiaVersion}{#if __DXREGIA_REVISION__} · <code>{__DXREGIA_REVISION__}</code>{/if}</p>
        </div>
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
      <div class="workspace-content" hidden={page !== 'console'}>
        <TelnetConsole />
      </div>
      {#if page !== 'console'}
      {#key page}
      <div class="workspace-content glass-enter">
        {#if page === 'stato'}
          {@render nodeStatus()}
        {:else if page === 'configurazione'}
          <Setup {onSaved} />
        {:else if page === 'accesso-web'}
          <WebSettings />
        {:else if page === 'opzioni-telnet'}
          <TelnetOptions />
        {:else if page === 'log-dxspider'}
          <MojoLog />
        {:else if page === 'partner'}
          <Partners {draft} />
        {:else if page === 'rbn'}
          {#await import('./ReverseBeacon.svelte')}
            <p role="status">{t('loading')}</p>
          {:then section}
            <section.default admin />
          {:catch}
            <p role="alert">{t('rbnUnavailable')}</p>
          {/await}
        {:else if page === 'email-partner'}
          <PartnerMailHistory />
        {:else if page === 'richieste'}
          <ContactRequests onDraft={request => {
            draft = { ...request }
            location.hash = 'partner'
            page = 'partner'
          }} />
        {:else if page === 'analytics'}
          <Analytics />
        {:else if page === 'messaggio'}
          <NodeMessage />
        {:else if page === 'utenti'}
          <NodeRegistrations />
        {:else if page === 'accessi'}
          <NodeAccesses />
        {:else if page === 'avvio'}
          <NodeStartup />
        {:else if page === 'liste'}
          <NodeLists />
        {:else if page === 'hop'}
          <NodeHops />
        {:else if page === 'posta'}
          <EmailSettings />
        {:else if page === 'backup'}
          <BackupVerify />
        {:else if page === 'aggiornamenti'}
          <Updates />
        {:else if page === 'password'}
          <PasswordSettings />
        {/if}
      </div>
      {/key}
      {/if}
    </div>
    <footer class="login-footer admin-footer">
      {#if engineVersion}<span>DXSpider <strong>{engineVersion}</strong></span>{/if}
      <a href="https://github.com/wifi75/dxregia" target="_blank" rel="noopener noreferrer" title="Progetto open source">DXRegia su GitHub</a>
      <span>By IU3WJQ - Tiziano Cassone</span>
    </footer>
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
        <div class="password-field">
          <label for="sysop-password">
            {session.mustChange ? 'Nuova password' : challenge ? (recovery ? 'Codice di recupero' : 'Codice a sei cifre') : 'Password'}
          </label>
          <div class="password-input" class:with-toggle={!challenge}>
            <input id="sysop-password" type={(challenge ? !recovery : revealPassword) ? 'text' : 'password'} required bind:value={password}
              inputmode={challenge && !recovery ? 'numeric' : undefined}
              pattern={challenge ? (recovery ? '[0-9a-f]{48}' : '[0-9]{6}') : undefined}
              autocomplete={session.mustChange ? 'new-password' : session.mustVerify ? 'one-time-code' : 'current-password'}
              autocapitalize="off" spellcheck="false" onkeydown={updateCapsLock} onkeyup={updateCapsLock} onblur={() => { capsLock = false }}
              minlength={session.mustChange ? 12 : undefined} />
            {#if !challenge}
              <button class="password-toggle" type="button" aria-controls="sysop-password"
                aria-label={revealPassword ? 'Nascondi password' : 'Mostra password'}
                title={revealPassword ? 'Nascondi password' : 'Mostra password'}
                onclick={() => { revealPassword = !revealPassword }}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" aria-hidden="true">
                  <path d="M2 12s3.6-6 10-6 10 6 10 6-3.6 6-10 6-10-6-10-6Z" /><circle cx="12" cy="12" r="2.5" />
                  {#if revealPassword}<path d="m3 3 18 18" />{/if}
                </svg>
              </button>
            {/if}
          </div>
          <div role="status">
            {#if capsLock}
              <span class="caps-lock"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round" aria-hidden="true"><path d="m12 3 8 8h-4v6H8v-6H4l8-8Z M8 21h8" /></svg>Bloc Maiusc attivo</span>
            {/if}
          </div>
        </div>
        {#if session.mustChange}<p>Usa almeno 12 caratteri.</p>{/if}
        {#if error}<p role="alert">{error}</p>{/if}
        <button disabled={busy}>{session.mustChange ? 'Salva password' : session.mustVerify ? 'Verifica codice' : 'Accedi'}</button>
      </form>
      {#if challenge && !session.enrollmentKey}
        <button type="button" class="secondary" onclick={toggleRecovery} disabled={busy}>
          {recovery ? 'Usa l’app autenticatore' : 'Usa un codice di recupero'}
        </button>
      {/if}
      <footer class="login-footer">
        {#if engineVersion}<span>DXSpider <strong>{engineVersion}</strong></span>{/if}
        <a href="https://github.com/wifi75/dxregia" target="_blank" rel="noopener noreferrer" title="Progetto open source">DXRegia su GitHub</a>
        <span>By IU3WJQ - Tiziano Cassone</span>
      </footer>
    </section>
  {/if}
{/if}

<style>
  .auth-card { max-width: 440px; margin: 40px auto; padding: clamp(22px, 5vw, 36px); text-align: center; }
  .auth-card h2 { margin: 6px 0 22px; font-size: 1.25rem; text-align: center; }
  .auth-card form { display: grid; gap: 18px; text-align: left; max-width: 340px; margin: 0 auto; }
  .auth-card label { display: grid; max-width: none; gap: 8px; font-weight: 650; }
  .login-footer { display: flex; flex-wrap: wrap; justify-content: center; gap: .35rem .75rem; margin: 20px auto 0; padding-top: 12px; border-top: 1px solid #c4d6e3; color: #405b70; font-size: .78rem; }
  .login-footer a { color: #174e79; }
  .admin-footer { margin-top: 28px; }
  .auth-card input { width: 100%; min-height: 46px; padding: 10px 12px; font: inherit; }
  .auth-card form button { justify-self: center; width: auto; min-width: 120px; min-height: 40px; padding: 0 28px; margin-top: 4px; }
  .password-field { display: grid; }
  .password-field label { margin-bottom: 8px; }
  .password-input { position: relative; }
  .password-input.with-toggle input { padding-inline-end: 50px; }
  .password-input input::-ms-reveal { display: none; }
  .auth-card form .password-toggle { position: absolute; inset-inline-end: 3px; top: 3px; bottom: 3px; display: grid; place-items: center; width: 40px; min-width: 0; min-height: 0; margin: 0; padding: 0; border: 0; border-radius: 5px; background: transparent; color: #214e80; box-shadow: none; }
  .auth-card form .password-toggle:hover { background: #dfedf7; }
  .caps-lock { display: inline-flex; gap: 6px; align-items: center; margin-top: 8px; color: #79470c; font-size: .85rem; font-weight: 650; }
  .auth-card > button { margin-top: 14px; }
  .auth-card img { display: block; margin: 10px auto; }
  .auth-card details, .auth-card > p { text-align: left; }
  .workspace { display: grid; grid-template-columns: clamp(190px, 19vw, 264px) minmax(0, 1fr); align-items: start; gap: clamp(14px, 2vw, 28px); }
  .workspace-content { min-width: 0; }
  .sidebar { position: sticky; top: 18px; display: flex; flex-direction: column; padding: 16px 12px; min-height: calc(100vh - 112px); border-radius: 14px; background: #0f2438; color: #c9d9e8; box-shadow: 0 8px 24px #0f243833; }
  .sidebar-heading { margin: 0 0 14px; padding-inline: .7rem; }
  .sidebar-version { margin: .2rem 0 0; color: #b9cde0; font-size: .75rem; overflow-wrap: anywhere; }
  .sidebar-title { margin: 0; font-size: 1.02rem; color: #ffffff; font-weight: 750; letter-spacing: .02em; }
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
