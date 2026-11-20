<script lang="ts">
  import { onMount } from 'svelte'
  import { dxccFlags } from './dxccFlags'
  import { lang } from './i18n.svelte'

  type Status = { address: string; state: 'ready' | 'unexpected' | 'unreachable'; checkedAt: string }
  type Channel = { Call: string; Kind: string; State: string; Since: number; Outbound?: boolean; DXCC?: number }
  type Snapshot = { At: number; Channels: Channel[] | null }
  type ResourcePoint = { At: number; DiskUsed: number; DiskTotal: number; MemoryUsed: number; MemoryTotal: number }
  type NodeHealth = {
    MemoryFreeBytes: number; MemoryTotalBytes: number; MemoryAvailable: boolean; Resources?: ResourcePoint[]

    LogBytes: number; FreeBytes: number; TotalBytes: number; GrowthBytes: number
    GrowthIntervalSeconds: number; DiskAvailable: boolean; LogsAvailable: boolean
    DXRegiaPID: number; DXSpiderPID: number; LatestSpotAt: number; RetentionDays: number
  }
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
  let nodeHealth = $state<NodeHealth>()
  let presenceDelay = $state<number>()
  const number = new Intl.NumberFormat('it-IT', { maximumFractionDigits: 1 })
  function size(bytes: number) {
    const unit = Math.abs(bytes) >= 1024 ** 3 ? 3 : Math.abs(bytes) >= 1024 ** 2 ? 2 : Math.abs(bytes) >= 1024 ? 1 : 0
    return number.format(bytes / 1024 ** unit) + ' ' + ['B', 'KiB', 'MiB', 'GiB'][unit]
  }
  const diskHistory = $derived((nodeHealth?.Resources ?? []).filter(point => point.DiskTotal > 0))
  const memoryHistory = $derived((nodeHealth?.Resources ?? []).filter(point => point.MemoryTotal > 0))
  const percent = (used: number, total: number) => total > 0 ? Math.min(100, Math.max(0, used / total * 100)) : 0
  function series(history: ResourcePoint[], kind: 'Disk' | 'Memory') {
    const start = history[0]?.At ?? 0
    const elapsed = Math.max(1, (history.at(-1)?.At ?? start) - start)
    return history.map(point => ((point.At - start) / elapsed * 480).toFixed(1) + ',' +
      (64 - percent(point[kind + 'Used' as 'DiskUsed' | 'MemoryUsed'], point[kind + 'Total' as 'DiskTotal' | 'MemoryTotal']) * .60).toFixed(1)).join(' ')
  }
  const since = (at: number, current: number) => at > 0 && at <= current / 1000
    ? Math.floor(current / 1000 - at) + ' s' : '—'
  async function loadHealth() {
    try {
      const response = await fetch('/api/node-health', { cache: 'no-store', signal: AbortSignal.timeout(5000) })
      nodeHealth = response.ok ? await response.json() : undefined
    } catch { nodeHealth = undefined }
  }

  const time = new Intl.DateTimeFormat('it-IT', { timeStyle: 'medium' })

  let fresh = $derived(Boolean(snapshot?.At) && now / 1000 - (snapshot?.At ?? 0) < 20)
  let channels = $derived(fresh ? snapshot?.Channels ?? [] : [])
  let nodes = $derived(channels.filter(channel => channel.Kind === 'node'))
  let inboundNodes = $derived(nodes.filter(node => node.State === 'normal' && node.Outbound === false))
  let sessions = $derived(channels.filter(channel => channel.Kind === 'telnet' || channel.Kind === 'web')
    .sort((a, b) => a.Call.localeCompare(b.Call) || a.Kind.localeCompare(b.Kind)))
  let sessionsWidth = $state(0)
  let sessionPage = $state(1)
  const pageSize = $derived(sessionsWidth < 560 ? 12 : sessionsWidth < 960 ? 20 : 30)
  const pageCount = $derived(Math.max(1, Math.ceil(sessions.length / pageSize)))
  const currentPage = $derived(Math.min(sessionPage, pageCount))
  const visibleSessions = $derived(sessions.slice((currentPage - 1) * pageSize, currentPage * pageSize))
  const regionNames = $derived(new Intl.DisplayNames([lang()], { type: 'region' }))
  $effect(() => { if (fresh && sessionPage > pageCount) sessionPage = pageCount })
  let partnerCalls = $derived(Object.keys(directory).sort())
  const partnerRows = $derived([...new Set([...partnerCalls, ...nodes.map(node => node.Call)])].sort())
  let partnerPage = $state(1)
  const partnersPerPage = 8
  const partnerPageCount = $derived(Math.max(1, Math.ceil(partnerRows.length / partnersPerPage)))
  const currentPartnerPage = $derived(Math.min(partnerPage, partnerPageCount))
  const visiblePartners = $derived(partnerRows.slice((currentPartnerPage - 1) * partnersPerPage, currentPartnerPage * partnersPerPage))
  $effect(() => { if (fresh && partnerPage > partnerPageCount) partnerPage = partnerPageCount })
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
      : { tone: 'red', title: 'Nessun dato', detail: 'DXSpider non invia dati recenti.' })
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
      const age = Date.now() / 1000 - (snapshot?.At ?? 0)
      presenceDelay = snapshot?.At && age >= 0 ? Math.round(age) : undefined
      if (partners.ok) directory = await partners.json()
    } catch {
      snapshot = undefined
      presenceDelay = undefined
    } finally {
      now = Date.now()
      updatedAt = now
      loading = false
      if (again) { again = false; void load() }
    }
  }

  onMount(() => {
    void load()
    void loadHealth()
    const healthTimer = setInterval(loadHealth, 15000)
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
    return () => { source?.close(); clearInterval(timer); clearInterval(safety); clearInterval(probe); clearInterval(clock); clearInterval(healthTimer) }
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
    { label: 'Cluster DXSpider', ...cluster },
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
    <header class="box-head"><h2>Nodi partner</h2><span class="count">{partnerRows.length}</span></header>
    {#if lockedCalls.length}
      <p class="blocked" role="status">Nodi bloccati in DXSpider · {lockedCalls.join(', ')}. <a href="#partner">Apri Nodi partner</a></p>
    {/if}
    {#if partnerRows.length}
      <div class="partner-table">
        <table>
          <thead><tr><th scope="col">Nodo</th><th scope="col">Stato</th><th scope="col">Direzione</th></tr></thead>
          <tbody>
            {#each visiblePartners as call (call)}
              {@const node = nodes.find(item => item.Call === call)}
              {@const locked = directory[call]?.Locked}
              {@const tone = locked || !node ? 'red' : node.State === 'normal' ? 'green' : 'amber'}
              <tr>
                <td><strong title={directory[call]?.Name || undefined}>{call}</strong></td>
                <td><span class="partner-state"><span class="led" data-tone={tone} aria-hidden="true"></span>{locked ? 'Nodo bloccato in DXSpider' : !node ? 'Disconnesso' : node.State === 'normal' ? 'Attivo' : 'Sessione aperta'}
                  {#if !locked && node?.State === 'normal'}<small>· dalle {time.format(new Date(node.Since * 1000))}</small>{/if}</span></td>
                <td>{#if node}<span class="tag">{node.Outbound ? 'In uscita ↗' : 'In ingresso ↙'}</span>{:else}—{/if}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      {@render pagination('Pagine dei nodi partner', currentPartnerPage, partnerPageCount, page => partnerPage = page)}
    {:else}
      <p class="muted">Nessun nodo partner registrato.</p>
    {/if}
  </article>

  <article class="glass-panel box inbound">
    <header class="box-head"><h2>Nodi collegati in ingresso</h2><span class="count">{fresh ? inboundNodes.length : '—'}</span></header>
    {#if !fresh}
      <p class="muted">Stato delle sessioni non disponibile.</p>
    {:else if inboundNodes.length}
      <ul class="rows">
        {#each inboundNodes as node (node.Call)}
          <li>
            <span class="led" data-tone="green" aria-hidden="true"></span>
            <strong>{node.Call}</strong>
            {#if directory[node.Call]?.Name && directory[node.Call].Name !== node.Call}
              <span class="muted">{directory[node.Call].Name}</span>
            {/if}
            {#if node.Since > 0}<span class="state">dalle {time.format(new Date(node.Since * 1000))}</span>{/if}
          </li>
        {/each}
      </ul>
    {:else}
      <p class="muted">Nessun nodo collegato in ingresso.</p>
    {/if}
  </article>
</div>

<article class="glass-panel box users" aria-labelledby="active-users-title" bind:clientWidth={sessionsWidth}>
  <header class="box-head"><h2 id="active-users-title">Radioamatori collegati al tuo nodo</h2><span class="count">{fresh ? sessions.length : '—'}</span></header>
  <p class="sub">Utenti connessi in questo momento via Telnet o dal web.</p>
  {#if !fresh}
    <p class="muted">Stato delle sessioni non disponibile.</p>
  {:else if visibleSessions.length}
    <ul class="session-grid">
      {#each visibleSessions as session (session.Call + ':' + session.Kind)}
        {@const flag = dxccFlags[session.DXCC ?? 0]}
        <li>
          <div class="session-call">
            <span class="led" data-tone="green" aria-hidden="true"></span>
            {#if flag}
              {@const country = regionNames.of(flag.slice(0, 2).toUpperCase()) ?? flag.toUpperCase()}
              <img src={'/flags/' + flag + '.svg'} alt={country} title={country} width="22" height="16" loading="lazy" />
            {/if}
            <strong title={session.Call}>{session.Call}</strong>
          </div>
          <span class="tag" title={session.Kind === 'telnet' ? 'Telnet' : 'Web DXSpider'}>{session.Kind === 'telnet' ? 'Telnet' : 'Web'}</span>
          {#if session.Since > 0}
            <time datetime={new Date(session.Since * 1000).toISOString()}>{time.format(new Date(session.Since * 1000))}</time>
          {/if}
        </li>
      {/each}
    </ul>
    {@render pagination('Pagine delle sessioni', currentPage, pageCount, page => sessionPage = page)}
  {:else}
    <p class="muted">Nessuna sessione Telnet o Web DXSpider attiva.</p>
  {/if}
</article>

{#snippet pagination(label: string, current: number, total: number, go: (page: number) => void)}
  {#if total > 1}
    <nav class="pagination" aria-label={label}>
      <button type="button" onclick={() => go(current - 1)} disabled={current === 1}>Precedente</button>
      <span aria-live="polite">{'Pagina ' + current + ' di ' + total}</span>
      <button type="button" onclick={() => go(current + 1)} disabled={current === total}>Successiva</button>
    </nav>
  {/if}
{/snippet}

<section class="server-resources" aria-label="Risorse del server">
  {#snippet resource(title: string, free: number, total: number, history: ResourcePoint[], kind: 'Disk' | 'Memory')}
    <article class="glass-panel box resource">
      <h2>{title}</h2>
      {#if total > 0}
        {@const used = Math.max(0, total - free)}
        <div class="usage-head"><strong>{number.format(percent(used, total))}%</strong><span>{size(used)} / {size(total)}</span></div>
        <meter aria-label={title} min="0" max={total} value={used}></meter>
        <p>{size(free)} <span>disponibili</span></p>
        {#if history.length > 1}
          <svg viewBox="0 0 480 68" preserveAspectRatio="none" aria-hidden="true">
            <path d="M0 4H480 M0 34H480 M0 64H480" class="guides" />
            <polyline points={series(history, kind)} vector-effect="non-scaling-stroke" />
          </svg>
          <div class="usage-range"><time datetime={new Date(history[0].At * 1000).toISOString()}>{time.format(new Date(history[0].At * 1000))}</time>
            <time datetime={new Date(history.at(-1)!.At * 1000).toISOString()}>{time.format(new Date(history.at(-1)!.At * 1000))}</time></div>
        {:else}
          <small>Il grafico si aggiorna ogni 30 secondi.</small>
        {/if}
      {:else}
        <p>Misura non disponibile.</p>
      {/if}
    </article>
  {/snippet}
  {@render resource('Disco · volume dati', nodeHealth?.FreeBytes ?? 0, nodeHealth?.DiskAvailable ? nodeHealth.TotalBytes : 0, diskHistory, 'Disk')}
  {@render resource('Memoria RAM', nodeHealth?.MemoryFreeBytes ?? 0, nodeHealth?.MemoryAvailable ? nodeHealth.MemoryTotalBytes : 0, memoryHistory, 'Memory')}
</section>

<article class="glass-panel box diagnostics">
  <h2>Diagnostica del nodo</h2>
  <dl>
    <div><dt>Ritardo aggiornamento sessioni</dt><dd>{fresh && presenceDelay !== undefined ? presenceDelay + ' s' : '—'}</dd></div>
    <div><dt>Ultimo spot</dt><dd>{since(nodeHealth?.LatestSpotAt ?? 0, now)}</dd></div>
    <div><dt>Log e storico email</dt><dd>
      {nodeHealth?.LogsAvailable ? size(nodeHealth.LogBytes) : '—'}
      {#if nodeHealth?.GrowthIntervalSeconds}
        <small>{nodeHealth.GrowthBytes > 0 ? '+' : ''}{size(nodeHealth.GrowthBytes)} in {nodeHealth.GrowthIntervalSeconds} s</small>
      {/if}
      {#if nodeHealth?.RetentionDays}<small>Conservazione eventi: {nodeHealth.RetentionDays} giorni</small>{/if}
    </dd></div>
    <div><dt>Processo DXRegia</dt><dd>{nodeHealth?.DXRegiaPID ? 'PID ' + nodeHealth.DXRegiaPID : '—'}</dd></div>
    <div><dt>Processo DXSpider</dt><dd>{nodeHealth?.DXSpiderPID ? 'PID ' + nodeHealth.DXSpiderPID : fresh ? 'Dati attivi · PID non rilevato' : 'Non rilevato'}</dd></div>
  </dl>
</article>

<style>
  .server-resources { display: grid; grid-template-columns: repeat(auto-fit,minmax(min(100%,280px),1fr)); gap: 14px; margin-top: 14px; }
  .usage-head, .usage-range { display: flex; justify-content: space-between; align-items: baseline; gap: 8px; }
  .usage-head strong { font-size: 1.45rem; font-variant-numeric: tabular-nums; }
  .resource p { margin: .35rem 0 .6rem; }
  meter { display: block; width: 100%; height: .65rem; accent-color: #227392; }
  .resource svg { display: block; width: 100%; height: 68px; color: #227392; }
  .resource svg > * { stroke: currentColor; fill: none; stroke-width: 2; }
  .resource svg .guides { opacity: .24; stroke-width: 1; }
  .usage-range, .resource small { font-size: .8rem; }
  .diagnostics { margin-top: 14px; }
  .diagnostics dl { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 180px), 1fr)); gap: 14px; margin: 12px 0 0; }
  .diagnostics dt { color: #485b6d; font-size: .86rem; }
  .diagnostics dd { margin: 4px 0 0; font-weight: 700; }
  .diagnostics small { display: block; font-weight: 400; }

  .blocked, .blocked a { color: #9d2027; font-weight: 700; }
  .heading { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 16px; }
  h1 { margin: 0; font-size: 1.65rem; line-height: 1.2; }
  .health { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 216px), 1fr)); gap: 10px; margin-top: 6px; }
  .tile { padding: 8px 12px; border-left: 4px solid #708391; }
  .tile[data-tone=green] { border-left-color: #167b50; }
  .tile[data-tone=amber] { border-left-color: #ad6a10; }
  .tile[data-tone=red] { border-left-color: #b32731; }
  .tile p { margin: 0; }
  .label { display: flex; align-items: center; gap: 9px; color: #485b6d; font-size: .88rem; font-weight: 650; text-transform: uppercase; letter-spacing: .03em; }
  .tile .led { width: 8px; height: 8px; flex-basis: 8px; }
  .value { margin-top: 3px !important; font-size: 1.08rem; font-weight: 750; color: #143a60; line-height: 1.2; }
  .detail { margin-top: 2px !important; color: #485b6d; font-size: .81rem; overflow-wrap: anywhere; }
  .led { width: 13px; height: 13px; flex: 0 0 13px; border-radius: 50%; background: #708391; box-shadow: 0 0 0 4px #70839126; }
  .led[data-tone=green] { background: #1fa86b; box-shadow: 0 0 0 4px #1fa86b33, 0 0 10px #1fa86b88; }
  .led[data-tone=red] { background: #d0343f; box-shadow: 0 0 0 4px #d0343f33, 0 0 10px #d0343f88; }
  .led[data-tone=amber] { background: #e08a1a; box-shadow: 0 0 0 4px #e08a1a33, 0 0 10px #e08a1a88; }
  .box-head { display: flex; align-items: center; justify-content: space-between; gap: 10px; }
  .box-head h2 { margin: 0; }
  .count { min-width: 1.8em; padding: 2px 10px; border-radius: 999px; background: #e3eef6; color: #143a60; font-weight: 800; text-align: center; font-variant-numeric: tabular-nums; }
  .partners { border-top: 3px solid #3a82c0; }
  .partner-table { margin-top: 7px; overflow-x: auto; }
  .partner-table table { width: 100%; border-collapse: collapse; text-align: left; }
  .partner-table th, .partner-table td { padding: .32rem .5rem; border-bottom: 1px solid #cfdee7; }
  .partner-table th { color: #3f5c71; background: #e8f1f7; font-size: .8rem; }
  .partner-state { display: inline-flex; align-items: center; flex-wrap: wrap; gap: .2rem .45rem; }
  .partner-state small { color: #506678; white-space: nowrap; }
  .partner-state .led { width: 8px; height: 8px; flex-basis: 8px; }
  .partner-table .tag { white-space: nowrap; }
  .inbound { border-top: 3px solid #167b50; }
  .users { margin-top: 14px; border-top: 4px solid #2fa36b; }
  .session-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 204px), 1fr)); gap: 4px 6px; padding: 0; margin: 0; list-style: none; }
  .session-grid li { display: flex; align-items: center; gap: 6px; min-width: 0; padding: 3px 7px; border: 1px solid #cfdee7; border-radius: 6px; background: #ffffff60; white-space: nowrap; }
  .session-call { display: flex; align-items: center; gap: 5px; flex: 1; min-width: 0; }
  .session-call strong { overflow: hidden; text-overflow: ellipsis; font-size: .86rem; }
  .session-call .led, .inbound .led { width: 7px; height: 7px; flex-basis: 7px; }
  .session-call img { width: 18px; height: 13px; flex: none; object-fit: cover; border: 1px solid #15263826; border-radius: 2px; }
  .session-grid .tag { padding: 1px 5px; font-size: .73rem; }
  .session-grid time { color: #485b6d; font-size: .73rem; font-variant-numeric: tabular-nums; }
  .pagination { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; justify-content: end; margin-top: 7px; font-size: .8rem; }
  .pagination button { min-height: 28px; padding: 2px 8px; }

  .sub { margin: 4px 0 10px; color: #5b6e80; font-size: .88rem; }
  .columns { display: grid; grid-template-columns: minmax(0, 3fr) minmax(0, 2fr); gap: 10px; margin-top: 10px; align-items: start; }
  @media (max-width: 1000px) { .columns { grid-template-columns: minmax(0, 1fr); } }
  .box { padding: 16px; }
  .partners, .inbound, .users { padding: 10px 12px; }
  h2 { margin: 0 0 8px; font-size: 1.05rem; }
  h2:not(:first-child) { margin-top: 20px; }
  .muted { color: #485b6d; margin: 6px 0; }
  .rows { margin: 7px 0 0; padding: 0; list-style: none; max-height: 14rem; overflow-y: auto; }
  .rows li { display: flex; align-items: center; flex-wrap: wrap; gap: 3px 8px; padding: 4px 3px; border-top: 1px solid #d7dfe7; }
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
