<script lang="ts">
  import { onMount } from 'svelte'
  import Hint from './Hint.svelte'

  let words = $state('')
  let ips = $state('')
  let dx = $state('')
  let nodes = $state('')
  let spotters = $state('')
  let baseNodes = $state(0)
  let baseSpotters = $state(0)
  let otherWords = $state(0)
  let otherIps = $state(0)
  let status = $state('')
  let busy = $state(false)
  const lines = (text: string) => text.split(/[\n,;]+/).map(line => line.trim()).filter(Boolean)

  onMount(async () => {
    try {
      const response = await fetch('/api/badlists')
      if (!response.ok) return
      const data = await response.json()
      words = data.Words.join('\n'); ips = data.IPs.join('\n'); dx = data.DX.join('\n')
      nodes = (data.Nodes ?? []).join('\n'); spotters = (data.Spotters ?? []).join('\n')
      baseNodes = data.BaseNodes ?? 0; baseSpotters = data.BaseSpotters ?? 0
      otherWords = data.OtherWords; otherIps = data.OtherIPs
    } catch { status = 'Impostazioni non disponibili.' }
  })

  async function save(event: SubmitEvent) {
    event.preventDefault()
    busy = true
    status = ''
    try {
      const response = await fetch('/api/badlists', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Words: lines(words), IPs: lines(ips), DX: lines(dx), Nodes: lines(nodes), Spotters: lines(spotters) }),
      })
      status = response.ok ? 'Liste salvate. Il nodo le rilegge entro pochi secondi, senza riavvio.' : (await response.text()).trim()
    } catch { status = 'Connessione interrotta.' } finally { busy = false }
  }
</script>

<section class="glass-panel" aria-labelledby="lists-title">
  <h1 id="lists-title">Liste di blocco<Hint text="Elenchi che il nodo usa per difendersi da abusi. Si scrive un elemento per riga; le maiuscole non contano. Dopo il salvataggio DXSpider ricarica le liste da solo." /></h1>
  <form onsubmit={save}>
    <label>Parole vietate<Hint text="Parole che DXSpider rifiuta nei messaggi, negli annunci e nei comandi degli utenti. Solo lettere e cifre, senza spazi." />
      <textarea rows="8" bind:value={words} disabled={busy} spellcheck="false"></textarea>
    </label>
    {#if otherWords}<small>Nel file ci sono anche {otherWords} regole avanzate (espressioni regolari): restano come sono.</small>{/if}
    <label>Nominativi DX bloccati<Hint text="Gli spot con questi nominativi non vengono mostrati né inoltrati. Utile per nominativi di prova o inventati. Esempi: TEST, ID10T, K1/T." />
      <textarea rows="5" bind:value={dx} disabled={busy} spellcheck="false"></textarea>
    </label>
    <label>Nodi bloccati al login<Hint text="Un nodo con questo nominativo (anche senza SSID) viene rifiutato quando prova a collegarsi. Si aggiunge ai nominativi già presenti nell'elenco di base del nodo e a quelli messi con il comando set/badnode." />
      <textarea rows="4" bind:value={nodes} disabled={busy} spellcheck="false"></textarea>
    </label>
    {#if baseNodes}<small>L’elenco di base del nodo ha già {baseNodes} voci: qui si aggiungono le tue.</small>{/if}
    <label>Utenti bloccati al login<Hint text="Chi si collega con questo nominativo, anche con un SSID diverso, viene rifiutato. Si aggiunge all'elenco di base del nodo e al comando set/badspotter." />
      <textarea rows="4" bind:value={spotters} disabled={busy} spellcheck="false"></textarea>
    </label>
    {#if baseSpotters}<small>L’elenco di base del nodo ha già {baseSpotters} voci: qui si aggiungono le tue.</small>{/if}
    <label>Indirizzi IP bloccati<Hint text="Chi si collega da questi indirizzi viene rifiutato. Puoi indicare un indirizzo (192.0.2.1) o un'intera rete (192.0.2.0/24). Sono ammessi anche gli indirizzi IPv6." />
      <textarea rows="5" bind:value={ips} disabled={busy} spellcheck="false"></textarea>
    </label>
    {#if otherIps}<small>Nel file ci sono anche {otherIps} righe non riconosciute: restano come sono.</small>{/if}
    <button disabled={busy}>Salva liste</button>
  </form>
  {#if status}<p class="note" role="status">{status}</p>{/if}
</section>

<style>
  section { padding: 22px; }
  h1 { margin: 0 0 1rem; }
  form { display: grid; gap: .85rem; justify-items: start; }
  label { display: grid; gap: .35rem; width: min(100%, 560px); font-weight: 650; }
  textarea { width: 100%; box-sizing: border-box; border: 1px solid #a4bbc9; border-radius: 6px; padding: .45rem .65rem; font: inherit; font-family: ui-monospace, monospace; resize: vertical; }
  small, .note { color: #465e70; }
  .note { margin: .8rem 0 0; }
</style>
