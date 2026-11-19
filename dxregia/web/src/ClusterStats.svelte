<script lang="ts">
  import { onMount } from 'svelte'
  import { dxccFlags } from './dxccFlags'
  import { lang, t } from './i18n.svelte'
  import BarList from './BarList.svelte'

  type Item = { Name: string; Count: number; DXCC?: number }
  type Report = { Date: string; Spots: {
    Total: number; UniqueDX: number; UniqueSpotters: number; UniqueEntities: number; UniqueModes: number
    BandHours: Record<string, Record<string, number[]>>; Bands: Item[] | null; Entities: Item[] | null; TopDX: Item[] | null; TopSpotters: Item[] | null
  } }
  const continents = ['all', 'AF', 'AN', 'AS', 'EU', 'NA', 'OC', 'SA']
  const bands = ['160m', '80m', '60m', '40m', '30m', '20m', '17m', '15m', '12m', '10m', '6m', '4m', '2m', '70cm', 'SHF']
  const palette = ['#aa4933', '#ac631c', '#926d10', '#8c800b', '#39804c', '#177959', '#1d7783', '#2e6eb0', '#5360ab', '#7554a1', '#994698', '#a54473', '#9b4f61', '#47728d', '#197487']
  const hours = Array.from({ length: 24 }, (_, index) => String(index).padStart(2, '0'))
  const blankHours = Array<number>(24).fill(0)
  let report = $state<Report>()
  let failed = $state(false)
  let continent = $state('all')
  let loading = false

  const number = $derived(new Intl.NumberFormat(lang()))
  const rows = $derived.by(() => {
    const entries = report?.Spots.BandHours[continent] ?? {}
    return [...bands, ...Object.keys(entries).filter(band => !bands.includes(band))].map((band, i) => {
      const values = entries[band] ?? blankHours
      return { band, values, max: Math.max(1, ...values), color: palette[i % palette.length] }
    })
  })
  const bars = (items: Item[] | null) => (items ?? []).map(item => ({ name: item.Name, count: item.Count, flag: item.DXCC ? dxccFlags[item.DXCC] : undefined }))

  async function load() {
    if (loading) return
    loading = true
    try {
      const response = await fetch('/api/cluster-stats', { signal: AbortSignal.timeout(8000) })
      if (!response.ok) throw new Error()
      report = await response.json()
      failed = false
    } catch {
      failed = true
    } finally {
      loading = false
    }
  }

  onMount(() => {
    void load()
    const timer = setInterval(load, 30000)
    return () => clearInterval(timer)
  })
</script>

