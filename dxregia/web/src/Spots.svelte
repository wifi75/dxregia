<script lang="ts">
  import { dxccFlags } from './dxccFlags'
  import { onMount } from 'svelte'
  import { t, locale } from './i18n.svelte'

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
  let selectedBand = $state('')
  let mode = $state('')
  let noFt8 = $state(false)
  let query = $state('')
  const spotKey = (spot: Spot) => spot.At + '|' + spot.Call + '|' + spot.Frequency
  const soundPref = () => { try { return localStorage.getItem('dxregia-sound') !== 'off' } catch { return true } }
  let soundOn = $state(soundPref())
  let freshKeys = $state<Set<string>>(new Set())
  let knownKeys: Set<string> | undefined
  let audio: AudioContext | undefined

  // I browser permettono l'audio solo dopo un gesto dell'utente: il primo tocco o tasto lo sblocca.
  function unlockAudio() {
    try {
      audio ??= new AudioContext()
      if (audio.state === 'suspended') void audio.resume()
    } catch { /* senza audio il resto funziona uguale */ }
  }

  // Un breve «bip» morbido: due note ravvicinate, volume basso.
  function chirp(at = 0) {
    if (!audio || audio.state !== 'running') return
    const start = audio.currentTime + at
    const gain = audio.createGain()
    gain.connect(audio.destination)
    gain.gain.setValueAtTime(0.0001, start)
    gain.gain.exponentialRampToValueAtTime(0.09, start + 0.015)
    gain.gain.exponentialRampToValueAtTime(0.0001, start + 0.2)
    const tone = audio.createOscillator()
    tone.type = 'sine'
    tone.frequency.setValueAtTime(880, start)
    tone.frequency.exponentialRampToValueAtTime(1320, start + 0.09)
    tone.connect(gain)
    tone.start(start)
    tone.stop(start + 0.22)
  }

  function toggleSound() {
    soundOn = !soundOn
    try { localStorage.setItem('dxregia-sound', soundOn ? 'on' : 'off') } catch { /* solo una comodità */ }
    if (soundOn) { unlockAudio(); chirp() }
  }
  let reading = false
  let views = $state<number>()
  const bandFor = (frequency: number) => bands.find(([, low, high]) => frequency >= low && frequency <= high)?.[0] ?? ''
  const modeFor = (comment: string) => comment.match(/\b(FT8|FT4|SSB|CW|RTTY|PSK31|USB|LSB)\b/i)?.[1].toUpperCase() ?? ''
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
    (!selectedBand || bandFor(spot.Frequency) === selectedBand) &&
    (!mode || groupOf(modeFor(spot.Comment)) === mode) &&
    (!noFt8 || modeFor(spot.Comment) !== 'FT8') &&
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
          if (soundOn) for (let i = 0; i < Math.min(added.length, 4); i++) chirp(i * 0.16)
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
  onMount(() => {
    const unlock = () => unlockAudio()
    window.addEventListener('pointerdown', unlock, { once: true })
    window.addEventListener('keydown', unlock, { once: true })
    void refresh()
    void fetch('/api/visits').then(async response => {
      if (response.ok) views = (await response.json()).Views
    }).catch(() => {})
    const timer = setInterval(refresh, 5000)
    return () => { clearInterval(timer); window.removeEventListener('pointerdown', unlock); window.removeEventListener('keydown', unlock) }
  })
</script>

