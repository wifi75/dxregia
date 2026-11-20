<script lang="ts">
  import { onMount } from 'svelte'
  import { bandFor } from './spotTools'
  import { locale, t } from './i18n.svelte'
  import { dxccFlags } from './dxccFlags'

  type Spot = { Call: string; Skimmer: string; Mode: 'CW' | 'RTTY'; Frequency: number; At: number; SNR: number; Quality: number; DXCC?: number; DXCCPrefix?: string }
  type Feed = { Enabled: boolean; State: 'disabled' | 'waiting' | 'connecting' | 'connected' | 'conflict' | 'error'; Spots: Spot[] }
  let { admin = false }: { admin?: boolean } = $props()
  let feed = $state<Feed>()
  let failed = $state(false)
  let busy = $state(false)
  let mode = $state('')
  let search = $state('')
  let request: AbortController | undefined
  const status = {
    disabled: 'rbnDisabled', waiting: 'rbnWaiting', connecting: 'rbnConnecting',
    connected: 'rbnConnected', conflict: 'rbnConflict', error: 'rbnUnavailable',
  } as const
  const stamp = $derived(new Intl.DateTimeFormat(locale(), { timeZone: 'UTC', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }))
  const frequency = $derived(new Intl.NumberFormat(locale(), { maximumFractionDigits: 1, useGrouping: false }))
  const regionNames = $derived(new Intl.DisplayNames([locale()], { type: 'region' }))
  const term = $derived(search.trim().toUpperCase())
  const spots = $derived((feed?.Spots ?? []).filter(spot =>
    (!mode || spot.Mode === mode) && (!term || spot.Call.includes(term) || spot.Skimmer.includes(term))))

  async function load(change?: boolean) {
    if (request && change === undefined) return
    request?.abort()
    const controller = new AbortController()
    request = controller
    busy = change !== undefined
    try {
      const response = await fetch('/api/rbn', {
        cache: 'no-store',
        signal: AbortSignal.any([controller.signal, AbortSignal.timeout(6000)]),
        ...(change === undefined ? {} : {
          method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ Enabled: change }),
        }),
      })
      if (!response.ok) throw new Error()
      const data: Feed = await response.json()
      if (controller.signal.aborted) return
      feed = data
      failed = false
    } catch {
      if (!controller.signal.aborted) failed = true
    } finally {
      if (request === controller) {
        busy = false
        request = undefined
      }
    }
  }

  onMount(() => {
    void load()
    const timer = setInterval(() => { if (!document.hidden) void load() }, 10000)
    return () => { clearInterval(timer); request?.abort() }
  })
</script>

<section class="glass-panel rbn-panel" aria-labelledby="rbn-title" data-no-translate>
  <header>
    <div><h2 id="rbn-title">{t('rbnTitle')}</h2><p>{t('rbnHelp')}</p></div>
    {#if admin && feed}
      <button type="button" disabled={busy} onclick={() => load(!feed!.Enabled)}>
        {t(feed.Enabled ? 'rbnDisable' : 'rbnEnable')}
      </button>
    {/if}
  </header>
  {#if admin}<p>{t('rbnTelnet')} <code>SET/SKIMMER CW RTTY</code></p>{/if}
  {#if failed}
    <p role="alert">{t('rbnUnavailable')}</p>
  {:else}
    <p class="status" class:connected={feed?.State === 'connected'} role="status">{t(feed ? (!admin && feed.State === 'conflict' ? 'rbnUnavailable' : status[feed.State]) : 'loading')}</p>
  {/if}
  {#if !admin && feed?.Enabled && feed.Spots.length}
    <div class="filters">
      <label>{t('thMode')} <select bind:value={mode}><option value="">{t('allModes')}</option><option>CW</option><option>RTTY</option></select></label>
      <input type="search" aria-label={t('rbnSearch')} placeholder={t('rbnSearch')} bind:value={search} maxlength="20" />
    </div>
    {#if spots.length}
      <div class="spot-scroll">
        <table>
          <thead><tr>
            <th scope="col">{t('thUtc')}</th><th scope="col">{t('thDx')}</th><th scope="col">{t('thEntity')}</th>
            <th scope="col">{t('rbnFlag')}</th><th scope="col">{t('rbnCountry')}</th><th scope="col">{t('thFreq')} (kHz)</th>
            <th scope="col">{t('thMode')}</th><th scope="col">Skimmer</th><th scope="col">SNR</th><th scope="col">Q</th>
          </tr></thead>
          <tbody>
            {#each spots as spot}
              {@const flag = dxccFlags[spot.DXCC ?? 0]}
              {@const country = flag ? regionNames.of(flag.slice(0, 2).toUpperCase()) : undefined}
              {@const band = bandFor(spot.Frequency)}
              <tr>
                <td><time datetime={new Date(spot.At * 1000).toISOString()}>{stamp.format(spot.At * 1000)}</time></td>
                <th scope="row">{spot.Call}</th><td>{spot.DXCCPrefix || '—'}</td>
                <td>{#if flag}<img src={'/flags/' + flag + '.svg'} alt="" title={country} width="22" height="16" loading="lazy" />{:else}—{/if}</td>
                <td>{country ?? '—'}</td>
                <td>{frequency.format(spot.Frequency)} {#if band}<small class="band">{band}</small>{/if}</td>
                <td><span class="mode" data-kind={spot.Mode}>{spot.Mode}</span></td><td>{spot.Skimmer}</td><td>{spot.SNR} dB</td><td>{spot.Quality || '—'}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p>{t('rbnEmpty')}</p>
    {/if}
  {:else if !admin && feed?.State === 'connected'}
    <p>{t('rbnEmpty')}</p>
  {/if}
  {#if !admin && feed?.Enabled}
    <footer>{t('solarSource')} <a href="https://www.reversebeacon.net/" target="_blank" rel="noopener noreferrer">Reverse Beacon Network</a></footer>
  {/if}
</section>

<style>
  .rbn-panel { border-top: 4px solid #138478; }
  header, .filters { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 8px 14px; }
  h2 { margin: 0; }
  p { margin: 6px 0; }
  .status { display: inline-block; padding: 4px 9px; border-radius: 6px; background: #e5eef4; color: #25465f; font-weight: 650; }
  .status.connected { background: #d0efe3; color: #12563e; }
  .filters { justify-content: flex-start; margin: 8px 0; }
  .filters select, .filters input { padding: 6px 8px; }
  .spot-scroll { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; font-variant-numeric: tabular-nums; white-space: nowrap; }
  th, td { padding: 6px 8px; text-align: start; border-bottom: 1px solid #c4d8e3; }
  thead { background: #d2ebe9; color: #124640; }
  tbody tr:nth-child(even) { background: #e1f1f680; }
  td img { display: block; object-fit: cover; border: 1px solid #17384c26; border-radius: 2px; }
  .mode, .band { display: inline-block; padding: 1px 6px; border-radius: 4px; font-weight: 650; }
  .mode[data-kind=CW] { background: #ffe3a9; color: #68430a; }
  .mode[data-kind=RTTY] { background: #ebdcff; color: #513682; }
  .band { margin-inline-start: 4px; background: #d1edea; color: #174b48; }
  footer { margin-top: 10px; font-size: .82rem; }
</style>
