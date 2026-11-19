<script lang="ts">
  import { dxccFlags } from './dxccFlags'
  import { t, locale } from './i18n.svelte'
  import { type Spot } from './spotTools'

  type DXCC = { Code: number; Name: string; Prefix: string; Continent: string; CQ: number; ITU: number; ReferenceGrid: string }
  let { spots }: { spots: Spot[] } = $props()
  let entities = $state<DXCC[]>([])
  let catalogFailed = $state(false)
  let name = $state('')
  let dx = $state('')
  let de = $state('')
  let cq = $state('')
  let itu = $state('')
  let grid = $state('')
  let special = $state(false)
  let selected = $state<{ call?: string; code?: number }>()
  let history = $state<Spot[]>()
  let activity = $state<Record<string, number>>({})
  let historyFailed = $state(false)
  let historyRequest: AbortController | undefined
  let utc = $derived(new Intl.DateTimeFormat(locale(), {dateStyle:'short', timeStyle:'short', timeZone:'UTC'}))
  const continents = ['AF', 'AN', 'AS', 'EU', 'NA', 'OC', 'SA']
  const calls = $derived([...new Set(spots.map(spot => spot.Call))].sort())
  const suggestions = $derived(calls.filter(call => call.toUpperCase().startsWith(name.trim().toUpperCase())).slice(0, 12))
  const recent = $derived(spots.filter(spot =>
    (!name.trim() || spot.Call.toUpperCase().startsWith(name.trim().toUpperCase())) &&
    (!dx || spot.Continent === dx) && (!de || spot.SpotterContinent === de) &&
    (!cq || spot.CQ === Number(cq)) && (!itu || spot.ITU === Number(itu)) &&
    (!grid.trim() || spot.Grids?.some(value => value.startsWith(grid.trim().toUpperCase()))) &&
    (!special || spot.Special)
  ))
  const results = $derived.by(() => {
    const unique = new Map<string, Spot>()
    for (const spot of recent) if (!unique.has(spot.Call)) unique.set(spot.Call, spot)
    return [...unique.values()]
  })
  const entityCodes = $derived(new Set(entities.map(item => item.Code)))
  const entity = $derived(entities.find(entry => entry.Code === selected?.code))
  const bandActivity = $derived(Object.entries(activity).filter(([band]) => band !== 'Altre'))

  $effect(() => {
    const controller = new AbortController()
    fetch('/api/dxcc', {signal: controller.signal}).then(async response => {
      if (!response.ok) throw new Error()
      entities = await response.json()
    }).catch(() => { if (!controller.signal.aborted) catalogFailed = true })
    return () => { controller.abort(); historyRequest?.abort() }
  })

  async function openHistory(call?: string, code?: number) {
    historyRequest?.abort()
    if (!call && !code) { selected = undefined; return }
    const controller = new AbortController()
    historyRequest = controller
    selected = { call, code }
    history = undefined
    activity = {}
    historyFailed = false
    try {
      const params = new URLSearchParams(call ? {call} : {dxcc: String(code)})
      const response = await fetch('/api/spot-history?' + params, {signal: controller.signal})
      if (!response.ok) throw new Error()
      const data: { Spots: Spot[]; Bands: Record<string, number> } = await response.json()
      if (!controller.signal.aborted) { history = data.Spots; activity = data.Bands }
    } catch {
      if (!controller.signal.aborted) historyFailed = true
    }
  }
</script>