<main aria-labelledby="spots-title">
  <section class="filters glass-panel glass-enter" aria-label={t('filters')}>
    <div class="chip-row">
      <div class="chips" role="group" aria-label={t('bands')}>
        <button class="chip" type="button" aria-pressed={selectedBand === ''} onclick={() => selectedBand = ''}>{t('allShort')}</button>
        {#each bands as [band]}
          <button class="chip" type="button" aria-pressed={selectedBand === band} onclick={() => selectedBand = selectedBand === band ? '' : band}>
            <span class="band-color" style:background={colors[band]}></span>{band.replace(' ', '')}
          </button>
        {/each}
      </div>
      <div class="chips" role="group" aria-label={t('modes')}>
        {#each [['CW', 'CW'], ['SSB', 'SSB'], ['DIGI', t('digi')]] as [value, label]}
          <button class="chip" type="button" aria-pressed={mode === value} onclick={() => mode = mode === value ? '' : value}>{label}</button>
        {/each}
        <button class="chip" type="button" aria-pressed={noFt8} onclick={() => noFt8 = !noFt8}>{t('noFt8')}</button>
      </div>
    </div>
    <div class="search-row">
      <span class="search">
        <input type="search" bind:value={query} aria-label={t('searchLabel')} placeholder={t('searchPlaceholder')} />
        {#if query}<button class="clear" type="button" aria-label={t('clear')} title={t('clear')} onclick={() => query = ''}>×</button>{/if}
      </span>
      <span class="count">{t(filtered.length === 1 ? 'shownOne' : 'shownMany', { n: filtered.length })}</span>
      <button class="sound" type="button" aria-pressed={soundOn} title={soundOn ? t('soundOn') : t('soundOff')} aria-label={soundOn ? t('soundOn') : t('soundOff')} onclick={toggleSound}>
        <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 9.5v5h3.5L12 18.5v-13L7.5 9.5z" fill="currentColor" stroke="none"/>
          {#if soundOn}<path d="M15.5 9a4.2 4.2 0 0 1 0 6"/><path d="M18 6.5a8 8 0 0 1 0 11"/>{:else}<path d="M16 9.5l5 5M21 9.5l-5 5"/>{/if}
        </svg>
      </button>
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
        <thead><tr><th scope="col">{t('thDe')}</th><th scope="col">{t('thFreq')} <small>kHz</small></th><th>{t('thMode')}</th><th>{t('thDx')}</th><th>{t('thEntity')}</th><th>{t('thComment')}</th><th>{t('thUtc')}</th></tr></thead>
        <tbody>
          {#each rowsKeyed(filtered) as { spot, key, id } (key)}
            <tr class:fresh={freshKeys.has(id)}>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Spotter} onclick={() => query = spot.Spotter}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button>{spot.Spotter}</span></td>
              <td><strong class="frequency" style:background={colors[bandFor(spot.Frequency)] || '#edf3f9'}>{khz.format(spot.Frequency)}</strong></td>
              <td data-empty={!modeFor(spot.Comment) || undefined}><span class="mode" data-mode={modeFor(spot.Comment)}>{modeFor(spot.Comment) || '—'}</span></td>
              <td><span class="callcell"><button class="lens" type="button" title={t('lens')} aria-label={t('lens') + ': ' + spot.Call} onclick={() => query = spot.Call}><svg viewBox="0 0 20 20" width="15" height="15" aria-hidden="true"><circle cx="8.5" cy="8.5" r="5.5" fill="none" stroke="currentColor" stroke-width="2"/><path d="M13 13l5 5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg></button><strong>{spot.Call}</strong></span></td>
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
  {#if views !== undefined}<p class="visits">{t('visits', { n: khz.format(views) })}</p>{/if}
    </section>
  </div>
</main>

<style>
  main { width: 100%; max-width: 1240px; margin: 0 auto; padding: 14px clamp(12px, 2vw, 32px) 24px; }
  h1 { margin: 0; font-size: clamp(1.65rem, 3vw, 2.25rem); color: #102e42; letter-spacing: -.04em; }
  @media (prefers-reduced-motion: reduce) { tbody tr.fresh, tbody tr.fresh:nth-child(even) { animation: none; background: #fff1a0; } }
  .filters { display: grid; grid-template-columns: minmax(0, 1fr); gap: 10px; padding: 12px 14px; margin-bottom: 14px; min-width: 0; }
  .filters > * { min-width: 0; }
  .chip-row { display: flex; flex-wrap: wrap; gap: 8px 22px; align-items: center; }
  .chips { display: flex; flex-wrap: wrap; gap: 7px; }
  .chip { display: inline-flex; align-items: center; gap: .4rem; min-height: 34px; padding: .25rem .8rem; border-radius: 999px; font-size: .88rem; --button-face: #fff; --button-edge: #a4bccb; --button-ink: #274760; box-shadow: none; }
  .chip[aria-pressed=true] { --button-face: #175c82; --button-edge: #0c3a56; --button-ink: #fff; }
  .band-color { width: 9px; height: 9px; border-radius: 3px; box-shadow: inset 0 0 0 1px #223c4c30; }
  .search-row { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; min-width: 0; }
  .count { margin-left: auto; }
  .search { position: relative; flex: 1 1 280px; max-width: 460px; min-width: 0; }
  .search input { width: 100%; padding-right: 36px; }
  .clear { position: absolute; right: 4px; top: 4px; width: 32px; min-height: 32px; padding: 0; border-radius: 6px; font-size: 1.2rem; line-height: 1; --button-face: transparent; --button-edge: transparent; --button-ink: #4c6374; box-shadow: none; }
  .count { color: #496174; font-size: .9rem; }
  .desk { min-width: 0; }
  .log { min-width: 0; padding: 0; overflow: hidden; }
  .log > p { margin: 14px 16px; }
  .toolbar { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: .65rem; padding: 12px 16px; border-bottom: 1px solid #ccdbe7; background: #ffffff80; color: #496174; }
  input { width: min(100%, 350px); min-height: 40px; border: 1px solid #acbfce; border-radius: 6px; background: #fff; color: #173653; padding: 6px 10px; font: inherit; }
  .table-wrap { overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; white-space: nowrap; }
  th, td { padding: 8px clamp(10px, 1.15vw, 20px); border-bottom: 1px solid #c4d6e3; }
  th { background: #ffffffa3; color: #1d5d8c; border-bottom: 2px solid #2972a0; font-size: .86rem; font-weight: 800; letter-spacing: .06em; text-transform: uppercase; }
  th small { color: #536b7d; text-transform: none; }
  tbody tr:nth-child(odd) { background: #fbfdff; }
  tbody tr:nth-child(even) { background: #d6e8f3; }
  tbody tr.fresh, tbody tr.fresh:nth-child(even) { animation: fresh 14s ease-out both; }
  @keyframes fresh { 0%, 71% { background: #fff1a0; box-shadow: inset 4px 0 #e0a020; } 100% { background: transparent; box-shadow: inset 4px 0 transparent; } }
  .sound { display: inline-grid; place-items: center; width: 40px; min-height: 40px; padding: 0; border-radius: 50%; --button-face: #fff; --button-edge: #a4bccb; --button-ink: #214e80; box-shadow: none; }
  .sound[aria-pressed=false] { --button-ink: #7d8fa0; }
  .callcell { display: inline-flex; align-items: center; gap: 6px; }
  .lens { display: inline-grid; place-items: center; width: 26px; min-height: 26px; padding: 0; border-radius: 6px; --button-face: transparent; --button-edge: transparent; --button-ink: #6d8798; box-shadow: none; }
  .lens:hover:not(:disabled) { --button-ink: #175c82; --button-face: #e3eef6; transform: none; filter: none; }
  tbody tr:hover, tbody tr:nth-child(even):hover { background: #c3dbec; }
  .frequency, .mode { display: inline-block; padding: 3px 8px; border-radius: 5px; text-align: center; }
  .frequency { width: 10.5ch; color: #142b39; font-weight: 700; }
  .frequency, td:last-child { font-variant-numeric: tabular-nums; }
  td:first-child, td:last-child, .comment { color: #465f72; }
  .comment { width: 100%; }
  img { object-fit: cover; vertical-align: middle; margin-right: 7px; border: 1px solid #15263826; border-radius: 2px; }
  .mode { min-width: 4.6em; color: #4c6374; font-size: .92rem; font-weight: 700; }
  .mode[data-mode=CW] { background: #fff0d7; color: #7d430e; }
  .mode[data-mode=SSB], .mode[data-mode=USB], .mode[data-mode=LSB] { background: #d7f0e3; color: #1a6046; }
  .mode[data-mode=FT8], .mode[data-mode=FT4], .mode[data-mode=RTTY], .mode[data-mode=PSK31] { background: #eee4ff; color: #573894; }
  .visits { margin-block: 14px 0; color: #526c7f; font-size: .85rem; }
  @media (max-width: 700px) {
    main { padding: 0 10px 20px; }
    .filters { padding: 10px; gap: 8px; margin-bottom: 10px; }
    .chip-row { flex-direction: column; align-items: stretch; gap: 8px; }
    .chips { flex-wrap: nowrap; overflow-x: auto; padding-bottom: 4px; scrollbar-width: thin; }
    .chip { flex: 0 0 auto; min-height: 40px; }
    .search { max-width: none; flex: 1 1 100%; }
    .count { margin-left: 0; }
    .sound { margin-left: auto; }
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
</style>
