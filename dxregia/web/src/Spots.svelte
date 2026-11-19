<script lang="ts">
  import { dxccFlags } from './dxccFlags'
  import { onMount } from 'svelte'
  import Sponsor from './Sponsor.svelte'
  import { t, locale, lang } from './i18n.svelte'
  import { soundOn, chirp, unlockAudio } from './sound.svelte'

  import SpotResearch from './SpotResearch.svelte'
  import { alertStorageKey, bandFor, bands, localDay, migrateSpotSearch, spotMatcher, corroboratedSpots, groupOf, modeInfo, exportSpotCSV, spotKey, type Spot } from './spotTools'

  type SavedFilter = { name: string; bands: string[]; modes: string[]; noFt8: boolean; query: string; comment?: string; confirmed: boolean; expires?: string }
  const filterKey = 'dxregia-spot-filters'
  let savedFilters = $state<SavedFilter[]>([])
  let selectedFilter = $state('')
  let saveName = $state('')
  let saveExpiry = $state('')
  let showSave = $state(false)
  let saveError = $state(false)
  let { activePage }: { activePage: 'feed' | 'research' | 'alerts' } = $props()
  let researchOpened = $state(false)
  let RadioAlerts = $state<typeof import('./RadioAlerts.svelte').default>()
  let alertsFailed = $state(false)
  $effect(() => {
    if (activePage === 'research') researchOpened = true
    if (activePage === 'alerts') void loadAlerts()
  })

  async function loadAlerts() {
    if (RadioAlerts) return
    try { RadioAlerts = (await import('./RadioAlerts.svelte')).default }
    catch { alertsFailed = true }
  }
  let view = $state<'table' | 'compact' | 'band' | 'compare'>('table')
  let bandA = $state('20 m')
  let bandB = $state('15 m')
  let onlyConfirmed = $state(false)
  let highlight = $state(true)
  let paused = $state(false)
  let frozen = $state<Spot[]>([])
  let pending = $state<Spot[]>([])
  let reviewing = $state(false)
  let resuming = $state(false)
  let recovered = $state(0)
  let pauseGap = $state(false)
  let newEntityKeys = $state(new Set<string>())
  let seenEntities = new Set<number>()
  let entityDay = ''
  let shown = $derived(paused || reviewing ? frozen : spots)
  const confirmedKeys = $derived(corroboratedSpots(shown))
  const matchesSearch = $derived(spotMatcher(query))
  const limitPending = 1000

  function persistFilters(next: SavedFilter[]) {
    try { localStorage.setItem(filterKey, JSON.stringify(next)) } catch { saveError = true; return false }
    saveError = false
    savedFilters = next
    return true
  }
  function commitFilter(event: SubmitEvent) {
    event.preventDefault()
    const name = saveName.trim()
    if (!name || (saveExpiry && saveExpiry < localDay(new Date()))) return
    const next = savedFilters.filter(filter => filter.name !== name).slice(-19)
    next.push({ name, bands: [...bandSel], modes: [...modeSel], noFt8, query, confirmed: onlyConfirmed, expires: saveExpiry })
    if (persistFilters(next)) { selectedFilter = name; showSave = false; saveName = ''; saveExpiry = '' }
  }
  function restoreFilter(name: string) {
    selectedFilter = name
    const filter = savedFilters.find(filter => filter.name === name)
    if (!filter || (filter.expires && filter.expires < localDay(new Date()))) return
    bandSel = [...filter.bands]; modeSel = [...filter.modes]; noFt8 = filter.noFt8
    query = migrateSpotSearch(filter.query, filter.comment ?? ''); onlyConfirmed = filter.confirmed
  }
  async function togglePause() {
    if (!paused) {
      frozen = [...spots]
      pending = []
      pauseGap = false
      recovered = 0
      paused = true
      return
    }
    resuming = true
    const cursor = frozen.find(spot => spot.Cursor)?.Cursor
    let archived: Spot[] = []
    if (cursor) {
      try {
        const response = await fetch('/api/spot-resume?' + new URLSearchParams({ after: cursor }), {
          signal: AbortSignal.timeout(7000),
        })
        if (!response.ok) throw new Error()
        const data: {Spots: Spot[]; Complete: boolean} = await response.json()
        archived = data.Spots
        pauseGap = !data.Complete
      } catch { pauseGap = true }
    }
    const before = new Set(frozen.map(spotKey))
    const combined = new Map([...archived, ...pending].filter(spot => !before.has(spotKey(spot))).map(spot => [spotKey(spot), spot]))
    recovered = combined.size
    if (combined.size) frozen = [...combined.values()].sort((a, b) => a.At - b.At)
    paused = false
    reviewing = true
    resuming = false
  }
  function returnLive() {
    frozen = []
    pending = []
    reviewing = false
    pauseGap = false
  }

  let utc = $derived(new Intl.DateTimeFormat(locale(), { hour: '2-digit', minute: '2-digit', timeZone: 'UTC' }))
  let khz = $derived(new Intl.NumberFormat(locale(), { maximumFractionDigits: 3 }))
  let spots = $state<Spot[]>([])
  let available = $state<boolean>()
  let bandSel = $state<string[]>([])
  let modeSel = $state<string[]>([])
  let noFt8 = $state(false)
  const toggle = (list: string[], value: string) => list.includes(value) ? list.filter(item => item !== value) : [...list, value]
  let query = $state('')
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
  let latestAt = $state(0)
  let spotClock = $state(Date.now())
  const filterToday = $derived(localDay(new Date(spotClock)))
  let arrivalDelay = $state<number>()
  let archiveUpdatedAt = 0
  let views = $state<number>()
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
  let filtered = $derived(shown.filter(spot =>
    (!bandSel.length || bandSel.includes(bandFor(spot.Frequency))) &&
    (!modeSel.length || modeSel.includes(groupOf(modeInfo(spot).mode))) &&
    (!noFt8 || modeInfo(spot).mode !== 'FT8') &&
    (!onlyConfirmed || confirmedKeys.has(spotKey(spot))) && matchesSearch(spot)
  ))

  const comparisonBands = $derived(view === 'compare' ? [...new Set([bandA, bandB])] : view === 'band' ? [bandA] : [])
  const displayed = $derived(comparisonBands.length ? filtered.filter(spot => comparisonBands.includes(bandFor(spot.Frequency))) : filtered)

  async function refresh() {
    if (reading) return
    reading = true
    try {
      const response = await fetch('/api/spots', { signal: AbortSignal.timeout(6000) })
      if (!response.ok) throw new Error()
      const data = await response.json()
      const rows: Spot[] = data.Spots ?? []
      const keys = new Set(rows.map(spotKey))
      const previousKeys = knownKeys
      if (data.Available) {
        const day = new Date().toISOString().slice(0, 10)
        if (entityDay !== day) {
          entityDay = day
          seenEntities = new Set()
          newEntityKeys = new Set()
        }
        const newEntities = new Set(newEntityKeys)
        for (const spot of [...rows].reverse()) {
          if (spot.DXCC <= 0 || spot.DXCC === 666 || seenEntities.has(spot.DXCC) || new Date(spot.At * 1000).toISOString().slice(0, 10) !== day) continue
          seenEntities.add(spot.DXCC)
          if (previousKeys && !previousKeys.has(spotKey(spot))) newEntities.add(spotKey(spot))
        }
        newEntityKeys = new Set([...newEntities].filter(key => keys.has(key)))
      }
      if (paused && data.Available) {
        const existing = new Set([...pending, ...frozen].map(spotKey))
        const arrival = rows.filter(spot => !previousKeys?.has(spotKey(spot)) && !existing.has(spotKey(spot)))
        if ((rows.length >= 100 && !rows.some(spot => previousKeys?.has(spotKey(spot)))) || pending.length + arrival.length > limitPending) pauseGap = true
        pending = [...arrival, ...pending].slice(0, limitPending)
      }
      if (previousKeys) {
        const added = [...keys].filter(key => !previousKeys.has(key))
        if (added.length) {
          const writtenAt: number = data.ArchiveUpdatedAt ?? 0
          const delay = Date.now() - writtenAt
          arrivalDelay = writtenAt > archiveUpdatedAt && delay >= 0 ? Math.floor(delay / 1000) : undefined
          freshKeys = new Set([...freshKeys, ...added])
          // Giallo pieno per 10 secondi, poi sfuma: la scheda resta evidenziata ancora 4 secondi durante la dissolvenza.
          setTimeout(() => { freshKeys = new Set([...freshKeys].filter(key => !added.includes(key))) }, 14500)
          if (!paused && !reviewing && soundOn()) for (let i = 0; i < Math.min(added.length, 4); i++) chirp(i * 0.16)
        }
      }
      if (!data.Available && paused) pauseGap = true
      knownKeys = data.Available ? keys : undefined
      archiveUpdatedAt = data.ArchiveUpdatedAt ?? 0
      latestAt = data.LatestAt ?? rows.reduce((at, spot) => Math.max(at, spot.At), 0)
      if (!data.Available) arrivalDelay = undefined
      spots = rows
      available = data.Available
    } catch {
      available = false
      if (paused) pauseGap = true
      knownKeys = undefined
      arrivalDelay = undefined
    } finally {
      reading = false
    }
  }
  // Su smartphone: chiude la tastiera e porta alla tabella.
  function showResults() {
    (document.activeElement as HTMLElement | null)?.blur()
    document.querySelector('.table-wrap')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
  let menu = $state<'bands' | 'modes' | undefined>()
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
    const syncAlerts = (event: StorageEvent) => { if (event.key === alertStorageKey && event.newValue) void loadAlerts() }
    window.addEventListener('storage', syncAlerts)
    try { if (activePage === 'alerts' || localStorage.getItem(alertStorageKey)) void loadAlerts() }
    catch { if (activePage === 'alerts') void loadAlerts() }
    try {
      const stored: unknown = JSON.parse(localStorage.getItem(filterKey) || '[]')
      if (Array.isArray(stored)) savedFilters = stored.slice(-20).filter(item => item && typeof item.name === 'string' &&
        Array.isArray(item.bands) && Array.isArray(item.modes) && typeof item.query === 'string')
    } catch { /* storage may be unavailable */ }
    const unlock = () => unlockAudio()
    window.addEventListener('pointerdown', unlock, { once: true })
    window.addEventListener('keydown', unlock, { once: true })
    void refresh()
    void loadAdxo()
    const adxoTimer = setInterval(loadAdxo, 30 * 60 * 1000)
    const closePopup = () => { popup = undefined; menu = undefined }
    window.addEventListener('click', closePopup)
    window.addEventListener('keydown', event => { if (event.key === 'Escape') popup = undefined })
    void fetch('/api/visits').then(async response => {
      if (response.ok) views = (await response.json()).Views
    }).catch(() => {})
    const timer = setInterval(refresh, 5000)
    const clockTimer = setInterval(() => { spotClock = Date.now() }, 1000)
    return () => { clearInterval(timer); clearInterval(adxoTimer); clearInterval(clockTimer); window.removeEventListener('storage', syncAlerts); window.removeEventListener('click', closePopup); window.removeEventListener('pointerdown', unlock); window.removeEventListener('keydown', unlock) }
  })
