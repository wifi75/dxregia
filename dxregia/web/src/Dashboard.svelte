<script lang="ts">
  import { onMount } from 'svelte'

  type Status = { address: string; state: 'ready' | 'unexpected' | 'unreachable'; checkedAt: string }
  type Channel = { Call: string; Kind: string; State: string; Since: number; Outbound?: boolean }
  type Snapshot = { At: number; Channels: Channel[] | null }
  type Tone = 'green' | 'amber' | 'red' | 'neutral'

  let { status, failed, refresh }: { status?: Status; failed: boolean; refresh: () => void } = $props()

  let snapshot = $state<Snapshot>()
  let directory = $state<Record<string, { Name: string; Locked?: boolean }>>({})
  let now = $state(Date.now())
  let loading = false
  let again = false
  let liveState = $state<'live' | 'reconnecting' | 'polling'>('polling')
  let updatedAt = $state<number>()
  let publicTelnet = $state('')
  let knownKeys: Set<string> | undefined

  const time = new Intl.DateTimeFormat('it-IT', { timeStyle: 'medium' })

  let fresh = $derived(Boolean(snapshot?.At) && now / 1000 - (snapshot?.At ?? 0) < 20)
  let channels = $derived(fresh ? snapshot?.Channels ?? [] : [])
  let nodes = $derived(channels.filter(channel => channel.Kind === 'node'))
  let sessions = $derived(channels.filter(channel => channel.Kind === 'telnet' || channel.Kind === 'web'))
  let partnerCalls = $derived(Object.keys(directory).sort())
  let lockedCalls = $derived(partnerCalls.filter(call => directory[call].Locked))
  let connected = $derived(partnerCalls.filter(call => nodes.some(node => node.Call === call && node.State === 'normal')))
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
  async function load() {
    if (loading) { again = true; return }
    loading = true
    try {
      const [presence, partners] = await Promise.all([
        fetch('/api/presence', { cache: 'no-store' }), fetch('/api/partners', { cache: 'no-store' }),
      ])
      snapshot = presence.ok ? await presence.json() : undefined
      if (partners.ok) directory = await partners.json()
    } catch {
      snapshot = undefined
    } finally {
      now = Date.now()
      updatedAt = now
      loading = false
      if (again) { again = false; void load() }
    }
  }

  onMount(() => {
    void load()
    void fetch('/api/node-settings', { cache: 'no-store' }).then(async response => {
      if (response.status !== 200) return
      const { advertisedHost, port } = await response.json()
      if (advertisedHost && port) {
        publicTelnet = `${advertisedHost.includes(':') ? `[${advertisedHost}]` : advertisedHost}:${port}`
      }
    }).catch(() => {})
    // Flusso live: il server avvisa appena arriva uno spot o cambia la presenza. Senza flusso si interroga ogni 5 secondi.
    let source: EventSource | undefined
    if (typeof EventSource !== 'undefined') {
      source = new EventSource('/api/live')
      source.onopen = () => { liveState = 'live' }
      source.addEventListener('tick', () => { liveState = 'live'; void load() })
      source.onerror = () => { liveState = source?.readyState === EventSource.CLOSED ? 'polling' : 'reconnecting' }
    }
    const timer = setInterval(() => { if (liveState !== 'live') void load() }, 5000)
    const safety = setInterval(load, 20000)
    const probe = setInterval(refresh, 30000)
    const clock = setInterval(() => { now = Date.now() }, 1000)
    return () => { source?.close(); clearInterval(timer); clearInterval(safety); clearInterval(probe); clearInterval(clock) }
  })
</script>

