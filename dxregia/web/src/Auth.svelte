<script lang="ts">
  import { onMount } from 'svelte'
  import { toDataURL } from 'qrcode'
  import Setup from './Setup.svelte'
  import Partners from './Partners.svelte'
  import Presence from './Presence.svelte'
  import BackupVerify from './BackupVerify.svelte'
  import Traffic from './Traffic.svelte'
  import ContactRequests from './ContactRequests.svelte'
  import type { ContactRequest } from './ContactRequests.svelte'
  let draft = $state<ContactRequest>()

  type Session = { signedIn: boolean; mustChange: boolean; mustVerify?: boolean; enrollmentKey?: string; recoveryCodes?: string[] }
  let session = $state<Session>()
  let { onSaved }: { onSaved: () => void } = $props()
  let password = $state('')
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
        body: JSON.stringify(password),
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

{#if session}
  {#if codes.length}
    <section>
      <h2>Salva i codici di recupero</h2>
      <p>Conserva questi codici in un posto sicuro. Un codice e la password permettono di configurare una nuova app di autenticazione.</p>
      <ul class="recovery-codes">
        {#each codes as code}<li><code>{code}</code></li>{/each}
      </ul>
      <button type="button" onclick={() => { codes = []; if (session) session.recoveryCodes = undefined }}>Ho salvato i codici</button>
    </section>
  {:else if session.signedIn && !session.mustChange && !session.mustVerify}
    <Setup {onSaved} />
    <Presence />
    <Traffic />
    <BackupVerify />
    <ContactRequests onDraft={request => { draft = { ...request } }} />
    <Partners {draft} />
    <button type="button" onclick={logout}>Esci</button>
    {#if error}<p role="alert">{error}</p>{/if}
  {:else}
    <section>
      <h2>{session.mustChange ? 'Cambia la password temporanea' : session.mustVerify ? 'Verifica in due passaggi' : 'Accesso sysop'}</h2>
      {#if session.enrollmentKey}
        <p>Aggiungi DXRegia alla tua app di autenticazione.</p>
        {#if qr}<img src={qr} alt="Codice QR per configurare DXRegia nell’app di autenticazione" width="180" height="180" />{/if}
        <details><summary>Inserisci la chiave a mano</summary><p><code>{session.enrollmentKey}</code></p></details>
      {/if}
      <form onsubmit={submit}>
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
  section { margin-top: 20px; padding: 16px; background: #fff; border: 1px solid #d7dfe7; }
  h2 { margin: 0 0 12px; font-size: 1.15rem; }
  label { display: grid; max-width: 330px; gap: 5px; font-weight: 600; }
  input { border: 1px solid #aab8c5; border-radius: 4px; min-height: 44px; padding: 8px; font: inherit; }
  button { margin-top: 14px; min-height: 44px; padding: 7px 13px; border: 1px solid #214e80; border-radius: 4px; background: #214e80; color: #fff; font: inherit; font-weight: 650; cursor: pointer; }
  :is(button, input):focus-visible { outline: 2px solid #214e80; outline-offset: 3px; }
  .secondary { margin-left: 8px; background: #fff; color: #214e80; }
  .recovery-codes { padding-left: 20px; }
  code { overflow-wrap: anywhere; }
  summary { cursor: pointer; color: #214e80; }
  [role=alert] { color: #9b2424; }
</style>
