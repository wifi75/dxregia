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
    <h1 id="spots-title">Spot DX</h1>
    <a href="/contatti">Contatta il sysop o richiedi un collegamento</a>
    <button type="button" onclick={() => { paused = !paused; if (!paused) void refresh() }}>
      {paused ? 'Riprendi aggiornamenti' : 'Metti in pausa'}
    </button>
  </div>
  <div class="filters" aria-label="Filtra per banda">
    <button type="button" aria-pressed={selectedBand === ''} onclick={() => selectedBand = ''}>Tutte le bande</button>
    {#each bands as [band]}
      <button type="button" aria-pressed={selectedBand === band} onclick={() => selectedBand = band}>{band}</button>
    {/each}
  </div>
  <div class="filters">
    <input type="search" bind:value={query} aria-label="Cerca negli spot" placeholder="Nominativo o commento" />
    <select bind:value={mode} aria-label="Modo riportato nel commento">
      <option value="">Tutti i modi</option>
      {#each ['CW', 'SSB', 'FT8', 'FT4', 'RTTY', 'PSK31', 'USB', 'LSB'] as value}<option>{value}</option>{/each}
    </select>
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
        <thead><tr><th>Spotter</th><th>kHz</th><th>Modo</th><th>DX</th><th>Entità DXCC</th><th>Commento</th><th>UTC</th></tr></thead>
        <tbody>
          {#each filtered as spot}
            <tr>
              <td>{spot.Spotter}</td>
              <td><strong class="frequency" style:background={colors[bandFor(spot.Frequency)] || '#edf3f9'}>{khz.format(spot.Frequency)}</strong></td>
              <td><span class="mode">{modeFor(spot.Comment) || '—'}</span></td>
              <td><strong>{spot.Call}</strong></td>
              <td>
                {#if entities[spot.DXCC]}
                  <img src={'/flags/' + entities[spot.DXCC][1] + '.png'} alt="" width="26" height="17" />
                  {entities[spot.DXCC][0]}
                {:else}
                  {spot.Entity || (spot.DXCC > 0 && spot.DXCC !== 666 ? 'DXCC ' + spot.DXCC : '—')}
                {/if}
              </td>
              <td>{spot.Comment}</td>
              <td>{utc.format(new Date(spot.At * 1000))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if available}
    <p>{spots.length ? 'Nessuno spot corrisponde ai filtri.' : 'Nessuno spot recente.'}</p>
  {/if}
  {#if views !== undefined}<p>Visite al portale: {khz.format(views)}</p>{/if}
</main>

<style>
  main { width: 100%; padding: 22px clamp(14px, 2vw, 36px); }
  .heading, .filters { display: flex; align-items: center; flex-wrap: wrap; gap: 7px; margin-bottom: 10px; }
  .heading { justify-content: space-between; }
  h1 { margin: 0; font-size: 1.65rem; }
  input, select { border: 1px solid #aebfcc; border-radius: 4px; background: white; color: #173653; padding: 6px 10px; min-height: 38px; font: inherit; }
  .filters button { min-height: 36px; padding: .3rem .75rem; border-radius: 9px; }
  .filters button[aria-pressed=false] { --button-face: #fff; --button-edge: #a7bdcd; --button-ink: #234968; }
  .filters button[aria-pressed=true] { --button-face: #17648e; --button-edge: #103f61; }
  :focus-visible { outline: 2px solid #175586; outline-offset: 2px; }
  .table-wrap { overflow: auto; margin-top: 16px; }
  table { width: 100%; border-collapse: collapse; text-align: left; white-space: nowrap; }
  th, td { padding: 7px 10px; border-bottom: 1px solid #d7e2ed; }
  th { background: #dce9f5; color: #174e7a; border-bottom: 2px solid #4078a6; font-size: .85rem; }
  tbody tr:nth-child(odd) { background: white; }
  tbody tr:nth-child(even) { background: #eaf0f6; }
  .frequency, .mode { padding: 2px 6px; border-radius: 4px; }
  .frequency, td:last-child { font-variant-numeric: tabular-nums; }
  img { object-fit: cover; vertical-align: middle; margin-right: 6px; }
  .mode { color: #245d4c; font-size: .8rem; }
  @media (max-width: 600px) { button, input, select { min-height: 44px; } }
</style>
