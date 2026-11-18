<script lang="ts">
  let kind = $state('message')
  let call = $state('')
  let contact = $state('')
  let email = $state('')
  let name = $state('')
  let host = $state('')
  let port = $state<number>()
  let message = $state('')
  let pending = $state(false)
  let sent = $state(false)
  let problem = $state('')

  async function submit(event: SubmitEvent) {
    event.preventDefault()
    if (pending) return
    pending = true
    problem = ''
    try {
      const result = await fetch('/api/contact', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Kind: kind, Call: call, Contact: contact, Email: email, Name: name, Host: host, Port: port, Message: message }),
      })
      if (!result.ok) {
        problem = result.status === 404 ? 'Il modulo non è disponibile su questo nodo.' : (await result.text()).trim()
        return
      }
      sent = true
    } catch {
      problem = 'Il server non risponde. Riprova tra poco.'
    } finally {
      pending = false
    }
  }
</script>

<main>
  <h1>Contatta il sysop</h1>
  {#if sent}
    <p role="status">Richiesta registrata. Il sysop può consultarla e rispondere all’indirizzo indicato.</p>
    <a href="/spots">Torna agli spot</a>
  {:else}
    <form onsubmit={submit}>
      <label>Motivo del contatto
        <select bind:value={kind}>
          <option value="message">Scrivi al sysop</option>
          <option value="partner">Richiedi un collegamento tra nodi</option>
        </select>
      </label>
      <div class="fields">
        <label>Il tuo indicativo <input bind:value={call} required maxlength="12" autocapitalize="characters" spellcheck="false" /></label>
        <label>Nome del referente <input bind:value={contact} maxlength="120" autocomplete="name" /></label>
        <label>Email per la risposta <input bind:value={email} type="email" required maxlength="254" autocomplete="email" /></label>
      </div>
      {#if kind === 'partner'}
        <h2>Dati del nodo remoto</h2>
        <p>Usa l’indicativo del nodo, comprensivo di SSID se presente, nel campo dell’indicativo.</p>
        <div class="fields">
          <label>Nome del nodo <input bind:value={name} maxlength="120" required /></label>
          <label>Host o indirizzo IP <input bind:value={host} maxlength="253" required spellcheck="false" /></label>
          <label>Porta Telnet remota <input type="number" bind:value={port} min="1" max="65535" required /></label>
        </div>
      {/if}
      <label>{kind === 'partner' ? 'Note per il sysop' : 'Messaggio'}
        <textarea bind:value={message} maxlength="3000" rows="5" required={kind === 'message'}></textarea>
      </label>
      {#if problem}<p role="alert">{problem}</p>{/if}
      <button disabled={pending}>{pending ? 'Invio in corso…' : 'Invia richiesta'}</button>
    </form>
  {/if}
</main>

<style>
  main { max-width: 840px; margin: auto; padding: 22px clamp(14px, 4vw, 34px); }
  h1 { margin: 0 0 16px; font-size: 1.65rem; }
  h2 { margin: 6px 0 0; font-size: 1.12rem; }
  p { margin: 0; }
  form { background: white; border: 1px solid #d7dfe7; padding: 16px; display: grid; gap: 14px; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 195px), 1fr)); gap: 12px; }
  label { display: grid; gap: 5px; font-weight: 600; min-width: 0; }
  input, select, textarea { padding: 8px; min-height: 44px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-width: 0; }
  button { justify-self: start; min-height: 44px; border: 1px solid #214e80; border-radius: 4px; background: #214e80; color: white; padding: 7px 14px; font: inherit; font-weight: 650; cursor: pointer; }
  :focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
  [role=alert] { color: #9b2424; }
</style>
