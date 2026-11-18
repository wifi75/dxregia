<script lang="ts">
  import { dxccFlags } from './dxccFlags'
  import { onMount } from 'svelte'
  import { t, locale, lang } from './i18n.svelte'
  import { soundOn, chirp, unlockAudio } from './sound.svelte'

  type Spot = { Frequency: number; Call: string; At: number; Comment: string; Spotter: string; DXCC: number; Entity?: string }
  const bands = [
    ['160 m', 1800, 2000], ['80 m', 3500, 4000], ['60 m', 5250, 5450],
    ['40 m', 7000, 7300], ['30 m', 10100, 10150], ['20 m', 14000, 14350],
    ['17 m', 18068, 18168], ['15 m', 21000, 21450], ['12 m', 24890, 24990],
    ['10 m', 28000, 29700], ['6 m', 50000, 54000], ['VHF+', 70000, Infinity],
  ] as const
  let utc = $derived(new Intl.DateTimeFormat(locale(), { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }))
  let khz = $derived(new Intl.NumberFormat(locale(), { maximumFractionDigits: 3 }))
  let spots = $state<Spot[]>([])
  let available = $state<boolean>()
  let bandSel = $state<string[]>([])
  let modeSel = $state<string[]>([])
  let noFt8 = $state(false)
  const toggle = (list: string[], value: string) => list.includes(value) ? list.filter(item => item !== value) : [...list, value]
  let query = $state('')
  const spotKey = (spot: Spot) => spot.At + '|' + spot.Call + '|' + spot.Frequency
  let freshKeys = $state<Set<string>>(new Set())

  // Spedizioni DX annunciate (feed pubblico NG3K ADXO, scaricato dal nodo): icona accanto al nominativo e popup con i dettagli.
  type Adxo = { Call: string; Entity: string; Dates: string; QSL: string; Info: string; Link: string }
  let adxo = $state<{ Source: string; Entries: Adxo[] | null }>()
  let popup = $state<{ entry: Adxo; x: number; y: number }>()
  let adxoMap = $derived(new Map((adxo?.Entries ?? []).map(entry => [entry.Call, entry])))
  let adxoPrefixes = $derived((adxo?.Entries ?? []).filter(entry => /\d$/.test(entry.Call)))
  function adxoFor(call: string): Adxo | undefined {
    if (!adxoMap.size) return undefined
    const full = call.toUpperCase()
    const base = full.replace(/\/(P|QRP|M|MM|AM)$/, '')
    return adxoMap.get(full) ?? adxoMap.get(base) ?? adxoPrefixes.find(entry => base.startsWith(entry.Call))
  }
  let closeTimer: ReturnType<typeof setTimeout> | undefined
  const keepInfo = () => { clearTimeout(closeTimer) }
  const hideInfoSoon = () => { clearTimeout(closeTimer); closeTimer = setTimeout(() => { popup = undefined }, 180) }
  function openInfo(event: Event, entry: Adxo) {
    event.stopPropagation()
    clearTimeout(closeTimer)
    const button = event.currentTarget as HTMLElement
    const host = button.closest('main')!.getBoundingClientRect()
    const rect = button.getBoundingClientRect()
    const width = Math.min(380, host.width - 16)
    popup = { entry, x: Math.max(8, Math.min(rect.left - host.left - 10, host.width - width - 8)), y: rect.bottom - host.top + 8 }
  }
  async function loadAdxo() {
    try {
      const response = await fetch('/api/adxo')
      if (response.ok) adxo = await response.json()
    } catch { /* senza il feed la pagina funziona uguale */ }
  }
  let knownKeys: Set<string> | undefined
  let reading = false
  let views = $state<number>()
  const bandFor = (frequency: number) => bands.find(([, low, high]) => frequency >= low && frequency <= high)?.[0] ?? ''
  const modeFor = (comment: string) => comment.match(/\b(FT8|FT4|SSB|CW|RTTY|PSK31|USB|LSB)\b/i)?.[1].toUpperCase() ?? ''
  // Se il commento non indica il modo, lo si deduce dalla frequenza (piano di banda IARU regione 1).
  const ft8 = [1840, 3573, 5357, 7074, 10136, 14074, 18100, 21074, 24915, 28074, 50313]
  const ft4 = [3575.5, 7047.5, 10140, 14080, 18104, 21140, 24919, 28180, 50318]
  const segments: [number, number, string][] = [
    [1800, 1838, 'CW'], [1838, 1843, 'DIGI'], [1843, 2000, 'SSB'],
    [3500, 3570, 'CW'], [3570, 3600, 'DIGI'], [3600, 3800, 'SSB'],
    [5351, 5367, 'DIGI'],
    [7000, 7040, 'CW'], [7040, 7060, 'DIGI'], [7060, 7200, 'SSB'],
    [10100, 10130, 'CW'], [10130, 10150, 'DIGI'],
    [14000, 14070, 'CW'], [14070, 14099, 'DIGI'], [14112, 14350, 'SSB'],
    [18068, 18095, 'CW'], [18095, 18109, 'DIGI'], [18111, 18168, 'SSB'],
    [21000, 21070, 'CW'], [21070, 21150, 'DIGI'], [21151, 21450, 'SSB'],
    [24890, 24915, 'CW'], [24915, 24931, 'DIGI'], [24931, 24990, 'SSB'],
    [28000, 28070, 'CW'], [28070, 28190, 'DIGI'], [28225, 29700, 'SSB'],
    [50000, 50100, 'CW'], [50100, 50300, 'SSB'], [50300, 50500, 'DIGI'],
  ]
  function inferMode(khz: number) {
    if (ft8.some(value => Math.abs(khz - value) <= 1.2)) return 'FT8'
    if (ft4.some(value => Math.abs(khz - value) <= 1.2)) return 'FT4'
    return segments.find(([low, high]) => khz >= low && khz < high)?.[2] ?? ''
  }
  const modeInfo = (spot: Spot) => {
    const declared = modeFor(spot.Comment)
    return declared ? { mode: declared, inferred: false } : { mode: inferMode(spot.Frequency), inferred: true }
  }
  const colors: Record<string, string> = {
    '160 m': '#f1bfae', '80 m': '#f6cf9b', '60 m': '#f2dbaf', '40 m': '#f2e39a',
    '30 m': '#cbe19c', '20 m': '#ffd48a', '17 m': '#dfcbf7', '15 m': '#aee1ed',
    '12 m': '#c6dafa', '10 m': '#c2e7a7', '6 m': '#facdd4', 'VHF+': '#d8e5ef',
  }
  const entities: Record<number, [string]> = {
    23: ['Portogallo'], 25: ['Madeira'], 36: ['Canarie'], 95: ['Giappone'],
    124: ['Argentina'], 129: ['Austria'], 142: ['Brasile'], 226: ['Stati Uniti'], 277: ['Figi'],
  }
  const rowsKeyed = (list: Spot[]) => {
    const seen = new Map<string, number>()
    return list.map(spot => {
      const base = spotKey(spot)
      const n = (seen.get(base) ?? 0) + 1
      seen.set(base, n)
      return { spot, key: base + '#' + n, id: base }
    })
  }
  const groupOf = (value: string) => ['FT8', 'FT4', 'RTTY', 'PSK31'].includes(value) ? 'DIGI' : ['SSB', 'USB', 'LSB'].includes(value) ? 'SSB' : value
  let filtered = $derived(spots.filter(spot =>
    (!bandSel.length || bandSel.includes(bandFor(spot.Frequency))) &&
    (!modeSel.length || modeSel.includes(groupOf(modeInfo(spot).mode))) &&
    (!noFt8 || modeInfo(spot).mode !== 'FT8') &&
    (!query.trim() || [spot.Call, spot.Spotter, spot.Comment].some(value => value.toLowerCase().includes(query.trim().toLowerCase())))
  ))

  async function refresh() {
    if (reading) return
    reading = true
    try {
      const response = await fetch('/api/spots', { signal: AbortSignal.timeout(6000) })
      if (!response.ok) throw new Error()
      const data = await response.json()
      const rows: Spot[] = data.Spots ?? []
      const keys = new Set(rows.map(spotKey))
      if (knownKeys) {
        const added = [...keys].filter(key => !knownKeys!.has(key))
        if (added.length) {
          freshKeys = new Set([...freshKeys, ...added])
          // Giallo pieno per 10 secondi, poi sfuma: la scheda resta evidenziata ancora 4 secondi durante la dissolvenza.
          setTimeout(() => { freshKeys = new Set([...freshKeys].filter(key => !added.includes(key))) }, 14500)
          if (soundOn()) for (let i = 0; i < Math.min(added.length, 4); i++) chirp(i * 0.16)
        }
      }
      knownKeys = keys
      spots = rows
      available = data.Available
    } catch {
      available = false
    } finally {
      reading = false
    }
  }
  let engine = $state('')
  onMount(async () => {
    try {
      const response = await fetch('/api/engine')
      if (response.ok) {
        const info = await response.json()
        if (info.Version) engine = info.Version + (info.Build ? ' (build ' + info.Build + ')' : '')
      }
    } catch { /* il piè di pagina funziona anche senza */ }
  })
  onMount(() => {
    const unlock = () => unlockAudio()
    window.addEventListener('pointerdown', unlock, { once: true })
    window.addEventListener('keydown', unlock, { once: true })
    void refresh()
    void loadAdxo()
    const adxoTimer = setInterval(loadAdxo, 30 * 60 * 1000)
    const closePopup = () => { popup = undefined }
    window.addEventListener('click', closePopup)
    window.addEventListener('keydown', event => { if (event.key === 'Escape') popup = undefined })
    void fetch('/api/visits').then(async response => {
      if (response.ok) views = (await response.json()).Views
    }).catch(() => {})
    const timer = setInterval(refresh, 5000)
    return () => { clearInterval(timer); clearInterval(adxoTimer); window.removeEventListener('click', closePopup); window.removeEventListener('pointerdown', unlock); window.removeEventListener('keydown', unlock) }
  })