</script>

<main aria-labelledby="spots-title">
  <Sponsor />
  <div hidden={activePage !== 'alerts'}>
    {#if RadioAlerts}<RadioAlerts {spots} {available} />
    {:else}<p role="status">{alertsFailed ? t('archiveDown') : t('loading')}</p>{/if}
  </div>
  {#if researchOpened}<div hidden={activePage !== 'research'}><SpotResearch {spots} /></div>{/if}
  <div hidden={activePage !== 'feed'}>
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
    <div class="pickers">
      <div class="picker">
        <button type="button" class="pick-btn" aria-haspopup="true" aria-expanded={menu === 'bands'} onclick={event => { event.stopPropagation(); menu = menu === 'bands' ? undefined : 'bands' }}>
          {t('bands')}{#if bandSel.length}&nbsp;· {bandSel.length}{/if}<svg viewBox="0 0 20 20" width="14" height="14" aria-hidden="true" fill="currentColor"><path d="M5 7l5 6 5-6z" /></svg>
        </button>
        {#if menu === 'bands'}
          <div class="pick-panel bands-panel" role="group" aria-label={t('bands')} onclick={event => event.stopPropagation()}>
            <button type="button" class="pick-all" onclick={() => bandSel = []}>{t('allShort')}</button>
            {#each bands as [band]}
              <label class="pick-item"><input type="checkbox" checked={bandSel.includes(band)} onchange={() => bandSel = toggle(bandSel, band)} />{band.replace(' ', '')}</label>
            {/each}
          </div>
        {/if}
      </div>
      <div class="picker">
        <button type="button" class="pick-btn" aria-haspopup="true" aria-expanded={menu === 'modes'} onclick={event => { event.stopPropagation(); menu = menu === 'modes' ? undefined : 'modes' }}>
          {t('modes')}{#if modeSel.length + (noFt8 ? 1 : 0)}&nbsp;· {modeSel.length + (noFt8 ? 1 : 0)}{/if}<svg viewBox="0 0 20 20" width="14" height="14" aria-hidden="true" fill="currentColor"><path d="M5 7l5 6 5-6z" /></svg>
        </button>
        {#if menu === 'modes'}
          <div class="pick-panel modes-panel" role="group" aria-label={t('modes')} onclick={event => event.stopPropagation()}>
            {#each [['CW', 'CW'], ['SSB', 'SSB'], ['DIGI', t('digi')]] as [value, label]}
              <label class="pick-item"><input type="checkbox" checked={modeSel.includes(value)} onchange={() => modeSel = toggle(modeSel, value)} />{label}</label>
            {/each}
            <label class="pick-item"><input type="checkbox" checked={noFt8} onchange={() => noFt8 = !noFt8} />{t('noFt8')}</label>
          </div>
        {/if}
      </div>
    </div>
    <div class="search-row">
      <span class="search">
        <input type="search" bind:value={query} aria-label={t('searchLabel')} placeholder={t('searchShort')} title={t('spotSearchSyntax')} maxlength="300" />
        {#if query}<button class="clear" type="button" aria-label={t('clear')} title={t('clear')} onclick={() => query = ''}>×</button>{/if}
      </span>
      <button class="funnel" type="button" aria-label={t('filters')} title={t('filters')} onclick={showResults}>
        <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true" fill="currentColor"><path d="M3 4h18l-7 8.5V19l-4 2v-8.5z" /></svg>
      </button>
    </div>
  </section>
  <div class="feed-controls glass-panel">
    <label>{t('spotSavedFilters')}
      <select value={selectedFilter} onchange={event => restoreFilter(event.currentTarget.value)}>
        <option value="">{t('spotChooseFilter')}</option>
        {#each savedFilters as filter (filter.name)}<option value={filter.name} disabled={Boolean(filter.expires && filter.expires < filterToday)}>{filter.name}{filter.expires && filter.expires < filterToday ? ' · ' + t('radioExpired') : ''}</option>{/each}
      </select>
    </label>
    <button type="button" aria-expanded={showSave} onclick={() => { showSave = !showSave; saveName = selectedFilter; saveExpiry = savedFilters.find(item => item.name === selectedFilter)?.expires ?? ''; saveError = false }}>{t('spotSaveFilter')}</button>
    <label>{t('spotView')}
      <select bind:value={view}>
        <option value="table">{t('spotTable')}</option><option value="compact">{t('spotCompact')}</option>
        <option value="band">{t('spotByBand')}</option><option value="compare">{t('spotBandCompare')}</option>
      </select>
    </label>
    {#if comparisonBands.length}
      <label>{t('bands')} A<select bind:value={bandA}>{#each bands as [band]}<option>{band}</option>{/each}</select></label>
      {#if view === 'compare'}<label>{t('bands')} B<select bind:value={bandB}>{#each bands as [band]}<option>{band}</option>{/each}</select></label>{/if}
    {/if}
    {#if reviewing}
      <button type="button" onclick={returnLive}>{t('spotBackLive')}</button>
    {:else}
      <button type="button" onclick={togglePause} disabled={resuming}>{resuming ? t('loading') : paused ? t('spotResume') : t('spotPause')}</button>
    {/if}
    <button type="button" onclick={() => exportSpotCSV(displayed)} disabled={!displayed.length}>{t('spotExportCSV')}</button>
    {#if showSave}
      <form class="save-filter" onsubmit={commitFilter}>
        <label>{t('spotFilterName')}<input bind:value={saveName} maxlength="40" required /></label>
        <label>{t('radioExpiry')}<input type="date" bind:value={saveExpiry} min={filterToday} /></label>
        <button type="submit">{t('spotSaveFilter')}</button>
        {#if savedFilters.some(filter => filter.name === saveName.trim())}
          <button type="button" onclick={() => { if (persistFilters(savedFilters.filter(filter => filter.name !== saveName.trim()))) { selectedFilter = ''; saveName = ''; showSave = false } }}>{t('spotRemoveFilter')}</button>
        {/if}
      </form>
    {/if}
    {#if saveError}<p role="status">{t('spotStorageUnavailable')}</p>{/if}
  </div>
  <div class="feed-controls glass-panel">
    <label class="check" title={t('spotConfirmedHint')}><input type="checkbox" bind:checked={onlyConfirmed} /> {t('spotOnlyConfirmed')}</label>
    <label class="check"><input type="checkbox" bind:checked={highlight} /> {t('spotMarkNew')}</label>
  </div>
  {#if paused || reviewing}
    <p class="spot-timing" role="status">{paused ? t('spotWaiting', { n: pending.length }) : t('spotReviewing', { n: recovered })}</p>
    {#if pauseGap}<p class="spot-timing" role="status">{t('spotGap')}</p>{/if}
  {/if}
  {#if available && latestAt > 0}
    <p class="spot-timing">
      {t('lastSpotAge')} · {Math.max(0, Math.floor(spotClock / 1000 - latestAt))} s
      {#if arrivalDelay !== undefined}<span>{t('spotArrivalDelay')} · {arrivalDelay} s</span>{/if}
    </p>
  {/if}
  <div class="desk glass-enter">
    <section class="log glass-panel" aria-label="Registro degli spot">
  {#if available === false}
    <p role="status">{t('archiveDown')}</p>
  {:else if available === undefined}
    <p role="status">{t('loading')}</p>
  {/if}
  {#snippet spotTable(rows: Spot[], label?: string)}
    {#if label}<h2 class="band-heading">{label} · {rows.length}</h2>{/if}
    {#if rows.length}
    <!-- svelte-ignore a11y_no_noninteractive_tabindex (scrollable table needs keyboard focus) -->
    <div class="table-wrap" tabindex="0" role="region" aria-label={t('tableLabel')}>
      <table class:compact={view === 'compact'}>
        <thead><tr><th scope="col">{t('thDe')}</th><th scope="col">{t('thFreq')}</th><th>{t('thMode')}</th><th>{t('thDx')}</th><th>{t('thEntity')}</th><th>{t('thComment')}</th><th>{t('thUtc')}</th></tr></thead>
        <tbody>
          {#each rowsKeyed(rows) as { spot, key, id } (key)}
            {@const m = modeInfo(spot)}
            <tr class:fresh={freshKeys.has(id)} class:new-entity={highlight && newEntityKeys.has(id)}>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Spotter} onclick={() => query = spot.Spotter}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button>{spot.Spotter}</span></td>
              <td><strong class="frequency" style:--c={colors[bandFor(spot.Frequency)] || '#edf3f9'}>{khz.format(spot.Frequency)}</strong></td>
              <td data-empty={!m.mode || undefined}><span class="mode" data-mode={m.mode} data-inferred={m.inferred && m.mode ? '' : undefined} title={m.inferred && m.mode ? t('modeInferred') : undefined}>{m.mode || '—'}</span></td>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Call} onclick={() => query = spot.Call}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button><strong>{spot.Call}</strong>{#if confirmedKeys.has(id)}<span class="confirmed" role="img" aria-label={t('spotConfirmedHint')} title={t('spotConfirmedHint')}>✓</span>{/if}{#if adxoFor(spot.Call)}{@const entry = adxoFor(spot.Call)!}<button class="info" type="button" title={t('adxoTitle') + ': ' + entry.Entity} aria-label={t('adxoTitle') + ': ' + entry.Entity} onmouseenter={event => openInfo(event, entry)} onfocus={event => openInfo(event, entry)} onmouseleave={hideInfoSoon} onblur={hideInfoSoon} onclick={event => openInfo(event, entry)}><svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3c2.6 2.6 3.9 5.6 3.9 9s-1.3 6.4-3.9 9c-2.6-2.6-3.9-5.6-3.9-9S9.4 5.6 12 3z"/></svg></button>{/if}</span></td>
              <td>
                {#if dxccFlags[spot.DXCC]}
                  <img src={'/flags/' + dxccFlags[spot.DXCC] + '.svg'} alt="" width="26" height="19" loading="lazy" />
                {/if}
                {(lang() === 'it' ? entities[spot.DXCC]?.[0] : undefined) ?? (spot.Entity || (spot.DXCC > 0 && spot.DXCC !== 666 ? 'DXCC ' + spot.DXCC : '—'))}
                {#if highlight && newEntityKeys.has(id)}<span class="spot-tag">{t('spotNewEntity')}</span>{/if}
              </td>
              <td class="comment" title={spot.Comment}>{spot.Comment}</td>
              <td>{utc.format(new Date(spot.At * 1000))}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
    {:else if available}<p class="empty-spots">{shown.length ? t('noMatch') : t('noRecent')}</p>{/if}
  {/snippet}
  {#if comparisonBands.length}
    <div class="comparison" class:paired={comparisonBands.length > 1}>
      {#each comparisonBands as band (band)}
        <div>{@render spotTable(displayed.filter(spot => bandFor(spot.Frequency) === band), band)}</div>
      {/each}
    </div>
  {:else}
    {@render spotTable(displayed)}
  {/if}
    </section>
  </div>
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
      <span class="stat" title="DXSpider">DXSpider <strong>{engine}</strong></span>
      <span class="dot" aria-hidden="true"></span>
    {/if}
    <a class="project" href="https://github.com/wifi75/dxregia" target="_blank" rel="noopener noreferrer" title={t('projectNote')}>
      <svg viewBox="0 0 24 24" width="17" height="17" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 7l-5 5 5 5M16 7l5 5-5 5M14 4l-4 16"/></svg>
      {t('projectLink')}
    </a>
    <span class="dot" aria-hidden="true"></span>
    <span class="author">By IU3WJQ - Tiziano Cassone</span>
  </footer>
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
  .feed-controls { display: flex; flex-wrap: wrap; align-items: end; gap: 8px; padding: 8px 10px; margin-bottom: 8px; }
  .feed-controls label { display: grid; gap: 3px; font-size: .8rem; font-weight: 650; min-width: 0; }
  .feed-controls select, .feed-controls input:not([type=checkbox]), .feed-controls button { min-height: 34px; min-width: 0; }
  .feed-controls input:not([type=checkbox]) { width: 100%; }
  .feed-controls .check { display: flex; align-items: center; gap: 5px; min-height: 34px; }
  .check input { width: auto; min-height: 0; }
  .save-filter { display: flex; flex-wrap: wrap; gap: 8px; align-items: end; flex-basis: 100%; }
  .band-heading { font-size: 1rem; margin: 8px 12px; }
  .comparison { display: grid; }
  .comparison.paired { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .comparison > div { min-width: 0; }
  .comparison.paired table { table-layout: auto; }
  .comparison.paired th, .comparison.paired td { width: auto; padding-inline: 5px; }
  .empty-spots { margin: 10px; }
  table.compact td { padding-block: 2px; }
  table.compact .lens, table.compact .mode, table.compact .frequency { min-height: 21px; }
  .new-entity td:nth-child(5) { box-shadow: inset 3px 0 #c28522; font-weight: 700; }
  .spot-tag { margin-inline-start: 4px; color: #55360b; background: #ffe5a0; font-size: .72rem; border-radius: 5px; padding: 2px 4px; }
  .confirmed { color: #197849; font-weight: 800; }
  @media (max-width: 900px) { .comparison.paired { grid-template-columns: minmax(0, 1fr); } }
  .spot-timing { display: flex; flex-wrap: wrap; gap: 4px 16px; margin: 6px 2px; color: #385569; font-size: .82rem; font-variant-numeric: tabular-nums; }
  main { position: relative; width: 100%; max-width: 1240px; margin: 0 auto; padding: 14px clamp(12px, 2vw, 32px) 24px; }
  @media (prefers-reduced-motion: reduce) { tbody tr.fresh, tbody tr.fresh:nth-child(even) { animation: none; background: #fff1a0; } }
  .filters { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 10px; padding: 7px 10px; margin-bottom: 12px; min-width: 0; }
  .chip-row, .search-row { display: contents; }
  .chips { display: flex; flex-wrap: wrap; gap: 4px; }
  .chips + .chips { padding-left: 10px; border-left: 1px solid #b6c9d7; }
  .chip { display: inline-flex; align-items: center; gap: .3rem; min-height: 26px; padding: 0 .5rem; border-radius: 999px; font-size: .78rem; --button-face: #fff; --button-edge: #a4bccb; --button-ink: #274760; box-shadow: none; }
  .chip[aria-pressed=true] { --button-face: #175c82; --button-edge: #0c3a56; --button-ink: #fff; }
  .band-color { width: 7px; height: 7px; border-radius: 2px; box-shadow: inset 0 0 0 1px #223c4c30; }
  .search { position: relative; flex: 1 1 17ch; min-width: 0; }
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
    th:nth-child(4) { width: 200px; }
    th:nth-child(5) { width: 210px; }
    th:nth-child(4), td:nth-child(4) { padding-left: calc(clamp(10px, 1.15vw, 20px) + 20px); }
    th:nth-child(5), td:nth-child(5) { padding-left: calc(clamp(10px, 1.15vw, 20px) + 3px); }
    th:nth-child(6), td:nth-child(6) { padding-left: calc(clamp(10px, 1.15vw, 20px) - 2px); }
    td:nth-child(5), td.comment { font-size: .94em; }
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
  .pickers { display: none; }
  @media (max-width: 700px) {
    .chip-row { display: none; }
    .filters { position: relative; z-index: 20; overflow: visible; }
    .pickers { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; position: relative; }
    .picker { position: relative; min-width: 0; }
    .pick-btn { display: flex; align-items: center; justify-content: space-between; gap: 8px; width: 100%; min-height: 44px; padding: 0 16px; border-radius: 999px; font-size: 1rem; font-weight: 650; }
    .pick-panel { position: absolute; z-index: 25; top: calc(100% + 6px); left: 0; display: grid; gap: 2px; min-width: 100%; max-height: 60vh; overflow: auto; padding: 8px; border: 1px solid #9db8cc; border-radius: 14px; background: #fff; box-shadow: 0 14px 34px #143a5940; }
    .bands-panel { grid-template-columns: 1fr 1fr; width: calc(200% + 10px); }
    .picker:nth-child(2) .pick-panel { left: auto; right: 0; }
    .pick-all { grid-column: 1 / -1; min-height: 38px; font-size: 1rem; }
    .pick-item { display: flex; align-items: center; gap: 10px; min-height: 42px; padding: 0 10px; border-radius: 8px; color: #14293a; font-size: 1rem; }
    .pick-item:hover { background: #e3eef6; }
    .pick-item input { width: 20px; min-height: 20px; margin: 0; accent-color: #1f6fa8; }
  }
  .funnel { display: none; }
  @media (max-width: 700px) {
    main { padding: 0 8px 20px; }
    .filters { display: grid; grid-template-columns: minmax(0, 1fr); padding: 10px; gap: 10px; margin-bottom: 10px; }
    .chips { flex-wrap: wrap; gap: 8px; overflow: visible; }
    .chips + .chips { padding-left: 0; border-left: 0; }
    .chip { flex: 0 0 auto; min-height: 40px; padding: 0 16px; font-size: 1rem; }
    .band-color { display: none; }
    .search-row { display: flex; align-items: center; gap: 10px; }
    .search { flex: 1 1 auto; }
    .search input { width: 100%; min-height: 44px; border-radius: 999px; padding: 6px 16px; font-size: 1rem; }
    .funnel { display: grid; flex: none; place-items: center; width: 44px; min-height: 44px; padding: 0; border: 0; border-radius: 50%; background: #1f86ff; color: #fff; box-shadow: none; }
    .table-wrap { overflow-x: auto; -webkit-overflow-scrolling: touch; }
    table { table-layout: auto; width: 100%; min-width: 0; }
    th { width: auto !important; position: sticky; top: 0; padding: 7px 3px; font-size: .6rem; letter-spacing: 0; }
    td { padding: 3px 3px; font-size: .8rem; }
    th:first-child, td:first-child { padding-left: 8px; }
    td:nth-child(1) .lens { display: none; }
    td:nth-child(2), td:nth-child(3) { overflow: visible; text-overflow: clip; }
    th:nth-child(6), td:nth-child(6) { display: none; }
    td:nth-child(5) { font-size: 0; padding-left: 2px; padding-right: 2px; }
    td:nth-child(5) img { width: 28px; height: 20px; margin: 0; }
    td:nth-child(7) { font-size: .74rem; padding-right: 8px; }
    .callcell { gap: 2px; }
    .frequency { width: auto; min-width: 0; padding: 0 4px; font-size: .78rem; }
    .mode { min-width: 0; padding: 0 4px; font-size: .72rem; }
    .lens { width: 18px; min-height: 18px; }
    .lens svg { width: 14px; height: 14px; }
    .info { width: 18px; min-height: 18px; }
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
</style>
