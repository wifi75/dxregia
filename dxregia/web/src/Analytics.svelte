<script lang="ts">
  import { onMount } from 'svelte'
  import { dxccFlags } from './dxccFlags'
  import ColumnChart from './ColumnChart.svelte'
  import BarList from './BarList.svelte'

  type Item = { Name: string; Count: number; DXCC?: number }
  type Point = { Date: string; Views?: number; Visitors?: number; Spots?: number }
  type Report = {
    GeneratedAt: string; Days: number
    Portal: { AllTime: number; Active: number; Views: number; Visitors: number; Days: Point[]; Hours: number[]; Pages: Item[] | null; Referrers: Item[] | null; Languages: Item[] | null; Devices: Item[] | null }
    Spots: { Total: number; Days: Point[]; Hours: number[]; Bands: Item[] | null; Modes: Item[] | null; Entities: Item[] | null; TopDX: Item[] | null; TopSpotters: Item[] | null }
  }

  const periods = [[1, '24 ore'], [7, '7 giorni'], [30, '30 giorni']] as const
  let days = $state<number>(7)
  let report = $state<Report>()
  let failed = $state(false)
  const number = new Intl.NumberFormat('it-IT')
  const time = new Intl.DateTimeFormat('it-IT', { timeStyle: 'medium' })

  const bandColors: Record<string, string> = {
    '160m': '#e8a58f', '80m': '#eeb878', '60m': '#e6c98f', '40m': '#e2cf6b', '30m': '#a9cc6a', '20m': '#f2b254',
    '17m': '#b99be6', '15m': '#6fc6dc', '12m': '#8fb0ee', '10m': '#7fcb7a', '6m': '#ee9cab', 'VHF+': '#a8bccb', 'Altre': '#b9c5ce',
  }
  const languageNames: Record<string, string> = { it: 'Italiano', en: 'English', es: 'Español', de: 'Deutsch', fr: 'Français', pt: 'Português', altre: 'Altre lingue' }
  const pageNames: Record<string, string> = { '/spots': 'Spot in diretta', '/contatti': 'Contatti' }

  async function load() {
    try {
      const response = await fetch('/api/analytics?days=' + days, { cache: 'no-store' })
      if (!response.ok) throw new Error()
      report = await response.json()
      failed = false
    } catch {
      failed = true
    }
  }

  function choose(value: number) {
    days = value
    report = undefined
    void load()
  }

  onMount(() => {
    void load()
    const timer = setInterval(load, 15000)
    return () => clearInterval(timer)
  })

  const shortDate = (date: string) => date.slice(8, 10) + '/' + date.slice(5, 7)
  const hours = (values: number[]) => values.map((value, hour) => ({ label: String(hour).padStart(2, '0'), value, hint: String(hour).padStart(2, '0') + ':00 UTC · ' + number.format(value) }))
  const items = (list: Item[] | null) => list ?? []

  let portal = $derived(report?.Portal)
  let spots = $derived(report?.Spots)
  let topMode = $derived(items(spots?.Modes ?? null)[0])
  let perHour = $derived(spots ? Math.round((spots.Total / (days * 24)) * 10) / 10 : 0)
</script>

