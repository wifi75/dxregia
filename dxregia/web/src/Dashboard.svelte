<script lang="ts">
  import { onMount } from 'svelte'

  type Status = { address: string; state: 'ready' | 'unexpected' | 'unreachable'; checkedAt: string }
  type Channel = { Call: string; Kind: string; State: string; Since: number }
  type Snapshot = { At: number; Channels: Channel[] | null }
  type Spot = { Frequency: number; Call: string; At: number; Spotter: string; Entity?: string }
  type Tone = 'green' | 'amber' | 'red' | 'neutral'

  let { status, failed, busy, refresh }: { status?: Status; failed: boolean; busy: boolean; refresh: () => void } = $props()

  let snapshot = $state<Snapshot>()
  let directory = $state<Record<string, { Name: string }>>({})
  let spots = $state<Spot[]>([])
  let spotsAvailable = $state(true)
  let views = $state<number>()
  let now = $state(Date.now())
  let loading = false

  const time = new Intl.DateTimeFormat('it-IT', { timeStyle: 'medium' })
  const utc = new Intl.DateTimeFormat('it-IT', { timeZone: 'UTC', timeStyle: 'medium' })

  const bands: [string, number, number][] = [
    ['160 m', 1800, 2000], ['80 m', 3500, 4000], ['60 m', 5300, 5410], ['40 m', 7000, 7300], ['30 m', 10100, 10150],
    ['20 m', 14000, 14350], ['17 m', 18068, 18168], ['15 m', 21000, 21450], ['12 m', 24890, 24990], ['10 m', 28000, 29700],
    ['6 m', 50000, 54000], ['2 m', 144000, 148000], ['70 cm', 430000, 440000],
  ]
  const bandOf = (khz: number) => bands.find(([, low, high]) => khz >= low && khz <= high)?.[0] ?? 'Altre'

  let fresh = $derived(Boolean(snapshot?.At) && now / 1000 - (snapshot?.At ?? 0) < 20)
  let channels = $derived(fresh ? snapshot?.Channels ?? [] : [])
  let nodes = $derived(channels.filter(channel => channel.Kind === 'node'))
  let sessions = $derived(channels.filter(channel => channel.Kind === 'telnet' || channel.Kind === 'web'))
  let partnerCalls = $derived(Object.keys(directory).sort())
  let connected = $derived(partnerCalls.filter(call => nodes.some(node => node.Call === call && node.State === 'normal')))
  let lastHour = $derived(spots.filter(spot => now / 1000 - spot.At < 3600))
  let lastSpot = $derived(spots[0])
  let byBand = $derived.by(() => {
    const counts = new Map<string, number>()
    for (const spot of lastHour) counts.set(bandOf(spot.Frequency), (counts.get(bandOf(spot.Frequency)) ?? 0) + 1)
    const rows = [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 8)
    return { rows, max: Math.max(1, ...rows.map(([, n]) => n)) }
  })

  let listener = $derived<{ tone: Tone; title: string; detail: string }>(
    failed ? { tone: 'red', title: 'Non verificabile', detail: 'Il pannello non risponde.' }
      : !status ? { tone: 'neutral', title: 'Verifica…', detail: 'Interrogazione in corso.' }
      : status.state === 'ready' ? { tone: 'green', title: 'Pronto', detail: status.address }
      : status.state === 'unexpected' ? { tone: 'amber', title: 'Prompt assente', detail: status.address }
      : { tone: 'red', title: 'Irraggiungibile', detail: status.address })
  let cluster = $derived<{ tone: Tone; title: string; detail: string }>(
    fresh ? { tone: 'green', title: 'Attivo', detail: 'Dati aggiornati ' + time.format(new Date((snapshot?.At ?? 0) * 1000)) }
      : { tone: 'red', title: 'Nessun dato', detail: 'Mojo non invia dati recenti.' })
  let partnerTile = $derived<{ tone: Tone; title: string; detail: string }>(
    !partnerCalls.length ? { tone: 'neutral', title: 'Nessun partner', detail: 'Aggiungi un nodo dalla pagina Nodi partner.' }
      : connected.length === partnerCalls.length ? { tone: 'green', title: `${connected.length} su ${partnerCalls.length} collegati`, detail: 'Tutti i nodi sono attivi.' }
      : connected.length ? { tone: 'amber', title: `${connected.length} su ${partnerCalls.length} collegati`, detail: 'Alcuni nodi sono disconnessi.' }
      : { tone: 'red', title: `0 su ${partnerCalls.length} collegati`, detail: 'Nessun nodo partner è attivo.' })
  let spotTile = $derived<{ tone: Tone; title: string; detail: string }>(
    !spotsAvailable ? { tone: 'red', title: 'Archivio non leggibile', detail: 'Gli spot non sono disponibili.' }
      : !lastSpot ? { tone: 'neutral', title: 'Nessuno spot', detail: 'Non risultano spot recenti.' }
      : now / 1000 - lastSpot.At < 900 ? { tone: 'green', title: `${lastHour.length >= 100 ? '100+' : lastHour.length} nell’ultima ora`, detail: 'Ultimo spot ' + ago(lastSpot.At) }
      : now / 1000 - lastSpot.At < 3600 ? { tone: 'amber', title: `${lastHour.length} nell’ultima ora`, detail: 'Ultimo spot ' + ago(lastSpot.At) }
      : { tone: 'red', title: 'Fermi da oltre un’ora', detail: 'Ultimo spot ' + ago(lastSpot.At) })

  function ago(at: number) {
    const seconds = Math.max(0, Math.round(now / 1000 - at))
    if (seconds < 60) return seconds + ' s fa'
    if (seconds < 3600) return Math.round(seconds / 60) + ' min fa'
    return Math.round(seconds / 3600) + ' h fa'
  }

  async function load() {
    if (loading) return
    loading = true
    try {
      const [presence, partners, spotRows, traffic] = await Promise.all([
        fetch('/api/presence', { cache: 'no-store' }), fetch('/api/partners', { cache: 'no-store' }),
        fetch('/api/spots', { cache: 'no-store' }), fetch('/api/traffic', { cache: 'no-store' }),
      ])
      snapshot = presence.ok ? await presence.json() : undefined
      if (partners.ok) directory = await partners.json()
      if (spotRows.ok) {
        const data = await spotRows.json()
        spots = data.Spots ?? []
        spotsAvailable = data.Available !== false
      }
      if (traffic.ok) views = (await traffic.json()).Active
    } catch {
      snapshot = undefined
    } finally {
      now = Date.now()
      loading = false
    }
  }

  onMount(() => {
    void load()
    const timer = setInterval(load, 5000)
    const clock = setInterval(() => { now = Date.now() }, 1000)
    return () => { clearInterval(timer); clearInterval(clock) }
  })
