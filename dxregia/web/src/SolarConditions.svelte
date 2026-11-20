<script lang="ts">
  import { onMount } from 'svelte'
  import { locale, t } from './i18n.svelte'

  type Solar = {
    UpdatedAt: string; SFI: string; A: string; K: string; Sunspots: string
    AuroraLatitude: string; SolarWind: string; MagneticField: string; GeomagneticField: string
    SignalNoise: string; FOF2: string
    HF: { Band: string; Day: string; Night: string }[]
    VHF: { Name: string; Location: string; Condition: string }[]
  }
  const mapURL = 'https://prop.kc2g.com/renders/current/mufd-normal-now.svg'
  let report = $state<Solar>()
  let unavailable = $state(false)
  let mapVersion = $state(Math.floor(Date.now() / 900000))
  let mapFailed = $state(false)
  const date = $derived(new Intl.DateTimeFormat(locale(), { timeZone: 'UTC', dateStyle: 'medium', timeStyle: 'short' }))
  const metrics = [
    ['SFI', 'propSFI', '', 'solar'], ['A', 'propIndexA', '', 'geo'], ['K', 'propIndexK', '', 'geo'],
    ['Sunspots', 'solarSunspots', '', 'solar'], ['AuroraLatitude', 'solarAuroraLatitude', '°', 'vhf'],
    ['SolarWind', 'solarWind', ' km/s', 'solar'], ['MagneticField', 'solarMagnetic', ' nT', 'geo'],
    ['GeomagneticField', 'solarGeomagnetic', '', 'geo'], ['SignalNoise', 'solarNoise', '', 'hf'], ['FOF2', 'solarFoF2', ' MHz', 'hf'],
  ] as const
  const conditions = {
    good: 'solarGood', fair: 'solarFair', poor: 'solarPoor', quiet: 'solarQuiet',
    'band closed': 'solarClosed', 'band open': 'solarOpen',
  } as const
  const places = {
    northern_hemi: 'solarNorth', europe: 'propEurope', north_america: 'propNorthAmerica',
    europe_6m: 'solarEurope6m', europe_4m: 'solarEurope4m',
  } as const
  const keyOf = (value: string) => (value || '').trim().toLowerCase()
  function quality(value: string) {
    const normalized = keyOf(value)
    if (['', 'none', 'norpt', 'no report'].includes(normalized)) return '—'
    return Object.hasOwn(conditions, normalized) ? t(conditions[normalized as keyof typeof conditions]) : value
  }
  function area(value: string) {
    return Object.hasOwn(places, value) ? t(places[value as keyof typeof places]) : value.replaceAll('_', ' ')
  }
  function tone(value: string) {
    switch (keyOf(value)) {
      case 'good': case 'band open': return 'good'
      case 'fair': return 'fair'
      case 'poor': case 'band closed': return 'poor'
      default: return 'neutral'
    }
  }

  onMount(() => {
    const lifecycle = new AbortController()
    async function refresh() {
      if (document.hidden) return
      try {
        const response = await fetch('/api/solar-conditions', {
          signal: AbortSignal.any([lifecycle.signal, AbortSignal.timeout(6500)]),
        })
        if (!response.ok) throw new Error()
        report = await response.json() as Solar
        unavailable = false
      } catch {
        if (!lifecycle.signal.aborted) unavailable = true
      }
      const version = Math.floor(Date.now() / 900000)
      if (version !== mapVersion) {
        mapFailed = false
        mapVersion = version
      }
    }
    void refresh()
    const timer = setInterval(refresh, 5 * 60 * 1000)
    document.addEventListener('visibilitychange', refresh)
    return () => {
      lifecycle.abort()
      clearInterval(timer)
      document.removeEventListener('visibilitychange', refresh)
    }
  })
</script>

