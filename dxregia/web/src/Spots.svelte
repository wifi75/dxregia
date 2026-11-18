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
      <button class="secondary" type="button" onclick={() => { paused = !paused; if (!paused) void refresh() }}>
        {paused ? 'Riprendi' : 'Pausa'}
      </button>
    </div>
  </div>
  <div class="toolbar">
  <div class="filters" aria-label="Filtra per banda">
    <button type="button" aria-pressed={selectedBand === ''} onclick={() => selectedBand = ''}>Tutte le bande</button>
    {#each bands as [band]}
      <button type="button" aria-pressed={selectedBand === band} onclick={() => selectedBand = band}>{band}</button>
    {/each}
  </div>
  <div class="filters search">
    <input type="search" bind:value={query} aria-label="Cerca negli spot" placeholder="Nominativo o commento" />
    <select bind:value={mode} aria-label="Modo riportato nel commento">
      <option value="">Tutti i modi</option>
      {#each ['CW', 'SSB', 'FT8', 'FT4', 'RTTY', 'PSK31', 'USB', 'LSB'] as value}<option>{value}</option>{/each}
    </select>
  </div>
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
  {#if views !== undefined}<p class="visits">Visite al portale: {khz.format(views)}</p>{/if}
</main>

<style>
  main { width: 100%; padding: 0 0 24px; }
  .heading, .toolbar, .table-wrap, main > p { margin-inline: clamp(12px, 2vw, 32px); }
  .heading { display: flex; align-items: center; justify-content: space-between; gap: 12px; flex-wrap: wrap; min-height: 71px; }
  h1 { margin: 0; font-size: clamp(1.35rem, 2vw, 1.95rem); color: #14334d; letter-spacing: -.025em; }
  .feed-controls, .filters { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
  .feed-state { display: inline-flex; align-items: center; gap: .5rem; color: #53687a; font-weight: 600; font-size: .85rem; }
  .feed-state::before { content: ''; width: 8px; height: 8px; border-radius: 50%; background: #ad7118; }
  .feed-state.live::before { background: #20845a; box-shadow: 0 0 0 3px #dcefe5; }
  .toolbar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: .8rem; padding: 12px 14px 14px; background: #edf3f8; border: 1px solid #ccdbe7; border-radius: 8px 8px 0 0; }
  .filters button { min-height: 34px; min-width: 48px; padding: .3rem .68rem; border-radius: 6px; font-variant-numeric: tabular-nums; }
  .filters button[aria-pressed=false] { --button-face: #fff; --button-edge: #acbfce; --button-ink: #274760; }
  .filters button[aria-pressed=true] { --button-face: #125e8b; --button-edge: #104165; }
  input, select { border: 1px solid #acbfce; border-radius: 6px; background: #fff; color: #173653; padding: 6px 10px; min-height: 36px; font: inherit; }
  .table-wrap { overflow: auto; background: #fff; border: 1px solid #ccdbe7; border-top: 0; border-radius: 0 0 8px 8px; }
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
  @media (max-width: 600px) {
    .filters button, input, select { min-height: 44px; }
    .search, .search input { width: 100%; }
  }
</style>