<main class="cluster-stats">
  <header class="stats-heading">
    <h1>{t('navStats')}</h1>
    <p>{t('statsDay')}{#if report} · <time datetime={report.Date}>{report.Date}</time>{/if}</p>
  </header>
  {#if failed}<p class="glass-panel" role="alert">{t('statsUnavailable')}</p>{/if}
  {#if report}
    {@const spots = report.Spots}
    <div class="totals">
      {#each [['Spot', spots.Total], [t('statsUniqueDx'), spots.UniqueDX], [t('statsUniqueSpotters'), spots.UniqueSpotters], [t('statsEntities'), spots.UniqueEntities], [t('bands'), spots.Bands?.length ?? 0], [t('statsModes'), spots.UniqueModes]] as [label, value]}
        <div class="glass-panel metric"><strong>{number.format(Number(value))}</strong><span>{label}</span></div>
      {/each}
    </div>

    <section class="glass-panel activity" aria-labelledby="band-hour-title">
      <div class="activity-head">
        <h2 id="band-hour-title">{t('statsBandHour')} · UTC</h2>
        <div class="continents" role="group" aria-label={t('statsContinent')}>
          <span>{t('statsContinent')}</span>
          {#each continents as code}
            <button type="button" aria-pressed={continent === code} onclick={() => { continent = code }}>{code === 'all' ? t('allShort') : code}</button>
          {/each}
        </div>
      </div>
      <div class="heat-scroll">
        <table>
          <thead><tr><th scope="col">{t('bands')}</th>{#each hours as hour}<th scope="col">{hour}</th>{/each}</tr></thead>
          <tbody>
            {#each rows as row (row.band)}
              <tr><th scope="row">{row.band}</th>
                {#each row.values as value, hour}
                  <td title={`${row.band} · ${hours[hour]}:00 UTC · ${number.format(value)} spot`}>
                    <span class="cell" style:--heat={value ? row.color : 'transparent'} style:opacity={value ? .25 + .75 * value / row.max : 1}></span>
                    <span class="sr-only">{number.format(value)} spot</span>
                  </td>
                {/each}
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </section>

    <section class="glass-panel card"><h2>{t('statsBands')}</h2><BarList items={bars(spots.Bands)} empty={t('noRecent')} /></section>
    <div class="rankings">
      <section class="glass-panel card"><h2>{t('statsTopDx')}</h2><BarList items={bars(spots.TopDX)} total={spots.Total} empty={t('noRecent')} /></section>
      <section class="glass-panel card"><h2>{t('statsTopSpotters')}</h2><BarList items={bars(spots.TopSpotters)} total={spots.Total} empty={t('noRecent')} /></section>
    </div>
    <section class="glass-panel card entities"><h2>{t('statsEntities')}</h2><BarList items={bars(spots.Entities)} total={spots.Total} empty={t('noRecent')} /></section>
  {:else if !failed}
    <p role="status">{t('loading')}</p>
  {/if}
</main>

<style>
  .cluster-stats { width: min(100% - 28px, 1176px); margin: 20px auto 32px; display: grid; gap: 14px; }
  .stats-heading h1 { margin: 0; font-size: clamp(1.5rem, 4vw, 2rem); }
  .stats-heading p { margin: 4px 0 0; }
  .totals, .rankings { display: grid; gap: 12px; }
  .totals { grid-template-columns: repeat(6, minmax(0, 1fr)); }
  .rankings { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .metric { display: grid; gap: 3px; padding: 12px; text-align: center; }
  .metric strong { font-size: clamp(1.35rem, 3vw, 2rem); font-variant-numeric: tabular-nums; }
  .metric span { font-size: .82rem; }
  .activity, .card { min-width: 0; padding: 14px 18px; }
  h2 { margin: 0 0 12px; font-size: 1.08rem; }
  .activity-head { display: grid; gap: .25rem; }
  .continents { display: flex; flex-wrap: wrap; align-items: center; gap: 5px; margin-bottom: 10px; }
  .continents span { font-size: .82rem; font-weight: 600; margin-right: 4px; }
  .continents button { min-height: 30px; padding: 0 8px; border: 1px solid #aac0d0; border-radius: 6px; background: #f8fcff; color: #143a60; box-shadow: none; }
  .continents button[aria-pressed='true'] { background: #1c5276; color: #fff; border-color: #1c5276; }
  .heat-scroll { overflow-x: auto; }
  table { width: 100%; min-width: 660px; border-collapse: collapse; table-layout: fixed; font-variant-numeric: tabular-nums; }
  th, td { padding: 3px; text-align: center; }
  thead th { background: #dbe9f4; color: #21475f; font-size: .72rem; }
  th:first-child { width: 58px; text-align: left; padding-left: 6px; }
  tbody th { font-size: .78rem; }
  .cell { display: block; height: 13px; border-radius: 3px; border: 1px solid #52718824; background: var(--heat); }
  .sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; clip-path: inset(50%); overflow: hidden; white-space: nowrap; }
  .entities { max-height: 590px; overflow-y: auto; }
  @media (max-width: 760px) { .totals { grid-template-columns: repeat(3, minmax(0, 1fr)); } .rankings { grid-template-columns: minmax(0, 1fr); } }
</style>
