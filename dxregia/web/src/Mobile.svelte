<script lang="ts">
  import { onMount } from 'svelte'
  import { version as dxregiaVersion } from '../package.json'
  import BarList from './BarList.svelte'
  import ColumnChart from './ColumnChart.svelte'
  import { dxccFlags } from './dxccFlags'
  import { currentTheme, toggleTheme } from './theme.svelte'

  // Applicazione per smartphone del sysop: menu in basso con più sezioni. Per aggiungerne una basta una voce in `tabs`
  // e un blocco {#if tab === …} più sotto.
  type Tab = 'stato' | 'statistiche' | 'altro'
  const tabs: { id: Tab; label: string; icon: string }[] = [
    { id: 'stato', label: 'Stato', icon: 'M3 12h4l3-8 4 16 3-8h4' },
    { id: 'statistiche', label: 'Statistiche', icon: 'M5 20V10M12 20V4M19 20v-7' },
    { id: 'altro', label: 'Altro', icon: 'M5 12h.01M12 12h.01M19 12h.01' },
  ]
  const fromHash = (): Tab => { const id = location.hash.slice(1); return tabs.some(tab => tab.id === id) ? id as Tab : 'stato' }
  let tab = $state<Tab>(fromHash())
  function open(id: Tab) {
    tab = id
    history.replaceState(null, '', '#' + id)
    window.scrollTo({ top: 0 })
  }

  type Channel = { Call: string; Kind: string; State: string; Since: number; DXCC?: number; SpotAt?: number }
  type Health = {
    MemoryFreeBytes: number; MemoryTotalBytes: number; MemoryAvailable: boolean
    FreeBytes: number; TotalBytes: number; DiskAvailable: boolean
    LatestSpotAt: number; ServiceStartedAt: number; EngineStartedAt: number; EngineRestarts: number
  }
  type LogEvent = { At: number; Kind: string; Message: string; Call?: string }
  type RecordEntry = { Count: number; At: number }
  type Item = { Name: string; Count: number; DXCC?: number }
  type Cluster = { Date: string; Spots: { Total: number; UniqueDX: number; UniqueSpotters: number; UniqueEntities: number; Bands: Item[] | null; Entities: Item[] | null; TopDX: Item[] | null; TopSpotters: Item[] | null } }
  type Portal = { Views: number; Visitors: number; Active: number; AllTime: number; Days: { Date: string; Views?: number; Visitors?: number }[]; Hours: number[]; Pages: Item[] | null; Devices: Item[] | null }

  let now = $state(Date.now())
  let updatedAt = $state<number>()
  let listener = $state<{ state: string; address: string }>()
  let snapshot = $state<{ At: number; Channels: Channel[] | null }>()
  let directory = $state<Record<string, { Name: string; Locked?: boolean }>>({})
  let health = $state<Health>()
  let events = $state<LogEvent[]>([])
  let records = $state<{ Telnet: RecordEntry; Web: RecordEntry }>()
  let publicTelnet = $state('')
  let cluster = $state<Cluster>()
  let portal = $state<Portal>()
  let days = $state(7)
  let offline = $state(false)

  const nf = new Intl.NumberFormat('it-IT', { maximumFractionDigits: 1 })
  const clock = new Intl.DateTimeFormat('it-IT', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  const hm = new Intl.DateTimeFormat('it-IT', { hour: '2-digit', minute: '2-digit' })
  const day = new Intl.DateTimeFormat('it-IT', { day: 'numeric', month: 'short', year: 'numeric' })
  const size = (bytes: number) => {
    const unit = bytes >= 1024 ** 3 ? 3 : bytes >= 1024 ** 2 ? 2 : bytes >= 1024 ? 1 : 0
    return nf.format(bytes / 1024 ** unit) + ' ' + ['B', 'KiB', 'MiB', 'GiB'][unit]
  }
  const span = (at: number) => {
    if (!(at > 0) || at > now / 1000) return '—'
    const minutes = Math.floor((now / 1000 - at) / 60)
    const d = Math.floor(minutes / 1440), h = Math.floor(minutes % 1440 / 60)
    return d ? d + ' g ' + h + ' h' : h ? h + ' h ' + minutes % 60 + ' min' : minutes + ' min'
  }
  const ago = (at: number) => {
    if (!(at > 0)) return '—'
    const seconds = Math.max(0, Math.floor(now / 1000 - at))
    return seconds < 90 ? seconds + ' s fa' : seconds < 5400 ? Math.round(seconds / 60) + ' min fa' : Math.round(seconds / 3600) + ' h fa'
  }

  const fresh = $derived(Boolean(snapshot?.At) && now / 1000 - (snapshot?.At ?? 0) < 20)
  const channels = $derived(fresh ? snapshot?.Channels ?? [] : [])
  const nodes = $derived(channels.filter(channel => channel.Kind === 'node'))
  const users = $derived(channels.filter(channel => channel.Kind === 'telnet' || channel.Kind === 'web')
    .sort((a, b) => (b.Since || 0) - (a.Since || 0) || a.Call.localeCompare(b.Call)))
  const telnetUsers = $derived(users.filter(user => user.Kind === 'telnet').length)
  const webUsers = $derived(users.length - telnetUsers)
  const upCalls = $derived(new Set(nodes.filter(node => node.State === 'normal' && !directory[node.Call]?.Locked).map(node => node.Call)))
  const partnerCalls = $derived(Object.keys(directory).sort())
  const partnerRows = $derived([...new Set([...partnerCalls, ...nodes.map(node => node.Call)])]
    .sort((a, b) => Number(upCalls.has(b)) - Number(upCalls.has(a)) || a.localeCompare(b)))
  const silent = $derived(new Set(nodes.filter(node => node.State === 'normal' && !directory[node.Call]?.Locked
    && now / 1000 - Math.max(node.SpotAt ?? 0, node.Since) > 30 * 60).map(node => node.Call)))
  let showAllUsers = $state(false)
  const shownUsers = $derived(showAllUsers ? users : users.slice(0, 24))

  type Tone = 'green' | 'amber' | 'red'
  const overall = $derived<{ tone: Tone; title: string; detail: string }>(
    offline ? { tone: 'red', title: 'Pannello non raggiungibile', detail: 'Controlla la connessione del telefono.' }
      : listener?.state !== 'ready' && listener ? { tone: 'red', title: 'Telnet non risponde', detail: listener.address }
      : !fresh ? { tone: 'red', title: 'Motore senza dati', detail: 'DXSpider non invia dati recenti.' }
      : partnerCalls.length && !upCalls.size ? { tone: 'amber', title: 'Nessun partner collegato', detail: 'Il nodo funziona ma è isolato.' }
      : silent.size ? { tone: 'amber', title: 'Nodo attivo, ' + silent.size + (silent.size === 1 ? ' partner' : ' partner') + ' senza spot', detail: [...silent].slice(0, 3).join(', ') }
      : partnerCalls.length && upCalls.size < partnerCalls.length ? { tone: 'amber', title: 'Nodo attivo', detail: (partnerCalls.length - upCalls.size) + ' partner disconnessi' }
      : { tone: 'green', title: 'Tutto in ordine', detail: publicTelnet || 'Nodo operativo' })

  const memoryUsed = $derived(health?.MemoryAvailable ? 1 - health.MemoryFreeBytes / health.MemoryTotalBytes : undefined)
  const diskUsed = $derived(health?.DiskAvailable ? 1 - health.FreeBytes / health.TotalBytes : undefined)
  const level = (value?: number) => value === undefined ? 'neutral' : value > .9 ? 'red' : value > .75 ? 'amber' : 'green'

  async function json<T>(url: string): Promise<T | undefined> {
    try {
      const response = await fetch(url, { cache: 'no-store', signal: AbortSignal.timeout(8000) })
      return response.ok ? await response.json() as T : undefined
    } catch { return undefined }
  }
  async function loadStatus() {
    const [l, p, d, h, e, r] = await Promise.all([
      json<{ state: string; address: string }>('/api/listener'), json<typeof snapshot>('/api/presence'),
      json<typeof directory>('/api/partners'), json<Health>('/api/node-health'),
      json<LogEvent[]>('/api/mojo-log?anonymous=false'), json<typeof records>('/api/records'),
    ])
    offline = !l && !p
    now = Date.now()
    if (l) listener = l
    snapshot = p
    if (d) directory = d
    health = h
    if (e) events = e.filter(event => event.At > 0 && (event.Kind === 'Avvio' || event.Kind === 'Collegamenti')).slice(0, 6)
    if (r) records = r
    if (!offline) updatedAt = now
  }
  async function loadStats() {
    const [c, a] = await Promise.all([json<Cluster>('/api/cluster-stats'), json<{ Portal: Portal }>('/api/analytics?days=' + days)])
    if (c) cluster = c
    if (a) portal = a.Portal
  }
  function choosePeriod(value: number) { days = value; portal = undefined; void loadStats() }
  const bars = (items: Item[] | null, color?: string) => (items ?? []).slice(0, 8).map(item => ({ name: item.Name, count: item.Count, color, flag: item.DXCC ? dxccFlags[item.DXCC] : undefined }))
  const shortDate = (date: string) => date.slice(8, 10) + '/' + date.slice(5, 7)

  async function logout() {
    try { await fetch('/api/auth/logout', { method: 'POST', headers: { 'Content-Type': 'application/json' } }) } catch { /* si ricarica comunque */ }
    location.reload()
  }

  const menu = [
    { label: 'Nodo', items: [['stato', 'Stato e sessioni'], ['accessi', 'Accessi al nodo'], ['console', 'Console Telnet'], ['messaggio', 'Messaggio e pianificazione']] },
    { label: 'Collegamenti', items: [['partner', 'Nodi partner'], ['utenti', 'Utenti registrati'], ['richieste', 'Richieste ricevute'], ['rbn', 'Reverse Beacon Network']] },
    { label: 'Gestione', items: [['analytics', 'Analytics'], ['backup', 'Backup'], ['aggiornamenti', 'Aggiornamenti'], ['sicurezza', 'Sicurezza'], ['password', 'Password']] },
  ] as const

  onMount(() => {
    void loadStatus(); void loadStats()
    void json<{ advertisedHost?: string; port?: number }>('/api/node-settings').then(s => {
      if (s?.advertisedHost && s.port) publicTelnet = (s.advertisedHost.includes(':') ? '[' + s.advertisedHost + ']' : s.advertisedHost) + ':' + s.port
    })
    const timers = [setInterval(() => { now = Date.now() }, 1000), setInterval(loadStatus, 10000), setInterval(loadStats, 30000)]
    // Tornando sull'app dopo un po' (telefono sbloccato) si aggiorna subito.
    const wake = () => { if (document.visibilityState === 'visible') { void loadStatus(); void loadStats() } }
    document.addEventListener('visibilitychange', wake)
    return () => { timers.forEach(clearInterval); document.removeEventListener('visibilitychange', wake) }
  })
</script>

<svelte:window onhashchange={() => { tab = fromHash() }} />

<div class="phone">
  <header class="bar">
    <img class="logo" src={currentTheme() === 'dark' ? '/brand/dxregia-logo-dark.svg' : '/brand/dxregia-logo.svg'} alt="DXRegia" width="96" height="29">
    <p class="stamp" role="status"><span class="dot" data-tone={offline ? 'red' : updatedAt ? 'green' : 'neutral'}></span>{updatedAt ? clock.format(new Date(updatedAt)) : '…'}</p>
  </header>

  <main>
    {#if tab === 'stato'}
      <section class="card">
        <h2>Stato del nodo</h2>
        <ul class="rows">
          <li><span class="dot" data-tone={offline ? 'red' : !listener ? 'neutral' : listener.state === 'ready' ? 'green' : listener.state === 'unexpected' ? 'amber' : 'red'}></span>
            <div><b>Listener Telnet</b><small>{offline ? 'Non verificabile' : !listener ? 'Verifica…' : listener.state === 'ready' ? 'Pronto' : listener.state === 'unexpected' ? 'Prompt assente' : 'Irraggiungibile'}</small>
              <small>Pubblico: {publicTelnet || 'non configurato'}</small></div></li>
          <li><span class="dot" data-tone={fresh ? 'green' : 'red'}></span>
            <div><b>Cluster DXSpider</b><small>{fresh ? 'Attivo' : 'Nessun dato recente'}</small>
              {#if snapshot?.At}<small>Aggiornato {clock.format(new Date(snapshot.At * 1000))}</small>{/if}</div></li>
          <li><span class="dot" data-tone={!fresh || !upCalls.size ? 'neutral' : silent.size ? 'amber' : 'green'}></span>
            <div><b>Partner silenziosi</b><small>{!fresh ? 'Non disponibile' : !upCalls.size ? 'Nessun nodo collegato' : silent.size ? silent.size + (silent.size === 1 ? ' nodo senza spot' : ' nodi senza spot') + ' da oltre 30 min' : 'Tutti i nodi inviano spot'}</small>
              {#if silent.size}<small>{[...silent].slice(0, 6).join(', ')}</small>{/if}</div></li>
        </ul>
      </section>

      <section class="card">
        <h2>Nodi partner <span class="badge">{upCalls.size}/{partnerCalls.length}</span></h2>
        {#if partnerRows.length}
          <ul class="partners">
            {#each partnerRows as call (call)}
              {@const isUp = upCalls.has(call)}
              <li><span class="dot" data-tone={directory[call]?.Locked ? 'neutral' : isUp ? (silent.has(call) ? 'amber' : 'green') : 'red'}></span>
                <span class="call">{call}</span>
                <small>{directory[call]?.Locked ? 'bloccato' : !isUp ? 'scollegato' : silent.has(call) ? 'senza spot' : 'collegato'}</small></li>
            {/each}
          </ul>
        {:else}<p class="empty">Nessun partner configurato.</p>{/if}
      </section>

      <section class="card">
        <h2>Radioamatori collegati al tuo nodo <span class="badge">{users.length}</span></h2>
        {#if users.length}
          <ul class="chips">
            {#each shownUsers as user (user.Kind + user.Call)}
              <li>{#if user.DXCC && dxccFlags[user.DXCC]}<img src={'/flags/' + dxccFlags[user.DXCC] + '.svg'} alt="" width="18" height="13">{/if}<span>{user.Call}</span><i>{user.Kind === 'web' ? 'W' : 'T'}</i></li>
            {/each}
          </ul>
          {#if users.length > shownUsers.length || showAllUsers && users.length > 24}
            <button class="more" type="button" onclick={() => { showAllUsers = !showAllUsers }}>{showAllUsers ? 'Mostra meno' : 'Mostra tutti (' + users.length + ')'}</button>
          {/if}
        {:else}<p class="empty">Nessuno collegato ora.</p>{/if}
      </section>

      <section class="card">
        <h2>Stato del server</h2>
        {#each [['Memoria', memoryUsed, health ? size(health.MemoryTotalBytes - health.MemoryFreeBytes) + ' di ' + size(health.MemoryTotalBytes) : ''], ['Disco', diskUsed, health ? size(health.TotalBytes - health.FreeBytes) + ' di ' + size(health.TotalBytes) : '']] as [label, value, text] (label)}
          <div class="meter">
            <p><span>{label}</span><b>{value === undefined ? '—' : Math.round((value as number) * 100) + '%'}</b></p>
            <div class="track"><i data-tone={level(value as number | undefined)} style:width={((value as number | undefined) ?? 0) * 100 + '%'}></i></div>
            <small>{text}</small>
          </div>
        {/each}
        <dl class="facts">
          <div><dt>Servizio attivo da</dt><dd>{health ? span(health.ServiceStartedAt) : '—'}</dd></div>
          <div><dt>Motore attivo da</dt><dd>{health ? span(health.EngineStartedAt) : '—'}</dd></div>
          <div><dt>Riavvii del motore</dt><dd>{health ? health.EngineRestarts : '—'}</dd></div>
        </dl>
      </section>


    {:else if tab === 'statistiche'}
      <h1 class="title">Statistiche</h1>
      <section class="card">
        <h2>Spot di oggi {#if cluster}<span class="badge">{cluster.Date}</span>{/if}</h2>
        {#if cluster}
          <div class="kpis tight">
            <div class="kpi"><span>Spot</span><strong>{nf.format(cluster.Spots.Total)}</strong></div>
            <div class="kpi"><span>DX diversi</span><strong>{nf.format(cluster.Spots.UniqueDX)}</strong></div>
            <div class="kpi"><span>Spotter</span><strong>{nf.format(cluster.Spots.UniqueSpotters)}</strong></div>
            <div class="kpi"><span>Paesi</span><strong>{nf.format(cluster.Spots.UniqueEntities)}</strong></div>
          </div>
        {:else}<p class="empty">Caricamento…</p>{/if}
      </section>
      {#if cluster}
        <section class="card"><h2>Bande</h2><BarList items={bars(cluster.Spots.Bands, '#2f78b3')} /></section>
        <section class="card"><h2>Paesi più spottati</h2><BarList items={bars(cluster.Spots.Entities, '#1fa86b')} /></section>
        <section class="card"><h2>DX più spottati</h2><BarList items={bars(cluster.Spots.TopDX, '#e08a1a')} /></section>
        <section class="card"><h2>Spotter più attivi</h2><BarList items={bars(cluster.Spots.TopSpotters, '#7a64c8')} /></section>
      {/if}

      <h2 class="group">Portale pubblico</h2>
      <div class="periods" role="group" aria-label="Periodo">
        {#each [[1, '24 ore'], [7, '7 giorni'], [30, '30 giorni']] as [value, label] (value)}
          <button type="button" aria-pressed={days === value} onclick={() => choosePeriod(value as number)}>{label}</button>
        {/each}
      </div>
      {#if portal}
        <div class="kpis tight">
          <div class="kpi"><span>Visite</span><strong>{nf.format(portal.Views)}</strong></div>
          <div class="kpi"><span>Visitatori</span><strong>{nf.format(portal.Visitors)}</strong></div>
          <div class="kpi"><span>Attivi ora</span><strong>{nf.format(portal.Active)}</strong></div>
          <div class="kpi"><span>Totali</span><strong>{nf.format(portal.AllTime)}</strong></div>
        </div>
        <section class="card">
          <h2>{days === 1 ? 'Visite di oggi' : 'Visite per giorno'}</h2>
          <ColumnChart items={portal.Days.map(d => ({ label: shortDate(d.Date), value: d.Views ?? 0, second: d.Visitors ?? 0, hint: d.Date + ' · ' + (d.Views ?? 0) + ' visite' }))} every={days > 14 ? 5 : days > 7 ? 2 : 1} height={140} />
        </section>
        <section class="card"><h2>Dispositivi</h2><BarList items={(portal.Devices ?? []).map(item => ({ name: item.Name, count: item.Count, color: '#e08a1a' }))} empty="Si raccoglie dalle prossime visite." /></section>
      {:else}<p class="empty center">Caricamento…</p>{/if}

    {:else}
      <h1 class="title">Altro</h1>
      <p class="hint">Queste voci aprono il pannello completo, che su telefono si legge meglio in orizzontale.</p>
      {#each menu as group (group.label)}
        <section class="card menu">
          <h2>{group.label}</h2>
          <ul>
            {#each group.items as [id, label] (id)}
              <li><a href={'/admin#' + id}>{label}<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M9 6l6 6-6 6"/></svg></a></li>
            {/each}
          </ul>
        </section>
      {/each}
      <section class="card menu">
        <h2>Applicazione</h2>
        <ul>
          <li><button type="button" onclick={toggleTheme}>Tema {currentTheme() === 'dark' ? 'chiaro' : 'scuro'}</button></li>
          <li><a href="/spots">Portale pubblico</a></li>
          <li><button type="button" class="danger" onclick={logout}>Esci</button></li>
        </ul>
      </section>
      <p class="version">DXRegia v{dxregiaVersion}</p>
    {/if}
  </main>

  <nav class="tabs" aria-label="Menu principale">
    {#each tabs as item (item.id)}
      <button type="button" aria-current={tab === item.id ? 'page' : undefined} onclick={() => open(item.id)}>
        <svg viewBox="0 0 24 24" width="24" height="24" aria-hidden="true"><path d={item.icon}/></svg>
        <span>{item.label}</span>
      </button>
    {/each}
  </nav>
</div>

<style>
  .phone { font-size: clamp(14px, 3.9vw, 16px); --ink: #14293a; --muted: #4d6578; --line: #c9d9e5; --brand: #175c82; --green: #1fa86b; --amber: #e0a21a; --red: #d0343f;
    min-height: 100dvh; max-width: 640px; margin: 0 auto; padding-bottom: calc(76px + env(safe-area-inset-bottom)); color: var(--ink); }
  .bar { position: sticky; top: 0; z-index: 5; display: flex; align-items: center; justify-content: space-between; min-height: 48px; padding: calc(8px + env(safe-area-inset-top)) 16px 8px;
    background: #ffffffe6; backdrop-filter: blur(14px); border-bottom: 1px solid var(--line); }
  .logo { display: block; height: 29px; width: auto; }
  .stamp { display: inline-flex; align-items: center; gap: 7px; margin: 0; font-size: .85rem; font-weight: 600; color: var(--muted); font-variant-numeric: tabular-nums; }
  main { display: grid; gap: 12px; padding: 14px max(14px, env(safe-area-inset-left)) 0 max(14px, env(safe-area-inset-right)); }
  .title { margin: 4px 2px 0; font-size: 1.5rem; text-align: center; }
  .group { margin: 10px 2px 0; font-size: 1.15rem; text-align: center; }
  .card { background: #ffffffe6; border: 1px solid var(--line); border-radius: 16px; padding: 14px 16px; box-shadow: 0 2px 10px #143a5914; min-width: 0; }
  .card h2 { display: flex; align-items: center; justify-content: center; flex-wrap: wrap; text-align: center; gap: 4px 10px; margin: 0 0 10px; font-size: 1.02rem; color: #143a60; }
  .badge { padding: 2px 10px; border-radius: 999px; background: #dcebf6; color: #143a60; font-size: .8rem; font-weight: 700; }
  .empty { margin: 2px 0; text-align: center; color: var(--muted); } .empty.center { text-align: center; }
  .dot { width: 10px; height: 10px; flex: none; border-radius: 50%; background: #708391; display: inline-block; }
  .dot[data-tone=green] { background: var(--green); box-shadow: 0 0 0 3px #1fa86b33; }
  .dot[data-tone=amber] { background: var(--amber); box-shadow: 0 0 0 3px #e0a21a33; }
  .dot[data-tone=red] { background: var(--red); box-shadow: 0 0 0 3px #d0343f33; }


  .kpis { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .kpi { display: grid; gap: 1px; justify-items: center; text-align: center; align-content: center; padding: 12px 10px; background: #ffffffe6; border: 1px solid var(--line); border-radius: 14px; border-top: 5px solid #8fb0c8; min-width: 0; }
  .kpi span { font-size: .72rem; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; color: var(--muted); }
  .kpi strong { font-size: 2rem; line-height: 1.15; color: #143a60; font-variant-numeric: tabular-nums; }
  .kpis.tight .kpi strong { font-size: 1.6rem; }

  .meter { margin-bottom: 12px; }
  .meter p { display: flex; justify-content: space-between; margin: 0 0 5px; font-weight: 600; }
  .track { height: 12px; border-radius: 6px; background: #dbe7f0; overflow: hidden; }
  .track i { display: block; height: 100%; border-radius: 6px; background: #708391; transition: width .5s; }
  .track i[data-tone=green] { background: var(--green); } .track i[data-tone=amber] { background: var(--amber); } .track i[data-tone=red] { background: var(--red); }
  .meter small { color: var(--muted); }
  .facts { margin: 4px 0 0; display: grid; gap: 6px; }
  .facts div { display: flex; justify-content: space-between; gap: 10px; padding-top: 6px; border-top: 1px solid #e3edf4; }
  .facts dt { color: var(--muted); } .facts dd { margin: 0; font-weight: 700; font-variant-numeric: tabular-nums; }

  .rows { margin: 0; padding: 0; list-style: none; display: grid; gap: 12px; }
  .rows li { display: grid; grid-template-columns: auto 1fr; align-items: start; gap: 12px; padding-bottom: 12px; border-bottom: 1px solid #e3edf4; }
  .rows li:last-child { padding-bottom: 0; border-bottom: 0; }
  .rows .dot { margin-top: 6px; }
  .rows b { display: block; }
  .rows small { display: block; color: var(--muted); overflow-wrap: anywhere; }
  .partners { margin: 0; padding: 0; list-style: none; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 6px 14px; }
  .partners li { display: grid; grid-template-columns: auto 1fr; align-items: center; column-gap: 8px; padding: 6px 0; border-bottom: 1px solid #e3edf4; }
  .partners .call { font-weight: 700; overflow-wrap: anywhere; }
  .partners small { grid-column: 2; color: var(--muted); }

  .chips { margin: 0; padding: 0; list-style: none; display: flex; flex-wrap: wrap; justify-content: center; gap: 6px; }
  .chips li { display: inline-flex; align-items: center; gap: 6px; padding: 5px 9px; border: 1px solid var(--line); border-radius: 999px; background: #f4f9fc; font-weight: 700; font-size: .88rem; }
  .chips img { border: 1px solid #15263826; border-radius: 2px; }
  .chips i { font-style: normal; font-size: .7rem; font-weight: 800; color: var(--brand); }
  .more { width: 100%; margin-top: 10px; min-height: 44px; border: 1px solid var(--line); border-radius: 12px; background: #fff; color: var(--brand); font-weight: 700; font-size: .95rem; }



  .periods { display: flex; gap: 8px; }
  .periods button { flex: 1; min-height: 44px; border: 1px solid var(--line); border-radius: 12px; background: #fff; color: var(--brand); font-weight: 700; }
  .periods button[aria-pressed=true] { background: var(--brand); border-color: var(--brand); color: #fff; }

  .hint { margin: 0 2px; text-align: center; color: var(--muted); font-size: .9rem; }
  .menu ul { margin: 0; padding: 0; list-style: none; }
  .menu li + li { border-top: 1px solid #e3edf4; }
  .menu a, .menu button { display: flex; align-items: center; justify-content: space-between; width: 100%; min-height: 48px; padding: 0; border: 0; background: none; color: var(--ink); font: inherit; font-weight: 600; text-decoration: none; text-align: left; }
  .menu svg { fill: none; stroke: #7d93a5; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
  .menu .danger { color: var(--red); }
  .version { text-align: center; color: var(--muted); font-size: .8rem; margin: 4px 0 0; }

  .tabs { position: fixed; left: 0; right: 0; bottom: 0; z-index: 10; display: flex; max-width: 640px; margin: 0 auto; padding-bottom: env(safe-area-inset-bottom);
    background: #ffffffee; backdrop-filter: blur(14px); border-top: 1px solid var(--line); }
  .phone button { box-shadow: none; filter: none; }
  .phone button:hover:not(:disabled), .phone button:active:not(:disabled) { transform: none; filter: none; }
  .tabs button { border-radius: 0; background: none; flex: 1; display: grid; justify-items: center; gap: 2px; padding: 8px 4px 9px; min-height: 60px; border: 0; background: none; color: #6a8196; font: inherit; font-size: .78rem; font-weight: 700; }
  .tabs svg { fill: none; stroke: currentColor; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
  .tabs button[aria-current=page] { color: var(--brand); }
  @media (prefers-reduced-motion: reduce) { .track i { transition: none; } }
  @media (min-width: 420px) { .partners { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