</script>

<div class="heading">
  <h1>Stato del nodo</h1>
  <button onclick={refresh} disabled={busy}>{busy ? 'Verifica in corso…' : 'Verifica ora'}</button>
</div>

<section class="health" aria-label="Salute del nodo">
  {#each [
    { label: 'Listener Telnet', ...listener },
    { label: 'Cluster Mojo', ...cluster },
    { label: 'Nodi partner', ...partnerTile },
    { label: 'Spot', ...spotTile },
  ] as tile (tile.label)}
    <article class="tile glass-panel" data-tone={tile.tone}>
      <p class="label"><span class="led" data-tone={tile.tone} aria-hidden="true"></span>{tile.label}</p>
      <p class="value">{tile.title}</p>
      <p class="detail">{tile.detail}</p>
    </article>
  {/each}
</section>

<section class="counters" aria-label="Numeri">
  <div class="glass-panel counter"><strong>{sessions.length}</strong><span>Radioamatori collegati</span></div>
  <div class="glass-panel counter"><strong>{nodes.length}</strong><span>Nodi con sessione aperta</span></div>
  <div class="glass-panel counter"><strong>{lastHour.length >= 100 ? '100+' : lastHour.length}</strong><span>Spot nell’ultima ora</span></div>
  <div class="glass-panel counter"><strong>{views ?? '—'}</strong><span>Visitatori attivi sul portale</span></div>
</section>

