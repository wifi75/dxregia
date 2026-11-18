<script lang="ts">
  import { onMount } from 'svelte'

  type Spot = { Frequency: number; Call: string; At: number; Comment: string; Spotter: string; DXCC: number; Entity?: string }
  const bands = [
    ['160 m', 1800, 2000], ['80 m', 3500, 4000], ['60 m', 5250, 5450],
    ['40 m', 7000, 7300], ['30 m', 10100, 10150], ['20 m', 14000, 14350],
    ['17 m', 18068, 18168], ['15 m', 21000, 21450], ['12 m', 24890, 24990],
    ['10 m', 28000, 29700], ['6 m', 50000, 54000], ['VHF+', 70000, Infinity],
  ] as const
  const utc = new Intl.DateTimeFormat('it-IT', { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' })
  const khz = new Intl.NumberFormat('it-IT', { maximumFractionDigits: 3 })
  let spots = $state<Spot[]>([])
  let available = $state<boolean>()
  let selectedBand = $state('')
  let mode = $state('')
  let query = $state('')
  let paused = $state(false)
  let reading = false
  let lastUpdate = $state<Date>()
  let views = $state<number>()
  const bandFor = (frequency: number) => bands.find(([, low, high]) => frequency >= low && frequency <= high)?.[0] ?? ''
  const modeFor = (comment: string) => comment.match(/\b(FT8|FT4|SSB|CW|RTTY|PSK31|USB|LSB)\b/i)?.[1].toUpperCase() ?? ''
  const colors: Record<string, string> = {
    '160 m': '#f1bfae', '80 m': '#f6cf9b', '60 m': '#f2dbaf', '40 m': '#f2e39a',
    '30 m': '#cbe19c', '20 m': '#ffd48a', '17 m': '#dfcbf7', '15 m': '#aee1ed',
    '12 m': '#c6dafa', '10 m': '#c2e7a7', '6 m': '#facdd4', 'VHF+': '#d8e5ef',
  }
  const entities: Record<number, [string, string]> = {
    23: ['Portogallo', 'pt'], 25: ['Madeira', 'pt'], 36: ['Canarie', 'es'], 95: ['Giappone', 'jp'],
    124: ['Argentina', 'ar'], 129: ['Austria', 'at'], 142: ['Brasile', 'br'], 226: ['Stati Uniti', 'us'], 277: ['Figi', 'fj'],
  }
  let filtered = $derived(spots.filter(spot =>
    (!selectedBand || bandFor(spot.Frequency) === selectedBand) &&
    (!mode || modeFor(spot.Comment) === mode) &&
    (!query.trim() || [spot.Call, spot.Spotter, spot.Comment].some(value => value.toLowerCase().includes(query.trim().toLowerCase())))
  ))

  async function refresh() {
    if (paused || reading) return
    reading = true
    try {
      const response = await fetch('/api/spots', { signal: AbortSignal.timeout(6000) })
      if (!response.ok) throw new Error()
      const data = await response.json()
      if (!paused) {
        spots = data.Spots
        available = data.Available
        if (available) lastUpdate = new Date()
      }
    } catch {
      if (!paused) available = false
    } finally {
      reading = false
    }
  }
  onMount(() => {
    void refresh()
    void fetch('/api/visits').then(async response => {
      if (response.ok) views = (await response.json()).Views
    }).catch(() => {})
    const timer = setInterval(refresh, 5000)
    return () => clearInterval(timer)
  })
</script>

<main aria-labelledby="spots-title">
  <div class="heading">
    <h1 id="spots-title">Cluster DX</h1>
    <div class="feed-controls">
      <span class="feed-state" class:live={available && !paused}>
        {paused ? 'In pausa' : available === false ? 'Feed non disponibile' : available ? 'In diretta' : 'Connessione…'}
      </span>
      {#if lastUpdate}<time title="Ultimo aggiornamento del feed" datetime={lastUpdate.toISOString()}>{utc.format(lastUpdate)} UTC</time>{/if}
      <button class="secondary" type="button" onclick={() => { paused = !paused; if (!paused) void refresh() }}>
        {paused ? 'Riprendi' : 'Pausa'}
      </button>
    </div>
  </div>
  <div class="desk">
    <aside class="receiver" aria-label="Filtri degli spot">
      <section>
        <h2>Bande</h2>
        <div class="band-grid">
          <button class="all-bands" type="button" aria-pressed={selectedBand === ''} onclick={() => selectedBand = ''}>Tutte le bande</button>
          {#each bands as [band]}
            <button type="button" aria-pressed={selectedBand === band} onclick={() => selectedBand = band}>
              <span class="band-color" style:background={colors[band]}></span>{band}
            </button>
          {/each}
        </div>
      </section>
      <section class="mode-filter">
        <h2>Modo</h2>
        <div class="band-grid">
          <button class="all-bands" type="button" aria-pressed={mode === ''} onclick={() => mode = ''}>Tutti i modi</button>
          {#each ['CW', 'SSB', 'FT8', 'FT4', 'RTTY', 'PSK31', 'USB', 'LSB'] as value}
            <button type="button" aria-pressed={mode === value} onclick={() => mode = value}>{value}</button>
          {/each}
        </div>
      </section>
    </aside>
    <section class="log" aria-label="Registro degli spot">
      <div class="toolbar">
        <input type="search" bind:value={query} aria-label="Cerca negli spot" placeholder="Cerca nominativo o commento" />
        <span>{filtered.length} {filtered.length === 1 ? 'spot mostrato' : 'spot mostrati'}</span>
      </div>
  {#if available === false}
    <p role="status">Archivio spot non disponibile.</p>
  {:else if available === undefined}
    <p role="status">Caricamento degli spot…</p>
  {/if}
  {#if filtered.length}
    <!-- svelte-ignore a11y_no_noninteractive_tabindex (scrollable table needs keyboard focus) -->
    <div class="table-wrap" tabindex="0" role="region" aria-label="Elenco degli spot">
      <table>
        <thead><tr><th scope="col">DE</th><th scope="col">Frequenza <small>kHz</small></th><th>Modo</th><th>DX</th><th>Entità DXCC</th><th>Commento</th><th>UTC</th></tr></thead>
        <tbody>
          {#each filtered as spot}
            <tr>
              <td>{spot.Spotter}</td>
              <td><strong class="frequency" style:background={colors[bandFor(spot.Frequency)] || '#edf3f9'}>{khz.format(spot.Frequency)}</strong></td>
              <td><span class="mode" data-mode={modeFor(spot.Comment)}>{modeFor(spot.Comment) || '—'}</span></td>
              <td><strong>{spot.Call}</strong></td>
              <td>
                {#if entities[spot.DXCC]}
                  <img src={'/flags/' + entities[spot.DXCC][1] + '.png'} alt="" width="26" height="17" />
                  {entities[spot.DXCC][0]}
                {:else}
                  {spot.Entity || (spot.DXCC > 0 && spot.DXCC !== 666 ? 'DXCC ' + spot.DXCC : '—')}
                {/if}
              </td>
              <td class="comment">{spot.Comment}</td>
              <td>{utc.format(new Date(spot.At * 1000))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if available}
    <p>{spots.length ? 'Nessuno spot corrisponde ai filtri.' : 'Nessuno spot recente.'}</p>
  {/if}
  {#if views !== undefined}<p class="visits">Visite al portale {khz.format(views)}</p>{/if}
    </section>
  </div>
</main>

<style>
  main { width: 100%; padding: 0 clamp(12px, 2vw, 32px) 24px; }
  .heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; min-height: 72px; }
  h1 { margin: 0; font-size: clamp(1.65rem, 3vw, 2.25rem); color: #102e42; letter-spacing: -.04em; }
  .feed-controls { display: flex; align-items: center; flex-wrap: wrap; gap: 14px; }
  .feed-controls time { color: #496174; font-variant-numeric: tabular-nums; }
  .feed-state { display: inline-flex; align-items: center; gap: .45rem; color: #395468; font-weight: 650; }
  .feed-state::before { content: ''; width: 8px; height: 8px; border-radius: 50%; background: #ad7118; }
  .feed-state.live::before { background: #20845a; box-shadow: 0 0 0 3px #dcefe5; animation: signal-pulse 2.6s ease-out infinite; }
  @keyframes signal-pulse { 0%, 100% { box-shadow: 0 0 0 2px #20845a24; } 48% { box-shadow: 0 0 0 5px #20845a0a; } }
  @media (prefers-reduced-motion: reduce) { .feed-state.live::before { animation: none; } }
  .desk { display: grid; grid-template-columns: 218px minmax(0, 1fr); align-items: start; gap: 18px; }
  .receiver { position: sticky; top: 12px; min-width: 0; border: 1px solid #b7c9d5; border-top: 4px solid #186187; border-radius: 7px; background: linear-gradient(160deg, #ffffffdf, #e2eef5cc); backdrop-filter: blur(12px); box-shadow: 0 8px 22px #26526d16; padding: 14px; }
  .receiver h2 { margin: 0 0 .6rem; color: #214963; font-size: .78rem; text-transform: uppercase; letter-spacing: .1em; }
  .mode-filter { margin-top: 16px; padding-top: 14px; border-top: 1px solid #c8d7e0; }
  .band-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .55rem; }
  .band-grid button { display: flex; justify-content: center; gap: .4rem; align-items: center; min-width: 0; min-height: 34px; padding: .35rem .4rem; border-radius: 5px; font-size: .83rem; }
  .band-grid .all-bands { grid-column: 1 / -1; }
  .band-grid button[aria-pressed=false] { --button-face: #fff; --button-edge: #a4bccb; --button-ink: #274760; }
  .band-grid button[aria-pressed=true] { --button-face: #175c82; --button-edge: #0c3a56; }
  .band-color { width: 9px; height: 9px; border-radius: 2px; box-shadow: inset 0 0 0 1px #223c4c30; }
  .log { min-width: 0; background: #fff; border: 1px solid #b7c9d5; border-radius: 7px; box-shadow: 0 8px 24px #26526d12; }
  .log > p { margin: 14px 16px; }
  .toolbar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: .65rem; padding: 12px 16px; border-bottom: 1px solid #ccdbe7; background: #f5f8fa; border-radius: 6px 6px 0 0; color: #496174; }
  input { width: min(100%, 350px); min-height: 40px; border: 1px solid #acbfce; border-radius: 6px; background: #fff; color: #173653; padding: 6px 10px; font: inherit; }
  .table-wrap { overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; white-space: nowrap; }
  th, td { padding: 8px clamp(10px, 1.15vw, 20px); border-bottom: 1px solid #dce5ed; }
  th { background: #f9fbfd; color: #285879; border-bottom: 2px solid #2972a0; font-size: .78rem; letter-spacing: .065em; text-transform: uppercase; }
  th small { color: #536b7d; text-transform: none; }
  tbody tr:nth-child(even) { background: #edf3f8; }
  tbody tr:first-child td:first-child { box-shadow: inset 4px 0 #20845a; }
  tbody tr:hover { background: #e0eff8; }
  .frequency, .mode { display: inline-block; padding: 2px 7px; border-radius: 4px; }
  .frequency { min-width: 7ch; text-align: right; color: #142b39; }
  .frequency, td:last-child { font-variant-numeric: tabular-nums; }
  td:first-child, td:last-child, .comment { color: #465f72; }
  .comment { width: 100%; }
  img { object-fit: cover; vertical-align: middle; margin-right: 7px; border: 1px solid #15263826; border-radius: 2px; }
  .mode { color: #4c6374; font-size: .76rem; font-weight: 650; }
  .mode[data-mode=CW] { background: #fff0d7; color: #7d430e; }
  .mode[data-mode=SSB], .mode[data-mode=USB], .mode[data-mode=LSB] { background: #d7f0e3; color: #1a6046; }
  .mode[data-mode=FT8], .mode[data-mode=FT4], .mode[data-mode=RTTY], .mode[data-mode=PSK31] { background: #eee4ff; color: #573894; }
  .visits { margin-block: 14px 0; color: #526c7f; font-size: .85rem; }
  @media (max-width: 800px) {
    .desk { grid-template-columns: minmax(0, 1fr); gap: 12px; }
    .receiver { position: static; }
    .band-grid { grid-template-columns: repeat(auto-fit, minmax(53px, 1fr)); }
    .band-grid .all-bands { grid-column: auto; }
    .band-grid button { min-height: 40px; }
  }
</style>
