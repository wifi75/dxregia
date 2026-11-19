<script lang="ts">
  import { onMount } from 'svelte'
  import MailSettings from './MailSettings.svelte'

  type Preferences = { Recipient: string; PortalMessages: boolean; PartnerRequests: boolean; PartnerDown: boolean; PartnerRecovered: boolean; PartnerDownMinutes: number }
  let active = $state<'email' | 'smtp'>('email')
  let preferences = $state<Preferences>()
  let sysopEmail = $state('')
  let smtpReady = $state(false)
  let busy = $state(false)
  let error = $state('')
  let message = $state('')

  async function load() {
    busy = true
    error = ''
    try {
      const response = await fetch('/api/email-preferences', { cache: 'no-store', signal: AbortSignal.timeout(8000) })
      if (!response.ok) throw new Error()
      const { SysopEmail, SMTPReady, ...settings } = await response.json()
      preferences = settings
      sysopEmail = SysopEmail
      smtpReady = SMTPReady
    } catch {
      error = 'Impossibile leggere le preferenze email.'
    } finally {
      busy = false
    }
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (!preferences || busy) return
    busy = true
    message = error = ''
    try {
      const response = await fetch('/api/email-preferences', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(preferences), signal: AbortSignal.timeout(8000),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      message = 'Preferenze email salvate.'
    } catch {
      error = 'Impossibile salvare le preferenze email.'
    } finally {
      busy = false
    }
  }

  function select(tab: 'email' | 'smtp') {
    if (tab === active) return
    active = tab
    message = ''
    if (tab === 'email') void load()
  }

  onMount(() => { void load() })
</script>

<nav class="mail-tabs" aria-label="Impostazioni di posta">
  <button type="button" aria-pressed={active === 'email'} onclick={() => select('email')}>Email</button>
  <button type="button" aria-pressed={active === 'smtp'} onclick={() => select('smtp')}>SMTP</button>
</nav>

{#if active === 'smtp'}
  <MailSettings />
{:else}
  <section class="glass-panel" aria-labelledby="email-notifications">
    <h1 id="email-notifications">Notifiche email al sysop</h1>
    {#if error}<p role="alert">{error}</p>{/if}
    {#if preferences}
      {#if !smtpReady}<p class="notice">Configura il server di invio nella linguetta SMTP.</p>{/if}
      <form onsubmit={save}>
        <label class="destination">Destinatario delle notifiche
          <input type="email" maxlength="254" bind:value={preferences.Recipient} placeholder={sysopEmail || 'Email del sysop'} disabled={busy} />
        </label>
        <p class="hint">Se lasci il campo vuoto verrà usato l’indirizzo del sysop nel profilo del nodo.</p>
        <fieldset disabled={busy}>
          <legend>Richieste dal portale</legend>
          <label class="choice"><input type="checkbox" bind:checked={preferences.PortalMessages} /> Nuovi messaggi al sysop</label>
          <label class="choice"><input type="checkbox" bind:checked={preferences.PartnerRequests} /> Nuove richieste per collegare un nodo partner</label>
        </fieldset>
        <fieldset disabled={busy}>
          <legend>Stato dei nodi partner</legend>
          <label class="choice"><input type="checkbox" bind:checked={preferences.PartnerDown} /> Avvisami se un partner resta scollegato</label>
          <label class="minutes">Attendi <input type="number" min="1" max="1440" step="1" required bind:value={preferences.PartnerDownMinutes} /> minuti</label>
          <label class="choice"><input type="checkbox" bind:checked={preferences.PartnerRecovered} disabled={!preferences.PartnerDown} /> Avvisami quando il collegamento torna attivo</label>
          <p class="hint">L’avviso riporta l’ultimo esito di DXSpider, quando disponibile. Viene inviato una volta per interruzione.</p>
        </fieldset>
        <button disabled={busy}>Salva preferenze email</button>
      </form>
    {:else if busy}
      <p role="status">Caricamento…</p>
    {:else}
      <button type="button" class="secondary" onclick={load}>Riprova</button>
    {/if}
    {#if message}<p role="status">{message}</p>{/if}
  </section>
{/if}

<style>
  .mail-tabs { display: flex; gap: .4rem; margin-bottom: .85rem; }
  .mail-tabs button { min-width: 92px; border: 1px solid #a8bfce; background: #f9fdff; color: #244762; box-shadow: none; }
  .mail-tabs button[aria-pressed='true'] { background: #1c5276; color: #fff; border-color: #1c5276; }
  section h1 { margin: 0 0 1rem; }
  form { display: grid; justify-items: start; gap: .85rem; }
  .destination { display: grid; gap: .4rem; width: min(100%, 28rem); }
  fieldset { min-width: 0; width: 100%; margin: 0; padding: .75rem 1rem; border: 1px solid #b7cad7; border-radius: 8px; }
  legend { font-weight: 700; padding: 0 .35rem; }
  .choice, .minutes { display: flex; align-items: center; gap: .55rem; margin: .35rem 0; }
  .minutes input { width: 6rem; }
  .hint { margin: 0; color: #4a6174; font-size: .86rem; }
  fieldset .hint { margin-top: .55rem; }
  [role='alert'], .notice { color: #a02527; }
</style>