</script>

<main aria-labelledby="spots-title">
  <section class="filters glass-panel glass-enter" aria-label={t('filters')}>
    <div class="chip-row">
      <div class="chips" role="group" aria-label={t('bands')}>
        <button class="chip" type="button" aria-pressed={bandSel.length === 0} onclick={() => bandSel = []}>{t('allShort')}</button>
        {#each bands as [band]}
          <button class="chip" type="button" aria-pressed={bandSel.includes(band)} onclick={() => bandSel = toggle(bandSel, band)}>
            <span class="band-color" style:background={colors[band]}></span>{band.replace(' ', '')}
          </button>
        {/each}
      </div>
      <div class="chips" role="group" aria-label={t('modes')}>
        {#each [['CW', 'CW'], ['SSB', 'SSB'], ['DIGI', t('digi')]] as [value, label]}
          <button class="chip" type="button" aria-pressed={modeSel.includes(value)} onclick={() => modeSel = toggle(modeSel, value)}>{label}</button>
        {/each}
        <button class="chip" type="button" aria-pressed={noFt8} onclick={() => noFt8 = !noFt8}>{t('noFt8')}</button>
      </div>
    </div>
    <div class="search-row">
      <span class="search">
        <input type="search" bind:value={query} aria-label={t('searchLabel')} placeholder={t('searchShort')} title={t('searchPlaceholder')} />
        {#if query}<button class="clear" type="button" aria-label={t('clear')} title={t('clear')} onclick={() => query = ''}>×</button>{/if}
      </span>
    </div>
  </section>
  <div class="desk glass-enter">
    <section class="log glass-panel" aria-label="Registro degli spot">
  {#if available === false}
    <p role="status">{t('archiveDown')}</p>
  {:else if available === undefined}
    <p role="status">{t('loading')}</p>
  {/if}
  {#if filtered.length}
    <!-- svelte-ignore a11y_no_noninteractive_tabindex (scrollable table needs keyboard focus) -->
    <div class="table-wrap" tabindex="0" role="region" aria-label={t('tableLabel')}>
      <table>
        <thead><tr><th scope="col">{t('thDe')}</th><th scope="col">{t('thFreq')}</th><th>{t('thMode')}</th><th>{t('thDx')}</th><th>{t('thEntity')}</th><th>{t('thComment')}</th><th>{t('thUtc')}</th></tr></thead>
        <tbody>
          {#each rowsKeyed(filtered) as { spot, key, id } (key)}
            {@const m = modeInfo(spot)}
            <tr class:fresh={freshKeys.has(id)}>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Spotter} onclick={() => query = spot.Spotter}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button>{spot.Spotter}</span></td>
              <td><strong class="frequency" style:--c={colors[bandFor(spot.Frequency)] || '#edf3f9'}>{khz.format(spot.Frequency)}</strong></td>
              <td data-empty={!m.mode || undefined}><span class="mode" data-mode={m.mode} data-inferred={m.inferred && m.mode ? '' : undefined} title={m.inferred && m.mode ? t('modeInferred') : undefined}>{m.mode || '—'}</span></td>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Call} onclick={() => query = spot.Call}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button><strong>{spot.Call}</strong>{#if adxoFor(spot.Call)}{@const entry = adxoFor(spot.Call)!}<button class="info" type="button" title={t('adxoTitle') + ': ' + entry.Entity} aria-label={t('adxoTitle') + ': ' + entry.Entity} onmouseenter={event => openInfo(event, entry)} onfocus={event => openInfo(event, entry)} onmouseleave={hideInfoSoon} onblur={hideInfoSoon} onclick={event => openInfo(event, entry)}><svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3c2.6 2.6 3.9 5.6 3.9 9s-1.3 6.4-3.9 9c-2.6-2.6-3.9-5.6-3.9-9S9.4 5.6 12 3z"/></svg></button>{/if}</span></td>
              <td>
                {#if dxccFlags[spot.DXCC]}
                  <img src={'/flags/' + dxccFlags[spot.DXCC] + '.svg'} alt="" width="26" height="19" loading="lazy" />
                {/if}
                {entities[spot.DXCC]?.[0] ?? (spot.Entity || (spot.DXCC > 0 && spot.DXCC !== 666 ? 'DXCC ' + spot.DXCC : '—'))}
              </td>
              <td class="comment">{spot.Comment}</td>
              <td>{utc.format(new Date(spot.At * 1000))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if available}
    <p>{spots.length ? t('noMatch') : t('noRecent')}</p>
  {/if}
    </section>
  </div>
  <footer class="foot">
    {#if views !== undefined}
      <span class="stat" title={t('visitsLabel')}>
        <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/></svg>
        <strong>{khz.format(views)}</strong> {t('visitsShort')}
      </span>
      <span class="dot" aria-hidden="true"></span>
    {/if}
    {#if engine}
      <span class="stat" title="DXSpider Mojo">DXSpider <strong>{engine}</strong></span>
      <span class="dot" aria-hidden="true"></span>
    {/if}
    <a class="project" href="https://github.com/wifi75/dxregia" target="_blank" rel="noopener noreferrer" title={t('projectNote')}>
      <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 7l-5 5 5 5M16 7l5 5-5 5M14 4l-4 16"/></svg>
      {t('projectLink')}
    </a>
    <span class="dot" aria-hidden="true"></span>
    <span class="author">By IU3WJQ - Tiziano Cassone</span>
  </footer>
  <p class="sponsor">
    <span>{t('sponsorLabel')}</span>
    <a href={'https://aurora.iu3wjq.com/' + (['it', 'es', 'de', 'fr'].includes(lang()) ? lang() : 'en') + '/'} target="_blank" rel="noopener noreferrer"><strong>Aurora</strong> · {t('sponsorText')}</a>
  </p>
  {#if popup}
    <div class="popup" role="dialog" aria-label={t('adxoTitle')} style:left={popup.x + 'px'} style:top={popup.y + 'px'} onclick={event => event.stopPropagation()} onmouseenter={keepInfo} onmouseleave={hideInfoSoon}>
      <h4>{t('adxoTitle')}: {popup.entry.Entity}</h4>
      <p class="dates"><strong>{popup.entry.Call}</strong> · {popup.entry.Dates}{#if popup.entry.QSL} · QSL: {popup.entry.QSL}{/if}</p>
      <p>{popup.entry.Info}</p>
      <p class="src"><a href={popup.entry.Link} target="_blank" rel="noopener noreferrer">{popup.entry.Call}</a> · <a href={adxo?.Source ?? 'https://www.ng3k.com/Misc/adxo.html'} target="_blank" rel="noopener noreferrer">{t('adxoSource')}</a></p>
    </div>
  {/if}
</main>

<style>
  main { position: relative; width: 100%; max-width: 1240px; margin: 0 auto; padding: 14px clamp(12px, 2vw, 32px) 24px; }
  @media (prefers-reduced-motion: reduce) { tbody tr.fresh, tbody tr.fresh:nth-child(even) { animation: none; background: #fff1a0; } }
  .filters { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 10px; padding: 7px 10px; margin-bottom: 12px; min-width: 0; }
  .chip-row, .search-row { display: contents; }
  .chips { display: flex; flex-wrap: wrap; gap: 4px; }
  .chips + .chips { padding-left: 10px; border-left: 1px solid #b6c9d7; }
  .chip { display: inline-flex; align-items: center; gap: .3rem; min-height: 26px; padding: 0 .5rem; border-radius: 999px; font-size: .78rem; --button-face: #fff; --button-edge: #a4bccb; --button-ink: #274760; box-shadow: none; }
  .chip[aria-pressed=true] { --button-face: #175c82; --button-edge: #0c3a56; --button-ink: #fff; }
  .band-color { width: 7px; height: 7px; border-radius: 2px; box-shadow: inset 0 0 0 1px #223c4c30; }
  .search { position: relative; flex: 0 0 15ch; min-width: 0; }
  .search input { width: 100%; min-height: 28px; padding: 3px 26px 3px 9px; font-size: .85rem; }
  .clear { position: absolute; right: 2px; top: 2px; width: 26px; min-height: 26px; padding: 0; border-radius: 6px; font-size: 1.1rem; line-height: 1; --button-face: transparent; --button-edge: transparent; --button-ink: #4c6374; box-shadow: none; }
  .desk { min-width: 0; }
  .log { min-width: 0; padding: 0; overflow: hidden; }
  .log > p { margin: 14px 16px; }
  input { width: min(100%, 350px); min-height: 40px; border: 1px solid #acbfce; border-radius: 6px; background: #fff; color: #173653; padding: 6px 10px; font: inherit; }
  .table-wrap { overflow: auto; }
  table { width: 100%; table-layout: fixed; border-collapse: collapse; text-align: left; white-space: nowrap; }
  td { overflow: hidden; text-overflow: ellipsis; }
  th:nth-child(1) { width: 150px; } th:nth-child(2) { width: 150px; } th:nth-child(3) { width: 100px; }
  th:nth-child(4) { width: 220px; } th:nth-child(5) { width: 210px; } th:nth-child(7) { width: 96px; }
  @media (min-width: 701px) {
    th:nth-child(4), td:nth-child(4) { padding-left: calc(clamp(10px, 1.15vw, 20px) + 20px); }
    th:nth-child(5), td:nth-child(5) { padding-left: calc(clamp(10px, 1.15vw, 20px) + 8px); }
    th:nth-child(6), td:nth-child(6) { padding-left: calc(clamp(10px, 1.15vw, 20px) - 2px); }
  }
  td:nth-child(2), td:nth-child(3) { overflow: visible; text-overflow: clip; }
  th, td { padding: 8px clamp(10px, 1.15vw, 20px); border-bottom: 1px solid #c4d6e3; }
  th { background: #ffffffa3; color: #1d5d8c; border-bottom: 2px solid #2972a0; font-size: .86rem; font-weight: 800; letter-spacing: .06em; text-transform: uppercase; }
  tbody tr:nth-child(odd) { background: #fbfdff; }
  tbody tr:nth-child(even) { background: #d6e8f3; }
  tbody tr.fresh, tbody tr.fresh:nth-child(even) { animation: fresh 14s ease-out both; }
  @keyframes fresh { 0%, 71% { background: #fff1a0; box-shadow: inset 4px 0 #e0a020; } 100% { background: transparent; box-shadow: inset 4px 0 transparent; } }
  .callcell { display: inline-flex; align-items: center; gap: 6px; }
  .lens { display: inline-grid; place-items: center; width: 26px; min-height: 26px; padding: 0; border-radius: 6px; --button-face: transparent; --button-edge: transparent; --button-ink: #3d5a70; box-shadow: none; }
  .lens:hover:not(:disabled) { --button-ink: #175c82; --button-face: #e3eef6; transform: none; filter: none; }
  tbody tr:hover, tbody tr:nth-child(even):hover { background: #c3dbec; }
  .frequency, .mode { display: inline-block; padding: 2px 7px; border: 1.5px solid transparent; border-radius: 5px; text-align: center; }
  .frequency { width: 10ch; color: #142b39; font-weight: 700; background: var(--c); border-color: color-mix(in srgb, var(--c), #000 30%); }
  .frequency, td:last-child { font-variant-numeric: tabular-nums; }
  td { color: #14293a; font-weight: 500; }
  td:first-child, td:last-child, .comment { color: #1f3a4e; }
  img { object-fit: cover; vertical-align: middle; margin-right: 7px; border: 1px solid #15263826; border-radius: 2px; }
  .mode[data-inferred] { border-style: dashed; }
  .mode { min-width: 4.6em; color: #4c6374; font-size: .92rem; font-weight: 700; }
  .mode[data-mode=CW] { --c: #fff0d7; background: var(--c); border-color: color-mix(in srgb, var(--c), #000 28%); color: #7d430e; }
  .mode[data-mode=SSB], .mode[data-mode=USB], .mode[data-mode=LSB] { --c: #d7f0e3; background: var(--c); border-color: color-mix(in srgb, var(--c), #000 28%); color: #1a6046; }
  .mode[data-mode=FT8], .mode[data-mode=FT4], .mode[data-mode=RTTY], .mode[data-mode=PSK31], .mode[data-mode=DIGI] { --c: #eee4ff; background: var(--c); border-color: color-mix(in srgb, var(--c), #000 28%); color: #573894; }
  @media (max-width: 700px) {
    main { padding: 0 10px 20px; }
    .filters { display: grid; grid-template-columns: minmax(0, 1fr); padding: 10px; gap: 8px; margin-bottom: 10px; }
    .chip-row { display: grid; grid-template-columns: minmax(0, 1fr); gap: 8px; }
    .search-row { display: flex; align-items: center; gap: 8px; }
    .chips { flex-wrap: nowrap; overflow-x: auto; padding-bottom: 4px; scrollbar-width: thin; }
    .chip { flex: 0 0 auto; min-height: 38px; font-size: .88rem; }
    .search { flex: 1 1 auto; }
    .search input { min-height: 38px; }
        .table-wrap { overflow: visible; }
    table, tbody { display: block; white-space: normal; }
    thead { display: none; }
    tbody tr { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; grid-template-areas: "freq dx utc" "mode entity entity" "de comment comment"; align-items: center; gap: 3px 10px; padding: 10px 12px; border-bottom: 1px solid #dce5ed; }
    td { padding: 0; border: 0; min-width: 0; }
    td:nth-child(1) { grid-area: de; font-size: .82rem; }
    td:nth-child(1) .lens { display: none; }
    td:nth-child(1)::before { content: 'de '; color: #7a93a6; }
    td:nth-child(2) { grid-area: freq; }
    td:nth-child(3) { grid-area: mode; }
    td:nth-child(3)[data-empty] { display: none; }
    td:nth-child(4) { grid-area: dx; font-size: 1.08rem; }
    td:nth-child(5) { grid-area: entity; font-size: .9rem; }
    td:nth-child(6) { grid-area: comment; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: .85rem; }
    td:nth-child(7) { grid-area: utc; font-size: .85rem; }
    .frequency { width: 10ch; }
  }
  .info { display: inline-grid; place-items: center; width: 24px; min-height: 24px; padding: 0; border-radius: 50%; --button-face: #e3f0fa; --button-edge: #7fb1d9; --button-ink: #175c82; box-shadow: none; flex: 0 0 auto; }
  .popup { position: absolute; z-index: 30; width: min(380px, calc(100% - 16px)); padding: 12px 14px; border: 1px solid #9db8cc; border-radius: 12px; background: #ffffffee; backdrop-filter: blur(10px); box-shadow: 0 14px 34px #143a5940; color: #14293a; font-size: .88rem; line-height: 1.45; white-space: normal; }
  .popup h4 { margin: 0 0 6px; font-size: .95rem; color: #143a60; }
  .popup p { margin: 0 0 6px; color: #1f3a4e; overflow-wrap: anywhere; }
  .popup .dates { font-weight: 600; }
  .popup .src { margin: 8px 0 0; font-size: .82rem; }
  .popup a { color: #175c82; font-weight: 600; }
  .foot { display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 8px 14px; margin: 22px auto 0; padding: 9px 18px; width: fit-content; max-width: 100%; border: 1px solid #b9cbd9; border-radius: 999px; background: #ffffffb3; backdrop-filter: blur(10px); color: #274760; font-size: .88rem; }
  .stat, .project { display: inline-flex; align-items: center; gap: 7px; white-space: nowrap; }
  .stat strong { font-variant-numeric: tabular-nums; color: #143a60; font-weight: 800; }
  .stat svg, .project svg { color: #3a82c0; }
  .project { color: #175c82; font-weight: 650; text-decoration: none; }
  .project:hover { text-decoration: underline; }
  .dot { width: 4px; height: 4px; border-radius: 50%; background: #8aa3b5; }
  .author { color: inherit; opacity: .85; white-space: nowrap; }
  .sponsor { display: flex; flex-wrap: wrap; justify-content: center; gap: 6px 10px; margin: 10px auto 0; max-width: 100%; font-size: .85rem; color: #274760; text-align: center; }
  .sponsor span { text-transform: uppercase; letter-spacing: .06em; font-size: .72rem; opacity: .75; align-self: center; }
  .sponsor a { color: #175c82; text-decoration: none; }
  .sponsor a:hover { text-decoration: underline; }
</style>