<section class="glass-panel explorer" aria-label={t('spotResearch')}>
  <div class="search-options">
    <form class="field call-search" onsubmit={event => { event.preventDefault(); if (/^[a-z0-9][a-z0-9/-]{1,23}$/i.test(name.trim())) void openHistory(name.trim().toUpperCase()) }}>
      <label for="spot-research-callsign">{t('spotCallPrefix')}</label>
      <div><input id="spot-research-callsign" type="search" bind:value={name} list="spot-research-calls" maxlength="24" autocomplete="off" />
        <button type="submit" disabled={!/^[a-z0-9][a-z0-9/-]{1,23}$/i.test(name.trim())}>{t('spotHistory')}</button></div>
      <datalist id="spot-research-calls">{#each suggestions as call}<option value={call}></option>{/each}</datalist>
    </form>
    <label class="field">{t('spotDXCCDirectory')}
      <select value={selected?.code ?? ''} onchange={event => void openHistory(undefined, Number(event.currentTarget.value))} disabled={!entities.length}>
        <option value="">{t('spotChooseEntity')}</option>
        {#each entities as item (item.Code)}<option value={item.Code}>{item.Name} · {item.Prefix || item.Code}</option>{/each}
      </select>
    </label>
  </div>
  {#if catalogFailed}<p role="status">{t('archiveDown')}</p>{/if}
  <div class="field-grid">
    <label class="field">{t('spotDXContinent')}<select bind:value={dx}><option value="">{t('allShort')}</option>{#each continents as continent}<option>{continent}</option>{/each}</select></label>
    <label class="field">{t('spotterContinent')}<select bind:value={de}><option value="">{t('allShort')}</option>{#each continents as continent}<option>{continent}</option>{/each}</select></label>
    <label class="field">{t('spotCQZone')}<select bind:value={cq}><option value="">{t('allShort')}</option>{#each Array.from({ length: 40 }, (_, i) => i + 1) as zone}<option value={String(zone)}>{zone}</option>{/each}</select></label>
    <label class="field">{t('spotITUZone')}<select bind:value={itu}><option value="">{t('allShort')}</option>{#each Array.from({ length: 90 }, (_, i) => i + 1) as zone}<option value={String(zone)}>{zone}</option>{/each}</select></label>
    <label class="field">{t('spotReportedGrid')}<input bind:value={grid} placeholder="JN55" maxlength="8" /></label>
    <label class="check"><input bind:checked={special} type="checkbox" /> {t('spotSpecialPrefixes')}</label>
  </div>
  <div class="research-grid">
    <section aria-label={t('spotRecentResults')}>
      <h2>{t('spotRecentResults')}</h2>
      {#if results.length}
        <div class="overflow">
          <table>
            <thead><tr><th scope="col">{t('thDx')}</th><th scope="col">DXCC</th><th scope="col">CQ / ITU</th><th scope="col">QRA</th></tr></thead>
            <tbody>{#each results as spot (spot.Call)}
              <tr>
                <td><button type="button" class="text-button" onclick={() => void openHistory(spot.Call)}>{spot.Call}</button></td>
                <td>{#if dxccFlags[spot.DXCC]}<img src={'/flags/' + dxccFlags[spot.DXCC] + '.svg'} alt="" width="21" height="15" loading="lazy" />{/if}
                  {#if entityCodes.has(spot.DXCC)}<button type="button" class="text-button" onclick={() => void openHistory(undefined, spot.DXCC)}>{spot.Entity || spot.DXCC}</button>{:else}{spot.Entity || '—'}{/if}
                </td>
                <td>{spot.CQ || '—'} / {spot.ITU || '—'}</td>
                <td>{spot.Grids?.join(', ') || '—'}</td>
              </tr>
            {/each}</tbody>
          </table>
        </div>
      {:else}<p>{t('noMatch')}</p>{/if}
    </section>
    {#if selected}
      <section class="history" aria-label={t('spotHistory')}>
        <div class="history-head"><h2>{entity?.Name || selected.call || 'DXCC ' + selected.code}</h2><button type="button" onclick={() => { historyRequest?.abort(); selected = undefined }}>{t('clear')}</button></div>
        {#if entity}
          <p>{[entity.Prefix, entity.Continent, entity.CQ && 'CQ ' + entity.CQ, entity.ITU && 'ITU ' + entity.ITU].filter(Boolean).join(' · ')}</p>
          {#if entity.ReferenceGrid}<p>{t('spotReferenceGrid')} · {entity.ReferenceGrid}</p>{/if}
          {#if bandActivity.length}
            <h3>{t('spotObservedBands')}</h3>
            <ul class="bands">{#each bandActivity as [band, count] (band)}<li>{band} · {count}</li>{/each}</ul>
          {/if}
        {/if}
        <h3>{t('spotHistory')}</h3>
        <p class="muted">{t('spotHistoryLimit')}</p>
        {#if historyFailed}<p role="status">{t('archiveDown')}</p>
        {:else if !history}<p role="status">{t('loading')}</p>
        {:else if history.length}
          <div class="overflow"><table>
            <thead><tr><th scope="col">UTC</th><th scope="col">{t('thDx')}</th><th scope="col">kHz</th><th scope="col">DE</th></tr></thead>
            <tbody>{#each history as spot}
              <tr><td>{utc.format(new Date(spot.At * 1000))}</td><td>{spot.Call}</td><td>{spot.Frequency}</td><td>{spot.Spotter}</td></tr>
            {/each}</tbody>
          </table></div>
        {:else}<p>{t('noRecent')}</p>{/if}
      </section>
    {/if}
  </div>
</section>

<style>
  .explorer { padding: clamp(12px, 2vw, 20px); }
  .search-options, .field-grid, .research-grid { display: grid; gap: 12px; margin-bottom: 14px; }
  .search-options { grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr)); }
  .field-grid { grid-template-columns: repeat(auto-fit, minmax(min(100%, 138px), 1fr)); align-items: end; }
  .research-grid { grid-template-columns: minmax(0, 1.45fr) minmax(0, 1fr); margin-bottom: 0; }
  .research-grid > section:only-child { grid-column: 1 / -1; }
  .field { display: grid; min-width: 0; gap: 4px; font-weight: 650; }
  .field input, .field select { min-width: 0; width: 100%; min-height: 40px; }
  .call-search div, .history-head { display: flex; gap: 6px; align-items: center; }
  .check { display: flex; align-items: center; min-height: 40px; gap: 6px; }
  .check input { width: auto; }
  .history { border-left: 1px solid #7798ae55; padding-left: 12px; min-width: 0; }
  .history-head { justify-content: space-between; flex-wrap: wrap; }
  .history p { margin: .35rem 0; }
  h2, h3 { font-size: 1rem; margin: 0 0 6px; }
  .history h3 { margin-top: 12px; }
  .muted { font-size: .87rem; opacity: .85; }
  .bands { list-style: none; display: flex; flex-wrap: wrap; gap: 6px 14px; margin: .3rem 0; padding: 0; }
  .overflow { overflow: auto; }
  table { width: 100%; border-collapse: collapse; text-align: left; }
  td, th { padding: 6px; border-bottom: 1px solid #7798ae55; }
  td:first-child, th { white-space: nowrap; }
  td img { vertical-align: middle; margin-right: 3px; }
  .text-button { box-shadow: none; padding: 2px; --button-face: transparent; --button-edge: transparent; --button-ink: inherit; text-decoration: underline; text-underline-offset: 3px; }
  @media (max-width: 760px) { .research-grid { grid-template-columns: minmax(0, 1fr); } .history { border-left: 0; padding-left: 0; border-top: 1px solid #7798ae55; padding-top: 12px; } }
</style>
