<script lang="ts">
  import { onMount } from 'svelte'

  type View = { Address: string; From: string; User: string; TLS: string; PasswordSet: boolean; Source: string }
  let form = $state({ Address: '', From: '', User: '', TLS: 'implicit' })
  let password = $state('')
  let changePassword = $state(false)
  let passwordSet = $state(false)
  let source = $state('')
  let testTo = $state('')
  let busy = $state(false)
  let loaded = $state(false)
  let error = $state('')
  let message = $state('')

  const presets: [string, string, string][] = [
    ['Gmail', 'smtp.gmail.com:465', 'implicit'],
    ['Aruba', 'smtps.aruba.it:465', 'implicit'],
    ['Office 365', 'smtp.office365.com:587', 'starttls'],
  ]

  function apply(view: View) {
    form = { Address: view.Address, From: view.From, User: view.User, TLS: view.TLS || 'implicit' }
    passwordSet = view.PasswordSet
    source = view.Source
    password = ''
    changePassword = false
    if (!testTo) testTo = view.From
  }

  onMount(async () => {
    try {
      const response = await fetch('/api/smtp-settings', { cache: 'no-store' })
      if (!response.ok) throw new Error()
      apply(await response.json())
    } catch {
      error = 'Impossibile leggere le impostazioni di posta.'
    } finally {
      loaded = true
    }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (busy) return
    busy = true
    error = message = ''
    try {
      const response = await fetch('/api/smtp-settings', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, ...(changePassword || !passwordSet ? { Password: password } : {}) }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      apply(await response.json())
      message = 'Impostazioni salvate. Vale subito, senza riavviare. Invia un’email di prova per controllarle.'
    } catch {
      error = 'Impossibile salvare le impostazioni.'
    } finally {
      busy = false
    }
  }

  async function sendTest() {
    if (busy) return
    busy = true
    error = message = ''
    try {
      const response = await fetch('/api/smtp-settings/test', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ To: testTo }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      message = 'Email di prova inviata a ' + testTo + '. Controlla la casella, anche tra la posta indesiderata.'
    } catch {
      error = 'Impossibile inviare la prova.'
    } finally {
      busy = false
    }
  }
</script>

<section class="glass-panel" aria-labelledby="mail-title">
  <h2 id="mail-title">Server di posta (SMTP)</h2>
  <p class="intro">Serve per avvisarti delle richieste dal portale e per mandare ai partner le istruzioni di collegamento.</p>
  <p class="observed-state" role="status">
    <span class="led" data-tone={!loaded ? 'neutral' : form.Address ? 'green' : 'amber'} aria-hidden="true"></span>
    {!loaded ? 'Lettura in corso…' : form.Address ? 'Configurato' + (source === 'ambiente' ? ' (da variabili d’ambiente: salvando qui le sostituisci)' : '') : 'Non configurato: nessuna email viene inviata.'}
  </p>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if message}<p role="status">{message}</p>{/if}

  <form onsubmit={save}>
    <div class="presets">
      <span>Compila in fretta:</span>
      {#each presets as [name, address, tls] (name)}
        <button class="secondary" type="button" onclick={() => { form.Address = address; form.TLS = tls }}>{name}</button>
      {/each}
    </div>
    <div class="fields">
      <label>Server e porta <input bind:value={form.Address} placeholder="smtp.esempio.it:465" maxlength="253" required autocomplete="off" /></label>
      <label>Cifratura
        <select bind:value={form.TLS}>
          <option value="implicit">TLS implicito (porta 465)</option>
          <option value="starttls">STARTTLS (porta 587)</option>
        </select>
      </label>
      <label>Indirizzo mittente <input type="email" bind:value={form.From} placeholder="dx@esempio.it" maxlength="254" required autocomplete="off" /></label>
      <label>Utente <input bind:value={form.User} maxlength="254" autocomplete="off" placeholder="Di solito è l’email" /></label>
      {#if passwordSet && !changePassword}
        <div class="stored">
          <span>Password</span>
          <p class="muted">Salvata. <button class="link" type="button" onclick={() => { changePassword = true }}>Cambia</button></p>
        </div>
      {:else}
        <label>Password <input type="password" bind:value={password} maxlength="254" autocomplete="new-password" /></label>
      {/if}
    </div>
    <div class="actions">
      <button disabled={busy}>{busy ? 'Attendi…' : 'Salva impostazioni'}</button>
    </div>
  </form>

  <h3>Prova l’invio</h3>
  <div class="test">
    <label>Invia un’email di prova a <input type="email" bind:value={testTo} maxlength="254" /></label>
    <button class="secondary" type="button" disabled={busy || !testTo || !form.Address} onclick={sendTest}>Invia prova</button>
  </div>
  <p class="muted">La password è salvata solo sul server, in un file riservato, e non viene mai mostrata di nuovo.</p>
</section>

<style>
  section { margin-top: 20px; padding: 16px; }
  h2 { margin: 0; font-size: 1.15rem; }
  h3 { margin: 22px 0 8px; font-size: 1rem; }
  .intro, .muted { color: #485b6d; margin: 6px 0; }
  .observed-state { display: flex; align-items: center; gap: .55rem; }
  .led { width: 11px; height: 11px; flex: 0 0 11px; border-radius: 50%; background: #708391; box-shadow: 0 0 0 3px #70839123; }
  .led[data-tone=green] { background: #1fa86b; box-shadow: 0 0 0 3px #1fa86b33; }
  .led[data-tone=amber] { background: #e08a1a; box-shadow: 0 0 0 3px #e08a1a33; }
  form { display: grid; gap: 14px; margin-top: 10px; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 260px), 1fr)); gap: 12px 16px; }
  label { display: grid; min-width: 0; gap: 5px; font-weight: 600; }
  .stored span { font-weight: 600; }
  .stored p { margin: 8px 0 0; }
  .presets { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
  .actions, .test { display: flex; flex-wrap: wrap; align-items: end; gap: 10px; }
  .test label { flex: 1 1 280px; }
  input, select { min-width: 0; padding: 8px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-height: 44px; }
  .link { background: none; border: 0; box-shadow: none; padding: 2px; min-height: 0; color: #214e80; text-decoration: underline; font-weight: 600; }
  [role=alert] { color: #9b2424; }
  :is(button,input,select):focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
</style>
