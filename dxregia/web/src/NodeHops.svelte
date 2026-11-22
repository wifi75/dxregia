<script lang="ts">
  import ConfirmDialog from './ConfirmDialog.svelte'
  let confirmer: ConfirmDialog
  import { tr } from './adminI18n'
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'

  type Rule = { Action: string; Band: string; Mode: string; Continent: string; Call: string }
  type Slot = { Number: number; Action: string; Text: string; Editor?: Rule }
  type Filter = { Sort: string; Name: string; Text: string; Size: number; Rules?: Slot[] }
  const emptyRule: Rule = { Action: 'reject', Band: '', Mode: '', Continent: '', Call: '' }
  let rule = $state<Rule>({ ...emptyRule })
  let bandOptions = $state<Record<string, string[]>>({})
  let modeOptions = $derived(rule.Band ? bandOptions[rule.Band] ?? [] : [...new Set(Object.values(bandOptions).flat())].sort())
  let target = $state('')
  let inbound = $state(false)
  let number = $state(1)
  let editing = $state('')
  let previousAction = ''
  let savingRule = $state(false)
  let ruleEditor = $state<HTMLFormElement>()
  const filterPath = (name: string, slot: number) => '/api/filters/spots/' + encodeURIComponent(name) + '/' + slot

  function editRule(filter: Filter, slot: Slot) {
    if (!slot.Editor) return
    editing = filter.Name
    inbound = filter.Name.startsWith('in_')
    target = filter.Name.replace(/^in_|\.pl$/g, '')
    number = slot.Number
    previousAction = slot.Action
    rule = { ...slot.Editor }
    ruleEditor?.scrollIntoView({ block: 'nearest' })
    ruleEditor?.querySelector<HTMLSelectElement>('#rule-band')?.focus()
  }
  function resetEditor() {
    editing = ''
    previousAction = ''
    rule = { ...emptyRule }
  }
  async function saveRule(event: SubmitEvent) {
    event.preventDefault()
    const input = target.trim()
    const normalized = ['node_default', 'user_default'].includes(input.toLowerCase()) ? input.toLowerCase() : input.toUpperCase()
    const name = editing || (inbound ? 'in_' : '') + normalized + '.pl'
    savingRule = true
    filterStatus = ''
    try {
      const response = await fetch(filterPath(name, number), {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...rule, PreviousAction: previousAction }),
      })
      if (!response.ok) {
        filterStatus = (await response.text()).trim()
        return
      }
      filterStatus = 'Regola salvata. Si applica entro pochi secondi.'
      resetEditor()
      await loadFilters()
    } catch { filterStatus = 'Connessione interrotta.' }
    finally { savingRule = false }
  }
  async function removeRule(filter: Filter, slot: Slot) {
    if (!await confirmer.ask(tr('Elimina filtro'), tr('Eliminare il filtro {}?', filter.Name + ' · ' + slot.Number + ' ' + slot.Action), tr('Elimina filtro'), true)) return
    try {
      const response = await fetch(filterPath(filter.Name, slot.Number) + '?action=' + slot.Action, { method: 'DELETE' })
      filterStatus = response.ok ? 'Filtro eliminato.' : (await response.text()).trim()
      if (response.ok) {
        if (editing === filter.Name && number === slot.Number && previousAction === slot.Action) resetEditor()
        await loadFilters()
      }
    } catch { filterStatus = 'Connessione interrotta.' }
  }
  let fallback = $state(30)
  let rows = $state<{ pc: string; hops: number }[]>([])
  let custom = $state(false)
  let filters = $state<Filter[]>([])
  let status = $state('')
  let filterStatus = $state('')
  let busy = $state(false)
  const kinds: Record<string, string> = { ann: 'Annunci', route: 'Instradamento', spots: 'Spot', wcy: 'WCY', wwv: 'WWV' }
  const packetTypes: Record<string, string> = { '16': 'Utente collegato a un nodo', '17': 'Utente scollegato da un nodo', '19': 'Nodo annunciato nella rete', '21': 'Nodo rimosso dalle rotte' }

  async function loadFilters() {
    const response = await fetch('/api/filters')
    if (response.ok) filters = await response.json()
  }
  onMount(async () => {
    try {
      const response = await fetch('/api/hops')
      if (response.ok) {
        const data = await response.json()
        fallback = data.Default; custom = data.Custom
        rows = Object.entries<number>(data.Table).map(([pc, hops]) => ({ pc, hops })).sort((a, b) => Number(a.pc) - Number(b.pc))
      }
      await loadFilters()
      const options = await fetch('/api/filters/spot-options')
      if (options.ok) bandOptions = await options.json()
    } catch { status = 'Impostazioni non disponibili.' }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    status = ''
    const Table: Record<string, number> = {}
    for (const row of rows) if (row.pc.trim()) Table[row.pc.trim()] = Number(row.hops)
    try {
      const response = await fetch('/api/hops', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ Default: Number(fallback), Table }),
      })
      status = response.ok ? 'Hop salvati. Si applicano entro pochi secondi, senza riavvio.' : (await response.text()).trim()
      if (response.ok) custom = true
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }
  async function remove(filter: Filter) {
    if (!await confirmer.ask(tr('Elimina filtro'), tr('Eliminare il filtro {}?', filter.Name), tr('Elimina filtro'), true)) return
    const response = await fetch('/api/filters/' + filter.Sort + '/' + filter.Name, { method: 'DELETE' })
    filterStatus = response.ok ? 'Filtro eliminato.' : (await response.text()).trim()
    if (response.ok) await loadFilters()
  }