<div class="columns">
  <article class="glass-panel box">
    <h2>Nodi partner</h2>
    {#if partnerCalls.length || nodes.length}
      <ul class="rows">
        {#each [...new Set([...partnerCalls, ...nodes.map(node => node.Call)])] as call (call)}
          {@const node = nodes.find(item => item.Call === call)}
          {@const tone = !node ? 'red' : node.State === 'normal' ? 'green' : 'amber'}
          <li>
            <span class="led" data-tone={tone} aria-hidden="true"></span>
            <strong>{call}</strong>
            <span class="muted">{directory[call]?.Name || ''}</span>
            <span class="state">{!node ? 'Disconnesso' : node.State === 'normal' ? 'Attivo dalle ' + time.format(new Date(node.Since * 1000)) : 'Sessione aperta'}</span>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">Nessun nodo partner registrato.</p>
    {/if}

    <h2>Radioamatori collegati</h2>
    {#if sessions.length}
      <ul class="rows">
        {#each sessions as session (session.Call + session.Kind)}
          <li>
            <span class="led" data-tone="green" aria-hidden="true"></span>
            <strong>{session.Call}</strong>
            <span class="muted">{session.Kind === 'telnet' ? 'Telnet' : 'Web Mojo'}</span>
            <span class="state">dalle {time.format(new Date(session.Since * 1000))}</span>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">Nessuna sessione Telnet o Web Mojo attiva.</p>
    {/if}
  </article>

  <article class="glass-panel box">
    <h2>Spot per banda · ultima ora</h2>
    {#if byBand.rows.length}
      <ul class="bars">
        {#each byBand.rows as [band, count] (band)}
          <li>
            <span class="band">{band}</span>
            <span class="bar"><i style:width={(count / byBand.max) * 100 + '%'}></i></span>
            <span class="num">{count}</span>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">Nessuno spot nell’ultima ora.</p>
    {/if}

    <h2>Ultimi spot</h2>
    {#if spots.length}
      <div class="overflow">
        <table>
          <thead><tr><th>UTC</th><th>DX</th><th>kHz</th><th>Da</th></tr></thead>
          <tbody>
            {#each spots.slice(0, 8) as spot (spot.At + spot.Call + spot.Frequency)}
              <tr><td>{utc.format(new Date(spot.At * 1000))}</td><td><strong>{spot.Call}</strong></td><td>{spot.Frequency.toFixed(1)}</td><td>{spot.Spotter}</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="muted">Nessuno spot recente.</p>
    {/if}
  </article>
</div>

<style>
  .heading { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 16px; }
  h1 { margin: 0; font-size: 1.65rem; line-height: 1.2; }
  .health { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 230px), 1fr)); gap: 14px; margin-top: 8px; }
  .tile { padding: 14px 16px; border-left: 5px solid #708391; }
  .tile[data-tone=green] { border-left-color: #167b50; }
  .tile[data-tone=amber] { border-left-color: #ad6a10; }
  .tile[data-tone=red] { border-left-color: #b32731; }
  .tile p { margin: 0; }
  .label { display: flex; align-items: center; gap: 9px; color: #485b6d; font-size: .88rem; font-weight: 650; text-transform: uppercase; letter-spacing: .03em; }
  .value { margin-top: 6px !important; font-size: 1.3rem; font-weight: 750; color: #143a60; line-height: 1.2; }
  .detail { margin-top: 4px !important; color: #485b6d; font-size: .9rem; overflow-wrap: anywhere; }
  .led { width: 13px; height: 13px; flex: 0 0 13px; border-radius: 50%; background: #708391; box-shadow: 0 0 0 4px #70839126; }
  .led[data-tone=green] { background: #1fa86b; box-shadow: 0 0 0 4px #1fa86b33, 0 0 10px #1fa86b88; }
  .led[data-tone=red] { background: #d0343f; box-shadow: 0 0 0 4px #d0343f33, 0 0 10px #d0343f88; }
  .led[data-tone=amber] { background: #e08a1a; box-shadow: 0 0 0 4px #e08a1a33, 0 0 10px #e08a1a88; }
  .counters { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 170px), 1fr)); gap: 14px; margin-top: 14px; }
  .counter { display: grid; padding: 12px 16px; }
  .counter strong { font-size: 1.9rem; line-height: 1.1; color: #143a60; font-variant-numeric: tabular-nums; }
  .counter span { color: #485b6d; font-size: .9rem; }
  .columns { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr)); gap: 14px; margin-top: 14px; align-items: start; }
  .box { padding: 16px; }
  h2 { margin: 0 0 8px; font-size: 1.05rem; }
  h2:not(:first-child) { margin-top: 20px; }
  .muted { color: #485b6d; margin: 6px 0; }
  .rows, .bars { margin: 0; padding: 0; list-style: none; }
  .rows li { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 10px; padding: 8px 2px; border-top: 1px solid #d7dfe7; }
  .rows .state { margin-left: auto; color: #485b6d; font-size: .9rem; }
  .bars li { display: grid; grid-template-columns: 52px 1fr 34px; align-items: center; gap: 10px; padding: 4px 0; }
  .band { font-weight: 650; }
  .bar { height: 14px; border-radius: 7px; background: #dde8f1; overflow: hidden; }
  .bar i { display: block; height: 100%; border-radius: 7px; background: linear-gradient(90deg, #2f78b3, #1fa86b); }
  .num { text-align: right; font-variant-numeric: tabular-nums; }
  .overflow { overflow: auto; }
  table { width: 100%; text-align: left; border-collapse: collapse; white-space: nowrap; font-variant-numeric: tabular-nums; }
  th, td { padding: 5px 8px; border-bottom: 1px solid #d7dfe7; }
  th { background: #edf3f9; }
</style>