<div class="head">
  <h1>Analytics</h1>
  <div class="tools">
    <div class="chips" role="group" aria-label="Periodo">
      {#each periods as [value, label]}
        <button class="chip" type="button" aria-pressed={days === value} onclick={() => choose(value)}>{label}</button>
      {/each}
    </div>
    <span class="stamp"><span class="led" data-tone={failed ? 'red' : report ? 'green' : 'neutral'}></span>{failed ? 'Dati non disponibili' : report ? 'Aggiornato ' + time.format(new Date(report.GeneratedAt)) : 'Caricamento…'}</span>
  </div>
</div>

{#if report && portal && spots}
  <h2 class="section"><span class="mark" style:background="#2f78b3"></span>Portale pubblico</h2>
  <div class="kpis">
    <div class="kpi glass-panel"><span class="label">Visite</span><strong>{number.format(portal.Views)}</strong><span class="sub">nel periodo</span></div>
    <div class="kpi glass-panel"><span class="label">Visitatori unici</span><strong>{number.format(portal.Visitors)}</strong><span class="sub">somma dei giorni</span></div>
    <div class="kpi glass-panel"><span class="label">Attivi ora</span><strong>{number.format(portal.Active)}</strong><span class="sub">schede aperte (90 s)</span></div>
    <div class="kpi glass-panel"><span class="label">Visite totali</span><strong>{number.format(portal.AllTime)}</strong><span class="sub">dall’inizio</span></div>
  </div>

  <div class="grid">
    <section class="card glass-panel wide">
      <h3>{days === 1 ? 'Visite di oggi' : 'Visite per giorno'} <span class="legend"><i style:background="#7fb1d9"></i>visite <i style:background="#1f6fa8"></i>visitatori unici</span></h3>
      <ColumnChart items={portal.Days.map(day => ({ label: shortDate(day.Date), value: day.Views ?? 0, second: day.Visitors ?? 0, hint: day.Date + ' · ' + number.format(day.Views ?? 0) + ' visite · ' + number.format(day.Visitors ?? 0) + ' visitatori' }))} every={days > 14 ? 3 : 1} />
    </section>
    <section class="card glass-panel">
      <h3>Orari di visita <span class="legend">UTC</span></h3>
      <ColumnChart items={hours(portal.Hours)} height={150} every={3} color="#8fcdb0" />
    </section>
    <section class="card glass-panel"><h3>Pagine</h3><BarList items={items(portal.Pages).map(item => ({ name: pageNames[item.Name] ?? item.Name, count: item.Count }))} /></section>
    <section class="card glass-panel"><h3>Provenienza</h3><BarList items={items(portal.Referrers).map(item => ({ name: item.Name, count: item.Count, color: '#1fa86b' }))} /></section>
    <section class="card glass-panel"><h3>Lingua del browser</h3><BarList items={items(portal.Languages).map(item => ({ name: languageNames[item.Name] ?? item.Name, count: item.Count, color: '#7a64c8' }))} empty="Si raccoglie dalle prossime visite." /></section>
    <section class="card glass-panel"><h3>Dispositivi</h3><BarList items={items(portal.Devices).map(item => ({ name: item.Name, count: item.Count, color: '#e08a1a' }))} empty="Si raccoglie dalle prossime visite." /></section>
  </div>

  <h2 class="section"><span class="mark" style:background="#1fa86b"></span>Attività del cluster</h2>
  <div class="kpis">
    <div class="kpi glass-panel"><span class="label">Spot</span><strong>{number.format(spots.Total)}</strong><span class="sub">nel periodo</span></div>
    <div class="kpi glass-panel"><span class="label">Spot all’ora</span><strong>{number.format(perHour)}</strong><span class="sub">media</span></div>
    <div class="kpi glass-panel"><span class="label">Entità DXCC</span><strong>{number.format(items(spots.Entities).length)}{items(spots.Entities).length >= 12 ? '+' : ''}</strong><span class="sub">più spottate</span></div>
    <div class="kpi glass-panel"><span class="label">Modo più usato</span><strong>{topMode?.Name ?? '—'}</strong><span class="sub">{topMode ? number.format(topMode.Count) + ' spot' : 'nessun dato'}</span></div>
  </div>

  <div class="grid">
    <section class="card glass-panel" class:wide={days > 1} class:full={days === 1}>
      <h3>{days === 1 ? 'Spot per ora' : 'Spot per giorno'} <span class="legend">{days === 1 ? 'UTC, ultime 24 ore' : 'UTC'}</span></h3>
      {#if days === 1}
        <ColumnChart items={hours(spots.Hours)} color="#8fcdb0" every={2} />
      {:else}
        <ColumnChart items={spots.Days.map(day => ({ label: shortDate(day.Date), value: day.Spots ?? 0, hint: day.Date + ' · ' + number.format(day.Spots ?? 0) + ' spot' }))} color="#8fcdb0" every={days > 14 ? 3 : 1} />
      {/if}
    </section>
    {#if days > 1}
      <section class="card glass-panel">
        <h3>Spot per ora del giorno <span class="legend">UTC</span></h3>
        <ColumnChart items={hours(spots.Hours)} height={150} every={3} color="#8fcdb0" />
      </section>
    {/if}
    <section class="card glass-panel"><h3>Bande</h3><BarList items={items(spots.Bands).map(item => ({ name: item.Name, count: item.Count, color: bandColors[item.Name] ?? '#2f78b3' }))} /></section>
    <section class="card glass-panel"><h3>Modi</h3><BarList items={items(spots.Modes).map(item => ({ name: item.Name, count: item.Count, color: '#7a64c8' }))} empty="Nessun modo indicato nei commenti." /></section>
    <section class="card glass-panel half"><h3>Entità più spottate</h3><BarList items={items(spots.Entities).map(item => ({ name: item.Name, count: item.Count, flag: item.DXCC ? dxccFlags[item.DXCC] : undefined }))} total={spots.Total} /></section>
    <section class="card glass-panel half"><h3>DX più spottati</h3><BarList items={items(spots.TopDX).map(item => ({ name: item.Name, count: item.Count, color: '#1fa86b' }))} total={spots.Total} /></section>
    <section class="card glass-panel half"><h3>Spottatori più attivi</h3><BarList items={items(spots.TopSpotters).map(item => ({ name: item.Name, count: item.Count, color: '#e08a1a' }))} total={spots.Total} /></section>
  </div>
{:else if failed}
  <p class="glass-panel notice" role="alert">Impossibile leggere i dati di analytics. Riprova tra poco.</p>
{:else}
  <p class="glass-panel notice" role="status">Caricamento dei dati…</p>
{/if}

<style>
  .head { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 10px 16px; margin-bottom: 4px; }
  h1 { margin: 0; font-size: 1.65rem; line-height: 1.2; }
  .tools { display: flex; align-items: center; flex-wrap: wrap; gap: 10px 14px; }
  .chips { display: flex; gap: 6px; }
  .chip { min-height: 32px; padding: 0 .9rem; border-radius: 999px; font-size: .85rem; --button-face: #fff; --button-edge: #a4bccb; --button-ink: #274760; box-shadow: none; }
  .chip[aria-pressed=true] { --button-face: #175c82; --button-edge: #0c3a56; --button-ink: #fff; }
  .stamp { display: inline-flex; align-items: center; gap: 8px; font-size: .85rem; color: #2f4b5f; font-weight: 600; }
  .led { width: 10px; height: 10px; border-radius: 50%; background: #708391; }
  .led[data-tone=green] { background: #1fa86b; box-shadow: 0 0 0 3px #1fa86b33; }
  .led[data-tone=red] { background: #d0343f; box-shadow: 0 0 0 3px #d0343f33; }
  .section { display: flex; align-items: center; gap: 10px; margin: 22px 0 10px; font-size: 1.1rem; color: #143a60; }
  .mark { width: 5px; height: 20px; border-radius: 3px; }
  .kpis { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 190px), 1fr)); gap: 12px; }
  .kpi { display: grid; gap: 2px; padding: 12px 16px; }
  .kpi .label { color: #2f4b5f; font-size: .78rem; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; }
  .kpi strong { font-size: 2rem; line-height: 1.15; color: #143a60; font-variant-numeric: tabular-nums; }
  .kpi .sub { color: #4d6578; font-size: .85rem; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr)); gap: 12px; margin-top: 12px; align-items: start; }
  .card { padding: 14px 16px; min-width: 0; }
  .card.wide, .card.full, .card.half { grid-column: 1 / -1; }
  h3 { display: flex; align-items: baseline; justify-content: space-between; gap: 10px; flex-wrap: wrap; margin: 0 0 12px; font-size: 1rem; color: #143a60; }
  .legend { display: inline-flex; align-items: center; gap: 6px; color: #4d6578; font-size: .78rem; font-weight: 500; }
  .legend i { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-left: 6px; }
  .notice { margin-top: 14px; padding: 16px; }
  @media (min-width: 1100px) {
    .grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }
    .card.wide { grid-column: span 3; }
    .card.half { grid-column: span 2; }
  }
</style>