</script>

<section class="glass-panel" aria-labelledby="hops-title">
  <h1 id="hops-title">Hop<Hint text="L'hop count limita quanti nodi può attraversare un'informazione prima di essere scartata. Valori bassi evitano che spot e annunci girino all'infinito; valori alti li fanno arrivare più lontano." /></h1>
  <div class="guide">
    <p>Un <strong>hop</strong> è un passaggio da un nodo DX Cluster al successivo. Il valore indica quanti passaggi può ancora fare un messaggio del protocollo. Quando si esaurisce, il messaggio smette di essere inoltrato.</p>
    <p><strong>Hop predefiniti</strong> è il valore generale. Ogni eccezione lo sostituisce per un tipo di pacchetto <strong>PC</strong>. Ad esempio, con valore generale 30 e PC16 a 15, gli aggiornamenti PC16 hanno un limite di 15 passaggi. Rimuovere una riga con ✕ fa tornare quel tipo al valore generale dopo il salvataggio.</p>
  </div>
  <form onsubmit={save}>
    <label>Hop predefiniti<Hint text="Valore usato per ogni tipo di messaggio che non ha una riga propria qui sotto. Da 1 a 99; DXSpider usa 30." />
      <input type="number" min="1" max="99" bind:value={fallback} disabled={busy} />
    </label>
    <p>Eccezioni per tipo di messaggio<Hint text="Ogni tipo di pacchetto del protocollo si chiama «PC» seguito da un numero (ad esempio PC16 e PC17 per gli utenti collegati, PC19 e PC21 per i nodi). Qui puoi dargli un hop count diverso dal predefinito." /></p>
    {#each rows as row, i}
      <div class="row">
        <label class="small">PC<input aria-label="Numero PC" title="Numero del pacchetto, da 1 a 999" bind:value={row.pc} disabled={busy} /></label>
        <label class="small">Hop<input aria-label="Hop" type="number" min="1" max="99" bind:value={row.hops} disabled={busy} /></label>
        <button type="button" class="ghost" aria-label="Elimina la riga" onclick={() => rows.splice(i, 1)}>✕</button>
        <span class="packet-type">{packetTypes[row.pc.trim()] ?? 'Altro tipo di pacchetto DXSpider'}</span>
      </div>
    {/each}
    <div class="actions">
      <button type="button" class="ghost" onclick={() => rows.push({ pc: '', hops: 15 })}>Aggiungi eccezione</button>
      <button disabled={busy}>Salva hop</button>
    </div>
    {#if !custom}<small>Valori originali di DXSpider: nessuna modifica salvata finora.</small>{/if}
  </form>
  {#if status}<p class="note" role="status">{status}</p>{/if}

  <h2>Filtri</h2>
  <p class="note"><span>Scegli il nodo o l’utente a cui applicare la regola.</span> <span>{tr('Usa {} per i nodi o {} per gli utenti che non hanno un filtro specifico.', 'node_default', 'user_default')}</span></p>
  <p class="note"><span>I criteri della stessa regola devono corrispondere insieme.</span> <span>Il modo usa i segmenti di frequenza del piano DXSpider.</span> <span>Le regole create dalla console si possono consultare o eliminare qui.</span></p>
  <form class="rule-editor" bind:this={ruleEditor} onsubmit={saveRule}>
    <div class="row">
      <label>Destinatario<input required maxlength="12" placeholder="EA7JXH-6" bind:value={target} disabled={Boolean(editing) || savingRule} /></label>
      <label class="small">Direzione<select bind:value={inbound} disabled={Boolean(editing) || savingRule}><option value={false}>Uscita</option><option value={true}>Ingresso (solo nodi)</option></select></label>
      <label class="small">Regola<input type="number" min="0" max="9" required bind:value={number} disabled={Boolean(editing) || savingRule} /></label>
      <label class="small">Azione<select bind:value={rule.Action} disabled={savingRule}><option value="reject">Rifiuta</option><option value="accept">Accetta</option></select></label>
    </div>
    <div class="row">
      <label class="small">Banda<select id="rule-band" bind:value={rule.Band} disabled={savingRule} onchange={() => { if (rule.Band && !bandOptions[rule.Band]?.includes(rule.Mode)) rule.Mode = '' }}>
        <option value="">Tutte</option>
        {#each Object.keys(bandOptions).sort((a, b) => a.localeCompare(b, undefined, { numeric: true })) as band}<option value={band}>{band}</option>{/each}
      </select></label>
      <label class="small">Modo<select bind:value={rule.Mode} disabled={savingRule}><option value="">Tutti</option>{#each modeOptions as mode}<option value={mode}>{mode.toUpperCase()}</option>{/each}</select></label>
      <label class="small">Continente DX<select bind:value={rule.Continent} disabled={savingRule}><option value="">Tutti</option>{#each ['AF', 'AN', 'AS', 'EU', 'NA', 'OC', 'SA'] as continent}<option value={continent}>{continent}</option>{/each}</select></label>
      <label>Nominativo DX esatto<input maxlength="24" placeholder="IU3WJQ" bind:value={rule.Call} disabled={savingRule} /></label>
    </div>
    <div class="actions">
      <button disabled={savingRule || !target.trim() || !(rule.Band || rule.Mode || rule.Continent || rule.Call.trim())}>{editing ? 'Salva modifiche' : 'Salva regola'}</button>
      {#if editing}<button class="ghost" type="button" onclick={resetEditor}>Annulla</button>{/if}
    </div>
  </form>
  {#if filters.length}
    <ul class="filters">
      {#each filters as filter}
        <li>
          <details>
            <summary><strong>{kinds[filter.Sort] ?? filter.Sort}</strong> · {filter.Name.replace(/\.pl$/, '')} <small>{filter.Size} byte</small></summary>
            {#if filter.Rules?.length}
              <ul class="rules">
                {#each filter.Rules as slot}
                  <li>
                    <span><strong>{slot.Number} · {slot.Action === 'accept' ? 'Accetta' : 'Rifiuta'}</strong> <code>{slot.Text}</code></span>
                    <div class="actions">
                      {#if slot.Editor}<button type="button" class="ghost" onclick={() => editRule(filter, slot)}>Modifica</button>{/if}
                      <button type="button" class="ghost" onclick={() => removeRule(filter, slot)}>Elimina regola</button>
                    </div>
                  </li>
                {/each}
              </ul>
            {:else}
              <pre>{filter.Text}</pre>
            {/if}
          </details>
          <button type="button" class="ghost" onclick={() => remove(filter)}>Elimina</button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="note">Nessun filtro attivo.</p>
  {/if}
  {#if filterStatus}<p class="note" role="status">{filterStatus}</p>{/if}
  <p class="note">Le liste «nodx», «noann» e «nowwv» del vecchio DXSpider non sono più lette da questo motore: per escludere un tipo di informazione usa i filtri.</p>
</section>

<ConfirmDialog bind:this={confirmer} />

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  h2 { margin: 1.8rem 0 .8rem; font-size: 1.1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  form p, .note { margin: 0; color: #465e70; }
  .note { margin-top: .6rem; }
  label { display: grid; gap: .35rem; font-weight: 650; width: min(100%, 220px); }
  .guide { max-width: 74ch; border-inline-start: 3px solid #3279ac; padding-inline-start: .85rem; margin-bottom: 1.15rem; }
  .guide p { margin: 0 0 .5rem; }
  .row { display: flex; flex-wrap: wrap; gap: .6rem; align-items: end; }
  .packet-type { align-self: center; color: #465e70; font-size: .88rem; }
  .small { width: 7rem; }
  input, select { min-height: 40px; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; width: 100%; }
  .actions { display: flex; gap: .6rem; flex-wrap: wrap; }
  .filters { list-style: none; margin: 0 0 .8rem; padding: 0; display: grid; gap: .5rem; }
  .filters li { display: flex; gap: .6rem; align-items: start; justify-content: space-between; }
  details { flex: 1; min-width: 0; }
  summary { cursor: pointer; }
  pre { margin: .4rem 0 0; padding: .6rem .8rem; overflow-x: auto; border-radius: 6px; background: #0001; max-height: 14rem; }
  .rule-editor { margin: 1rem 0; }
  .rules { margin: .6rem 0; padding: 0; list-style: none; }
  .rules li { padding: .5rem 0; border-bottom: 1px solid #a4bbc9; flex-wrap: wrap; }
  .rules code { overflow-wrap: anywhere; }
  small { color: #465e70; }
</style>
