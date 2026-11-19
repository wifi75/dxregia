<script lang="ts">
  import { onMount } from 'svelte'
  import { t, locale } from './i18n.svelte'

  type Bulletin = {
    At: number; Call: string; Text: string
    SFI?: string; A?: string; K?: string; ExpectedK?: string; Sunspots?: string
    Solar?: string; Geomagnetic?: string; Aurora?: string
  }
  let { view }: { view: 'ann' | 'propagation' } = $props()
  let data = $state<Partial<Record<'ann' | 'wwv' | 'wcy', Bulletin[]>>>({})
  let unavailable = $state<Partial<Record<'ann' | 'wwv' | 'wcy', boolean>>>({})
  let stamp = $derived(new Intl.DateTimeFormat(locale(), {
    timeZone: 'UTC', dateStyle: 'short', timeStyle: 'short',
  }))
  let request: AbortController | undefined

  async function refresh(section: typeof view) {
    request?.abort()
    const controller = new AbortController()
    request = controller
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(6000)])
    const kinds: ('ann' | 'wwv' | 'wcy')[] = section === 'ann' ? ['ann'] : ['wwv', 'wcy']
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
  }

  $effect(() => { void refresh(view) })
  onMount(() => {
    const timer = setInterval(() => { if (!document.hidden) void refresh(view) }, 30000)
    return () => { clearInterval(timer); request?.abort() }
  })
</script>

{#snippet list(kind: 'ann' | 'wwv' | 'wcy', title: string)}
  <section class="glass-panel bulletins" aria-label={title}>
    <h2>{title}</h2>
    {#if unavailable[kind]}<p role="status">{t('bulletinsUnavailable')}</p>{/if}
    {#if data[kind]?.length}
      <ul>
        {#each data[kind] ?? [] as row}
          <li>
            <header><strong>{row.Call}</strong><time datetime={new Date(row.At * 1000).toISOString()}>{stamp.format(row.At * 1000)} UTC</time></header>
            {#if kind !== 'ann'}
              <dl>
                {#each [['SFI', row.SFI], ['A', row.A], ['K', row.K], ['Kp (exp.)', row.ExpectedK], ['R', row.Sunspots], ['SA', row.Solar], ['GMF', row.Geomagnetic], ['Au', row.Aurora]] as [label, value]}
                  {#if value}<div><dt>{label}</dt><dd>{value}</dd></div>{/if}
                {/each}
              </dl>
            {/if}
            {#if row.Text}<p>{row.Text}</p>{/if}
          </li>
        {/each}
      </ul>
    {:else if !unavailable[kind]}
      <p role="status">{data[kind] ? t('bulletinsEmpty') : t('loading')}</p>
    {/if}
  </section>
{/snippet}

{#if view === 'ann'}
  {@render list('ann', t('navAnnouncements'))}
{:else}
  <div class="propagation" aria-label={t('navPropagation')}>
    {@render list('wwv', 'WWV')}
    {@render list('wcy', 'WCY')}
  </div>
{/if}

<style>
  .propagation { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 370px), 1fr)); align-items: start; gap: 12px; }
  .bulletins { padding: 16px; margin-bottom: 12px; }
  h2 { margin: 0 0 10px; font-size: 1.1rem; }
  ul { list-style: none; margin: 0; padding: 0; }
  li { padding: 10px 0; border-top: 1px solid var(--glass-border, #a4bac580); }
  header { display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: .25rem 1rem; }
  time { font-variant-numeric: tabular-nums; font-size: .86rem; }
  p { margin: .3rem 0 0; overflow-wrap: anywhere; }
  dl { display: flex; flex-wrap: wrap; gap: .5rem .9rem; margin: .45rem 0 0; }
  dl div { display: flex; gap: .3rem; }
  dt { font-size: .85rem; align-self: center; }
  dd { margin: 0; font-weight: 650; }
</style>
