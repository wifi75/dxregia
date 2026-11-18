<script lang="ts">
  import { onMount } from 'svelte'
  import Setup from './Setup.svelte'
  import Partners from './Partners.svelte'
  import Presence from './Presence.svelte'

  type Session = { signedIn: boolean; mustChange: boolean; mustVerify?: boolean; enrollmentKey?: string }
  let session = $state<Session>()
  let { onSaved }: { onSaved: () => void } = $props()
  let password = $state('')
  let error = $state('')
  let busy = $state(false)

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
      const endpoint = session.mustChange ? 'password' : session.mustVerify ? 'totp' : 'login'
      const response = await fetch('/api/auth/' + endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(password),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      session = await response.json()
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
  {#if session.signedIn && !session.mustChange && !session.mustVerify}
    <Setup {onSaved} />
    <Presence />
    <Partners />
    <button type="button" onclick={logout}>Esci</button>
    {#if error}<p role="alert">{error}</p>{/if}
  {:else}
    <section>
      <h2>{session.mustChange ? 'Cambia la password temporanea' : session.mustVerify ? 'Verifica in due passaggi' : 'Accesso sysop'}</h2>
      {#if session.enrollmentKey}
        <p>Aggiungi una voce DXRegia nella tua app di autenticazione, usando questa chiave per un codice a tempo.</p>
        <p><code>{session.enrollmentKey}</code></p>
      {/if}
      <form onsubmit={submit}>
        <label>
          {session.mustChange ? 'Nuova password' : session.mustVerify ? 'Codice a sei cifre' : 'Password'}
          <input type={session.mustVerify ? 'text' : 'password'} required bind:value={password}
            inputmode={session.mustVerify ? 'numeric' : undefined}
            pattern={session.mustVerify ? '[0-9]{6}' : undefined}
            autocomplete={session.mustChange ? 'new-password' : session.mustVerify ? 'one-time-code' : 'current-password'}
            minlength={session.mustChange ? 12 : undefined} />
        </label>
        {#if session.mustChange}<p>Usa almeno 12 caratteri. Dopo il cambio imposterai il secondo fattore.</p>{/if}
        {#if error}<p role="alert">{error}</p>{/if}
        <button disabled={busy}>{session.mustChange ? 'Salva password' : session.mustVerify ? 'Verifica codice' : 'Accedi'}</button>
      </form>
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
  [role=alert] { color: #9b2424; }
</style>