<section class="glass-panel solar-summary" aria-labelledby="solar-summary-title">
  <div class="summary-heading">
    <h2 id="solar-summary-title">{t('solarOverview')}</h2>
    {#if report}<time datetime={report.UpdatedAt}>{date.format(new Date(report.UpdatedAt))} UTC</time>{/if}
  </div>
  {#if !report || unavailable}
    <p role="status">{unavailable ? t('solarUnavailable') : t('loading')}</p>
  {/if}
  {#if report}
    <div class="overview-grid">
      <dl class="metrics">
        {#each metrics as [field, title, unit, group]}
          <div data-group={group}>
            <dt>{t(title)}</dt>
            <dd class:text-reading={field === 'GeomagneticField'}>{quality(report[field])}{#if report[field] && unit}<small>{unit}</small>{/if}</dd>
          </div>
        {/each}
      </dl>

      <div class="conditions">
        <table>
          <caption>{t('solarHF')}</caption>
          <thead><tr><th scope="col">{t('bands')}</th><th scope="col">{t('solarDay')}</th><th scope="col">{t('solarNight')}</th></tr></thead>
          <tbody>
            {#each report.HF as row (row.Band)}
              <tr><th scope="row">{row.Band.replace('-', '–')}</th>
                <td><span class="reading" data-tone={tone(row.Day)}>{quality(row.Day)}</span></td>
                <td><span class="reading" data-tone={tone(row.Night)}>{quality(row.Night)}</span></td></tr>
            {:else}
              <tr><td colspan="3">{t('bulletinsEmpty')}</td></tr>
            {/each}
          </tbody>
        </table>
        <table>
          <caption>{t('solarVHF')}</caption>
          <thead><tr><th scope="col">{t('solarPhenomenon')}</th><th scope="col">{t('solarRegion')}</th><th scope="col">{t('solarCondition')}</th></tr></thead>
          <tbody>
            {#each report.VHF as row}
              <tr><th scope="row">{keyOf(row.Name) === 'vhf-aurora' ? t('solarVHFAurora') : row.Name}</th>
                <td>{area(row.Location)}</td><td><span class="reading" data-tone={tone(row.Condition)}>{quality(row.Condition)}</span></td></tr>
            {:else}
              <tr><td colspan="3">{t('bulletinsEmpty')}</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
  <footer>{t('solarSource')} <a href="https://www.hamqsl.com/solar.html" target="_blank" rel="noopener noreferrer">HamQSL · N0NBH</a>
    <span> · NOAA SWPC, Space Weather Canada, DXrobot, Make More Miles on VHF, G7IZU</span>
  </footer>
</section>

<section class="glass-panel muf" aria-labelledby="muf-title">
  <h2 id="muf-title">{t('solarMUF')}</h2>
  <p>{t('solarMUFHelp')}</p>
  {#if !mapFailed}
    <a class="map-link" href="https://prop.kc2g.com/" target="_blank" rel="noopener noreferrer">
      <img src={mapURL + '?v=' + mapVersion} alt={t('solarMUF')} width="1526" height="905"
        referrerpolicy="no-referrer" loading="lazy" onerror={() => mapFailed = true} />
    </a>
  {:else}
    <p role="status">{t('solarMapUnavailable')}</p>
  {/if}
  <footer>{t('solarSource')} <a href="https://prop.kc2g.com/" target="_blank" rel="noopener noreferrer">KC2G · prop.kc2g.com</a>
    <span> · <a href="https://prop.kc2g.com/acknowledgments/" target="_blank" rel="noopener noreferrer">GIRO, INGV, WWROF</a></span>
  </footer>
</section>

<style>
  section { padding: clamp(12px, 2vw, 22px); margin-bottom: 14px; }
  h2 { margin: 0; color: #153b57; font-size: 1.3rem; }
  p { margin: 6px 0 12px; color: #456074; }
  .summary-heading { display: flex; justify-content: space-between; align-items: baseline; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
  time, footer { color: #456074; font-size: .82rem; }
  footer { border-top: 1px solid #c4d8e3; margin-top: 12px; padding-top: 8px; }
  a { color: #175c82; font-weight: 650; }
  .overview-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 16px; align-items: start; }
  .metrics { display: grid; grid-template-columns: repeat(auto-fit,minmax(min(100%,118px),1fr)); gap: 8px; margin: 0; }
  .metrics > div { display: flex; flex-direction: column; min-height: 78px; padding: 9px 10px; border: 1px solid #c5d9e5; border-top: 3px solid #32759b; border-radius: 8px; background: #f9fcffbb; }
  .metrics > div[data-group=solar] { border-top-color: #bd821e; }
  .metrics > div[data-group=geo] { border-top-color: #2c766d; }
  .metrics > div[data-group=vhf] { border-top-color: #934c87; }
  dt { color: #456074; font-size: .77rem; }
  dd { order: -1; margin: 0 0 2px; color: #163c53; font-size: 1.3rem; font-weight: 800; overflow-wrap: anywhere; }
  dd.text-reading { font-size: 1.02rem; line-height: 1.35; }
  dd small { font-size: .72rem; font-weight: 600; }
  .conditions { display: grid; gap: 12px; min-width: 0; }
  table { width: 100%; border-collapse: collapse; font-size: .85rem; }
  caption { text-align: left; font-weight: 750; color: #173b56; margin-bottom: 5px; }
  th, td { padding: 5px 7px; text-align: start; border-bottom: 1px solid #c4d8e3; }
  thead { background: #e5f0f6; }
  tbody tr:nth-child(odd) { background: #f8fcff8a; }
  .reading { display: inline-block; background: #e3eef5; color: #3c5466; border-radius: 5px; padding: 1px 6px; font-weight: 650; }
  .reading[data-tone=good] { background: #d6efe1; color: #205b43; }
  .reading[data-tone=fair] { background: #ffeab6; color: #714900; }
  .reading[data-tone=poor] { background: #f8dddd; color: #832d33; }
  .map-link, img { display: block; width: 100%; height: auto; border-radius: 8px; }
  @media (min-width: 1100px) {
    .overview-grid { align-items: stretch; }
    .metrics { grid-template-columns: repeat(5, minmax(0, 1fr)); grid-auto-rows: 1fr; }
  }
  @media (max-width: 900px) { .overview-grid { grid-template-columns: 1fr; } }
</style>
