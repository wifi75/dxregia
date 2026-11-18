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
  import Traffic from './Traffic.svelte'
  import ContactRequests from './ContactRequests.svelte'
  import type { ContactRequest } from './ContactRequests.svelte'
  let draft = $state<ContactRequest>()

  type Session = { signedIn: boolean; mustChange: boolean; mustVerify?: boolean; enrollmentKey?: string; recoveryCodes?: string[] }
  let session = $state<Session>()
  let { onSaved, nodeStatus }: { onSaved: () => void; nodeStatus: Snippet } = $props()
  const groups = [
    { label: 'Nodo', items: [['stato', 'Stato e sessioni'], ['configurazione', 'Configurazione'], ['accesso-web', 'Accesso web'], ['console', 'Console Telnet'], ['opzioni-telnet', 'Opzioni Telnet'], ['log-dxspider', 'Log DXSpider']] },
    { label: 'Collegamenti', items: [['partner', 'Nodi partner'], ['richieste', 'Richieste ricevute']] },
    { label: 'Gestione', items: [['statistiche', 'Statistiche del portale'], ['posta', 'Server di posta'], ['backup', 'Backup'], ['password', 'Password']] },
  ] as const
  type Page = typeof groups[number]['items'][number][0]
  let page = $state<Page>('stato')

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
      <aside class="sidebar glass-panel">
        <p class="sidebar-title">Amministrazione</p>
        <nav aria-label="Amministrazione DXRegia">
          {#each groups as group}
            <div class="nav-group">
              <h2>{group.label}</h2>
              <ul>
                {#each group.items as [id, label]}
                  <li><a href={'#' + id} aria-current={page === id ? 'page' : undefined}
                    onclick={(event) => selectPage(event, id)}>{label}</a></li>
                {/each}
              </ul>
            </div>
          {/each}
        </nav>
        <button class="secondary" type="button" onclick={logout}>Esci</button>
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
        {:else if page === 'statistiche'}
          <Traffic />
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
      <h2>{session.mustChange ? 'Cambia la password temporanea' : session.mustVerify ? 'Verifica in due passaggi' : 'Accesso sysop'}</h2>
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
          <p class="hint">Il tuo indicativo da sysop. Se non l’hai ancora impostato: sysop.</p>
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
  .hint { margin: -6px 0 0; font-size: .88rem; }
  .auth-card { max-width: 560px; margin: 24px auto; padding: clamp(18px, 4vw, 32px); }
  .workspace { display: grid; grid-template-columns: clamp(190px, 19vw, 264px) minmax(0, 1fr); align-items: start; gap: clamp(14px, 2vw, 28px); }
  .workspace-content { min-width: 0; }
  .sidebar { position: sticky; top: 18px; padding: 16px 12px; min-height: calc(100vh - 112px); }
  .sidebar-title { margin: 0 0 1.2rem; padding-inline: .7rem; font-size: 1.06rem; color: #143a54; font-weight: 760; }
  .sidebar ul { margin: 0; padding: 0; list-style: none; }
  .nav-group + .nav-group { margin-top: 14px; padding-top: 12px; border-top: 1px solid #e2eaf0; }
  .nav-group h2 { margin: 0 0 .25rem; padding-left: .7rem; color: #425d70; font-size: .75rem; letter-spacing: .075em; text-transform: uppercase; }
  .nav-group a { display: flex; align-items: center; min-height: 40px; padding: .45rem .65rem; text-decoration: none; color: #244662; border-inline-start: 3px solid transparent; border-radius: 5px; }
  .nav-group a:hover { background: #ffffffa8; }
  .nav-group a[aria-current=page] { border-inline-start-color: #176c98; background: linear-gradient(100deg, #cceaf4, #e4f6f1); color: #103e60; font-weight: 700; }
  @media (max-width: 650px) {
    .workspace { grid-template-columns: minmax(0, 1fr); }
    .sidebar { position: static; min-height: 0; }
    .sidebar nav { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(145px, 100%), 1fr)); gap: .5rem; }
    .nav-group + .nav-group { margin-top: 0; padding-top: 0; border: 0; }
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
