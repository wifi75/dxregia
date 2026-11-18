<script lang="ts">
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'
  import { tr } from './adminI18n'
  import { lang } from './i18n.svelte'

  type Row = { Call: string; Email: string; Lang: string; Status: string; RequestedAt: string; ConfirmedAt?: string }
  let enabled = $state(false)
  let publicURL = $state('')
  let smtp = $state(true)
  let rows = $state<Row[]>([])
  let registrationMode = $state<number>()
  let status = $state('')
  let busy = $state(false)

  const date = $derived(new Intl.DateTimeFormat(lang() === 'it' ? 'it-IT' : lang(), { dateStyle: 'short', timeStyle: 'short' }))

  async function load() {
    try {
      const [settings, list, options] = await Promise.all([fetch('/api/register/settings'), fetch('/api/registrations'), fetch('/api/telnet-options')])
      if (settings.ok) {
        const data = await settings.json()
        enabled = data.Enabled
        publicURL = data.PublicURL || location.origin
        smtp = data.SMTP
      }
      if (list.ok) rows = await list.json()
      if (options.ok) registrationMode = (await options.json()).Options?.Registration
    } catch {
      status = 'Impostazioni non disponibili.'
    }
  }
  onMount(load)

  async function save(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    status = ''
    try {
      const response = await fetch('/api/register/settings', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ Enabled: enabled, PublicURL: publicURL }),
      })
      status = response.ok ? 'Impostazioni salvate.' : (await response.text()).trim()
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }

  // Per questa funzione serve la modalità «solo utenti registrati dal sysop inviano spot».
  async function useRegisteredMode() {
    busy = true
    try {
      const current = await (await fetch('/api/telnet-options')).json()
      const response = await fetch('/api/telnet-options', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...current.Options, Registration: 1 }),
      })
      if (response.ok) { registrationMode = 1; status = 'Opzione impostata: solo gli utenti registrati inviano spot.' }
      else status = (await response.text()).trim()
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }

  async function revoke(row: Row) {
    if (!confirm(tr('Revocare la registrazione di {}? L’utente potrà ancora collegarsi ma non inviare spot, e la sua password non varrà più.', row.Call))) return
    const response = await fetch('/api/registrations/' + encodeURIComponent(row.Call), { method: 'DELETE' })
    status = response.ok ? 'Registrazione revocata.' : (await response.text()).trim()
    if (response.ok) await load()
  }
</script>

<section class="glass-panel" aria-labelledby="users-title">
  <h1 id="users-title">Utenti registrati<Hint text="Chi si registra dal portale sceglie nominativo, email e password, conferma l'email e può poi inviare spot collegandosi via Telnet con quella password." /></h1>

  {#if registrationMode !== undefined && registrationMode !== 1}
    <p class="warn" role="alert">
      Per far inviare spot solo agli utenti registrati serve l'opzione «Solo utenti registrati dal sysop inviano spot».
      <button type="button" class="secondary" onclick={useRegisteredMode} disabled={busy}>Imposta ora</button>
    </p>
  {/if}
  {#if !smtp}<p class="warn" role="alert">Serve il server di posta: configuralo in Gestione → Server di posta.</p>{/if}

  <form onsubmit={save}>
    <label class="check"><input type="checkbox" bind:checked={enabled} />Registrazione aperta dal portale<Hint text="Se attiva, nel portale compare il pulsante «Registrati». Se la disattivi, chi è già registrato resta tale." /></label>
    <label>Indirizzo pubblico del sito<Hint text="È la base dei link nelle email di conferma, ad esempio https://dxregia.iu3wjq.com. Va indicato qui per sicurezza: non lo ricaviamo dalla richiesta, così nessuno può far inviare link a siti diversi dal tuo." />
      <input bind:value={publicURL} placeholder="https://dxregia.example.org" required disabled={busy} />
    </label>
    <button disabled={busy}>Salva</button>
  </form>
  {#if status}<p class="note" role="status">{status}</p>{/if}

  <h2>Registrazioni<Hint text="Da confermare: l'utente non ha ancora cliccato il link. Registrato: può inviare spot. Scaduta: il link non è stato usato entro 24 ore." /></h2>
  {#if rows.length}
    <div class="table-wrap">
      <table>
        <thead><tr><th>Nominativo</th><th>Email</th><th>Stato</th><th>Richiesta</th><th></th></tr></thead>
        <tbody>
          {#each rows as row (row.Call)}
            <tr>
              <td><strong>{row.Call}</strong></td><td>{row.Email}</td><td>{row.Status}</td>
              <td>{date.format(new Date(row.RequestedAt))}</td>
              <td><button type="button" class="secondary" onclick={() => revoke(row)}>Revoca</button></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else}
    <p class="note">Nessuna registrazione.</p>
  {/if}
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.6rem 0 .6rem; font-size: 1.1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  label { display: grid; gap: .3rem; font-weight: 650; width: min(100%, 520px); }
  label.check { display: flex; align-items: center; gap: .5rem; }
  input:not([type=checkbox]) { min-height: 42px; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; width: 100%; box-sizing: border-box; }
  .warn { margin: 0 0 1rem; padding: .7rem .9rem; border-left: 4px solid #b8650d; background: #fdf0e0; color: #6b3d08; border-radius: 4px; display: flex; flex-wrap: wrap; gap: .6rem; align-items: center; }
  .note { margin: .8rem 0 0; color: #465e70; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 7px 12px; text-align: left; border-bottom: 1px solid #c4d6e3; }
  th { font-size: .8rem; text-transform: uppercase; letter-spacing: .05em; color: #1d5d8c; }
</style>