<div class="heading">
  <h1>Stato del nodo</h1>
  <p class="live" data-state={liveState} role="status">
    <span class="pulse" aria-hidden="true"></span>
    <strong>{liveState === 'live' ? 'LIVE' : liveState === 'reconnecting' ? 'Riconnessione…' : 'Aggiornamento ogni 5 s'}</strong>
    {#if updatedAt}<span class="muted">· aggiornato {time.format(new Date(updatedAt))}</span>{/if}
  </p>
</div>

<section class="health" aria-label="Salute del nodo">
  {#each [
    { label: 'Listener Telnet', ...listener },
    { label: 'Cluster Mojo', ...cluster },
    { label: 'Nodi partner', ...partnerTile },
  ] as tile (tile.label)}
    <article class="tile glass-panel" data-tone={tile.tone}>
      <p class="label"><span class="led" data-tone={tile.tone} aria-hidden="true"></span>{tile.label}</p>
      <p class="value">{tile.title}</p>
      {#if tile.label === 'Listener Telnet'}
        <p class="detail"><span>Indirizzo pubblico</span> · {publicTelnet || 'Non configurato'}</p>
        <p class="detail"><span>Controllo locale</span> · {tile.detail}</p>
      {:else}
        <p class="detail">{tile.detail}</p>
      {/if}
    </article>
  {/each}
</section>

<div class="columns">
  <article class="glass-panel box partners">
    <header class="box-head"><h2>Nodi partner</h2><span class="count">{[...new Set([...partnerCalls, ...nodes.map(node => node.Call)])].length}</span></header>
    <p class="sub">Altri cluster collegati al tuo come partner.</p>
    {#if lockedCalls.length}
      <p class="blocked" role="status">Nodi bloccati in Mojo · {lockedCalls.join(', ')}. <a href="#partner">Apri Nodi partner</a></p>
    {/if}
    {#if partnerCalls.length || nodes.length}
      <div class="partner-table">
        <table>
          <thead><tr><th scope="col">Nodo</th><th scope="col">Stato</th><th scope="col">Direzione</th></tr></thead>
          <tbody>
            {#each [...new Set([...partnerCalls, ...nodes.map(node => node.Call)])] as call (call)}
              {@const node = nodes.find(item => item.Call === call)}
              {@const locked = directory[call]?.Locked}
              {@const tone = locked || !node ? 'red' : node.State === 'normal' ? 'green' : 'amber'}
              <tr>
                <td><strong>{call}</strong>{#if directory[call]?.Name && directory[call].Name !== call}<small>{directory[call].Name}</small>{/if}</td>
                <td><span class="partner-state"><span class="led" data-tone={tone} aria-hidden="true"></span>{locked ? 'Nodo bloccato in Mojo' : !node ? 'Disconnesso' : node.State === 'normal' ? 'Attivo' : 'Sessione aperta'}</span>
                  {#if node?.State === 'normal'}<small>{'Dalle ' + time.format(new Date(node.Since * 1000))}</small>{/if}</td>
                <td>{#if node}<span class="tag">{node.Outbound ? 'In uscita ↗' : 'In ingresso ↙'}</span>{:else}—{/if}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="muted">Nessun nodo partner registrato.</p>
    {/if}
  </article>

  <article class="glass-panel box users">
    <header class="box-head"><h2>Radioamatori collegati al tuo nodo</h2><span class="count">{sessions.length}</span></header>
    <p class="sub">Utenti connessi in questo momento via Telnet o dal web.</p>
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

</div>

<style>
  .blocked, .blocked a { color: #9d2027; font-weight: 700; }
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
  .box-head { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
  .box-head h2 { margin: 0; }
  .count { min-width: 1.8em; padding: 2px 10px; border-radius: 999px; background: #e3eef6; color: #143a60; font-weight: 800; text-align: center; font-variant-numeric: tabular-nums; }
  .partners { border-top: 3px solid #3a82c0; }
  .partner-table { overflow-x: auto; }
  .partner-table table { width: 100%; border-collapse: collapse; text-align: left; }
  .partner-table th, .partner-table td { padding: .65rem .55rem; border-bottom: 1px solid #cfdee7; }
  .partner-table th { color: #3f5c71; background: #e8f1f7; font-size: .8rem; }
  .partner-table small { display: block; color: #506678; margin-top: .18rem; }
  .partner-state { display: inline-flex; align-items: center; gap: .55rem; }
  .partner-state .led { width: 8px; height: 8px; flex-basis: 8px; }
  .partner-table .tag { white-space: nowrap; }
  .users { border-top: 4px solid #2fa36b; }
  .sub { margin: 4px 0 10px; color: #5b6e80; font-size: .88rem; }
  .columns { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 14px; margin-top: 14px; align-items: start; }
  @media (max-width: 1000px) { .columns { grid-template-columns: minmax(0, 1fr); } }
  .box { padding: 16px; }
  h2 { margin: 0 0 8px; font-size: 1.05rem; }
  h2:not(:first-child) { margin-top: 20px; }
  .muted { color: #485b6d; margin: 6px 0; }
  .rows { margin: 0; padding: 0; list-style: none; }
  .rows li { display: flex; align-items: center; flex-wrap: wrap; gap: 4px 10px; padding: 8px 2px; border-top: 1px solid #d7dfe7; }
  .tag { padding: 1px 8px; border-radius: 10px; background: #e7f1fa; color: #214e80; font-size: .82rem; font-weight: 650; }
  .rows .state { margin-left: auto; color: #485b6d; font-size: .9rem; }
  .live { display: flex; align-items: center; gap: 8px; margin: 0; padding: 5px 12px; border-radius: 16px; background: #ffffffa6; border: 1px solid #cdd9e4; font-size: .9rem; }
  .live strong { letter-spacing: .06em; color: #167b50; }
  .live[data-state=reconnecting] strong { color: #b8650d; }
  .live[data-state=polling] strong { color: #485b6d; letter-spacing: 0; }
  .pulse { width: 10px; height: 10px; border-radius: 50%; background: #1fa86b; animation: beat 1.6s ease-out infinite; }
  .live[data-state=reconnecting] .pulse { background: #e08a1a; }
  .live[data-state=polling] .pulse { background: #708391; animation: none; }
  @keyframes beat { 0% { box-shadow: 0 0 0 0 #1fa86b88; } 70%, 100% { box-shadow: 0 0 0 9px #1fa86b00; } }
  @media (prefers-reduced-motion: reduce) { .pulse { animation: none; } }
</style>
