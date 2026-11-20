<script lang="ts">
  import { onMount } from 'svelte'
  import { t, locale } from './i18n.svelte'
  import PropagationMap from './PropagationMap.svelte'
  import SolarConditions from './SolarConditions.svelte'

  type Bulletin = {
    At: number; Call: string; Text: string
    SFI?: string; A?: string; K?: string; ExpectedK?: string; Sunspots?: string
    Solar?: string; Geomagnetic?: string; Aurora?: string
  }
  let data = $state<Partial<Record<'wwv' | 'wcy', Bulletin[]>>>({})
  let unavailable = $state<Partial<Record<'wwv' | 'wcy', boolean>>>({})
  let stamp = $derived(new Intl.DateTimeFormat(locale(), {
    timeZone: 'UTC', dateStyle: 'short', timeStyle: 'short',
  }))
  let request: AbortController | undefined
  let showObserved = $state(false)

  async function refresh() {
    if (request) return
    const controller = new AbortController()
    request = controller
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(6000)])
    const kinds = ['wwv', 'wcy'] as const
    await Promise.all(kinds.map(async kind => {
      try {
        const response = await fetch('/api/cluster-bulletins/' + kind, { signal })
        if (!response.ok) throw new Error()
        const rows = await response.json() as Bulletin[]
        if (controller.signal.aborted) return
        data[kind] = rows
        unavailable[kind] = false
      } catch {
        if (!controller.signal.aborted) unavailable[kind] = true
      }
    }))
    request = undefined
  }

  onMount(() => {
    void refresh()
    const timer = setInterval(() => { if (!document.hidden) void refresh() }, 30000)
    return () => { clearInterval(timer); request?.abort() }
  })
</script>

{#snippet list(kind: 'wwv' | 'wcy', title: string)}
  <section class="glass-panel bulletins" aria-label={title}>
    <h2>{title}</h2>
    {#if unavailable[kind]}<p role="status">{t('bulletinsUnavailable')}</p>{/if}
    {#if data[kind]?.length}
      <ul>
        {#each data[kind] ?? [] as row}
          <li>
            <header><strong>{row.Call}</strong><time datetime={new Date(row.At * 1000).toISOString()}>{stamp.format(row.At * 1000)} UTC</time></header>
            <dl>
                {#each [[t('propSFI'), row.SFI], [t('propIndexA'), row.A], [t('propIndexK'), row.K], ['Kp (exp.)', row.ExpectedK], ['R', row.Sunspots], ['SA', row.Solar], ['GMF', row.Geomagnetic], ['Au', row.Aurora]] as [label, value]}
                  {#if value}<div><dt>{label}</dt><dd>{value}</dd></div>{/if}
                {/each}
            </dl>
            {#if row.Text}<p>{row.Text}</p>{/if}
          </li>
        {/each}
      </ul>
    {:else if !unavailable[kind]}
      <p role="status">{data[kind] ? t('bulletinsEmpty') : t('loading')}</p>
    {/if}
  </section>
{/snippet}

<SolarConditions />
<details class="glass-panel solar observed-spots" bind:open={showObserved}>
  <summary>{t('propTitle')}</summary>
  {#if showObserved}<PropagationMap />{/if}
</details>
<details class="glass-panel solar">
  <summary>{t('propSolarBulletins')}</summary>
  <p class="intro">{t('propSolarIntro')}</p>
  <div class="propagation">
    {@render list('wwv', 'WWV')}
    {@render list('wcy', 'WCY')}
  </div>
</details>

<style>
  .propagation { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 370px), 1fr)); align-items: start; gap: 12px; }
  .bulletins { padding: 16px; margin-bottom: 12px; }
  .solar { padding: 15px; }
  .observed-spots { margin-bottom: 12px; }
  .observed-spots summary { margin-bottom: 4px; }
  summary { cursor: pointer; font-weight: 750; color: #153a59; }
  .intro { margin: 0 0 12px; color: #456074; }
  .solar .intro { margin-top: 8px; }
  .propagation .bulletins { box-shadow: none; margin-bottom: 0; }
  .bulletins[aria-label] ul { display: grid; gap: 7px; }
  h2 { margin: 0 0 10px; font-size: 1.1rem; }
  ul { list-style: none; margin: 0; padding: 0; }
  li { padding: 9px 11px; border: 1px solid #bfd2df; background: #ffffff63; border-radius: 8px; }
  header { display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: .25rem 1rem; }
  time { font-variant-numeric: tabular-nums; font-size: .86rem; }
  p { margin: .3rem 0 0; overflow-wrap: anywhere; white-space: pre-wrap; }
  dl { display: flex; flex-wrap: wrap; gap: .5rem .9rem; margin: .45rem 0 0; }
  dl div { display: flex; gap: .3rem; }
  dt { font-size: .85rem; align-self: center; }
  dd { margin: 0; font-weight: 650; }
</style>
