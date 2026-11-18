<script lang="ts">
  import { onMount } from 'svelte'

  type Details = { Call: string; Name: string; Contact: string; Email: string; Host: string; Port?: number }
  type Partner = Details & { RegisteredAt: string; FirstConnectedAt: string | null; Notes: { Text: string; At: string }[] | null }
  const empty = (): Details => ({ Call: '', Name: '', Contact: '', Email: '', Host: '', Port: undefined })
  const date = new Intl.DateTimeFormat('it-IT', { dateStyle: 'short', timeStyle: 'short' })
  let partners = $state<Partner[]>([])
  let selected = $state<Partner>()
  let form = $state<Details>(empty())
  let editing = $state(false)
  let note = $state('')
  let error = $state('')
  let message = $state('')
  let busy = $state(true)

  onMount(async () => {
    try {
      const response = await fetch('/api/partners')
      if (!response.ok) throw new Error()
      partners = Object.values(await response.json() as Record<string, Partner>)
    } catch {
      error = 'Impossibile caricare i nodi partner.'
    } finally {
      busy = false
    }
  })

  function open(partner?: Partner) {
    selected = partner
    form = partner
      ? { Call: partner.Call, Name: partner.Name, Contact: partner.Contact, Email: partner.Email, Host: partner.Host, Port: partner.Port }
      : empty()
    note = ''
    error = ''
    message = ''
    editing = true
  }

  async function save(event: SubmitEvent) {
    event.preventDefault()
    if (busy) return
    busy = true
    error = ''
    message = ''
    try {
      const response = await fetch('/api/partners' + (selected ? '/' + encodeURIComponent(selected.Call) : ''), {
        method: selected ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, Note: note }),
      })
      if (!response.ok) {
        error = (await response.text()).trim()
        return
      }
      const saved: Partner = await response.json()
      partners = [...partners.filter(partner => partner.Call !== saved.Call), saved]
      open(saved)
      message = 'Anagrafica salvata.'
    } catch {
      error = 'Impossibile salvare il nodo partner.'
    } finally {
      busy = false
    }
  }
</script>

<section aria-labelledby="partners-title">
  <div class="toolbar">
    <h2 id="partners-title">Nodi partner</h2>
    <button type="button" onclick={() => open()} disabled={busy}>Aggiungi nodo</button>
  </div>
  {#if error}<p role="alert">{error}</p>{/if}
  {#if message}<p role="status">{message}</p>{/if}
  {#if partners.length}
    <div class="table-wrap">
      <table>
        <thead><tr><th>Indicativo</th><th>Nome nodo</th><th>Referente</th><th>Registrato il</th></tr></thead>
        <tbody>
          {#each partners.toSorted((a, b) => a.Call.localeCompare(b.Call)) as partner (partner.Call)}
            <tr>
              <td><button class="call" type="button" onclick={() => open(partner)} disabled={busy}>{partner.Call}</button></td>
              <td>{partner.Name || '—'}</td>
              <td>{partner.Contact || '—'}</td>
              <td>{date.format(new Date(partner.RegisteredAt))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if !busy && !error}
    <p>Nessun nodo partner registrato.</p>
  {/if}

  {#if editing}
    <h3>{selected ? selected.Call : 'Nuovo nodo partner'}</h3>
    {#if selected}
      <p>Prima connessione: {selected.FirstConnectedAt ? date.format(new Date(selected.FirstConnectedAt)) : 'Da rilevare'}</p>
    {/if}
    <form onsubmit={save}>
      <div class="fields">
        <label>Indicativo <input bind:value={form.Call} maxlength="12" readonly={Boolean(selected)} required /></label>
        <label>Nome del nodo <input bind:value={form.Name} maxlength="120" /></label>
        <label>Nome del referente <input bind:value={form.Contact} maxlength="120" /></label>
        <label>Email <input type="email" bind:value={form.Email} maxlength="254" /></label>
        <label>Host <input bind:value={form.Host} maxlength="253" required /></label>
        <label>Porta remota <input type="number" bind:value={form.Port} min="1" max="65535" required /></label>
      </div>
      <label class="note">Aggiungi una nota <textarea bind:value={note} maxlength="4000" rows="3" placeholder="Es. Inviata email al referente"></textarea></label>
      <button disabled={busy}>{busy ? 'Salvataggio…' : 'Salva anagrafica'}</button>
    </form>
    {#if selected?.Notes?.length}
      <h3>Storico note</h3>
      <ul>
        {#each selected.Notes.toReversed() as entry}
          <li><time datetime={entry.At}>{date.format(new Date(entry.At))}</time><p>{entry.Text}</p></li>
        {/each}
      </ul>
    {/if}
  {/if}
</section>

<style>
  section { margin-top: 20px; padding: 16px; background: white; border: 1px solid #d7dfe7; }
  .toolbar { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px; }
  h2 { margin: 0; font-size: 1.15rem; }
  h3 { margin: 18px 0 6px; font-size: 1rem; }
  .table-wrap { overflow: auto; margin-top: 12px; }
  table { width: 100%; text-align: left; border-collapse: collapse; white-space: nowrap; }
  th, td { padding: 5px 8px; border-bottom: 1px solid #d7dfe7; }
  th { background: #edf3f9; }
  .fields { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr)); gap: 12px 16px; }
  label { display: grid; min-width: 0; gap: 5px; font-weight: 600; }
  .note { margin: 12px 0; }
  input, textarea { min-width: 0; padding: 8px; border: 1px solid #aab8c5; border-radius: 4px; font: inherit; min-height: 44px; }
  button { min-height: 44px; padding: 7px 13px; border: 1px solid #214e80; border-radius: 4px; background: #214e80; color: #fff; font: inherit; font-weight: 650; cursor: pointer; }
  .call { background: none; color: #214e80; border: 0; padding: 4px; min-height: 34px; text-decoration: underline; }
  :is(button,input,textarea):focus-visible { outline: 2px solid #214e80; outline-offset: 2px; }
  [role=alert] { color: #9b2424; }
  ul { padding: 0; list-style: none; }
  li { border-top: 1px solid #d7dfe7; padding: 8px 0; }
  li p { white-space: pre-wrap; overflow-wrap: anywhere; margin: 3px 0; }
  time { color: #485b6d; font-size: .85rem; }
</style>
